package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ServiceAccountName            = "kubesphere.io/service-account.name"
	ServiceAccountUID             = "kubesphere.io/service-account.uid"
	ServiceAccountToken           = "token"
	SecretTypeServiceAccountToken = "kubesphere.io/service-account-token"

	ServiceAccountGroup                     = "kubesphere:serviceaccount"
	ServiceAccountTokenPrefix               = ServiceAccountGroup + ":"
	ServiceAccountTokenSubFormat            = ServiceAccountTokenPrefix + "%s:%s"
	ServiceAccountTokenExtraSecretNamespace = "secret-namespace"
	ServiceAccountTokenExtraSecretName      = "secret-name"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Namespaced"

type ServiceAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Secrets []corev1.ObjectReference `json:"secrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// +kubebuilder:object:root=true

type ServiceAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccount `json:"items"`
}
