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
	rootConfig := configSource{
		value:     config,
		valueType: valueTypeRaw,
	}
	lex := &Lexer{
		rootConfig: rootConfig,
		logger:     &mocklogger.Logger{},
	}
	tokens, _ := lex.loadFromDataSource(rootConfig, nil, 0)

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
	rootConfig := configSource{
		value:     config,
		valueType: valueTypeRaw,
	}
	lex := &Lexer{
		rootConfig: rootConfig,
		logger:     &mocklogger.Logger{},
	}
	tokens, _ := lex.loadFromDataSource(rootConfig, nil, 0)

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
	rootConfig := configSource{
		value:     config,
		valueType: valueTypeRaw,
	}
	lex := &Lexer{
		rootConfig: rootConfig,
		logger:     &mocklogger.Logger{},
	}
	tokens, _ := lex.loadFromDataSource(rootConfig, nil, 0)
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for invalid user, got %d", len(tokens))
	}
}

func TestLexer_Tokenize_InvalidPort(t *testing.T) {
	const config = `
Port notaport
`
	rootConfig := configSource{
		value:     config,
		valueType: valueTypeRaw,
	}
	lex := &Lexer{
		rootConfig: rootConfig,
		logger:     &mocklogger.Logger{},
	}
	tokens, _ := lex.loadFromDataSource(rootConfig, nil, 0)
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

	rootConfig := configSource{
		value:     includedConfig1,
		valueType: valueTypeFile,
	}
	lex := &Lexer{
		rootConfig: rootConfig,
		logger:     &mocklogger.Logger{},
	}
	tokens, _ := lex.loadFromDataSource(rootConfig, nil, 0)
	require.Len(t, tokens, 1, "expected 1 token for included host")
}

func TestLexer_Tokenize_IncludeDepthLimit(t *testing.T) {
	// Simulate include depth limit reached by recursive call
	tmpDir := t.TempDir()
	includedConfig := filepath.Join(tmpDir, "config_included1")

	// Create a config file with a single line - Include pointing to another file.
	content := "Host mock-host\n"
	content += fmt.Sprintf("Include %s\n", includedConfig)
	if err := os.WriteFile(includedConfig, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	rootConfig := configSource{
		value:     includedConfig,
		valueType: valueTypeFile,
	}
	logger := &mocklogger.Logger{}
	lex := &Lexer{
		rootConfig: rootConfig,
		logger:     logger,
	}

	tokens, err := lex.loadFromDataSource(rootConfig, nil, 0)
	require.Len(t, tokens, maxFileIncludeDepth, "expected tokens is not equal to max include depth")
	require.NoError(t, err, "should not error on max include depth")
	require.Contains(t, logger.Logs, "[SSHCONFIG] Max include depth reached", "expected log about max include depth")
}

func Test_matchToken(t *testing.T) {
	tests := []struct {
		str   string
		token string
		want  bool
	}{
		{"Host test", "host", true},
		{"HOST test", "host", true},
		{"\tUser alice", "USER", true},
		{" Port 22", "port", true},
		{"\tIdentityFile foo", "identityfile", true},
		{"\tSomethingElse", "host", false},
		{"\t# GG:GROUP test", "# GG:GROUP", true},
		{"\t# GG:GROUP: test", "# GG:GROUP", true},
	}
	for _, tt := range tests {
		if got := matchToken(tt.str, tt.token); got != tt.want {
			t.Errorf("hasPrefixIgnoreCase(%q, %q) = %v, want %v", tt.str, tt.token, got, tt.want)
		}
	}
}

func TestLexer_handleIncludeToken_localFile(t *testing.T) {
	// Create the included file, this file will be read by handleIncludeToken.
	tmpDir := t.TempDir()

	// Re-define HOME and USERPROFILE to point to our temp directory, that's for tilde path test.
	t.Setenv("USERPROFILE", tmpDir) // For Windows
	t.Setenv("HOME", tmpDir)        // For Unix

	tmpFile := filepath.Join(tmpDir, "config_included")
	// Create a config file with a single line - Host. That's enough for the test.
	content := "Host mock-local-host\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		tokenValue string
		expected   configSource
	}{
		{
			tokenValue: tmpFile,
			expected: configSource{
				value:     tmpFile,
				valueType: valueTypeFile,
			},
		},
		{
			tokenValue: "~/" + "config_included",
			expected: configSource{
				value:     tmpFile,
				valueType: valueTypeFile,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.tokenValue, func(t *testing.T) {
			rootConfig := configSource{
				value:     "it won't be read",
				valueType: valueTypeFile,
			}

			lex := &Lexer{
				rootConfig: rootConfig,
				logger:     &mocklogger.Logger{},
			}

			// That's our starting token which includes the file we created above.
			token := SSHToken{
				kind:  tokenKind.IncludeFile,
				value: tt.tokenValue,
			}

			configSources := lex.handleIncludeToken(token, rootConfig)
			require.Len(t, configSources, 1, "expected 1 include token")
			require.Equal(t, tt.expected.value, configSources[0].value, "unexpected included file path")
			require.Equal(t, tt.expected.valueType, configSources[0].valueType,
				"unexpected value type for included file")
		})
	}
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
			rootConfig := configSource{
				value:     tt.baseURL,
				valueType: valueTypeURL,
			}

			lex := &Lexer{
				rootConfig: rootConfig,
				logger:     &mocklogger.Logger{},
			}

			startingPointToken := SSHToken{
				kind:  tokenKind.IncludeFile,
				value: tt.sourceTokenURL,
			}

			configSource := lex.handleIncludeToken(startingPointToken, rootConfig)
			require.Len(t, configSource, 1, "expected 1 include token")
			require.Equal(t, tt.expectedTokenURL, configSource[0].value, "unexpected included file URL")
		})
	}
}

