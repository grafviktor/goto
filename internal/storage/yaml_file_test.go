package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
)

type testLogger struct{}

func (l *testLogger) Debug(_ string, _ ...any) {}
func (l *testLogger) Info(_ string, _ ...any)  {}
func (l *testLogger) Error(_ string, _ ...any) {}

func TestYAMLFile_SaveAndGetAll(t *testing.T) {
	tmpDir := t.TempDir()
	st := newYAMLStorage(context.TODO(), tmpDir, &testLogger{})

	host1 := model.Host{Title: "host1", Address: "host1.com"}
	host2 := model.Host{Title: "host2", Address: "host2.com"}

	saved1, err := st.Save(host1)
	require.NoError(t, err)
	require.NotZero(t, saved1.ID)

	saved2, err := st.Save(host2)
	require.NoError(t, err)
	require.NotZero(t, saved2.ID)
	require.NotEqual(t, saved1.ID, saved2.ID)

	hosts, err := st.GetAll()
	require.NoError(t, err)
	require.Len(t, hosts, 2)
	titles := []string{hosts[0].Title, hosts[1].Title}
	require.Contains(t, titles, "host1")
	require.Contains(t, titles, "host2")
}

func TestYAMLFile_Get(t *testing.T) {
	tmpDir := t.TempDir()
	st := newYAMLStorage(context.TODO(), tmpDir, &testLogger{})

	host := model.Host{Title: "host1", Address: "host1.com"}
	saved, err := st.Save(host)
	require.NoError(t, err)

	got, err := st.Get(saved.ID)
	require.NoError(t, err)
	require.Equal(t, saved.Title, got.Title)
	require.Equal(t, saved.Address, got.Address)
}

func TestYAMLFile_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	st := newYAMLStorage(context.TODO(), tmpDir, &testLogger{})

	host := model.Host{Title: "host1", Address: "host1.com"}
	saved, err := st.Save(host)
	require.NoError(t, err)

	err = st.Delete(saved.ID)
	require.NoError(t, err)

	_, err = st.Get(saved.ID)
	require.ErrorIs(t, err, constant.ErrNotFound)
}

func TestYAMLFile_GetAll_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	st := newYAMLStorage(context.TODO(), tmpDir, &testLogger{})

	hosts, err := st.GetAll()
	require.NoError(t, err)
	require.Empty(t, hosts)
}

func TestYAMLFile_GetAll_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	st := newYAMLStorage(context.TODO(), tmpDir, &testLogger{})

	// Write invalid YAML to the file
	err := os.WriteFile(filepath.Join(tmpDir, "hosts.yaml"), []byte("not: [valid"), 0o600)
	require.NoError(t, err)

	hosts, err := st.GetAll()
	require.Error(t, err)
	require.Nil(t, hosts)
}

func TestYAMLFile_Type(t *testing.T) {
	st := newYAMLStorage(context.TODO(), t.TempDir(), &testLogger{})
	require.Equal(t, constant.HostStorageType.YAMLFile, st.Type())
}
