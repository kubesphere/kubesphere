// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package bundle implements bundle loading.
package bundle

import (
	"io"

	"github.com/open-policy-agent/opa/ast"
	v1 "github.com/open-policy-agent/opa/v1/bundle"
)

// Common file extensions and file names.
const (
	RegoExt        = v1.RegoExt
	WasmFile       = v1.WasmFile
	PlanFile       = v1.PlanFile
	ManifestExt    = v1.ManifestExt
	SignaturesFile = v1.SignaturesFile

	DefaultSizeLimitBytes = v1.DefaultSizeLimitBytes
	DeltaBundleType       = v1.DeltaBundleType
	SnapshotBundleType    = v1.SnapshotBundleType
)

// Bundle represents a loaded bundle. The bundle can contain data and policies.
type Bundle = v1.Bundle

// Raw contains raw bytes representing the bundle's content
type Raw = v1.Raw

// Patch contains an array of objects wherein each object represents the patch operation to be
// applied to the bundle data.
type Patch = v1.Patch

// PatchOperation models a single patch operation against a document.
type PatchOperation = v1.PatchOperation

// SignaturesConfig represents an array of JWTs that encapsulate the signatures for the bundle.
type SignaturesConfig = v1.SignaturesConfig

// DecodedSignature represents the decoded JWT payload.
type DecodedSignature = v1.DecodedSignature

// FileInfo contains the hashing algorithm used, resulting digest etc.
type FileInfo = v1.FileInfo

// NewFile returns a new FileInfo.
func NewFile(name, hash, alg string) FileInfo {
	return v1.NewFile(name, hash, alg)
}

// Manifest represents the manifest from a bundle. The manifest may contain
// metadata such as the bundle revision.
type Manifest = v1.Manifest

// WasmResolver maps a wasm module to an entrypoint ref.
type WasmResolver = v1.WasmResolver

// ModuleFile represents a single module contained in a bundle.
type ModuleFile = v1.ModuleFile

// WasmModuleFile represents a single wasm module contained in a bundle.
type WasmModuleFile = v1.WasmModuleFile

// PlanModuleFile represents a single plan module contained in a bundle.
//
// NOTE(tsandall): currently the plans are just opaque binary blobs. In the
// future we could inject the entrypoints so that the plans could be executed
// inside of OPA proper like we do for Wasm modules.
type PlanModuleFile = v1.PlanModuleFile

// Reader contains the reader to load the bundle from.
type Reader = v1.Reader

// NewReader is deprecated. Use NewCustomReader instead.
func NewReader(r io.Reader) *Reader {
	return v1.NewReader(r).WithRegoVersion(ast.DefaultRegoVersion)
}

// NewCustomReader returns a new Reader configured to use the
// specified DirectoryLoader.
func NewCustomReader(loader DirectoryLoader) *Reader {
	return v1.NewCustomReader(loader).WithRegoVersion(ast.DefaultRegoVersion)
}

// Write is deprecated. Use NewWriter instead.
func Write(w io.Writer, bundle Bundle) error {
	return v1.Write(w, bundle)
}

// Writer implements bundle serialization.
type Writer = v1.Writer

// NewWriter returns a bundle writer that writes to w.
func NewWriter(w io.Writer) *Writer {
	return v1.NewWriter(w)
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
	if regoVersion == ast.RegoUndefined {
		regoVersion = ast.DefaultRegoVersion
	}

	return v1.MergeWithRegoVersion(bundles, regoVersion, usePath)
}

// RootPathsOverlap takes in two bundle root paths and returns true if they overlap.
func RootPathsOverlap(pathA string, pathB string) bool {
	return v1.RootPathsOverlap(pathA, pathB)
}

// RootPathsContain takes a set of bundle root paths and returns true if the path is contained.
func RootPathsContain(roots []string, path string) bool {
	return v1.RootPathsContain(roots, path)
}
