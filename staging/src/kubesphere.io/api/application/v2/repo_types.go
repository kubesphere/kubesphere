package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/constants"
)

type RepoCredential struct {
	// chart repository username
	Username string `json:"username,omitempty"`
	// chart repository password
	Password string `json:"password,omitempty"`
	// identify HTTPS client using this SSL certificate file
	CertFile string `json:"certFile,omitempty"`
	// identify HTTPS client using this SSL key file
	KeyFile string `json:"keyFile,omitempty"`
	// verify certificates of HTTPS-enabled servers using this CA bundle
	CAFile string `json:"caFile,omitempty"`
	// skip tls certificate checks for the repository, default is ture
	InsecureSkipTLSVerify *bool `json:"insecureSkipTLSVerify,omitempty"`
}

// RepoSpec defines the desired state of Repo
type RepoSpec struct {
	Url         string         `json:"url"`
	Credential  RepoCredential `json:"credential,omitempty"`
	Description string         `json:"description,omitempty"`
	SyncPeriod  int            `json:"syncPeriod,omitempty"`
}

// RepoStatus defines the observed state of Repo
type RepoStatus struct {
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	State          string      `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,path=repos,shortName=repo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="url",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Repo is the Schema for the repoes API
type Repo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepoSpec   `json:"spec,omitempty"`
	Status RepoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepoList contains a list of Repo
type RepoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repo `json:"items"`
}

func (in *Repo) GetWorkspace() string {
	return getValue(in.Labels, constants.WorkspaceLabelKey)
}

func (in *Repo) GetCreator() string {
	return getValue(in.Annotations, constants.CreatorAnnotationKey)
}
