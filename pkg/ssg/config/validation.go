package config

import (
	"fmt"
	"net/url"
	"strings"
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

func (fs *FS) validate() error {
	if fs == nil {
		return fmt.Errorf("no fs configuration supplied")
	}

	if fs.Root == "" {
		return fmt.Errorf("no root filesystem path provided")
	}

	if !strings.HasPrefix(fs.Root, "/") {
		return fmt.Errorf("root filesystem path provided as relative path (must be absolute)")
	}

	return nil
}

func (s3 *S3) validate() error {
	if s3 == nil {
		return fmt.Errorf("no s3 configuration supplied")
	}

	if s3.Region == "" {
		return fmt.Errorf("no region provided")
	}

	if s3.Bucket == "" {
		return fmt.Errorf("no bucket provided")
	}

	iam := s3.InstanceMetadata
	aki := s3.AccessKeyID != "" && s3.SecretAccessKey != ""

	if iam && aki {
		return fmt.Errorf("instance metadata and access key authentication are mutually exclusive")
	}
	if !iam && !aki {
		return fmt.Errorf("no authentication mechanism defined")
	}

	return nil
}

func (webdav *WebDAV) validate() error {
	if webdav == nil {
		return fmt.Errorf("no webdav configuration supplied")
	}

	if webdav.URL == "" {
		return fmt.Errorf("no webdav url provided")
	}

	u, err := url.Parse(webdav.URL)
	if err != nil {
		return fmt.Errorf("webdav url '%s' is malformed: %s", webdav.URL, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("webdav url '%s' is malformed", webdav.URL)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("webdav url '%s' is malformed", webdav.URL)
	}

	return nil
}
