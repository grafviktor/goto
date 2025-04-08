package storage

import (
	"context"
	"errors"
	"os"
	"path"

	"golang.org/x/exp/slices"

	"github.com/samber/lo"
	"gopkg.in/yaml.v2"

	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
)

var _ HostStorage = &yamlFile{}

const (
	hostsFile = "hosts.yaml"
	// Yaml storage specific: if host has id which is equal to "0"
	// that means that this host doesn't yet exist. It's a hack,
	// but simplifies the application. That's why idEmpty = "0".
	idEmpty = 0
)

// newYAMLStorage creates new YAML storage.
func newYAMLStorage(_ context.Context, appFolder string, logger iLogger) (*yamlFile, error) {
	logger.Debug("[STORAGE] Init YAML storage. Config folder %s", appFolder)
	fsDataPath := path.Join(appFolder, hostsFile)

	return &yamlFile{
		innerStorage: make(map[int]yamlHostWrapper),
		fsDataPath:   fsDataPath,
		logger:       logger,
	}, nil
}

type yamlFile struct {
	innerStorage map[int]yamlHostWrapper
	nextID       int
	fsDataPath   string
	logger       iLogger
}

type yamlHostWrapper struct {
	Host model.Host `yaml:"host"`
}

func (s *yamlFile) flushToDisk() error {
	// map contains values in shuffled order
	mapValues := lo.Values(s.innerStorage)
	// sorting slice by index
	slices.SortFunc(mapValues, func(a, b yamlHostWrapper) int {
		if a.Host.ID < b.Host.ID {
			return -1
		}
		return 1
	})

	result, err := yaml.Marshal(mapValues)
	if err != nil {
		return err
	}

	err = os.WriteFile(s.fsDataPath, result, 0o600)
	if err != nil {
		panic(err)
	}

	return nil
}

func (s *yamlFile) Save(host model.Host) (model.Host, error) {
	if host.ID == idEmpty {
		s.logger.Debug("[STORAGE] Generate new id for new host with title: %s", host.Title)
		s.nextID++
		host.ID = s.nextID
	}

	s.logger.Info("[STORAGE] Save host with id: %d, title: %s", host.ID, host.Title)
	s.innerStorage[host.ID] = yamlHostWrapper{host}

	err := s.flushToDisk()
	if err != nil {
		s.logger.Error("[STORAGE] Cannot flush database changes to disk. %v", err)
	}

	return host, err
}

func (s *yamlFile) Delete(hostID int) error {
	s.logger.Info("[STORAGE] Delete host with id: %d", hostID)
	delete(s.innerStorage, hostID)

	err := s.flushToDisk()
	if err != nil {
		s.logger.Error("[STORAGE] Error deleting host id: %d from the database. %v", hostID, err)
	}
	return err
}

func (s *yamlFile) GetAll() ([]model.Host, error) {
	// re-create innerStorage before reading file data
	s.innerStorage = make(map[int]yamlHostWrapper)
	s.logger.Debug("[STORAGE] Read hosts from file: %s\n", s.fsDataPath)
	fileData, err := os.ReadFile(s.fsDataPath)
	if err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			s.logger.Info("[STORAGE] Path no found: %s. Assuming it's not created yet", s.fsDataPath)

			return make([]model.Host, 0), nil
		}

		s.logger.Error("[STORAGE] Read hosts error. %v", err)
		return nil, err
	}

	var yamlHosts []yamlHostWrapper
	s.logger.Debug("[STORAGE] Unmarshal hosts data from yaml storage")
	err = yaml.Unmarshal(fileData, &yamlHosts)
	if err != nil {
		s.logger.Error("[STORAGE] Could not unmarshal hosts data. %v", err)
		return nil, err
	}

	s.nextID = idEmpty
	for _, wrapped := range yamlHosts {
		s.nextID++
		wrapped.Host.ID = s.nextID

		// Maintain an internal map which is keyed by int
		s.innerStorage[s.nextID] = wrapped
	}

	hosts := lo.MapToSlice(s.innerStorage, func(key int, value yamlHostWrapper) model.Host {
		return value.Host
	})

	s.logger.Debug("[STORAGE] Read %d items from the database", len(hosts))
	return hosts, nil
}

func (s *yamlFile) Get(hostID int) (model.Host, error) {
	s.logger.Debug("[STORAGE] Read host with id %d from the database", hostID)
	found, ok := s.innerStorage[hostID]

	if !ok {
		s.logger.Debug("[STORAGE] Host id %d NOT found in the database", hostID)
		return model.Host{}, constant.ErrNotFound
	}

	s.logger.Debug("[STORAGE] Host id %d found in the database", hostID)
	return found.Host, nil
}

func (s *yamlFile) Type() constant.HostStorageEnum {
	return constant.HostStorageType.YAML_FILE
}
