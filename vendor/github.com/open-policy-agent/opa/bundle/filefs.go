//go:build go1.16
// +build go1.16

package bundle

import (
	"io/fs"

	v1 "github.com/open-policy-agent/opa/v1/bundle"
)

// NewFSLoader returns a basic DirectoryLoader implementation
// that will load files from a fs.FS interface
func NewFSLoader(filesystem fs.FS) (DirectoryLoader, error) {
	return v1.NewFSLoader(filesystem)
}

// NewFSLoaderWithRoot returns a basic DirectoryLoader implementation
// that will load files from a fs.FS interface at the supplied root
func NewFSLoaderWithRoot(filesystem fs.FS, root string) DirectoryLoader {
	return v1.NewFSLoaderWithRoot(filesystem, root)
}
