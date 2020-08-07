package config

import (
	"fmt"
	"net/url"
)

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

	// Timeout determines how long HTTP requests can take to
	// connect, issue the request, and read the full response
	// body before they are forcibly disconnected.
	//
	Timeout int `yaml:"timeout"`
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

	if webdav.Timeout < 0 {
		return fmt.Errorf("webdav timeout '%d' is negative", webdav.Timeout)
	}

	return nil
}
