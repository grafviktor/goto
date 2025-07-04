package sshconfig

import (
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
		sourceType: "string",
		source:     config,
		logger:     &mocklogger.Logger{},
	}
	parent := sshToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: config,
	}
	tokens := lex.loadFromDataSource(parent, nil, 0)

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
		sourceType: "string",
		source:     config,
		logger:     &mocklogger.Logger{},
	}
	parent := sshToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: config,
	}
	tokens := lex.loadFromDataSource(parent, nil, 0)

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
		sourceType: "string",
		source:     config,
		logger:     &mocklogger.Logger{},
	}
	parent := sshToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: config,
	}
	tokens := lex.loadFromDataSource(parent, nil, 0)
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for invalid user, got %d", len(tokens))
	}
}

func TestLexer_Tokenize_InvalidPort(t *testing.T) {
	const config = `
Port notaport
`
	lex := &Lexer{
		sourceType: "string",
		source:     config,
		logger:     &mocklogger.Logger{},
	}
	parent := sshToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: config,
	}
	tokens := lex.loadFromDataSource(parent, nil, 0)
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for invalid port, got %d", len(tokens))
	}
}

func TestLexer_Tokenize_IncludeDepthLimit(t *testing.T) {
	// Simulate include depth limit by recursive call
	lex := &Lexer{
		sourceType: "string",
		source:     "irrelevant",
		logger:     &mocklogger.Logger{},
	}
	parent := sshToken{
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
	_ = lex.loadFromDataSource(parent, nil, maxFileIncludeDepth)
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

func TestLexer_handleIncludeToken(t *testing.T) {
	// tmpDir will be automatically cleaned removed after the test
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config_included")
	content := "Host mock-host\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	lex := &Lexer{
		sourceType: "file",
		source:     filepath.Join(tmpDir, "config"),
		logger:     &mocklogger.Logger{},
	}

	token := sshToken{
		kind:  tokenKind.IncludeFile,
		key:   "Include",
		value: filepath.Join(tmpDir, "*_included"),
	}

	tokens := lex.handleIncludeToken(token)
	require.Len(t, tokens, 1, "expected 1 include token")
}

func TestLexer_MetaDataToken(t *testing.T) {
	lex := &Lexer{}
	line := "# GG:GROUP mock_group"
	token := lex.metaDataToken(tokenKind.Group, line)
	require.Equal(t, token.key, "GROUP", "wrong token key")
	require.Equal(t, token.value, "mock_group", "wrong token value")
}
