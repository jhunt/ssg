package s3

import (
	"github.com/jhunt/go-s3"
)

type Uploader struct {
	key string
	up  *s3.Upload
	n   int64

	bufn int
	buf  []byte
}

func (out *Uploader) Write(b []byte) (int, error) {
	left := len(out.buf) - out.bufn
	nwrit := 0
	for len(b) >= left {
		copy(out.buf[out.bufn:], b[:left])
		b = b[left:]

		if err := out.up.Write(out.buf); err != nil {
			return nwrit, err
		}

		out.n += int64(nwrit)
		nwrit += len(out.buf)
		left = len(out.buf)
		out.bufn = 0
	}

	copy(out.buf[out.bufn:], b)
	out.bufn += len(b)
	nwrit += len(b)

	return nwrit, nil
}

func (out *Uploader) Close() error {
	if out.bufn > 0 {
		if err := out.up.Write(out.buf[:out.bufn]); err != nil {
			return err
		}
		out.n += int64(out.bufn)
	}
	return out.up.Done()
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
