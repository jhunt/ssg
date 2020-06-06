package vault

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cloudfoundry-community/vaultkv"
)

type Parameters struct {
	Key  string `json:"key"`
	IV   string `json:"iv"`
	Type string `json:"type"`
	UUID string `json:"uuid"`
}

func keygen(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func gen(t string, k, i int) (Parameters, error) {
	key, err := keygen(k)
	if err != nil {
		return Parameters{}, err
	}

	iv, err := keygen(i)
	if err != nil {
		return Parameters{}, err
	}

	return Parameters{
		Type: t,
		Key:  key,
		IV:   iv,
	}, nil
}

func (c *Client) NewParameters(id, typ string) (Parameters, error) {
	enc, err := GenerateRandomParameters(typ)
	if err != nil {
		return Parameters{}, err
	}
	return enc, c.Store(id, enc)
}

func GenerateRandomParameters(typ string) (Parameters, error) {
	cipher := strings.Split(typ, "-")[0]
	switch cipher {
	case "aes128":
		return gen(typ, 128/8, aes.BlockSize)

	case "aes256":
		return gen(typ, 256/8, aes.BlockSize)

	default:
		return Parameters{}, fmt.Errorf("unrecognized cipher/mode '%s'", typ)
	}
}

func (c *Client) pathTo(id string) string {
	return fmt.Sprintf("%s/archives/%s", c.Prefix, id)
}

func (c *Client) Store(id string, params Parameters) error {
	_, err := c.kv.Set(c.pathTo(id), map[string]string{
		"uuid": id,
		"key":  params.Key,
		"iv":   params.IV,
		"type": params.Type,
	}, nil)
	return err
}

func (c *Client) Retrieve(id string) (Parameters, error) {
	var params Parameters
	_, err := c.kv.Get(c.pathTo(id), &params, nil)
	if err != nil {
		return params, fmt.Errorf("failed to retrieve encryption parameters for [%s]: %s", id, err)
	}
	return params, nil
}

func (c *Client) Delete(id string) error {
	return c.kv.Delete(c.pathTo(id), &vaultkv.KVDeleteOpts{V1Destroy: true})
}
