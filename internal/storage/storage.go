// Package storage contains methods for interaction with database.
package storage

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/samber/lo"
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

var _ HostStorage = &CombinedStorage{}

type CombinedStorage struct {
	// hosts    []model.Host
	storages map[constant.HostStorageEnum]HostStorage
	logger   iLogger
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
	combinedStorage := CombinedStorage{storages: make(map[constant.HostStorageEnum]HostStorage)}
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
	storageType := c.GetHostById(hostID).StorageType
	storage := c.storages[storageType]
	return storage.Delete(c.toInnerStorageID(hostID))
}

// Get implements HostStorage.
func (c *CombinedStorage) Get(hostID int) (model.Host, error) {
	storageType := c.GetHostById(hostID).StorageType
	storage := c.storages[storageType]
	host, err := storage.Get(c.toInnerStorageID(hostID))
	if err != nil {
		return model.Host{}, err
	}

	host.StorageType = storage.Type()
	// Re-assign host ID to the external value. Can use fromInnerStorageID(...),
	// but that's not necessary because we already have the value.
	host.ID = hostID
	return host, nil
}

// GetAll implements HostStorage.
func (c *CombinedStorage) GetAll() ([]model.Host, error) {
	// c.hosts = nil
	hosts := make([]model.Host, 0)
	for _, storage := range c.storages {
		storageHosts, err := storage.GetAll()
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(storageHosts); i++ {
			storageHosts[i].ID = c.fromInnerStorageID(storage.Type(), storageHosts[i].ID)
			storageHosts[i].StorageType = storage.Type()
			hosts = append(hosts, storageHosts[i])

			c.logger.Debug("[STORAGE] Storage type: %s -> host id: %d", storage.Type(), storageHosts[i].ID)
		}
	}

	// return c.hosts, nil
	return hosts, nil
}

// Save implements HostStorage.
func (c *CombinedStorage) Save(host model.Host) (model.Host, error) {
	storage := c.storages[host.StorageType]
	host.ID = c.toInnerStorageID(host.ID)
	host, err := storage.Save(host)
	host.ID = c.fromInnerStorageID(host.StorageType, host.ID)
	return host, err
}

// Type implements HostStorage.
func (c *CombinedStorage) Type() constant.HostStorageEnum {
	return constant.HostStorageType.COMBINED
}

const HOST_ID_SHIFT_PER_STORAGE = 10000

func (c *CombinedStorage) fromInnerStorageID(storageEnum constant.HostStorageEnum, hostID int) int {
	// Sort the storage types to ensure consistent ordering
	storageTypes := make([]constant.HostStorageEnum, 0, len(c.storages))
	for k := range maps.Keys(c.storages) {
		storageTypes = append(storageTypes, k)
	}

	slices.SortFunc(storageTypes, func(a, b constant.HostStorageEnum) int {
		return lo.Ternary(string(a) < string(b), -1, 1)
	})

	// Loop through the sorted storage types to find the index of the current storage type
	for n, v := range storageTypes { // iterate over keys in insertion order
		if storageEnum == v {
			return n*HOST_ID_SHIFT_PER_STORAGE + hostID
		}

		n++
	}

	return hostID
}

func (c *CombinedStorage) toInnerStorageID(hostID int) int {
	return hostID % HOST_ID_SHIFT_PER_STORAGE
}

func (c *CombinedStorage) GetHostById(hostID int) model.Host {
	hosts, err := c.GetAll()
	// FIXME: BUG - unsyncs with internal storage when copy and delete host several times
	if err != nil {
		c.logger.Error("Failed to get all hosts: %v", err)
		panic(err)
	}

	host, found := lo.Find(hosts, func(host model.Host) bool {
		return host.ID == hostID
	})

	if !found {
		errMsg := fmt.Sprintf("Host with id %d not found", hostID)
		c.logger.Error(errMsg)
		panic(errMsg)
	}

	return host
}
