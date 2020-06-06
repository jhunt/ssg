package crypt

import (
	"crypto/cipher"
	"encoding/hex"
	"io"
	"strings"

	"github.com/jhunt/shield-storage-gateway/vault"
)

func Encrypt(out io.Writer, key, IV, enctype string) (cipher.StreamWriter, error) {
	var encStream cipher.Stream

	keyRaw, err := hex.DecodeString(strings.Replace(key, "-", "", -1))
	if err != nil {
		return cipher.StreamWriter{}, err
	}
	ivRaw, err := hex.DecodeString(strings.Replace(IV, "-", "", -1))
	if err != nil {
		return cipher.StreamWriter{}, err
	}

	encStream, _, err = vault.Stream(enctype, []byte(keyRaw), []byte(ivRaw))
	if err != nil {
		return cipher.StreamWriter{}, err
	}

	encrypter := cipher.StreamWriter{
		S: encStream,
		W: out,
	}
	return encrypter, nil
}

func Decrypt(in io.Reader, key, IV, enctype string) (cipher.StreamReader, error) {
	var decStream cipher.Stream

	keyRaw, err := hex.DecodeString(strings.Replace(key, "-", "", -1))
	if err != nil {
		return cipher.StreamReader{}, err
	}
	ivRaw, err := hex.DecodeString(strings.Replace(IV, "-", "", -1))
	if err != nil {
		return cipher.StreamReader{}, err
	}

	_, decStream, err = vault.Stream(enctype, []byte(keyRaw), []byte(ivRaw))
	if err != nil {
		return cipher.StreamReader{}, err
	}

	decrypter := cipher.StreamReader{
		S: decStream,
		R: in,
	}
	return decrypter, nil
}
