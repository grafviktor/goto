package storage

import (
	"context"
	"errors"
	"os"
	"path"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/utils"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

var _ HostStorage = &yamlStorage{}

const hostsFile = "hosts.yaml"
const idEmpty = 0

type Logger interface {
	Debug(format string, args ...any)
}

func NewYAML(ctx context.Context, appName string, logger Logger) (*yamlStorage, error) {
	appFolder, err := utils.GetAppDir(logger, appName)
	if err != nil {
		logger.Debug("Error %s", err.Error())
		return nil, err
	}

	logger.Debug("Config folder %s", appFolder)
	fsDataPath := path.Join(appFolder, hostsFile)

	return &yamlStorage{
		innerStorage: make(map[int]yamlHostWrapper),
		fsDataPath:   fsDataPath,
		logger:       logger,
	}, nil
}

type yamlStorage struct {
	innerStorage map[int]yamlHostWrapper
	nextID       int
	fsDataPath   string
	logger       Logger
}

type yamlHostWrapper struct {
	Host model.Host `yaml:"host"`
}

func (s *yamlStorage) flushToDisk() error {
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

func (s *yamlStorage) Save(host model.Host) error {
	if host.ID == idEmpty {
		s.nextID++
		host.ID = s.nextID
	}

	s.innerStorage[host.ID] = yamlHostWrapper{host}

	return s.flushToDisk()
}

func (s *yamlStorage) Delete(id int) error {
	delete(s.innerStorage, id)

	return s.flushToDisk()
}

func (s *yamlStorage) GetAll() ([]model.Host, error) {
	// re-create innerStorage before reading file data
	s.innerStorage = make(map[int]yamlHostWrapper)
	s.logger.Debug("Read hosts file list %s\n", s.fsDataPath)
	fileData, err := os.ReadFile(s.fsDataPath)
	if err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			return make([]model.Host, 0), nil
		}

		return nil, err
	}

	var yamlHosts []yamlHostWrapper
	err = yaml.Unmarshal(fileData, &yamlHosts)
	if err != nil {
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

	return hosts, nil
}

func (s *yamlStorage) Get(hostID int) (model.Host, error) {
	found, ok := s.innerStorage[hostID]

	if !ok {
		return model.Host{}, constant.ErrNotFound
	}

	return found.Host, nil
}
