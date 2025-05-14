package bundle

import (
	"io"

	"github.com/open-policy-agent/opa/storage"
	v1 "github.com/open-policy-agent/opa/v1/bundle"
)

// Descriptor contains information about a file and
// can be used to read the file contents.
type Descriptor = v1.Descriptor

func NewDescriptor(url, path string, reader io.Reader) *Descriptor {
	return v1.NewDescriptor(url, path, reader)
}

type PathFormat = v1.PathFormat

const (
	Chrooted    = v1.Chrooted
	SlashRooted = v1.SlashRooted
	Passthrough = v1.Passthrough
)

// DirectoryLoader defines an interface which can be used to load
// files from a directory by iterating over each one in the tree.
type DirectoryLoader = v1.DirectoryLoader

// NewDirectoryLoader returns a basic DirectoryLoader implementation
// that will load files from a given root directory path.
func NewDirectoryLoader(root string) DirectoryLoader {
	return v1.NewDirectoryLoader(root)
}

// NewTarballLoader is deprecated. Use NewTarballLoaderWithBaseURL instead.
func NewTarballLoader(r io.Reader) DirectoryLoader {
	return v1.NewTarballLoader(r)
}

// NewTarballLoaderWithBaseURL returns a new DirectoryLoader that reads
// files out of a gzipped tar archive. The file URLs will be prefixed
// with the baseURL.
func NewTarballLoaderWithBaseURL(r io.Reader, baseURL string) DirectoryLoader {
	return v1.NewTarballLoaderWithBaseURL(r, baseURL)
}

func NewIterator(raw []Raw) storage.Iterator {
	return v1.NewIterator(raw)
}
