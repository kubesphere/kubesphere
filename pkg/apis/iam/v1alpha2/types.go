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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// User is the Schema for the users API
// +kubebuilder:printcolumn:name="Email",type="string",JSONPath=".spec.email"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:resource:categories="iam",scope="Cluster"
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserSpec `json:"spec"`
	// +optional
	Status UserStatus `json:"status,omitempty"`
}

type FinalizerName string

// UserSpec defines the desired state of User
type UserSpec struct {
	// Unique email address.
	Email string `json:"email"`
	// The preferred written or spoken language for the user.
	// +optional
	Lang string `json:"lang,omitempty"`
	// Description of the user.
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	DisplayName string `json:"displayName,omitempty"`
	// +optional
	Groups            []string `json:"groups,omitempty"`
	EncryptedPassword string   `json:"password"`

	// Finalizers is an opaque list of values that must be empty to permanently remove object from storage.
	// +optional
	Finalizers []FinalizerName `json:"finalizers,omitempty"`
}

type UserState string

// These are the valid phases of a user.
const (
	// UserActive means the user is available.
	UserActive UserState = "Active"
	// UserDisabled means the user is disabled.
	UserDisabled UserState = "Disabled"
)

// UserStatus defines the observed state of User
type UserStatus struct {
	// The user status
	// +optional
	State UserState `json:"state,omitempty"`

	// Represents the latest available observations of a namespace's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []UserCondition `json:"conditions,omitempty"`
}

type UserCondition struct {
	// Type of namespace controller condition.
	Type UserConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

type UserConditionType string

// These are valid conditions of a user.
const (
	// UserLoginFailure contains information about user login.
	UserLoginFailure UserConditionType = "UserLoginFailure"
)

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue  ConditionStatus = "True"
	ConditionFalse ConditionStatus = "False"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".target.scope"
// +kubebuilder:printcolumn:name="Target",type="string",JSONPath=".target.name"
// +kubebuilder:resource:categories="iam",scope="Cluster"
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Target Target    `json:"target"`
	Rules  []RuleRef `json:"rules"`
}

type Target struct {
	Scope Scope  `json:"scope"`
	Name  string `json:"name"`
}

type Scope string

const (
	GlobalScope     Scope = "Global"
	ClusterScope    Scope = "Cluster"
	WorkspaceScope  Scope = "Workspace"
	NamespaceScope  Scope = "Namespace"
	UserKind              = "User"
	PolicyRuleKind        = "PolicyRule"
	RoleKind              = "Role"
	RoleBindingKind       = "RoleBinding"
)

// RuleRef contains information that points to the role being used
type RuleRef struct {
	// APIGroup is the group for the resource being referenced
	APIGroup string `json:"apiGroup"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind"`
	// Name is the name of resource being referenced
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoleList contains a list of Role
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".scope"
// +kubebuilder:resource:categories="iam",scope="Cluster"
type PolicyRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Scope Scope  `json:"scope"`
	Rego  string `json:"rego"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyRuleList contains a list of PolicyRule
type PolicyRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyRule `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoleBinding is the Schema for the rolebindings API
// +kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".scope"
// +kubebuilder:printcolumn:name="RoleRef",type="string",JSONPath=".roleRef.name"
// +kubebuilder:printcolumn:name="Subjects",type="string",JSONPath=".subjects[*].name"
// +kubebuilder:resource:categories="iam",scope="Cluster"
type RoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Scope   Scope   `json:"scope"`
	RoleRef RoleRef `json:"roleRef"`
	// Subjects holds references to the users the role applies to.
	// +optional
	Subjects []Subject `json:"subjects,omitempty"`
}

// RoleRef contains information that points to the role being used
type RoleRef struct {
	// APIGroup is the group for the resource being referenced
	APIGroup string `json:"apiGroup"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind"`
	// Name is the name of resource being referenced
	Name string `json:"name"`
}

// or a value for non-objects such as user and group names.
type Subject struct {
	// Kind of object being referenced. Values defined by this API group are "User", "Group", and "ServiceAccount".
	// If the Authorizer does not recognized the kind value, the Authorizer should report an error.
	Kind string `json:"kind"`
	// APIGroup holds the API group of the referenced subject.
	APIGroup string `json:"apiGroup"`
	// Name of the object being referenced.
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoleBindingList contains a list of RoleBinding
type RoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleBinding `json:"items"`
}

type UserDetail struct {
	*User
	GlobalRole *Role `json:"globalRole"`
}
