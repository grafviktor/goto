package sshconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

func TestLexer_Tokenize_General(t *testing.T) {
	const config = `
Host test
    # Just a comment
    # GG:GROUP mock_group
    # GG:DESCRIPTION mock_description
		Unsupported
    HostName example.com # comment
    User alice
    Port 2222 # comment
    IdentityFile ~/.ssh/id_rsa
	HostkeyAlgorithms +ssh-dss,ssh-rsa
`
	lex := &Lexer{
		pathType:    "string",
		currentPath: config,
		logger:      &mocklogger.Logger{},
	}
	parent := SSHToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: config,
	}
	tokens, _ := lex.loadFromDataSource(parent, nil, 0)

	wantKinds := []tokenEnum{
		tokenKind.Host,
		tokenKind.Group,
		tokenKind.Description,
		tokenKind.Hostname,
		tokenKind.User,
		tokenKind.NetworkPort,
		tokenKind.IdentityFile,
	}

	wantValues := []string{
		"test",
		"mock_group",
		"mock_description",
		"example.com",
		"alice",
		"2222",
		"~/.ssh/id_rsa",
	}

	if len(tokens) != len(wantKinds) {
		t.Fatalf("expected %d tokens, got %d", len(wantKinds), len(tokens))
	}

	for i, tk := range tokens {
		if tk.kind != wantKinds[i] {
			t.Errorf("token %d: expected kind %v, got %v", i, wantKinds[i], tk.kind)
		}

		if tk.value != wantValues[i] {
			t.Errorf("token %d: expected value %q, got %q", i, wantValues[i], tk.value)
		}
	}
}

func TestLexer_Tokenize_Unsupported(t *testing.T) {
	const config = `
Host test
    UnknownKey value
`
	lex := &Lexer{
		pathType:    "string",
		currentPath: config,
		logger:      &mocklogger.Logger{},
	}
	parent := SSHToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: config,
	}
	tokens, _ := lex.loadFromDataSource(parent, nil, 0)

	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].kind != tokenKind.Host {
		t.Errorf("expected Host token, got %v", tokens[0].kind)
	}
}

func TestLexer_Tokenize_InvalidUser(t *testing.T) {
	const config = `
User invalid!user
`
	lex := &Lexer{
		pathType:    "string",
		currentPath: config,
		logger:      &mocklogger.Logger{},
	}
	parent := SSHToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: config,
	}
	tokens, _ := lex.loadFromDataSource(parent, nil, 0)
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for invalid user, got %d", len(tokens))
	}
}

func TestLexer_Tokenize_InvalidPort(t *testing.T) {
	const config = `
Port notaport
`
	lex := &Lexer{
		pathType:    "string",
		currentPath: config,
		logger:      &mocklogger.Logger{},
	}
	parent := SSHToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: config,
	}
	tokens, _ := lex.loadFromDataSource(parent, nil, 0)
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for invalid port, got %d", len(tokens))
	}
}

func TestLexer_Tokenize_IncludeFile(t *testing.T) {
	// Explanation of what's going on here:
	// Create a starting point token, which includes a file (includedConfig1).
	// That file includes another file (includedConfig2).
	// The second file contains a Host definition.
	// We expect the lexer to follow the includes and return the Host token.

	// tmpDir will be automatically cleaned removed after the test
	tmpDir := t.TempDir()
	includedConfig1 := filepath.Join(tmpDir, "config_included1")
	includedConfig2 := filepath.Join(tmpDir, "config_included2")

	// Create a config file with a single line - Include pointing to another file.
	content := fmt.Sprintf("Include %s\n", includedConfig2)
	if err := os.WriteFile(includedConfig1, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Create a config file with a single line - Host.
	content = "Host mock-included-host\n"
	if err := os.WriteFile(includedConfig2, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	lex := &Lexer{
		pathType:    "file",
		currentPath: filepath.Join(tmpDir, "config"),
		logger:      &mocklogger.Logger{},
	}

	// That's our starting token which includes the file we created above.
	parent := SSHToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: includedConfig1,
	}

	tokens, _ := lex.loadFromDataSource(parent, nil, 0)
	require.Len(t, tokens, 1, "expected 1 token for included host")
}

func TestLexer_Tokenize_IncludeDepthLimit(t *testing.T) {
	// Simulate include depth limit by recursive call
	lex := &Lexer{
		pathType:    "string",
		currentPath: "irrelevant",
		logger:      &mocklogger.Logger{},
	}
	parent := SSHToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: "irrelevant",
	}
	// Should not panic or recurse infinitely
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not panic on max include depth, got %v", r)
		}
	}()
	_, _ = lex.loadFromDataSource(parent, nil, maxFileIncludeDepth)
}

