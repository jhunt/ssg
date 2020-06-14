package vault

import (
	"fmt"
)

var Nil Vault

type NilVault struct{}

func (NilVault) Get(_ string) ([]byte, error) {
	return nil, fmt.Errorf("no vault configured")
}

func (NilVault) SetCipher(_ string, _ Cipher) error {
	return fmt.Errorf("no vault configured")
}

func (NilVault) GetCipher(_ string) (Cipher, error) {
	return Cipher{}, fmt.Errorf("no vault configured")
}

func (NilVault) Delete(_ string) error {
	return fmt.Errorf("no vault configured")
}

func init() {
	Nil = Vault{
		Provider: NilVault{},
	}
}
