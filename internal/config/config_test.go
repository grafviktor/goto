package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
)

func Test_Initialize(t *testing.T) {
	os.Args = []string{"app_name"} //nolint:reassign // reset os.Args to avoid interference from test runner args
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
		require.Empty(t, envConfig.SSHConfigPath)
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
		require.Equal(t, "/tmp/custom_config", envConfig.SSHConfigPath)
		require.Equal(t, "debug", envConfig.LogLevel)
	})
}

func Test_parseCommandLineFlags(t *testing.T) {
	envConfig := &Configuration{
		AppHome:       "/tmp/home",
		LogLevel:      "info",
		SSHConfigPath: "/tmp/custom_config",
	}

	tests := []struct {
		name       string
		args       []string
		wantConfig *Configuration
		wantError  bool
	}{
		{
			name: "No command line flags",
			args: []string{},
			wantConfig: &Configuration{
				AppHome:        "/tmp/home",
				AppMode:        "",
				DisableFeature: "",
				EnableFeature:  "",
				LogLevel:       "info",
				SSHConfigPath:  "/tmp/custom_config",
			},
			wantError: false,
		}, {
			name: "Display help",
			args: []string{"-h"},
			wantConfig: &Configuration{
				AppHome:        "/tmp/home",
				AppMode:        "",
				DisableFeature: "",
				EnableFeature:  "",
				LogLevel:       "info",
				SSHConfigPath:  "/tmp/custom_config",
				SetTheme:       "",
			},
			wantError: true, // returns flag.ErrHelp, read the os.flag documentation for details
		}, {
			name: "Display version",
			args: []string{"-v"},
			wantConfig: &Configuration{
				AppHome:        "/tmp/home",
				AppMode:        "DISPLAY_INFO",
				DisableFeature: "",
				EnableFeature:  "",
				LogLevel:       "info",
				SSHConfigPath:  "/tmp/custom_config",
				SetTheme:       "",
			},
			wantError: false,
		}, {
			name: "Set home app folder",
			args: []string{"-f", "/tmp/home2"},
			wantConfig: &Configuration{
				AppHome:        "/tmp/home2",
				AppMode:        "",
				DisableFeature: "",
				EnableFeature:  "",
				LogLevel:       "info",
				SSHConfigPath:  "/tmp/custom_config",
				SetTheme:       "",
			},
			wantError: false,
		}, {
			name: "Set log level",
			args: []string{"-l", "debug"},
			wantConfig: &Configuration{
				AppHome:        "/tmp/home",
				AppMode:        "",
				DisableFeature: "",
				EnableFeature:  "",
				LogLevel:       "debug",
				SSHConfigPath:  "/tmp/custom_config",
				SetTheme:       "",
			},
			wantError: false,
		}, {
			name: "Set SSH config file path",
			args: []string{"-s", "/tmp/custom_config2"},
			wantConfig: &Configuration{
				AppHome:        "/tmp/home",
				AppMode:        "",
				DisableFeature: "",
				EnableFeature:  "",
				LogLevel:       "info",
				SSHConfigPath:  "/tmp/custom_config2",
				SetTheme:       "",
			},
			wantError: false,
		}, {
			name: "Set theme",
			args: []string{"--set-theme", "dark"},
			wantConfig: &Configuration{
				AppHome:        "/tmp/home",
				AppMode:        "HANDLE_PARAM",
				DisableFeature: "",
				EnableFeature:  "",
				LogLevel:       "info",
				SSHConfigPath:  "/tmp/custom_config",
				SetTheme:       "dark",
			},
			wantError: false,
		}, {
			name: "Enable feature",
			args: []string{"-e", "ssh_config"},
			wantConfig: &Configuration{
				AppHome:        "/tmp/home",
				AppMode:        "HANDLE_PARAM",
				DisableFeature: "",
				EnableFeature:  "ssh_config",
				LogLevel:       "info",
				SSHConfigPath:  "/tmp/custom_config",
				SetTheme:       "",
			},
			wantError: false,
		}, {
			name: "Disable feature",
			args: []string{"-d", "ssh_config"},
			wantConfig: &Configuration{
				AppHome:        "/tmp/home",
				AppMode:        "HANDLE_PARAM",
				DisableFeature: "ssh_config",
				EnableFeature:  "",
				LogLevel:       "info",
				SSHConfigPath:  "/tmp/custom_config",
				SetTheme:       "",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"app_name"}, tt.args...)
			cfg, err := parseCommandLineFlags(envConfig, args, false)

			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, cfg)
				return
			}

			require.NotNil(t, cfg)
			require.NoError(t, err)
			require.Equal(t, tt.wantConfig.AppMode, cfg.AppMode)
			require.Equal(t, tt.wantConfig.AppHome, cfg.AppHome)
			require.Equal(t, tt.wantConfig.DisableFeature, cfg.DisableFeature)
			require.Equal(t, tt.wantConfig.EnableFeature, cfg.EnableFeature)
			require.Equal(t, tt.wantConfig.LogLevel, cfg.LogLevel)
			require.Equal(t, tt.wantConfig.SSHConfigPath, cfg.SSHConfigPath)
			require.Equal(t, tt.wantConfig.SetTheme, cfg.SetTheme)
		})
	}
}

func Test_setConfigDefaults(t *testing.T) {
	// Test with valid config
	tempDir := t.TempDir()
	config := &Configuration{
		AppHome:  tempDir,
		LogLevel: constant.LogLevelType.INFO,
	}

	finalConfig, err := setConfigDefaults(config)
	require.NoError(t, err)
	require.Equal(t, appName, finalConfig.AppName)
	require.Equal(t, tempDir, finalConfig.AppHome)
	require.Equal(t, constant.AppModeType.StartUI, finalConfig.AppMode)
	require.Equal(t, constant.LogLevelType.INFO, finalConfig.LogLevel)

	// Test with invalid app home
	config = &Configuration{
		AppHome:  "/tmp/nonexistent",
		LogLevel: constant.LogLevelType.INFO,
	}

	_, err = setConfigDefaults(config)
	require.Error(t, err)

	// Test with unsupported log level
	config = &Configuration{
		AppHome:  tempDir,
		LogLevel: "unsupported_level",
	}

	_, err = setConfigDefaults(config)
	require.Error(t, err)
}
