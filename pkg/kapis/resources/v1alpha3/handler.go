/*
Copyright 2020 KubeSphere Authors

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

package v1alpha3

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/components"
	v2 "kubesphere.io/kubesphere/pkg/models/registries/v2"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	resourcev1alpha2 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/resource"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/server/params"
)

type Handler struct {
	resourceGetterV1alpha3  *resourcev1alpha3.ResourceGetter
	resourcesGetterV1alpha2 *resourcev1alpha2.ResourceGetter
	componentsGetter        components.ComponentsGetter
	registryHelper          v2.RegistryHelper
}

func New(resourceGetterV1alpha3 *resourcev1alpha3.ResourceGetter, resourcesGetterV1alpha2 *resourcev1alpha2.ResourceGetter, componentsGetter components.ComponentsGetter) *Handler {
	return &Handler{
		resourceGetterV1alpha3:  resourceGetterV1alpha3,
		resourcesGetterV1alpha2: resourcesGetterV1alpha2,
		componentsGetter:        componentsGetter,
		registryHelper:          v2.NewRegistryHelper(),
	}
}

func (h *Handler) handleGetResources(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	resourceType := request.PathParameter("resources")
	name := request.PathParameter("name")

	// use informers to retrieve resources
	result, err := h.resourceGetterV1alpha3.Get(resourceType, namespace, name)
	if err == nil {
		response.WriteEntity(result)
		return
	}

	if err != resourcev1alpha3.ErrResourceNotSupported {
		klog.Error(err, resourceType)
		api.HandleInternalError(response, nil, err)
		return
	}

	// fallback to v1alpha2
	resultV1alpha2, err := h.resourcesGetterV1alpha2.GetResource(namespace, resourceType, name)
	if err != nil {
		if err == resourcev1alpha2.ErrResourceNotSupported {
			api.HandleNotFound(response, request, err)
			return
		}
		klog.Error(err)
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(resultV1alpha2)

}

// handleListResources retrieves resources
func (h *Handler) handleListResources(request *restful.Request, response *restful.Response) {
	query := query.ParseQueryParameter(request)
	resourceType := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")

	result, err := h.resourceGetterV1alpha3.List(resourceType, namespace, query)
	if err == nil {
		response.WriteEntity(result)
		return
	}

	if err != resourcev1alpha3.ErrResourceNotSupported {
		klog.Error(err, resourceType)
		api.HandleInternalError(response, request, err)
		return
	}

	// fallback to v1alpha2
	result, err = h.fallback(resourceType, namespace, query)
	if err != nil {
		if err == resourcev1alpha2.ErrResourceNotSupported {
			api.HandleNotFound(response, request, err)
			return
		}
		klog.Error(err)
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func (h *Handler) fallback(resourceType string, namespace string, q *query.Query) (*api.ListResult, error) {
	orderBy := string(q.SortBy)
	limit, offset := q.Pagination.Limit, q.Pagination.Offset
	reverse := !q.Ascending
	conditions := &params.Conditions{Match: make(map[string]string, 0), Fuzzy: make(map[string]string, 0)}
	for field, value := range q.Filters {
		switch field {
		case query.FieldName:
			conditions.Fuzzy[v1alpha2.Name] = string(value)
			break
		case query.FieldNames:
			conditions.Match[v1alpha2.Name] = string(value)
			break
		case query.FieldCreationTimeStamp:
			conditions.Match[v1alpha2.CreateTime] = string(value)
			break
		case query.FieldLastUpdateTimestamp:
			conditions.Match[v1alpha2.UpdateTime] = string(value)
			break
		case query.FieldLabel:
			values := strings.SplitN(string(value), ":", 2)
			if len(values) == 2 {
				conditions.Match[values[0]] = values[1]
			} else {
				conditions.Match[v1alpha2.Label] = values[0]
			}
			break
		case query.FieldAnnotation:
			values := strings.SplitN(string(value), ":", 2)
			if len(values) == 2 {
				conditions.Match[v1alpha2.Annotation] = values[1]
			} else {
				conditions.Match[v1alpha2.Annotation] = values[0]
			}
			break
		case query.FieldStatus:
			conditions.Match[v1alpha2.Status] = string(value)
			break
		case query.FieldOwnerReference:
			conditions.Match[v1alpha2.Owner] = string(value)
			break
		default:
			conditions.Match[string(field)] = string(value)
			break
		}
	}

	result, err := h.resourcesGetterV1alpha2.ListResources(namespace, resourceType, conditions, orderBy, reverse, limit, offset)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &api.ListResult{
		Items:      result.Items,
		TotalItems: result.TotalCount,
	}, nil
}

func (h *Handler) handleGetComponentStatus(request *restful.Request, response *restful.Response) {
	component := request.PathParameter("component")
	result, err := h.componentsGetter.GetComponentStatus(component)
	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}
	response.WriteEntity(result)
}

func (h *Handler) handleGetSystemHealthStatus(request *restful.Request, response *restful.Response) {
	result, err := h.componentsGetter.GetSystemHealthStatus()
	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}

// get all componentsHandler
func (h *Handler) handleGetComponents(request *restful.Request, response *restful.Response) {
	result, err := h.componentsGetter.GetAllComponentsStatus()
	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}

// handleVerifyImageRepositorySecret verifies image secret against registry, it takes k8s.io/api/core/v1/types.Secret
// as input, and authenticate registry with credential specified. Returns http.StatusOK if authenticate successfully,
// returns http.StatusUnauthorized if failed.
func (h *Handler) handleVerifyImageRepositorySecret(request *restful.Request, response *restful.Response) {
	secret := &v1.Secret{}
	err := request.ReadEntity(secret)
	if err != nil {
		api.HandleBadRequest(response, request, err)
	}

	ok, err := h.registryHelper.Auth(secret)
	if !ok {
		klog.Error(err)
		api.HandleUnauthorized(response, request, err)
	} else {
		response.WriteHeaderAndJson(http.StatusOK, secret, restful.MIME_JSON)
	}
}

// handleGetImageConfig fetches container image spec described in https://github.com/opencontainers/image-spec/blob/main/manifest.md
func (h *Handler) handleGetImageConfig(request *restful.Request, response *restful.Response) {
	secretName := request.QueryParameter("secret")
	namespace := request.PathParameter("namespace")
	image := request.QueryParameter("image")
	var secret *v1.Secret

	// empty secret means anoymous fetching
	if len(secretName) != 0 {
		object, err := h.resourceGetterV1alpha3.Get("secrets", namespace, secretName)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
		}
		secret = object.(*v1.Secret)
	}

	config, err := h.registryHelper.Config(secret, image)
	if err != nil {
		canonicalizeRegistryError(request, response, err)
		return
	}

	response.WriteHeaderAndJson(http.StatusOK, config, restful.MIME_JSON)
}

// handleGetRepositoryTags fetchs all tags of given repository, no paging.
func (h *Handler) handleGetRepositoryTags(request *restful.Request, response *restful.Response) {
	secretName := request.QueryParameter("secret")
	namespace := request.PathParameter("namespace")
	repository := request.QueryParameter("repository")
	var secret *v1.Secret

	if len(repository) == 0 {
		api.HandleBadRequest(response, request, fmt.Errorf("empty repository name"))
		return
	}

	if len(secretName) != 0 {
		object, err := h.resourceGetterV1alpha3.Get("secrets", namespace, secretName)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
		}
		secret = object.(*v1.Secret)
	}

	tags, err := h.registryHelper.ListRepositoryTags(secret, repository)
	if err != nil {
		canonicalizeRegistryError(request, response, err)
		return
	}

	response.WriteHeaderAndJson(http.StatusOK, tags, restful.MIME_JSON)
}

func canonicalizeRegistryError(request *restful.Request, response *restful.Response, err error) {
	if strings.Contains(err.Error(), "Unauthorized") {
		api.HandleUnauthorized(response, request, err)
	} else if strings.Contains(err.Error(), "MANIFEST_UNKNOWN") {
		api.HandleNotFound(response, request, err)
	} else {
		api.HandleBadRequest(response, request, err)
	}
}
