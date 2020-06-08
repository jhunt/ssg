package s3

import (
	"io/ioutil"

	"github.com/jhunt/go-s3"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
	"github.com/jhunt/shield-storage-gateway/pkg/rand"
)

const RandomKey = ""

type Endpoint struct {
	Region string
	Bucket string
	Prefix string
	AccessKeyID string
	SecretAccessKey string
}

type Provider struct {
	prefix string
	client *s3.Client
}

func Configure(e Endpoint) (Provider, error) {
	client, err := s3.NewClient(&s3.Client{
		Bucket: e.Bucket,
		Region: e.Region,
		AccessKeyID: e.AccessKeyID,
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
		up: up,
		key: key,
		n: 0,
		buf: make([]byte, 5*1024*1024),
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
