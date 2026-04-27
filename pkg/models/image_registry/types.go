/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package image_registry

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RegistryType represents the type of image registry
type RegistryType string

const (
	RegistryTypePublic  RegistryType = "public"
	RegistryTypePrivate RegistryType = "private"
)

// Registry represents a configured image registry
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistrySpec   `json:"spec,omitempty"`
	Status RegistryStatus `json:"status,omitempty"`
}

// RegistrySpec defines the specification of a registry
type RegistrySpec struct {
	Type        RegistryType             `json:"type"`                   // "public" or "private"
	Domain      string                   `json:"domain"`                 // Registry address (e.g., Docker Hub, Harbor)
	Host        string                   `json:"host,omitempty"`         // API load balancer address (may differ from Domain)
	Insecure    bool                     `json:"insecure,omitempty"`     // Skip SSL verification
	PullSecret  *corev1.SecretReference `json:"pullSecret,omitempty"`   // K8s Secret reference for pulling
	PushSecret  *corev1.SecretReference `json:"pushSecret,omitempty"`   // K8s Secret reference for pushing
	SyncClusters []string               `json:"syncClusters,omitempty"` // List of clusters to sync with
}

// RegistryStatus represents the status of a registry
type RegistryStatus struct {
	Connected       bool   `json:"connected"`                  // Connection status
	LastCheckTime   string `json:"lastCheckTime,omitempty"`     // Last check timestamp
	RepositoryCount int    `json:"repositoryCount,omitempty"`  // Number of Images in the repository
}

// ImageMetadata holds image metadata
type ImageMetadata struct {
	Digest         string    `json:"digest"`                    // Image digest
	ManifestDigest string    `json:"manifestDigest,omitempty"`  // Manifest digest
	Size           int64     `json:"size"`                      // Image size in bytes
	Created        string    `json:"created,omitempty"`         // Creation timestamp
	Tags           []string  `json:"tags"`                      // List of tags
	Layers         []string  `json:"layers,omitempty"`          // Layer digests
	Architecture   string    `json:"architecture,omitempty"`    // Image architecture (amd64, arm64, etc.)
	OS             string    `json:"os,omitempty"`              // Operating system
}

// ImageSearchFilter defines filter criteria for searching images
type ImageSearchFilter struct {
	RegistryName string    `json:"registryName,omitempty"`   // Registry name to search in
	Query        string    `json:"query,omitempty"`          // Search query
	Tags         []string  `json:"tags,omitempty"`           // Filter by specific tags
	MinSize      int64     `json:"minSize,omitempty"`        // Minimum image size
	MaxSize      int64     `json:"maxSize,omitempty"`        // Maximum image size
	CreatedAfter string    `json:"createdAfter,omitempty"`   // Created after this time
	CreatedBefore string   `json:"createdBefore,omitempty"`   // Created before this time
	Limit        int       `json:"limit,omitempty"`          // Maximum number of results
	Offset       int       `json:"offset,omitempty"`         // Result offset for pagination
}

// ImageUploadRequest represents an image upload request
type ImageUploadRequest struct {
	SourceImage  string   `json:"sourceImage"`            // Source image name (e.g., foo:bar)
	TargetImage  string   `json:"targetImage"`            // Target image name (e.g., org/app:tag)
	TargetRegistries []string `json:"targetRegistries"`   // Target registry names
	SyncClusters []string `json:"syncClusters,omitempty"` // Clusters to sync with
	DryRun       bool     `json:"dryRun,omitempty"`       // Validate without actually pushing
	Force        bool     `json:"force,omitempty"`        // Force overwrite existing images
}

// ImageUploadTask represents an upload task
type ImageUploadTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageUploadTaskSpec   `json:"spec,omitempty"`
	Status ImageUploadTaskStatus `json:"status,omitempty"`
}

// ImageUploadTaskSpec defines the specification of an upload task
type ImageUploadTaskSpec struct {
	SourceImage   string   `json:"sourceImage"`            // Source image name (e.g., foo:bar)
	TargetRegistry string   `json:"targetRegistry"`        // Target registry name
	TargetImage    string   `json:"targetImage"`           // Target image name (e.g., org/app:tag)
	SyncClusters   []string `json:"syncClusters,omitempty"` // Clusters to sync with
	DryRun         bool     `json:"dryRun,omitempty"`      // Validate without actually pushing
}

