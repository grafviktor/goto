package sshconfig

import (
	"bufio"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/model/sshconfig"
	"github.com/grafviktor/goto/internal/utils"
)

const maxFileIncludeDepth = 16

// Lexer is responsible for reading and tokenizing an SSH config file.
type Lexer struct {
	rawData    []byte
	logger     iLogger
	rootConfig configSource
}

// NewFileLexer creates a new instance of Lexer for the given SSH config file path.
func NewFileLexer(sshConfigPath string, log iLogger) *Lexer {
	var pathType valueTypeEnum
	if utils.IsSupportedURL(sshConfigPath) {
		pathType = valueTypeURL
	} else {
		pathType = valueTypeFile
	}

	parent := configSource{
		valueType: pathType,
		value:     sshConfigPath,
	}

	return &Lexer{
		rootConfig: parent,
		rawData:    []byte{},
		logger:     log,
	}
}

func (l *Lexer) GetRawData() []byte {
	return l.rawData
}

// Tokenize reads the SSH config file and returns a slice of tokens representing the contents.
func (l *Lexer) Tokenize() ([]SSHToken, error) {
	tokens, err := l.loadFromDataSource(l.rootConfig, []SSHToken{}, 0)
	if sshconfig.IsUserDefinedPath() && err != nil {
		// That's a bit hacky. If user explicitly set ssh/config file path via env var or CLI flag
		// we should NOT ignore errors occurred during file reading.
		return nil, err
	}

	// In case of default ssh/config file path, we can ignore errors
	return tokens, nil
}

func (l *Lexer) loadFromDataSource(
	src configSource,
	children []SSHToken,
	currentDepth int,
) ([]SSHToken, error) {
	currentDepth++
	if currentDepth > maxFileIncludeDepth {
		l.logger.Error("[SSHCONFIG] Max include depth reached")

		return children, nil
	}

	l.logger.Info("[SSHCONFIG] Load file: %s", src.value)
	rdr, err := newReader(src.value, src.valueType)
	if err != nil {
		l.logger.Error("[SSHCONFIG] Error opening file %s: %+v", src.value, err)
		return nil, err
	}

	defer func() {
		if closeErr := rdr.Close(); closeErr != nil {
			l.logger.Error("[SSHCONFIG] Error closing file %s: %v", src.value, closeErr)
		}
	}()

	scanner := bufio.NewScanner(rdr)

	for scanner.Scan() {
		line := scanner.Text()
		if shouldSkipLine(line) {
			continue
		}

		line = stripInlineComments(line)
		token := l.readToken(line)

		if token.kind == tokenKind.IncludeFile {
			for _, includeToken := range l.handleIncludeToken(token, src) {
				children, err = l.loadFromDataSource(includeToken, children, currentDepth)
				if err != nil {
					return children, err
				}
			}

			continue
		}

		l.rawData = append(l.rawData, []byte(line+"\n")...)
		if token.kind != tokenKind.Unsupported {
			children = append(children, token)
		}
	}

	if err = scanner.Err(); err != nil {
		// Ideally, should add a line number which is failing to the error message
		l.logger.Error("[SSHCONFIG] Error reading file %+v", err)
	}

	return children, err
}

func (l *Lexer) readToken(line string) SSHToken {
	line = strings.TrimSpace(line)

	var token SSHToken
	switch {
	case matchToken(line, "User"):
		token = l.usernameToken(line)
	case matchToken(line, "HostName"):
		token = l.hostnameToken(line)
	case matchToken(line, "Host"):
		token = l.hostToken(line)
	case matchToken(line, "Port"):
		token = l.networkPortToken(line)
	case matchToken(line, "Include"):
		token = l.keyValuesToken(tokenKind.IncludeFile, line)
	case matchToken(line, "IdentityFile"):
		token = l.identityFileToken(line)
	case matchToken(line, "# GG:GROUP"):
		token = l.metaDataToken(tokenKind.Group, line)
	case matchToken(line, "# GG:DESCRIPTION"):
		token = l.metaDataToken(tokenKind.Description, line)
	default:
		token = SSHToken{kind: tokenKind.Unsupported}
	}
	return token
}

