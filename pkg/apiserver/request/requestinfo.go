/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

// NOTE: This file is copied from k8s.io/apiserver/pkg/endpoints/request.
// We expanded requestInfo.

package request

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation/path"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metainternalversionscheme "k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	k8srequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
)

const (
	VerbCreate = "create"
	VerbGet    = "get"
	VerbList   = "list"
	VerbUpdate = "update"
	VerbDelete = "delete"
	VerbWatch  = "watch"
	VerbPatch  = "patch"
)

type RequestInfoResolver interface {
	NewRequestInfo(req *http.Request) (*RequestInfo, error)
}

// specialVerbs contains just strings which are used in REST paths for special actions that don't fall under the normal
// CRUDdy GET/POST/PUT/DELETE actions on REST objects.
// master's Mux.
var specialVerbs = sets.New("proxy", "watch")

// specialVerbsNoSubresources contains root verbs which do not allow subresources
var specialVerbsNoSubresources = sets.New("proxy")

// namespaceSubresources contains subresources of namespace
// this list allows the parser to distinguish between a namespace subresource, and a namespaced resource
var namespaceSubresources = sets.New("status", "finalize")

var kubernetesAPIPrefixes = sets.New("api", "apis")

// RequestInfo holds information parsed from the http.Request,
// extended from k8s.io/apiserver/pkg/endpoints/request/requestinfo.go
type RequestInfo struct {
	*k8srequest.RequestInfo

	// IsKubernetesRequest indicates whether or not the request should be handled by kubernetes or kubesphere
	IsKubernetesRequest bool

	// Workspace of requested resource, for non-workspaced resources, this may be empty
	Workspace string

	// Cluster of requested resource, this is empty in single-cluster environment
	Cluster string

	// Scope of requested resource.
	ResourceScope string

	// Source IP
	SourceIP string

	// User agent
	UserAgent string
}

type RequestInfoFactory struct {
	APIPrefixes          sets.Set[string]
	GrouplessAPIPrefixes sets.Set[string]
	GlobalResources      []schema.GroupResource
}

