package storage

import (
	"context"
	"errors"
	"os"
	"path"
	"slices"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/storage/sshconfig"
)

var (
	_             HostStorage = &SSHConfigFile{}
	sshConfigFile             = "hosts.config"
)

// ErrNotSupported - is an error which is returned when trying to save or
// delete host in SSHConfigFile storage.
var ErrNotSupported = errors.New("readonly storage, edit ssh config directly")

type sshParser interface {
	Parse() ([]model.Host, error)
}

// SSHConfigFile - is a storage which contains hosts loaded from SSH config file.
type SSHConfigFile struct {
	innerStorage  map[int]model.Host
	parser        sshParser
	sshConfigPath string
	writer        *sshconfig.Writer
	logger        iLogger
}

// newSSHConfigStorage - constructs new SSHStorage.
func newSSHConfigStorage(_ context.Context, appFolder string, sshConfigPath string, logger iLogger) (*SSHConfigFile, error) {
	_, err := os.Stat(sshConfigPath)
	if err != nil {
		return nil, err
	}

	// Read from ~/.ssh/config
	lexer := sshconfig.NewFileLexer(sshConfigPath, logger)
	parser := sshconfig.NewParser(lexer, logger)
	// But write to $GOTO_HOME/hosts.config
	sshWriter := sshconfig.NewWriter(path.Join(appFolder, sshConfigFile), logger)
	return &SSHConfigFile{
		parser:        parser,
		sshConfigPath: sshConfigPath,
		writer:        sshWriter,
		logger:        logger,
	}, nil
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

func (s *SSHConfigFile) SaveAll(hosts []model.Host) error {
	return s.flushToDisk(hosts)
}

func (s *SSHConfigFile) flushToDisk(hosts []model.Host) error {
	// sorting slice by index
	slices.SortFunc(hosts, func(a, b model.Host) int {
		if a.ID < b.ID {
			return -1
		}
		return 1
	})

	return s.writer.WriteToFile(hosts)
}

// Delete - throws not supported error.
func (s *SSHConfigFile) Delete(_ int) error {
	return ErrNotSupported
}

// Type - returns storage type.
func (s *SSHConfigFile) Type() constant.HostStorageEnum {
	return constant.HostStorageType.SSHConfig
}
