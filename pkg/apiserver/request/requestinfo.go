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

// NOTE: This file is copied from k8s.io/apiserver/pkg/endpoints/request.
// We expanded requestInfo.

package request

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/validation/path"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metainternalversionscheme "k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"
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

// specialVerbsNoSubresources contains root verbs which do not allow subresources
var specialVerbsNoSubresources = sets.NewString("proxy")

// namespaceSubresources contains subresources of namespace
// this list allows the parser to distinguish between a namespace subresource, and a namespaced resource
var namespaceSubresources = sets.NewString("status", "finalize")

var kubernetesAPIPrefixes = sets.NewString("api", "apis")

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

	// DevOps project of requested resource
	DevOps string

	// Scope of requested resource.
	ResourceScope string
}

type RequestInfoFactory struct {
	APIPrefixes          sets.String
	GrouplessAPIPrefixes sets.String
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
// /kapis/clusters/{cluster}/{api-group}/{version}/namespaces/{namespace}/{resource}
// /kapis/clusters/{cluster}/{api-group}/{version}/namespaces/{namespace}/{resource}/{resourceName}
//
func (r *RequestInfoFactory) NewRequestInfo(req *http.Request) (*RequestInfo, error) {
	requestInfo := RequestInfo{
		IsKubernetesRequest: false,
		RequestInfo: &k8srequest.RequestInfo{
			Path: req.URL.Path,
			Verb: req.Method,
		},
		Workspace: api.WorkspaceNone,
		Cluster:   api.ClusterNone,
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

	// URL forms: /clusters/{cluster}/*
	if currentParts[0] == "clusters" {
		if len(currentParts) > 1 {
			requestInfo.Cluster = currentParts[1]
		}
		if len(currentParts) > 2 {
			currentParts = currentParts[2:]
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

	// URL forms: /workspaces/{workspace}/*
	if currentParts[0] == "workspaces" {
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

			// if there is another step after the namespace name and it is not a known namespace subresource
			// move currentParts to include it as a resource in its own right
			if len(currentParts) > 2 && !namespaceSubresources.Has(currentParts[2]) {
				currentParts = currentParts[2:]
			}
		}
	} else if currentParts[0] == "devops" {
		if len(currentParts) > 1 {
			requestInfo.DevOps = currentParts[1]

			// if there is another step after the devops name
			// move currentParts to include it as a resource in its own right
			if len(currentParts) > 2 {
				currentParts = currentParts[2:]
			}
		}
	} else {
		requestInfo.Namespace = metav1.NamespaceNone
		requestInfo.DevOps = metav1.NamespaceNone
	}

	// parsing successful, so we now know the proper value for .Parts
	requestInfo.Parts = currentParts

	requestInfo.ResourceScope = r.resolveResourceScope(requestInfo)

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

	// if there's no name on the request and we thought it was a get before, then the actual verb is a list or a watch
	if len(requestInfo.Name) == 0 && requestInfo.Verb == "get" {
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
			requestInfo.Verb = "watch"
		} else {
			requestInfo.Verb = "list"
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
	if requestInfo.Verb == "watch" {
		selector := req.URL.Query().Get("labelSelector")
		if strings.HasPrefix(selector, workspaceSelectorPrefix) {
			workspace := strings.TrimPrefix(selector, workspaceSelectorPrefix)
			requestInfo.Workspace = workspace
			requestInfo.ResourceScope = WorkspaceScope
		}
	}

	// if there's no name on the request and we thought it was a delete before, then the actual verb is deletecollection
	if len(requestInfo.Name) == 0 && requestInfo.Verb == "delete" {
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

const (
	GlobalScope             = "Global"
	ClusterScope            = "Cluster"
	WorkspaceScope          = "Workspace"
	NamespaceScope          = "Namespace"
	DevOpsScope             = "DevOps"
	workspaceSelectorPrefix = constants.WorkspaceLabelKey + "="
)

func (r *RequestInfoFactory) resolveResourceScope(request RequestInfo) string {
	if r.isGlobalScopeResource(request.APIGroup, request.Resource) {
		return GlobalScope
	}

	if request.Namespace != "" {
		return NamespaceScope
	}

	if request.DevOps != "" {
		return DevOpsScope
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
