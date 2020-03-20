package api

import (
	"fmt"
	"io"
	"time"

	"github.com/shieldproject/shield-storage-gateway/backend"
)

const (
	StreamIDLength = 96
	FileNameLength = 64
)

type Disposition int

type Stream struct {
	ID       string
	Path     string
	Received uint64

	token Token
	backend backend.Backend
}

func (s Stream) Token() string {
	return s.token.Secret
}

func (s Stream) Expires() time.Time {
	return s.token.Expires
}

func NewStream(path string, builder backend.BackendBuilder) (Stream, error) {
	id, err := NewRandomString(StreamIDLength)
	if err != nil {
		return Stream{}, err
	}

	return Stream{
		ID:    id,
		Path:  path,

		token:   ExpiredToken,
		backend: builder(path),
	}, nil
}

func (s *Stream) Lease(lifetime time.Duration) error {
	tok, err := NewToken(lifetime)
	if err != nil {
		return err
	}

	s.token = tok
	return nil
}

func (s Stream) Authorize(token string) bool {
	return s.token.Secret == token && !s.token.Expired()
}

func (s Stream) Expired() bool {
	return s.token.Expired()
}

func (s Stream) Cancel() error {
	return s.backend.Cancel()
}

func (s *Stream) AuthorizedWrite(token string, b []byte) (int, error) {
	if !s.Authorize(token) {
		return 0, fmt.Errorf("unauthorized attempt to upload")
	}
	s.token.Renew()
	s.Received += uint64(len(b))
	return s.backend.Write(b)
}

func (s *Stream) AuthorizedRetrieve(token string) (io.ReadCloser, error) {
	if !s.Authorize(token) {
		return nil, fmt.Errorf("unauthorized to download")
	}
	s.token.Renew()
	return s.backend.Retrieve()
}
