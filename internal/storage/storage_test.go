package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/application"
	"github.com/grafviktor/goto/internal/constant"
	model "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

type fakeHostStorage struct {
	hosts   []model.Host
	hostMap map[int]model.Host
	saveErr error
	getErr  error
	delErr  error
	typ     constant.HostStorageEnum
}

func (f *fakeHostStorage) GetAll() ([]model.Host, error) {
	return f.hosts, nil
}

func (f *fakeHostStorage) Get(id int) (model.Host, error) {
	if f.getErr != nil {
		return model.Host{}, f.getErr
	}
	h, ok := f.hostMap[id]
	if !ok {
		return model.Host{}, errors.New("not found")
	}
	return h, nil
}

func (f *fakeHostStorage) Save(h model.Host) (model.Host, error) {
	if f.saveErr != nil {
		return model.Host{}, f.saveErr
	}
	f.hosts = append(f.hosts, h)
	if f.hostMap == nil {
		f.hostMap = make(map[int]model.Host)
	}
	f.hostMap[h.ID] = h
	return h, nil
}

func (f *fakeHostStorage) Type() constant.HostStorageEnum {
	return f.typ
}

func (f *fakeHostStorage) Delete(id int) error {
	if f.delErr != nil {
		return f.delErr
	}
	delete(f.hostMap, id)
	return nil
}

func TestCombinedStorage_GetAll(t *testing.T) {
	logger := &mocklogger.Logger{}

	cs := combinedStorage{
		storages:       getMockStorages(context.TODO(), application.Configuration{}, logger),
		hostStorageMap: make(map[int]hostStorageMapping),
		hosts:          make(map[int]model.Host),
		nextID:         0,
		logger:         logger,
	}

	hosts, err := cs.GetAll()
	require.NoError(t, err)
	require.Len(t, hosts, 3)
}

func getMockStorages(
	_ context.Context,
	_ application.Configuration,
	_ iLogger,
) map[constant.HostStorageEnum]HostStorage {
	// Setup fake storages
	yamlStorage := &fakeHostStorage{
		hosts: []model.Host{
			{ID: 1, Title: "host1"},
			{ID: 2, Title: "host2"},
		},
		typ: constant.HostStorageType.YAMLFile,
	}
	sshStorage := &fakeHostStorage{
		hosts: []model.Host{
			{ID: 1, Title: "sshhost"},
		},
		typ: constant.HostStorageType.SSHConfig,
	}

	return map[constant.HostStorageEnum]HostStorage{
		constant.HostStorageType.YAMLFile:  yamlStorage,
		constant.HostStorageType.SSHConfig: sshStorage,
	}
}

func TestCombinedStorage_SaveAndGet(t *testing.T) {
	cs := &combinedStorage{
		storages:       make(map[constant.HostStorageEnum]HostStorage),
		hostStorageMap: make(map[int]hostStorageMapping),
		hosts:          make(map[int]model.Host),
		nextID:         0,
	}
	yamlStorage := &fakeHostStorage{typ: constant.HostStorageType.YAMLFile, hostMap: make(map[int]model.Host)}
	cs.storages[constant.HostStorageType.YAMLFile] = yamlStorage

	host := model.Host{Title: "test"}
	saved, err := cs.Save(host)
	require.NoError(t, err, "expected no error on save")
	require.NotEqual(t, 0, saved.ID, "expected ID to be non-zero after save")
	got, err := cs.Get(saved.ID)
	require.NoError(t, err)
	require.Equal(t, "test", got.Title, "expected title 'test'")
}

func TestCombinedStorage_Delete(t *testing.T) {
	cs := &combinedStorage{
		storages:       make(map[constant.HostStorageEnum]HostStorage),
		hostStorageMap: make(map[int]hostStorageMapping),
		hosts:          make(map[int]model.Host),
		nextID:         0,
	}
	yamlStorage := &fakeHostStorage{typ: constant.HostStorageType.YAMLFile, hostMap: make(map[int]model.Host)}
	cs.storages[constant.HostStorageType.YAMLFile] = yamlStorage

	host := model.Host{Title: "test"}
	saved, _ := cs.Save(host)
	err := cs.Delete(saved.ID)
	require.NoError(t, err, "expected no error on delete")
	require.NotContains(t, cs.hosts, saved.ID, "host not deleted from combined storage")
}
