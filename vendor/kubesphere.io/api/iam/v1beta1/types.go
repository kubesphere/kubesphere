package v1beta1

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CategorySpec defines the desired state of Category
type CategorySpec struct {
	DisplayName map[string]string `json:"displayName,omitempty"`
	Description map[string]string `json:"description,omitempty"`
	Icon        string            `json:"icon,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster

// Category is the Schema for the categories API
type Category struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CategorySpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster

// CategoryList contains a list of Category
type CategoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Category `json:"items"`
}

// AggregationRoleTemplates indicates which roleTemplate the role is composed of.
// If the TemplateNames is not empty, Selector will be ignored.
type AggregationRoleTemplates struct {
	// TemplateNames select rules from RoleTemplate`s rules by RoleTemplate name
	//+listType=set
	TemplateNames []string `json:"templateNames,omitempty"`

	// +optional
	// RoleSelectors select rules from RoleTemplate`s rules by labels
	RoleSelector *metav1.LabelSelector `json:"roleSelector,omitempty"`
}

// +kubebuilder:object:root=true

// SubjectAccessReview checks whether a user or group can perform an action.
// NOTE: This type does not require crd, so we omit the metav1.ObjectMeta
type SubjectAccessReview struct {
	metav1.TypeMeta `json:",inline"`

	// Spec holds information about the request being evaluated
	Spec SubjectAccessReviewSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// Status is filled in by the server and indicates whether the request is allowed or not
	// +optional
	Status SubjectAccessReviewStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:object:root=true

// SubjectAccessReviewList contains a list of SubjectAccessReview
type SubjectAccessReviewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubjectAccessReview `json:"items"`
}

// SubjectAccessReviewSpec is a description of the access request.  Exactly one of ResourceAuthorizationAttributes
// and NonResourceAuthorizationAttributes must be set
type SubjectAccessReviewSpec struct {
	// ResourceAuthorizationAttributes describes information for a resource access request
	// +optional
	ResourceAttributes *ResourceAttributes `json:"resourceAttributes,omitempty" protobuf:"bytes,1,opt,name=resourceAttributes"`
	// NonResourceAttributes describes information for a non-resource access request
	// +optional
	NonResourceAttributes *NonResourceAttributes `json:"nonResourceAttributes,omitempty" protobuf:"bytes,2,opt,name=nonResourceAttributes"`

	// User is the user you're testing for.
	// If you specify "User" but not "Groups", then is it interpreted as "What if User were not a member of any groups
	// +optional
	User string `json:"user,omitempty" protobuf:"bytes,3,opt,name=user"`
	// Groups is the groups you're testing for.
	// +optional
	Groups []string `json:"groups,omitempty" protobuf:"bytes,4,rep,name=groups"`
	// Extra corresponds to the user.Info.GetExtra() method from the authenticator.  Since that is input to the authorizer
	// it needs a reflection here.
	// +optional
	Extra map[string]ExtraValue `json:"extra,omitempty" protobuf:"bytes,5,rep,name=extra"`
	// UID information about the requesting user.
	// +optional
	UID string `json:"uid,omitempty" protobuf:"bytes,6,opt,name=uid"`
}

// ExtraValue masks the value so protobuf can generate
type ExtraValue []string

func (t ExtraValue) String() string {
	return fmt.Sprintf("%v", []string(t))
}

// ResourceAttributes includes the authorization attributes available for resource requests to the Authorizer interface
type ResourceAttributes struct {
	// ResourceScope indicate which scope the resource belongs to
	// +optional
	ResourceScope string `json:"resourceScope,omitempty"`
	// +optional
	Workspace string `json:"workspace,omitempty"`
	// Namespace is the namespace of the action being requested.  Currently, there is no distinction between no namespace and all namespaces
	// "" (empty) is defaulted for LocalSubjectAccessReviews
	// "" (empty) is empty for cluster-scoped resources
	// "" (empty) means "all" for namespace scoped resources from a SubjectAccessReview or SelfSubjectAccessReview
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	// Verb is a kubernetes resource API verb, like: get, list, watch, create, update, delete, proxy.  "*" means all.
	// +optional
	Verb string `json:"verb,omitempty" protobuf:"bytes,2,opt,name=verb"`
	// Group is the API Group of the Resource.  "*" means all.
	// +optional
	Group string `json:"group,omitempty" protobuf:"bytes,3,opt,name=group"`
	// Version is the API Version of the Resource.  "*" means all.
	// +optional
	Version string `json:"version,omitempty" protobuf:"bytes,4,opt,name=version"`
	// Resource is one of the existing resource types.  "*" means all.
	// +optional
	Resource string `json:"resource,omitempty" protobuf:"bytes,5,opt,name=resource"`
	// Subresource is one of the existing resource types.  "" means none.
	// +optional
	Subresource string `json:"subresource,omitempty" protobuf:"bytes,6,opt,name=subresource"`
	// Name is the name of the resource being requested for a "get" or deleted for a "delete". "" (empty) means all.
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,7,opt,name=name"`
}

