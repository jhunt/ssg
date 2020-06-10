package provider

import (
	"compress/zlib"
	"io"
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
	if err := z.w.Close(); err != nil {
		return err
	}
	// zlib.Writer's Close() does NOT close the underlying io.Writer...
	return z.inner.Close()
}

func (z *ZlibUploader) WroteCompressed() int64 {
	return z.inner.WroteCompressed()
}

func (z *ZlibUploader) WroteUncompressed() int64 {
	return z.n
}

func (z *ZlibUploader) Path() string {
	return z.inner.Path()
}

func (z *ZlibUploader) Cancel() error {
	return z.inner.Cancel()
}

type ZlibDownloader struct {
	r     io.ReadCloser
	inner Downloader
	n     int64
}

func (z *ZlibDownloader) Read(b []byte) (int, error) {
	n, err := z.r.Read(b)
	if err == nil || err == io.EOF {
		z.n += int64(n)
	}
	return n, err
}

func (z *ZlibDownloader) Close() error {
	return z.r.Close()
}

func (z *ZlibDownloader) ReadCompressed() int64 {
	return z.inner.ReadCompressed()
}

func (z *ZlibDownloader) ReadUncompressed() int64 {
	return z.n
}
