package provider

import (
	"compress/zlib"
)

type ZlibUploader struct {
	w     *zlib.Writer
	inner Uploader
	n     int64
}

func (z *ZlibUploader) Write(b []byte) (int, error) {
	n, err := z.w.Write(b)
	if err != nil {
		return n, err
	}
	z.n += int64(n)
	return n, nil
}

func (z *ZlibUploader) Close() error {
	return z.w.Close()
}

func (z *ZlibUploader) SentCompressed() int64 {
	return z.inner.SentCompressed()
}

func (z *ZlibUploader) SentUncompressed() int64 {
	return z.n
}

func (z *ZlibUploader) Path() string {
	return z.inner.Path()
}

func (z *ZlibUploader) Cancel() error {
	return z.inner.Cancel()
}