// NonResourceAttributes includes the authorization attributes available for non-resource requests to the Authorizer interface
type NonResourceAttributes struct {
	// Path is the URL path of the request
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,1,opt,name=path"`
	// Verb is the standard HTTP verb
	// +optional
	Verb string `json:"verb,omitempty" protobuf:"bytes,2,opt,name=verb"`
}

// SubjectAccessReviewStatus includes the authorization result
type SubjectAccessReviewStatus struct {
	// Allowed is required. True if the action would be allowed, false otherwise.
	Allowed bool `json:"allowed" protobuf:"varint,1,opt,name=allowed"`
	// Denied is optional. True if the action would be denied, otherwise
	// false. If both allowed is false and denied is false, then the
	// authorizer has no opinion on whether to authorize the action. Denied
	// may not be true if Allowed is true.
	// +optional
	Denied bool `json:"denied,omitempty" protobuf:"varint,4,opt,name=denied"`
	// Reason is optional.  It indicates why a request was allowed or denied.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,2,opt,name=reason"`
	// EvaluationError is an indication that some error occurred during the authorization check.
	// It is entirely possible to get an error and be able to continue determine authorization status in spite of it.
	// For instance, RBAC can be missing a role, but enough roles are still present and bound to reason about the request.
	// +optional
	EvaluationError string `json:"evaluationError,omitempty" protobuf:"bytes,3,opt,name=evaluationError"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster
// +kubebuilder:storageversion

// GlobalRole is the Schema for the globalroles API
type GlobalRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	// AggregationRoleTemplates means which RoleTemplates are composed this Role
	AggregationRoleTemplates *AggregationRoleTemplates `json:"aggregationRoleTemplates,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	Rules []rbacv1.PolicyRule `json:"rules"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"

// GlobalRoleList contains a list of GlobalRole
type GlobalRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRole `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +kubebuilder:storageversion

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

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="Alias",type="string",JSONPath=".metadata.annotations.kubesphere\\.io/alias-name"
// +kubebuilder:storageversion

// WorkspaceRole is the Schema for the workspaceroles API
type WorkspaceRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// AggregationRoleTemplates means which RoleTemplates are composed this Role
	AggregationRoleTemplates *AggregationRoleTemplates `json:"aggregationRoleTemplates,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	Rules []rbacv1.PolicyRule `json:"rules,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster

// WorkspaceRoleList contains a list of WorkspaceRole
type WorkspaceRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceRole `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:storageversion
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

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Namespaced

// Role is the Schema for the roles API
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// AggregationRoleTemplates means which RoleTemplates are composed this Role
	AggregationRoleTemplates *AggregationRoleTemplates `json:"aggregationRoleTemplates,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	Rules []rbacv1.PolicyRule `json:"rules,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Namespaced

// RoleList contains a list of Role
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Namespaced

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

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Namespaced

type RoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleBinding `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster

// ClusterRole is the Schema for the clusterroles API
type ClusterRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// AggregationRoleTemplates means which RoleTemplates are composed this Role
	AggregationRoleTemplates *AggregationRoleTemplates `json:"aggregationRoleTemplates,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	Rules []rbacv1.PolicyRule `json:"rules,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster

// ClusterRoleList contains a list of ClusterRole
type ClusterRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterRole `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster

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

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster

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

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=iam,scope=Cluster

// RoleTemplateList contains a list of RoleTemplate
type RoleTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleTemplate `json:"items"`
}
