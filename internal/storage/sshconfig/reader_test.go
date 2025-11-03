package sshconfig

import (
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newReader_url(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := newReader("bad_url", "url")
	require.Error(t, err)

	_, err = newReader(ts.URL, "url")
	require.NoError(t, err)
}

func Test_newReader_file(t *testing.T) {
	// Test case: File does not exist
	noSuchFile := path.Join(t.TempDir(), "no_such_file")
	_, err := newReader(noSuchFile, "file")
	require.Error(t, err)

	// Test case: Path is a directory
	folderInsteadOfFile := t.TempDir()
	_, err = newReader(folderInsteadOfFile, "file")
	require.Error(t, err)
}
