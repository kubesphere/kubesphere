package v1alpha2

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindUser                      = "User"
	ResourcesSingularUser                 = "user"
	ResourcesPluralUser                   = "users"
	ResourceKindLoginRecord               = "LoginRecord"
	ResourcesSingularLoginRecord          = "loginrecord"
	ResourcesPluralLoginRecord            = "loginrecords"
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
	GrantedClustersAnnotation             = "iam.kubesphere.io/granted-clusters"
	UninitializedAnnotation               = "iam.kubesphere.io/uninitialized"
	LastPasswordChangeTimeAnnotation      = "iam.kubesphere.io/last-password-change-time"
	RoleAnnotation                        = "iam.kubesphere.io/role"
	RoleTemplateLabel                     = "iam.kubesphere.io/role-template"
	ScopeLabelFormat                      = "scope.kubesphere.io/%s"
	UserReferenceLabel                    = "iam.kubesphere.io/user-ref"
	IdentifyProviderLabel                 = "iam.kubesphere.io/identify-provider"
	OriginUIDLabel                        = "iam.kubesphere.io/origin-uid"
	ServiceAccountReferenceLabel          = "iam.kubesphere.io/serviceaccount-ref"
	FieldEmail                            = "email"
	ExtraEmail                            = FieldEmail
	ExtraIdentityProvider                 = "idp"
	ExtraUID                              = "uid"
	ExtraUsername                         = "username"
	ExtraDisplayName                      = "displayName"
	ExtraUninitialized                    = "uninitialized"
	InGroup                               = "ingroup"
	NotInGroup                            = "notingroup"
	AggregateTo                           = "aggregateTo"
	ScopeWorkspace                        = "workspace"
	ScopeCluster                          = "cluster"
	ScopeNamespace                        = "namespace"
	ScopeDevOps                           = "devops"
	PlatformAdmin                         = "platform-admin"
	NamespaceAdmin                        = "admin"
	ClusterAdmin                          = "cluster-admin"
	PreRegistrationUser                   = "system:pre-registration"
	PreRegistrationUserGroup              = "pre-registration"
)

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:deprecatedversion
// +kubebuilder:printcolumn:name="Email",type="string",JSONPath=".spec.email"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:resource:categories="iam",scope="Cluster"

// User is the Schema for the users API
type User struct {
	metav1.TypeMeta `json:",inline"`
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

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:deprecatedversion
// +kubebuilder:resource:categories="iam",scope="Cluster"

type GlobalRole struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Rules holds all the PolicyRules for this GlobalRole
	// +optional
	Rules []rbacv1.PolicyRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`
}

// +kubebuilder:object:root=true

// GlobalRoleList contains a list of GlobalRole
type GlobalRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRole `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:deprecatedversion
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

// GlobalRoleBindingList contains a list of GlobalRoleBinding
type GlobalRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRoleBinding `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:deprecatedversion
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="Alias",type="string",JSONPath=".metadata.annotations.kubesphere\\.io/alias-name"
// +kubebuilder:resource:categories="iam",scope="Cluster"

type WorkspaceRole struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Rules holds all the PolicyRules for this WorkspaceRole
	// +optional
	Rules []rbacv1.PolicyRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`
}

// +kubebuilder:object:root=true

// WorkspaceRoleList contains a list of WorkspaceRole
type WorkspaceRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceRole `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:deprecatedversion
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

// WorkspaceRoleBindingList contains a list of WorkspaceRoleBinding
type WorkspaceRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceRoleBinding `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:deprecatedversion
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
// +kubebuilder:printcolumn:name="From",type="string",JSONPath=".spec.sourceIP"
// +kubebuilder:printcolumn:name="Success",type="string",JSONPath=".spec.success"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".spec.reason"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories="iam",scope="Cluster"

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

// LoginRecordList contains a list of LoginRecord
type LoginRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoginRecord `json:"items"`
}

// GroupSpec defines the desired state of Group
type GroupSpec struct {
}

// GroupStatus defines the observed state of Group
type GroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:deprecatedversion
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:resource:categories="group",scope="Cluster"

// Group is the Schema for the groups API
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec,omitempty"`
	Status GroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

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
// +kubebuilder:deprecatedversion
// +kubebuilder:printcolumn:name="Group",type="string",JSONPath=".groupRef.name"
// +kubebuilder:printcolumn:name="Users",type="string",JSONPath=".users"
// +kubebuilder:resource:categories="group",scope="Cluster"

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
