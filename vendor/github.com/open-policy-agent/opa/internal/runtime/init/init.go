// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package init is an internal package with helpers for data and policy loading during initialization.
package init

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	storedversion "github.com/open-policy-agent/opa/internal/version"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/storage"
)

// InsertAndCompileOptions contains the input for the operation.
type InsertAndCompileOptions struct {
	Store                 storage.Store
	Txn                   storage.Transaction
	Files                 loader.Result
	Bundles               map[string]*bundle.Bundle
	MaxErrors             int
	EnablePrintStatements bool
	ParserOptions         ast.ParserOptions
}

// InsertAndCompileResult contains the output of the operation.
type InsertAndCompileResult struct {
	Compiler *ast.Compiler
	Metrics  metrics.Metrics
}

// InsertAndCompile writes data and policy into the store and returns a compiler for the
// store contents.
func InsertAndCompile(ctx context.Context, opts InsertAndCompileOptions) (*InsertAndCompileResult, error) {
	if len(opts.Files.Documents) > 0 {
		if err := opts.Store.Write(ctx, opts.Txn, storage.AddOp, storage.Path{}, opts.Files.Documents); err != nil {
			return nil, fmt.Errorf("storage error: %w", err)
		}
	}

	policies := make(map[string]*ast.Module, len(opts.Files.Modules))

	for id, parsed := range opts.Files.Modules {
		policies[id] = parsed.Parsed
	}

	compiler := ast.NewCompiler().
		SetErrorLimit(opts.MaxErrors).
		WithPathConflictsCheck(storage.NonEmpty(ctx, opts.Store, opts.Txn)).
		WithEnablePrintStatements(opts.EnablePrintStatements)
	m := metrics.New()

	activation := &bundle.ActivateOpts{
		Ctx:           ctx,
		Store:         opts.Store,
		Txn:           opts.Txn,
		Compiler:      compiler,
		Metrics:       m,
		Bundles:       opts.Bundles,
		ExtraModules:  policies,
		ParserOptions: opts.ParserOptions,
	}

	err := bundle.Activate(activation)
	if err != nil {
		return nil, err
	}

	// Policies in bundles will have already been added to the store, but
	// modules loaded outside of bundles will need to be added manually.
	for id, parsed := range opts.Files.Modules {
		if err := opts.Store.UpsertPolicy(ctx, opts.Txn, id, parsed.Raw); err != nil {
			return nil, fmt.Errorf("storage error: %w", err)
		}
	}

	// Set the version in the store last to prevent data files from overwriting.
	if err := storedversion.Write(ctx, opts.Store, opts.Txn); err != nil {
		return nil, fmt.Errorf("storage error: %w", err)
	}

	return &InsertAndCompileResult{Compiler: compiler, Metrics: m}, nil
}

// LoadPathsResult contains the output loading a set of paths.
type LoadPathsResult struct {
	Bundles map[string]*bundle.Bundle
	Files   loader.Result
}

// WalkPathsResult contains the output loading a set of paths.
type WalkPathsResult struct {
	BundlesLoader   []BundleLoader
	FileDescriptors []*Descriptor
}

// BundleLoader contains information about files in a bundle
type BundleLoader struct {
	DirectoryLoader bundle.DirectoryLoader
	IsDir           bool
}

// Descriptor contains information about a file
type Descriptor struct {
	Root string
	Path string
}

// LoadPaths reads data and policy from the given paths and returns a set of bundles or
// raw loader file results.
func LoadPaths(paths []string,
	filter loader.Filter,
	asBundle bool,
	bvc *bundle.VerificationConfig,
	skipVerify bool,
	processAnnotations bool,
	caps *ast.Capabilities,
	fsys fs.FS) (*LoadPathsResult, error) {
	return LoadPathsForRegoVersion(ast.RegoV0, paths, filter, asBundle, bvc, skipVerify, processAnnotations, false, caps, fsys)
}

func LoadPathsForRegoVersion(regoVersion ast.RegoVersion,
	paths []string,
	filter loader.Filter,
	asBundle bool,
	bvc *bundle.VerificationConfig,
	skipVerify bool,
	processAnnotations bool,
	followSymlinks bool,
	caps *ast.Capabilities,
	fsys fs.FS) (*LoadPathsResult, error) {

	if caps == nil {
		caps = ast.CapabilitiesForThisVersion()
	}

	// tar.gz files are automatically loaded as bundles
	var likelyBundles, nonBundlePaths []string
	if !asBundle {
		likelyBundles, nonBundlePaths = splitByTarGzExt(paths)
		paths = likelyBundles
	}

	var result LoadPathsResult
	var err error
	if asBundle || len(likelyBundles) > 0 {
		result.Bundles = make(map[string]*bundle.Bundle, len(paths))
		for _, path := range paths {
			result.Bundles[path], err = loader.NewFileLoader().
				WithFS(fsys).
				WithBundleVerificationConfig(bvc).
				WithSkipBundleVerification(skipVerify).
				WithFilter(filter).
				WithProcessAnnotation(processAnnotations).
				WithCapabilities(caps).
				WithRegoVersion(regoVersion).
				WithFollowSymlinks(followSymlinks).
				AsBundle(path)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(nonBundlePaths) == 0 {
		return &result, nil
	}

	files, err := loader.NewFileLoader().
		WithFS(fsys).
		WithProcessAnnotation(processAnnotations).
		WithCapabilities(caps).
		WithRegoVersion(regoVersion).
		Filtered(nonBundlePaths, filter)

	if err != nil {
		return nil, err
	}

	result.Files = *files

	return &result, nil
}

// splitByTarGzExt splits the paths in 2 groups. Ones with .tar.gz and another with
// non .tar.gz extensions.
func splitByTarGzExt(paths []string) (targzs []string, nonTargzs []string) {
	for _, path := range paths {
		if strings.HasSuffix(path, ".tar.gz") {
			targzs = append(targzs, path)
		} else {
			nonTargzs = append(nonTargzs, path)
		}
	}
	return
}

// WalkPaths reads data and policy from the given paths and returns a set of bundle directory loaders
// or descriptors that contain information about files.
func WalkPaths(paths []string, filter loader.Filter, asBundle bool) (*WalkPathsResult, error) {

	var result WalkPathsResult

	if asBundle {
		result.BundlesLoader = make([]BundleLoader, len(paths))
		for i, path := range paths {
			bundleLoader, isDir, err := loader.GetBundleDirectoryLoader(path)
			if err != nil {
				return nil, err
			}

			result.BundlesLoader[i] = BundleLoader{
				DirectoryLoader: bundleLoader,
				IsDir:           isDir,
			}
		}
		return &result, nil
	}

	result.FileDescriptors = []*Descriptor{}
	for _, path := range paths {
		filePaths, err := loader.FilteredPaths([]string{path}, filter)
		if err != nil {
			return nil, err
		}

		for _, fp := range filePaths {
			// Trim off the root directory and return path as if chrooted
			cleanedPath := strings.TrimPrefix(fp, path)
			if path == "." && filepath.Base(fp) == bundle.ManifestExt {
				cleanedPath = fp
			}

			if !strings.HasPrefix(cleanedPath, "/") {
				cleanedPath = "/" + cleanedPath
			}

			result.FileDescriptors = append(result.FileDescriptors, &Descriptor{
				Root: path,
				Path: cleanedPath,
			})
		}
	}

	return &result, nil
}
