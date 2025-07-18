package sshconfig

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/application"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

func TestGetCurrentOSUser(t *testing.T) {
	username := currentUsername()
	require.NotEmpty(t, username, "GetCurrentOSUser should return a non-empty string")
}

var (
	windowsMockSSHConfig = "user mock_domain\\mock_user\r\nhostname mock_hostname\r\nport 22\r\naddkeystoagent false\r\naddressfamily any\r\nbatchmode no\r\ncanonicalizefallbacklocal yes\r\ncanonicalizehostname false\r\nchallengeresponseauthentication yes\r\ncheckhostip yes\r\ncompression no\r\ncontrolmaster false\r\nenablesshkeysign no\r\nclearallforwardings no\r\nexitonforwardfailure no\r\nfingerprinthash SHA256\r\nforwardagent no\r\nforwardx11 no\r\nforwardx11trusted no\r\ngatewayports no\r\ngssapiauthentication no\r\ngssapidelegatecredentials no\r\nhashknownhosts no\r\nhostbasedauthentication no\r\nidentitiesonly no\r\nkbdinteractiveauthentication yes\r\nnohostauthenticationforlocalhost no\r\npasswordauthentication yes\r\npermitlocalcommand no\r\nproxyusefdpass no\r\npubkeyauthentication yes\r\nrequesttty auto\r\nstreamlocalbindunlink no\r\nstricthostkeychecking ask\r\ntcpkeepalive yes\r\ntunnel false\r\nverifyhostkeydns false\r\nvisualhostkey no\r\nupdatehostkeys false\r\ncanonicalizemaxdots 1\r\nconnectionattempts 1\r\nforwardx11timeout 1200\r\nnumberofpasswordprompts 3\r\nserveralivecountmax 3\r\nserveraliveinterval 0\r\nciphers chacha20-poly1305@openssh.com,aes128-ctr,aes192-ctr,aes256-ctr,aes128-gcm@openssh.com,aes256-gcm@openssh.com\r\nhostkeyalgorithms ecdsa-sha2-nistp256-cert-v01@openssh.com,ecdsa-sha2-nistp384-cert-v01@openssh.com,ecdsa-sha2-nistp521-cert-v01@openssh.com,ssh-ed25519-cert-v01@openssh.com,rsa-sha2-512-cert-v01@openssh.com,rsa-sha2-256-cert-v01@openssh.com,ssh-rsa-cert-v01@openssh.com,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,ssh-ed25519,rsa-sha2-512,rsa-sha2-256,ssh-rsa\r\nhostbasedkeytypes ecdsa-sha2-nistp256-cert-v01@openssh.com,ecdsa-sha2-nistp384-cert-v01@openssh.com,ecdsa-sha2-nistp521-cert-v01@openssh.com,ssh-ed25519-cert-v01@openssh.com,rsa-sha2-512-cert-v01@openssh.com,rsa-sha2-256-cert-v01@openssh.com,ssh-rsa-cert-v01@openssh.com,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,ssh-ed25519,rsa-sha2-512,rsa-sha2-256,ssh-rsa\r\nkexalgorithms curve25519-sha256,curve25519-sha256@libssh.org,ecdh-sha2-nistp256,ecdh-sha2-nistp384,ecdh-sha2-nistp521,diffie-hellman-group-exchange-sha256,diffie-hellman-group16-sha512,diffie-hellman-group18-sha512,diffie-hellman-group14-sha256,diffie-hellman-group14-sha1\r\ncasignaturealgorithms ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,ssh-ed25519,rsa-sha2-512,rsa-sha2-256,ssh-rsa\r\nloglevel INFO\r\nmacs umac-64-etm@openssh.com,umac-128-etm@openssh.com,hmac-sha2-256-etm@openssh.com,hmac-sha2-512-etm@openssh.com,hmac-sha1-etm@openssh.com,umac-64@openssh.com,umac-128@openssh.com,hmac-sha2-256,hmac-sha2-512,hmac-sha1\r\npubkeyacceptedkeytypes ecdsa-sha2-nistp256-cert-v01@openssh.com,ecdsa-sha2-nistp384-cert-v01@openssh.com,ecdsa-sha2-nistp521-cert-v01@openssh.com,ssh-ed25519-cert-v01@openssh.com,rsa-sha2-512-cert-v01@openssh.com,rsa-sha2-256-cert-v01@openssh.com,ssh-rsa-cert-v01@openssh.com,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,ssh-ed25519,rsa-sha2-512,rsa-sha2-256,ssh-rsa\r\nxauthlocation /usr/X11R6/bin/xauth\r\nidentityfile c:/temp/mock_rsa_file\r\nidentityfile ~/.ssh/id_dsa\r\nidentityfile ~/.ssh/id_ecdsa\r\nidentityfile ~/.ssh/id_ed25519\r\nidentityfile ~/.ssh/id_xmss\r\ncanonicaldomains\r\nglobalknownhostsfile __PROGRAMDATA__\\ssh/ssh_known_hosts __PROGRAMDATA__\\ssh/ssh_known_hosts2\r\nuserknownhostsfile ~/.ssh/known_hosts ~/.ssh/known_hosts2\r\nconnecttimeout none\r\ntunneldevice any:any\r\ncontrolpersist no\r\nescapechar ~\r\nipqos af21 cs1\r\nrekeylimit 0 0\r\nstreamlocalbindmask 0177\r\nsyslogfacility USER"
	unixMockSSHConfig    = "user mock_user\nhostname mock_hostname\nport 22\naddkeystoagent false\naddressfamily any\nbatchmode no\ncanonicalizefallbacklocal yes\ncanonicalizehostname false\nchallengeresponseauthentication yes\ncheckhostip yes\ncompression no\ncontrolmaster false\nenablesshkeysign no\nclearallforwardings no\nexitonforwardfailure no\nfingerprinthash SHA256\nforwardx11 no\nforwardx11trusted yes\ngatewayports no\ngssapiauthentication yes\ngssapikeyexchange no\ngssapidelegatecredentials no\ngssapitrustdns no\ngssapirenewalforcesrekey no\ngssapikexalgorithms gss-gex-sha1-,gss-group14-sha1-\nhashknownhosts yes\nhostbasedauthentication no\nidentitiesonly no\nkbdinteractiveauthentication yes\nnohostauthenticationforlocalhost no\npasswordauthentication yes\npermitlocalcommand no\nproxyusefdpass no\npubkeyauthentication yes\nrequesttty auto\nstreamlocalbindunlink no\nstricthostkeychecking ask\ntcpkeepalive yes\ntunnel false\nverifyhostkeydns false\nvisualhostkey no\nupdatehostkeys false\ncanonicalizemaxdots 1\nconnectionattempts 1\nforwardx11timeout 1200\nnumberofpasswordprompts 3\nserveralivecountmax 5\nserveraliveinterval 30\nciphers chacha20-poly1305@openssh.com,aes128-ctr,aes192-ctr,aes256-ctr,aes128-gcm@openssh.com,aes256-gcm@openssh.com\nhostkeyalgorithms ecdsa-sha2-nistp256-cert-v01@openssh.com,ecdsa-sha2-nistp384-cert-v01@openssh.com,ecdsa-sha2-nistp521-cert-v01@openssh.com,sk-ecdsa-sha2-nistp256-cert-v01@openssh.com,ssh-ed25519-cert-v01@openssh.com,sk-ssh-ed25519-cert-v01@openssh.com,rsa-sha2-512-cert-v01@openssh.com,rsa-sha2-256-cert-v01@openssh.com,ssh-rsa-cert-v01@openssh.com,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,sk-ecdsa-sha2-nistp256@openssh.com,ssh-ed25519,sk-ssh-ed25519@openssh.com,rsa-sha2-512,rsa-sha2-256,ssh-rsa\nhostbasedkeytypes ecdsa-sha2-nistp256-cert-v01@openssh.com,ecdsa-sha2-nistp384-cert-v01@openssh.com,ecdsa-sha2-nistp521-cert-v01@openssh.com,sk-ecdsa-sha2-nistp256-cert-v01@openssh.com,ssh-ed25519-cert-v01@openssh.com,sk-ssh-ed25519-cert-v01@openssh.com,rsa-sha2-512-cert-v01@openssh.com,rsa-sha2-256-cert-v01@openssh.com,ssh-rsa-cert-v01@openssh.com,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,sk-ecdsa-sha2-nistp256@openssh.com,ssh-ed25519,sk-ssh-ed25519@openssh.com,rsa-sha2-512,rsa-sha2-256,ssh-rsa\nkexalgorithms curve25519-sha256,curve25519-sha256@libssh.org,ecdh-sha2-nistp256,ecdh-sha2-nistp384,ecdh-sha2-nistp521,diffie-hellman-group-exchange-sha256,diffie-hellman-group16-sha512,diffie-hellman-group18-sha512,diffie-hellman-group14-sha256,diffie-hellman-group14-sha1\ncasignaturealgorithms ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,sk-ecdsa-sha2-nistp256@openssh.com,ssh-ed25519,sk-ssh-ed25519@openssh.com,rsa-sha2-512,rsa-sha2-256\nloglevel INFO\nmacs umac-64-etm@openssh.com,umac-128-etm@openssh.com,hmac-sha2-256-etm@openssh.com,hmac-sha2-512-etm@openssh.com,hmac-sha1-etm@openssh.com,umac-64@openssh.com,umac-128@openssh.com,hmac-sha2-256,hmac-sha2-512,hmac-sha1\nsecuritykeyprovider internal\npubkeyacceptedkeytypes ecdsa-sha2-nistp256-cert-v01@openssh.com,ecdsa-sha2-nistp384-cert-v01@openssh.com,ecdsa-sha2-nistp521-cert-v01@openssh.com,sk-ecdsa-sha2-nistp256-cert-v01@openssh.com,ssh-ed25519-cert-v01@openssh.com,sk-ssh-ed25519-cert-v01@openssh.com,rsa-sha2-512-cert-v01@openssh.com,rsa-sha2-256-cert-v01@openssh.com,ssh-rsa-cert-v01@openssh.com,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,sk-ecdsa-sha2-nistp256@openssh.com,ssh-ed25519,sk-ssh-ed25519@openssh.com,rsa-sha2-512,rsa-sha2-256,ssh-rsa\nxauthlocation /usr/bin/xauth\nidentityfile ~/.ssh/mock_rsa_file\nidentityfile ~/.ssh/id_dsa\nidentityfile ~/.ssh/id_ecdsa\nidentityfile ~/.ssh/id_ecdsa_sk\nidentityfile ~/.ssh/id_ed25519\nidentityfile ~/.ssh/id_ed25519_sk\nidentityfile ~/.ssh/id_xmss\ncanonicaldomains\nglobalknownhostsfile /etc/ssh/ssh_known_hosts /etc/ssh/ssh_known_hosts2\nuserknownhostsfile ~/.ssh/known_hosts ~/.ssh/known_hosts2\nsendenv LANG\nsendenv LC_*\nforwardagent no\nconnecttimeout none\ntunneldevice any:any\ncontrolpersist no\nescapechar ~\nipqos lowdelay throughput\nrekeylimit 0 0\nstreamlocalbindmask 0177\nsyslogfacility USER"
)

