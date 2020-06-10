package provider

import (
	"fmt"

	"compress/zlib"

	"github.com/jhunt/ssg/pkg/meter"
)

func Compress(ul Uploader, alg string) (Uploader, error) {
	switch alg {
	case "none", "":
		return ul, nil
	case "zlib":
		return &ZlibUploader{
			w:     zlib.NewWriter(ul),
			inner: ul,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported compression algorithem: '%s'", alg)
	}
}

func Decompress(dl Downloader, alg string) (Downloader, error) {
	switch alg {
	case "none", "":
		return dl, nil
	case "zlib":
		zr, err := zlib.NewReader(dl)
		if err != nil {
			return nil, err
		}
		return &ZlibDownloader{
			r:     meter.NewReader(zr),
			inner: dl,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported compression algorithem: '%s'", alg)
	}
}
