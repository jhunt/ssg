package config

// Vault represents the configuration of secure
// credentials storage system that exists outside
// of the storage gateway.
//
type Vault struct {
	FixedKey struct {
		Enabled bool   `yaml:"enabled"`
		PBKDF2  string `yaml:"pbkdf2"`
		AES128  struct {
			Key string `yaml:"key"`
			IV  string `yaml:"iv"`
		} `yaml:"aes128"`
		AES192 struct {
			Key string `yaml:"key"`
			IV  string `yaml:"iv"`
		} `yaml:"aes192"`
		AES256 struct {
			Key string `yaml:"key"`
			IV  string `yaml:"iv"`
		} `yaml:"aes256"`
	} `yaml:"fixedKey"`

	// Kind identifies what type of secure storage
	// system this configuration represents.
	//
	// Currently, the only supported value is
	// "hashicorp"
	//
	Kind string `yaml:"kind"`

	// Hashicorp contains the configuration for
	// Vaults whose `Kind` is set to "hashicorp".
	//
	Hashicorp struct {
		// URL is the base URL of the Vault instance,
		// including the scheme.  Normally this will
		// be an HTTPS URL, but for test / dev purposes,
		// you may want to use a non-TLS endpoint.
		//
		URL string `yaml:"url"`

		// Prefix specifies the path prefix at which
		// to store credentials, and must be specified
		// since it also includes the mountpoint of
		// the KV v2 backend.
		//
		Prefix string `yaml:"prefix"`

		// Token contains a (root) token for accessing
		// the Vault.  This token will not be renewed,
		// so pragmatically, only a root token works.
		//
		// For more secure authentication, use AppRole,
		// by specifying a Role and a Secret.
		//
		// Token is mutually exclusive with Role / Secret,
		// and if you specify both, the configuration will
		// be considered invalid.
		//
		Token string `yaml:"token"`

		// Role contains the AppRole `role_id` value
		// to use when authenticating to this Vault.
		//
		// Role / Secret are mutually exclusive with Token,
		// and if you specify both, the configuration will
		// be considered invalid.
		//
		Role string `yaml:"role"`

		// Secret contains the AppRole `secret_id` value
		// to use when authenticating to this Vault.
		//
		// Role / Secret are mutually exclusive with Token,
		// and if you specify both, the configuration will
		// be considered invalid.
		//
		Secret string `yaml:"secret"`

		// CA provides authority configuration for
		// validating the TLS certificates presented
		// by the Vault instance during normal operation.
		//
		CA CA `yaml:"ca"`
	} `yaml:"hashicorp"`
}