// NewRequestInfo returns the information from the http request.  If error is not nil, RequestInfo holds the information as best it is known before the failure
// It handles both resource and non-resource requests and fills in all the pertinent information for each.
// Valid Inputs:
//
// /apis/{api-group}/{version}/namespaces
// /api/{version}/namespaces
// /api/{version}/namespaces/{namespace}
// /api/{version}/namespaces/{namespace}/{resource}
// /api/{version}/namespaces/{namespace}/{resource}/{resourceName}
// /api/{version}/{resource}
// /api/{version}/{resource}/{resourceName}
//
// Special verbs without subresources:
// /api/{version}/proxy/{resource}/{resourceName}
// /api/{version}/proxy/namespaces/{namespace}/{resource}/{resourceName}
//
// Special verbs with subresources:
// /api/{version}/watch/{resource}
// /api/{version}/watch/namespaces/{namespace}/{resource}
//
// /kapis/{api-group}/{version}/workspaces/{workspace}/{resource}/{resourceName}
// /
// /kapis/{api-group}/{version}/namespaces/{namespace}/{resource}
// /kapis/{api-group}/{version}/namespaces/{namespace}/{resource}/{resourceName}
// With workspaces:
// /clusters/{cluster}/kapis/{api-group}/{version}/namespaces/{namespace}/{resource}
// /clusters/{cluster}/kapis/{api-group}/{version}/namespaces/{namespace}/{resource}/{resourceName}
func (r *RequestInfoFactory) NewRequestInfo(req *http.Request) (*RequestInfo, error) {
	requestInfo := RequestInfo{
		IsKubernetesRequest: false,
		RequestInfo: &k8srequest.RequestInfo{
			Path: req.URL.Path,
			Verb: req.Method,
		},
		Workspace: api.WorkspaceNone,
		Cluster:   api.ClusterNone,
		SourceIP:  iputil.RemoteIp(req),
		UserAgent: req.UserAgent(),
	}

	defer func() {
		prefix := requestInfo.APIPrefix
		if prefix == "" {
			currentParts := splitPath(requestInfo.Path)
			// Proxy discovery API
			if len(currentParts) > 0 && len(currentParts) < 3 {
				prefix = currentParts[0]
			}
		}
		if kubernetesAPIPrefixes.Has(prefix) {
			requestInfo.IsKubernetesRequest = true
		}
	}()

	currentParts := splitPath(req.URL.Path)
	if len(currentParts) < 3 {
		return &requestInfo, nil
	}

	// URL forms: /clusters/{cluster}/*
	if currentParts[0] == "clusters" {
		if len(currentParts) > 1 {
			requestInfo.Cluster = currentParts[1]
			// resolve the real path behind the cluster dispatcher
			requestInfo.Path = strings.TrimPrefix(requestInfo.Path, fmt.Sprintf("/clusters/%s", requestInfo.Cluster))
		}
		if len(currentParts) > 2 {
			currentParts = currentParts[2:]
		}
	}

	if !r.APIPrefixes.Has(currentParts[0]) {
		// return a non-resource request
		return &requestInfo, nil
	}
	requestInfo.APIPrefix = currentParts[0]
	currentParts = currentParts[1:]

	// fallback to legacy cluster API
	// TODO remove the following codes
	if requestInfo.Cluster == "" {
		// URL forms: /(kapis|apis|api)/clusters/{cluster}/*
		if currentParts[0] == "clusters" {
			if len(currentParts) > 1 {
				requestInfo.Cluster = currentParts[1]
				// resolve the real path behind the cluster dispatcher
				requestInfo.Path = strings.Replace(requestInfo.Path, fmt.Sprintf("/clusters/%s", requestInfo.Cluster), "", 1)
			}
			if len(currentParts) > 2 {
				currentParts = currentParts[2:]
			}
		}
	}

	if !r.GrouplessAPIPrefixes.Has(requestInfo.APIPrefix) {
		// one part (APIPrefix) has already been consumed, so this is actually "do we have four parts?"
		if len(currentParts) < 3 {
			// return a non-resource request
			return &requestInfo, nil
		}

		requestInfo.APIGroup = currentParts[0]
		currentParts = currentParts[1:]
	}

	requestInfo.IsResourceRequest = true
	requestInfo.APIVersion = currentParts[0]
	currentParts = currentParts[1:]

	if len(currentParts) > 0 && specialVerbs.Has(currentParts[0]) {
		if len(currentParts) < 2 {
			return &requestInfo, fmt.Errorf("unable to determine kind and namespace from url: %v", req.URL)
		}

		requestInfo.Verb = currentParts[0]
		currentParts = currentParts[1:]
	} else {
		switch req.Method {
		case "POST":
			requestInfo.Verb = VerbCreate
		case "GET", "HEAD":
			requestInfo.Verb = VerbGet
		case "PUT":
			requestInfo.Verb = VerbUpdate
		case "PATCH":
			requestInfo.Verb = VerbPatch
		case "DELETE":
			requestInfo.Verb = VerbDelete
		default:
			requestInfo.Verb = ""
		}
	}

	// URL forms: /workspaces/{workspace}/*
	if currentParts[0] == "workspaces" || currentParts[0] == "workspacetemplates" {
		if len(currentParts) > 1 {
			requestInfo.Workspace = currentParts[1]
		}
		if len(currentParts) > 2 {
			currentParts = currentParts[2:]
		}
	}

	// URL forms: /namespaces/{namespace}/{kind}/*, where parts are adjusted to be relative to kind
	if currentParts[0] == "namespaces" {
		if len(currentParts) > 1 {
			requestInfo.Namespace = currentParts[1]

			// if there is another step after the namespace name, and it is not a known namespace subresource
			// move currentParts to include it as a resource in its own right
			if len(currentParts) > 2 && !namespaceSubresources.Has(currentParts[2]) {
				currentParts = currentParts[2:]
			}
		}
	}

	// parsing successful, so we now know the proper value for .Parts
	requestInfo.Parts = currentParts

	// parts look like: resource/resourceName/subresource/other/stuff/we/don't/interpret
	switch {
	case len(requestInfo.Parts) >= 3 && !specialVerbsNoSubresources.Has(requestInfo.Verb):
		requestInfo.Subresource = requestInfo.Parts[2]
		fallthrough
	case len(requestInfo.Parts) >= 2:
		requestInfo.Name = requestInfo.Parts[1]
		fallthrough
	case len(requestInfo.Parts) >= 1:
		requestInfo.Resource = requestInfo.Parts[0]
	}

	requestInfo.ResourceScope = r.resolveResourceScope(requestInfo)

	// if there's no name on the request and we thought it was a get before, then the actual verb is a list or a watch
	if len(requestInfo.Name) == 0 && requestInfo.Verb == VerbGet {
		opts := metainternalversion.ListOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), metav1.SchemeGroupVersion, &opts); err != nil {
			// An error in parsing request will result in default to "list" and not setting "name" field.
			klog.Errorf("Couldn't parse request %#v: %v", req.URL.Query(), err)
			// Reset opts to not rely on partial results from parsing.
			// However, if watch is set, let's report it.
			opts = metainternalversion.ListOptions{}
			if values := req.URL.Query()["watch"]; len(values) > 0 {
				switch strings.ToLower(values[0]) {
				case "false", "0":
				default:
					opts.Watch = true
				}
			}
		}

		if opts.Watch {
			requestInfo.Verb = VerbWatch
		} else {
			requestInfo.Verb = VerbList
		}

		if opts.FieldSelector != nil {
			if name, ok := opts.FieldSelector.RequiresExactMatch("metadata.name"); ok {
				if len(path.IsValidPathSegmentName(name)) == 0 {
					requestInfo.Name = name
				}
			}
		}
	}

	// URL forms: /api/v1/watch/namespaces?labelSelector=kubesphere.io/workspace=system-workspace
	if requestInfo.Verb == VerbWatch {
		selector := req.URL.Query().Get("labelSelector")
		if strings.HasPrefix(selector, workspaceSelectorPrefix) {
			workspace := strings.TrimPrefix(selector, workspaceSelectorPrefix)
			requestInfo.Workspace = workspace
			requestInfo.ResourceScope = WorkspaceScope
		}
	}

	// if there's no name on the request and we thought it was a delete before, then the actual verb is deletecollection
	if len(requestInfo.Name) == 0 && requestInfo.Verb == VerbDelete {
		requestInfo.Verb = "deletecollection"
	}

	return &requestInfo, nil
}

