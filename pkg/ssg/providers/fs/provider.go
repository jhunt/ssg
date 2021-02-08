package fs

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/jhunt/ssg/pkg/rand"
	"github.com/jhunt/ssg/pkg/ssg/provider"
)

const RandomFile = ""

type Provider struct {
	Root string
}

func Configure(root string) (Provider, error) {
	root = filepath.Clean(root)
	st, err := os.Stat(root)
	if err != nil {
		return Provider{}, fmt.Errorf("%s: %s", root, err)
	}
	if !st.IsDir() {
		return Provider{}, fmt.Errorf("%s: not a directory", root)
	}
	return Provider{
		Root: root,
	}, nil
}

func (f Provider) Upload(relpath string) (provider.Uploader, error) {
	if relpath == RandomFile {
		relpath = rand.Path()
		for {
			if _, err := os.Stat(filepath.Join(f.Root, relpath)); err != nil {
				break
			}
			relpath = rand.Path()
		}
	}
	relpath = filepath.Clean(relpath)
	abspath := filepath.Join(f.Root, relpath)

	if err := os.MkdirAll(path.Dir(abspath), 0777); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(abspath, os.O_CREATE|os.O_RDWR|os.O_TRUNC|os.O_EXCL, 0666)
	if err != nil {
		return nil, err
	}

	return &Uploader{
		file:    file,
		relpath: relpath,
		abspath: abspath,
	}, nil
}

func (f Provider) Download(relpath string) (provider.Downloader, error) {
	if relpath == "" {
		return nil, fmt.Errorf("no file specified")
	}

	relpath = filepath.Clean(relpath)
	file, err := os.OpenFile(filepath.Join(f.Root, relpath), os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	st, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("%s: not a regular file", relpath)
	}

	return provider.MeteredDownload(file)
}

func (f Provider) Expunge(relpath string) error {
	return os.Remove(filepath.Join(f.Root, filepath.Clean(relpath)))
}
