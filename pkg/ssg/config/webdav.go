package config

// WebDAV represents a storage backend that implements
// RFC-4918 Web Distributed Authoring and Versioning
// extensions for HTTP, a read-write version of a
// regular web server.
//
type WebDAV struct {
	// URL specifies the base URL at which to store and
	// retrieve files.  This may contains a request path,
	// to enable sharing of one WebDAV server amongst
	// many different buckets.
	//
	URL string `yaml:"url"`

	// BasicAuth provides the credentials for authenticating
	// to the WebDAV server using HTTP Basic Authentication,
	// a cleartext username / password scheme.
	//
	BasicAuth struct {
		// Username contains the username to authenticate with.
		//
		Username string `yaml:"username"`

		// Password contains the password to authenticate with.
		// Due to the nature of Basic Auth, this password will
		// be sent in the clear to the WebDAV server.
		//
		Password string `yaml:"password"`
	} `yaml:"basicAuth"`

	// CA provides the Certificate Authority configuration
	// to use when validating TLS X.509 Certificates
	// presented by the WebDAV server, during the course of
	// normal operation.
	//
	CA CA `yaml:"ca"`
}
