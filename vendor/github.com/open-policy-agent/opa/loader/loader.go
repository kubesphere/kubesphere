// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package loader contains utilities for loading files into OPA.
package loader

import (
	"io/fs"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	v1 "github.com/open-policy-agent/opa/v1/loader"
)

// Result represents the result of successfully loading zero or more files.
type Result = v1.Result

// RegoFile represents the result of loading a single Rego source file.
type RegoFile = v1.RegoFile

// Filter defines the interface for filtering files during loading. If the
// filter returns true, the file should be excluded from the result.
type Filter = v1.Filter

// GlobExcludeName excludes files and directories whose names do not match the
// shell style pattern at minDepth or greater.
func GlobExcludeName(pattern string, minDepth int) Filter {
	return v1.GlobExcludeName(pattern, minDepth)
}

// FileLoader defines an interface for loading OPA data files
// and Rego policies.
type FileLoader = v1.FileLoader

// NewFileLoader returns a new FileLoader instance.
func NewFileLoader() FileLoader {
	return v1.NewFileLoader().WithRegoVersion(ast.DefaultRegoVersion)
}

// GetBundleDirectoryLoader returns a bundle directory loader which can be used to load
// files in the directory
func GetBundleDirectoryLoader(path string) (bundle.DirectoryLoader, bool, error) {
	return v1.GetBundleDirectoryLoader(path)
}

// GetBundleDirectoryLoaderWithFilter returns a bundle directory loader which can be used to load
// files in the directory after applying the given filter.
func GetBundleDirectoryLoaderWithFilter(path string, filter Filter) (bundle.DirectoryLoader, bool, error) {
	return v1.GetBundleDirectoryLoaderWithFilter(path, filter)
}

// GetBundleDirectoryLoaderFS returns a bundle directory loader which can be used to load
// files in the directory.
func GetBundleDirectoryLoaderFS(fsys fs.FS, path string, filter Filter) (bundle.DirectoryLoader, bool, error) {
	return v1.GetBundleDirectoryLoaderFS(fsys, path, filter)
}

// FilteredPaths is the same as FilterPathsFS using the current diretory file
// system
func FilteredPaths(paths []string, filter Filter) ([]string, error) {
	return v1.FilteredPaths(paths, filter)
}

// FilteredPathsFS return a list of files from the specified
// paths while applying the given filters. If any filter returns true, the
// file/directory is excluded.
func FilteredPathsFS(fsys fs.FS, paths []string, filter Filter) ([]string, error) {
	return v1.FilteredPathsFS(fsys, paths, filter)
}

// Schemas loads a schema set from the specified file path.
func Schemas(schemaPath string) (*ast.SchemaSet, error) {
	return v1.Schemas(schemaPath)
}

// All returns a Result object loaded (recursively) from the specified paths.
// Deprecated: Use FileLoader.Filtered() instead.
func All(paths []string) (*Result, error) {
	return NewFileLoader().Filtered(paths, nil)
}

// Filtered returns a Result object loaded (recursively) from the specified
// paths while applying the given filters. If any filter returns true, the
// file/directory is excluded.
// Deprecated: Use FileLoader.Filtered() instead.
func Filtered(paths []string, filter Filter) (*Result, error) {
	return NewFileLoader().Filtered(paths, filter)
}

// AsBundle loads a path as a bundle. If it is a single file
// it will be treated as a normal tarball bundle. If a directory
// is supplied it will be loaded as an unzipped bundle tree.
// Deprecated: Use FileLoader.AsBundle() instead.
func AsBundle(path string) (*bundle.Bundle, error) {
	return NewFileLoader().AsBundle(path)
}

// AllRegos returns a Result object loaded (recursively) with all Rego source
// files from the specified paths.
func AllRegos(paths []string) (*Result, error) {
	return NewFileLoader().Filtered(paths, func(_ string, info os.FileInfo, _ int) bool {
		return !info.IsDir() && !strings.HasSuffix(info.Name(), bundle.RegoExt)
	})
}

// Rego is deprecated. Use RegoWithOpts instead.
func Rego(path string) (*RegoFile, error) {
	return RegoWithOpts(path, ast.ParserOptions{})
}

// RegoWithOpts returns a RegoFile object loaded from the given path.
func RegoWithOpts(path string, opts ast.ParserOptions) (*RegoFile, error) {
	if opts.RegoVersion == ast.RegoUndefined {
		opts.RegoVersion = ast.DefaultRegoVersion
	}

	return v1.RegoWithOpts(path, opts)
}

// CleanPath returns the normalized version of a path that can be used as an identifier.
func CleanPath(path string) string {
	return v1.CleanPath(path)
}

// Paths returns a sorted list of files contained at path. If recurse is true
// and path is a directory, then Paths will walk the directory structure
// recursively and list files at each level.
func Paths(path string, recurse bool) (paths []string, err error) {
	return v1.Paths(path, recurse)
}

// Dirs resolves filepaths to directories. It will return a list of unique
// directories.
func Dirs(paths []string) []string {
	return v1.Dirs(paths)
}

// SplitPrefix returns a tuple specifying the document prefix and the file
// path.
func SplitPrefix(path string) ([]string, string) {
	return v1.SplitPrefix(path)
}
