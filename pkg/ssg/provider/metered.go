package provider

import (
	"io"

	"github.com/jhunt/ssg/pkg/meter"
)

type MeteredDownloader struct {
	rd *meter.Reader
}

func (m MeteredDownloader) Read(b []byte) (int, error) {
	return m.rd.Read(b)
}

func (m MeteredDownloader) Close() error {
	return m.rd.Close()
}

func (m MeteredDownloader) ReadCompressed() int64 {
	return m.rd.Total()
}

func (m MeteredDownloader) ReadUncompressed() int64 {
	return m.rd.Total()
}

func MeteredDownload(r io.ReadCloser) (MeteredDownloader, error) {
	return MeteredDownloader{
		rd: meter.NewReader(r),
	}, nil
}
