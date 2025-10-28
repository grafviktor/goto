package storage

import (
	"context"
	"errors"
	"slices"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/storage/sshconfig"
	"github.com/grafviktor/goto/internal/utils"
)

var _ HostStorage = &SSHConfigFile{}

// ErrNotSupported - is an error which is returned when trying to save or
// delete host in SSHConfigFile storage.
var ErrNotSupported = errors.New("readonly storage, edit ssh config directly")

type sshParser interface {
	Parse() ([]model.Host, error)
}

// SSHConfigFile - is a storage which contains hosts loaded from SSH config file.
type SSHConfigFile struct {
	parser       sshParser
	innerStorage map[int]model.Host
}

// newSSHConfigStorage - constructs new SSHStorage.
func newSSHConfigStorage(_ context.Context, sshConfigPath string, logger iLogger) *SSHConfigFile {
	var sourceType string
	if utils.IsNetworkSchemeSupported(sshConfigPath) {
		sourceType = "url"
	} else {
		sourceType = "file"
	}

	lexer := sshconfig.NewFileLexer(sshConfigPath, sourceType, logger)
	parser := sshconfig.NewParser(lexer, logger)
	return &SSHConfigFile{parser: parser}
}

// GetAll - returns all hosts.
func (s *SSHConfigFile) GetAll() ([]model.Host, error) {
	hosts, err := s.parser.Parse()
	if err != nil {
		return nil, err
	}

	s.innerStorage = make(map[int]model.Host, len(hosts))
	for i := range hosts {
		// Make sure that not assigning '0' as host id, because '0' is empty host identifier.
		// Consider to use '-1' for all new hostnames.
		hosts[i].ID = i + 1
		s.innerStorage[i+1] = hosts[i]
	}

	values := lo.Values(s.innerStorage)
	// Map does not guarantee order, so we need to sort the collection.
	slices.SortFunc(values, func(a, b model.Host) int {
		if a.ID < b.ID {
			return -1
		}
		return 1
	})
	return values, nil
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
func (s *SSHConfigFile) Delete(_ int) error {
	return ErrNotSupported
}

// Type - returns storage type.
func (s *SSHConfigFile) Type() constant.HostStorageEnum {
	return constant.HostStorageType.SSHConfig
}
