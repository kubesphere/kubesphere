// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package bundle implements bundle loading.
package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/open-policy-agent/opa/internal/file/archive"
	"github.com/open-policy-agent/opa/internal/merge"
	"github.com/open-policy-agent/opa/metrics"

	"github.com/pkg/errors"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/util"
)

// Common file extensions and file names.
const (
	RegoExt      = ".rego"
	WasmFile     = "/policy.wasm"
	manifestExt  = ".manifest"
	dataFile     = "data.json"
	yamlDataFile = "data.yaml"
)

const bundleLimitBytes = (1024 * 1024 * 1024) + 1 // limit bundle reads to 1GB to protect against gzip bombs

// Bundle represents a loaded bundle. The bundle can contain data and policies.
type Bundle struct {
	Manifest Manifest
	Data     map[string]interface{}
	Modules  []ModuleFile
	Wasm     []byte
}

// Manifest represents the manifest from a bundle. The manifest may contain
// metadata such as the bundle revision.
type Manifest struct {
	Revision string    `json:"revision"`
	Roots    *[]string `json:"roots,omitempty"`
}

// Init initializes the manifest. If you instantiate a manifest
// manually, call Init to ensure that the roots are set properly.
func (m *Manifest) Init() {
	if m.Roots == nil {
		defaultRoots := []string{""}
		m.Roots = &defaultRoots
	}
}

