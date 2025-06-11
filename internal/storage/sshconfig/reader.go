package sshconfig

import (
	"io"
	"os"
	"strings"
)

func newReader(value, kind string) (*reader, error) {
	if kind == "file" {
		file, err := os.Open(value)
		if err != nil {
			return nil, err
		}

		return &reader{
			kind:   kind,
			reader: file,
			closer: nil,
		}, nil
	}

	return &reader{
		kind:   kind,
		reader: strings.NewReader(value),
		closer: nil,
	}, nil
}

type reader struct {
	kind   string
	reader io.Reader
	closer io.Closer
}

func (r reader) Read(p []byte) (n int, err error) {
	if r.reader == nil {
		return 0, io.EOF
	}
	return r.reader.Read(p)
}

func (r reader) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}
