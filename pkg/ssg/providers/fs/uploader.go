package fs

import (
	"os"
)

type Uploader struct {
	file *os.File
	relpath string
	abspath string
	n       int64
}

func (out *Uploader) Write(b []byte) (int, error) {
	n, err := out.file.Write(b)
	if err != nil {
		return n, err
	}
	out.n += int64(n)
	return n, nil
}

func (out *Uploader) Close() error {
	return out.file.Close()
}

func (out *Uploader) SentCompressed() int64 {
	return out.n
}

func (out *Uploader) SentUncompressed() int64 {
	return out.n
}

func (out *Uploader) Path() string {
	return out.relpath
}

func (out *Uploader) Cancel() error {
	return os.Remove(out.abspath)
}
