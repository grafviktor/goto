package host

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewHost(t *testing.T) {
	expectedHost := Host{
		ID:               1,
		Title:            "TestTitle",
		Description:      "TestDescription",
		Address:          "TestAddress",
		RemotePort:       "1234",
		LoginName:        "TestUser",
		IdentityFilePath: "/path/to/private/key",
	}

	// Create a new host using the NewHost function
	newHost := NewHost(expectedHost.ID, expectedHost.Title, expectedHost.Description, expectedHost.Address, expectedHost.LoginName, expectedHost.IdentityFilePath, expectedHost.RemotePort)

	// Check if the new host matches the expected host
	if !reflect.DeepEqual(newHost, expectedHost) {
		t.Errorf("NewHost function did not create the expected host. Expected: %v, Got: %v", expectedHost, newHost)
	}
}

func TestCloneHost(t *testing.T) {
	// Create a host to clone
	originalHost := Host{
		ID:               1,
		Title:            "TestTitle",
		Description:      "TestDescription",
		Address:          "TestAddress",
		RemotePort:       "1234",
		LoginName:        "TestUser",
		IdentityFilePath: "/path/to/private/key",
	}

	// Clone the host
	clonedHost := originalHost.Clone()

	// ID of the new host should always be "0", we should not copy the ID of the original host
	require.Equal(t,
		clonedHost.ID,
		0,
		"Clone function should create a new host, but host ID should be equal to '0'",
	)

	// Set the ID of the cloned host to the original host's ID just for the sake of using DeepEqual.
	// In reality IDs should always be different.
	clonedHost.ID = originalHost.ID
	// Check if the cloned host is equal to the original host
	if !reflect.DeepEqual(clonedHost, originalHost) {
		t.Errorf("Clone function did not create an equal host. Original: %v, Clone: %v", originalHost, clonedHost)
	}

	// Ensure that modifying the cloned host does not affect the original host
	clonedHost.Address = "ModifiedAddress"
	if clonedHost.Address == originalHost.Address {
		t.Error("Modifying the cloned host should not affect the original host")
	}
}

func TestCmdSSHConnect(t *testing.T) {
	tests := []struct {
		name     string
		host     Host
		expected bool
	}{
		{
			name: "NOT user defined ssh command",
			host: Host{
				Address: "localhost",
			},
			expected: false,
		},
		{
			name: "User defined ssh command - contains symbol: '@'",
			host: Host{
				Address: "user@localhost",
			},
			expected: true,
		},
		{
			name: "User defined ssh command - contains symbol: ' '",
			host: Host{
				Address: "localhost -p 2222",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.host.IsUserDefinedSSHCommand())
		})
	}
}

func TestIsUserDefinedSSHCommand(t *testing.T) {
	tests := []struct {
		name     string
		host     Host
		expected string
	}{
		{
			name: "NOT user defined ssh command",
			host: Host{
				Address:          "localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "/tmp",
			},
			expected: "ssh -i /tmp -p 2222 -l root localhost",
		},
		{
			name: "User defined ssh command",
			host: Host{
				Address: "username@localhost",
			},
			expected: "ssh username@localhost",
		},
		{
			name: "User defined ssh command - other parameters ignored",
			host: Host{
				Address:          "username@localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "/tmp",
			},
			expected: "ssh username@localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.host.CmdSSHConnect()
			require.Equal(t, tt.expected, actual)
		})
	}
}
