package compress

import (
	"compress/zlib"
	"fmt"
	"io"
)

type Compression string

const ZlibCompression Compression = "zlib"

func Compress(out io.Writer, compressionType string) (io.WriteCloser, error) {
	switch compressionType {
	case string(ZlibCompression):
		w := zlib.NewWriter(out)
		return w, nil
	default:
		return nil, fmt.Errorf("unsupported compression scheme %s", compressionType)
	}
}

func Decompress(in io.ReadCloser, compressionType string) (io.ReadCloser, error) {
	switch compressionType {
	case string(ZlibCompression):
		return zlib.NewReader(in)
	default:
		return nil, fmt.Errorf("unsupported compression scheme %s", compressionType)
	}
}
