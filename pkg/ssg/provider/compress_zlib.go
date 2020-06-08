package provider

import (
	"compress/zlib"
)

type ZlibUploader struct {
	w *zlib.Writer
	inner Uploader
}

func (z ZlibUploader) Write(b []byte) (int, error) {
	return z.w.Write(b)
}

func (z ZlibUploader) Close() error {
	return z.w.Close()
}

func (z ZlibUploader) Path() string {
	return z.inner.Path()
}

func (z ZlibUploader) Cancel() error {
	return z.inner.Cancel()
}
