package ssh

import (
	"os/user"
	"regexp"
)

// Config struct contains values loaded from ~/.ssh_config file.
type Config struct {
	// Values which should be extracted from 'ssh -G <hostname>' command:
	// 1. 'identityfile'
	// 2. 'user'
	// 3. 'port'
	IdentityFile string
	User         string
	Port         string
}

// currentUsername - returns current OS username or "n/a" if it can't be determined.
func currentUsername() string {
	// ssh [-vvv] -G <hostname> should be used to request settings for a hostname.
	user, err := user.Current()
	if err != nil {
		return "n/a"
	}

	return user.Username
}

// DefaultConfig - returns a stub SSH config. It is used on application startup when build application state and
// no hosts yet available. Consider to run real ssh process to request a config. See 'message.RunProcessLoadSSHConfig'.
func DefaultConfig() *Config {
	return &Config{
		IdentityFile: "$HOME/.ssh/id_rsa",
		Port:         "22",
		User:         currentUsername(),
	}
}

var (
	sshConfigUserRe         = regexp.MustCompile(`(?i)user\s+(.*)`)
	sshConfigPortRe         = regexp.MustCompile(`(?i)port\s+(.*)`)
	sshConfigIdentityFileRe = regexp.MustCompile(`(?i)identityfile\s+(.*)`)
)

func getRegexFirstMatchingGroup(groups []string) string {
	if len(groups) > 1 {
		return groups[1]
	}

	return ""
}

// ParseConfig - parses 'ssh -G <hostname> command' output and returns Config struct.
func ParseConfig(config string) *Config {
	return &Config{
		IdentityFile: getRegexFirstMatchingGroup(sshConfigIdentityFileRe.FindStringSubmatch(config)),
		Port:         getRegexFirstMatchingGroup(sshConfigPortRe.FindStringSubmatch(config)),
		User:         getRegexFirstMatchingGroup(sshConfigUserRe.FindStringSubmatch(config)),
	}
}
