package sshconfig

import (
	"bufio"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/model/sshconfig"
	"github.com/grafviktor/goto/internal/utils"
)

const maxFileIncludeDepth = 16

// Lexer is responsible for reading and tokenizing an SSH config file.
type Lexer struct {
	pathType    string
	currentPath string
	rawData     []byte
	logger      iLogger
}

// NewFileLexer creates a new instance of Lexer for the given SSH config file path.
func NewFileLexer(sshConfigPath string, log iLogger) *Lexer {
	var pathType string
	if utils.IsSupportedURL(sshConfigPath) {
		pathType = pathTypeURL
	} else {
		pathType = pathTypeFile
	}

	return &Lexer{
		pathType:    pathType,
		currentPath: sshConfigPath,
		rawData:     []byte{},
		logger:      log,
	}
}

func (l *Lexer) GetRawData() []byte {
	return l.rawData
}

// Tokenize reads the SSH config file and returns a slice of tokens representing the contents.
func (l *Lexer) Tokenize() ([]SSHToken, error) {
	parent := SSHToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: l.currentPath,
	}

	tokens, err := l.loadFromDataSource(parent, []SSHToken{}, 0)
	if sshconfig.IsUserDefinedPath() && err != nil {
		// That's a bit hacky. If user explicitly set ssh/config file path via env var or CLI flag
		// we should NOT ignore errors occurred during file reading.
		return nil, err
	}

	// In case of default ssh/config file path, we can ignore errors
	return tokens, nil
}

