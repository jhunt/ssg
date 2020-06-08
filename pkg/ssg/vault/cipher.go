package vault

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/cipher"
	"io"
	"fmt"
	"strings"
)

type Cipher struct {
	Algorithm string
	Key       []byte
	IV        []byte
}

func parse(in string) (string, string) {
	l := strings.Split(in, "-")
	if len(l) != 2 {
		return "", ""
	}
	return l[0], l[1]
}

func NewCipher(alg string) (Cipher, error) {
	c := Cipher{Algorithm: alg}

	algorithm, _ := parse(alg)
	switch algorithm {
	case "aes128":
		c.Key = make([]byte, 16)
		c.IV = make([]byte, aes.BlockSize)

	case "aes192":
		c.Key = make([]byte, 24)
		c.IV = make([]byte, aes.BlockSize)

	case "aes256":
		c.Key = make([]byte, 32)
		c.IV = make([]byte, aes.BlockSize)

	default:
		return Cipher{}, fmt.Errorf("unrecognized encryption algorithm: '%s'", alg)
	}

	if _, err := rand.Read(c.Key); err != nil {
		return Cipher{}, fmt.Errorf("failed to generate %s encryption key: %s", alg, err)
	}
	if _, err := rand.Read(c.IV); err != nil {
		return Cipher{}, fmt.Errorf("failed to generate %s initialization vector: %s", alg, err)
	}

	return c, nil
}

func (c Cipher) stream() (cipher.Stream, cipher.Stream, error) {
	algo, mode := parse(c.Algorithm)
	switch algo {
	case "ases128", "aes192", "aes256":
		block, err := aes.NewCipher(c.Key)
		if err != nil {
			return nil, nil, err
		}

		switch mode {
		case "cfb":
			return cipher.NewCFBEncrypter(block, c.IV), cipher.NewCFBDecrypter(block, c.IV), nil
		case "ofb":
			return cipher.NewOFB(block, c.IV), cipher.NewOFB(block, c.IV), nil
		case "ctr":
			return cipher.NewCTR(block, c.IV), cipher.NewCTR(block, c.IV), nil
		}
	}

	return nil, nil, fmt.Errorf("unrecognized encryption algorithm: '%s'", c.Algorithm)
}

func (c Cipher) Encrypt(wr io.Writer) (io.WriteCloser, error) {
	e, _, err := c.stream()
	if err != nil {
		return nil, err
	}

	return cipher.StreamWriter{
		S: e,
		W: wr,
	}, nil
}

func (c Cipher) Decrypt(rd io.Reader) (io.Reader, error) {
	_, d, err := c.stream()
	if err != nil {
		return nil, err
	}

	return cipher.StreamReader{
		S: d,
		R: rd,
	}, nil
}
