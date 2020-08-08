package hashicorp

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jhunt/go-log"

	"github.com/cloudfoundry-community/vaultkv"
	"github.com/jhunt/ssg/pkg/ssg/config"
	"github.com/jhunt/ssg/pkg/ssg/vault"
)

type Vault struct {
	prefix string
	client *vaultkv.Client
	kv     *vaultkv.KV
}

type Endpoint struct {
	Prefix  string
	URL     string
	Token   string
	CA      config.CA
	Timeout int
}

func Configure(e Endpoint) (Vault, error) {
	if e.Prefix == "" {
		return Vault{}, fmt.Errorf("no prefix supplied")
	}

	u, err := url.Parse(e.URL)
	if err != nil {
		return Vault{}, err
	}

	tlsConfig, err := e.CA.TLSConfig()
	if err != nil {
		return Vault{}, err
	}

	c := &vaultkv.Client{
		VaultURL:  u,
		AuthToken: e.Token,

		Client: &http.Client{
			Timeout: time.Duration(e.Timeout) * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}

	log.Infof(LOG+"configuring hashicorp vault at %v, with prefix=%v", e.URL, e.Prefix)
	return Vault{
		prefix: path.Clean(e.Prefix),
		client: c,
		kv:     c.NewKV(),
	}, nil
}

func (v Vault) SetCipher(id string, c vault.Cipher) error {
	log.Debugf(LOG+"persisting secret %v for cipher [alg=%v]", id, c.Algorithm)
	_, err := v.kv.Set(filepath.Join(v.prefix, id), map[string]string{
		"id":  id,
		"key": hex.EncodeToString(c.Key),
		"iv":  hex.EncodeToString(c.IV),
		"alg": c.Algorithm,
	}, nil)
	return err
}

func (v Vault) Get(id string) ([]byte, error) {
	key := "value"
	if strings.Contains(id, ":") {
		l := strings.Split(id, ":")
		id = strings.Join(l[0:len(l)-1], ":")
		key = l[len(l)-1]
	}

	log.Debugf(LOG+"retrieving raw secret path=%v, key=%v", filepath.Join(v.prefix, id), key)
	in := make(map[string]string)
	_, err := v.kv.Get(filepath.Join(v.prefix, id), &in, nil)
	if err != nil {
		return nil, err
	}
	if v, ok := in[key]; ok {
		return hex.DecodeString(v)
	}
	return nil, fmt.Errorf("key %v not found in path %v", key, id)
}

func (v Vault) GetCipher(id string) (vault.Cipher, error) {
	log.Debugf(LOG+"retrieving secret %v", id)

	var in struct {
		ID  string `json:"id"`
		Key string `json:"key"`
		IV  string `json:"iv"`
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

func (v Vault) Delete(id string) error {
	log.Debugf(LOG+"deleting secret %v", id)
	return v.kv.Delete(filepath.Join(v.prefix, id), &vaultkv.KVDeleteOpts{V1Destroy: true})
}
