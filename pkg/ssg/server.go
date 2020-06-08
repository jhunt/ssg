package ssg

import (
	"fmt"
	"time"

	"github.com/jhunt/shield-storage-gateway/pkg/rand"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/config"
	"github.com/jhunt/shield-storage-gateway/pkg/url"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/providers/fs"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/providers/s3"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/providers/webdav"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg/vault"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/vaults/hashicorp"
)

func (s *Server) startUpload(to *url.URL) (*stream, string, error) {
	bucket := s.bucket(to.Bucket)
	if bucket == nil {
		return nil, "", fmt.Errorf("bucket '%s' not found", to.Bucket)
	}

	uploader, err := bucket.provider.Upload("")
	if err != nil {
		return nil, "", err
	}
	to.Path = uploader.Path()

	upstream := &stream{
		id:     rand.String(96),
		canon:  to.String(),
		reader: nil,
		writer: uploader,
	}
	upstream.lease(s.MaxLease)

	s.lock.Lock()
	defer s.lock.Unlock()
	s.uploads[upstream.id] = upstream
	return upstream, uploader.Path(), nil
}

func (s *Server) getUpload(id, token string) (*stream, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	upstream, ok := s.uploads[id]
	return upstream, ok && upstream.authorize(token)
}

func (s *Server) startDownload(from *url.URL) (*stream, error) {
	bucket := s.bucket(from.Bucket)
	if bucket == nil {
		return nil, fmt.Errorf("bucket '%s' not found", from.Bucket)
	}

	downloader, err := bucket.provider.Download(from.Path)
	if err != nil {
		return nil, err
	}

	downstream := &stream{
		id:     rand.String(96),
		canon:  from.String(),
		reader: downloader,
		writer: nil,
	}
	downstream.lease(s.MaxLease)

	s.lock.Lock()
	defer s.lock.Unlock()
	s.downloads[downstream.id] = downstream
	return downstream, nil
}

func (s *Server) getDownload(id, token string) (*stream, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	downstream, ok := s.downloads[id]
	return downstream, ok && downstream.authorize(token)
}

func (s *Server) forget(x *stream) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.uploads[x.id]; ok {
		delete(s.uploads, x.id)
	}

	if _, ok := s.downloads[x.id]; ok {
		delete(s.downloads, x.id)
	}
}

func (s *Server) expunge(where *url.URL) error {
	bucket := s.bucket(where.Bucket)
	if bucket == nil {
		return fmt.Errorf("bucket '%s' not found", where.Bucket)
	}

	return bucket.provider.Expunge(where.Path)
}

func NewServer(c config.Config) (*Server, error) {
	var s Server
	s.uploads = make(map[string]*stream)
	s.downloads = make(map[string]*stream)

	s.Cluster = c.Cluster
	s.Bind = c.Bind

	s.ControlTokens = make([]string, len(c.ControlTokens))
	copy(s.ControlTokens, c.ControlTokens)

	s.MonitorTokens = make([]string, len(c.MonitorTokens))
	copy(s.MonitorTokens, c.MonitorTokens)

	s.MaxLease = time.Duration(c.MaxLease) * time.Second
	s.SweepInterval = time.Duration(c.SweepInterval) * time.Second

	s.buckets = make([]*bucket, len(c.Buckets))
	for i, b := range c.Buckets {
		var p provider.Provider
		switch b.Provider.Kind {
		case "fs":
			candidate, err := fs.Configure(b.Provider.FS.Root)
			if err != nil {
				return nil, err
			}
			p = candidate

		case "s3":
			candidate, err := s3.Configure(s3.Endpoint{
				Prefix:          b.Provider.S3.Prefix,
				Region:          b.Provider.S3.Region,
				Bucket:          b.Provider.S3.Bucket,
				AccessKeyID:     b.Provider.S3.AccessKeyID,
				SecretAccessKey: b.Provider.S3.SecretAccessKey,
			})
			if err != nil {
				return nil, err
			}
			p = candidate

		case "webdav":
			candidate, err := webdav.Configure(webdav.Endpoint{
				URL:      b.Provider.WebDAV.URL,
				Username: b.Provider.WebDAV.BasicAuth.Username,
				Password: b.Provider.WebDAV.BasicAuth.Password,
			})
			if err != nil {
				return nil, err
			}
			p = candidate

		default:
			return nil, fmt.Errorf("unrecognized bucket provider: '%s'", b.Provider.Kind)
		}

		var v vault.Vault
		if b.Vault == nil {
			v = vault.Nil
		} else {
			switch b.Vault.Kind {
			case "hashicorp":
				candidate, err := hashicorp.Configure(hashicorp.Endpoint{
					Prefix: b.Vault.Hashicorp.Prefix,
					URL:    b.Vault.Hashicorp.URL,
					Token:  b.Vault.Hashicorp.Token,
				})
				if err != nil {
					return nil, err
				}
				v = candidate
			}
		}

		s.buckets[i] = &bucket{
			key:         b.Key,
			name:        b.Name,
			description: b.Description,
			compression: b.Compression,
			encryption:  b.Encryption,
			provider:    p,
			vault:       v,
		}
	}

	return &s, nil
}