func (m *Manifest) validateAndInjectDefaults(b Bundle) error {

	m.Init()

	// Validate roots in bundle.
	roots := *m.Roots

	// Standardize the roots (no starting or trailing slash)
	for i := range roots {
		roots[i] = strings.Trim(roots[i], "/")
	}

	for i := 0; i < len(roots)-1; i++ {
		for j := i + 1; j < len(roots); j++ {
			if RootPathsOverlap(roots[i], roots[j]) {
				return fmt.Errorf("manifest has overlapped roots: %v and %v", roots[i], roots[j])
			}
		}
	}

	// Validate modules in bundle.
	for _, module := range b.Modules {
		found := false
		if path, err := module.Parsed.Package.Path.Ptr(); err == nil {
			for i := range roots {
				if strings.HasPrefix(path, roots[i]) {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("manifest roots %v do not permit '%v' in module '%v'", roots, module.Parsed.Package, module.Path)
		}
	}

	// Validate data in bundle.
	return dfs(b.Data, "", func(path string, node interface{}) (bool, error) {
		path = strings.Trim(path, "/")
		for i := range roots {
			if strings.HasPrefix(path, roots[i]) {
				return true, nil
			}
		}
		if _, ok := node.(map[string]interface{}); ok {
			for i := range roots {
				if strings.HasPrefix(roots[i], path) {
					return false, nil
				}
			}
		}
		return false, fmt.Errorf("manifest roots %v do not permit data at path '/%s' (hint: check bundle directory structure)", roots, path)
	})
}

// ModuleFile represents a single module contained a bundle.
type ModuleFile struct {
	Path   string
	Raw    []byte
	Parsed *ast.Module
}

// Reader contains the reader to load the bundle from.
type Reader struct {
	loader                DirectoryLoader
	includeManifestInData bool
	metrics               metrics.Metrics
	baseDir               string
}

// NewReader returns a new Reader which is configured for reading tarballs.
func NewReader(r io.Reader) *Reader {
	return NewCustomReader(NewTarballLoader(r))
}

// NewCustomReader returns a new Reader configured to use the
// specified DirectoryLoader.
func NewCustomReader(loader DirectoryLoader) *Reader {
	nr := Reader{
		loader:  loader,
		metrics: metrics.New(),
	}
	return &nr
}

// IncludeManifestInData sets whether the manifest metadata should be
// included in the bundle's data.
func (r *Reader) IncludeManifestInData(includeManifestInData bool) *Reader {
	r.includeManifestInData = includeManifestInData
	return r
}

// WithMetrics sets the metrics object to be used while loading bundles
func (r *Reader) WithMetrics(m metrics.Metrics) *Reader {
	r.metrics = m
	return r
}

// WithBaseDir sets a base directory for file paths of loaded Rego
// modules. This will *NOT* affect the loaded path of data files.
func (r *Reader) WithBaseDir(dir string) *Reader {
	r.baseDir = dir
	return r
}

// Read returns a new Bundle loaded from the reader.
func (r *Reader) Read() (Bundle, error) {

	var bundle Bundle

	bundle.Data = map[string]interface{}{}

	for {
		f, err := r.loader.NextFile()
		if err == io.EOF {
			break
		}
		if err != nil {
			return bundle, errors.Wrap(err, "bundle read failed")
		}

		var buf bytes.Buffer
		n, err := f.Read(&buf, bundleLimitBytes)
		f.Close() // always close, even on error
		if err != nil && err != io.EOF {
			return bundle, err
		} else if err == nil && n >= bundleLimitBytes {
			return bundle, fmt.Errorf("bundle exceeded max size (%v bytes)", bundleLimitBytes-1)
		}

		// Normalize the paths to use `/` separators
		path := filepath.ToSlash(f.Path())

		if strings.HasSuffix(path, RegoExt) {
			fullPath := r.fullPath(path)
			r.metrics.Timer(metrics.RegoModuleParse).Start()
			module, err := ast.ParseModule(fullPath, buf.String())
			r.metrics.Timer(metrics.RegoModuleParse).Stop()
			if err != nil {
				return bundle, err
			}

			mf := ModuleFile{
				Path:   fullPath,
				Raw:    buf.Bytes(),
				Parsed: module,
			}
			bundle.Modules = append(bundle.Modules, mf)

		} else if path == WasmFile {
			bundle.Wasm = buf.Bytes()

		} else if filepath.Base(path) == dataFile {
			var value interface{}

			r.metrics.Timer(metrics.RegoDataParse).Start()
			err := util.NewJSONDecoder(&buf).Decode(&value)
			r.metrics.Timer(metrics.RegoDataParse).Stop()

			if err != nil {
				return bundle, errors.Wrapf(err, "bundle load failed on %v", r.fullPath(path))
			}

			if err := insertValue(&bundle, path, value); err != nil {
				return bundle, err
			}

		} else if filepath.Base(path) == yamlDataFile {

			var value interface{}

			r.metrics.Timer(metrics.RegoDataParse).Start()
			err := util.Unmarshal(buf.Bytes(), &value)
			r.metrics.Timer(metrics.RegoDataParse).Stop()

			if err != nil {
				return bundle, errors.Wrapf(err, "bundle load failed on %v", r.fullPath(path))
			}

			if err := insertValue(&bundle, path, value); err != nil {
				return bundle, err
			}

		} else if strings.HasSuffix(path, manifestExt) {
			if err := util.NewJSONDecoder(&buf).Decode(&bundle.Manifest); err != nil {
				return bundle, errors.Wrap(err, "bundle load failed on manifest decode")
			}
		}
	}

	if err := bundle.Manifest.validateAndInjectDefaults(bundle); err != nil {
		return bundle, err
	}

	if r.includeManifestInData {
		var metadata map[string]interface{}

		b, err := json.Marshal(&bundle.Manifest)
		if err != nil {
			return bundle, errors.Wrap(err, "bundle load failed on manifest marshal")
		}

		err = util.UnmarshalJSON(b, &metadata)
		if err != nil {
			return bundle, errors.Wrap(err, "bundle load failed on manifest unmarshal")
		}

		// For backwards compatibility always write to the old unnamed manifest path
		// This will *not* be correct if >1 bundle is in use...
		if err := bundle.insert(legacyManifestStoragePath, metadata); err != nil {
			return bundle, errors.Wrapf(err, "bundle load failed on %v", legacyRevisionStoragePath)
		}
	}

	return bundle, nil
}

func (r *Reader) fullPath(path string) string {
	if r.baseDir != "" {
		path = filepath.Join(r.baseDir, path)
	}
	return path
}

// Write serializes the Bundle and writes it to w.
func Write(w io.Writer, bundle Bundle) error {
	gw := gzip.NewWriter(w)
	tw := tar.NewWriter(gw)

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(bundle.Data); err != nil {
		return err
	}

	if err := archive.WriteFile(tw, "data.json", buf.Bytes()); err != nil {
		return err
	}

	for _, module := range bundle.Modules {
		if err := archive.WriteFile(tw, module.Path, module.Raw); err != nil {
			return err
		}
	}

	if err := writeWasm(tw, bundle); err != nil {
		return err
	}

	if err := writeManifest(tw, bundle); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return gw.Close()
}

func writeWasm(tw *tar.Writer, bundle Bundle) error {
	if len(bundle.Wasm) == 0 {
		return nil
	}

	return archive.WriteFile(tw, WasmFile, bundle.Wasm)
}

func writeManifest(tw *tar.Writer, bundle Bundle) error {

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(bundle.Manifest); err != nil {
		return err
	}

	return archive.WriteFile(tw, manifestExt, buf.Bytes())
}

// ParsedModules returns a map of parsed modules with names that are
// unique and human readable for the given a bundle name.
func (b *Bundle) ParsedModules(bundleName string) map[string]*ast.Module {

	mods := make(map[string]*ast.Module, len(b.Modules))

	for _, mf := range b.Modules {
		mods[modulePathWithPrefix(bundleName, mf.Path)] = mf.Parsed
	}

	return mods
}

// Equal returns true if this bundle's contents equal the other bundle's
// contents.
func (b Bundle) Equal(other Bundle) bool {
	if !reflect.DeepEqual(b.Data, other.Data) {
		return false
	}
	if len(b.Modules) != len(other.Modules) {
		return false
	}
	for i := range b.Modules {
		if b.Modules[i].Path != other.Modules[i].Path {
			return false
		}
		if !b.Modules[i].Parsed.Equal(other.Modules[i].Parsed) {
			return false
		}
		if !bytes.Equal(b.Modules[i].Raw, other.Modules[i].Raw) {
			return false
		}
	}
	if (b.Wasm == nil && other.Wasm != nil) || (b.Wasm != nil && other.Wasm == nil) {
		return false
	}

	return bytes.Equal(b.Wasm, other.Wasm)
}

func (b *Bundle) insert(key []string, value interface{}) error {
	// Build an object with the full structure for the value
	obj, err := mktree(key, value)
	if err != nil {
		return err
	}

	// Merge the new data in with the current bundle data object
	merged, ok := merge.InterfaceMaps(b.Data, obj)
	if !ok {
		return fmt.Errorf("failed to insert data file from path %s", filepath.Join(key...))
	}

	b.Data = merged

	return nil
}

func mktree(path []string, value interface{}) (map[string]interface{}, error) {
	if len(path) == 0 {
		// For 0 length path the value is the full tree.
		obj, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("root value must be object")
		}
		return obj, nil
	}

	dir := map[string]interface{}{}
	for i := len(path) - 1; i > 0; i-- {
		dir[path[i]] = value
		value = dir
		dir = map[string]interface{}{}
	}
	dir[path[0]] = value

	return dir, nil
}

// RootPathsOverlap takes in two bundle root paths and returns
// true if they overlap.
func RootPathsOverlap(pathA string, pathB string) bool {

	// Special case for empty prefixes, they always overlap
	if pathA == "" || pathB == "" {
		return true
	}

	aParts := strings.Split(pathA, "/")
	bParts := strings.Split(pathB, "/")

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		if aParts[i] != bParts[i] {
			// Found diverging path segments, no overlap
			return false
		}
	}
	return true
}

