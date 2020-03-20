package backend

import (
	"io"
	"io/ioutil"

	"github.com/jhunt/go-s3"
)

type S3 struct {
	key    string
	client *s3.Client
	upload io.WriteCloser
	final chan error
}

func S3Builder(config s3.Client) BackendBuilder {
	return func(path string) Backend {
		return &S3{
			key:    path,
			client: &config,
		}
	}
}

func (s *S3) Write(b []byte) (int, error) {
	if s.upload == nil {
		c, err := s3.NewClient(s.client)
		if err != nil {
			return 0, err
		}
		u, err := c.NewUpload(s.key, nil)
		if err != nil {
			return 0, err
		}

		s.final = make(chan error)
		rd, wr := io.Pipe()
		go func () {
			_, err := u.Stream(rd, 5*1024*1024*1024)
			if err == nil {
				err = u.Done()
			}
			s.final <- err
		}()
		s.upload = wr
	}

	return s.upload.Write(b)
}

func (s *S3) Retrieve() (io.ReadCloser, error) {
	c, err := s3.NewClient(s.client)
	if err != nil {
		return nil, err
	}

	reader, err := c.Get(s.key)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(reader), nil
}

func (s *S3) Close() error {
	if s.upload != nil {
		s.upload.Close()
		return <- s.final
	}
	return nil
}

func (s *S3) Cancel() error {
	c, err := s3.NewClient(s.client)
	if err != nil {
		return err
	}

	return c.Delete(s.key)
}
