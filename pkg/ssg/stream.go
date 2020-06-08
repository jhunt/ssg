package ssg

import (
	"time"

	"github.com/jhunt/shield-storage-gateway/pkg/rand"
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

func (s *stream) expire() {
	s.expires = time.Now().Add(-1 * time.Second)
}

func (s *stream) renew() {
	s.expires = time.Now().Add(s.renewal)
}
