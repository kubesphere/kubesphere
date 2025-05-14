// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package bundle implements bundle loading.
package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gobwas/glob"
	"github.com/open-policy-agent/opa/internal/file/archive"
	"github.com/open-policy-agent/opa/internal/merge"
	"github.com/open-policy-agent/opa/v1/ast"
	astJSON "github.com/open-policy-agent/opa/v1/ast/json"
	"github.com/open-policy-agent/opa/v1/format"
	"github.com/open-policy-agent/opa/v1/metrics"
	"github.com/open-policy-agent/opa/v1/util"
)

// Common file extensions and file names.
const (
	RegoExt               = ".rego"
	WasmFile              = "policy.wasm"
	PlanFile              = "plan.json"
	ManifestExt           = ".manifest"
	SignaturesFile        = "signatures.json"
	patchFile             = "patch.json"
	dataFile              = "data.json"
	yamlDataFile          = "data.yaml"
	ymlDataFile           = "data.yml"
	defaultHashingAlg     = "SHA-256"
	DefaultSizeLimitBytes = (1024 * 1024 * 1024) // limit bundle reads to 1GB to protect against gzip bombs
	DeltaBundleType       = "delta"
	SnapshotBundleType    = "snapshot"
)

// Bundle represents a loaded bundle. The bundle can contain data and policies.
type Bundle struct {
	Signatures  SignaturesConfig
	Manifest    Manifest
	Data        map[string]interface{}
	Modules     []ModuleFile
	Wasm        []byte // Deprecated. Use WasmModules instead
	WasmModules []WasmModuleFile
	PlanModules []PlanModuleFile
	Patch       Patch
	Etag        string
	Raw         []Raw

	lazyLoadingMode bool
	sizeLimitBytes  int64
}

// Raw contains raw bytes representing the bundle's content
type Raw struct {
	Path   string
	Value  []byte
	module *ModuleFile
}

// Patch contains an array of objects wherein each object represents the patch operation to be
// applied to the bundle data.
type Patch struct {
	Data []PatchOperation `json:"data,omitempty"`
}

// PatchOperation models a single patch operation against a document.
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// SignaturesConfig represents an array of JWTs that encapsulate the signatures for the bundle.
type SignaturesConfig struct {
	Signatures []string `json:"signatures,omitempty"`
	Plugin     string   `json:"plugin,omitempty"`
}

// isEmpty returns if the SignaturesConfig is empty.
func (s SignaturesConfig) isEmpty() bool {
	return reflect.DeepEqual(s, SignaturesConfig{})
}

// DecodedSignature represents the decoded JWT payload.
type DecodedSignature struct {
	Files    []FileInfo `json:"files"`
	KeyID    string     `json:"keyid"` // Deprecated, use kid in the JWT header instead.
	Scope    string     `json:"scope"`
	IssuedAt int64      `json:"iat"`
	Issuer   string     `json:"iss"`
}

// FileInfo contains the hashing algorithm used, resulting digest etc.
type FileInfo struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
}

// NewFile returns a new FileInfo.
func NewFile(name, hash, alg string) FileInfo {
	return FileInfo{
		Name:      name,
		Hash:      hash,
		Algorithm: alg,
	}
}

// Manifest represents the manifest from a bundle. The manifest may contain
// metadata such as the bundle revision.
type Manifest struct {
	Revision      string         `json:"revision"`
	Roots         *[]string      `json:"roots,omitempty"`
	WasmResolvers []WasmResolver `json:"wasm,omitempty"`
	// RegoVersion is the global Rego version for the bundle described by this Manifest.
	// The Rego version of individual files can be overridden in FileRegoVersions.
	// We don't use ast.RegoVersion here, as this iota type's order isn't guaranteed to be stable over time.
	// We use a pointer so that we can support hand-made bundles that don't have an explicit version appropriately.
	// E.g. in OPA 0.x if --v1-compatible is used when consuming the bundle, and there is no specified version,
	// we should default to v1; if --v1-compatible isn't used, we should default to v0. In OPA 1.0, no --x-compatible
	// flag and no explicit bundle version should default to v1.
	RegoVersion *int `json:"rego_version,omitempty"`
	// FileRegoVersions is a map from file paths to Rego versions.
	// This allows individual files to override the global Rego version specified by RegoVersion.
	FileRegoVersions map[string]int         `json:"file_rego_versions,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`

	compiledFileRegoVersions []fileRegoVersion
}

type fileRegoVersion struct {
	path    glob.Glob
	version int
}

// WasmResolver maps a wasm module to an entrypoint ref.
type WasmResolver struct {
	Entrypoint  string             `json:"entrypoint,omitempty"`
	Module      string             `json:"module,omitempty"`
	Annotations []*ast.Annotations `json:"annotations,omitempty"`
}

// Init initializes the manifest. If you instantiate a manifest
// manually, call Init to ensure that the roots are set properly.
func (m *Manifest) Init() {
	if m.Roots == nil {
		defaultRoots := []string{""}
		m.Roots = &defaultRoots
	}
}

// AddRoot adds r to the roots of m. This function is idempotent.
func (m *Manifest) AddRoot(r string) {
	m.Init()
	if !RootPathsContain(*m.Roots, r) {
		*m.Roots = append(*m.Roots, r)
	}
}

func (m *Manifest) SetRegoVersion(v ast.RegoVersion) {
	m.Init()
	regoVersion := 0
	if v == ast.RegoV1 {
		regoVersion = 1
	}
	m.RegoVersion = &regoVersion
}

// Equal returns true if m is semantically equivalent to other.
func (m Manifest) Equal(other Manifest) bool {

	// This is safe since both are passed by value.
	m.Init()
	other.Init()

	if m.Revision != other.Revision {
		return false
	}

	if m.RegoVersion == nil && other.RegoVersion != nil {
		return false
	}
	if m.RegoVersion != nil && other.RegoVersion == nil {
		return false
	}
	if m.RegoVersion != nil && other.RegoVersion != nil && *m.RegoVersion != *other.RegoVersion {
		return false
	}

	// If both are nil, or both are empty, we consider them equal.
	if !(len(m.FileRegoVersions) == 0 && len(other.FileRegoVersions) == 0) &&
		!reflect.DeepEqual(m.FileRegoVersions, other.FileRegoVersions) {
		return false
	}

	if !reflect.DeepEqual(m.Metadata, other.Metadata) {
		return false
	}

	return m.equalWasmResolversAndRoots(other)
}

