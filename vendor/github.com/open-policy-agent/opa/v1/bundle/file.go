package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/v1/loader/filter"

	"github.com/open-policy-agent/opa/v1/storage"
)

const maxSizeLimitBytesErrMsg = "bundle file %s size (%d bytes) exceeds configured size_limit_bytes (%d bytes)"

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

func NewDescriptor(url, path string, reader io.Reader) *Descriptor {
	return &Descriptor{
		url:    url,
		path:   path,
		reader: reader,
	}
}

func (d *Descriptor) WithCloser(closer io.Closer) *Descriptor {
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

type PathFormat int64

const (
	Chrooted PathFormat = iota
	SlashRooted
	Passthrough
)

// DirectoryLoader defines an interface which can be used to load
// files from a directory by iterating over each one in the tree.
type DirectoryLoader interface {
	// NextFile must return io.EOF if there is no next value. The returned
	// descriptor should *always* be closed when no longer needed.
	NextFile() (*Descriptor, error)
	WithFilter(filter filter.LoaderFilter) DirectoryLoader
	WithPathFormat(PathFormat) DirectoryLoader
	WithSizeLimitBytes(sizeLimitBytes int64) DirectoryLoader
	WithFollowSymlinks(followSymlinks bool) DirectoryLoader
}

type dirLoader struct {
	root              string
	files             []string
	idx               int
	filter            filter.LoaderFilter
	pathFormat        PathFormat
	maxSizeLimitBytes int64
	followSymlinks    bool
}

// Normalize root directory, ex "./src/bundle" -> "src/bundle"
// We don't need an absolute path, but this makes the joined/trimmed
// paths more uniform.
func normalizeRootDirectory(root string) string {
	if len(root) > 1 {
		if root[0] == '.' && root[1] == filepath.Separator {
			if len(root) == 2 {
				root = root[:1] // "./" -> "."
			} else {
				root = root[2:] // remove leading "./"
			}
		}
	}
	return root
}

// NewDirectoryLoader returns a basic DirectoryLoader implementation
// that will load files from a given root directory path.
func NewDirectoryLoader(root string) DirectoryLoader {
	d := dirLoader{
		root:       normalizeRootDirectory(root),
		pathFormat: Chrooted,
	}
	return &d
}

// WithFilter specifies the filter object to use to filter files while loading bundles
func (d *dirLoader) WithFilter(filter filter.LoaderFilter) DirectoryLoader {
	d.filter = filter
	return d
}

// WithPathFormat specifies how a path is formatted in a Descriptor
func (d *dirLoader) WithPathFormat(pathFormat PathFormat) DirectoryLoader {
	d.pathFormat = pathFormat
	return d
}

// WithSizeLimitBytes specifies the maximum size of any file in the directory to read
func (d *dirLoader) WithSizeLimitBytes(sizeLimitBytes int64) DirectoryLoader {
	d.maxSizeLimitBytes = sizeLimitBytes
	return d
}

// WithFollowSymlinks specifies whether to follow symlinks when loading files from the directory
func (d *dirLoader) WithFollowSymlinks(followSymlinks bool) DirectoryLoader {
	d.followSymlinks = followSymlinks
	return d
}

func formatPath(fileName string, root string, pathFormat PathFormat) string {
	switch pathFormat {
	case SlashRooted:
		if !strings.HasPrefix(fileName, string(filepath.Separator)) {
			return string(filepath.Separator) + fileName
		}
		return fileName
	case Chrooted:
		// Trim off the root directory and return path as if chrooted
		result := strings.TrimPrefix(fileName, filepath.FromSlash(root))
		if root == "." && filepath.Base(fileName) == ManifestExt {
			result = fileName
		}
		if !strings.HasPrefix(result, string(filepath.Separator)) {
			result = string(filepath.Separator) + result
		}
		return result
	case Passthrough:
		fallthrough
	default:
		return fileName
	}
}

// NextFile iterates to the next file in the directory tree
// and returns a file Descriptor for the file.
func (d *dirLoader) NextFile() (*Descriptor, error) {
	// build a list of all files we will iterate over and read, but only one time
	if d.files == nil {
		d.files = []string{}
		err := filepath.Walk(d.root, func(path string, info os.FileInfo, _ error) error {
			if info == nil {
				return nil
			}

			if info.Mode().IsRegular() {
				if d.filter != nil && d.filter(filepath.ToSlash(path), info, getdepth(path, false)) {
					return nil
				}
				if d.maxSizeLimitBytes > 0 && info.Size() > d.maxSizeLimitBytes {
					return fmt.Errorf(maxSizeLimitBytesErrMsg, strings.TrimPrefix(path, "/"), info.Size(), d.maxSizeLimitBytes)
				}
				d.files = append(d.files, path)
			} else if d.followSymlinks && info.Mode().Type()&fs.ModeSymlink == fs.ModeSymlink {
				if d.filter != nil && d.filter(filepath.ToSlash(path), info, getdepth(path, false)) {
					return nil
				}
				if d.maxSizeLimitBytes > 0 && info.Size() > d.maxSizeLimitBytes {
					return fmt.Errorf(maxSizeLimitBytesErrMsg, strings.TrimPrefix(path, "/"), info.Size(), d.maxSizeLimitBytes)
				}
				d.files = append(d.files, path)
			} else if info.Mode().IsDir() {
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

	cleanedPath := formatPath(fileName, d.root, d.pathFormat)
	f := NewDescriptor(filepath.Join(d.root, cleanedPath), cleanedPath, fh).WithCloser(fh)
	return f, nil
}

type tarballLoader struct {
	baseURL           string
	r                 io.Reader
	tr                *tar.Reader
	files             []file
	idx               int
	filter            filter.LoaderFilter
	skipDir           map[string]struct{}
	pathFormat        PathFormat
	maxSizeLimitBytes int64
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
		r:          r,
		pathFormat: Passthrough,
	}
	return &l
}

// NewTarballLoaderWithBaseURL returns a new DirectoryLoader that reads
// files out of a gzipped tar archive. The file URLs will be prefixed
// with the baseURL.
func NewTarballLoaderWithBaseURL(r io.Reader, baseURL string) DirectoryLoader {
	l := tarballLoader{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		r:          r,
		pathFormat: Passthrough,
	}
	return &l
}

// WithFilter specifies the filter object to use to filter files while loading bundles
func (t *tarballLoader) WithFilter(filter filter.LoaderFilter) DirectoryLoader {
	t.filter = filter
	return t
}

// WithPathFormat specifies how a path is formatted in a Descriptor
func (t *tarballLoader) WithPathFormat(pathFormat PathFormat) DirectoryLoader {
	t.pathFormat = pathFormat
	return t
}

// WithSizeLimitBytes specifies the maximum size of any file in the tarball to read
func (t *tarballLoader) WithSizeLimitBytes(sizeLimitBytes int64) DirectoryLoader {
	t.maxSizeLimitBytes = sizeLimitBytes
	return t
}

// WithFollowSymlinks is a no-op for tarballLoader
func (t *tarballLoader) WithFollowSymlinks(_ bool) DirectoryLoader {
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

				if t.maxSizeLimitBytes > 0 && header.Size > t.maxSizeLimitBytes {
					return nil, fmt.Errorf(maxSizeLimitBytesErrMsg, header.Name, header.Size, t.maxSizeLimitBytes)
				}

				f := file{name: header.Name}

				// Note(philipc): We rely on the previous size check in this loop for safety.
				buf := bytes.NewBuffer(make([]byte, 0, header.Size))
				if _, err := io.Copy(buf, t.tr); err != nil {
					return nil, fmt.Errorf("failed to copy file %s: %w", header.Name, err)
				}

				f.reader = buf

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

	cleanedPath := formatPath(f.name, "", t.pathFormat)
	d := NewDescriptor(filepath.Join(t.baseURL, cleanedPath), cleanedPath, f.reader)
	return d, nil
}

// Next implements the storage.Iterator interface.
// It iterates to the next policy or data file in the directory tree
// and returns a storage.Update for the file.
func (it *iterator) Next() (*storage.Update, error) {
	if it.files == nil {
		it.files = []file{}

		for _, item := range it.raw {
			f := file{name: item.Path}

			p, err := getFileStoragePath(f.name)
			if err != nil {
				return nil, err
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

func getFileStoragePath(path string) (storage.Path, error) {
	fpath := strings.TrimLeft(normalizePath(filepath.Dir(path)), "/.")
	if strings.HasSuffix(path, RegoExt) {
		fpath = strings.Trim(normalizePath(path), "/")
	}

	p, ok := storage.ParsePathEscaped("/" + fpath)
	if !ok {
		return nil, fmt.Errorf("storage path invalid: %v", path)
	}
	return p, nil
}
