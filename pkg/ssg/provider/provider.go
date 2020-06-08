package provider

import (
	"io"
)

type Provider interface {
	Upload(string) (Uploader, error)
	Download(string) (Downloader, error)
	Expunge(string) error
}

type Uploader interface {
	io.Writer
	io.Closer
	Path() string
	Cancel() error
}

type Downloader interface {
	io.Reader
	io.Closer
}
