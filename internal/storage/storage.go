// Package storage contains methods for interaction with database.
package storage

import (
	"context"
	"fmt"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/samber/lo"
)

var _ HostStorage = &CombinedStorage{}
var defaultHostStorageType = constant.HostStorageType.YAML_FILE

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
func Get(ctx context.Context, appConfig config.Application, logger iLogger) (HostStorage, error) {
	// TODO: Merge Get and newCombinedStorage into one function.
	yamlStorage, err := newYAMLStorage(ctx, appConfig.Config.AppHome, appConfig.Logger)
	if err != nil {
		return nil, err
	}

	// TODO: This storage type should be enabled by config flag
	sshConfigStorage, err := newSSHConfigStorage(ctx, appConfig.Config.SSHConfigFile, appConfig.Logger)
	if err != nil {
		return nil, err
	}

	return newCombinedStorage(logger, yamlStorage, sshConfigStorage), nil
}

func newCombinedStorage(logger iLogger, storages ...HostStorage) HostStorage {
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

	return &combinedStorage
}

// Delete implements HostStorage.
func (c *CombinedStorage) Delete(hostID int) error {
	// storageType := c.hosts[hostID].StorageType
	// storage := c.storages[storageType]
	storage := c.getHostOrDefaultStorage(c.hosts[hostID])
	if storage == nil {
		return fmt.Errorf("storage type %q not found", storage.Type())
	}

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
	c.hosts = make(map[int]model.Host, 0)
	for _, storage := range c.storages {
		storageHosts, err := storage.GetAll()
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(storageHosts); i++ {
			// storageHosts[i].ID = c.fromInnerStorageID(storage.Type(), storageHosts[i].ID)
			// storageHosts[i].StorageType = storage.Type()
			// c.hosts[storageHosts[i].ID] = storageHosts[i]
			// c.hosts[c.nextID] = HostWrapper{storageHosts[i]}
			c.addHost(storageHosts[i], storage.Type())
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

// const HOST_ID_SHIFT_PER_STORAGE = 10000

// func (c *CombinedStorage) fromInnerStorageID(storageEnum constant.HostStorageEnum, hostID int) int {
// 	storageTypes := lo.Keys(c.storages)

// 	// Sort the storage types to ensure consistent ordering
// 	slices.SortFunc(storageTypes, func(a, b constant.HostStorageEnum) int {
// 		return lo.Ternary(string(a) < string(b), -1, 1)
// 	})

// 	// Loop through the sorted storage types to find the index of the current storage type
// 	for n, v := range storageTypes { // iterate over keys in insertion order
// 		if storageEnum == v {
// 			return n*HOST_ID_SHIFT_PER_STORAGE + hostID
// 		}

// 		n++
// 	}

// 	return hostID
// }

// func (c *CombinedStorage) toInnerStorageID(hostID int) int {
// 	return hostID % HOST_ID_SHIFT_PER_STORAGE
// }

func (c *CombinedStorage) getHostOrDefaultStorage(host model.Host) HostStorage {
	storageType := c.hostStorageMap[host.ID].storageType
	if storageType != "" {
		return c.storages[storageType]
	}

	// Falling back to yaml file if storage is not set
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
	c.logger.Debug("[STORAGE] Storage type: %s -> host id: %d", host.StorageType, c.nextID)

	// BUG: Overrides host ID for new hosts
	host.ID = c.nextID
	c.hosts[c.nextID] = host

	return c.nextID
}
