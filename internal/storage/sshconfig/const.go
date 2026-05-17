package sshconfig

type pathType string

const (
	pathTypeFile pathType = "file"
	pathTypeURL  pathType = "url"
)

type configSource struct {
	path string
	kind pathType
}
