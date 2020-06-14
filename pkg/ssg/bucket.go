package ssg

import (
	"github.com/jhunt/go-log"

	"github.com/jhunt/ssg/pkg/ssg/provider"
	"github.com/jhunt/ssg/pkg/ssg/vault"
)

func (b *bucket) Upload(s string) (provider.Uploader, error) {
	uploader, err := b.provider.Upload(s)
	if err != nil {
		return nil, err
	}

	log.Debugf(LOG+"blobs in bucket %v use encryption algorithm %v", b.key, b.encryption)
	if b.encryption != "none" {
		uploader, err = vault.Encrypt(b.vault, uploader.Path(), b.encryption, uploader)
		if err != nil {
			return nil, err
		}
	}

	log.Debugf(LOG+"blobs in bucket %v use compression algorithm %v", b.key, b.compression)
	uploader, err = provider.Compress(uploader, b.compression)
	if err != nil {
		return nil, err
	}

	return uploader, nil
}

func (b *bucket) Download(s string) (provider.Downloader, error) {
	downloader, err := b.provider.Download(s)
	if err != nil {
		return nil, err
	}

	log.Debugf(LOG+"blobs in bucket %v use encryption algorithm %v", b.key, b.encryption)
	if b.encryption != "none" {
		downloader, err = vault.Decrypt(b.vault, s, downloader)
		if err != nil {
			return nil, err
		}
	}

	log.Debugf(LOG+"blobs in bucket %v use compression algorithm %v", b.key, b.compression)
	downloader, err = provider.Decompress(downloader, b.compression)
	if err != nil {
		return nil, err
	}

	return downloader, nil
}

func (b *bucket) Expunge(s string) error {
	log.Debugf(LOG+"expunging %s from bucket", s)
	if b.encryption != "none" {
		log.Debugf(LOG+"blobs in bucket %v are encrypted; removing cipher parameters from vault", b.key)
		if err := b.vault.Provider.Delete(s); err != nil {
			return err
		}
	}
	return b.provider.Expunge(s)
}
