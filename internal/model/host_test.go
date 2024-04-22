// Package model contains description of data models. For now there is only 'Host' model
package model

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewHost(t *testing.T) {
	// Create a new host using the NewHost function
	expectedHost := Host{
		ID:               1,
		Title:            "TestTitle",
		Description:      "TestDescription",
		Address:          "TestAddress",
		RemotePort:       "1234",
		LoginName:        "TestUser",
		IdentityFilePath: "/path/to/private/key",
	}

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
