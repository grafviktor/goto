package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/application"
	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/storage/sshconfig"
	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

type mockSSHLexer struct{}

func (m *mockSSHLexer) Tokenize() ([]sshconfig.SSHToken, error) {
	return nil, nil
}

func (m *mockSSHLexer) GetRawData() []byte {
	return []byte{}
}

type mockSSHParser struct {
	hosts []model.Host
	err   error
}

func (m *mockSSHParser) Parse() ([]model.Host, error) {
	return m.hosts, m.err
}

func TestNewSSHConfigStorageLocalFile(t *testing.T) {
	mockAppConfig := application.Configuration{}
	mockLogger := mocklogger.Logger{}
	s := newSSHConfigStorage(context.TODO(), &mockAppConfig, &mockLogger)
	require.NotNil(t, s)

	s.Close()
}

func TestSSHConfigFile_GetAll(t *testing.T) {
	mockHosts := []model.Host{
		{Title: "host1", Address: "host1.com"},
		{Title: "host2", Address: "host2.com"},
	}

	mockAppConfig := application.Configuration{}
	s := &SSHConfigFile{
		fileLexer:  &mockSSHLexer{},
		fileParser: &mockSSHParser{hosts: mockHosts},
		appConfig:  &mockAppConfig,
	}
	hosts, err := s.GetAll()
	require.NoError(t, err)
	require.Len(t, hosts, 2)
	require.Equal(t, "host1", hosts[0].Title)
	require.Equal(t, "host2", hosts[1].Title)
	require.Equal(t, 1, hosts[0].ID)
	require.Equal(t, 2, hosts[1].ID)
	// It's required for Windows to release the temp file, we're closing it in storage.Close().
	s.Close()
}

func TestSSHConfigFile_GetAll_Error(t *testing.T) {
	s := &SSHConfigFile{
		fileParser: &mockSSHParser{err: errors.New("parse error")},
	}
	hosts, err := s.GetAll()
	require.Error(t, err)
	require.Nil(t, hosts)
}

func TestSSHConfigFile_Get(t *testing.T) {
	mockHosts := []model.Host{
		{Title: "host1", Address: "host1.com"},
	}
	mockAppConfig := application.Configuration{}
	s := &SSHConfigFile{
		fileLexer:  &mockSSHLexer{},
		fileParser: &mockSSHParser{hosts: mockHosts},
		appConfig:  &mockAppConfig,
	}
	_, _ = s.GetAll()
	h, err := s.Get(1)
	require.NoError(t, err)
	require.Equal(t, "host1", h.Title)
	s.Close() // It's required for Windows to release the temp file
}

func TestSSHConfigFile_Save_Delete(t *testing.T) {
	s := &SSHConfigFile{}
	h, err := s.Save(model.Host{})
	require.ErrorIs(t, err, ErrNotSupported)
	require.Equal(t, model.Host{}, h)

	err = s.Delete(1)
	require.ErrorIs(t, err, ErrNotSupported)
}

func TestSSHConfigFile_Type(t *testing.T) {
	s := &SSHConfigFile{}
	require.Equal(t, constant.HostStorageType.SSHConfig, s.Type())
}