func (m Manifest) Empty() bool {
	return m.Equal(Manifest{})
}

// Copy returns a deep copy of the manifest.
func (m Manifest) Copy() Manifest {
	m.Init()
	roots := make([]string, len(*m.Roots))
	copy(roots, *m.Roots)
	m.Roots = &roots

	wasmModules := make([]WasmResolver, len(m.WasmResolvers))
	copy(wasmModules, m.WasmResolvers)
	m.WasmResolvers = wasmModules

	metadata := m.Metadata

	if metadata != nil {
		m.Metadata = make(map[string]interface{})
		for k, v := range metadata {
			m.Metadata[k] = v
		}
	}

	return m
}

func (m Manifest) String() string {
	m.Init()
	if m.RegoVersion != nil {
		return fmt.Sprintf("<revision: %q, rego_version: %d, roots: %v, wasm: %+v, metadata: %+v>",
			m.Revision, *m.RegoVersion, *m.Roots, m.WasmResolvers, m.Metadata)
	}
	return fmt.Sprintf("<revision: %q, roots: %v, wasm: %+v, metadata: %+v>",
		m.Revision, *m.Roots, m.WasmResolvers, m.Metadata)
}

func (m Manifest) rootSet() stringSet {
	rs := map[string]struct{}{}

	for _, r := range *m.Roots {
		rs[r] = struct{}{}
	}

	return stringSet(rs)
}

func (m Manifest) equalWasmResolversAndRoots(other Manifest) bool {
	if len(m.WasmResolvers) != len(other.WasmResolvers) {
		return false
	}

	for i := range len(m.WasmResolvers) {
		if !m.WasmResolvers[i].Equal(&other.WasmResolvers[i]) {
			return false
		}
	}

	return m.rootSet().Equal(other.rootSet())
}

func (wr *WasmResolver) Equal(other *WasmResolver) bool {
	if wr == nil && other == nil {
		return true
	}

	if wr == nil || other == nil {
		return false
	}

	if wr.Module != other.Module {
		return false
	}

	if wr.Entrypoint != other.Entrypoint {
		return false
	}

	annotLen := len(wr.Annotations)
	if annotLen != len(other.Annotations) {
		return false
	}

	for i := range annotLen {
		if wr.Annotations[i].Compare(other.Annotations[i]) != 0 {
			return false
		}
	}

	return true
}

type stringSet map[string]struct{}

func (ss stringSet) Equal(other stringSet) bool {
	if len(ss) != len(other) {
		return false
	}
	for k := range other {
		if _, ok := ss[k]; !ok {
			return false
		}
	}
	return true
}

func (m *Manifest) validateAndInjectDefaults(b Bundle) error {

	m.Init()

	// Validate roots in bundle.
	roots := *m.Roots

	// Standardize the roots (no starting or trailing slash)
	for i := range roots {
		roots[i] = strings.Trim(roots[i], "/")
	}

	for i := range len(roots) - 1 {
		for j := i + 1; j < len(roots); j++ {
			if RootPathsOverlap(roots[i], roots[j]) {
				return fmt.Errorf("manifest has overlapped roots: '%v' and '%v'", roots[i], roots[j])
			}
		}
	}

	// Validate modules in bundle.
	for _, module := range b.Modules {
		found := false
		if path, err := module.Parsed.Package.Path.Ptr(); err == nil {
			found = RootPathsContain(roots, path)
		}
		if !found {
			return fmt.Errorf("manifest roots %v do not permit '%v' in module '%v'", roots, module.Parsed.Package, module.Path)
		}
	}

	// Build a set of wasm module entrypoints to validate
	wasmModuleToEps := map[string]string{}
	seenEps := map[string]struct{}{}
	for _, wm := range b.WasmModules {
		wasmModuleToEps[wm.Path] = ""
	}

	for _, wmConfig := range b.Manifest.WasmResolvers {
		_, ok := wasmModuleToEps[wmConfig.Module]
		if !ok {
			return fmt.Errorf("manifest references wasm module '%s' but the module file does not exist", wmConfig.Module)
		}

		// Ensure wasm module entrypoint in within bundle roots
		if !RootPathsContain(roots, wmConfig.Entrypoint) {
			return fmt.Errorf("manifest roots %v do not permit '%v' entrypoint for wasm module '%v'", roots, wmConfig.Entrypoint, wmConfig.Module)
		}

		if _, ok := seenEps[wmConfig.Entrypoint]; ok {
			return fmt.Errorf("entrypoint '%s' cannot be used by more than one wasm module", wmConfig.Entrypoint)
		}
		seenEps[wmConfig.Entrypoint] = struct{}{}

		wasmModuleToEps[wmConfig.Module] = wmConfig.Entrypoint
	}

	// Validate data patches in bundle.
	for _, patch := range b.Patch.Data {
		path := strings.Trim(patch.Path, "/")
		if !RootPathsContain(roots, path) {
			return fmt.Errorf("manifest roots %v do not permit data patch at path '%s'", roots, path)
		}
	}

	if b.lazyLoadingMode {
		return nil
	}

	// Validate data in bundle.
	return dfs(b.Data, "", func(path string, node interface{}) (bool, error) {
		path = strings.Trim(path, "/")
		if RootPathsContain(roots, path) {
			return true, nil
		}

		if _, ok := node.(map[string]interface{}); ok {
			for i := range roots {
				if RootPathsContain(strings.Split(path, "/"), roots[i]) {
					return false, nil
				}
			}
		}
		return false, fmt.Errorf("manifest roots %v do not permit data at path '/%s' (hint: check bundle directory structure)", roots, path)
	})
}

// ModuleFile represents a single module contained in a bundle.
type ModuleFile struct {
	URL          string
	Path         string
	RelativePath string
	Raw          []byte
	Parsed       *ast.Module
}

// WasmModuleFile represents a single wasm module contained in a bundle.
type WasmModuleFile struct {
	URL         string
	Path        string
	Entrypoints []ast.Ref
	Raw         []byte
}