func TestParseConfig(t *testing.T) {
	// Windows uses '\r\n' for lines ending.
	expected := &Config{
		Hostname:     "mock_hostname",
		IdentityFile: "c:/temp/mock_rsa_file",
		User:         "mock_domain\\mock_user",
		Port:         "22",
	}

	actual := Parse(windowsMockSSHConfig)
	require.Equal(t, expected, actual)

	// UNIX uses '\n' for lines ending.
	expected = &Config{
		Hostname:     "mock_hostname",
		IdentityFile: "~/.ssh/mock_rsa_file",
		User:         "mock_user",
		Port:         "22",
	}

	actual = Parse(unixMockSSHConfig)
	require.Equal(t, expected, actual)
}

func TestIsAlternativeFilePathDefined(t *testing.T) {
	// Create a mock logger for testing
	mockLogger := mocklogger.Logger{}
	state.Create(context.TODO(), application.Configuration{}, &mockLogger)

	// No custom path is set
	state.Get().ApplicationConfig.SSHConfigFilePath = ""
	actual := IsAlternativeFilePathDefined()
	require.False(t, actual, "IsAlternativeFilePathDefined should return false when no custom path is set")

	// Custom path is set
	state.Get().ApplicationConfig.SSHConfigFilePath = "/custom/path/to/ssh_config"
	actual = IsAlternativeFilePathDefined()
	require.True(t, actual, "IsAlternativeFilePathDefined should return true when custom path is set")
}
