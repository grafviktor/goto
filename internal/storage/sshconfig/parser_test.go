package sshconfig

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

type mockLexer struct {
	tokens []sshToken
}

func (m *mockLexer) Tokenize() []sshToken {
	return m.tokens
}

func TestParser_Parse_SingleHost(t *testing.T) {
	lexer := &mockLexer{
		tokens: []sshToken{
			{kind: tokenKind.Host, value: "testhost"},
			{kind: tokenKind.Hostname, value: "example.com"},
			{kind: tokenKind.User, value: "alice"},
			{kind: tokenKind.NetworkPort, value: "2222"},
			{kind: tokenKind.IdentityFile, value: "~/.ssh/id_rsa"},
			{kind: tokenKind.Group, value: "devops"},
			{kind: tokenKind.Description, value: "desc"},
		},
	}
	parser := NewParser(lexer, &mocklogger.Logger{})
	hosts, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, hosts, 1)
	h := hosts[0]
	require.Equal(t, "testhost", h.Title)
	require.Equal(t, "example.com", h.Address)
	require.Equal(t, "alice", h.LoginName)
	require.Equal(t, "2222", h.RemotePort)
	require.Equal(t, "~/.ssh/id_rsa", h.IdentityFilePath)
	require.Equal(t, "devops", h.Group)
	require.Equal(t, "desc", h.Description)
}

func TestParser_Parse_MultipleHosts(t *testing.T) {
	lexer := &mockLexer{
		tokens: []sshToken{
			{kind: tokenKind.Host, value: "host1"},
			{kind: tokenKind.Hostname, value: "host1.com"},
			{kind: tokenKind.Host, value: "host2"},
			{kind: tokenKind.Hostname, value: "host2.com"},
		},
	}
	parser := NewParser(lexer, &mocklogger.Logger{})
	hosts, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, hosts, 2)
	require.Equal(t, "host1", hosts[0].Title)
	require.Equal(t, "host1.com", hosts[0].Address)
	require.Equal(t, "host2", hosts[1].Title)
	require.Equal(t, "host2.com", hosts[1].Address)
}

func TestParser_Parse_InvalidHost(t *testing.T) {
	lexer := &mockLexer{
		tokens: []sshToken{
			// This token is invalid because it does not have a hostname
			{kind: tokenKind.Host, value: "*"},
			{kind: tokenKind.Hostname, value: "bad.com"},
			// This token is valid and will be added to the hosts
			{kind: tokenKind.Host, value: "good"},
			{kind: tokenKind.Hostname, value: "good.com"},
		},
	}
	parser := NewParser(lexer, &mocklogger.Logger{})
	hosts, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, hosts, 1)
	require.Equal(t, "good", hosts[0].Title)
	require.Equal(t, "good.com", hosts[0].Address)
}

func TestParser_Parse_EmptyLexer(t *testing.T) {
	lexer := &mockLexer{}
	parser := NewParser(lexer, &mocklogger.Logger{})
	hosts, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, hosts, 0)
}

func TestParser_Parse_DefaultGroup(t *testing.T) {
	lexer := &mockLexer{
		tokens: []sshToken{
			{kind: tokenKind.Host, value: "host1"},
			{kind: tokenKind.Hostname, value: "host1.com"},
		},
	}
	parser := NewParser(lexer, &mocklogger.Logger{})
	hosts, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, hosts, 1)
	require.Equal(t, "ssh_config", hosts[0].Group)
}

func TestParser_Parse_NoLexer(t *testing.T) {
	parser := NewParser(nil, &mocklogger.Logger{})
	hosts, err := parser.Parse()
	require.Error(t, err)
	require.Nil(t, hosts)
}
