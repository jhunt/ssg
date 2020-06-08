package s3

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/jhunt/go-s3"

	"github.com/jhunt/shield-storage-gateway/pkg/rand"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
)

const RandomKey = ""

type Endpoint struct {
	URL             string
	Region          string
	Bucket          string
	Prefix          string
	UsePath         bool
	AccessKeyID     string
	SecretAccessKey string
}

type Provider struct {
	prefix string
	client *s3.Client
}

func Configure(e Endpoint) (Provider, error) {
	var host, scheme string

	if e.URL != "" {
		u, err := url.Parse(e.URL)
		if err != nil {
			return Provider{}, err
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return Provider{}, fmt.Errorf("invalid s3 base url '%s': no http/https scheme", e.URL)
		}
		scheme = u.Scheme
		host = u.Host
	}

	client, err := s3.NewClient(&s3.Client{
		Domain:          host,
		Protocol:        scheme,
		Bucket:          e.Bucket,
		Region:          e.Region,
		UsePathBuckets:  e.UsePath,
		AccessKeyID:     e.AccessKeyID,
		SecretAccessKey: e.SecretAccessKey,
	})
	if err != nil {
		return Provider{}, err
	}

	return Provider{
		prefix: e.Prefix,
		client: client,
	}, nil
}

func (p Provider) Upload(hint string) (provider.Uploader, error) {
	key := hint
	if key == RandomKey {
		key = rand.Path()
	}

	up, err := p.client.NewUpload(key, nil)
	if err != nil {
		return nil, err
	}

	return &Uploader{
		up:  up,
		key: key,
		buf: make([]byte, 5*1024*1024), // FIXME make this configurable
	}, nil
}

func (p Provider) Download(path string) (provider.Downloader, error) {
	dl, err := p.client.Get(path)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(dl), nil
}

func (p Provider) Expunge(path string) error {
	return p.client.Delete(path)
}