func (l *Lexer) loadFromDataSource(
	includeToken SSHToken,
	children []SSHToken,
	currentDepth int,
) ([]SSHToken, error) {
	currentDepth++
	if currentDepth > maxFileIncludeDepth {
		l.logger.Error("[SSHCONFIG] Max include depth reached")

		return children, nil
	}

	if includeToken.kind != tokenKind.IncludeFile {
		return children, nil
	}

	l.logger.Info("[SSHCONFIG] Loading included file: %s", includeToken.value)
	rdr, err := newReader(includeToken.value, l.pathType)
	if err != nil {
		l.logger.Error("[SSHCONFIG] Error opening file %s: %+v", includeToken.value, err)
		return nil, err
	}

	defer func() {
		if closeErr := rdr.Close(); closeErr != nil {
			l.logger.Error("[SSHCONFIG] Error closing file %s: %v", includeToken.value, closeErr)
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
			includeTokens := l.handleIncludeToken(token)
			for _, includeToken := range includeTokens {
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
	var token SSHToken
	switch {
	case matchToken(line, "User", true):
		token = l.usernameToken(line)
	case matchToken(line, "HostName", true):
		token = l.hostnameToken(line)
	case matchToken(line, "Host", false):
		token = l.hostToken(line)
	case matchToken(line, "Port", true):
		token = l.networkPortToken(line)
	case matchToken(line, "Include", false):
		token = l.keyValuesToken(tokenKind.IncludeFile, line)
	case matchToken(line, "IdentityFile", true):
		token = l.identityFileToken(line)
	case matchToken(line, "# GG:GROUP", true):
		token = l.metaDataToken(tokenKind.Group, line)
	case matchToken(line, "# GG:DESCRIPTION", true):
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

func matchToken(line, prefix string, shouldBeIndented bool) bool {
	trimmedLine := strings.TrimSpace(line)
	if utils.StringEmpty(&trimmedLine) {
		return false
	}

	if len(prefix) >= len(trimmedLine) {
		return false
	}

	if isTokenIndented(line) != shouldBeIndented {
		return false
	}

	if !isTokenFollowedDelimiter(trimmedLine, prefix) {
		return false
	}

	return strings.HasPrefix(strings.ToLower(trimmedLine), strings.ToLower(prefix))
}

func isTokenIndented(str string) bool {
	return len(str) > 0 && (str[0] == ' ' || str[0] == '\t')
}

func isTokenFollowedDelimiter(str, prefix string) bool {
	prefixLen := len(prefix)
	delimiters := []rune{' ', '\t'}

	// Should support metadata token which ends with space or color.
	// For instance "# GG:GROUP value" or "# GG:GROUP: value" are valid
	if strings.HasPrefix(str, "# GG:") {
		delimiters = append(delimiters, ':')
	}

	_, found := lo.Find(delimiters, func(d rune) bool {
		return str[prefixLen] == byte(d)
	})

	return found
}

func (l *Lexer) hostToken(line string) SSHToken {
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  tokenKind.Host,
		key:   key,
		value: value,
	}
}

func (l *Lexer) usernameToken(rawLine string) SSHToken {
	line := strings.TrimSpace(rawLine)
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	if !sshUsernameRegex.MatchString(value) {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  tokenKind.User,
		key:   key,
		value: value,
	}
}

const maxHostnameLength = 253

func (l *Lexer) hostnameToken(rawLine string) SSHToken {
	line := strings.TrimSpace(rawLine)
	key, value, err := parseKeyValuesLine(line)
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
		key:   key,
		value: value,
	}
}

func (l *Lexer) networkPortToken(rawLine string) SSHToken {
	line := strings.TrimSpace(rawLine)
	key, value, err := parseKeyValuesLine(line)
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
		key:   key,
		value: value,
	}
}

func (l *Lexer) identityFileToken(line string) SSHToken {
	trimmedLine := strings.TrimSpace(line)
	key, value, err := parseKeyValuesLine(trimmedLine)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  tokenKind.IdentityFile,
		key:   key,
		value: value,
	}
}

func (l *Lexer) handleIncludeToken(token SSHToken) []SSHToken {
	if token.kind != tokenKind.IncludeFile {
		return []SSHToken{}
	}

	if l.pathType == pathTypeURL {
		return l.includeRemoteFileToken(token.value)
	}

	return l.includeLocalFileToken(token.value)
}

func (l *Lexer) includeLocalFileToken(localPath string) []SSHToken {
	tokens := []SSHToken{}
	if !filepath.IsAbs(localPath) {
		localPath = filepath.Join(filepath.Dir(l.currentPath), localPath)
	}

	matches, err := filepath.Glob(localPath)
	if err != nil {
		return tokens
	}

	for _, path := range matches {
		var info os.FileInfo
		info, err = os.Stat(path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			continue
		}

		if !isTextFileMime(path) {
			continue
		}

		tokens = append(tokens, SSHToken{
			kind:  tokenKind.IncludeFile,
			key:   "Include",
			value: path,
		})
	}

	return tokens
}

func (l *Lexer) includeRemoteFileToken(remotePath string) []SSHToken {
	if utils.IsSupportedURL(remotePath) {
		// If remotePath is already a full URL, use it as is.
		return []SSHToken{{
			kind:  tokenKind.IncludeFile,
			key:   "Include",
			value: remotePath,
		}}
	}

	var err error
	// If remotePath is not a full URL, we need to construct the full URL taking lexer.currentPath as base.
	if path.IsAbs(remotePath) {
		// If remotePath is absolute, extract base URL from lexer.currentPath and try to fetch the file
		// from the server root. I.e. "http://127.0.0.1:8080" + "/path/to/resource".
		var baseURL string
		baseURL, err = utils.ExtractBaseURL(l.currentPath)
		remotePath = baseURL + remotePath
	} else {
		// If remotePath is relative, take the base URL as the directory part of the lexer.currentPath.
		// Example:
		// l.currentPath = "http://127.0.0.1:8080/path/ssh_config"
		// remotePath = "ssh_config_included"
		// Result
		// remotePath = "http://127.0.0.1:8080/path/ssh_config_included"
		var u *url.URL
		u, err = url.Parse(l.currentPath)
		if err == nil {
			u.Path = path.Join(path.Dir(u.Path), remotePath)
			remotePath = u.String()
		}
	}

	if err != nil {
		l.logger.Error("[SSHCONFIG]: Cannot parse resource URL: %v", err)
		return []SSHToken{}
	}

	return []SSHToken{{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: remotePath,
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
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return SSHToken{kind: tokenKind.Unsupported}
	}

	return SSHToken{
		kind:  kind,
		key:   key,
		value: value,
	}
}
