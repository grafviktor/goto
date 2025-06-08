package sshconfig

import (
	"io"
	"os"
	"strings"
)

func newReader(value string, kind string) (*closeableReader, error) {
	if kind == "file" {
		file, err := os.Open(value)
		if err != nil {
			return nil, err
		}

		return &closeableReader{
			kind:   kind,
			reader: file,
			closer: nil,
		}, nil
	}

	return &closeableReader{
		kind:   kind,
		reader: strings.NewReader(value),
		closer: nil,
	}, nil
}

type closeableReader struct {
	kind   string
	reader io.Reader
	closer io.Closer
}

func (r closeableReader) Read(p []byte) (n int, err error) {
	if r.reader == nil {
		return 0, io.EOF
	}
	return r.reader.Read(p)
}

func (r closeableReader) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}
