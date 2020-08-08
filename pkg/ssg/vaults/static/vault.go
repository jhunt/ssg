package static

import (
	"github.com/jhunt/ssg/pkg/ssg/vault"
)

type Static struct {
	alg string
	fixed vault.FixedKeySource
}

func Configure(alg string, fks vault.FixedKeySource) (Static, error) {
	return Static{
		alg:   alg,
		fixed: fks,
	}, nil
}

func (s Static) SetCipher(string, vault.Cipher) error {
	return nil
}

func (s Static) GetCipher(string) (vault.Cipher, error) {
	return s.fixed.Derive(s.alg, nil)
}

func (s Static) FixedKeyResolver() vault.FixedKeyResolver {
	return vault.PassThroughResolver
}

func (s Static) Delete(string) error {
	return nil
}
