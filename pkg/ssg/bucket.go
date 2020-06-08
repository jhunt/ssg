package ssg

import (
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/vault"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
)

func (b *bucket) Upload(s string) (provider.Uploader, error) {
	uploader, err := b.provider.Upload(s)
	if err != nil {
		return nil, err
	}

	if b.encryption != "none" {
		uploader, err = vault.Encrypt(b.vault, uploader.Path(), b.encryption, uploader)
		if err != nil {
			return nil, err
		}
	}

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

func (b *bucket) encrypt(up provider.Uploader) (provider.Uploader, error) {
	return up, nil
}

func (b *bucket) decrypt(down provider.Downloader) (provider.Downloader, error) {
	return down, nil
}
