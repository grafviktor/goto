package storage

import (
	"context"
	"errors"
	"os"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/storage/sshconfig"
)

var _ HostStorage = &SSHConfigFile{}

var ErrNotSupported = errors.New("readonly storage, edit ssh config directly")

type SSHParser interface {
	Parse() ([]model.Host, error)
}

type SSHConfigFile struct {
	parser       SSHParser
	innerStorage map[int]model.Host
}

// newSSHConfigStorage - constructs new SSHStorage.
func newSSHConfigStorage(_ context.Context, sshConfigPath string, logger iLogger) (*SSHConfigFile, error) {
	_, err := os.Stat(sshConfigPath)
	if err != nil {
		return nil, err
	}

	lexer := sshconfig.NewFileLexer(sshConfigPath, logger)
	parser := sshconfig.NewParser(lexer, logger)
	return &SSHConfigFile{parser: parser}, nil
}

// GetAll - returns all hosts.
func (s *SSHConfigFile) GetAll() ([]model.Host, error) {
	hosts, err := s.parser.Parse()
	if err != nil {
		return nil, err
	}

	s.innerStorage = make(map[int]model.Host, len(hosts))
	for i := 0; i < len(hosts); i++ {
		// Make sure that not assigning '0' as host id, because '0' is empty host identifier.
		// Consider to use '-1' for all new hostnames.
		hosts[i].ID = i + 1
		s.innerStorage[i+1] = hosts[i]
	}

	return lo.Values(s.innerStorage), nil
}

// Get - returns host by ID.
func (s *SSHConfigFile) Get(hostID int) (model.Host, error) {
	return s.innerStorage[hostID], nil
}

// Save - throws not supported error.
func (s *SSHConfigFile) Save(host model.Host) (model.Host, error) {
	return host, ErrNotSupported
}

// Delete - throws not supported error.
func (s *SSHConfigFile) Delete(id int) error {
	return ErrNotSupported
}

func (s *SSHConfigFile) Type() constant.HostStorageEnum {
	return constant.HostStorageType.SSH_CONFIG
}
