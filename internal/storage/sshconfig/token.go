package sshconfig

type tokenEnum string

var TokenType = struct {
	HOST          tokenEnum
	USER          tokenEnum
	HOSTNAME      tokenEnum
	NETWORK_PORT  tokenEnum
	UNSUPPORTED   tokenEnum
	INCLUDE_FILE  tokenEnum
	IDENTITY_FILE tokenEnum
	GROUP         tokenEnum
	DESCRIPTION   tokenEnum
}{
	HOST:          "Host",
	USER:          "User",
	HOSTNAME:      "HostName",
	NETWORK_PORT:  "Port",
	UNSUPPORTED:   "Unsupported",
	INCLUDE_FILE:  "Include",
	IDENTITY_FILE: "IdentityFile",
	GROUP:         "Group",
	DESCRIPTION:   "Description",
}

type Token struct {
	key   string
	value string
	Type  tokenEnum
}

func (t *Token) Key() string {
	return t.key
}

func (t *Token) Value() string {
	return t.value
}

func (t *Token) String() string {
	return t.key + ": " + t.value
}
