package sshconfig

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const MAX_DEPTH = 16

type FileLexer struct {
	filePath string
}

func NewFileLexer(filePath string) *FileLexer {
	return &FileLexer{
		filePath: filePath,
	}
}

func (fl *FileLexer) Tokenize() []Token {
	parent := Token{
		Type:  TokenType.INCLUDE_FILE,
		key:   "Include",
		value: fl.filePath,
	}

	var tokens = []Token{}
	return fl.loadFromFile(parent, tokens, 0)
}

func (fl *FileLexer) loadFromFile(includeToken Token, children []Token, currentDepth int) []Token {
	currentDepth++
	if currentDepth > MAX_DEPTH {
		log.Println("MAX DEPTH EXCEEDED")

		return children
	}

	if includeToken.Type != TokenType.INCLUDE_FILE {
		return children
	}

	file, err := os.Open(includeToken.value)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return children
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
			token = fl.metaInfoToken(TokenType.GROUP, line)
		case strings.HasPrefix(line, "# GG:DESCRIPTION"):
			token = fl.metaInfoToken(TokenType.DESCRIPTION, line)
		default:
			token = Token{Type: TokenType.UNSUPPORTED}
		}

		if token.Type == TokenType.INCLUDE_FILE {
			includeTokens := fl.parseIncludeTokens(token) // TODO: rename to handleIncludeStatement
			for _, includeToken := range includeTokens {
				includedTokens := fl.loadFromFile(includeToken, children, currentDepth)
				children = append(children, includedTokens...)
			}

			continue
		}

		if token.Type != TokenType.UNSUPPORTED {
			children = append(children, token)
		}
	}

	if err := scanner.Err(); err != nil {
		// TODO: Add line number to error message
		fmt.Println("Error reading file:", err)
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

func (fl *FileLexer) parseIncludeTokens(token Token) []Token {
	tokens := []Token{}
	if token.Type != TokenType.INCLUDE_FILE {
		return tokens
	}

	matches, err := filepath.Glob(token.value)
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
			value: token.value,
		})
	}

	return tokens
}

func (fl *FileLexer) metaInfoToken(tokenType tokenEnum, line string) Token {
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
