/*
Copyright 2019 The KubeSphere Authors.

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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindUser                      = "User"
	ResourcesSingularUser                 = "user"
	ResourcesPluralUser                   = "users"
	ResourceKindGlobalRoleBinding         = "GlobalRoleBinding"
	ResourcesSingularGlobalRoleBinding    = "globalrolebinding"
	ResourcesPluralGlobalRoleBinding      = "globalrolebindings"
	ResourceKindClusterRoleBinding        = "ClusterRoleBinding"
	ResourcesSingularClusterRoleBinding   = "clusterrolebinding"
	ResourcesPluralClusterRoleBinding     = "clusterrolebindings"
	ResourceKindRoleBinding               = "RoleBinding"
	ResourcesSingularRoleBinding          = "rolebinding"
	ResourcesPluralRoleBinding            = "rolebindings"
	ResourceKindGlobalRole                = "GlobalRole"
	ResourcesSingularGlobalRole           = "globalrole"
	ResourcesPluralGlobalRole             = "globalroles"
	ResourceKindWorkspaceRoleBinding      = "WorkspaceRoleBinding"
	ResourcesSingularWorkspaceRoleBinding = "workspacerolebinding"
	ResourcesPluralWorkspaceRoleBinding   = "workspacerolebindings"
	ResourceKindWorkspaceRole             = "WorkspaceRole"
	ResourcesSingularWorkspaceRole        = "workspacerole"
	ResourcesPluralWorkspaceRole          = "workspaceroles"
	ResourceKindClusterRole               = "ClusterRole"
	ResourcesSingularClusterRole          = "clusterrole"
	ResourcesPluralClusterRole            = "clusterroles"
	ResourceKindRole                      = "Role"
	ResourcesSingularRole                 = "role"
	ResourcesPluralRole                   = "roles"
	RegoOverrideAnnotation                = "iam.kubesphere.io/rego-override"
	AggregationRolesAnnotation            = "iam.kubesphere.io/aggregation-roles"
	GlobalRoleAnnotation                  = "iam.kubesphere.io/globalrole"
	WorkspaceRoleAnnotation               = "iam.kubesphere.io/workspacerole"
	ClusterRoleAnnotation                 = "iam.kubesphere.io/clusterrole"
	RoleAnnotation                        = "iam.kubesphere.io/role"
	RoleTemplateLabel                     = "iam.kubesphere.io/role-template"
	UserReferenceLabel                    = "iam.kubesphere.io/user-ref"
	IdentifyProviderLabel                 = "iam.kubesphere.io/identify-provider"
	PasswordEncryptedAnnotation           = "iam.kubesphere.io/password-encrypted"
	FieldEmail                            = "email"
	AggregateTo                           = "aggregateTo"
	ScopeWorkspace                        = "workspace"
	ScopeCluster                          = "cluster"
	ScopeNamespace                        = "namespace"
	LocalCluster                          = "local"
	GlobalAdmin                           = "global-admin"
	ClusterAdmin                          = "cluster-admin"
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
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserSpec `json:"spec"`
	// +optional
	Status UserStatus `json:"status,omitempty"`
}

type FinalizerName string

// UserSpec defines the desired state of User
type UserSpec struct {
	// Unique email address(https://www.ietf.org/rfc/rfc5322.txt).
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
	Groups []string `json:"groups,omitempty"`
	// password will be encrypted by mutating admission webhook
	EncryptedPassword string `json:"password,omitempty"`
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
	// Represents the latest available observations of a user's current state.
	// +optional
	Conditions []UserCondition `json:"conditions,omitempty"`
}

type UserCondition struct {
	// Type of user controller condition.
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
	LoginFailure UserConditionType = "LoginFailure"
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
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:resource:categories="iam",scope="Cluster"
type GlobalRole struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Rules holds all the PolicyRules for this GlobalRole
	// +optional
	Rules []rbacv1.PolicyRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalRoleList contains a list of GlobalRole
type GlobalRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRole `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalRoleBinding is the Schema for the globalrolebindings API
// +kubebuilder:resource:categories="iam",scope="Cluster"
type GlobalRoleBinding struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Subjects holds references to the objects the role applies to.
	// +optional
	Subjects []rbacv1.Subject `json:"subjects,omitempty" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef can only reference a GlobalRole.
	// If the RoleRef cannot be resolved, the Authorizer must return an error.
	RoleRef rbacv1.RoleRef `json:"roleRef" protobuf:"bytes,3,opt,name=roleRef"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalRoleBindingList contains a list of GlobalRoleBinding
type GlobalRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRoleBinding `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="Alias",type="string",JSONPath=".metadata.annotations.kubesphere\\.io/alias-name"
// +kubebuilder:resource:categories="iam",scope="Cluster"
type WorkspaceRole struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	// +optional
	Rules []rbacv1.PolicyRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkspaceRoleList contains a list of WorkspaceRole
type WorkspaceRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceRole `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkspaceRoleBinding is the Schema for the workspacerolebindings API
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:resource:categories="iam",scope="Cluster"
type WorkspaceRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Subjects holds references to the objects the role applies to.
	// +optional
	Subjects []rbacv1.Subject `json:"subjects,omitempty" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef can only reference a WorkspaceRole.
	// If the RoleRef cannot be resolved, the Authorizer must return an error.
	RoleRef rbacv1.RoleRef `json:"roleRef" protobuf:"bytes,3,opt,name=roleRef"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkspaceRoleBindingList contains a list of WorkspaceRoleBinding
type WorkspaceRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceRoleBinding `json:"items"`
}
