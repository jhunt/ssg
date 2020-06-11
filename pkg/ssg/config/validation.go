package config

import (
	"fmt"
)

func validCompression(alg string) bool {
	switch alg {
	case "none",
		"zlib":
		return true
	}

	return false
}

func validEncryption(alg string) bool {
	switch alg {
	case "none",
		"aes128-ctr", "aes128-cfb", "aes128-ofb",
		"aes192-ctr", "aes192-cfb", "aes192-ofb",
		"aes256-ctr", "aes256-cfb", "aes256-ofb":
		return true
	}

	return false
}

func (v *Vault) validate() error {
	if v.Kind != "hashicorp" {
		return fmt.Errorf("unrecognized vault kind '%s'", v.Kind)
	}

	if v.Hashicorp.URL == "" {
		return fmt.Errorf("no vault url specified")
	}

	if v.Hashicorp.Prefix == "" {
		return fmt.Errorf("no vault prefix specified")
	}

	role := v.Hashicorp.Role != "" && v.Hashicorp.Secret != ""
	token := v.Hashicorp.Token != ""
	if token && role {
		return fmt.Errorf("token and approle authentication are mutually exclusive")
	}
	if !token && !role {
		return fmt.Errorf("no authentication mechanism defined")
	}

	return nil
}
