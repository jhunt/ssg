package api

import (
	"fmt"
	"io"
	"time"

	"github.com/jhunt/shield-storage-gateway/backend"
	"github.com/jhunt/shield-storage-gateway/compress"
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

	token   Token
	backend backend.Backend

	reader io.ReadCloser
	writer io.WriteCloser
}

func (s Stream) Token() string {
	return s.token.Secret
}

func (s Stream) Expires() time.Time {
	return s.token.Expires
}

func NewUploadStream(path string, builder backend.BackendBuilder) (Stream, error) {
	id, err := NewRandomString(StreamIDLength)
	if err != nil {
		return Stream{}, err
	}

	be := builder(path)
	return Stream{
		ID:   id,
		Path: path,

		token:   ExpiredToken,
		backend: be,
		writer:  be,
	}, nil
}

func NewDownloadStream(path string, builder backend.BackendBuilder) (Stream, error) {
	id, err := NewRandomString(StreamIDLength)
	if err != nil {
		return Stream{}, err
	}

	be := builder(path)
	r, err := be.Retrieve()
	if err != nil {
		return Stream{}, err
	}
	return Stream{
		ID:   id,
		Path: path,

		token:   ExpiredToken,
		backend: be,
		reader:  r,
	}, nil
}

func AuthorizeDelete(path string, builder backend.BackendBuilder) (Stream, error) {
	id, err := NewRandomString(StreamIDLength)
	if err != nil {
		return Stream{}, err
	}

	be := builder(path)
	return Stream{
		ID:   id,
		Path: path,

		token:   ExpiredToken,
		backend: be,
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

func (s Stream) Close() error {
	return s.writer.Close()
}

func (s Stream) Cancel() error {
	if s.writer != nil {
		s.writer.Close()
	}
	return s.backend.Cancel()
}

func (s *Stream) Compress(typ string) error {
	w, err := compress.Compress(s.writer, typ)
	if err != nil {
		return err
	}

	s.writer = w
	return nil
}

func (s *Stream) Decompress(typ string) error {
	r, err := compress.Decompress(s.reader, typ)
	if err != nil {
		return err
	}
	s.reader = r
	return nil
}

func (s *Stream) AuthorizedWrite(token string, b []byte) (int, error) {
	if !s.Authorize(token) {
		return 0, fmt.Errorf("unauthorized attempt to upload")
	}
	s.token.Renew()
	s.Received += uint64(len(b))

	return s.writer.Write(b)
}

func (s *Stream) AuthorizedRetrieve(token string) (io.ReadCloser, error) {
	if !s.Authorize(token) {
		return nil, fmt.Errorf("unauthorized to download")
	}
	s.token.Renew()

	return s.reader, nil
}
