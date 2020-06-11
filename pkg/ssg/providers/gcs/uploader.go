package gcs

import (
	"io"

	"google.golang.org/api/storage/v1"
)

type Uploader struct {
	errors chan error
	writer io.WriteCloser
	object *storage.ObjectsInsertCall
	key    string
	n      int64
}

func (out *Uploader) Write(b []byte) (int, error) {
	if out.errors == nil {
		out.errors = make(chan error, 1)
		rd, wr := io.Pipe()
		go func() {
			_, err := out.object.Media(rd).Do()
			if err != nil {
				out.errors <- err
			}
			close(out.errors)
		}()
		out.writer = wr
	}

	n, err := out.writer.Write(b)
	if err == nil || err == io.EOF {
		out.n += int64(n)
	}
	return n, err
}

func (out *Uploader) Close() error {
	return out.writer.Close()
}

func (out *Uploader) WroteCompressed() int64 {
	return out.n
}

func (out *Uploader) WroteUncompressed() int64 {
	return out.n
}

func (out *Uploader) Path() string {
	return out.key
}

func (out *Uploader) Cancel() error {
	return nil
}
