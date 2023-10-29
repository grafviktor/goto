package model

func NewHost(id int, title, description, address, loginName, privateKeyPath, remotePort string) Host {
	return Host{
		ID:             id,
		Title:          title,
		Description:    description,
		Address:        address,
		LoginName:      loginName,
		RemotePort:     remotePort,
		PrivateKeyPath: privateKeyPath,
	}
}

type Host struct {
	ID             int    `yaml:"-"`
	Title          string `yaml:"title"`
	Description    string `yaml:"description,omitempty"`
	Address        string `yaml:"address"`
	RemotePort     string `yaml:"network_port,omitempty"`
	LoginName      string `yaml:"username,omitempty"`
	PrivateKeyPath string `yaml:"identity_file_path,omitempty"`
}

func (h Host) Clone() Host {
	newHost := Host{
		Title:          h.Title,
		Description:    h.Description,
		Address:        h.Address,
		LoginName:      h.LoginName,
		PrivateKeyPath: h.PrivateKeyPath,
		RemotePort:     h.RemotePort,
	}

	return newHost
}
