package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
