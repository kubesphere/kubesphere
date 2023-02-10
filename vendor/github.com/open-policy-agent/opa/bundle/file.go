package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/loader/filter"

	"github.com/open-policy-agent/opa/storage"
)

// Descriptor contains information about a file and
// can be used to read the file contents.
type Descriptor struct {
	url       string
	path      string
	reader    io.Reader
	closer    io.Closer
	closeOnce *sync.Once
}

// lazyFile defers reading the file until the first call of Read
type lazyFile struct {
	path string
	file *os.File
}

// newLazyFile creates a new instance of lazyFile
func newLazyFile(path string) *lazyFile {
	return &lazyFile{path: path}
}

// Read implements io.Reader. It will check if the file has been opened
// and open it if it has not before attempting to read using the file's
// read method
func (f *lazyFile) Read(b []byte) (int, error) {
	var err error

	if f.file == nil {
		if f.file, err = os.Open(f.path); err != nil {
			return 0, fmt.Errorf("failed to open file %s: %w", f.path, err)
		}
	}

	return f.file.Read(b)
}

// Close closes the lazy file if it has been opened using the file's
// close method
func (f *lazyFile) Close() error {
	if f.file != nil {
		return f.file.Close()
	}

	return nil
}

func newDescriptor(url, path string, reader io.Reader) *Descriptor {
	return &Descriptor{
		url:    url,
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

// URL returns the url of the file.
func (d *Descriptor) URL() string {
	return d.url
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
	WithFilter(filter filter.LoaderFilter) DirectoryLoader
}

type dirLoader struct {
	root   string
	files  []string
	idx    int
	filter filter.LoaderFilter
}

// NewDirectoryLoader returns a basic DirectoryLoader implementation
// that will load files from a given root directory path.
func NewDirectoryLoader(root string) DirectoryLoader {

	if len(root) > 1 {
		// Normalize relative directories, ex "./src/bundle" -> "src/bundle"
		// We don't need an absolute path, but this makes the joined/trimmed
		// paths more uniform.
		if root[0] == '.' && root[1] == filepath.Separator {
			if len(root) == 2 {
				root = root[:1] // "./" -> "."
			} else {
				root = root[2:] // remove leading "./"
			}
		}
	}

	d := dirLoader{
		root: root,
	}
	return &d
}

// WithFilter specifies the filter object to use to filter files while loading bundles
func (d *dirLoader) WithFilter(filter filter.LoaderFilter) DirectoryLoader {
	d.filter = filter
	return d
}

// NextFile iterates to the next file in the directory tree
// and returns a file Descriptor for the file.
func (d *dirLoader) NextFile() (*Descriptor, error) {
	// build a list of all files we will iterate over and read, but only one time
	if d.files == nil {
		d.files = []string{}
		err := filepath.Walk(d.root, func(path string, info os.FileInfo, err error) error {
			if info != nil && info.Mode().IsRegular() {
				if d.filter != nil && d.filter(filepath.ToSlash(path), info, getdepth(path, false)) {
					return nil
				}
				d.files = append(d.files, path)
			} else if info != nil && info.Mode().IsDir() {
				if d.filter != nil && d.filter(filepath.ToSlash(path), info, getdepth(path, true)) {
					return filepath.SkipDir
				}
			}
			return nil
		})
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
	fh := newLazyFile(fileName)

	// Trim off the root directory and return path as if chrooted
	cleanedPath := strings.TrimPrefix(fileName, filepath.FromSlash(d.root))
	if d.root == "." && filepath.Base(fileName) == ManifestExt {
		cleanedPath = fileName
	}

	if !strings.HasPrefix(cleanedPath, string(os.PathSeparator)) {
		cleanedPath = string(os.PathSeparator) + cleanedPath
	}

	f := newDescriptor(path.Join(d.root, cleanedPath), cleanedPath, fh).withCloser(fh)
	return f, nil
}

type tarballLoader struct {
	baseURL string
	r       io.Reader
	tr      *tar.Reader
	files   []file
	idx     int
	filter  filter.LoaderFilter
	skipDir map[string]struct{}
}

type file struct {
	name   string
	reader io.Reader
	path   storage.Path
	raw    []byte
}

// NewTarballLoader is deprecated. Use NewTarballLoaderWithBaseURL instead.
func NewTarballLoader(r io.Reader) DirectoryLoader {
	l := tarballLoader{
		r: r,
	}
	return &l
}

// NewTarballLoaderWithBaseURL returns a new DirectoryLoader that reads
// files out of a gzipped tar archive. The file URLs will be prefixed
// with the baseURL.
func NewTarballLoaderWithBaseURL(r io.Reader, baseURL string) DirectoryLoader {
	l := tarballLoader{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		r:       r,
	}
	return &l
}

// WithFilter specifies the filter object to use to filter files while loading bundles
func (t *tarballLoader) WithFilter(filter filter.LoaderFilter) DirectoryLoader {
	t.filter = filter
	return t
}

// NextFile iterates to the next file in the directory tree
// and returns a file Descriptor for the file.
func (t *tarballLoader) NextFile() (*Descriptor, error) {
	if t.tr == nil {
		gr, err := gzip.NewReader(t.r)
		if err != nil {
			return nil, fmt.Errorf("archive read failed: %w", err)
		}

		t.tr = tar.NewReader(gr)
	}

	if t.files == nil {
		t.files = []file{}

		if t.skipDir == nil {
			t.skipDir = map[string]struct{}{}
		}

		for {
			header, err := t.tr.Next()
			if err == io.EOF {
				break
			}

			if err != nil {
				return nil, err
			}

			// Keep iterating on the archive until we find a normal file
			if header.Typeflag == tar.TypeReg {

				if t.filter != nil {

					if t.filter(filepath.ToSlash(header.Name), header.FileInfo(), getdepth(header.Name, false)) {
						continue
					}

					basePath := strings.Trim(filepath.Dir(filepath.ToSlash(header.Name)), "/")

					// check if the directory is to be skipped
					if _, ok := t.skipDir[basePath]; ok {
						continue
					}

					match := false
					for p := range t.skipDir {
						if strings.HasPrefix(basePath, p) {
							match = true
							break
						}
					}

					if match {
						continue
					}
				}

				f := file{name: header.Name}

				var buf bytes.Buffer
				if _, err := io.Copy(&buf, t.tr); err != nil {
					return nil, fmt.Errorf("failed to copy file %s: %w", header.Name, err)
				}

				f.reader = &buf

				t.files = append(t.files, f)
			} else if header.Typeflag == tar.TypeDir {
				cleanedPath := filepath.ToSlash(header.Name)
				if t.filter != nil && t.filter(cleanedPath, header.FileInfo(), getdepth(header.Name, true)) {
					t.skipDir[strings.Trim(cleanedPath, "/")] = struct{}{}
				}
			}
		}
	}

	// If done reading files then just return io.EOF
	// errors for each NextFile() call
	if t.idx >= len(t.files) {
		return nil, io.EOF
	}

	f := t.files[t.idx]
	t.idx++

	return newDescriptor(path.Join(t.baseURL, f.name), f.name, f.reader), nil
}

// Next implements the storage.Iterator interface.
// It iterates to the next policy or data file in the directory tree
// and returns a storage.Update for the file.
func (it *iterator) Next() (*storage.Update, error) {

	if it.files == nil {
		it.files = []file{}

		for _, item := range it.raw {
			f := file{name: item.Path}

			fpath := strings.TrimLeft(filepath.ToSlash(filepath.Dir(f.name)), "/.")
			if strings.HasSuffix(f.name, RegoExt) {
				fpath = strings.Trim(f.name, "/")
			}

			p, ok := storage.ParsePathEscaped("/" + fpath)
			if !ok {
				return nil, fmt.Errorf("storage path invalid: %v", f.name)
			}
			f.path = p

			f.raw = item.Value

			it.files = append(it.files, f)
		}

		sortFilePathAscend(it.files)
	}

	// If done reading files then just return io.EOF
	// errors for each NextFile() call
	if it.idx >= len(it.files) {
		return nil, io.EOF
	}

	f := it.files[it.idx]
	it.idx++

	isPolicy := false
	if strings.HasSuffix(f.name, RegoExt) {
		isPolicy = true
	}

	return &storage.Update{
		Path:     f.path,
		Value:    f.raw,
		IsPolicy: isPolicy,
	}, nil
}

type iterator struct {
	raw   []Raw
	files []file
	idx   int
}

func NewIterator(raw []Raw) storage.Iterator {
	it := iterator{
		raw: raw,
	}
	return &it
}

func sortFilePathAscend(files []file) {
	sort.Slice(files, func(i, j int) bool {
		return len(files[i].path) < len(files[j].path)
	})
}

func getdepth(path string, isDir bool) int {
	if isDir {
		cleanedPath := strings.Trim(filepath.ToSlash(path), "/")
		return len(strings.Split(cleanedPath, "/"))
	}

	basePath := strings.Trim(filepath.Dir(filepath.ToSlash(path)), "/")
	return len(strings.Split(basePath, "/"))
}
