package hostedit

import (
	model "github.com/grafviktor/goto/internal/model/host"
)

func wrap(host *model.Host) hostModelWrapper {
	return hostModelWrapper{Host: host}
}

type hostModelWrapper struct{ *model.Host }

func (m *hostModelWrapper) getHostAttributeValueByIndex(inputType int) string {
	switch inputType {
	case inputTitle:
		return m.Title
	case inputAddress:
		return m.Address
	case inputGroup:
		return m.Group
	case inputDescription:
		return m.Description
	case inputLogin:
		return m.LoginName
	case inputNetworkPort:
		return m.RemotePort
	case inputIdentityFile:
		return m.IdentityFilePath
	default:
		return ""
	}
}

func (m *hostModelWrapper) setHostAttributeByIndex(inputType int, value string) {
	switch inputType {
	case inputTitle:
		m.Title = value
	case inputAddress:
		m.Address = value
	case inputGroup:
		m.Group = value
	case inputDescription:
		m.Description = value
	case inputLogin:
		m.LoginName = value
	case inputNetworkPort:
		m.RemotePort = value
	case inputIdentityFile:
		m.IdentityFilePath = value
	}
}

func (m *hostModelWrapper) unwrap() model.Host {
	return *m.Host
}
