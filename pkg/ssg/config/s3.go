package config

// S3 represents the configuration for many blob storage
// providers that export an API similar or identical to
// that of Amazon's Simple Scalable Storage service, S3.
//
type S3 struct {
	// URL identifies where the S3 (or S3-like) API endpoint
	// can be found.  This is mostly used for non-Amazon
	// implementations, like Minio or Linode OBJ.
	//
	URL string `yaml:"url"`

	// Region identifies the Amazon region in which S3
	// bucket operations are to be carried out.  Usually,
	// this is the region in which the bucket was created.
	//
	Region string `yaml:"region"`

	// Bucket specifies the name of the S3 bucket to
	// store blobs in.
	//
	Bucket string `yaml:"bucket"`

	// Prefix allows operators to share S3 buckets amongst
	// multiple storage providers without fear of collision.
	//
	// Note that if you wish this to appear filesystem-like,
	// you will need to explicitly end the prefix value
	// with a trailing forward slash ('/').
	//
	Prefix string `yaml:"prefix"`

	// UsePath indicates that the bucket should be sent in
	// the request URL path, not in the hostname, when
	// communicating with the backend.  Official S3 uses
	// DNS-based bucket addressing, but most work-alikes
	// do not.
	//
	UsePath bool `yaml:"usePath"`

	// PartSize sets the size of the pieces to send to the
	// S3 API server, in MiB (1024 * 1024 bytes).
	// Amazon AWS requires this to be at // least 5MiB,
	// but allows it to be larger.
	//
	PartSize int `json:"partSize"`

	// AccessKeyID contains the Access Key ID to use for
	// authenticating to the S3 API.
	//
	// This configuration is ignored if InstanceMetadata
	// is set to true.
	//
	AccessKeyID string `yaml:"accessKeyID"`

	// SecretAccessKey contains the Secret Access Key
	// that corresponds to the given Access Key ID, for
	// authenticating to the S3 API.
	//
	// This configuration is ignored if InstanceMetadata
	// is set to true.
	//
	SecretAccessKey string `yaml:"secretAccessKey"`

	// InstanceMetadata instructs the S3 provider to
	// dynamically acquire S3 authentication tokens
	// using the Amazon EC2 instance metadata API,
	// by crafting a specific HTTP request to a known
	// 169.x.x.x endpoint.
	//
	InstanceMetadata bool `yaml:"instanceMetadata"`
}
