package storage

import (
	"errors"
	"os"
	"path"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/logger"
	"github.com/grafviktor/goto/internal/model"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

var _ HostStorage = &yamlStorage{}

const hostsFile = "hosts.yaml"
const ID_EMPTY = 0

func NewYAML(config config.Application) (*yamlStorage, error) {
	appConfigDir, err := getAppConfigDir(config)
	if err != nil {
		return nil, err
	}
	ctxLogger, _ := logger.FromContext(config.Context)

	_, err = os.Stat(appConfigDir)

	if os.IsNotExist(err) {
		ctxLogger.Log("App config folder does not exist. Creating %s\n", appConfigDir)
		err = os.MkdirAll(appConfigDir, 0o700)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		ctxLogger.Log("Failed to open or create App config folder %s\n", appConfigDir)
		return nil, err
	}

	fsDataPath := path.Join(appConfigDir, hostsFile)
	ctxLogger.Log("Reading hosts file list %s\n", fsDataPath)

	return &yamlStorage{
		innerStorage: make(map[int]yamlHostWrapper),
		fsDataPath:   fsDataPath,
		logger:       ctxLogger,
	}, nil
}

type yamlStorage struct {
	innerStorage map[int]yamlHostWrapper
	nextID       int
	fsDataPath   string
	logger       *logger.Logger
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
	if host.ID == ID_EMPTY {
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

	s.nextID = ID_EMPTY
	for _, wrapped := range yamlHosts {
		s.nextID++
		wrapped.Host.ID = s.nextID

		// Maintain an internal map which is keyed by int
		s.innerStorage[s.nextID] = wrapped
	}

	hosts := lo.MapToSlice(s.innerStorage, func(key int, value yamlHostWrapper) model.Host {
		return value.Host
	})

	slices.SortFunc(hosts, func(a, b model.Host) int {
		if a.ID < b.ID {
			return -1
		}
		return 1
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
