package api

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/jhunt/shield-storage-gateway/backend"
)

const (
	StreamIDLength = 96
	FileNameLength = 64
)

type Disposition int

type Stream struct {
	ID          string
	Path        string
	Received    uint64
	Compression string

	token   Token
	backend backend.Backend
}

func (s Stream) Token() string {
	return s.token.Secret
}

func (s Stream) Expires() time.Time {
	return s.token.Expires
}

func NewStream(path string, builder backend.BackendBuilder, config StreamConfig) (Stream, error) {
	id, err := NewRandomString(StreamIDLength)
	if err != nil {
		return Stream{}, err
	}

	return Stream{
		ID:          id,
		Path:        path,
		Compression: config.Compression,

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

func (s Stream) Close() error {
	return s.backend.Close()
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

	if s.Compression != "" {
		r, w := io.Pipe()
		errors := make(chan error, 2)
		size := make(chan int, 1)
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer func() {
				w.Close()
				wg.Done()
			}()
			_, err := backend.Compress(w, b, s.Compression)
			if err != nil {
				errors <- fmt.Errorf("failed to compress data: %s", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(r)
			n := 8192
			split := func(data []byte, atEOF bool) (int, []byte, error) {

				if atEOF && len(data) == 0 {
					return 0, nil, nil
				}

				if len(data) >= n {
					return n, data[0:n], nil
				}

				if atEOF {
					return len(data), data, nil
				}

				return 0, nil, nil
			}
			scanner.Split(split)
			t := 0
			for scanner.Scan() {
				n, err := s.backend.Write(scanner.Bytes())
				if err != nil {
					errors <- err
					return
				}
				t += n
			}
			size <- t
		}()

		wg.Wait()
		close(size)
		close(errors)

		select {
		case err := <-errors:
			return 0, err
		default:
			return <-size, nil
		}
	}

	return s.backend.Write(b)
}

func (s *Stream) AuthorizedRetrieve(token string) (io.ReadCloser, error) {
	if !s.Authorize(token) {
		return nil, fmt.Errorf("unauthorized to download")
	}
	s.token.Renew()

	reader, err := s.backend.Retrieve()
	if err != nil {
		return nil, err
	}
	if s.Compression != "" {
		return backend.Decompress(reader, s.Compression)
	}

	return reader, nil
}
