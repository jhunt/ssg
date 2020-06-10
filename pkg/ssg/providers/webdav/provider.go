package webdav

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/jhunt/shield-storage-gateway/pkg/rand"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
)

const RandomFile = ""

type Endpoint struct {
	URL      string
	Username string
	Password string
}

type Provider struct {
	base     *url.URL
	username string
	password string
	client   *http.Client
}

func Configure(e Endpoint) (Provider, error) {
	base, err := url.Parse(e.URL)
	if err != nil {
		return Provider{}, err
	}

	return Provider{
		base:     base,
		username: e.Username,
		password: e.Password,
		client:   http.DefaultClient,
	}, nil
}

func (p Provider) Upload(hint string) (provider.Uploader, error) {
	relpath := hint
	if relpath == "" {
		relpath = rand.Path()
	}

	for _, prefix := range ancestors(relpath) {
		req, err := http.NewRequest("MKCOL", p.url(prefix)+"/", nil)
		if err != nil {
			return nil, fmt.Errorf("MKCOL %s: %s", prefix, err)
		}
		res, err := p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("MKCOL %s: %s", prefix, err)
		}
		res.Body.Close()
		if res.StatusCode != 201 && res.StatusCode != 405 {
			return nil, fmt.Errorf("MKCOL %s: HTTP %s", prefix, res.Status)
		}
	}

	rd, wr := io.Pipe()
	req, err := http.NewRequest("PUT", p.url(relpath), rd)
	if err != nil {
		return nil, err
	}
	done := make(chan int)
	go func() {
		res, err := p.client.Do(req)
		if err == nil {
			res.Body.Close()
		}
		done <- 1
	}()
	return &Uploader{
		writer:  wr,
		relpath: relpath,
		done:    done,
	}, nil
}

func (p Provider) Download(path string) (provider.Downloader, error) {
	req, err := http.NewRequest("GET", p.url(path), nil)
	if err != nil {
		return nil, err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %s", req.URL, res.Status)
	}
	return provider.MeteredDownload(res.Body)
}

func (p Provider) Expunge(path string) error {
	req, err := http.NewRequest("DELETE", p.url(path), nil)
	if err != nil {
		return err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != 410 && res.StatusCode != 200 && res.StatusCode != 204 {
		return fmt.Errorf("%s: HTTP %s", req.URL, res.Status)
	}
	return nil
}

func (p Provider) url(rel string) string {
	u, _ := url.Parse(p.base.String())
	u.Path = filepath.Join(u.Path, path.Clean(rel))
	return u.String()
}

func ancestors(dir string) []string {
	parts := strings.Split(strings.TrimPrefix(path.Clean(dir), "/"), "/")
	if len(parts) < 2 {
		return []string{}
	}
	l := make([]string, len(parts)-1)
	for i := range parts[1:] {
		l[i] = filepath.Join(parts[:i+1]...)
	}
	return l
}
