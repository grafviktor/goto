// Package storage contains methods for interaction with database.
package storage

import (
	"context"
	"fmt"
	"maps"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
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
	hosts    []model.Host
	storages map[constant.HostStorageEnum]HostStorage
}

// Get returns new data service.
func Get(ctx context.Context, appConfig config.Application) (HostStorage, error) {
	yamlStorage, err := newYAMLStorage(ctx, appConfig.Config.AppHome, appConfig.Logger)
	if err != nil {
		return nil, err
	}

	// TODO: This storage type should be enabled by config flag
	sshConfigStorage, err := newSSHConfigStorage(ctx, appConfig.Config.SSHConfigFile, appConfig.Logger)
	if err != nil {
		return nil, err
	}

	return newCombinedStorage(yamlStorage, sshConfigStorage), nil
}

func newCombinedStorage(storages ...HostStorage) HostStorage {
	combinedStorage := CombinedStorage{storages: make(map[constant.HostStorageEnum]HostStorage)}
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
	storage := c.storages[c.hosts[hostID].StorageType]
	return storage.Delete(c.toInnerStorageID(hostID))
}

// Get implements HostStorage.
func (c *CombinedStorage) Get(hostID int) (model.Host, error) {
	storage := c.storages[c.hosts[hostID].StorageType]
	return storage.Get(c.toInnerStorageID(hostID))
}

// GetAll implements HostStorage.
func (c *CombinedStorage) GetAll() ([]model.Host, error) {
	c.hosts = nil
	for _, storage := range c.storages {
		hosts, err := storage.GetAll()
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(hosts); i++ {
			hosts[i].ID = c.fromInnerStorageID(storage.Type(), hosts[i].ID)
			c.hosts = append(c.hosts, hosts[i])
		}
	}

	return c.hosts, nil
}

// Save implements HostStorage.
func (c *CombinedStorage) Save(host model.Host) (model.Host, error) {
	storage := c.storages[host.StorageType]
	host.ID = c.toInnerStorageID(host.ID)
	return storage.Save(host)
}

// Type implements HostStorage.
func (c *CombinedStorage) Type() constant.HostStorageEnum {
	return constant.HostStorageType.COMBINED
}

const HOST_ID_SHIFT_PER_STORAGE = 10000

func (c *CombinedStorage) fromInnerStorageID(storageEnum constant.HostStorageEnum, hostID int) int {
	n := 0
	for st := range maps.Keys(c.storages) { // iterate over keys in insertion order
		if st == storageEnum {
			return n*HOST_ID_SHIFT_PER_STORAGE + hostID
		}

		n++
	}

	return hostID
}

func (c *CombinedStorage) toInnerStorageID(hostID int) int {
	return hostID % HOST_ID_SHIFT_PER_STORAGE
}
