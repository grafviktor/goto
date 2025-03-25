package sshconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const MAX_DEPTH = 16

type FileLexer struct {
	filePath string
	logger   iLogger
}

func NewFileLexer(filePath string, log iLogger) *FileLexer {
	return &FileLexer{
		filePath: filePath,
		logger:   log,
	}
}

func (fl *FileLexer) Tokenize() []Token {
	parent := Token{
		Type:  TokenType.INCLUDE_FILE,
		key:   "Include",
		value: fl.filePath,
	}

	tokens := []Token{}
	return fl.loadFromFile(parent, tokens, 0)
}

func (fl *FileLexer) loadFromFile(includeToken Token, children []Token, currentDepth int) []Token {
	currentDepth++
	if currentDepth > MAX_DEPTH {
		fl.logger.Error("[STORAGE]: Max include depth reached")

		return children
	}

	if includeToken.Type != TokenType.INCLUDE_FILE {
		return children
	}

	file, err := os.Open(includeToken.value)
	if err != nil {
		fl.logger.Error("[STORAGE] Error opening file: %+v", err)
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		var token Token
		switch {
		case strings.HasPrefix(line, "User"):
			token = fl.usernameToken(line)
		case strings.HasPrefix(line, "HostName"):
			token = fl.hostnameToken(line)
		case strings.HasPrefix(line, "Host"): // Host should be checked after HostName
			token = fl.hostToken(line)
		case strings.HasPrefix(line, "Port"):
			token = fl.networkPortToken(line)
		case strings.HasPrefix(line, "Include"):
			token = fl.keyValuesToken(TokenType.INCLUDE_FILE, line)
		case strings.HasPrefix(line, "IdentityFile"):
			token = fl.identityFileToken(line)
		case strings.HasPrefix(line, "# GG:GROUP"):
			token = fl.metaDataToken(TokenType.GROUP, line)
		case strings.HasPrefix(line, "# GG:DESCRIPTION"):
			token = fl.metaDataToken(TokenType.DESCRIPTION, line)
		default:
			token = Token{Type: TokenType.UNSUPPORTED}
		}

		if token.Type == TokenType.INCLUDE_FILE {
			includeTokens := fl.handleIncludeToken(token)
			for _, includeToken := range includeTokens {
				children = fl.loadFromFile(includeToken, children, currentDepth)
			}

			continue
		}

		if token.Type != TokenType.UNSUPPORTED {
			children = append(children, token)
		}
	}

	if err := scanner.Err(); err != nil {
		// TODO: Add line number to error message
		fl.logger.Error("[STORAGE] Error reading file %+v", err)
		panic(err)
	}

	return children
}

func (fl *FileLexer) hostToken(line string) Token {
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	return Token{
		Type:  TokenType.HOST,
		key:   key,
		value: value,
	}
}

func (fl *FileLexer) usernameToken(rawLine string) Token {
	line := strings.TrimSpace(rawLine)
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	if !sshUsernameRegex.MatchString(value) {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	return Token{
		Type:  TokenType.USER,
		key:   key,
		value: value,
	}
}

const MAX_HOSTNAME_LENGTH = 253

func (fl *FileLexer) hostnameToken(rawLine string) Token {
	line := strings.TrimSpace(rawLine)
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	if len(value) > MAX_HOSTNAME_LENGTH {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	if !hostnameRegex.MatchString(value) {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	return Token{
		Type:  TokenType.HOSTNAME,
		key:   key,
		value: value,
	}
}

func (fl *FileLexer) networkPortToken(rawLine string) Token {
	line := strings.TrimSpace(rawLine)
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	networkPort, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	if !isNetworkPortNumberValid(int(networkPort)) {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	return Token{
		Type:  TokenType.NETWORK_PORT,
		key:   key,
		value: value,
	}
}

func (fl *FileLexer) identityFileToken(line string) Token {
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	return Token{
		Type:  TokenType.IDENTITY_FILE,
		key:   key,
		value: value,
	}
}

func (fl *FileLexer) handleIncludeToken(token Token) []Token {
	tokens := []Token{}
	if token.Type != TokenType.INCLUDE_FILE {
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

		tokens = append(tokens, Token{
			Type:  TokenType.INCLUDE_FILE,
			key:   "Include",
			value: path,
		})
	}

	return tokens
}

func (fl *FileLexer) metaDataToken(tokenType tokenEnum, line string) Token {
	tokenFound := false
	line, tokenFound = strings.CutPrefix(line, "# GG:")
	if !tokenFound {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	return fl.keyValuesToken(tokenType, line)
}

func (fl *FileLexer) keyValuesToken(tokenType tokenEnum, line string) Token {
	key, value, err := parseKeyValuesLine(line)
	if err != nil {
		return Token{Type: TokenType.UNSUPPORTED}
	}

	return Token{
		Type:  tokenType,
		key:   key,
		value: value,
	}
}
