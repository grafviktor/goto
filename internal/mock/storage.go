package mock

import (
	"errors"

	"github.com/grafviktor/goto/internal/model"
)

// =============================================== Storage

func NewMockStorage(shouldFail bool) *mockStorage {
	hosts := []model.Host{
		model.NewHost(0, "Mock Host 1", "", "localhost", "root", "id_rsa", "2222"),
		model.NewHost(1, "Mock Host 2", "", "localhost", "root", "id_rsa", "2222"),
		model.NewHost(2, "Mock Host 3", "", "localhost", "root", "id_rsa", "2222"),
	}

	return &mockStorage{
		shouldFail: shouldFail,
		Hosts:      hosts,
	}
}

type mockStorage struct {
	shouldFail bool
	Hosts      []model.Host
}

// Delete implements storage.HostStorage.
func (ms *mockStorage) Delete(id int) error {
	if ms.shouldFail {
		return errors.New("mock error")
	}

	ms.Hosts = append(ms.Hosts[:id], ms.Hosts[id+1:]...)

	return nil
}

// Get implements storage.HostStorage.
func (ms *mockStorage) Get(hostID int) (model.Host, error) {
	if ms.shouldFail {
		return model.Host{}, errors.New("mock error")
	}

	return ms.Hosts[hostID], nil
}

// GetAll implements storage.HostStorage.
func (ms *mockStorage) GetAll() ([]model.Host, error) {
	if ms.shouldFail {
		return ms.Hosts, errors.New("mock error")
	}

	return ms.Hosts, nil
}

// Save implements storage.HostStorage.
func (ms *mockStorage) Save(m model.Host) (model.Host, error) {
	if ms.shouldFail {
		return m, errors.New("mock error")
	}

	ms.Hosts = append(ms.Hosts, m)

	return m, nil
}
