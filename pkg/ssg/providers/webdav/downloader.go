package webdav

import (
	"io"
)

type Downloader struct {
	path string
}

func (in *Downloader) Read(b []byte) (int, error) {
	return 0, io.EOF
}
