// Package sshconfig contains SSH related models and methods.
package sshconfig

import (
	"os/user"
	"regexp"

	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/utils"
)

// Config struct contains values loaded from ~/.ssh_config file.
type Config struct {
	// Values which should be extracted from 'ssh -G <hostname>' command:
	// 1. 'hostname'
	// 2. 'identityfile'
	// 3. 'port'
	// 4. 'user'
	Hostname     string
	IdentityFile string
	Port         string
	User         string
}

// Parse - parses 'ssh -G <hostname> command' output and returns Config struct.
func Parse(config string) *Config {
	return &Config{
		Hostname:     getRegexFirstMatchingGroup(sshConfigHostnameRe.FindStringSubmatch(config)),
		IdentityFile: getRegexFirstMatchingGroup(sshConfigIdentityFileRe.FindStringSubmatch(config)),
		Port:         getRegexFirstMatchingGroup(sshConfigPortRe.FindStringSubmatch(config)),
		User:         getRegexFirstMatchingGroup(sshConfigUserRe.FindStringSubmatch(config)),
	}
}

// StubConfig - returns a stub SSH config. It is used on application startup when build application state and
// no hosts yet available. Consider to run real ssh process to request a config. See 'message.RunProcessLoadSSHConfig'.
func StubConfig() *Config {
	return &Config{
		Hostname:     "Loading, please wait...",
		IdentityFile: "~/.ssh/id_rsa",
		Port:         "22",
		User:         currentUsername(),
	}
}

// currentUsername - returns current OS username or "n/a" if it can't be determined.
func currentUsername() string {
	// ssh [-vvv] -G <hostname> is used to request settings for a hostname.
	// for a stub config use u.Current()
	u, err := user.Current()
	if err != nil {
		return "n/a"
	}

	return u.Username
}

var (
	sshConfigHostnameRe     = regexp.MustCompile(`(?i)hostname\s+(.*[^\r\n])`)
	sshConfigIdentityFileRe = regexp.MustCompile(`(?i)identityfile\s+(.*[^\r\n])`)
	sshConfigPortRe         = regexp.MustCompile(`(?i)port\s+(.*[^\r\n])`)
	sshConfigUserRe         = regexp.MustCompile(`(?i)user\s+(.*[^\r\n])`)
)

func getRegexFirstMatchingGroup(groups []string) string {
	if len(groups) > 1 {
		return groups[1]
	}

	return ""
}

func IsAlternativeFilePathDefined() bool {
	userDefinedConfig := state.Get().ApplicationConfig.UserConfig.SSHConfigFilePath
	defaultConfig, _ := utils.SSHConfigFilePath("")

	return userDefinedConfig != defaultConfig
}

func GetFilePath() string {
	return state.Get().ApplicationConfig.UserConfig.SSHConfigFilePath
}
