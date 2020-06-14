package gcs

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"

	"github.com/jhunt/ssg/pkg/rand"
	"github.com/jhunt/ssg/pkg/ssg/provider"
)

const RandomKey = ""

type Endpoint struct {
	Key    interface{}
	Bucket string
	Prefix string
}

type Provider struct {
	svc    *storage.Service
	bucket string
	prefix string
}

func Configure(e Endpoint) (Provider, error) {
	var c *http.Client

	scope := storage.DevstorageFullControlScope
	ctx := context.Background()
	if e.Key != nil {
		b, err := json.Marshal(j2y(e.Key))
		if err != nil {
			return Provider{}, err
		}

		conf, err := google.JWTConfigFromJSON(b, scope)
		if err != nil {
			return Provider{}, err
		}

		c = conf.Client(ctx)

	} else {
		maybe, err := google.DefaultClient(ctx, scope)
		if err != nil {
			return Provider{}, err
		}
		c = maybe
	}

	svc, err := storage.New(c)
	if err != nil {
		return Provider{}, err
	}

	return Provider{
		svc:    svc,
		bucket: e.Bucket,
		prefix: e.Prefix,
	}, nil
}

func (p Provider) Upload(hint string) (provider.Uploader, error) {
	key := hint
	if key == RandomKey {
		key = rand.Path()
	}
	key = p.prefix + key

	return &Uploader{
		key:    key,
		object: p.svc.Objects.Insert(p.bucket, &storage.Object{Name: key}),
	}, nil
}

func (p Provider) Download(path string) (provider.Downloader, error) {
	res, err := p.svc.Objects.Get(p.bucket, path).Download()
	if err != nil {
		return nil, err
	}

	return provider.MeteredDownload(res.Body)
}

func (p Provider) Expunge(path string) error {
	return p.svc.Objects.Delete(p.bucket, path).Do()
}
