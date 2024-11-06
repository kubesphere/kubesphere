package authorizer

import (
	"net/http"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apiserver/pkg/authentication/user"
)

// NOTE: This file is copied from k8s.io/kubernetes/apiserver/pkg/authorization/authorizer.
// We have expanded some attributes, such as workspace and cluster.

// Attributes is an interface used by an Authorizer to get information about a request
// that is used to make an authorization decision.
type Attributes interface {
	// GetUser returns the user.Info object to authorize
	GetUser() user.Info

	// GetVerb returns the kube verb associated with API requests (this includes get, list, watch, create, update, patch, delete, deletecollection, and proxy),
	// or the lowercased HTTP verb associated with non-API requests (this includes get, put, post, patch, and delete)
	GetVerb() string

	// When IsReadOnly() == true, the request has no side effects, other than
	// caching, logging, and other incidentals.
	IsReadOnly() bool

	// Indicates whether or not the request should be handled by kubernetes or kubesphere
	IsKubernetesRequest() bool

	// The cluster of the object, if a request is for a REST object.
	GetCluster() string

	// The workspace of the object, if a request is for a REST object.
	GetWorkspace() string

	// The namespace of the object, if a request is for a REST object.
	GetNamespace() string

	// The kind of object, if a request is for a REST object.
	GetResource() string

	// GetSubresource returns the subresource being requested, if present
	GetSubresource() string

	// GetName returns the name of the object as parsed off the request.  This will not be present for all request types, but
	// will be present for: get, update, delete
	GetName() string

	// The group of the resource, if a request is for a REST object.
	GetAPIGroup() string

	// GetAPIVersion returns the version of the group requested, if a request is for a REST object.
	GetAPIVersion() string

	// IsResourceRequest returns true for requests to API resources, like /api/v1/nodes,
	// and false for non-resource endpoints like /api, /healthz
	IsResourceRequest() bool

	// GetResourceScope returns the scope of the resource requested, if a request is for a REST object.
	GetResourceScope() string

	// GetPath returns the path of the request
	GetPath() string

	// ParseFieldSelector is lazy, thread-safe, and stores the parsed result and error.
	// It returns an error if the field selector cannot be parsed.
	// The returned requirements must be treated as readonly and not modified.
	GetFieldSelector() (fields.Requirements, error)

	// ParseLabelSelector is lazy, thread-safe, and stores the parsed result and error.
	// It returns an error if the label selector cannot be parsed.
	// The returned requirements must be treated as readonly and not modified.
	GetLabelSelector() (labels.Requirements, error)
}

// Authorizer makes an authorization decision based on information gained by making
// zero or more calls to methods of the Attributes interface.  It returns nil when an action is
// authorized, otherwise it returns an error.
type Authorizer interface {
	Authorize(a Attributes) (authorized Decision, reason string, err error)
}

type AuthorizerFunc func(a Attributes) (Decision, string, error)

func (f AuthorizerFunc) Authorize(a Attributes) (Decision, string, error) {
	return f(a)
}

// RuleResolver provides a mechanism for resolving the list of rules that apply to a given user within a namespace.
type RuleResolver interface {
	// RulesFor get the list of cluster wide rules, the list of rules in the specific namespace, incomplete status and errors.
	RulesFor(user user.Info, namespace string) ([]ResourceRuleInfo, []NonResourceRuleInfo, bool, error)
}

// RequestAttributesGetter provides a function that extracts Attributes from an http.Request
type RequestAttributesGetter interface {
	GetRequestAttributes(user.Info, *http.Request) Attributes
}

// AttributesRecord implements Attributes interface.
type AttributesRecord struct {
	User              user.Info
	Verb              string
	Cluster           string
	Workspace         string
	Namespace         string
	APIGroup          string
	APIVersion        string
	Resource          string
	Subresource       string
	Name              string
	KubernetesRequest bool
	ResourceRequest   bool
	Path              string
	ResourceScope     string
}

func (a AttributesRecord) GetFieldSelector() (fields.Requirements, error) {
	return fields.Requirements{}, nil
}

func (a AttributesRecord) GetLabelSelector() (labels.Requirements, error) {
	return labels.Requirements{}, nil
}

func (a AttributesRecord) GetUser() user.Info {
	return a.User
}

func (a AttributesRecord) GetVerb() string {
	return a.Verb
}

func (a AttributesRecord) IsReadOnly() bool {
	return a.Verb == VerbGet || a.Verb == VerbList || a.Verb == VerbWatch
}

func (a AttributesRecord) GetCluster() string {
	return a.Cluster
}

func (a AttributesRecord) GetWorkspace() string {
	return a.Workspace
}

func (a AttributesRecord) GetNamespace() string {
	return a.Namespace
}

func (a AttributesRecord) GetResource() string {
	return a.Resource
}

func (a AttributesRecord) GetSubresource() string {
	return a.Subresource
}

func (a AttributesRecord) GetName() string {
	return a.Name
}

func (a AttributesRecord) GetAPIGroup() string {
	return a.APIGroup
}

func (a AttributesRecord) GetAPIVersion() string {
	return a.APIVersion
}

func (a AttributesRecord) IsResourceRequest() bool {
	return a.ResourceRequest
}

func (a AttributesRecord) IsKubernetesRequest() bool {
	return a.KubernetesRequest
}

func (a AttributesRecord) GetPath() string {
	return a.Path
}

func (a AttributesRecord) GetResourceScope() string {
	return a.ResourceScope
}

type Decision int

const (
	// DecisionDeny means that an authorizer decided to deny the action.
	DecisionDeny Decision = iota
	// DecisionAllow means that an authorizer decided to allow the action.
	DecisionAllow
	// DecisionNoOpionion means that an authorizer has no opinion on whether
	// to allow or deny an action.
	DecisionNoOpinion
)

const (
	// VerbList represents the verb of listing resources
	VerbList = "list"
	// VerbCreate represents the verb of creating a resource
	VerbCreate = "create"
	// VerbGet represents the verb of getting a resource or resources
	VerbGet = "get"
	// VerbWatch represents the verb of watching a resource
	VerbWatch = "watch"
	// VerbDelete represents the verb of deleting a resource
	VerbDelete = "delete"
)