func shouldSkipLine(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	if utils.StringEmpty(&trimmedLine) {
		return true
	}

	if strings.HasPrefix(trimmedLine, "# GG:") {
		// This is a metadata comments, which should be processed
		return false
	}

	if strings.HasPrefix(trimmedLine, "#") {
		return true
	}

	return false
}

func stripInlineComments(line string) string {
	line = strings.TrimRight(line, " \t")

	if strings.Contains(line, "# GG:") {
		return line
	}

	parts := strings.Split(line, "#")
	if len(parts) > 0 {
		return parts[0]
	}

	return line
}

func matchToken(line, token string) bool {
	// Remove leading and trailing whitespace: "  User root" -> "User root"
	line = strings.TrimSpace(line)
	if utils.StringEmpty(&line) {
		return false
	}

	// If token prefix is longer than the line itself, it can't match.
	if len(token) >= len(line) {
		return false
	}

	if !isTokenFollowedDelimiter(line, token) {
		return false
	}

	return strings.HasPrefix(strings.ToLower(line), strings.ToLower(token))
}

func isTokenFollowedDelimiter(line, token string) bool {
	prefixLen := len(token)
	delimiters := []byte{' ', '\t'}

	// Should support metadata token which ends with space or colon.
	// For instance "# GG:GROUP value" or "# GG:GROUP: value" are valid
	if strings.HasPrefix(line, "# GG:") {
		delimiters = append(delimiters, ':')
	}

	// Check if token is followed one of the delimiters. For example,
	// "User root" and "User\troot" are valid, but "User=root" is not.
	_, found := lo.Find(delimiters, func(d byte) bool {
		return line[prefixLen] == d
	})

	return found
}

func (l *Lexer) hostToken(line string) SSHToken {
	_, value, err := parseKeyValuesLine(line)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  tokenKind.Host,
		value: value,
	}
}

func (l *Lexer) usernameToken(line string) SSHToken {
	_, value, err := parseKeyValuesLine(line)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	if !sshUsernameRegex.MatchString(value) {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  tokenKind.User,
		value: value,
	}
}

const maxHostnameLength = 253

func (l *Lexer) hostnameToken(line string) SSHToken {
	_, value, err := parseKeyValuesLine(line)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	if len(value) > maxHostnameLength {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	if !hostnameRegex.MatchString(value) {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  tokenKind.Hostname,
		value: value,
	}
}

func (l *Lexer) networkPortToken(line string) SSHToken {
	_, value, err := parseKeyValuesLine(line)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	networkPort, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	if !isNetworkPortNumberValid(int(networkPort)) {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  tokenKind.NetworkPort,
		value: value,
	}
}

func (l *Lexer) identityFileToken(line string) SSHToken {
	trimmedLine := strings.TrimSpace(line)
	_, value, err := parseKeyValuesLine(trimmedLine)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  tokenKind.IdentityFile,
		value: value,
	}
}

func (l *Lexer) handleIncludeToken(token SSHToken, parent configSource) []configSource {
	switch {
	// Order matters! Check for tilde prefix first.
	case strings.HasPrefix(token.value, "~"):
		// If path starts from tilde, we load the included file from the local file system.
		// This allows to set some user default values, even if config is stored remotely.
		expandedPath := l.expandTildePath(token.value)
		return l.includeLocalFileToken(expandedPath, parent)
	case parent.valueType == valueTypeURL:
		return l.includeRemoteFileToken(token.value, parent)
	default:
		return l.includeLocalFileToken(token.value, parent)
	}
}

