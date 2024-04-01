package ssh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetCurrentOSUser(t *testing.T) {
	username := currentUsername()
	require.NotEmpty(t, username, "GetCurrentOSUser should return a non-empty string")
}
