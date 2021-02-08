package vault

import (
	"io"

	"github.com/jhunt/ssg/pkg/ssg/provider"
)

type Vault struct {
	FixedKey FixedKeySource
	Provider VaultProvider
}

type VaultProvider interface {
	FixedKeyResolver() FixedKeyResolver
	SetCipher(string, Cipher) error
	GetCipher(string) (Cipher, error)
	Delete(string) error
}

type FixedKeyResolver func(in string) ([]byte, error)

var PassThroughResolver FixedKeyResolver = func(in string) ([]byte, error) {
	return []byte(in), nil
}

// FixedKeySource represents operator configuration for the
// source and derivation of a single fixed cipher's static
// key and initialization vector.
//
// We currently support the following methods:
//
//    PBKDF2   Password-Based Key Derivation; the operator
//             points us at a secret *in* the vault, and
//             we use that as an input to a deterministic
//             function for deriving key + iv.
//
//    Literal  The key and the iv, encoded as fixed-length
//             hexadecimal values, are to be found in the
//             vault at fixed locations.  Different paths
//             (ids) in the vault are used for different
//             algorithms.
//
type FixedKeySource struct {
	// Enabled turns on fixed key derivation.
	//
	Enabled bool

	// PBKDF2 is the id of a secret, stored in the provider
	// backend vault, from which we will derive encrpytion
	// parameters.
	//
	PBKDF2 string

	// Literal provides paths to algorithm-specific key and
	// initialization vector values stored in the vault.
	//
	Literal struct {
		AES128 struct {
			Key string
			IV  string
		}
		AES192 struct {
			Key string
			IV  string
		}
		AES256 struct {
			Key string
			IV  string
		}
	}
}

type EncryptedUploader struct {
	id    string
	v     Vault
	wr    io.WriteCloser
	inner provider.Uploader
}

func (e EncryptedUploader) Write(b []byte) (int, error) {
	return e.wr.Write(b)
}

func (e EncryptedUploader) Close() error {
	return e.wr.Close()
}

func (e EncryptedUploader) WroteCompressed() int64 {
	return e.inner.WroteCompressed()
}

func (e EncryptedUploader) WroteUncompressed() int64 {
	return e.inner.WroteUncompressed()
}

func (e EncryptedUploader) Path() string {
	return e.inner.Path()
}

func (e EncryptedUploader) Cancel() error {
	err := e.inner.Cancel()
	if err != nil {
		return err
	}
	return e.v.Provider.Delete(e.id)
}

func Encrypt(v Vault, id, alg string, up provider.Uploader) (provider.Uploader, error) {
	c, err := v.Cipher(alg)
	if err != nil {
		return nil, err
	}

	err = v.Provider.SetCipher(id, c)
	if err != nil {
		return nil, err
	}

	wr, err := c.Encrypt(up)
	if err != nil {
		return nil, err
	}

	return EncryptedUploader{
		id:    id,
		v:     v,
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

func (d DecryptedDownloader) ReadCompressed() int64 {
	return d.inner.ReadCompressed()
}

func (d DecryptedDownloader) ReadUncompressed() int64 {
	return d.inner.ReadUncompressed()
}

func Decrypt(v Vault, id string, down provider.Downloader) (provider.Downloader, error) {
	c, err := v.Provider.GetCipher(id)
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
