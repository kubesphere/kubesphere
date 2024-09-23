package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UpdateStrategy struct {
	RegistryPoll `json:"registryPoll,omitempty"`
}

type RegistryPoll struct {
	Interval metav1.Duration `json:"interval"`
}

type BasicAuth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type RepositorySpec struct {
	// DEPRECATED: the field will remove in future versions, please use url.
	Image          string          `json:"image,omitempty"`
	URL            string          `json:"url,omitempty"`
	Description    string          `json:"description,omitempty"`
	BasicAuth      *BasicAuth      `json:"basicAuth,omitempty"`
	UpdateStrategy *UpdateStrategy `json:"updateStrategy,omitempty"`
	// The caBundle (base64 string) is used in helmExecutor to verify the helm server.
	// +optional
	CABundle string `json:"caBundle,omitempty"`
	// --insecure-skip-tls-verify. default false
	Insecure bool `json:"insecure,omitempty"`
}

type RepositoryStatus struct {
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty'"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="extensions",scope="Cluster"

// Repository declared a docker image containing the extension helm chart.
// The extension manager controller will deploy and synchronizes the extensions from the image repository.
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec,omitempty"`
	Status RepositoryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}
