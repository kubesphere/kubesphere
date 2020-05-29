package bundle

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Descriptor contains information about a file and
// can be used to read the file contents.
type Descriptor struct {
	path      string
	reader    io.Reader
	closer    io.Closer
	closeOnce *sync.Once
}

func newDescriptor(path string, reader io.Reader) *Descriptor {
	return &Descriptor{
		path:   path,
		reader: reader,
	}
}

func (d *Descriptor) withCloser(closer io.Closer) *Descriptor {
	d.closer = closer
	d.closeOnce = new(sync.Once)
	return d
}

// Path returns the path of the file.
func (d *Descriptor) Path() string {
	return d.path
}

// Read will read all the contents from the file the Descriptor refers to
// into the dest writer up n bytes. Will return an io.EOF error
// if EOF is encountered before n bytes are read.
func (d *Descriptor) Read(dest io.Writer, n int64) (int64, error) {
	n, err := io.CopyN(dest, d.reader, n)
	return n, err
}

// Close the file, on some Loader implementations this might be a no-op.
// It should *always* be called regardless of file.
func (d *Descriptor) Close() error {
	var err error
	if d.closer != nil {
		d.closeOnce.Do(func() {
			err = d.closer.Close()
		})
	}
	return err
}

// DirectoryLoader defines an interface which can be used to load
// files from a directory by iterating over each one in the tree.
type DirectoryLoader interface {
	// NextFile must return io.EOF if there is no next value. The returned
	// descriptor should *always* be closed when no longer needed.
	NextFile() (*Descriptor, error)
}

type dirLoader struct {
	root  string
	files []string
	idx   int
}

// NewDirectoryLoader returns a basic DirectoryLoader implementation
// that will load files from a given root directory path.
func NewDirectoryLoader(root string) DirectoryLoader {
	d := dirLoader{
		root: root,
	}
	return &d
}

// NextFile iterates to the next file in the directory tree
// and returns a file Descriptor for the file.
func (d *dirLoader) NextFile() (*Descriptor, error) {
	// build a list of all files we will iterate over and read, but only one time
	if d.files == nil {
		d.files = []string{}
		err := filepath.Walk(d.root, func(path string, info os.FileInfo, err error) error {
			if info != nil && info.Mode().IsRegular() {
				d.files = append(d.files, filepath.ToSlash(path))
			}
			return nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list files")
		}
	}

	// If done reading files then just return io.EOF
	// errors for each NextFile() call
	if d.idx >= len(d.files) {
		return nil, io.EOF
	}

	fileName := d.files[d.idx]
	d.idx++
	fh, err := os.Open(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %s", fileName)
	}

	// Trim off the root directory and return path as if chrooted
	cleanedPath := strings.TrimPrefix(fileName, d.root)
	if !strings.HasPrefix(cleanedPath, "/") {
		cleanedPath = "/" + cleanedPath
	}

	f := newDescriptor(cleanedPath, fh).withCloser(fh)
	return f, nil
}

type tarballLoader struct {
	r  io.Reader
	tr *tar.Reader
}

// NewTarballLoader returns a new DirectoryLoader that reads
// files out of a gzipped tar archive.
func NewTarballLoader(r io.Reader) DirectoryLoader {
	l := tarballLoader{
		r: r,
	}
	return &l
}

// NextFile iterates to the next file in the directory tree
// and returns a file Descriptor for the file.
func (t *tarballLoader) NextFile() (*Descriptor, error) {
	if t.tr == nil {
		gr, err := gzip.NewReader(t.r)
		if err != nil {
			return nil, errors.Wrap(err, "archive read failed")
		}

		t.tr = tar.NewReader(gr)
	}

	for {
		header, err := t.tr.Next()
		// Eventually we will get an io.EOF error when finished
		// iterating through the archive
		if err != nil {
			return nil, err
		}

		// Keep iterating on the archive until we find a normal file
		if header.Typeflag == tar.TypeReg {
			// no need to close this descriptor after reading
			f := newDescriptor(header.Name, t.tr)
			return f, nil
		}
	}
}
