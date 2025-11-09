package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFeatureFlag_Set_Valid(t *testing.T) {
	var ff FeatureFlag
	err := ff.Set("ssh_config")
	require.NoError(t, err)
	require.Equal(t, FeatureFlag("ssh_config"), ff)
}

func TestFeatureFlag_Set_Invalid(t *testing.T) {
	var ff FeatureFlag
	err := ff.Set("invalid_feature")
	require.Error(t, err)
	require.Contains(t, err.Error(), "supported values")
	require.Equal(t, FeatureFlag(""), ff)
}

func TestFeatureFlag_String(t *testing.T) {
	ff := FeatureFlag("ssh_config")
	require.Equal(t, "ssh_config", ff.String())
}
