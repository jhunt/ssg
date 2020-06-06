package vault

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cloudfoundry-community/vaultkv"
)

type Client struct {
	Prefix string

	vault vaultkv.Client
	kv    *vaultkv.KV
}

type Credentials struct {
	SealKey   string `json:"seal_key"`
	RootToken string `json:"root_token"`
}

func Connect(uri, rootToken string) (*Client, error) {
	vaultURI, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("Invalid or malformed Vault URI '%s': %s", uri, err)
	}

	c := &Client{
		Prefix: "secret/",
		vault: vaultkv.Client{
			AuthToken: rootToken,
			VaultURL:  vaultURI,
			Client: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			},
		},
	}
	c.kv = c.vault.NewKV()

	return c, nil
}

type Status int

const (
	Unknown Status = iota
	Blank
	Locked
	Ready
)

func (c *Client) Status() (Status, error) {
	if err := c.vault.Health(false); err == nil {
		return Ready, nil

	} else if vaultkv.IsUninitialized(err) {
		return Blank, nil

	} else if vaultkv.IsSealed(err) || c.vault.TokenIsValid() != nil {
		return Locked, nil

	} else {
		return Unknown, nil
	}
}
func (c *Client) StatusString() (string, error) {
	st, err := c.Status()
	if err != nil {
		return "unknown", err
	}
	switch st {
	case Blank:
		return "uninitialized", nil
	case Locked:
		return "locked", nil
	case Ready:
		return "unlocked", nil
	}
	return "unknown", nil
}
