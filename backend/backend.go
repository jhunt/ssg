package backend

import (
	"compress/zlib"
	"fmt"
	"io"
)

type Compression string

const ZlibCompression Compression = "zlib"

type Backend interface {
	io.Writer
	io.Closer

	Retrieve() (io.ReadCloser, error)
	Cancel() error
}

func Compress(out io.Writer, b []byte, compressionType string) (int, error) {
	switch compressionType {
	case string(ZlibCompression):
		w := zlib.NewWriter(out)
		size, err := w.Write(b)
		if err != nil {
			return 0, err
		}
		w.Close()
		return size, nil
	default:
		return 0, fmt.Errorf("unsupported compression scheme %s", compressionType)
	}
}

func Decompress(in io.Reader, compressionType string) (io.ReadCloser, error) {
	switch compressionType {
	case string(ZlibCompression):
		return zlib.NewReader(in)
	default:
		return nil, fmt.Errorf("unsupported compression scheme %s", compressionType)
	}
}

type BackendBuilder func(string) Backend
