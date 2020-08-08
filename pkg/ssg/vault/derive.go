package vault

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

func (fks FixedKeySource) Derive(alg string, resolve func(string) ([]byte, error)) (Cipher, error) {
	if resolve == nil {
		resolve = func(in string) ([]byte, error) {
			return []byte(in), nil
		}
	}

	c := Cipher{Algorithm: alg}
	var key, salt []byte
	if fks.PBKDF2 != "" {
		k, err := resolve(fks.PBKDF2)
		if err != nil {
			return Cipher{}, err
		}
		if len(k) < 24 {
			return Cipher{}, fmt.Errorf("insufficient keying material provided for pbkdf2: only %d bytes found (need at least 24 bytes)", len(k))
		}
		key = k[len(k)/3:]
		salt = k[:len(k)/2]
	}

	algorithm, _ := parse(alg)
	switch algorithm {
	case "aes128":
		if key != nil && salt != nil {
			c.Key = pbkdf2.Key(key, salt, 4096, 16, sha256.New)
			c.IV = pbkdf2.Key(key, salt, 4096, aes.BlockSize, sha256.New)
			return c, nil
		}

		if fks.Literal.AES128.Key != "" && fks.Literal.AES128.IV != "" {
			keysz := 16
			key, err := resolve(fks.Literal.AES128.Key)
			if err != nil {
				return Cipher{}, err
			}
			iv, err := resolve(fks.Literal.AES128.IV)
			if err != nil {
				return Cipher{}, err
			}
			if len(key) != keysz {
				return Cipher{}, fmt.Errorf("insufficient key size (%d bytes): want exactly %d bytes", len(key), keysz)
			}
			if len(iv) != aes.BlockSize {
				return Cipher{}, fmt.Errorf("insufficient initialization vector size (%d bytes): want exactly %d bytes", len(iv), aes.BlockSize)
			}
			c.Key = key
			c.IV = iv
			return c, nil
		}

	case "aes192":
		if key != nil && salt != nil {
			c.Key = pbkdf2.Key(key, salt, 4096, 24, sha256.New)
			c.IV = pbkdf2.Key(key, salt, 4096, aes.BlockSize, sha256.New)
			return c, nil
		}

		if fks.Literal.AES192.Key != "" && fks.Literal.AES192.IV != "" {
			keysz := 24
			key, err := resolve(fks.Literal.AES192.Key)
			if err != nil {
				return Cipher{}, err
			}
			iv, err := resolve(fks.Literal.AES192.IV)
			if err != nil {
				return Cipher{}, err
			}
			if len(key) != keysz {
				return Cipher{}, fmt.Errorf("insufficient key size (%d bytes): want exactly %d bytes", len(key), keysz)
			}
			if len(iv) != aes.BlockSize {
				return Cipher{}, fmt.Errorf("insufficient initialization vector size (%d bytes): want exactly %d bytes", len(iv), aes.BlockSize)
			}
			c.Key = key
			c.IV = iv
			return c, nil
		}

	case "aes256":
		if key != nil && salt != nil {
			c.Key = pbkdf2.Key(key, salt, 4096, 32, sha256.New)
			c.IV = pbkdf2.Key(key, salt, 4096, aes.BlockSize, sha256.New)
			return c, nil
		}

		if fks.Literal.AES256.Key != "" && fks.Literal.AES256.IV != "" {
			keysz := 32
			key, err := resolve(fks.Literal.AES256.Key)
			if err != nil {
				return Cipher{}, err
			}
			iv, err := resolve(fks.Literal.AES256.IV)
			if err != nil {
				return Cipher{}, err
			}
			if len(key) != keysz {
				return Cipher{}, fmt.Errorf("insufficient key size (%d bytes): want exactly %d bytes", len(key), keysz)
			}
			if len(iv) != aes.BlockSize {
				return Cipher{}, fmt.Errorf("insufficient initialization vector size (%d bytes): want exactly %d bytes", len(iv), aes.BlockSize)
			}
			c.Key = key
			c.IV = iv
			return c, nil
		}

	default:
		return Cipher{}, fmt.Errorf("unrecognized encryption algorithm: '%s'", alg)
	}

	return Cipher{}, fmt.Errorf("unable to derive %s fixed cipher: no methods left to try", algorithm)
}

func (v Vault) Cipher(alg string) (Cipher, error) {
	if v.FixedKey.Enabled {
		return v.FixedKey.Derive(alg, v.Provider.FixedKeyResolver())
	}

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
