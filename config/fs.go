package config

// FS represents a local-filesystem storage provider,
// where blobs are persisted to local disk, on the SSG.
//
// This is not a very scalable solution, and it has
// terrible availability prospects, but it does work
// well in test / dev environments, and small deployments.
//
type FS struct {
	// Root specifies the topmost directory into which
	// blob files can be stored.  The FS provider will
	// create directories underneath this root, and
	// store files under those.
	//
	Root string `yaml:"root"`
}
