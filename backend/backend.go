package backend

import (
	"io"
)

type Backend interface {
	io.Writer
	io.Closer

	Retrieve() (io.ReadCloser, error)
	Cancel() error
}

type BackendBuilder func(string) Backend
