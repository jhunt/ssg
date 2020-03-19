package api

import (
	"time"
)

const TokenLength = 32

type Token struct {
	Secret  string
	Issued  time.Time
	Expires time.Time

	renewal time.Duration
}

func NewToken(lifetime time.Duration) (Token, error) {
	secret, err := NewRandomString(TokenLength)
	if err != nil {
		return Token{}, err
	}

	now := time.Now()
	return Token{
		Secret:  secret,
		Issued:  now,
		Expires: now.Add(lifetime),

		renewal: lifetime,
	}, nil
}

func (t Token) Expired() bool {
	return !t.Expires.After(time.Now())
}

func (t *Token) Expire() {
	t.Expires = time.Now().Add(-1 * time.Second)
}

func (t *Token) Renew() {
	t.Expires = time.Now().Add(t.renewal)
}
