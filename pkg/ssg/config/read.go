package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func ReadFile(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	return Read(b)
}

func Read(raw []byte) (Config, error) {
	var c Config
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return c, fmt.Errorf("failed to parse yaml: %s", err)
	}

	return c.Resolve()
}

func (c Config) Resolve() (Config, error) {
	// reconcile default configuration with overrides
	if c.Bind == "" {
		c.Bind = Default.Bind
	}
	if c.MaxLease <= 0 {
		c.MaxLease = Default.MaxLease
	}
	if c.SweepInterval <= 0 {
		c.SweepInterval = Default.SweepInterval
	}
	if c.DefaultBucket.Compression == "" {
		c.DefaultBucket.Compression = Default.DefaultBucket.Compression
	}
	if c.DefaultBucket.Encryption == "" {
		c.DefaultBucket.Encryption = Default.DefaultBucket.Encryption
	}

	// validate global configuration
	if c.Bind == "" {
		return c, fmt.Errorf("no bind address specified")
	}
	if c.Cluster == "" {
		return c, fmt.Errorf("no cluster identity specified")
	}
	if c.ControlTokens == nil || len(c.ControlTokens) == 0 {
		return c, fmt.Errorf("no controlTokens specified")
	}

	// validate default bucket configuration
	if !validCompression(c.DefaultBucket.Compression) {
		return c, fmt.Errorf("invalid default bucket compression: '%s'", c.DefaultBucket.Compression)
	}
	if !validEncryption(c.DefaultBucket.Encryption) {
		return c, fmt.Errorf("invalid default bucket encryption: '%s'", c.DefaultBucket.Encryption)
	}

	if len(c.Buckets) == 0 {
		return c, fmt.Errorf("no buckets configured")
	}

	for i, bucket := range c.Buckets {
		// reconcile default buckets with per-bucket overrides
		if bucket.Compression == "" {
			bucket.Compression = c.DefaultBucket.Compression
		}
		if bucket.Encryption == "" {
			bucket.Encryption = c.DefaultBucket.Encryption
		}
		if bucket.Vault == nil {
			bucket.Vault = c.DefaultBucket.Vault
		}

		// validate bucket configuration
		if bucket.Key == "" {
			return c, fmt.Errorf("no bucket key configured for bucket #%d", i+1)
		}
		if !validCompression(bucket.Compression) {
			return c, fmt.Errorf("invalid compression for bucket '%s': '%s'", bucket.Key, bucket.Compression)
		}
		if !validEncryption(bucket.Encryption) {
			return c, fmt.Errorf("invalid encryption for bucket '%s': '%s'", bucket.Key, bucket.Encryption)
		}
		if bucket.Vault == nil && bucket.Encryption != "none" {
			return c, fmt.Errorf("no vault configuration provided for encrypted bucket '%s'", bucket.Key)
		}
		if bucket.Vault != nil {
			if err := bucket.Vault.validate(); err != nil {
				return c, fmt.Errorf("invalid vault configuration for encrypted bucket '%s': %s", bucket.Key, err)
			}
		}

		// validate bucket provider
		switch bucket.Provider.Kind {
		case "fs":
			if err := bucket.Provider.FS.validate(); err != nil {
				return c, fmt.Errorf("invalid configuration for fs-backed bucket '%s': %s", bucket.Key, err)
			}
		case "s3":
			if err := bucket.Provider.S3.validate(); err != nil {
				return c, fmt.Errorf("invalid configuration for s3-backed bucket '%s': %s", bucket.Key, err)
			}
		case "webdav":
			if err := bucket.Provider.WebDAV.validate(); err != nil {
				return c, fmt.Errorf("invalid configuration for webdav-backed bucket '%s': %s", bucket.Key, err)
			}
		}

		// infer bucket defaults, post-validation
		if bucket.Name == "" {
			bucket.Name = bucket.Key
		}
	}

	return c, nil
}
