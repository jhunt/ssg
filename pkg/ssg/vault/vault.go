package vault

import (
	"io"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
)

type Vault interface {
	Set(string, Cipher) error
	Get(string) (Cipher, error)
}

type EncryptedUploader struct {
	wr    io.WriteCloser
	inner provider.Uploader
}

func (e EncryptedUploader) Write(b []byte) (int, error) {
	return e.wr.Write(b)
}

func (e EncryptedUploader) Close() error {
	return e.wr.Close()
}

func (e EncryptedUploader) Path() string {
	return e.inner.Path()
}

func (e EncryptedUploader) Cancel() error {
	return e.inner.Cancel()
}

func Encrypt(v Vault, id, alg string, up provider.Uploader) (provider.Uploader, error) {
	c, err := NewCipher(alg)
	if err != nil {
		return nil, err
	}

	err = v.Set(id, c)
	if err != nil {
		return nil, err
	}

	wr, err := c.Encrypt(up)
	if err != nil {
		return nil, err
	}

	return EncryptedUploader{
		wr:    wr,
		inner: up,
	}, nil
}

type DecryptedDownloader struct {
	rd    io.Reader
	inner provider.Downloader
}

func (d DecryptedDownloader) Read(b []byte) (int, error) {
	return d.rd.Read(b)
}

func (d DecryptedDownloader) Close() error {
	return d.inner.Close()
}

func Decrypt(v Vault, id string, down provider.Downloader) (provider.Downloader, error) {
	c, err := v.Get(id)
	if err != nil {
		return nil, err
	}

	rd, err := c.Decrypt(down)
	if err != nil {
		return nil, err
	}

	return DecryptedDownloader{
		rd:    rd,
		inner: down,
	}, nil
}
