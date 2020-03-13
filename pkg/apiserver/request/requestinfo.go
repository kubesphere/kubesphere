package request

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"
	"net/http"
	"strings"

	k8srequest "k8s.io/apiserver/pkg/endpoints/request"
)

type RequestInfoResolver interface {
	NewRequestInfo(req *http.Request) (*RequestInfo, error)
}

// specialVerbs contains just strings which are used in REST paths for special actions that don't fall under the normal
// CRUDdy GET/POST/PUT/DELETE actions on REST objects.
// master's Mux.
var specialVerbs = sets.NewString("proxy", "watch")

var kubernetesAPIPrefixes = sets.NewString("api", "apis")

// RequestInfo holds information parsed from the http.Request,
// extended from k8s.io/apiserver/pkg/endpoints/request/requestinfo.go
type RequestInfo struct {
	*k8srequest.RequestInfo

	// IsKubeSphereRequest indicates whether or not the request should be handled by kubernetes or kubesphere
	IsKubernetesRequest bool

	// Workspace of requested namespace, for non-workspaced resources, this may be empty
	Workspace string

	// Cluster of requested resource, this is empty in single-cluster environment
	Cluster string
}

type RequestInfoFactory struct {
	APIPrefixes           sets.String
	GrouplessAPIPrefixes  sets.String
	k8sRequestInfoFactory *k8srequest.RequestInfoFactory
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
// /kapis/{api-group}/{version}/clusters/{cluster}/namespaces/{namespace}/{resource}
// /kapis/{api-group}/{version}/clusters/{cluster}/namespaces/{namespace}/{resource}/{resourceName}
//
func (r *RequestInfoFactory) NewRequestInfo(req *http.Request) (*RequestInfo, error) {

	requestInfo := RequestInfo{
		IsKubernetesRequest: false,
		RequestInfo: &k8srequest.RequestInfo{
			Path: req.URL.Path,
			Verb: req.Method,
		},
	}

	defer func() {
		if kubernetesAPIPrefixes.Has(requestInfo.APIPrefix) {
			requestInfo.IsKubernetesRequest = true
		}
	}()

	currentParts := splitPath(req.URL.Path)
	if len(currentParts) < 3 {
		return &requestInfo, nil
	}

	if !r.APIPrefixes.Has(currentParts[0]) {
		// return a non-resource request
		return &requestInfo, nil
	}
	requestInfo.APIPrefix = currentParts[0]
	currentParts = currentParts[1:]

	if !r.GrouplessAPIPrefixes.Has(requestInfo.APIPrefix) {
		if len(currentParts) < 2 {
			return &requestInfo, nil
		}

		if currentParts[0] == "clusters" {
			requestInfo.Cluster = currentParts[1]
			currentParts = currentParts[2:]
		}

		if len(currentParts) < 3 {
			return &requestInfo, nil
		}

		requestInfo.APIGroup = currentParts[0]
		currentParts = currentParts[1:]
	}

	requestInfo.IsResourceRequest = true
	requestInfo.APIVersion = currentParts[0]
	currentParts = currentParts[1:]

	if specialVerbs.Has(currentParts[0]) {
		if len(currentParts) < 2 {
			return &requestInfo, fmt.Errorf("unable to determine kind and namespace from url: %v", req.URL)
		}

		requestInfo.Verb = currentParts[0]
		currentParts = currentParts[1:]
	} else {
		switch req.Method {
		case "POST":
			requestInfo.Verb = "create"
		case "GET", "HEAD":
			requestInfo.Verb = "get"
		case "PUT":
			requestInfo.Verb = "update"
		case "PATCH":
			requestInfo.Verb = "patch"
		case "DELETE":
			requestInfo.Verb = "delete"
		default:
			requestInfo.Verb = ""
		}
	}

	return &requestInfo, nil
}

type requestInfoKeyType int

// requestInfoKey is the RequestInfo key for the context. It's of private type here. Because
// keys are interfaces and interfaces are equal when the type and the value is equal, this
// does not conflict with the keys defined in pkg/api.
const requestInfoKey requestInfoKeyType = iota

func WithRequestInfo(parent context.Context, info *RequestInfo) context.Context {
	return k8srequest.WithValue(parent, requestInfoKey, info)
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
