/*
Copyright 2023.

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

package v1beta1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type AggregationRoleTemplates struct {
	// TemplateNames select rules from RoleTemplate`s rules by RoleTemplate name
	TemplateNames []string `json:"templateNames,omitempty"`

	// Selector select rules from RoleTemplate`s rules by labels
	Selector metav1.LabelSelector `json:"selector,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=iam,scope=Cluster

// GlobalRole is the Schema for the globalroles API
type GlobalRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// AggregationRoleTemplates means which RoleTemplates are composed this Role
	AggregationRoleTemplates AggregationRoleTemplates `json:"aggregationRoleTemplates,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	Rules rbacv1.PolicyRule `json:"rules"`
}

//+kubebuilder:object:root=true

// GlobalRoleList contains a list of GlobalRole
type GlobalRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRole `json:"items"`
}
