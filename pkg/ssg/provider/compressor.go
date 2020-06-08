package provider

import (
	"fmt"

	"compress/zlib"
)

func Compress(ul Uploader, alg string) (Uploader, error) {
	switch alg {
	case "none", "":
			return ul, nil
	case "zlib":
		return &ZlibUploader{
			w: zlib.NewWriter(ul),
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
		return zlib.NewReader(dl)
	default:
		return nil, fmt.Errorf("unsupported compression algorithem: '%s'", alg)
	}
}
