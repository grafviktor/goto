package sshconfig

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/grafviktor/goto/internal/utils"
)

func newReader(value, kind string) (*reader, error) {
	switch kind {
	case "url":
		urlReader, err := utils.FetchFromURL(value)
		if err != nil {
			return nil, err
		}

		return &reader{
			kind:   "url",
			reader: urlReader,
			closer: urlReader,
		}, nil
	case "file":
		stat, err := os.Stat(value)
		if err != nil {
			return nil, err
		}

		if stat.IsDir() {
			return nil, errors.New("SSH config file path is a directory")
		}

		file, err := os.Open(value)
		if err != nil {
			return nil, err
		}

		return &reader{
			kind:   kind,
			reader: file,
			closer: file,
		}, nil
	default:
		return &reader{
			kind:   kind,
			reader: strings.NewReader(value),
			closer: nil,
		}, nil
	}
}

type reader struct {
	kind   string
	reader io.Reader
	closer io.Closer
}

func (r reader) Read(p []byte) (int, error) {
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