func TestLexer_MetaDataToken(t *testing.T) {
	lex := &Lexer{}
	lines := []string{"# GG:GROUP mock_group", "# GG:GROUP: mock_group"}
	for _, line := range lines {
		token := lex.metaDataToken(tokenKind.Group, line)
		require.Equal(t, tokenKind.Group, token.kind, "wrong token kind")
		require.Equal(t, "mock_group", token.value, "wrong token value")
	}
}

func TestExpandTildePath(t *testing.T) {
	// I don't want to hardcode c:\users\test or /home/test, as I will have to split the test
	// for Windows and Unix. Using temp directory for simplicity.
	tmpDir := t.TempDir()
	t.Setenv("USERPROFILE", tmpDir) // For Windows
	t.Setenv("HOME", tmpDir)        // For Unix

	lex := &Lexer{
		rootConfig: configSource{},
		logger:     &mocklogger.Logger{},
	}

	expanded := lex.expandTildePath("~/config")
	expected := filepath.Join(tmpDir, "config")
	require.Equal(t, expected, expanded, "expanded path does not match expected")
}

func TestExpandTildePath_HomeIsBlank(t *testing.T) {
	tests := []struct {
		name               string
		userHomeEnv        string
		expectedLogMessage string
	}{
		{
			name:               "Test 1",
			userHomeEnv:        "",
			expectedLogMessage: "[SSHCONFIG]: Cannot expand tilde in path: $HOME is not defined",
		},
		{
			name:               "Test 2",
			userHomeEnv:        "   ",
			expectedLogMessage: "[SSHCONFIG]: Cannot find user home directory to expand tilde in path: ~/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("USERPROFILE", tt.userHomeEnv) // For Windows
			t.Setenv("HOME", tt.userHomeEnv)        // For Unix

			logger := &mocklogger.Logger{}
			lex := &Lexer{
				rootConfig: configSource{},
				logger:     logger,
			}

			expanded := lex.expandTildePath("~/config")
			require.Equal(t, "~/config", expanded, "expanded path does not match expected")
			require.Contains(t, logger.Logs[0], tt.expectedLogMessage, "expected log about missing HOME variable")
		})
	}
}
