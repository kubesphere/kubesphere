/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:prerelease:deprecated
// +kubebuilder:resource:categories=kubesphere
// +kubebuilder:resource:shortNames=reg
// +kubebuilder:subresource:status

// Registry is the Schema for the registries API
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistrySpec   `json:"spec,omitempty"`
	Status RegistryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RegistryList contains a list of Registry
type RegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Registry `json:"items"`
}

// RegistrySpec defines the desired state of Registry
type RegistrySpec struct {
	// Type of registry (public or private)
	// +kubebuilder:validation:Enum=public;private
	// +required
	Type string `json:"type"`

	// Registry address (e.g., Docker Hub, Harbor)
	// +required
	Domain string `json:"domain"`

	// API load balancer address (may differ from Domain)
	// +optional
	Host string `json:"host,omitempty"`

	// Skip SSL verification
	// +optional
	// +kubebuilder:default=false
	Insecure bool `json:"insecure,omitempty"`

	// K8s Secret reference for pulling
	// +optional
	PullSecret *corev1.SecretReference `json:"pullSecret,omitempty"`

	// K8s Secret reference for pushing
	// +optional
	PushSecret *corev1.SecretReference `json:"pushSecret,omitempty"`

	// List of clusters to sync with
	// +optional
	SyncClusters []string `json:"syncClusters,omitempty"`
}

// RegistryStatus defines the observed state of Registry
type RegistryStatus struct {
	// Connection status
	// +optional
	Connected bool `json:"connected"`

	// Last check timestamp
	// +optional
	LastCheckTime string `json:"lastCheckTime,omitempty"`

	// Number of images in the repository
	// +optional
	RepositoryCount int `json:"repositoryCount,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:prerelease:deprecated
// +kubebuilder:resource:categories=kubesphere
// +kubebuilder:resource:shortNames=uploadtask;upload
// +kubebuilder:subresource:status

// ImageUploadTask is the Schema for the imageuploadtasks API
type ImageUploadTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageUploadTaskSpec   `json:"spec,omitempty"`
	Status ImageUploadTaskStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImageUploadTaskList contains a list of ImageUploadTask
type ImageUploadTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageUploadTask `json:"items"`
}

// ImageUploadTaskSpec defines the desired state of ImageUploadTask
type ImageUploadTaskSpec struct {
	// Source image name (e.g., foo:bar)
	// +required
	SourceImage string `json:"sourceImage"`

	// Target registry name
	// +required
	TargetRegistry string `json:"targetRegistry"`

	// Target image name (e.g., org/app:tag)
	// +optional
	TargetImage string `json:"targetImage,omitempty"`

	// Clusters to sync with
	// +optional
	SyncClusters []string `json:"syncClusters,omitempty"`

	// Validate without actually pushing
	// +optional
	// +kubebuilder:default=false
	DryRun bool `json:"dryRun,omitempty"`
}

// ImageUploadTaskStatus defines the observed state of ImageUploadTask
type ImageUploadTaskStatus struct {
	// Phase of the task (pending, Running, Completed, Failed)
	// +kubebuilder:validation:Enum=pending;Running;Completed;Failed
	// +optional
	Phase string `json:"phase"`

	// Status message
	// +optional
	Message string `json:"message,omitempty"`

	// Task start time
	// +optional
	StartTime string `json:"startTime,omitempty"`

	// Task completion time
	// +optional
	CompletionTime string `json:"completionTime,omitempty"`

	// Push status for each cluster
	// +optional
	ClusterResults []ClusterResult `json:"clusterResults,omitempty"`
}

// ClusterResult holds the result of an image push to a cluster
type ClusterResult struct {
	// Cluster name
	ClusterName string `json:"clusterName"`

	// Status of the push (pending, Completed, Failed)
	// +kubebuilder:validation:Enum=pending;Completed;Failed
	Status string `json:"status"`

	// Status message
	// +optional
	Message string `json:"message,omitempty"`
}

// ClusterStatus holds sync status for a cluster
type ClusterStatus struct {
	ClusterName string `json:"clusterName,omitempty"`
	Status       string `json:"status,omitempty"`
	Digest       string `json:"digest,omitempty"`
	Message      string `json:"message,omitempty"`
}

// ImagesDetail holds detailed information about an image
type ImagesDetail struct {
	// Image digest
	Digest string `json:"digest,omitempty"`

	// Image size in bytes
	Size int64 `json:"size,omitempty"`

	// Creation time
	Created string `json:"created,omitempty"`

	// Creation time (alias)
	CreatedAt string `json:"createdAt,omitempty"`

	// List of tags
	Tags []string `json:"tags,omitempty"`

	// List of layer digests
	Layers []string `json:"layers,omitempty"`

	// Image architecture
	Arch string `json:"arch,omitempty"`

	// Operating system
	OS string `json:"os,omitempty"`
}

// ImageSyncStatus holds sync status for an image
type ImageSyncStatus struct {
	ImageName     string `json:"imageName,omitempty"`
	SyncProgress  float64 `json:"syncProgress,omitempty"`
	ClusterStatus []ClusterStatus `json:"clusterStatus,omitempty"`
	LastSyncTime  string `json:"lastSyncTime,omitempty"`
}

// DeepCopy generates a new Registry object
func (in *Registry) DeepCopy() *Registry {
	// Simplified for now
	return &Registry{Spec: in.Spec, Status: in.Status}
}

// DeepCopyObject implements the runtime.Object interface
func (in *Registry) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopy generates a new RegistryList object
func (in *RegistryList) DeepCopy() *RegistryList {
	// Simplified for now
	return &RegistryList{Items: in.Items}
}

// DeepCopyObject implements the runtime.Object interface
func (in *RegistryList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopy generates a new ImageUploadTask object
func (in *ImageUploadTask) DeepCopy() *ImageUploadTask {
	// Simplified for now
	return &ImageUploadTask{Spec: in.Spec, Status: in.Status}
}

// DeepCopyObject implements the runtime.Object interface
func (in *ImageUploadTask) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopy generates a new ImageUploadTaskList object
func (in *ImageUploadTaskList) DeepCopy() *ImageUploadTaskList {
	// Simplified for now
	return &ImageUploadTaskList{Items: in.Items}
}

// DeepCopyObject implements the runtime.Object interface
func (in *ImageUploadTaskList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}