// ImageUploadTaskStatus represents the status of an upload task
type ImageUploadTaskStatus struct {
	Phase          string           `json:"phase"`                    // "pending", "Running", "Completed", "Failed"
	Message        string           `json:"message,omitempty"`        // Status message
	StartTime      string           `json:"startTime,omitempty"`      // Task start time
	CompletionTime string           `json:"completionTime,omitempty"` // Task completion time
	ClusterResults []ClusterResult `json:"clusterResults,omitempty"` // Push status for each cluster
}

// ClusterResult holds the result of an image push to a cluster
type ClusterResult struct {
	ClusterName string `json:"clusterName"` // Cluster name
	Status      string `json:"status"`      // "pending", "Completed", "Failed"
	Message     string `json:"message,omitempty"` // Status message
}

// GlobalImageStore defines a global image store configuration
type GlobalImageStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalImageStoreSpec   `json:"spec,omitempty"`
	Status GlobalImageStoreStatus `json:"status,omitempty"`
}

// GlobalImageStoreSpec defines the specification of a global image store
type GlobalImageStoreSpec struct {
	Registries      []string `json:"registries,omitempty"`      // List of registry names to include in the global store
	SyncStrategy    string   `json:"syncStrategy,omitempty"`    // Sync strategy: "periodic", "on-demand", "manual"
	SyncInterval    string   `json:"syncInterval,omitempty"`    // Sync interval (e.g., "1h", "30m")
	ConflictSolution string  `json:"conflictSolution,omitempty"` // Conflict resolution: "last-write-wins", "merge"
}

// GlobalImageStoreStatus represents the status of a global image store
type GlobalImageStoreStatus struct {
	LastSyncTime      string   `json:"lastSyncTime,omitempty"`      // Last successful sync time
	TotalImages       int      `json:"totalImages"`                 // Total number of images across all registries
	SyncFailed        bool     `json:"syncFailed,omitempty"`        // Whether last sync failed
	SyncError         string   `json:"syncError,omitempty"`         // Error message if sync failed
	AvailableClusters []string `json:"availableClusters,omitempty"` // List of clusters the store is available on
}

// ImageSyncRequest defines a multi-cluster image sync request
type ImageSyncRequest struct {
	ImageName    string   `json:"imageName"`              // Image name to sync
	SourceCluster string  `json:"sourceCluster"`          // Source cluster name
	TargetClusters []string `json:"targetClusters"`       // Target cluster names
	DryRun       bool     `json:"dryRun,omitempty"`       // Validate without actually syncing
}

// ImageSyncStatus tracks the sync status of an image across clusters
type ImageSyncStatus struct {
	ImageName         string         `json:"imageName"`                // Image name
	SyncProgress      float64        `json:"syncProgress"`             // Sync progress (0.0 to 1.0)
	ClusterStatus    []ClusterStatus `json:"clusterStatus"`            // Status per cluster
	LastError        string         `json:"lastError,omitempty"`      // Last error message
	LastSyncTime     string         `json:"lastSyncTime,omitempty"`   // Last successful sync time
}

// ClusterStatus holds the image sync status for a specific cluster
type ClusterStatus struct {
	ClusterName  string `json:"clusterName"`  // Cluster name
	Status       string `json:"status"`       // "pending", "Syncing", "Completed", "Failed"
	Digest       string `json:"digest,omitempty"` // Image digest in this cluster
	Message      string `json:"message,omitempty"` // Status message
}

// ImageDetails holds details about an image (compatible with existing registries package)
type ImageDetails struct {
	ImageTag      string `json:"ImageTag"`
	ImageDigest   string `json:"ImageDigest,omitempty"`
	ImageManifest struct {
		ManifestConfig struct {
			Digest string `json:"Digest"`
		} `json:"ManifestConfig"`
	} `json:"ImageManifest,omitempty"`
	ImageBlob struct {
		Size    string `json:"Size"`
		Created string `json:"Created,omitempty"`
	} `json:"ImageBlob,omitempty"`
}
