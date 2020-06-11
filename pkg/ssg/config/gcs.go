package config

// GCS represents the configuration for Google's Cloud
// Storage solution (often called GCS) that makes up part
// of their GCS Cloud Platform.
//
type GCS struct {
	// Key represents the JSON service account key used to
	// access Google Cloud Services.  This should be given
	// as an inline JSON / YAML object, to make life easier
	// on operators -- SSG will handle its eventual conversion
	// into compact JSON.
	//
	Key interface{} `yaml:"key"`

	// Bucket specifies the name of the GCS bucket to
	// store blobs in.
	//
	Bucket string `yaml:"bucket"`

	// Prefix allows operators to share GCS buckets amongst
	// multiple storage providers without fear of collision.
	//
	// Note that if you wish this to appear filesystem-like,
	// you will need to explicitly end the prefix value
	// with a trailing forward slash ('/').
	//
	Prefix string `yaml:"prefix"`
}
