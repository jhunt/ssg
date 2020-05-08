package backend

import (
	"io"
	"os"
	"path"
)

type File struct {
	path string
}

func FileBuilder(root string) BackendBuilder {
	return func(path string) Backend {
		return &File{path}
	}
}

func (f *File) Write(b []byte) (int, error) {
	err := os.MkdirAll(path.Dir(f.path), 0777)
	if err != nil {
		return 0, err
	}

	out, err := os.OpenFile(f.path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return 0, err
	}
	size, err := out.Write(b)
	if err != nil {
		return 0, err
	}
	out.Close()
	return size, nil
}

func (f *File) Retrieve() (io.ReadCloser, error) {
	return os.Open(f.path)
}

func (f *File) Close() error {
	return nil
}

func (f *File) Cancel() error {
	err := os.Remove(f.path)
	if os.IsNotExist(err) {
		return nil // ENOENT is A-OKAY
	}
	return err
}
