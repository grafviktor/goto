package sshconfig

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/grafviktor/goto/internal/utils"
)

func newReader(value string, kind valueTypeEnum) (*reader, error) {
	switch kind {
	case valueTypeURL:
		urlReader, err := utils.FetchFromURL(value)
		if err != nil {
			return nil, err
		}

		return &reader{
			kind:   kind,
			reader: urlReader,
			closer: urlReader,
		}, nil
	case valueTypeFile:
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
		// For raw value, we can directly create a reader from the string. This is a unit test path.
		return &reader{
			kind:   kind,
			reader: strings.NewReader(value),
			closer: nil,
		}, nil
	}
}

type reader struct {
	kind   valueTypeEnum
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