func Test_matchToken(t *testing.T) {
	tests := []struct {
		str        string
		prefix     string
		identation bool
		want       bool
	}{
		{"Host test", "host", false, true},
		{"\tHost test", "host", false, false},
		{"HOST test", "host", false, true},
		{"\tUser alice", "USER", true, true},
		{" Port 22", "port", true, true},
		{"\tIdentityFile foo", "identityfile", true, true},
		{"\tSomethingElse", "host", true, false},
		{"\t# GG:GROUP test", "# GG:GROUP", true, true},
		{"\t# GG:GROUP: test", "# GG:GROUP", true, true},
	}
	for _, tt := range tests {
		if got := matchToken(tt.str, tt.prefix, tt.identation); got != tt.want {
			t.Errorf("hasPrefixIgnoreCase(%q, %q, %v) = %v, want %v", tt.str, tt.prefix, tt.identation, got, tt.want)
		}
	}
}

func TestLexer_handleIncludeToken_localFile(t *testing.T) {
	// tmpDir will be automatically cleaned removed after the test
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config_included")
	// Create a config file with a single line - Host. That's enough for the test.
	content := "Host mock-local-host\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	lex := &Lexer{
		pathType:    "file",
		currentPath: filepath.Join(tmpDir, "config"),
		logger:      &mocklogger.Logger{},
	}

	// That's our starting token which includes the file we created above.
	token := SSHToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: filepath.Join(tmpDir, "config_included"),
	}

	tokens := lex.handleIncludeToken(token)
	require.Len(t, tokens, 1, "expected 1 include token")
}

func TestLexer_handleIncludeToken_remoteFile(t *testing.T) {
	// That's our starting token which includes the file we created above.
	tests := []struct {
		name             string
		baseURL          string
		sourceTokenURL   string
		expectedTokenURL string
	}{
		{
			name:             "full url",
			baseURL:          "http://127.0.0.1/config",
			sourceTokenURL:   "http://127.0.0.1/config",
			expectedTokenURL: "http://127.0.0.1/config",
		},
		{
			name:    "absolute path",
			baseURL: "http://127.0.0.1/config",
			// That's an absolute path, so it should be fetched from the server root.
			sourceTokenURL:   "/path/to/config",
			expectedTokenURL: "http://127.0.0.1/path/to/config",
		},
		{
			name:    "relative path",
			baseURL: "http://127.0.0.1/config",
			// That's a relative path, so it should be fetched from baseURL + path.
			sourceTokenURL:   "path/to/config",
			expectedTokenURL: "http://127.0.0.1/path/to/config",
		},
		{
			name:    "absolute path, base URL with trailing slash",
			baseURL: "http://127.0.0.1/config/",
			// That's an absolute path, so it should be fetched from the server root.
			sourceTokenURL:   "/path/to/config",
			expectedTokenURL: "http://127.0.0.1/path/to/config",
		},
		{
			name:    "relative path, base URL with trailing slash",
			baseURL: "http://127.0.0.1/config/",
			// That's a relative path, so it should be fetched from baseURL + path.
			sourceTokenURL:   "path/to/config",
			expectedTokenURL: "http://127.0.0.1/config/path/to/config",
		},
		{
			name:             "relative path starting with ./",
			baseURL:          "http://127.0.0.1/config",
			sourceTokenURL:   "./path/to/config",
			expectedTokenURL: "http://127.0.0.1/path/to/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := &Lexer{
				pathType:    pathTypeURL,
				currentPath: tt.baseURL,
				logger:      &mocklogger.Logger{},
			}

			startingPointToken := SSHToken{
				kind:  tokenKind.IncludeFile,
				key:   "Include",
				value: tt.sourceTokenURL,
			}

			tokens := lex.handleIncludeToken(startingPointToken)
			require.Len(t, tokens, 1, "expected 1 include token")
			require.Equal(t, tt.expectedTokenURL, tokens[0].value, "unexpected included file URL")
		})
	}
}

func TestLexer_MetaDataToken(t *testing.T) {
	lex := &Lexer{}
	line := "# GG:GROUP mock_group"
	token := lex.metaDataToken(tokenKind.Group, line)
	require.Equal(t, "GROUP", token.key, "wrong token key")
	require.Equal(t, "mock_group", token.value, "wrong token value")
}