type requestInfoKeyType int

// requestInfoKey is the RequestInfo key for the context. It's of private type here. Because
// keys are interfaces and interfaces are equal when the type and the value is equal, this
// does not conflict with the keys defined in pkg/api.
const requestInfoKey requestInfoKeyType = iota

func WithRequestInfo(parent context.Context, info *RequestInfo) context.Context {
	return context.WithValue(parent, requestInfoKey, info)
}

func RequestInfoFrom(ctx context.Context) (*RequestInfo, bool) {
	info, ok := ctx.Value(requestInfoKey).(*RequestInfo)
	return info, ok
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

const (
	GlobalScope             = "Global"
	ClusterScope            = "Cluster"
	WorkspaceScope          = "Workspace"
	NamespaceScope          = "Namespace"
	workspaceSelectorPrefix = constants.WorkspaceLabelKey + "="
)

func (r *RequestInfoFactory) resolveResourceScope(request RequestInfo) string {
	if r.isGlobalScopeResource(request.APIGroup, request.Resource) {
		// GET /apis/tenant.kubesphere.io/v1beta1/workspaces/{workspace}
		if request.Workspace != "" {
			return WorkspaceScope
		}
		return GlobalScope
	}

	if request.Namespace != "" {
		return NamespaceScope
	}

	if request.Workspace != "" {
		return WorkspaceScope
	}

	return ClusterScope
}

func (r *RequestInfoFactory) isGlobalScopeResource(apiGroup, resource string) bool {
	for _, groupResource := range r.GlobalResources {
		if groupResource.Group == apiGroup && groupResource.Resource == resource {
			return true
		}
	}
	return false
}
