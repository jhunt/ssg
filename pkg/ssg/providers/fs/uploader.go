package fs

import (
	"os"
)

type Uploader struct {
	file *os.File
	relpath string
	abspath string
}

func (out *Uploader) Write(b []byte) (int, error) {
	return out.file.Write(b)
}

func (out *Uploader) Close() error {
	return out.file.Close()
}

func (out *Uploader) Path() string {
	return out.relpath
}

func (out *Uploader) Cancel() error {
	return os.Remove(out.abspath)
}
