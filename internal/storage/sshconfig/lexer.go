package sshconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/grafviktor/goto/internal/logger"
)

const maxFileIncludeDepth = 16

// FileLexer is responsible for reading and tokenizing an SSH config file.
type FileLexer struct {
	filePath string
	logger   iLogger
}

// NewFileLexer creates a new instance of FileLexer for the given SSH config file path.
func NewFileLexer(filePath string, log iLogger) *FileLexer {
	return &FileLexer{
		filePath: filePath,
		logger:   log,
	}
}

// Tokenize reads the SSH config file and returns a slice of tokens representing the contents.
func (fl *FileLexer) Tokenize() []sshToken {
	parent := sshToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: fl.filePath,
	}

	tokens := []sshToken{}
	return fl.loadFromFile(parent, tokens, 0)
}

func (fl *FileLexer) loadFromFile(includeToken sshToken, children []sshToken, currentDepth int) []sshToken {
	currentDepth++
	if currentDepth > maxFileIncludeDepth {
		fl.logger.Error("[STORAGE]: Max include depth reached")

		return children
	}

	if includeToken.kind != tokenKind.IncludeFile {
		return children
	}

	file, err := os.Open(includeToken.value)
	if err != nil {
		fl.logger.Error("[STORAGE] Error opening file: %+v", err)
		panic(err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger := logger.Get()
			logger.Error("[SSHCONFIG] Error closing file %s: %v", includeToken.value, closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		var token sshToken
		switch {
		case hasPrefixIgnoreCase(line, "User"):
			token = fl.usernameToken(line)
		case hasPrefixIgnoreCase(line, "HostName"):
			token = fl.hostnameToken(line)
		case hasPrefixIgnoreCase(line, "Host"): // Host should be checked after HostName
			token = fl.hostToken(line)
		case hasPrefixIgnoreCase(line, "Port"):
			token = fl.networkPortToken(line)
		case hasPrefixIgnoreCase(line, "Include"):
			token = fl.keyValuesToken(tokenKind.IncludeFile, line)
		case hasPrefixIgnoreCase(line, "IdentityFile"):
			token = fl.identityFileToken(line)
		case hasPrefixIgnoreCase(line, "# GG:GROUP"):
			token = fl.metaDataToken(tokenKind.Group, line)
		case hasPrefixIgnoreCase(line, "# GG:DESCRIPTION"):
			token = fl.metaDataToken(tokenKind.Description, line)
		default:
			token = sshToken{kind: tokenKind.Unsupported}
		}

		if token.kind == tokenKind.IncludeFile {
			includeTokens := fl.handleIncludeToken(token)
			for _, includeToken := range includeTokens {
				children = fl.loadFromFile(includeToken, children, currentDepth)
			}

			continue
		}

		if token.kind != tokenKind.Unsupported {
			children = append(children, token)
		}
	}

	if err := scanner.Err(); err != nil {
		// Ideally, should add a line number which is failing to the error message
		fl.logger.Error("[STORAGE] Error reading file %+v", err)
		panic(err)
	}

	return children
}

func hasPrefixIgnoreCase(str, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(prefix))
}

func (fl *FileLexer) hostToken(line string) sshToken {
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return sshToken{kind: tokenKind.Unsupported}
	}

	return sshToken{
		kind:  tokenKind.Host,
		key:   key,
		value: value,
	}
}

func (fl *FileLexer) usernameToken(rawLine string) sshToken {
	line := strings.TrimSpace(rawLine)
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return sshToken{kind: tokenKind.Unsupported}
	}

	if !sshUsernameRegex.MatchString(value) {
		return sshToken{kind: tokenKind.Unsupported}
	}

	return sshToken{
		kind:  tokenKind.User,
		key:   key,
		value: value,
	}
}

const maxHostnameLength = 253

func (fl *FileLexer) hostnameToken(rawLine string) sshToken {
	line := strings.TrimSpace(rawLine)
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return sshToken{kind: tokenKind.Unsupported}
	}

	if len(value) > maxHostnameLength {
		return sshToken{kind: tokenKind.Unsupported}
	}

	if !hostnameRegex.MatchString(value) {
		return sshToken{kind: tokenKind.Unsupported}
	}

	return sshToken{
		kind:  tokenKind.Hostname,
		key:   key,
		value: value,
	}
}

func (fl *FileLexer) networkPortToken(rawLine string) sshToken {
	line := strings.TrimSpace(rawLine)
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return sshToken{kind: tokenKind.Unsupported}
	}

	networkPort, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return sshToken{kind: tokenKind.Unsupported}
	}

	if !isNetworkPortNumberValid(int(networkPort)) {
		return sshToken{kind: tokenKind.Unsupported}
	}

	return sshToken{
		kind:  tokenKind.NetworkPort,
		key:   key,
		value: value,
	}
}

func (fl *FileLexer) identityFileToken(line string) sshToken {
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return sshToken{kind: tokenKind.Unsupported}
	}

	return sshToken{
		kind:  tokenKind.IdentityFile,
		key:   key,
		value: value,
	}
}

func (fl *FileLexer) handleIncludeToken(token sshToken) []sshToken {
	tokens := []sshToken{}
	if token.kind != tokenKind.IncludeFile {
		return tokens
	}

	filePath := token.value
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(filepath.Dir(fl.filePath), filePath)
	}

	matches, err := filepath.Glob(filePath)
	if err != nil {
		return tokens
	}

	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			continue
		}

		if !isTextFileMime(path) {
			continue
		}

		tokens = append(tokens, sshToken{
			kind:  tokenKind.IncludeFile,
			key:   "Include",
			value: path,
		})
	}

	return tokens
}

func (fl *FileLexer) metaDataToken(kind tokenEnum, line string) sshToken {
	tokenFound := false
	line, tokenFound = strings.CutPrefix(line, "# GG:")
	if !tokenFound {
		return sshToken{kind: tokenKind.Unsupported}
	}

	return fl.keyValuesToken(kind, line)
}

func (fl *FileLexer) keyValuesToken(kind tokenEnum, line string) sshToken {
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return sshToken{kind: tokenKind.Unsupported}
	}

	return sshToken{
		kind:  kind,
		key:   key,
		value: value,
	}
}
