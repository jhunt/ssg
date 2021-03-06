package config

type Config struct {
	// Cluster defines the named storage cluster
	// that this node belongs to.
	//
	// Names are arbitrary, and operator-defined,
	// but ought to be 64 characters or less, and
	// consist primarily of alphanumeric printable
	// characters.
	Cluster string `yaml:"cluster"`

	// Bind defines the network interface(s) that
	// the API should bind to and listen on.
	//
	Bind string `yaml:"bind"`

	// MaxLease defines how many seconds an upload
	// or download can be idle, before it is canceled,
	// and the token invalidated.
	//
	MaxLease int `yaml:"maxLease"`

	// SweepInterval defines how often (in seconds)
	// leases are examined to determine if any of
	// them have expired and need to be cancelled,
	// and have their token invalidated.
	//
	SweepInterval int `yaml:"sweepInterval"`

	// Metrics contains settings related to metrics,
	// monitoring, and measurements.
	Metrics struct {
		// ReservoirSize sets the number of samples
		// used for sampling segment bytes to get
		// the median.  This imposes an fixed upper
		// limit on the memory usage of the monitoring
		// engine, without losing useful measurements.
		//
		// Defaults to a reasonable value of 100.
		//
		ReservoirSize int `yaml:"reservoirSize"`
	} `yaml:"metrics"`

	// ControlTokens is a list of all control bearer
	// tokens, which should be given to systems that
	// are allowed to orchestrate upload, download,
	// and deletion of blobs.
	//
	ControlTokens []string `yaml:"controlTokens"`

	// MonitorTokens is a list of all monitor bearer
	// tokens, which should be given to systems that
	// track the health and wellbeing of the cluster.
	//
	MonitorTokens []string `yaml:"monitorTokens"`

	// DefaultBucket contains global defaults for
	// all buckets that don't explicitly override
	// them.
	//
	DefaultBucket struct {
		// Compression identifies the algorithm to use
		// for compressing blobs, before encryption.
		//
		// Valid values are: 'none', and 'zlib'.
		//
		Compression string `yaml:"compression"`

		// Encryption identifies the algorithm to use
		// for encrypting blobs, after compression.
		//
		// Valid values are: 'none', 'aes256-ctr',
		// 'aes256-cfb', and 'aes256-ofb'.
		//
		Encryption string `yaml:"encryption"`

		// Vault contains the configuration for storing
		// encryption keys securely.  This configuration
		// is ignored if Encryption is set to 'none'.
		//
		Vault *Vault `yaml:"vault"`
	} `yaml:"defaultBucket"`

	// Buckets defines one or more storage buckets, into
	// which SSG callers can place blobs.  Each Bucket
	// is backed by a single backend storage system
	// (like S3, local filesystem, webDAV, etc.), and
	// specifies the compression and encryption algorithms
	// used (if any).
	//
	Buckets []*struct {
		// Key is a durable, internal identifier for this
		// bucket, which will be used by callers to reference'
		// this bucket and any blobs inside of it.
		//
		Key string `yaml:"key"`

		// Name is a human-friendly identifier for this bucket.
		//
		Name string `yaml:"name"`

		// Description provides a human-friendly explanation
		// of this bucket, how it is configured, what it is
		// intended to store, etc.
		//
		Description string `yaml:"description"`

		// Compression identifies the algorithm to use
		// for compressing blobs, before encryption.
		//
		// Valid values are: 'none', and 'zlib'.
		//
		// This overrides DefaultBucket.Compression.
		//
		Compression string `yaml:"compression"`

		// Encryption identifies the algorithm to use
		// for encrypting blobs, after compression.
		//
		// Valid values are: 'none', 'aes256-ctr',
		// 'aes256-cfb', and 'aes256-ofb'.
		//
		// This overrides DefaultBucket.Encryption.
		//
		Encryption string `yaml:"encryption"`

		// Vault contains the configuration for storing
		// encryption keys securely.  This configuration
		// is ignored if Encryption is set to 'none'.
		//
		// This overrides DefaultBucket.Vault in its
		// entirety.
		//
		Vault *Vault `yaml:"vault"`

		// Provider specifies the configuration details
		// of the backing storage provider, and depends
		// quite heavily on the specific system being
		// employed.
		//
		Provider struct {
			// Kind identifies the type of provider in
			// use, and indicates which of the other
			// members of this object can and should be
			// consulted for the rest of the configuration.
			//
			// Valid values are 'fs', 'gcs', 's3', and 'webdav'.
			//
			Kind string `yaml:"kind"`

			// FS represents a local-filesystem storage provider,
			// where blobs are persisted to local disk, on the SSG.
			//
			// This is not a very scalable solution, and it has
			// terrible availability prospects, but it does work
			// well in test / dev environments, and small deployments.
			//
			FS *FS `yaml:"fs"`

			// GCS represents the configuration for Google's Cloud
			// Storage solution (often called GCS) that makes up part
			// of their Google Cloud Platform.
			//
			GCS *GCS `yaml:"gcs"`

			// S3 represents the configuration for many blob storage
			// providers that export an API similar or identical to
			// that of Amazon's Simple Scalable Storage service, S3.
			//
			S3 *S3 `yaml:"s3"`

			// WebDAV represents a storage backend that implements
			// RFC-4918 Web Distributed Authoring and Versioning
			// extensions for HTTP, a read-write version of a
			// regular web server.
			//
			WebDAV *WebDAV `yaml:"webdav"`
		} `yaml:"provider"`
	} `yaml:"buckets"`
}
