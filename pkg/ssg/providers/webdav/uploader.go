package webdav

import (
	"io"
)

type Uploader struct {
	relpath string
	writer io.WriteCloser
	done chan int
}

func (out *Uploader) Write(b []byte) (int, error) {
	return out.writer.Write(b)
}

func (out *Uploader) Close() error {
	err := out.writer.Close()
	<-out.done
	return err
}

func (out *Uploader) Path() string {
	return out.relpath
}

func (out *Uploader) Cancel() error {
	return nil
}
