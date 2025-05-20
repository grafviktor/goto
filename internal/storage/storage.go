// Package storage contains methods for interaction with database.
package storage

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"golang.org/x/exp/slices"

	"github.com/grafviktor/goto/internal/application"
	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/state"
)

var (
	_                      HostStorage = &CombinedStorage{}
	defaultHostStorageType             = constant.HostStorageType.YAML_FILE
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

// HostStorage defines CRUD operations for Host model.
type HostStorage interface {
	GetAll() ([]model.Host, error)
	Get(hostID int) (model.Host, error)
	Save(model.Host) (model.Host, error)
	Type() constant.HostStorageEnum
	Delete(id int) error
}

type hostStorageMapping struct {
	storageType       constant.HostStorageEnum
	combinedStorageID int
	innerStorageID    int
}

type CombinedStorage struct {
	hosts          map[int]model.Host
	storages       map[constant.HostStorageEnum]HostStorage
	logger         iLogger
	nextID         int
	hostStorageMap map[int]hostStorageMapping
}

// Get returns new data service.
func Get(ctx context.Context, appConfig application.Configuration, logger iLogger) (HostStorage, error) {
	storages, err := getStorages(ctx, appConfig, logger)
	if err != nil {
		return nil, err
	}

	combinedStorage := CombinedStorage{
		storages:       make(map[constant.HostStorageEnum]HostStorage),
		hostStorageMap: make(map[int]hostStorageMapping),
		hosts:          make(map[int]model.Host),
		nextID:         0,
	}
	combinedStorage.logger = logger
	for _, storage := range storages {
		if storage.Type() == constant.HostStorageType.COMBINED {
			errMsg := fmt.Sprintf("cannot use %s in combineStorages method", storage.Type())
			panic(errMsg)
		}
		combinedStorage.storages[storage.Type()] = storage
	}

	return &combinedStorage, nil
}

func getStorages(ctx context.Context, appConfig application.Configuration, logger iLogger) ([]HostStorage, error) {
	storages := []HostStorage{}
	yamlStorage, err := newYAMLStorage(ctx, appConfig.AppHome, logger)
	if err != nil {
		return nil, err
	}

	storages = append(storages, yamlStorage)

	sshConfigEnabled := state.Get().SSHConfigEnabled
	logger.Debug("[STORAGE] SSH config storage enable: '%t'", sshConfigEnabled)
	if sshConfigEnabled {
		logger.Info("[STORAGE] Load ssh hosts from ssh config file: '%s'", appConfig.SSHConfigFilePath)
		sshConfigStorage, err := newSSHConfigStorage(ctx, appConfig.SSHConfigFilePath, logger)
		if err != nil {
			return nil, err
		}

		storages = append(storages, sshConfigStorage)
	}

	return storages, nil
}

// Delete implements HostStorage.
func (c *CombinedStorage) Delete(hostID int) error {
	storage := c.getHostOrDefaultStorage(c.hosts[hostID])
	delete(c.hosts, hostID)
	return storage.Delete(c.hostStorageMap[hostID].innerStorageID)
}

// Get implements HostStorage.
func (c *CombinedStorage) Get(hostID int) (model.Host, error) {
	storage := c.getHostOrDefaultStorage(c.hosts[hostID])
	storageID := c.hostStorageMap[hostID].innerStorageID
	host, err := storage.Get(storageID)
	if err != nil {
		return model.Host{}, err
	}

	host.StorageType = storage.Type()
	// Re-assign host ID to the external value. Can use fromInnerStorageID(...),
	// but that's not necessary because we already have the value.
	host.ID = hostID
	return host, nil
}

// GetAll implements HostStorage. Warning: this method rebuilds the IDs.
func (c *CombinedStorage) GetAll() ([]model.Host, error) {
	storageTypes := lo.Keys(c.storages)
	slices.Sort(storageTypes)
	c.hosts = make(map[int]model.Host, 0)
	for _, storageType := range storageTypes {
		storageHosts, err := c.storages[storageType].GetAll()
		if err != nil {
			return nil, err
		}

		/*
		 * This sorting is required to preserve ID order between application restarts.
		 * If omit this, then almost every app restart loaded hosts will receive random IDs,
		 * thus it won't be possible to focus on previously selected host in the UI.
		 * In other words appState.Selected will point to different host with almost every restart.
		 *
		 * The reason why hosts come in different order is that the underlying storage contains
		 * all hosts in a map, because ... storage there is a hack which indicates that
		 * all new hosts IDs are equal to 0. Once this hack is removed, the sorting
		 * will be removed as well.
		 */
		slices.SortFunc(storageHosts, func(a, b model.Host) int {
			return lo.Ternary(a.Title < b.Title, -1, 1)
		})

		for i := 0; i < len(storageHosts); i++ {
			storageHosts[i].StorageType = storageType
			c.addHost(storageHosts[i], storageType)
		}
	}

	return lo.Values(c.hosts), nil
}

// Save implements HostStorage.
func (c *CombinedStorage) Save(host model.Host) (model.Host, error) {
	storage := c.getHostOrDefaultStorage(host)
	if isNewHost(host) {
		host, err := storage.Save(host)
		combinedStorageID := c.addHost(host, storage.Type())
		host.ID = combinedStorageID
		return host, err
	} else {
		mapping := c.hostStorageMap[host.ID]
		host.ID = mapping.innerStorageID
		host, err := storage.Save(host)
		host.ID = mapping.combinedStorageID
		return host, err
	}
}

// Type implements HostStorage.
func (c *CombinedStorage) Type() constant.HostStorageEnum {
	return constant.HostStorageType.COMBINED
}

func (c *CombinedStorage) getHostOrDefaultStorage(host model.Host) HostStorage {
	storageType := c.hostStorageMap[host.ID].storageType
	if storageType != "" {
		return c.storages[storageType]
	}

	// Falling back to yaml file if storage is not set. This is a new host.
	return c.storages[defaultHostStorageType]
}

func isNewHost(host model.Host) bool {
	return host.ID == 0
}

func (c *CombinedStorage) addHost(host model.Host, storageType constant.HostStorageEnum) int {
	c.nextID++
	c.hostStorageMap[c.nextID] = hostStorageMapping{
		storageType:       storageType,
		combinedStorageID: c.nextID,
		innerStorageID:    host.ID,
	}

	// BUG? - Overrides host ID for new hosts
	host.ID = c.nextID
	c.hosts[c.nextID] = host

	return c.nextID
}
