/*
Copyright 2020 The KubeSphere Authors.

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

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindCluster      = "ResourceQuota"
	ResourcesSingularCluster = "resourcequota"
	ResourcesPluralCluster   = "resourcequotas"
)

func init() {
	SchemeBuilder.Register(&ResourceQuota{}, &ResourceQuotaList{})
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="quota",scope="Cluster",path=resourcequotas
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// WorkspaceResourceQuota sets aggregate quota restrictions enforced per workspace
type ResourceQuota struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired quota
	Spec ResourceQuotaSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// Status defines the actual enforced quota and its current usage
	// +optional
	Status ResourceQuotaStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ResourceQuotaSpec defines the desired quota restrictions
type ResourceQuotaSpec struct {
	// LabelSelector is used to select projects by label.
	LabelSelector map[string]string `json:"selector" protobuf:"bytes,1,opt,name=selector"`

	// Quota defines the desired quota
	Quota corev1.ResourceQuotaSpec `json:"quota" protobuf:"bytes,2,opt,name=quota"`
}

// ResourceQuotaStatus defines the actual enforced quota and its current usage
type ResourceQuotaStatus struct {
	// Total defines the actual enforced quota and its current usage across all projects
	Total corev1.ResourceQuotaStatus `json:"total" protobuf:"bytes,1,opt,name=total"`

	// Namespaces slices the usage by project.
	Namespaces ResourceQuotasStatusByNamespace `json:"namespaces" protobuf:"bytes,2,rep,name=namespaces"`
}

// ResourceQuotasStatusByNamespace bundles multiple ResourceQuotaStatusByNamespace
type ResourceQuotasStatusByNamespace []ResourceQuotaStatusByNamespace

// ResourceQuotaStatusByNamespace gives status for a particular project
type ResourceQuotaStatusByNamespace struct {
	corev1.ResourceQuotaStatus `json:",inline"`

	// Namespace the project this status applies to
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ResourceQuotaList is a list of WorkspaceResourceQuota items.
type ResourceQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of WorkspaceResourceQuota objects.
	// More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/
	Items []ResourceQuota `json:"items" protobuf:"bytes,2,rep,name=items"`
}
