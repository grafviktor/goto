package sshconfig

type valueTypeEnum string

const (
	valueTypeFile valueTypeEnum = "file"
	valueTypeURL  valueTypeEnum = "url"
	valueTypeRaw  valueTypeEnum = "raw"
)

type configSource struct {
	value     string
	valueType valueTypeEnum
}
