package ssg

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jhunt/go-log"

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
	log.Debugf(LOG+"looking for bucket '%s' (from url '%s')", to.Bucket, to)
	bucket := s.bucket(to.Bucket)
	if bucket == nil {
		return nil, "", fmt.Errorf("bucket '%s' not found", to.Bucket)
	}

	log.Debugf(LOG+"generating random path in bucket '%s'", to.Bucket)
	uploader, err := bucket.Upload("")
	if err != nil {
		return nil, "", err
	}
	to.Path = uploader.Path()

	log.Infof(LOG+"starting upload to %v", to)
	upstream := &stream{
		id:     rand.String(96),
		canon:  to.String(),
		reader: nil,
		writer: uploader,
	}
	upstream.lease(s.MaxLease)
	log.Debugf(LOG+"stream %v -> %v will be valid until %v", upstream.id, upstream.canon, upstream.expires)

	s.lock.Lock()
	defer s.lock.Unlock()
	s.uploads[upstream.id] = upstream
	return upstream, uploader.Path(), nil
}

func (s *Server) getUpload(id, token string) (*stream, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Debugf(LOG+"looking for upload stream %v", id)
	upstream, ok := s.uploads[id]
	if ok {
		log.Debugf(LOG + "stream found; validating against supplied token")
	} else {
		log.Debugf(LOG + "stream not found in server records")
	}
	return upstream, ok && upstream.authorize(token)
}

func (s *Server) startDownload(from *url.URL) (*stream, error) {
	log.Debugf(LOG+"looking for bucket '%s' (from url '%s')", from.Bucket, from)
	bucket := s.bucket(from.Bucket)
	if bucket == nil {
		return nil, fmt.Errorf("bucket '%s' not found", from.Bucket)
	}

	log.Infof(LOG+"starting download from %v", from)
	downloader, err := bucket.Download(from.Path)
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
	log.Debugf(LOG+"stream %v <- %v will be valid until %v", downstream.id, downstream.canon, downstream.expires)

	s.lock.Lock()
	defer s.lock.Unlock()
	s.downloads[downstream.id] = downstream
	return downstream, nil
}

func (s *Server) getDownload(id, token string) (*stream, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Debugf(LOG+"looking for download stream %v", id)
	downstream, ok := s.downloads[id]
	if ok {
		log.Debugf(LOG + "stream found; validating against supplied token")
	} else {
		log.Debugf(LOG + "stream not found in server records")
	}
	return downstream, ok && downstream.authorize(token)
}

func (s *Server) forget(x *stream) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.uploads[x.id]; ok {
		log.Debugf(LOG+"forgeting upload stream %v", x.id)
		delete(s.uploads, x.id)
	}

	if _, ok := s.downloads[x.id]; ok {
		log.Debugf(LOG+"forgeting download stream %v", x.id)
		delete(s.downloads, x.id)
	}
}

func (s *Server) expunge(where *url.URL) error {
	log.Debugf(LOG+"looking for bucket '%s' (from url '%s')", where.Bucket, where)
	bucket := s.bucket(where.Bucket)
	if bucket == nil {
		return fmt.Errorf("bucket '%s' not found", where.Bucket)
	}

	log.Infof(LOG+"expunging %v", where)
	return bucket.Expunge(where.Path)
}

func (s *Server) Run(helo string) error {
	http.Handle("/", s.Router(helo))
	go s.Sweep()

	log.Infof(LOG+"http server starting up on %s", s.Bind)
	if err := http.ListenAndServe(s.Bind, nil); err != nil {
		return err
	}
	log.Infof(LOG + "http server shutting down")
	return nil
}

func NewServerFromFile(path string) (*Server, error) {
	cfg, err := config.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewServer(cfg)
}

func NewServer(c config.Config) (*Server, error) {
	var s Server
	s.uploads = make(map[string]*stream)
	s.downloads = make(map[string]*stream)

	s.Cluster = c.Cluster
	log.Infof(LOG+"set cluster identity to %v", s.Bind)

	s.Bind = c.Bind
	log.Infof(LOG+"set bind address to %v", s.Bind)

	s.ControlTokens = make([]string, len(c.ControlTokens))
	copy(s.ControlTokens, c.ControlTokens)
	log.Infof(LOG+"authorized %d control tokens", len(s.ControlTokens))

	s.MonitorTokens = make([]string, len(c.MonitorTokens))
	copy(s.MonitorTokens, c.MonitorTokens)
	log.Infof(LOG+"authorized %d monitor tokens", len(s.MonitorTokens))

	s.MaxLease = time.Duration(c.MaxLease) * time.Second
	log.Infof(LOG+"set maximum stream lease to %d seconds", c.MaxLease)

	s.SweepInterval = time.Duration(c.SweepInterval) * time.Second
	log.Infof(LOG+"set stream sweep interval to %d seconds", c.MaxLease)

	s.buckets = make([]*bucket, len(c.Buckets))
	for i, b := range c.Buckets {
		var p provider.Provider
		switch b.Provider.Kind {
		case "fs":
			log.Infof(LOG+"configuring bucket %v backed by fs (root=%v)", b.Key, b.Provider.FS.Root)
			candidate, err := fs.Configure(b.Provider.FS.Root)
			if err != nil {
				return nil, fmt.Errorf("fs bucket %v could not be configured: %s", b.Key, err)
			}
			p = candidate

		case "s3":
			log.Infof(LOG+"configuring bucket %v backed by s3 (region=%v, bucket=%v, prefix=%v)", b.Key, b.Provider.S3.Region, b.Provider.S3.Bucket, b.Provider.S3.Prefix)
			candidate, err := s3.Configure(s3.Endpoint{
				Prefix:          b.Provider.S3.Prefix,
				Region:          b.Provider.S3.Region,
				Bucket:          b.Provider.S3.Bucket,
				AccessKeyID:     b.Provider.S3.AccessKeyID,
				SecretAccessKey: b.Provider.S3.SecretAccessKey,
			})
			if err != nil {
				return nil, fmt.Errorf("s3 bucket %v could not be configured: %s", b.Key, err)
			}
			p = candidate

		case "webdav":
			log.Infof(LOG+"configuring bucket %v backed by webdav (url=%v)", b.Key, b.Provider.WebDAV.URL)
			candidate, err := webdav.Configure(webdav.Endpoint{
				URL:      b.Provider.WebDAV.URL,
				Username: b.Provider.WebDAV.BasicAuth.Username,
				Password: b.Provider.WebDAV.BasicAuth.Password,
			})
			if err != nil {
				return nil, fmt.Errorf("webdav bucket %v could not be configured: %s", b.Key, err)
			}
			p = candidate

		default:
			return nil, fmt.Errorf("unrecognized provider for bucket %v: '%s'", b.Key, b.Provider.Kind)
		}

		var v vault.Vault
		if b.Vault == nil {
			v = vault.Nil
		} else {
			switch b.Vault.Kind {
			case "hashicorp":
				log.Infof(LOG+"configuring bucket %v vault backed by hashicorp vault (url=%v, prefix=%v)", b.Key, b.Vault.Hashicorp.URL, b.Vault.Hashicorp.Prefix)
				candidate, err := hashicorp.Configure(hashicorp.Endpoint{
					Prefix: b.Vault.Hashicorp.Prefix,
					URL:    b.Vault.Hashicorp.URL,
					Token:  b.Vault.Hashicorp.Token,
				})
				if err != nil {
					return nil, fmt.Errorf("bucket %v hashicorp vault could not be configured: %s", b.Key, err)
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
	log.Infof(LOG+"configured %d buckets", len(s.buckets))

	return &s, nil
}
