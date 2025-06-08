package sshconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const maxFileIncludeDepth = 16

// Lexer is responsible for reading and tokenizing an SSH config file.
type Lexer struct {
	sourceType string
	source     string
	logger     iLogger
}

// NewFileLexer creates a new instance of Lexer for the given SSH config file path.
func NewFileLexer(filePath string, log iLogger) *Lexer {
	return &Lexer{
		sourceType: "file",
		source:     filePath,
		logger:     log,
	}
}

// Tokenize reads the SSH config file and returns a slice of tokens representing the contents.
func (l *Lexer) Tokenize() []sshToken {
	parent := sshToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: l.source,
	}

	var tokens []sshToken
	return l.loadFromFile(parent, tokens, 0)
}

func (l *Lexer) loadFromFile(includeToken sshToken, children []sshToken, currentDepth int) []sshToken {
	currentDepth++
	if currentDepth > maxFileIncludeDepth {
		l.logger.Error("[SSHCONFIG]: Max include depth reached")

		return children
	}

	if includeToken.kind != tokenKind.IncludeFile {
		return children
	}

	reader, err := newReader(includeToken.value, l.sourceType)
	if err != nil {
		l.logger.Error("[STORAGE] Error opening file: %+v", err)
		panic(err)
	}

	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			l.logger.Error("[SSHCONFIG] Error closing file %s: %v", includeToken.value, closeErr)
		}
	}()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		var token sshToken
		switch {
		case hasPrefixIgnoreCase(line, "User"):
			token = l.usernameToken(line)
		case hasPrefixIgnoreCase(line, "HostName"):
			token = l.hostnameToken(line)
		case hasPrefixIgnoreCase(line, "Host"): // Host should be checked after HostName
			token = l.hostToken(line)
		case hasPrefixIgnoreCase(line, "Port"):
			token = l.networkPortToken(line)
		case hasPrefixIgnoreCase(line, "Include"):
			token = l.keyValuesToken(tokenKind.IncludeFile, line)
		case hasPrefixIgnoreCase(line, "IdentityFile"):
			token = l.identityFileToken(line)
		case hasPrefixIgnoreCase(line, "# GG:GROUP"):
			token = l.metaDataToken(tokenKind.Group, line)
		case hasPrefixIgnoreCase(line, "# GG:DESCRIPTION"):
			token = l.metaDataToken(tokenKind.Description, line)
		default:
			token = sshToken{kind: tokenKind.Unsupported}
		}

		if token.kind == tokenKind.IncludeFile {
			includeTokens := l.handleIncludeToken(token)
			for _, includeToken := range includeTokens {
				children = l.loadFromFile(includeToken, children, currentDepth)
			}

			continue
		}

		if token.kind != tokenKind.Unsupported {
			children = append(children, token)
		}
	}

	if err = scanner.Err(); err != nil {
		// Ideally, should add a line number which is failing to the error message
		l.logger.Error("[SSHCONFIG] Error reading file %+v", err)
		panic(err)
	}

	return children
}

func hasPrefixIgnoreCase(str, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(prefix))
}

func (l *Lexer) hostToken(line string) sshToken {
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

func (l *Lexer) usernameToken(rawLine string) sshToken {
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

func (l *Lexer) hostnameToken(rawLine string) sshToken {
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

func (l *Lexer) networkPortToken(rawLine string) sshToken {
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

func (l *Lexer) identityFileToken(line string) sshToken {
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

func (l *Lexer) handleIncludeToken(token sshToken) []sshToken {
	var tokens []sshToken
	if token.kind != tokenKind.IncludeFile {
		return tokens
	}

	filePath := token.value
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(filepath.Dir(l.source), filePath)
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

func (l *Lexer) metaDataToken(kind tokenEnum, line string) sshToken {
	tokenFound := false
	line, tokenFound = strings.CutPrefix(line, "# GG:")
	if !tokenFound {
		return sshToken{kind: tokenKind.Unsupported}
	}

	return l.keyValuesToken(kind, line)
}

func (l *Lexer) keyValuesToken(kind tokenEnum, line string) sshToken {
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