// PlanModuleFile represents a single plan module contained in a bundle.
//
// NOTE(tsandall): currently the plans are just opaque binary blobs. In the
// future we could inject the entrypoints so that the plans could be executed
// inside of OPA proper like we do for Wasm modules.
type PlanModuleFile struct {
	URL  string
	Path string
	Raw  []byte
}

// Reader contains the reader to load the bundle from.
type Reader struct {
	loader                DirectoryLoader
	includeManifestInData bool
	metrics               metrics.Metrics
	baseDir               string
	verificationConfig    *VerificationConfig
	skipVerify            bool
	processAnnotations    bool
	capabilities          *ast.Capabilities
	files                 map[string]FileInfo // files in the bundle signature payload
	sizeLimitBytes        int64
	etag                  string
	lazyLoadingMode       bool
	name                  string
	persist               bool
	regoVersion           ast.RegoVersion
	followSymlinks        bool
}

// NewReader is deprecated. Use NewCustomReader instead.
func NewReader(r io.Reader) *Reader {
	return NewCustomReader(NewTarballLoader(r))
}

// NewCustomReader returns a new Reader configured to use the
// specified DirectoryLoader.
func NewCustomReader(loader DirectoryLoader) *Reader {
	nr := Reader{
		loader:         loader,
		metrics:        metrics.New(),
		files:          make(map[string]FileInfo),
		sizeLimitBytes: DefaultSizeLimitBytes + 1,
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

// WithBundleVerificationConfig sets the key configuration used to verify a signed bundle
func (r *Reader) WithBundleVerificationConfig(config *VerificationConfig) *Reader {
	r.verificationConfig = config
	return r
}

// WithSkipBundleVerification skips verification of a signed bundle
func (r *Reader) WithSkipBundleVerification(skipVerify bool) *Reader {
	r.skipVerify = skipVerify
	return r
}

// WithProcessAnnotations enables annotation processing during .rego file parsing.
func (r *Reader) WithProcessAnnotations(yes bool) *Reader {
	r.processAnnotations = yes
	return r
}

// WithCapabilities sets the supported capabilities when loading the files
func (r *Reader) WithCapabilities(caps *ast.Capabilities) *Reader {
	r.capabilities = caps
	return r
}

// WithJSONOptions sets the JSON options on the parser (now a no-op).
//
// Deprecated: Use SetOptions in the json package instead, where a longer description
// of why this is deprecated also can be found.
func (r *Reader) WithJSONOptions(*astJSON.Options) *Reader {
	return r
}

// WithSizeLimitBytes sets the size limit to apply to files in the bundle. If files are larger
// than this, an error will be returned by the reader.
func (r *Reader) WithSizeLimitBytes(n int64) *Reader {
	r.sizeLimitBytes = n + 1
	return r
}

// WithBundleEtag sets the given etag value on the bundle
func (r *Reader) WithBundleEtag(etag string) *Reader {
	r.etag = etag
	return r
}

// WithBundleName specifies the bundle name
func (r *Reader) WithBundleName(name string) *Reader {
	r.name = name
	return r
}

func (r *Reader) WithFollowSymlinks(yes bool) *Reader {
	r.followSymlinks = yes
	return r
}

// WithLazyLoadingMode sets the bundle loading mode. If true,
// bundles will be read in lazy mode. In this mode, data files in the bundle will not be
// deserialized and the check to validate that the bundle data does not contain paths
// outside the bundle's roots will not be performed while reading the bundle.
func (r *Reader) WithLazyLoadingMode(yes bool) *Reader {
	r.lazyLoadingMode = yes
	return r
}

// WithBundlePersistence specifies if the downloaded bundle will eventually be persisted to disk.
func (r *Reader) WithBundlePersistence(persist bool) *Reader {
	r.persist = persist
	return r
}

func (r *Reader) WithRegoVersion(version ast.RegoVersion) *Reader {
	r.regoVersion = version
	return r
}

func (r *Reader) ParserOptions() ast.ParserOptions {
	return ast.ParserOptions{
		ProcessAnnotation: r.processAnnotations,
		Capabilities:      r.capabilities,
		RegoVersion:       r.regoVersion,
	}
}

// Read returns a new Bundle loaded from the reader.
func (r *Reader) Read() (Bundle, error) {

	var bundle Bundle
	var descriptors []*Descriptor
	var err error
	var raw []Raw

	bundle.Signatures, bundle.Patch, descriptors, err = preProcessBundle(r.loader, r.skipVerify, r.sizeLimitBytes)
	if err != nil {
		return bundle, err
	}

	bundle.lazyLoadingMode = r.lazyLoadingMode
	bundle.sizeLimitBytes = r.sizeLimitBytes

	if bundle.Type() == SnapshotBundleType {
		err = r.checkSignaturesAndDescriptors(bundle.Signatures)
		if err != nil {
			return bundle, err
		}

		bundle.Data = map[string]interface{}{}
	}

	var modules []ModuleFile
	for _, f := range descriptors {
		buf, err := readFile(f, r.sizeLimitBytes)
		if err != nil {
			return bundle, err
		}

		// verify the file content
		if bundle.Type() == SnapshotBundleType && !bundle.Signatures.isEmpty() {
			path := f.Path()
			if r.baseDir != "" {
				path = f.URL()
			}
			path = strings.TrimPrefix(path, "/")

			// check if the file is to be excluded from bundle verification
			if r.isFileExcluded(path) {
				delete(r.files, path)
			} else {
				if err = r.verifyBundleFile(path, buf); err != nil {
					return bundle, err
				}
			}
		}

		// Normalize the paths to use `/` separators
		path := filepath.ToSlash(f.Path())

		if strings.HasSuffix(path, RegoExt) {
			fullPath := r.fullPath(path)
			bs := buf.Bytes()

			// Modules are parsed after we've had a chance to read the manifest
			mf := ModuleFile{
				URL:          f.URL(),
				Path:         fullPath,
				RelativePath: path,
				Raw:          bs,
			}
			modules = append(modules, mf)

			if r.lazyLoadingMode {
				p := fullPath
				if r.name != "" {
					p = modulePathWithPrefix(r.name, fullPath)
				}

				raw = append(raw, Raw{Path: p, Value: bs, module: &mf})
			}
		} else if filepath.Base(path) == WasmFile {
			bundle.WasmModules = append(bundle.WasmModules, WasmModuleFile{
				URL:  f.URL(),
				Path: r.fullPath(path),
				Raw:  buf.Bytes(),
			})
		} else if filepath.Base(path) == PlanFile {
			bundle.PlanModules = append(bundle.PlanModules, PlanModuleFile{
				URL:  f.URL(),
				Path: r.fullPath(path),
				Raw:  buf.Bytes(),
			})
		} else if filepath.Base(path) == dataFile {
			if r.lazyLoadingMode {
				raw = append(raw, Raw{Path: path, Value: buf.Bytes()})
				continue
			}

			var value interface{}

			r.metrics.Timer(metrics.RegoDataParse).Start()
			err := util.UnmarshalJSON(buf.Bytes(), &value)
			r.metrics.Timer(metrics.RegoDataParse).Stop()

			if err != nil {
				return bundle, fmt.Errorf("bundle load failed on %v: %w", r.fullPath(path), err)
			}

			if err := insertValue(&bundle, path, value); err != nil {
				return bundle, err
			}

		} else if filepath.Base(path) == yamlDataFile || filepath.Base(path) == ymlDataFile {
			if r.lazyLoadingMode {
				raw = append(raw, Raw{Path: path, Value: buf.Bytes()})
				continue
			}

			var value interface{}

			r.metrics.Timer(metrics.RegoDataParse).Start()
			err := util.Unmarshal(buf.Bytes(), &value)
			r.metrics.Timer(metrics.RegoDataParse).Stop()

			if err != nil {
				return bundle, fmt.Errorf("bundle load failed on %v: %w", r.fullPath(path), err)
			}

			if err := insertValue(&bundle, path, value); err != nil {
				return bundle, err
			}

		} else if strings.HasSuffix(path, ManifestExt) {
			if err := util.NewJSONDecoder(&buf).Decode(&bundle.Manifest); err != nil {
				return bundle, fmt.Errorf("bundle load failed on manifest decode: %w", err)
			}
		}
	}

	// Parse modules
	popts := r.ParserOptions()
	popts.RegoVersion = bundle.RegoVersion(popts.EffectiveRegoVersion())
	for _, mf := range modules {
		modulePopts := popts
		if regoVersion, err := bundle.RegoVersionForFile(mf.RelativePath, popts.EffectiveRegoVersion()); err != nil {
			return bundle, err
		} else if regoVersion != ast.RegoUndefined {
			// We don't expect ast.RegoUndefined here, but don't override configured rego-version if we do just to be extra protective
			modulePopts.RegoVersion = regoVersion
		}
		r.metrics.Timer(metrics.RegoModuleParse).Start()
		mf.Parsed, err = ast.ParseModuleWithOpts(mf.Path, string(mf.Raw), modulePopts)
		r.metrics.Timer(metrics.RegoModuleParse).Stop()
		if err != nil {
			return bundle, err
		}
		bundle.Modules = append(bundle.Modules, mf)
	}

	if bundle.Type() == DeltaBundleType {
		if len(bundle.Data) != 0 {
			return bundle, errors.New("delta bundle expected to contain only patch file but data files found")
		}

		if len(bundle.Modules) != 0 {
			return bundle, errors.New("delta bundle expected to contain only patch file but policy files found")
		}

		if len(bundle.WasmModules) != 0 {
			return bundle, errors.New("delta bundle expected to contain only patch file but wasm files found")
		}

		if r.persist {
			return bundle, errors.New("'persist' property is true in config. persisting delta bundle to disk is not supported")
		}
	}

	// check if the bundle signatures specify any files that weren't found in the bundle
	if bundle.Type() == SnapshotBundleType && len(r.files) != 0 {
		extra := []string{}
		for k := range r.files {
			extra = append(extra, k)
		}
		return bundle, fmt.Errorf("file(s) %v specified in bundle signatures but not found in the target bundle", extra)
	}

	if err := bundle.Manifest.validateAndInjectDefaults(bundle); err != nil {
		return bundle, err
	}

	// Inject the wasm module entrypoint refs into the WasmModuleFile structs
	epMap := map[string][]string{}
	for _, r := range bundle.Manifest.WasmResolvers {
		epMap[r.Module] = append(epMap[r.Module], r.Entrypoint)
	}
	for i := range len(bundle.WasmModules) {
		entrypoints := epMap[bundle.WasmModules[i].Path]
		for _, entrypoint := range entrypoints {
			ref, err := ast.PtrRef(ast.DefaultRootDocument, entrypoint)
			if err != nil {
				return bundle, fmt.Errorf("failed to parse wasm module entrypoint '%s': %s", entrypoint, err)
			}
			bundle.WasmModules[i].Entrypoints = append(bundle.WasmModules[i].Entrypoints, ref)
		}
	}

	if r.includeManifestInData {
		var metadata map[string]interface{}

		b, err := json.Marshal(&bundle.Manifest)
		if err != nil {
			return bundle, fmt.Errorf("bundle load failed on manifest marshal: %w", err)
		}

		err = util.UnmarshalJSON(b, &metadata)
		if err != nil {
			return bundle, fmt.Errorf("bundle load failed on manifest unmarshal: %w", err)
		}

		// For backwards compatibility always write to the old unnamed manifest path
		// This will *not* be correct if >1 bundle is in use...
		if err := bundle.insertData(legacyManifestStoragePath, metadata); err != nil {
			return bundle, fmt.Errorf("bundle load failed on %v: %w", legacyRevisionStoragePath, err)
		}
	}

	bundle.Etag = r.etag
	bundle.Raw = raw

	return bundle, nil
}

func (r *Reader) isFileExcluded(path string) bool {
	for _, e := range r.verificationConfig.Exclude {
		match, _ := filepath.Match(e, path)
		if match {
			return true
		}
	}
	return false
}

func (r *Reader) checkSignaturesAndDescriptors(signatures SignaturesConfig) error {
	if r.skipVerify {
		return nil
	}

	if signatures.isEmpty() && r.verificationConfig != nil && r.verificationConfig.KeyID != "" {
		return errors.New("bundle missing .signatures.json file")
	}

	if !signatures.isEmpty() {
		if r.verificationConfig == nil {
			return errors.New("verification key not provided")
		}

		// verify the JWT signatures included in the `.signatures.json` file
		if err := r.verifyBundleSignature(signatures); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) verifyBundleSignature(sc SignaturesConfig) error {
	var err error
	r.files, err = VerifyBundleSignature(sc, r.verificationConfig)
	return err
}

func (r *Reader) verifyBundleFile(path string, data bytes.Buffer) error {
	return VerifyBundleFile(path, data, r.files)
}

func (r *Reader) fullPath(path string) string {
	if r.baseDir != "" {
		path = filepath.Join(r.baseDir, path)
	}
	return path
}

// Write is deprecated. Use NewWriter instead.
func Write(w io.Writer, bundle Bundle) error {
	return NewWriter(w).
		UseModulePath(true).
		DisableFormat(true).
		Write(bundle)
}

// Writer implements bundle serialization.
type Writer struct {
	usePath       bool
	disableFormat bool
	w             io.Writer
}

// NewWriter returns a bundle writer that writes to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// UseModulePath configures the writer to use the module file path instead of the
// module file URL during serialization. This is for backwards compatibility.
func (w *Writer) UseModulePath(yes bool) *Writer {
	w.usePath = yes
	return w
}

// DisableFormat configures the writer to just write out raw bytes instead
// of formatting modules before serialization.
func (w *Writer) DisableFormat(yes bool) *Writer {
	w.disableFormat = yes
	return w
}

// Write writes the bundle to the writer's output stream.
func (w *Writer) Write(bundle Bundle) error {
	gw := gzip.NewWriter(w.w)
	tw := tar.NewWriter(gw)

	bundleType := bundle.Type()

	if bundleType == SnapshotBundleType {
		var buf bytes.Buffer

		if err := json.NewEncoder(&buf).Encode(bundle.Data); err != nil {
			return err
		}

		if err := archive.WriteFile(tw, "data.json", buf.Bytes()); err != nil {
			return err
		}

		for _, module := range bundle.Modules {
			path := module.URL
			if w.usePath {
				path = module.Path
			}

			if err := archive.WriteFile(tw, path, module.Raw); err != nil {
				return err
			}
		}

		if err := w.writeWasm(tw, bundle); err != nil {
			return err
		}

		if err := writeSignatures(tw, bundle); err != nil {
			return err
		}

		if err := w.writePlan(tw, bundle); err != nil {
			return err
		}
	} else if bundleType == DeltaBundleType {
		if err := writePatch(tw, bundle); err != nil {
			return err
		}
	}

	if err := writeManifest(tw, bundle); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return gw.Close()
}

func (w *Writer) writeWasm(tw *tar.Writer, bundle Bundle) error {
	for _, wm := range bundle.WasmModules {
		path := wm.URL
		if w.usePath {
			path = wm.Path
		}

		err := archive.WriteFile(tw, path, wm.Raw)
		if err != nil {
			return err
		}
	}

	if len(bundle.Wasm) > 0 {
		err := archive.WriteFile(tw, "/"+WasmFile, bundle.Wasm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) writePlan(tw *tar.Writer, bundle Bundle) error {
	for _, wm := range bundle.PlanModules {
		path := wm.URL
		if w.usePath {
			path = wm.Path
		}

		err := archive.WriteFile(tw, path, wm.Raw)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeManifest(tw *tar.Writer, bundle Bundle) error {

	if bundle.Manifest.Empty() {
		return nil
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(bundle.Manifest); err != nil {
		return err
	}

	return archive.WriteFile(tw, ManifestExt, buf.Bytes())
}

func writePatch(tw *tar.Writer, bundle Bundle) error {

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(bundle.Patch); err != nil {
		return err
	}

	return archive.WriteFile(tw, patchFile, buf.Bytes())
}

func writeSignatures(tw *tar.Writer, bundle Bundle) error {

	if bundle.Signatures.isEmpty() {
		return nil
	}

	bs, err := json.MarshalIndent(bundle.Signatures, "", " ")
	if err != nil {
		return err
	}

	return archive.WriteFile(tw, fmt.Sprintf(".%v", SignaturesFile), bs)
}

func hashBundleFiles(hash SignatureHasher, b *Bundle) ([]FileInfo, error) {

	files := []FileInfo{}

	bs, err := hash.HashFile(b.Data)
	if err != nil {
		return files, err
	}
	files = append(files, NewFile(strings.TrimPrefix("data.json", "/"), hex.EncodeToString(bs), defaultHashingAlg))

	if len(b.Wasm) != 0 {
		bs, err := hash.HashFile(b.Wasm)
		if err != nil {
			return files, err
		}
		files = append(files, NewFile(strings.TrimPrefix(WasmFile, "/"), hex.EncodeToString(bs), defaultHashingAlg))
	}

	for _, wasmModule := range b.WasmModules {
		bs, err := hash.HashFile(wasmModule.Raw)
		if err != nil {
			return files, err
		}
		files = append(files, NewFile(strings.TrimPrefix(wasmModule.Path, "/"), hex.EncodeToString(bs), defaultHashingAlg))
	}

	for _, planmodule := range b.PlanModules {
		bs, err := hash.HashFile(planmodule.Raw)
		if err != nil {
			return files, err
		}
		files = append(files, NewFile(strings.TrimPrefix(planmodule.Path, "/"), hex.EncodeToString(bs), defaultHashingAlg))
	}

	// If the manifest is essentially empty, don't add it to the signatures since it
	// won't be written to the bundle. Otherwise:
	// parse the manifest into a JSON structure;
	// then recursively order the fields of all objects alphabetically and then apply
	// the hash function to result to compute the hash.
	if !b.Manifest.Empty() {
		mbs, err := json.Marshal(b.Manifest)
		if err != nil {
			return files, err
		}

		var result map[string]interface{}
		if err := util.Unmarshal(mbs, &result); err != nil {
			return files, err
		}

		bs, err = hash.HashFile(result)
		if err != nil {
			return files, err
		}

		files = append(files, NewFile(strings.TrimPrefix(ManifestExt, "/"), hex.EncodeToString(bs), defaultHashingAlg))
	}

	return files, err
}

// FormatModules formats Rego modules
// Modules will be formatted to comply with [ast.DefaultRegoVersion], but Rego compatibility of individual parsed modules will be respected (e.g. if 'rego.v1' is imported).
func (b *Bundle) FormatModules(useModulePath bool) error {
	return b.FormatModulesForRegoVersion(ast.DefaultRegoVersion, true, useModulePath)
}

// FormatModulesForRegoVersion formats Rego modules to comply with a given Rego version
func (b *Bundle) FormatModulesForRegoVersion(version ast.RegoVersion, preserveModuleRegoVersion bool, useModulePath bool) error {
	var err error

	for i, module := range b.Modules {
		opts := format.Opts{}
		if preserveModuleRegoVersion {
			opts.RegoVersion = module.Parsed.RegoVersion()
			opts.ParserOptions = &ast.ParserOptions{
				RegoVersion: opts.RegoVersion,
			}
		} else {
			opts.RegoVersion = version
		}

		if module.Raw == nil {
			module.Raw, err = format.AstWithOpts(module.Parsed, opts)
			if err != nil {
				return err
			}
		} else {
			path := module.URL
			if useModulePath {
				path = module.Path
			}

			module.Raw, err = format.SourceWithOpts(path, module.Raw, opts)
			if err != nil {
				return err
			}
		}
		b.Modules[i].Raw = module.Raw
	}
	return nil
}

// GenerateSignature generates the signature for the given bundle.
func (b *Bundle) GenerateSignature(signingConfig *SigningConfig, keyID string, useModulePath bool) error {

	hash, err := NewSignatureHasher(HashingAlgorithm(defaultHashingAlg))
	if err != nil {
		return err
	}

	files := []FileInfo{}

	for _, module := range b.Modules {
		bytes, err := hash.HashFile(module.Raw)
		if err != nil {
			return err
		}

		path := module.URL
		if useModulePath {
			path = module.Path
		}
		files = append(files, NewFile(strings.TrimPrefix(path, "/"), hex.EncodeToString(bytes), defaultHashingAlg))
	}

	result, err := hashBundleFiles(hash, b)
	if err != nil {
		return err
	}
	files = append(files, result...)

	// generate signed token
	token, err := GenerateSignedToken(files, signingConfig, keyID)
	if err != nil {
		return err
	}

	if b.Signatures.isEmpty() {
		b.Signatures = SignaturesConfig{}
	}

	if signingConfig.Plugin != "" {
		b.Signatures.Plugin = signingConfig.Plugin
	}

	b.Signatures.Signatures = []string{token}

	return nil
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

func (b *Bundle) RegoVersion(def ast.RegoVersion) ast.RegoVersion {
	if v := b.Manifest.RegoVersion; v != nil {
		if *v == 0 {
			return ast.RegoV0
		} else if *v == 1 {
			return ast.RegoV1
		}
	}
	return def
}

func (b *Bundle) SetRegoVersion(v ast.RegoVersion) {
	b.Manifest.SetRegoVersion(v)
}

// RegoVersionForFile returns the rego-version for the specified file path.
// If there is no defined version for the given path, the default version def is returned.
// If the version does not correspond to ast.RegoV0 or ast.RegoV1, an error is returned.
func (b *Bundle) RegoVersionForFile(path string, def ast.RegoVersion) (ast.RegoVersion, error) {
	version, err := b.Manifest.numericRegoVersionForFile(path)
	if err != nil {
		return def, err
	} else if version == nil {
		return def, nil
	} else if *version == 0 {
		return ast.RegoV0, nil
	} else if *version == 1 {
		return ast.RegoV1, nil
	}
	return def, fmt.Errorf("unknown bundle rego-version %d for file '%s'", *version, path)
}

func (m *Manifest) RegoVersionForFile(path string) (ast.RegoVersion, error) {
	v, err := m.numericRegoVersionForFile(path)
	if err != nil {
		return ast.RegoUndefined, err
	}

	if v == nil {
		return ast.RegoUndefined, nil
	}

	return ast.RegoVersionFromInt(*v), nil
}

func (m *Manifest) numericRegoVersionForFile(path string) (*int, error) {
	var version *int

	if len(m.FileRegoVersions) != len(m.compiledFileRegoVersions) {
		m.compiledFileRegoVersions = make([]fileRegoVersion, 0, len(m.FileRegoVersions))
		for pattern, v := range m.FileRegoVersions {
			compiled, err := glob.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("failed to compile glob pattern %s: %s", pattern, err)
			}
			m.compiledFileRegoVersions = append(m.compiledFileRegoVersions, fileRegoVersion{compiled, v})
		}
	}

	for _, fv := range m.compiledFileRegoVersions {
		if fv.path.Match(path) {
			version = &fv.version
			break
		}
	}

	if version == nil {
		version = m.RegoVersion
	}
	return version, nil
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
		// To support bundles built from rootless filesystems we ignore a "/" prefix
		// for URLs and Paths, such that "/file" and "file" are equivalent
		if strings.TrimPrefix(b.Modules[i].URL, string(filepath.Separator)) !=
			strings.TrimPrefix(other.Modules[i].URL, string(filepath.Separator)) {
			return false
		}
		if strings.TrimPrefix(b.Modules[i].Path, string(filepath.Separator)) !=
			strings.TrimPrefix(other.Modules[i].Path, string(filepath.Separator)) {
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

// Copy returns a deep copy of the bundle.
func (b Bundle) Copy() Bundle {

	// Copy data.
	var x interface{} = b.Data

	if err := util.RoundTrip(&x); err != nil {
		panic(err)
	}

	if x != nil {
		b.Data = x.(map[string]interface{})
	}

	// Copy modules.
	for i := range b.Modules {
		bs := make([]byte, len(b.Modules[i].Raw))
		copy(bs, b.Modules[i].Raw)
		b.Modules[i].Raw = bs
		b.Modules[i].Parsed = b.Modules[i].Parsed.Copy()
	}

	// Copy manifest.
	b.Manifest = b.Manifest.Copy()

	return b
}

func (b *Bundle) insertData(key []string, value interface{}) error {
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

func (b *Bundle) readData(key []string) *interface{} {

	if len(key) == 0 {
		if len(b.Data) == 0 {
			return nil
		}
		var result interface{} = b.Data
		return &result
	}

	node := b.Data

	for i := range len(key) - 1 {

		child, ok := node[key[i]]
		if !ok {
			return nil
		}

		childObj, ok := child.(map[string]interface{})
		if !ok {
			return nil
		}

		node = childObj
	}

	child, ok := node[key[len(key)-1]]
	if !ok {
		return nil
	}

	return &child
}

// Type returns the type of the bundle.
func (b *Bundle) Type() string {
	if len(b.Patch.Data) != 0 {
		return DeltaBundleType
	}
	return SnapshotBundleType
}

func mktree(path []string, value interface{}) (map[string]interface{}, error) {
	if len(path) == 0 {
		// For 0 length path the value is the full tree.
		obj, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.New("root value must be object")
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

// Merge accepts a set of bundles and merges them into a single result bundle. If there are
// any conflicts during the merge (e.g., with roots) an error is returned. The result bundle
// will have an empty revision except in the special case where a single bundle is provided
// (and in that case the bundle is just returned unmodified.)
func Merge(bundles []*Bundle) (*Bundle, error) {
	return MergeWithRegoVersion(bundles, ast.DefaultRegoVersion, false)
}

// MergeWithRegoVersion creates a merged bundle from the provided bundles, similar to Merge.
// If more than one bundle is provided, the rego version of the result bundle is set to the provided regoVersion.
// Any Rego files in a bundle of conflicting rego version will be marked in the result's manifest with the rego version
// of its original bundle. If the Rego file already had an overriding rego version, it will be preserved.
// If a single bundle is provided, it will retain any rego version information it already had. If it has none, the
// provided regoVersion will be applied to it.
// If usePath is true, per-file rego-versions will be calculated using the file's ModuleFile.Path; otherwise, the file's
// ModuleFile.URL will be used.
func MergeWithRegoVersion(bundles []*Bundle, regoVersion ast.RegoVersion, usePath bool) (*Bundle, error) {

	if len(bundles) == 0 {
		return nil, errors.New("expected at least one bundle")
	}

	if regoVersion == ast.RegoUndefined {
		regoVersion = ast.DefaultRegoVersion
	}

	if len(bundles) == 1 {
		result := bundles[0]
		// We respect the bundle rego-version, defaulting to the provided rego version if not set.
		result.SetRegoVersion(result.RegoVersion(regoVersion))
		fileRegoVersions, err := bundleRegoVersions(result, result.RegoVersion(regoVersion), usePath)
		if err != nil {
			return nil, err
		}
		result.Manifest.FileRegoVersions = fileRegoVersions
		return result, nil
	}

	var roots []string
	var result Bundle

	for _, b := range bundles {

		if b.Manifest.Roots == nil {
			return nil, errors.New("bundle manifest not initialized")
		}

		roots = append(roots, *b.Manifest.Roots...)

		result.Modules = append(result.Modules, b.Modules...)

		for _, root := range *b.Manifest.Roots {
			key := strings.Split(root, "/")
			if val := b.readData(key); val != nil {
				if err := result.insertData(key, *val); err != nil {
					return nil, err
				}
			}
		}

		result.Manifest.WasmResolvers = append(result.Manifest.WasmResolvers, b.Manifest.WasmResolvers...)
		result.WasmModules = append(result.WasmModules, b.WasmModules...)
		result.PlanModules = append(result.PlanModules, b.PlanModules...)

		if b.Manifest.RegoVersion != nil || len(b.Manifest.FileRegoVersions) > 0 {
			if result.Manifest.FileRegoVersions == nil {
				result.Manifest.FileRegoVersions = map[string]int{}
			}

			fileRegoVersions, err := bundleRegoVersions(b, regoVersion, usePath)
			if err != nil {
				return nil, err
			}
			for k, v := range fileRegoVersions {
				result.Manifest.FileRegoVersions[k] = v
			}
		}
	}

	// We respect the bundle rego-version, defaulting to the provided rego version if not set.
	result.SetRegoVersion(result.RegoVersion(regoVersion))

	if result.Data == nil {
		result.Data = map[string]interface{}{}
	}

	result.Manifest.Roots = &roots

	if err := result.Manifest.validateAndInjectDefaults(result); err != nil {
		return nil, err
	}

	return &result, nil
}

func bundleRegoVersions(bundle *Bundle, regoVersion ast.RegoVersion, usePath bool) (map[string]int, error) {
	fileRegoVersions := map[string]int{}

	// we drop the bundle-global rego versions and record individual rego versions for each module.
	for _, m := range bundle.Modules {
		// We fetch rego-version by the path relative to the bundle root, as the complete path of the module might
		// contain the path between OPA working directory and the bundle root.
		v, err := bundle.RegoVersionForFile(bundleRelativePath(m, usePath), bundle.RegoVersion(regoVersion))
		if err != nil {
			return nil, err
		}

		// only record the rego version if it's different from the one applied globally to the result bundle
		if v != ast.RegoUndefined {
			if regoVersion == ast.RegoUndefined {
				// We store the rego version by the absolute path to the bundle root, as this will be the - possibly new - path
				// to the module inside the merged bundle.
				fileRegoVersions[bundleAbsolutePath(m, usePath)] = v.Int()
			} else {
				vInt := v.Int()
				gVInt := regoVersion.Int()
				if vInt != gVInt {
					fileRegoVersions[bundleAbsolutePath(m, usePath)] = vInt
				}
			}
		}
	}

	return fileRegoVersions, nil
}

func bundleRelativePath(m ModuleFile, usePath bool) string {
	p := m.RelativePath
	if p == "" {
		if usePath {
			p = m.Path
		} else {
			p = m.URL
		}
	}
	return p
}

func bundleAbsolutePath(m ModuleFile, usePath bool) string {
	var p string
	if usePath {
		p = m.Path
	} else {
		p = m.URL
	}
	if !path.IsAbs(p) {
		p = "/" + p
	}
	return path.Clean(p)
}

// RootPathsOverlap takes in two bundle root paths and returns true if they overlap.
func RootPathsOverlap(pathA string, pathB string) bool {
	a := rootPathSegments(pathA)
	b := rootPathSegments(pathB)
	return rootContains(a, b) || rootContains(b, a)
}

// RootPathsContain takes a set of bundle root paths and returns true if the path is contained.
func RootPathsContain(roots []string, path string) bool {
	segments := rootPathSegments(path)
	for i := range roots {
		if rootContains(rootPathSegments(roots[i]), segments) {
			return true
		}
	}
	return false
}

func rootPathSegments(path string) []string {
	return strings.Split(path, "/")
}

func rootContains(root []string, other []string) bool {

	// A single segment, empty string root always contains the other.
	if len(root) == 1 && root[0] == "" {
		return true
	}

	if len(root) > len(other) {
		return false
	}

	for j := range root {
		if root[j] != other[j] {
			return false
		}
	}

	return true
}

func insertValue(b *Bundle, path string, value interface{}) error {
	if err := b.insertData(getNormalizedPath(path), value); err != nil {
		return fmt.Errorf("bundle load failed on %v: %w", path, err)
	}
	return nil
}

func getNormalizedPath(path string) []string {
	// Remove leading / and . characters from the directory path. If the bundle
	// was written with OPA then the paths will contain a leading slash. On the
	// other hand, if the path is empty, filepath.Dir will return '.'.
	// Note: filepath.Dir can return paths with '\' separators, always use
	// filepath.ToSlash to keep them normalized.
	dirpath := strings.TrimLeft(normalizePath(filepath.Dir(path)), "/.")
	var key []string
	if dirpath != "" {
		key = strings.Split(dirpath, "/")
	}
	return key
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

	// Note: filepath.Join can return paths with '\' separators, always use
	// filepath.ToSlash to keep them normalized.
	return normalizePath(filepath.Join(prefix, modulePath))
}

// IsStructuredDoc checks if the file name equals a structured file extension ex. ".json"
func IsStructuredDoc(name string) bool {
	return filepath.Base(name) == dataFile || filepath.Base(name) == yamlDataFile ||
		filepath.Base(name) == SignaturesFile || filepath.Base(name) == ManifestExt
}

func preProcessBundle(loader DirectoryLoader, skipVerify bool, sizeLimitBytes int64) (SignaturesConfig, Patch, []*Descriptor, error) {
	descriptors := []*Descriptor{}
	var signatures SignaturesConfig
	var patch Patch

	for {
		f, err := loader.NextFile()
		if err == io.EOF {
			break
		}

		if err != nil {
			return signatures, patch, nil, fmt.Errorf("bundle read failed: %w", err)
		}

		// check for the signatures file
		if !skipVerify && strings.HasSuffix(f.Path(), SignaturesFile) {
			buf, err := readFile(f, sizeLimitBytes)
			if err != nil {
				return signatures, patch, nil, err
			}

			if err := util.NewJSONDecoder(&buf).Decode(&signatures); err != nil {
				return signatures, patch, nil, fmt.Errorf("bundle load failed on signatures decode: %w", err)
			}
		} else if !strings.HasSuffix(f.Path(), SignaturesFile) {
			descriptors = append(descriptors, f)

			if filepath.Base(f.Path()) == patchFile {

				var b bytes.Buffer
				tee := io.TeeReader(f.reader, &b)
				f.reader = tee

				buf, err := readFile(f, sizeLimitBytes)
				if err != nil {
					return signatures, patch, nil, err
				}

				if err := util.NewJSONDecoder(&buf).Decode(&patch); err != nil {
					return signatures, patch, nil, fmt.Errorf("bundle load failed on patch decode: %w", err)
				}

				f.reader = &b
			}
		}
	}
	return signatures, patch, descriptors, nil
}

func readFile(f *Descriptor, sizeLimitBytes int64) (bytes.Buffer, error) {
	// Case for pre-loaded byte buffers, like those from the tarballLoader.
	if bb, ok := f.reader.(*bytes.Buffer); ok {
		_ = f.Close() // always close, even on error

		if int64(bb.Len()) >= sizeLimitBytes {
			return *bb, fmt.Errorf("bundle file '%v' size (%d bytes) exceeded max size (%v bytes)",
				strings.TrimPrefix(f.Path(), "/"), bb.Len(), sizeLimitBytes-1)
		}

		return *bb, nil
	}

	// Case for *lazyFile readers:
	if lf, ok := f.reader.(*lazyFile); ok {
		var buf bytes.Buffer
		if lf.file == nil {
			var err error
			if lf.file, err = os.Open(lf.path); err != nil {
				return buf, fmt.Errorf("failed to open file %s: %w", f.path, err)
			}
		}
		// Bail out if we can't read the whole file-- there's nothing useful we can do at that point!
		fileSize, _ := fstatFileSize(lf.file)
		if fileSize > sizeLimitBytes {
			return buf, fmt.Errorf(maxSizeLimitBytesErrMsg, strings.TrimPrefix(f.Path(), "/"), fileSize, sizeLimitBytes-1)
		}
		// Prealloc the buffer for the file read.
		buffer := make([]byte, fileSize)
		_, err := io.ReadFull(lf.file, buffer)
		if err != nil {
			return buf, err
		}
		_ = lf.file.Close() // always close, even on error

		// Note(philipc): Replace the lazyFile reader in the *Descriptor with a
		// pointer to the wrapping bytes.Buffer, so that we don't re-read the
		// file on disk again by accident.
		buf = *bytes.NewBuffer(buffer)
		f.reader = &buf
		return buf, nil
	}

	// Fallback case:
	var buf bytes.Buffer
	n, err := f.Read(&buf, sizeLimitBytes)
	_ = f.Close() // always close, even on error

	if err != nil && err != io.EOF {
		return buf, err
	} else if err == nil && n >= sizeLimitBytes {
		return buf, fmt.Errorf(maxSizeLimitBytesErrMsg, strings.TrimPrefix(f.Path(), "/"), n, sizeLimitBytes-1)
	}

	return buf, nil
}

// Takes an already open file handle and invokes the os.Stat system call on it
// to determine the file's size. Passes any errors from *File.Stat on up to the
// caller.
func fstatFileSize(f *os.File) (int64, error) {
	fileInfo, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func normalizePath(p string) string {
	return filepath.ToSlash(p)
}
