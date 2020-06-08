package hashicorp

import (
	"fmt"
	"path"
	"path/filepath"
	"encoding/hex"
	"net/url"

	"github.com/cloudfoundry-community/vaultkv"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/vault"
)

type Vault struct {
	prefix string
	client vaultkv.Client
	kv     *vaultkv.KV
}

type Endpoint struct {
	Prefix string
	URL    string
	Token  string
}

func Configure(e Endpoint) (Vault, error) {
	if e.Prefix == "" {
		return Vault{}, fmt.Errorf("no prefix supplied")
	}

	u, err := url.Parse(e.URL)
	if err != nil {
		return Vault{}, err
	}

	c := vaultkv.Client{
		VaultURL:  u,
		AuthToken: e.Token,
	}

	return Vault{
		prefix: path.Clean(e.Prefix),
		client: c,
		kv:     c.NewKV(),
	}, nil
}

func (v Vault) Set(id string, c vault.Cipher) error {
	_, err := v.kv.Set(filepath.Join(v.prefix, id), map[string]string{
		"id": id,
		"key": hex.EncodeToString(c.Key),
		"iv": hex.EncodeToString(c.IV),
		"alg": c.Algorithm,
	}, nil)
	return err
}

func (v Vault) Get(id string) (vault.Cipher, error) {
	var in struct {
		ID string `json:"id"`
		Key string `json:"key"`
		IV string `json:"iv"`
		Alg string `json:"alg"`
	}
	_, err := v.kv.Get(filepath.Join(v.prefix, id), &in, nil)
	if err != nil {
		return vault.Cipher{}, err
	}
	if id != in.ID {
		return vault.Cipher{}, fmt.Errorf("id mismatch (credentials are for '%s', not '%s')", in.ID, id)
	}

	c := vault.Cipher{Algorithm: in.Alg}
	c.Key, err = hex.DecodeString(in.Key)
	if err != nil {
		return vault.Cipher{}, err
	}
	c.IV, err = hex.DecodeString(in.IV)
	if err != nil {
		return vault.Cipher{}, err
	}

	return c, nil
}