func (l *Lexer) includeLocalFileToken(localPath string, parent configSource) []configSource {
	sources := []configSource{}

	if !filepath.IsAbs(localPath) {
		localPath = filepath.Join(filepath.Dir(parent.value), localPath)
	}

	matches, err := filepath.Glob(localPath)
	if err != nil {
		l.logger.Error("[SSHCONFIG] Cannot process Include pattern %s: %v", localPath, err)
		return sources
	}

	if len(matches) == 0 {
		l.logger.Error("[SSHCONFIG] No files match Include pattern: %s", localPath)
		return sources
	}

	for _, path := range matches {
		var info os.FileInfo
		info, err = os.Stat(path)
		if err != nil {
			l.logger.Error("[SSHCONFIG] Cannot access file %s: %v", path, err)
			continue
		}

		if info.IsDir() {
			l.logger.Error("[SSHCONFIG] Path is a directory: %s", path)
			continue
		}

		if !isTextFileMime(path) {
			l.logger.Error("[SSHCONFIG] Not a text file: %s", path)
			continue
		}

		sources = append(sources, configSource{
			value:     path,
			valueType: valueTypeFile,
		})
	}

	return sources
}

func (l *Lexer) expandTildePath(localPath string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		l.logger.Error("[SSHCONFIG]: Cannot expand tilde in path: %v", err)
		return localPath
	}

	if home == "" {
		l.logger.Error("[SSHCONFIG]: Cannot find user home directory to expand tilde in path: %s", localPath)
		return localPath
	}

	// Linux:   "~/.path/config" => ".path/config."
	// Windows: "~\.path\config" => ".path\config".
	rest := strings.TrimLeft(strings.TrimPrefix(localPath, "~"), "/\\")
	if rest == "" {
		return home
	}

	// Linux:   /home/user, .  .path/config => /home/user/.path/config
	// Windows: C:\Users\user, .path\config => C:\Users\user\.path\config
	return filepath.Join(home, rest)
}

func (l *Lexer) includeRemoteFileToken(remotePath string, parent configSource) []configSource {
	if utils.IsSupportedURL(remotePath) {
		// If remotePath is already a full URL, use it as is.
		return []configSource{{
			valueType: valueTypeURL,
			value:     remotePath,
		}}
	}

	var err error
	// If remotePath is not a full URL, we need to construct the full URL taking parent.path as base.
	if path.IsAbs(remotePath) {
		// If remotePath is absolute, extract base URL from lexer.currentPath and try to fetch the file
		// from the server root. I.e. "http://127.0.0.1:8080" + "/path/to/resource".
		var baseURL string
		baseURL, err = utils.ExtractBaseURL(parent.value)
		remotePath = baseURL + remotePath
	} else {
		// If remotePath is relative, take the base URL as the directory part of the parent.path.
		// Example:
		// parent.path = "http://127.0.0.1:8080/path/ssh_config"
		// remotePath = "ssh_config_included"
		// Result
		// remotePath = "http://127.0.0.1:8080/path/ssh_config_included"
		var u *url.URL
		u, err = url.Parse(parent.value)
		if err == nil {
			u.Path = path.Join(path.Dir(u.Path), remotePath)
			remotePath = u.String()
		}
	}

	if err != nil {
		l.logger.Error("[SSHCONFIG]: Cannot parse resource URL: %v", err)
		return []configSource{}
	}

	return []configSource{{
		valueType: valueTypeURL,
		value:     remotePath,
	}}
}

func (l *Lexer) metaDataToken(kind tokenEnum, line string) SSHToken {
	line = strings.TrimSpace(line)
	line, tokenFound := strings.CutPrefix(line, "# GG:")
	if !tokenFound {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return l.keyValuesToken(kind, line)
}

func (l *Lexer) keyValuesToken(kind tokenEnum, line string) SSHToken {
	_, value, err := parseKeyValuesLine(line)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  kind,
		value: value,
	}
}

/*
Regex to match exactly two or more words.

"hello world",     // Valid.
"  foo   bar  ",   // Valid.
"oneword",         // Invalid.
"three word test", // Valid.
"",                // Invalid.
*/
var twoWordsRegex = regexp.MustCompile(`^(\S+)\s+(.+)$`)

func parseKeyValuesLine(line string) (string, string, error) {
	matches := twoWordsRegex.FindStringSubmatch(line)
	// Ideally it should be a loop, not regex.
	if len(matches) > 1 {
		return matches[1], matches[2], nil
	}

	return "", "", errors.New("not a key value string")
}
