// Copyright 2019 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package bundle

import (
	"context"

	"github.com/open-policy-agent/opa/storage"
	v1 "github.com/open-policy-agent/opa/v1/bundle"
)

// BundlesBasePath is the storage path used for storing bundle metadata
var BundlesBasePath = v1.BundlesBasePath

// Note: As needed these helpers could be memoized.

// ManifestStoragePath is the storage path used for the given named bundle manifest.
func ManifestStoragePath(name string) storage.Path {
	return v1.ManifestStoragePath(name)
}

// EtagStoragePath is the storage path used for the given named bundle etag.
func EtagStoragePath(name string) storage.Path {
	return v1.EtagStoragePath(name)
}

// ReadBundleNamesFromStore will return a list of bundle names which have had their metadata stored.
func ReadBundleNamesFromStore(ctx context.Context, store storage.Store, txn storage.Transaction) ([]string, error) {
	return v1.ReadBundleNamesFromStore(ctx, store, txn)
}

// WriteManifestToStore will write the manifest into the storage. This function is called when
// the bundle is activated.
func WriteManifestToStore(ctx context.Context, store storage.Store, txn storage.Transaction, name string, manifest Manifest) error {
	return v1.WriteManifestToStore(ctx, store, txn, name, manifest)
}

// WriteEtagToStore will write the bundle etag into the storage. This function is called when the bundle is activated.
func WriteEtagToStore(ctx context.Context, store storage.Store, txn storage.Transaction, name, etag string) error {
	return v1.WriteEtagToStore(ctx, store, txn, name, etag)
}

// EraseManifestFromStore will remove the manifest from storage. This function is called
// when the bundle is deactivated.
func EraseManifestFromStore(ctx context.Context, store storage.Store, txn storage.Transaction, name string) error {
	return v1.EraseManifestFromStore(ctx, store, txn, name)
}

// ReadWasmModulesFromStore will write Wasm module resolver metadata from the store.
func ReadWasmModulesFromStore(ctx context.Context, store storage.Store, txn storage.Transaction, name string) (map[string][]byte, error) {
	return v1.ReadWasmModulesFromStore(ctx, store, txn, name)
}

// ReadBundleRootsFromStore returns the roots in the specified bundle.
// If the bundle is not activated, this function will return
// storage NotFound error.
func ReadBundleRootsFromStore(ctx context.Context, store storage.Store, txn storage.Transaction, name string) ([]string, error) {
	return v1.ReadBundleRootsFromStore(ctx, store, txn, name)
}

// ReadBundleRevisionFromStore returns the revision in the specified bundle.
// If the bundle is not activated, this function will return
// storage NotFound error.
func ReadBundleRevisionFromStore(ctx context.Context, store storage.Store, txn storage.Transaction, name string) (string, error) {
	return v1.ReadBundleRevisionFromStore(ctx, store, txn, name)
}

// ReadBundleMetadataFromStore returns the metadata in the specified bundle.
// If the bundle is not activated, this function will return
// storage NotFound error.
func ReadBundleMetadataFromStore(ctx context.Context, store storage.Store, txn storage.Transaction, name string) (map[string]interface{}, error) {
	return v1.ReadBundleMetadataFromStore(ctx, store, txn, name)
}

// ReadBundleEtagFromStore returns the etag for the specified bundle.
// If the bundle is not activated, this function will return
// storage NotFound error.
func ReadBundleEtagFromStore(ctx context.Context, store storage.Store, txn storage.Transaction, name string) (string, error) {
	return v1.ReadBundleEtagFromStore(ctx, store, txn, name)
}

// ActivateOpts defines options for the Activate API call.
type ActivateOpts = v1.ActivateOpts

// Activate the bundle(s) by loading into the given Store. This will load policies, data, and record
// the manifest in storage. The compiler provided will have had the polices compiled on it.
func Activate(opts *ActivateOpts) error {
	return v1.Activate(opts)
}

// DeactivateOpts defines options for the Deactivate API call
type DeactivateOpts = v1.DeactivateOpts

// Deactivate the bundle(s). This will erase associated data, policies, and the manifest entry from the store.
func Deactivate(opts *DeactivateOpts) error {
	return v1.Deactivate(opts)
}

// LegacyWriteManifestToStore will write the bundle manifest to the older single (unnamed) bundle manifest location.
// Deprecated: Use WriteManifestToStore and named bundles instead.
func LegacyWriteManifestToStore(ctx context.Context, store storage.Store, txn storage.Transaction, manifest Manifest) error {
	return v1.LegacyWriteManifestToStore(ctx, store, txn, manifest)
}

// LegacyEraseManifestFromStore will erase the bundle manifest from the older single (unnamed) bundle manifest location.
// Deprecated: Use WriteManifestToStore and named bundles instead.
func LegacyEraseManifestFromStore(ctx context.Context, store storage.Store, txn storage.Transaction) error {
	return v1.LegacyEraseManifestFromStore(ctx, store, txn)
}

// LegacyReadRevisionFromStore will read the bundle manifest revision from the older single (unnamed) bundle manifest location.
// Deprecated: Use ReadBundleRevisionFromStore and named bundles instead.
func LegacyReadRevisionFromStore(ctx context.Context, store storage.Store, txn storage.Transaction) (string, error) {
	return v1.LegacyReadRevisionFromStore(ctx, store, txn)
}

// ActivateLegacy calls Activate for the bundles but will also write their manifest to the older unnamed store location.
// Deprecated: Use Activate with named bundles instead.
func ActivateLegacy(opts *ActivateOpts) error {
	return v1.ActivateLegacy(opts)
}
