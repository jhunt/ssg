package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
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
