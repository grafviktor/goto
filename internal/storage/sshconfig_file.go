package storage

import (
	"context"
	"errors"
	"os"
	"slices"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/application"
	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/storage/sshconfig"
)

var _ HostStorage = &SSHConfigFile{}

// ErrNotSupported - is an error which is returned when trying to save or
// delete host in SSHConfigFile storage.
var ErrNotSupported = errors.New("readonly storage, edit ssh config directly")

type sshLexer interface {
	Tokenize() ([]sshconfig.SSHToken, error)
	GetRawData() []byte
}

type sshParser interface {
	Parse() ([]model.Host, error)
}

// SSHConfigFile - is a storage which contains hosts loaded from SSH config file.
type SSHConfigFile struct {
	fileLexer     sshLexer
	fileParser    sshParser
	innerStorage  map[int]model.Host
	appConfig     *application.Configuration
	sshConfigCopy *os.File
}

// newSSHConfigStorage - constructs new SSHStorage.
func newSSHConfigStorage(
	_ context.Context,
	appConfig *application.Configuration,
	logger iLogger,
) *SSHConfigFile {
	lexer := sshconfig.NewFileLexer(appConfig.SSHConfigFilePath, logger)
	parser := sshconfig.NewParser(lexer, logger)
	return &SSHConfigFile{
		fileLexer:  lexer,
		fileParser: parser,
		appConfig:  appConfig,
	}
}

// GetAll - returns all hosts.
func (s *SSHConfigFile) GetAll() ([]model.Host, error) {
	hosts, err := s.fileParser.Parse()
	if err != nil {
		return nil, err
	}

	err = s.createSSHConfigCopy()
	if err != nil {
		return nil, err
	}

	s.updateApplicationState()
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

func (s *SSHConfigFile) updateApplicationState() {
	s.appConfig.SSHConfigFilePath = s.sshConfigCopy.Name()
}

func (s *SSHConfigFile) createSSHConfigCopy() error {
	rawData := s.fileLexer.GetRawData()
	sshConfigCopy, err := os.CreateTemp("", "goto_sshconfig_*")
	if err != nil {
		return err
	}
	s.sshConfigCopy = sshConfigCopy

	_, err = sshConfigCopy.Write(rawData)
	if err != nil {
		return err
	}

	return nil
}

func (s *SSHConfigFile) Close() {
	s.deleteSSHConfigCopy()
}

func (s *SSHConfigFile) deleteSSHConfigCopy() {
	if s.sshConfigCopy == nil {
		return
	}

	_ = s.sshConfigCopy.Close()
	_ = os.Remove(s.sshConfigCopy.Name())
}
