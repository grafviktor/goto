package sshconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/utils"
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

	tokens := []sshToken{}
	return l.loadFromDataSource(parent, tokens, 0)
}

func (l *Lexer) loadFromDataSource(includeToken sshToken, children []sshToken, currentDepth int) []sshToken {
	currentDepth++
	if currentDepth > maxFileIncludeDepth {
		l.logger.Error("[SSHCONFIG]: Max include depth reached")

		return children
	}

	if includeToken.kind != tokenKind.IncludeFile {
		return children
	}

	rdr, err := newReader(includeToken.value, l.sourceType)
	if err != nil {
		l.logger.Error("[STORAGE] Error opening file: %+v", err)
		panic(err)
	}

	defer func() {
		if closeErr := rdr.Close(); closeErr != nil {
			l.logger.Error("[SSHCONFIG] Error closing file %s: %v", includeToken.value, closeErr)
		}
	}()

	scanner := bufio.NewScanner(rdr)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t")
		var token sshToken
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
			token = sshToken{kind: tokenKind.Unsupported}
		}

		if token.kind == tokenKind.IncludeFile {
			includeTokens := l.handleIncludeToken(token)
			for _, includeToken := range includeTokens {
				children = l.loadFromDataSource(includeToken, children, currentDepth)
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
	trimmedLine := strings.TrimSpace(line)
	key, value, err := parseKeyValuesLine(trimmedLine)
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
	tokens := []sshToken{}
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
	line = strings.TrimSpace(line)
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
