package api

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

const (
	StreamIDLength = 96
	FileNameLength = 64
)

type Disposition int

type Stream struct {
	ID       string
	Token    Token
	Path     string
	Received uint64

}

func NewStream(path string, lifetime time.Duration) (Stream, error) {
	id, err := NewRandomString(StreamIDLength)
	if err != nil {
		return Stream{}, err
	}

	tok, err := NewToken(lifetime)
	if err != nil {
		return Stream{}, err
	}

	return Stream{
		ID:    id,
		Token: tok,
		Path:  path,
	}, nil
}

func (s Stream) Authorize(token string) bool {
	return s.Token.Secret == token && !s.Token.Expired()
}

func (s Stream) Expired() bool {
	return s.Token.Expired()
}

func (s Stream) Undo() error {
	err := os.Remove(s.Path)
	if os.IsNotExist(err) {
		return nil // ENOENT is A-OKAY
	}
	return err
}

func (s *Stream) UploadChunk(token string, b []byte) (int, error) {
	if !s.Authorize(token) {
		return 0, fmt.Errorf("unauthorized attempt to upload")
	}

	err := os.MkdirAll(path.Dir(s.Path), 0777)
	if err != nil {
		return 0, err
	}

	fmt.Printf("uploading %d bytes to file %s...\n", len(b), s.Path)
	f, err := os.OpenFile(s.Path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	n, err := f.Write(b)
	if err != nil {
		return n, err
	}

	s.Received += uint64(n)

	if n != len(b) {
		return n, fmt.Errorf("short write: had %d bytes but only wrote %d!", len(b), n)
	}

	return n, nil
}

func (s *Stream) Reader(token string) (io.ReadCloser, error) {
	if !s.Authorize(token) {
		return nil, fmt.Errorf("unauthorized to download")
	}
	return os.Open(s.Path)
}
