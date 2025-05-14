//go:build go1.16
// +build go1.16

package bundle

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/open-policy-agent/opa/v1/loader/filter"
)

const (
	defaultFSLoaderRoot = "."
)

type dirLoaderFS struct {
	sync.Mutex
	filesystem        fs.FS
	files             []string
	idx               int
	filter            filter.LoaderFilter
	root              string
	pathFormat        PathFormat
	maxSizeLimitBytes int64
	followSymlinks    bool
}

// NewFSLoader returns a basic DirectoryLoader implementation
// that will load files from a fs.FS interface
func NewFSLoader(filesystem fs.FS) (DirectoryLoader, error) {
	return NewFSLoaderWithRoot(filesystem, defaultFSLoaderRoot), nil
}

// NewFSLoaderWithRoot returns a basic DirectoryLoader implementation
// that will load files from a fs.FS interface at the supplied root
func NewFSLoaderWithRoot(filesystem fs.FS, root string) DirectoryLoader {
	d := dirLoaderFS{
		filesystem: filesystem,
		root:       normalizeRootDirectory(root),
		pathFormat: Chrooted,
	}

	return &d
}

func (d *dirLoaderFS) walkDir(path string, dirEntry fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if dirEntry != nil {
		info, err := dirEntry.Info()
		if err != nil {
			return err
		}

		if dirEntry.Type().IsRegular() {
			if d.filter != nil && d.filter(filepath.ToSlash(path), info, getdepth(path, false)) {
				return nil
			}

			if d.maxSizeLimitBytes > 0 && info.Size() > d.maxSizeLimitBytes {
				return fmt.Errorf("file %s size %d exceeds limit of %d", path, info.Size(), d.maxSizeLimitBytes)
			}

			d.files = append(d.files, path)
		} else if dirEntry.Type()&fs.ModeSymlink != 0 && d.followSymlinks {
			if d.filter != nil && d.filter(filepath.ToSlash(path), info, getdepth(path, false)) {
				return nil
			}

			if d.maxSizeLimitBytes > 0 && info.Size() > d.maxSizeLimitBytes {
				return fmt.Errorf("file %s size %d exceeds limit of %d", path, info.Size(), d.maxSizeLimitBytes)
			}

			d.files = append(d.files, path)
		} else if dirEntry.Type().IsDir() {
			if d.filter != nil && d.filter(filepath.ToSlash(path), info, getdepth(path, true)) {
				return fs.SkipDir
			}
		}
	}
	return nil
}

// WithFilter specifies the filter object to use to filter files while loading bundles
func (d *dirLoaderFS) WithFilter(filter filter.LoaderFilter) DirectoryLoader {
	d.filter = filter
	return d
}

// WithPathFormat specifies how a path is formatted in a Descriptor
func (d *dirLoaderFS) WithPathFormat(pathFormat PathFormat) DirectoryLoader {
	d.pathFormat = pathFormat
	return d
}

// WithSizeLimitBytes specifies the maximum size of any file in the filesystem directory to read
func (d *dirLoaderFS) WithSizeLimitBytes(sizeLimitBytes int64) DirectoryLoader {
	d.maxSizeLimitBytes = sizeLimitBytes
	return d
}

func (d *dirLoaderFS) WithFollowSymlinks(followSymlinks bool) DirectoryLoader {
	d.followSymlinks = followSymlinks
	return d
}

// NextFile iterates to the next file in the directory tree
// and returns a file Descriptor for the file.
func (d *dirLoaderFS) NextFile() (*Descriptor, error) {
	d.Lock()
	defer d.Unlock()

	if d.files == nil {
		err := fs.WalkDir(d.filesystem, d.root, d.walkDir)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}
	}

	// If done reading files then just return io.EOF
	// errors for each NextFile() call
	if d.idx >= len(d.files) {
		return nil, io.EOF
	}

	fileName := d.files[d.idx]
	d.idx++

	fh, err := d.filesystem.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", fileName, err)
	}

	cleanedPath := formatPath(fileName, d.root, d.pathFormat)
	f := NewDescriptor(cleanedPath, cleanedPath, fh).WithCloser(fh)
	return f, nil
}
