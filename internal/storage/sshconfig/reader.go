package sshconfig

import (
	"io"
	"os"
	"strings"

	"github.com/grafviktor/goto/internal/utils"
)

func newReader(value, kind string) (*reader, error) {
	if kind == "file" {
		// Check if the value is a URL
		if utils.IsURLPath(value) {
			urlReader, err := utils.FetchFromURL(value)
			if err != nil {
				return nil, err
			}

			return &reader{
				kind:   "url",
				reader: urlReader,
				closer: urlReader,
			}, nil
		}

		// Regular file handling
		file, err := os.Open(value)
		if err != nil {
			return nil, err
		}

		return &reader{
			kind:   kind,
			reader: file,
			closer: file,
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
