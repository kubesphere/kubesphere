/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha3

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/emicklei/go-restful/v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/components"
	v2 "kubesphere.io/kubesphere/pkg/models/registries/v2"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/simple/client/overview"
)

var (
	ClusterMetricNames = []string{
		overview.NamespaceCount, overview.PodCount, overview.DeploymentCount,
		overview.StatefulSetCount, overview.DaemonSetCount, overview.JobCount,
		overview.CronJobCount, overview.PersistentVolumeClaimCount, overview.ServiceCount,
		overview.IngressCount, overview.ClusterRoleBindingCount, overview.ClusterRoleCount,
	}

	NamespaceMetricNames = []string{
		overview.PodCount, overview.DeploymentCount, overview.StatefulSetCount,
		overview.DaemonSetCount, overview.JobCount, overview.CronJobCount,
		overview.PersistentVolumeClaimCount, overview.ServiceCount,
		overview.IngressCount, overview.RoleCount, overview.RoleBindingCount,
	}
)

type handler struct {
	resourceGetterV1alpha3 *resourcev1alpha3.Getter
	componentsGetter       components.Getter
	registryHelper         v2.RegistryHelper
	counter                overview.Counter
}

func (h *handler) GetResources(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	resourceType := request.PathParameter("resources")
	name := request.PathParameter("name")

	// use informers to retrieve resources
	result, err := h.resourceGetterV1alpha3.Get(resourceType, namespace, name)
	if err != nil {
		if err == resourcev1alpha3.ErrResourceNotSupported {
			api.HandleNotFound(response, request, err)
			return
		}
		klog.Error(err)
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

// ListResources retrieves resources
func (h *handler) ListResources(request *restful.Request, response *restful.Response) {
	query := query.ParseQueryParameter(request)
	resourceType := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")

	result, err := h.resourceGetterV1alpha3.List(resourceType, namespace, query)
	if err != nil {
		if err == resourcev1alpha3.ErrResourceNotSupported {
			api.HandleNotFound(response, request, err)
			return
		}
		klog.Error(err)
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *handler) GetComponentStatus(request *restful.Request, response *restful.Response) {
	component := request.PathParameter("component")
	result, err := h.componentsGetter.GetComponentStatus(component)
	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}
	response.WriteEntity(result)
}

func (h *handler) GetSystemHealthStatus(request *restful.Request, response *restful.Response) {
	result, err := h.componentsGetter.GetSystemHealthStatus()
	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}

// get all componentsHandler
func (h *handler) GetComponents(request *restful.Request, response *restful.Response) {
	result, err := h.componentsGetter.GetAllComponentsStatus()
	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}

// VerifyImageRepositorySecret verifies image secret against registry, it takes k8s.io/api/core/v1/types.Secret
// as input, and authenticate registry with credential specified. Returns http.StatusOK if authenticate successfully,
// returns http.StatusUnauthorized if failed.
func (h *handler) VerifyImageRepositorySecret(request *restful.Request, response *restful.Response) {
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

// GetImageConfig fetches container image spec described in https://github.com/opencontainers/image-spec/blob/main/manifest.md
func (h *handler) GetImageConfig(request *restful.Request, response *restful.Response) {
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

// GetRepositoryTags fetchs all tags of given repository, no paging.
func (h *handler) GetRepositoryTags(request *restful.Request, response *restful.Response) {
	secretName := request.QueryParameter("secret")
	namespace := request.PathParameter("namespace")
	repository := request.QueryParameter("repository")

	q := query.ParseQueryParameter(request)

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

	if !q.Ascending {
		sort.Sort(sort.Reverse(sort.StringSlice(tags.Tags)))
	}
	startIndex, endIndex := q.Pagination.GetValidPagination(len(tags.Tags))
	tags.Tags = tags.Tags[startIndex:endIndex]

	response.WriteHeaderAndJson(http.StatusOK, tags, restful.MIME_JSON)
}

func (h *handler) GetClusterOverview(request *restful.Request, response *restful.Response) {
	metrics, err := h.counter.GetMetrics(ClusterMetricNames, "", "", "cluster")
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	_ = response.WriteEntity(metrics)
}

func (h *handler) GetNamespaceOverview(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	metrics, err := h.counter.GetMetrics(NamespaceMetricNames, namespace, "", "namespace")
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	_ = response.WriteEntity(metrics)
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
