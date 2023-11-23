package utils

import (
	"os/exec"
	"testing"

	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/utils/ssh"
)

func TestCheckAppInstalled(t *testing.T) {
	tests := []struct {
		name          string
		appName       string
		expectedError bool
	}{
		{
			name:          "Installed App",
			appName:       "echo", // Assuming 'echo' is always installed
			expectedError: false,
		},
		{
			name:          "Uninstalled App",
			appName:       "nonexistentapp",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAppInstalled(tt.appName)

			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}

func TestHostModelToOptionsAdaptor(t *testing.T) {
	tests := []struct {
		name            string
		host            model.Host
		expectedOptions []ssh.CommandLineOption
	}{
		{
			name: "Valid Host",
			host: model.Host{
				Address:        "example.com",
				LoginName:      "user",
				RemotePort:     "22",
				PrivateKeyPath: "/path/to/private_key",
			},
			expectedOptions: []ssh.CommandLineOption{
				ssh.OptionAddress{Value: "example.com"},
				ssh.OptionLoginName{Value: "user"},
				ssh.OptionRemotePort{Value: "22"},
				ssh.OptionPrivateKey{Value: "/path/to/private_key"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HostModelToOptionsAdaptor(tt.host)

			if len(result) != len(tt.expectedOptions) {
				t.Errorf("Expected %d options, but got %d", len(tt.expectedOptions), len(result))
			}

			for i := range result {
				if result[i] != tt.expectedOptions[i] {
					t.Errorf("Expected option %v, but got %v", tt.expectedOptions[i], result[i])
				}
			}
		})
	}
}

func TestBuildProcess(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		expectedCmd *exec.Cmd
	}{
		{
			name:        "Simple Command",
			cmd:         "cd",
			expectedCmd: exec.Command("cd"),
		},
		{
			name:        "Command with Arguments",
			cmd:         "echo hello",
			expectedCmd: exec.Command("echo", "hello"),
		},
		{
			name:        "Empty Command",
			cmd:         "",
			expectedCmd: nil, // Expecting nil as there is no valid command
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildProcess(tt.cmd)

			switch {
			case tt.expectedCmd == nil && result != nil:
				t.Errorf("Expected nil, but got %+v", result)

			case tt.expectedCmd != nil && result == nil:
				t.Errorf("Expected %+v, but got nil", tt.expectedCmd)
			case tt.expectedCmd != nil && result != nil:
				// Compare relevant fields of the Cmd struct
				if tt.expectedCmd.Path != result.Path {
					t.Errorf("Expected Path %s, but got %s", tt.expectedCmd.Path, result.Path)
				}
				if len(tt.expectedCmd.Args) != len(result.Args) {
					t.Errorf("Expected %d arguments, but got %d", len(tt.expectedCmd.Args), len(result.Args))
				}
				for i := range tt.expectedCmd.Args {
					if tt.expectedCmd.Args[i] != result.Args[i] {
						t.Errorf("Expected argument %s, but got %s", tt.expectedCmd.Args[i], result.Args[i])
					}
				}
			}
		})
	}
}
