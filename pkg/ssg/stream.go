package ssg

import (
	"io"
	"time"

	"github.com/jhunt/ssg/pkg/rand"
)

func (s *stream) lease(life time.Duration) {
	s.secret = rand.String(32)
	s.leased = time.Now()
	s.expires = s.leased.Add(life)
	s.renewal = life
}

func (s *stream) authorize(token string) bool {
	if s.authorized(token) {
		s.renew()
		return true
	}
	return false
}

func (s *stream) authorized(token string) bool {
	return s.secret == token && !s.expired()
}

func (s *stream) expired() bool {
	return !s.expires.After(time.Now())
}

func (s *stream) renew() {
	s.expires = time.Now().Add(s.renewal)
}

func (s *stream) Read(b []byte) (int, error) {
	n, err := s.reader.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	}

	s.compressed.set(s.reader.ReadCompressed())
	s.uncompressed.set(s.reader.ReadUncompressed())
	s.bucket.metrics.OutFront(s.uncompressed.delta())
	s.bucket.metrics.InBack(s.compressed.delta())

	return n, err
}

func (s *stream) Write(b []byte) (int, error) {
	n, err := s.writer.Write(b)
	if err != nil {
		return n, err
	}

	s.segments++
	s.bucket.metrics.Segment(n)

	s.compressed.set(s.writer.WroteCompressed())
	s.uncompressed.set(s.writer.WroteUncompressed())
	s.bucket.metrics.InFront(s.uncompressed.delta())
	s.bucket.metrics.OutBack(s.compressed.delta())

	return n, nil
}

func (s *stream) Close() error {
	if s.writer != nil {
		err := s.writer.Close()
		if err != nil {
			return err
		}

		s.compressed.set(s.writer.WroteCompressed())
		s.uncompressed.set(s.writer.WroteUncompressed())
		s.bucket.metrics.OutBack(s.compressed.delta())
	}

	if s.reader != nil {
		err := s.reader.Close()
		if err != nil {
			return err
		}

		s.compressed.set(s.reader.ReadCompressed())
		s.uncompressed.set(s.reader.ReadUncompressed())
		s.bucket.metrics.OutFront(s.compressed.delta())
		s.bucket.metrics.InBack(s.compressed.delta())
	}

	return nil
}

func (s *stream) Cancel() error {
	if s.writer != nil {
		return s.writer.Cancel()
	}

	if s.reader != nil {
		return s.reader.Close()
	}

	return nil
}
