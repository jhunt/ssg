package s3

import (
	"github.com/jhunt/go-s3"
)

type Uploader struct {
	key string
	up *s3.Upload

	n int
	buf []byte
}

func (out *Uploader) Write(b []byte) (int, error) {
	left := len(out.buf) - out.n
	nwrit := 0
	for len(b) >= left {
		copy(out.buf[out.n:], b[:left])
		b = b[left:]

		if err := out.up.Write(out.buf); err != nil {
			return nwrit, err
		}

		nwrit += len(out.buf)
		left = len(out.buf)
		out.n = 0
	}

	copy(out.buf[out.n:], b)
	out.n += len(b)
	nwrit += len(b)

	return nwrit, nil
}

func (out *Uploader) Close() error {
	if out.n > 0 {
		if err := out.up.Write(out.buf[:out.n]); err != nil {
			return err
		}
	}
	return out.up.Done()
}

func (out *Uploader) Path() string {
	return out.key
}

func (out *Uploader) Cancel() error {
	return nil
}
