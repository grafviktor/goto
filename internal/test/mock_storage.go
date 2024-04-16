package test

import (
	"errors"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/model"
)

// =============================================== Storage

func NewMockStorage(shouldFail bool) *mockStorage {
	hosts := []model.Host{
		// Yaml storage specific: if host has id which is equal to "0"
		// that means that this host doesn't yet exist. It's a hack,
		// but simplifies the application. That's why we cound hosts from "1"
		model.NewHost(1, "Mock Host 1", "", "localhost", "root", "id_rsa", "2222"),
		model.NewHost(2, "Mock Host 2", "", "localhost", "root", "id_rsa", "2222"),
		model.NewHost(3, "Mock Host 3", "", "localhost", "root", "id_rsa", "2222"),
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

	_, id, found := lo.FindIndexOf[model.Host](ms.Hosts, func(h model.Host) bool {
		return h.ID == id
	})

	if !found {
		return errors.New("host not found")
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
