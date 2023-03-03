package v1beta1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	UserReferenceLabel  = "iam.kubesphere.io/user-ref"
	ResourcesPluralUser = "users"
)

// CategorySpec defines the desired state of Category
type CategorySpec struct {
	DisplayName map[string]string `json:"displayName,omitempty"`
	Description map[string]string `json:"description,omitempty"`
	Icon        string            `json:"icon,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster

// Category is the Schema for the categories API
type Category struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CategorySpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster

// CategoryList contains a list of Category
type CategoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Category `json:"items"`
}

// AggregationRoleTemplates indicates which roleTemplate the role is composed of.
// If the aggregation selector is not empty, the templateNames will be overwritten by the templates list by selector.
type AggregationRoleTemplates struct {
	// TemplateNames select rules from RoleTemplate`s rules by RoleTemplate name
	TemplateNames []string `json:"templateNames,omitempty"`

	// Selector select rules from RoleTemplate`s rules by labels
	Selector metav1.LabelSelector `json:"selector,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster

// GlobalRole is the Schema for the globalroles API
type GlobalRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// AggregationRoleTemplates means which RoleTemplates are composed this Role
	AggregationRoleTemplates AggregationRoleTemplates `json:"aggregationRoleTemplates,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	Rules []rbacv1.PolicyRule `json:"rules"`
}

//+kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"

// GlobalRoleList contains a list of GlobalRole
type GlobalRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRole `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"

// GlobalRoleBinding is the Schema for the globalrolebindings API
type GlobalRoleBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Subjects holds references to the objects the role applies to.
	// +optional
	Subjects []rbacv1.Subject `json:"subjects,omitempty" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef can only reference a GlobalRole.
	// If the RoleRef cannot be resolved, the Authorizer must return an error.
	RoleRef rbacv1.RoleRef `json:"roleRef" protobuf:"bytes,3,opt,name=roleRef"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster

// GlobalRoleBindingList contains a list of GlobalRoleBinding
type GlobalRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRoleBinding `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster
//+kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
//+kubebuilder:printcolumn:name="Alias",type="string",JSONPath=".metadata.annotations.kubesphere\\.io/alias-name"

// WorkspaceRole is the Schema for the workspaceroles API
type WorkspaceRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// AggregationRoleTemplates means which RoleTemplates are composed this Role
	AggregationRoleTemplates AggregationRoleTemplates `json:"aggregationRoleTemplates,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	Rules []rbacv1.PolicyRule `json:"rules,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster

// WorkspaceRoleList contains a list of WorkspaceRole
type WorkspaceRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceRole `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:resource:categories="iam",scope="Cluster"

// WorkspaceRoleBinding is the Schema for the workspacerolebindings API
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

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"

// WorkspaceRoleBindingList contains a list of WorkspaceRoleBinding
type WorkspaceRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceRoleBinding `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Namespaced

// Role is the Schema for the roles API
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// AggregationRoleTemplates means which RoleTemplates are composed this Role
	AggregationRoleTemplates AggregationRoleTemplates `json:"aggregationRoleTemplates,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	Rules []rbacv1.PolicyRule `json:"rules,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Namespaced

// RoleList contains a list of Role
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Namespaced

type RoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Subjects holds references to the objects the role applies to.
	// +optional
	Subjects []rbacv1.Subject `json:"subjects,omitempty" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef can only reference a WorkspaceRole.
	// If the RoleRef cannot be resolved, the Authorizer must return an error.
	RoleRef rbacv1.RoleRef `json:"roleRef" protobuf:"bytes,3,opt,name=roleRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Namespaced

type RoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleBinding `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster

// ClusterRole is the Schema for the clusterroles API
type ClusterRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// AggregationRoleTemplates means which RoleTemplates are composed this Role
	AggregationRoleTemplates AggregationRoleTemplates `json:"aggregationRoleTemplates,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	Rules []rbacv1.PolicyRule `json:"rules,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster

// ClusterRoleList contains a list of ClusterRole
type ClusterRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterRole `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster

type ClusterRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Subjects holds references to the objects the role applies to.
	// +optional
	Subjects []rbacv1.Subject `json:"subjects,omitempty" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef can only reference a WorkspaceRole.
	// If the RoleRef cannot be resolved, the Authorizer must return an error.
	RoleRef rbacv1.RoleRef `json:"roleRef" protobuf:"bytes,3,opt,name=roleRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster

type ClusterRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterRoleBinding `json:"items"`
}

// RoleTemplateSpec defines the desired state of RoleTemplate
type RoleTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// DisplayName represent the name displays at console, this field
	DisplayName map[string]string   `json:"displayName,omitempty"`
	Description map[string]string   `json:"description,omitempty"`
	Rules       []rbacv1.PolicyRule `json:"rules"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster
// +kubebuilder:storageversion
// RoleTemplate is the Schema for the roletemplates API
type RoleTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RoleTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories=iam,scope=Cluster

// RoleTemplateList contains a list of RoleTemplate
type RoleTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleTemplate `json:"items"`
}
