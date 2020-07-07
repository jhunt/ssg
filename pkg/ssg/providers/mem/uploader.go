package mem

import (
	"bytes"
)

type Uploader struct {
	buf    *bytes.Buffer
	path   string
	n      int64
	cancel func()
}

func (out *Uploader) Write(b []byte) (int, error) {
	n, err := out.buf.Write(b)
	if err != nil {
		return n, err
	}
	out.n += int64(n)
	return n, nil
}

func (out *Uploader) Close() error {
	return nil
}

func (out *Uploader) WroteCompressed() int64 {
	return out.n
}

func (out *Uploader) WroteUncompressed() int64 {
	return out.n
}

func (out *Uploader) Path() string {
	return out.path
}

func (out *Uploader) Cancel() error {
	out.cancel()
	return nil
}
