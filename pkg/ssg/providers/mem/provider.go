package mem

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/jhunt/ssg/pkg/rand"
	"github.com/jhunt/ssg/pkg/ssg/provider"
)

var RandomFile = ""

type Provider struct {
	Files map[string]*bytes.Buffer
}

func Configure() (Provider, error) {
	return Provider{
		Files: make(map[string]*bytes.Buffer),
	}, nil
}

func (f Provider) Upload(path string) (provider.Uploader, error) {
	if path == RandomFile {
		path = rand.Path()
	}

	if _, exists := f.Files[path]; exists {
		return nil, fmt.Errorf("%s: already exists", path)
	}

	f.Files[path] = bytes.NewBuffer(nil)
	return &Uploader{
		buf:  f.Files[path],
		path: path,
		n:    0,
		cancel: func() {
			delete(f.Files, path)
		},
	}, nil
}

func (f Provider) Download(path string) (provider.Downloader, error) {
	if path == "" {
		return nil, fmt.Errorf("no path specified")
	}

	buf, ok := f.Files[path]
	if !ok {
		return nil, fmt.Errorf("%s: not found", path)
	}
	return provider.MeteredDownload(ioutil.NopCloser(buf))
}

func (f Provider) Expunge(path string) error {
	delete(f.Files, path)
	return nil
}
