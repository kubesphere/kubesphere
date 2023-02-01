package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourcePluralGroup = "groups"
	GroupReferenceLabel = "iam.kubesphere.io/group-ref"
	GroupParent         = "iam.kubesphere.io/group-parent"
)

// GroupSpec defines the desired state of Group
type GroupSpec struct {
}

// GroupStatus defines the observed state of Group
type GroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:resource:categories="group",scope="Cluster"

// Group is the Schema for the groups API
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GroupSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +genclient:nonNamespaced

// GroupList contains a list of Group
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}
