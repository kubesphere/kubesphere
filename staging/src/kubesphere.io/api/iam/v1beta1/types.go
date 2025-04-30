/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1beta1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultInfo provides a simple user information exchange object
// for components that implement the UserInfo interface.
type DefaultInfo struct {
	Name   string
	UID    string
	Groups []string
	Extra  map[string][]string
}

func (i *DefaultInfo) GetName() string {
	return i.Name
}

func (i *DefaultInfo) GetUID() string {
	return i.UID
}

func (i *DefaultInfo) GetGroups() []string {
	return i.Groups
}

func (i *DefaultInfo) GetExtra() map[string][]string {
	return i.Extra
}

// User is the Schema for the users API
// +kubebuilder:printcolumn:name="Email",type="string",JSONPath=".spec.email"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
type User struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserSpec `json:"spec"`
	// +optional
	Status UserStatus `json:"status,omitempty"`
}

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
	// +kubebuilder:validation:MinLength=8
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Pattern=`^(.*[a-z].*[A-Z].*[0-9].*)$|^(.*[a-z].*[0-9].*[A-Z].*)$|^(.*[A-Z].*[a-z].*[0-9].*)$|^(.*[A-Z].*[0-9].*[a-z].*)$|^(.*[0-9].*[a-z].*[A-Z].*)$|^(.*[0-9].*[A-Z].*[a-z].*)$|^(\$2[ayb]\$.{56})$`
	// Password pattern is tricky here.
	// The rule is simple: length between [6,64], at least one uppercase letter, one lowercase letter, one digit.
	// The regexp in console(javascript) is quite straightforward: ^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[^]{6,64}$
	// But in Go, we don't have ?= (back tracking) capability in regexp (also in CRD validation pattern)
	// So we adopted an alternative scheme to achieve.
	// Use 6 different regexp to combine to achieve the same effect.
	// These six schemes enumerate the arrangement of numbers, uppercase letters, and lowercase letters that appear for the first time.
	// - ^(.*[a-z].*[A-Z].*[0-9].*)$ stands for lowercase letter comes first, then followed by an uppercase letter, then a digit.
	// - ^(.*[a-z].*[0-9].*[A-Z].*)$ stands for lowercase letter comes first, then followed by a digit, then an uppercase leeter.
	// - ^(.*[A-Z].*[a-z].*[0-9].*)$ ...
	// - ^(.*[A-Z].*[0-9].*[a-z].*)$ ...
	// - ^(.*[0-9].*[a-z].*[A-Z].*)$ ...
	// - ^(.*[0-9].*[A-Z].*[a-z].*)$ ...
	// Last but not least, the bcrypt string is also included to match the encrypted password. ^(\$2[ayb]\$.{56})$
	EncryptedPassword string `json:"password,omitempty"`
}

type UserState string

// These are the valid phases of a user.
const (
	// UserActive means the user is available.
	UserActive UserState = "Active"
	// UserDisabled means the user is disabled.
	UserDisabled UserState = "Disabled"
	// UserAuthLimitExceeded means restrict user login.
	UserAuthLimitExceeded UserState = "AuthLimitExceeded"

	AuthenticatedSuccessfully = "authenticated successfully"
)

// UserStatus defines the observed state of User
type UserStatus struct {
	// The user status
	// +optional
	State UserState `json:"state,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// Last login attempt timestamp
	// +optional
	LastLoginTime *metav1.Time `json:"lastLoginTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +kubebuilder:storageversion

type BuiltinRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:EmbeddedResource
	Role           runtime.RawExtension `json:"role"`
	TargetSelector metav1.LabelSelector `json:"targetSelector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="iam",scope="Cluster"

// BuiltinRoleList contains a list of BuiltinRole
type BuiltinRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BuiltinRole `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
// +kubebuilder:printcolumn:name="From",type="string",JSONPath=".spec.sourceIP"
// +kubebuilder:printcolumn:name="Success",type="string",JSONPath=".spec.success"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".spec.reason"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +kubebuilder:storageversion

type LoginRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LoginRecordSpec `json:"spec"`
}

type LoginRecordSpec struct {
	// Which authentication method used, Password/OAuth/Token
	Type LoginType `json:"type"`
	// Provider of authentication, Ldap/Github etc.
	Provider string `json:"provider"`
	// Source IP of client
	SourceIP string `json:"sourceIP"`
	// User agent of login attempt
	UserAgent string `json:"userAgent,omitempty"`
	// Successful login attempt or not
	Success bool `json:"success"`
	// States failed login attempt reason
	Reason string `json:"reason"`
}

type LoginType string

const (
	Password LoginType = "Password"
	OAuth    LoginType = "OAuth"
	Token    LoginType = "Token"
)

// +kubebuilder:object:root=true
// +kubebuilder:object:root=true

// LoginRecordList contains a list of LoginRecord
type LoginRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoginRecord `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:resource:categories="group",scope="Cluster"
// +kubebuilder:storageversion

// Group is the Schema for the groups API
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec,omitempty"`
	Status GroupStatus `json:"status,omitempty"`
}

// GroupSpec defines the desired state of Group
type GroupSpec struct{}

// GroupStatus defines the observed state of Group
type GroupStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="group",scope="Cluster"

// GroupList contains a list of Group
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}

// GroupRef defines the desired relation of GroupBinding
type GroupRef struct {
	APIGroup string `json:"apiGroup,omitempty"`
	Kind     string `json:"kind,omitempty"`
	Name     string `json:"name,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Group",type="string",JSONPath=".groupRef.name"
// +kubebuilder:printcolumn:name="Users",type="string",JSONPath=".users"
// +kubebuilder:resource:categories="group",scope="Cluster"
// +kubebuilder:storageversion

// GroupBinding is the Schema for the groupbindings API
type GroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	GroupRef GroupRef `json:"groupRef,omitempty"`
	Users    []string `json:"users,omitempty"`
}

// +kubebuilder:object:root=true

// GroupBindingList contains a list of GroupBinding
type GroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupBinding `json:"items"`
}

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
