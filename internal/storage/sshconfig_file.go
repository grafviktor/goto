package storage

import (
	"context"
	"errors"
	"os"
	"slices"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	sshConfigSettings "github.com/grafviktor/goto/internal/model/sshconfig"
	"github.com/grafviktor/goto/internal/state"
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
	appState      *state.State
	sshConfigCopy *os.File
}

// newSSHConfigStorage - constructs new SSHStorage.
func newSSHConfigStorage(
	_ context.Context,
	st *state.State,
	logger iLogger,
) *SSHConfigFile {
	lexer := sshconfig.NewFileLexer(st.SSHConfigPath, logger)
	parser := sshconfig.NewParser(lexer, logger)
	return &SSHConfigFile{
		fileLexer:  lexer,
		fileParser: parser,
		appState:   st,
	}
}

// GetAll - returns all hosts.
func (s *SSHConfigFile) GetAll() ([]model.Host, error) {
	// Optimization - this is a readonly storage. Therefore, it's pointless to reload
	// hosts from the file if they are already loaded. That especially increases
	// the performance when the app read hosts from a remote location.
	if s.innerStorage == nil {
		hosts, err := s.fileParser.Parse()
		if err != nil {
			return nil, err
		}

		err = s.createTempSSHConfigCopy()
		if err != nil {
			return nil, err
		}

		s.activateTempSSHConfig()

		s.innerStorage = make(map[int]model.Host, len(hosts))
		for i := range hosts {
			// Make sure that not assigning '0' as host id, because '0' is empty host identifier.
			// Consider to use '-1' for all new hostnames.
			hosts[i].ID = i + 1
			s.innerStorage[i+1] = hosts[i]
		}
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

func (s *SSHConfigFile) createTempSSHConfigCopy() error {
	rawData := s.fileLexer.GetRawData()
	sshConfigCopy, err := os.CreateTemp("", "goto_sshconfig_*")
	if err != nil {
		return err
	}
	s.sshConfigCopy = sshConfigCopy

	_, err = sshConfigCopy.Write(rawData)

	// We need to close this file to make sure that other processes (like ssh) can read it.
	// If not close now, Windows will fail to read the file. We close it anyway, independently
	// if the write call above was successful or not.
	_ = s.sshConfigCopy.Close()

	return err
}

func (s *SSHConfigFile) activateTempSSHConfig() {
	sshConfigSettings.SetPath(s.sshConfigCopy.Name())
}

func (s *SSHConfigFile) Close() {
	s.deleteSSHConfigCopy()
}

func (s *SSHConfigFile) deleteSSHConfigCopy() {
	if s.sshConfigCopy == nil {
		return
	}

	_ = os.Remove(s.sshConfigCopy.Name())
}
