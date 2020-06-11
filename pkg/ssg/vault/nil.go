package vault

import (
	"fmt"
)

var Nil NilVault

type NilVault struct{}

func (NilVault) Set(_ string, _ Cipher) error {
	return fmt.Errorf("no vault configured")
}

func (NilVault) Get(_ string) (Cipher, error) {
	return Cipher{}, fmt.Errorf("no vault configured")
}

func (NilVault) Delete(_ string) error {
	return fmt.Errorf("no vault configured")
}
