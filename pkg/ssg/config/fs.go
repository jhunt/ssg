package config

import (
	"fmt"
	"strings"
)

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

func (fs *FS) validate() error {
	if fs == nil {
		return fmt.Errorf("no fs configuration supplied")
	}

	if fs.Root == "" {
		return fmt.Errorf("no root filesystem path provided")
	}

	if !strings.HasPrefix(fs.Root, "/") {
		return fmt.Errorf("root filesystem path provided as relative path (must be absolute)")
	}

	return nil
}