func insertValue(b *Bundle, path string, value interface{}) error {

	// Remove leading / and . characters from the directory path. If the bundle
	// was written with OPA then the paths will contain a leading slash. On the
	// other hand, if the path is empty, filepath.Dir will return '.'.
	// Note: filepath.Dir can return paths with '\' separators, always use
	// filepath.ToSlash to keep them normalized.
	dirpath := strings.TrimLeft(filepath.ToSlash(filepath.Dir(path)), "/.")
	var key []string
	if dirpath != "" {
		key = strings.Split(dirpath, "/")
	}
	if err := b.insert(key, value); err != nil {
		return errors.Wrapf(err, "bundle load failed on %v", path)
	}

	return nil
}

func dfs(value interface{}, path string, fn func(string, interface{}) (bool, error)) error {
	if stop, err := fn(path, value); err != nil {
		return err
	} else if stop {
		return nil
	}
	obj, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}
	for key := range obj {
		if err := dfs(obj[key], path+"/"+key, fn); err != nil {
			return err
		}
	}
	return nil
}

func modulePathWithPrefix(bundleName string, modulePath string) string {
	// Default prefix is just the bundle name
	prefix := bundleName

	// Bundle names are sometimes just file paths, some of which
	// are full urls (file:///foo/). Parse these and only use the path.
	parsed, err := url.Parse(bundleName)
	if err == nil {
		prefix = filepath.Join(parsed.Host, parsed.Path)
	}

	return filepath.Join(prefix, modulePath)
}
