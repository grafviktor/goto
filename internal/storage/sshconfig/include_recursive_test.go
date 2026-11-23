package sshconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafviktor/goto/internal/testutils/mocklogger"
	"github.com/stretchr/testify/require"
)

func TestLexer_IncludeRecursiveGlob(t *testing.T) {
	tmpDir := t.TempDir()
	confDir := filepath.Join(tmpDir, "conf.d")
	require.NoError(t, os.MkdirAll(filepath.Join(confDir, "sub1", "deep"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(confDir, "sub2", "deeper", "even"), 0o755))

	parentConfig := filepath.Join(tmpDir, "parent_config.conf")
	require.NoError(t, os.WriteFile(parentConfig,
		[]byte(fmt.Sprintf("Include %s/**/*.conf\n", confDir)), 0o644))

	expected := map[string]bool{
		"depth0": false,
		"depth1": false,
		"depth2": false,
		"depth3": false,
	}
	files := map[string]string{
		"depth0": filepath.Join(confDir, "direct.conf"),
		"depth1": filepath.Join(confDir, "sub1", "a.conf"),
		"depth2": filepath.Join(confDir, "sub1", "deep", "b.conf"),
		"depth3": filepath.Join(confDir, "sub2", "deeper", "even", "c.conf"),
	}
	for name, path := range files {
		require.NoError(t, os.WriteFile(path, []byte("Host "+name+"\n"), 0o644))
	}

	root := configSource{value: parentConfig, valueType: valueTypeFile}
	lex := &Lexer{rootConfig: root, logger: &mocklogger.Logger{}}
	tokens, err := lex.loadFromDataSource(root, nil, 0)
	require.NoError(t, err)

	for _, token := range tokens {
		if _, ok := expected[token.value]; ok {
			expected[token.value] = true
		}
	}
	for name, found := range expected {
		require.True(t, found, "recursive Include should load %s from nested directory", name)
	}
	require.Len(t, tokens, len(files))
}

func TestExpandIncludePatternAbsoluteUnixPath(t *testing.T) {
	matches, err := expandIncludePattern("/home/user/.ssh/conf.d/**/*.conf")
	require.NoError(t, err)
	// No files exist, but the prefix must stay absolute (no CWD-relative match).
	require.Empty(t, matches)
}

func TestExpandIncludePatternAbsolutePrefix(t *testing.T) {
	tmpDir := t.TempDir()
	confDir := filepath.Join(tmpDir, "conf.d")
	require.NoError(t, os.MkdirAll(filepath.Join(confDir, "nested"), 0o755))
	expected := filepath.Join(confDir, "nested", "host.conf")
	require.NoError(t, os.WriteFile(expected, []byte("Host abs-host\n"), 0o644))

	matches, err := expandIncludePattern(filepath.ToSlash(confDir) + "/**/*.conf")
	require.NoError(t, err)
	require.Equal(t, []string{expected}, matches)
}
