package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
)

type mockSSHParser struct {
	hosts []model.Host
	err   error
}

func (m *mockSSHParser) Parse() ([]model.Host, error) {
	return m.hosts, m.err
}

func TestSSHConfigFile_GetAll(t *testing.T) {
	mockHosts := []model.Host{
		{Title: "host1", Address: "host1.com"},
		{Title: "host2", Address: "host2.com"},
	}
	s := &SSHConfigFile{
		fileParser: &mockSSHParser{hosts: mockHosts},
	}
	hosts, err := s.GetAll()
	require.NoError(t, err)
	require.Len(t, hosts, 2)
	require.Equal(t, "host1", hosts[0].Title)
	require.Equal(t, "host2", hosts[1].Title)
	require.Equal(t, 1, hosts[0].ID)
	require.Equal(t, 2, hosts[1].ID)
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
	s := &SSHConfigFile{
		fileParser: &mockSSHParser{hosts: mockHosts},
	}
	_, _ = s.GetAll()
	h, err := s.Get(1)
	require.NoError(t, err)
	require.Equal(t, "host1", h.Title)
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
