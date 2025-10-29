//nolint:mnd // Magic numbers are allowed in unittest
package testutils_test

import (
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/model/sshconfig"
)

type MockStorage struct {
	shouldFail bool
	Hosts      []host.Host
}

func NewMockStorage(shouldFail bool) *MockStorage {
	hosts := []host.Host{
		// Yaml storage specific: if host has id which is equal to "0"
		// that means that this host doesn't yet exist. It's a hack,
		// but simplifies the application. That's why we count hosts from "1"
		host.NewHost(1, "Mock Host 1", "", "localhost", "root", "id_rsa", "2222"),
		host.NewHost(2, "Mock Host 2", "", "localhost", "root", "id_rsa", "2222"),
		host.NewHost(3, "Mock Host 3", "", "localhost", "root", "id_rsa", "2222"),
	}

	for i := range hosts {
		hosts[i].SSHHostConfig = &sshconfig.Config{}
		hosts[i].Group = fmt.Sprintf("Group %d", i+1)
	}

	return &MockStorage{
		shouldFail: shouldFail,
		Hosts:      hosts,
	}
}

// Delete implements storage.HostStorage.
func (ms *MockStorage) Delete(id int) error {
	if ms.shouldFail {
		return errors.New("mock error")
	}

	_, id, found := lo.FindIndexOf(ms.Hosts, func(h host.Host) bool {
		return h.ID == id
	})

	if !found {
		return errors.New("host not found")
	}

	ms.Hosts = append(ms.Hosts[:id], ms.Hosts[id+1:]...)

	return nil
}

// Get implements storage.HostStorage.
func (ms *MockStorage) Get(hostID int) (host.Host, error) {
	if ms.shouldFail {
		return host.Host{}, errors.New("mock error")
	}

	return ms.Hosts[hostID], nil
}

// GetAll implements storage.HostStorage.
func (ms *MockStorage) GetAll() ([]host.Host, error) {
	if ms.shouldFail {
		return ms.Hosts, errors.New("mock error")
	}

	return ms.Hosts, nil
}

// Save implements storage.HostStorage.
func (ms *MockStorage) Save(m host.Host) (host.Host, error) {
	if ms.shouldFail {
		return m, errors.New("mock error")
	}

	ms.Hosts = append(ms.Hosts, m)

	return m, nil
}

func (ms *MockStorage) Type() constant.HostStorageEnum {
	return "MOCK STORAGE"
}

func (ms *MockStorage) Close() {
	// nop
}
