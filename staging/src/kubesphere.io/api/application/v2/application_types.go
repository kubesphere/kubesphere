package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/constants"
)

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	AppHome     string                 `json:"appHome,omitempty"`
	AppType     string                 `json:"appType,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Abstraction string                 `json:"abstraction,omitempty"`
	Attachments []string               `json:"attachments,omitempty"`
	Resources   []GroupVersionResource `json:"resources,omitempty"`
}

type GroupVersionResource struct {
	Group      string `json:"Group,omitempty"`
	Version    string `json:"Version,omitempty"`
	Resource   string `json:"Resource,omitempty"`
	Name       string `json:"Name,omitempty"`
	Desc       string `json:"Desc,omitempty"`
	ParentNode string `json:"ParentNode,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// the state of the helm application: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state,omitempty"`
	UpdateTime *metav1.Time `json:"updateTime"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=app
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="repo",type="string",JSONPath=".metadata.labels.application\\.kubesphere\\.io/repo-name"
// +kubebuilder:printcolumn:name="workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="appType",type="string",JSONPath=".spec.appType"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func getValue(m map[string]string, key string) string {
	if m == nil {
		return ""
	}
	return m[key]
}

func (in *Application) GetWorkspace() string {
	return getValue(in.Labels, constants.WorkspaceLabelKey)
}
