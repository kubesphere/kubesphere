/*
Copyright 2019 The KubeSphere authors.

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

package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DevOpsProjectRoleJenkinsFinalizerName = "devopsprojectrole.finalizers.kubesphere.io/jenkins"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DevOpsProjectRoleSpec defines the desired state of DevOpsProjectRole
type DevOpsProjectRoleSpec struct {
	// Rules holds all the PolicyRules for this Role
	Rules []rbacv1.PolicyRule `json:"rules"`
}

// DevOpsProjectRoleStatus defines the observed state of DevOpsProjectRole
type DevOpsProjectRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// DevOpsProjectRole is the Schema for the devopsprojectroles API
// +k8s:openapi-gen=true
type DevOpsProjectRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DevOpsProjectRoleSpec   `json:"spec,omitempty"`
	Status DevOpsProjectRoleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// DevOpsProjectRoleList contains a list of DevOpsProjectRole
type DevOpsProjectRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DevOpsProjectRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DevOpsProjectRole{}, &DevOpsProjectRoleList{})
}
