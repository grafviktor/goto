package sshconfig

type tokenEnum string

var tokenKind = struct {
	Host         tokenEnum
	User         tokenEnum
	Hostname     tokenEnum
	NetworkPort  tokenEnum
	Unsupported  tokenEnum
	IncludeFile  tokenEnum
	IdentityFile tokenEnum
	Group        tokenEnum
	Description  tokenEnum
}{
	Host:         "Host",
	User:         "User",
	Hostname:     "HostName",
	NetworkPort:  "Port",
	Unsupported:  "Unsupported",
	IncludeFile:  "Include",
	IdentityFile: "IdentityFile",
	Group:        "Group",
	Description:  "Description",
}

type SSHToken struct {
	key   string
	value string
	kind  tokenEnum
}
