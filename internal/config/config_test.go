package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
)

func Test_Initialize(t *testing.T) {
	appConfig, err := Initialize()
	if err != nil {
		t.Fatalf("failed to initialize configuration: %v", err)
	}

	require.NoError(t, err)
	require.Equal(t, "goto", appConfig.AppName)
	require.NotEmpty(t, appConfig.AppHome)
	require.Equal(t, "info", appConfig.LogLevel)
	require.Equal(t, constant.AppModeType.StartUI, appConfig.AppMode)
}

func Test_parseEnvironmentConfig(t *testing.T) {
	t.Run("No environment variables set", func(t *testing.T) {
		t.Setenv("GG_HOME", "")
		t.Setenv("GG_LOG_LEVEL", "")
		t.Setenv("GG_SSH_CONFIG_FILE_PATH", "")

		envConfig, err := parseEnvironmentVariables()
		require.NoError(t, err)
		require.Empty(t, envConfig.AppHome)
		require.Empty(t, envConfig.AppMode)
		require.Empty(t, envConfig.DisableFeature)
		require.Empty(t, envConfig.EnableFeature)
		require.Empty(t, envConfig.SetTheme)
		require.Empty(t, envConfig.SSHConfigFilePath)
		require.Equal(t, "info", envConfig.LogLevel)
	})

	t.Run("Environment variables set", func(t *testing.T) {
		t.Setenv("GG_HOME", "/root")
		t.Setenv("GG_LOG_LEVEL", "debug")
		t.Setenv("GG_SSH_CONFIG_FILE_PATH", "/tmp/custom_config")

		envConfig, err := parseEnvironmentVariables()
		require.NoError(t, err)
		require.Equal(t, "/root", envConfig.AppHome)
		require.Empty(t, envConfig.AppMode)
		require.Empty(t, envConfig.DisableFeature)
		require.Empty(t, envConfig.EnableFeature)
		require.Empty(t, envConfig.SetTheme)
		require.Equal(t, "/tmp/custom_config", envConfig.SSHConfigFilePath)
		require.Equal(t, "debug", envConfig.LogLevel)
	})
}

func Test_parseCommandLineFlags(t *testing.T) {
	envConfig := &Configuration{
		AppHome:           "/tmp/home",
		LogLevel:          "info",
		SSHConfigFilePath: "/tmp/custom_config",
	}

	args := []string{
		"-v",
		"-f", "/tmp/home2",
		"-l", "debug",
		"-s", "/tmp/custom_config2",
		"--set-theme", "dark",
		"-e", "ssh_config",
	}

	cfg, err := parseCommandLineFlags(envConfig, args, false)
	require.NotNil(t, cfg)
	require.NoError(t, err)
}
