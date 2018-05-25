/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package metadata

import (
	"github.com/boltdb/bolt"
	digest "github.com/opencontainers/go-digest"
)

// The layout where a "/" delineates a bucket is desribed in the following
// section. Please try to follow this as closely as possible when adding
// functionality. We can bolster this with helpers and more structure if that
// becomes an issue.
//
// Generically, we try to do the following:
//
// 	<version>/<namespace>/<object>/<key> -> <field>
//
// version: Currently, this is "v1". Additions can be made to v1 in a backwards
// compatible way. If the layout changes, a new version must be made, along
// with a migration.
//
// namespace: the namespace to which this object belongs.
//
// object: defines which object set is stored in the bucket. There are two
// special objects, "labels" and "indexes". The "labels" bucket stores the
// labels for the parent namespace. The "indexes" object is reserved for
// indexing objects, if we require in the future.
//
// key: object-specific key identifying the storage bucket for the objects
// contents.
var (
	bucketKeyVersion          = []byte(schemaVersion)
	bucketKeyDBVersion        = []byte("version")    // stores the version of the schema
	bucketKeyObjectLabels     = []byte("labels")     // stores the labels for a namespace.
	bucketKeyObjectImages     = []byte("images")     // stores image objects
	bucketKeyObjectContainers = []byte("containers") // stores container objects
	bucketKeyObjectSnapshots  = []byte("snapshots")  // stores snapshot references
	bucketKeyObjectContent    = []byte("content")    // stores content references
	bucketKeyObjectBlob       = []byte("blob")       // stores content links
	bucketKeyObjectIngest     = []byte("ingest")     // stores ingest links
	bucketKeyObjectLeases     = []byte("leases")     // stores leases

	bucketKeyDigest      = []byte("digest")
	bucketKeyMediaType   = []byte("mediatype")
	bucketKeySize        = []byte("size")
	bucketKeyImage       = []byte("image")
	bucketKeyRuntime     = []byte("runtime")
	bucketKeyName        = []byte("name")
	bucketKeyParent      = []byte("parent")
	bucketKeyChildren    = []byte("children")
	bucketKeyOptions     = []byte("options")
	bucketKeySpec        = []byte("spec")
	bucketKeySnapshotKey = []byte("snapshotKey")
	bucketKeySnapshotter = []byte("snapshotter")
	bucketKeyTarget      = []byte("target")
	bucketKeyExtensions  = []byte("extensions")
	bucketKeyCreatedAt   = []byte("createdat")
)

func getBucket(tx *bolt.Tx, keys ...[]byte) *bolt.Bucket {
	bkt := tx.Bucket(keys[0])

	for _, key := range keys[1:] {
		if bkt == nil {
			break
		}
		bkt = bkt.Bucket(key)
	}

	return bkt
}

func createBucketIfNotExists(tx *bolt.Tx, keys ...[]byte) (*bolt.Bucket, error) {
	bkt, err := tx.CreateBucketIfNotExists(keys[0])
	if err != nil {
		return nil, err
	}

	for _, key := range keys[1:] {
		bkt, err = bkt.CreateBucketIfNotExists(key)
		if err != nil {
			return nil, err
		}
	}

	return bkt, nil
}

func namespaceLabelsBucketPath(namespace string) [][]byte {
	return [][]byte{bucketKeyVersion, []byte(namespace), bucketKeyObjectLabels}
}

func withNamespacesLabelsBucket(tx *bolt.Tx, namespace string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, namespaceLabelsBucketPath(namespace)...)
	if err != nil {
		return err
	}

	return fn(bkt)
}

func getNamespaceLabelsBucket(tx *bolt.Tx, namespace string) *bolt.Bucket {
	return getBucket(tx, namespaceLabelsBucketPath(namespace)...)
}

func imagesBucketPath(namespace string) [][]byte {
	return [][]byte{bucketKeyVersion, []byte(namespace), bucketKeyObjectImages}
}

func createImagesBucket(tx *bolt.Tx, namespace string) (*bolt.Bucket, error) {
	return createBucketIfNotExists(tx, imagesBucketPath(namespace)...)
}

func getImagesBucket(tx *bolt.Tx, namespace string) *bolt.Bucket {
	return getBucket(tx, imagesBucketPath(namespace)...)
}

func createContainersBucket(tx *bolt.Tx, namespace string) (*bolt.Bucket, error) {
	bkt, err := createBucketIfNotExists(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectContainers)
	if err != nil {
		return nil, err
	}
	return bkt, nil
}

func getContainersBucket(tx *bolt.Tx, namespace string) *bolt.Bucket {
	return getBucket(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectContainers)
}

func getContainerBucket(tx *bolt.Tx, namespace, id string) *bolt.Bucket {
	return getBucket(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectContainers, []byte(id))
}

func createSnapshotterBucket(tx *bolt.Tx, namespace, snapshotter string) (*bolt.Bucket, error) {
	bkt, err := createBucketIfNotExists(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectSnapshots, []byte(snapshotter))
	if err != nil {
		return nil, err
	}
	return bkt, nil
}

func getSnapshottersBucket(tx *bolt.Tx, namespace string) *bolt.Bucket {
	return getBucket(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectSnapshots)
}

func getSnapshotterBucket(tx *bolt.Tx, namespace, snapshotter string) *bolt.Bucket {
	return getBucket(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectSnapshots, []byte(snapshotter))
}

func createBlobBucket(tx *bolt.Tx, namespace string, dgst digest.Digest) (*bolt.Bucket, error) {
	bkt, err := createBucketIfNotExists(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectContent, bucketKeyObjectBlob, []byte(dgst.String()))
	if err != nil {
		return nil, err
	}
	return bkt, nil
}

func getBlobsBucket(tx *bolt.Tx, namespace string) *bolt.Bucket {
	return getBucket(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectContent, bucketKeyObjectBlob)
}

func getBlobBucket(tx *bolt.Tx, namespace string, dgst digest.Digest) *bolt.Bucket {
	return getBucket(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectContent, bucketKeyObjectBlob, []byte(dgst.String()))
}

func createIngestBucket(tx *bolt.Tx, namespace string) (*bolt.Bucket, error) {
	bkt, err := createBucketIfNotExists(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectContent, bucketKeyObjectIngest)
	if err != nil {
		return nil, err
	}
	return bkt, nil
}

func getIngestBucket(tx *bolt.Tx, namespace string) *bolt.Bucket {
	return getBucket(tx, bucketKeyVersion, []byte(namespace), bucketKeyObjectContent, bucketKeyObjectIngest)
}
