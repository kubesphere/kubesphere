package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&Label{}, &LabelList{})
}

const (
	ClusterLabelIDsAnnotation = "cluster.kubesphere.io/label-ids"
	LabelFinalizer            = "finalizers.kubesphere.io/cluster-label"
	ClusterLabelFormat        = "label.cluster.kubesphere.io/%s"

	ResourcesPluralLabel = "labels"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Key",type=string,JSONPath=".spec.key"
// +kubebuilder:printcolumn:name="Value",type=string,JSONPath=".spec.value"

type Label struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec LabelSpec `json:"spec"`
}

type LabelSpec struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	// +optional
	BackgroundColor string `json:"backgroundColor,omitempty"`
	// +optional
	Clusters []string `json:"clusters,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type LabelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Label `json:"items"`
}
