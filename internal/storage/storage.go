// Package storage contains methods for interaction with database.
package storage

import (
	"context"
	"fmt"

	"github.com/grafviktor/goto/internal/config"
	model "github.com/grafviktor/goto/internal/model/host"
)

type StorageEnum string

var storageType = struct {
	COMBINED   StorageEnum
	SSH_CONFIG StorageEnum
	YAML_FILE  StorageEnum
}{
	COMBINED:   "COMBINED",
	SSH_CONFIG: "SSH_CONFIG",
	YAML_FILE:  "YAML_FILE",
}

// HostStorage defines CRUD operations for Host model.
type HostStorage interface {
	GetAll() ([]model.Host, error)
	Get(hostID int) (model.Host, error)
	Save(model.Host) (model.Host, error)
	Delete(id int) error
	Type() StorageEnum
}

var _ HostStorage = &CombinedStorage{}

const HOST_ID_SALT = 1000

type CombinedStorage struct {
	hosts    []model.Host
	storages map[StorageEnum]HostStorage
}

func NewCombinedStorage(storages ...HostStorage) HostStorage {
	combinedStorage := CombinedStorage{storages: make(map[StorageEnum]HostStorage)}
	for idx, storage := range storages {
		if storage.Type() == storageType.COMBINED {
			errMsg := fmt.Sprintf("cannot use %s in combineStorages method", storage.Type())
			panic(errMsg)
		}
		combinedStorage.storages[storage.Type()] = storage

		hosts, err := storage.GetAll()
		if err != nil {
			panic("failed to read from storage type " + storage.Type())
		}

		idShift := idx * HOST_ID_SALT
		for i := 0; i < len(hosts); i++ {
			hosts[i].ID = hosts[i].ID * idShift
			combinedStorage.hosts = append(combinedStorage.hosts, hosts[i])
		}
	}

	return &combinedStorage
}

// Delete implements HostStorage.
func (c *CombinedStorage) Delete(id int) error {
	panic("unimplemented")
}

// Get implements HostStorage.
func (c *CombinedStorage) Get(hostID int) (model.Host, error) {
	panic("unimplemented")
}

// GetAll implements HostStorage.
func (c *CombinedStorage) GetAll() ([]model.Host, error) {
	panic("unimplemented")
}

// Save implements HostStorage.
func (c *CombinedStorage) Save(model.Host) (model.Host, error) {
	panic("unimplemented")
}

// Type implements HostStorage.
func (c *CombinedStorage) Type() StorageEnum {
	return storageType.COMBINED
}

// Get returns new data service.
func Get(ctx context.Context, appConfig config.Application) (HostStorage, error) {
	yamlStorage, err := newYAMLStorage(ctx, appConfig.Config.AppHome, appConfig.Logger)
	if err != nil {
		return nil, err
	}

	// TODO: This storage type should be enabled by config flag
	sshConfigStorage, err := newSSHConfigStorage(ctx, appConfig.Config.SSHHome, appConfig.Logger)
	if err != nil {
		return nil, err
	}

	return NewCombinedStorage(yamlStorage, sshConfigStorage), nil
}
