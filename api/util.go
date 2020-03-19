package api

import (
	"crypto/rand"
	"fmt"
	"io"
)

func NewRandomString(n uint) (string, error) {
	b := make([]byte, (n+1)/2)
	if nread, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	} else if nread != len(b) {
		return "", fmt.Errorf("short read from random source -- wanted %d, got %d bytes!", len(b), nread)
	}
	return fmt.Sprintf("%x", b), nil
}
