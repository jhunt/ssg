package ssg

import (
	"github.com/jhunt/go-log"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/vault"
)

func (b *bucket) Upload(s string) (provider.Uploader, error) {
	uploader, err := b.provider.Upload(s)
	if err != nil {
		return nil, err
	}

	log.Debugf("bucket.Upload(%v): encryption is '%s'", s, b.encryption)
	if b.encryption != "none" {
		log.Debugf("bucket.Upload(%v): wrapping uploader with encrypting stream", s)
		uploader, err = vault.Encrypt(b.vault, uploader.Path(), b.encryption, uploader)
		if err != nil {
			return nil, err
		}
	}

	log.Debugf("bucket.Upload(%v): compression is '%s'", s, b.compression)
	if b.compression != "none" {
		log.Debugf("bucket.Upload(%v): wrapping uploader with compressing stream", s)
		uploader, err = provider.Compress(uploader, b.compression)
		if err != nil {
			return nil, err
		}
	}

	return uploader, nil
}

func (b *bucket) Download(s string) (provider.Downloader, error) {
	downloader, err := b.provider.Download(s)
	if err != nil {
		return nil, err
	}

	if b.encryption != "none" {
		downloader, err = vault.Decrypt(b.vault, s, downloader)
		if err != nil {
			return nil, err
		}
	}

	downloader, err = provider.Decompress(downloader, b.compression)
	if err != nil {
		return nil, err
	}

	return downloader, nil
}

func (b *bucket) Expunge(s string) error {
	if b.encryption != "none" {
		if err := b.vault.Delete(s); err != nil {
			return err
		}
	}
	return b.provider.Expunge(s)
}
