package config

// S3 represents the configuration for many blob storage
// providers that export an API similar or identical to
// that of Amazon's Simple Scalable Storage service, S3.
//
type S3 struct {
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
