package webdav

import (
	"io"
)

type Uploader struct {
	relpath string
	writer  io.WriteCloser
	done    chan int
	n       int64
}

func (out *Uploader) Write(b []byte) (int, error) {
	n, err := out.writer.Write(b)
	if err != nil {
		return n, err
	}
	out.n += int64(n)
	return n, nil
}

func (out *Uploader) Close() error {
	err := out.writer.Close()
	<-out.done
	return err
}

func (out *Uploader) WroteCompressed() int64 {
	return out.n
}

func (out *Uploader) WroteUncompressed() int64 {
	return out.n
}

func (out *Uploader) Path() string {
	return out.relpath
}

func (out *Uploader) Cancel() error {
	return nil
}
