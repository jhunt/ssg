package s3

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/jhunt/go-s3"

	"github.com/jhunt/shield-storage-gateway/pkg/rand"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
)

const (
	RandomKey       = ""
	DefaultPartSize = 5
)

type Endpoint struct {
	URL             string
	Region          string
	Bucket          string
	Prefix          string
	UsePath         bool
	PartSize        int
	AccessKeyID     string
	SecretAccessKey string
}

type Provider struct {
	prefix   string
	client   *s3.Client
	partsize int
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

	if e.PartSize == 0 {
		e.PartSize = DefaultPartSize
	}

	return Provider{
		prefix:   e.Prefix,
		client:   client,
		partsize: e.PartSize,
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
		buf: make([]byte, p.partsize*1024*1024),
	}, nil
}

func (p Provider) Download(path string) (provider.Downloader, error) {
	get, err := p.client.Get(path)
	if err != nil {
		return nil, err
	}
	return provider.MeteredDownload(ioutil.NopCloser(get))
}

func (p Provider) Expunge(path string) error {
	return p.client.Delete(path)
}
