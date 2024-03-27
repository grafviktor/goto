package ssh

import (
	"strings"
	"testing"
)

func Test_ConstructKeyValueOption(t *testing.T) {
	tests := []struct {
		name           string
		optionFlag     string
		optionValue    string
		expectedResult string
	}{
		{
			name:           "Option with value",
			optionFlag:     "-i",
			optionValue:    "private_key",
			expectedResult: " -i private_key",
		},
		{
			name:           "Option with empty value",
			optionFlag:     "-p",
			optionValue:    "",
			expectedResult: "",
		},
		{
			name:           "Option with space-padded value",
			optionFlag:     "-l",
			optionValue:    "  login_name  ",
			expectedResult: " -l login_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constructKeyValueOption(tt.optionFlag, tt.optionValue)

			if result != tt.expectedResult {
				t.Errorf("Expected result %s, but got %s", tt.expectedResult, result)
			}
		})
	}
}

func Test_AddOption(t *testing.T) {
	tests := []struct {
		name           string
		rawParameter   CommandLineOption
		expectedResult string
	}{
		{
			name:           "OptionPrivateKey with value",
			rawParameter:   OptionPrivateKey{Value: "private_key"},
			expectedResult: " -i private_key",
		},
		{
			name:           "OptionRemotePort with empty value",
			rawParameter:   OptionRemotePort{Value: ""},
			expectedResult: "",
		},
		{
			name:           "OptionLoginName with value",
			rawParameter:   OptionLoginName{Value: "login_name"},
			expectedResult: " -l login_name",
		},
		{
			name:           "OptionAddress with empty value",
			rawParameter:   OptionAddress{Value: ""},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			addOption(&sb, tt.rawParameter)

			result := sb.String()
			if result != tt.expectedResult {
				t.Errorf("Expected result %s, but got %s", tt.expectedResult, result)
			}
		})
	}
}

func Test_ConstructCMD(t *testing.T) {
	tests := []struct {
		name           string
		cmd            string
		options        []CommandLineOption
		expectedResult string
	}{
		{
			name:           "Command with Options",
			cmd:            "ssh",
			options:        []CommandLineOption{OptionPrivateKey{Value: "private_key"}, OptionRemotePort{Value: "22"}},
			expectedResult: "ssh -i private_key -p 22",
		},
		{
			name:           "Command without Options",
			cmd:            "ls",
			options:        []CommandLineOption{},
			expectedResult: "ls",
		},
		{
			name:           "Command with Address Option",
			cmd:            "ping",
			options:        []CommandLineOption{OptionAddress{Value: "example.com"}},
			expectedResult: "ping example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConstructCMD(tt.cmd, tt.options...)

			if result != tt.expectedResult {
				t.Errorf("Expected result %s, but got %s", tt.expectedResult, result)
			}
		})
	}
}
