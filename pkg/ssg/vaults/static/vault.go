package static

import (
	"github.com/jhunt/ssg/pkg/ssg/vault"
)

type Static struct {
	fixed vault.FixedKeySource
}

func Configure(fks vault.FixedKeySource) (Static, error) {
	return Static{
		fixed: fks,
	}, nil
}

func (s Static) SetCipher(string, vault.Cipher) error {
	return nil
}

func (s Static) GetCipher(alg string) (vault.Cipher, error) {
	return s.fixed.Derive(alg, nil)
}

func (s Static) FixedKeyResolver() vault.FixedKeyResolver {
	return vault.PassThroughResolver
}

func (s Static) Delete(string) error {
	return nil
}
