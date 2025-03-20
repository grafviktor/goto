package storage

import (
	"context"
	"errors"
	"fmt"

	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/storage/sshconfig"
)

var _ HostStorage = &SSHConfigFile{}

var ErrNotSupported = errors.New("operation not supported")

type SSHParser interface {
	Parse() ([]model.Host, error)
}

type SSHConfigFile struct {
	sshconfig SSHParser
	hosts     []model.Host
}

// newSSHConfigStorage - constructs new SSHStorage.
func newSSHConfigStorage(_ context.Context, sshHome string, logger iLogger) (*SSHConfigFile, error) {
	sshConfigPath := fmt.Sprintf("%s/ssh_config", sshHome)
	lexer := sshconfig.NewFileLexer(sshConfigPath)
	parser := sshconfig.NewParser(lexer)
	return &SSHConfigFile{sshconfig: parser}, nil
}

// GetAll - returns all hosts.
func (s *SSHConfigFile) GetAll() ([]model.Host, error) {
	var err error
	s.hosts, err = s.sshconfig.Parse()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(s.hosts); i++ {
		s.hosts[i].ID = i
	}

	return s.hosts, nil
}

// Get - returns host by ID.
func (s *SSHConfigFile) Get(hostID int) (model.Host, error) {
	return model.Host{}, nil
}

// Save - throws not supported error.
func (s *SSHConfigFile) Save(host model.Host) (model.Host, error) {
	return host, ErrNotSupported
}

// Delete - throws not supported error.
func (s *SSHConfigFile) Delete(id int) error {
	return ErrNotSupported
}

func (s *SSHConfigFile) Type() StorageEnum {
	return storageType.SSH_CONFIG
}
