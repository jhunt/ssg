package config

// A CA represents and X.509 PKI Authority for
// validating certificates presented by TLS and
// HTTPS endpoints.
//
//
type CA struct {
	// IgnoreSystem instructs the certificate validation
	// machinery to ignore system-provided root authorities,
	// and to instead only consider the authorities
	// specified explicitly by this configuriation object.
	//
	IgnoreSystem bool `yaml:"ignoreSystem"`

	// SkipVerification instructs the certificate
	// validation machinery to skip all X.509 verifications,
	// and to blindly trust any and all certificates.
	//
	SkipVerification bool `yaml:"skipVerification"`

	// Literal supplies an inline string consisting
	// of one or more PEM-encoded X.509 Certificate
	// Authority certificates to serve as trusted roots.
	//
	// Literal is mutually exclusive with File, and
	// if you specify both, the configuration will be
	// considered invalid.
	//
	Literal string `yaml:"literal"`

	// File supplies the path to a single file containing
	// one or more PEM-encoded X.509 Certificate Authority
	// certificates to serve as trusted roots.
	//
	// File is mutually exclusive with Literal, and
	// if you specify both, the configuration will be
	// considered invalid.
	//
	File string `yaml:"file"`
}
