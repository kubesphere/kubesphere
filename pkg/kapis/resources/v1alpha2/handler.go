package v1alpha2

import (
	"fmt"
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/components"
	"kubesphere.io/kubesphere/pkg/models/git"
	"kubesphere.io/kubesphere/pkg/models/quotas"
	"kubesphere.io/kubesphere/pkg/models/registries"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/resource"
	"kubesphere.io/kubesphere/pkg/models/revisions"
	"kubesphere.io/kubesphere/pkg/models/routers"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"net/http"
	"strconv"
	"strings"
)

type resourceHandler struct {
	resourcesGetter     *resource.ResourceGetter
	componentsGetter    components.ComponentsGetter
	resourceQuotaGetter quotas.ResourceQuotaGetter
	revisionGetter      revisions.RevisionGetter
	routerOperator      routers.RouterOperator
	gitVerifier         git.GitVerifier
	registryGetter      registries.RegistryGetter
}

func newResourceHandler(client k8s.Client) *resourceHandler {

	factory := informers.NewInformerFactories(client.Kubernetes(), client.KubeSphere(), client.S2i(), client.Application())

	return &resourceHandler{
		resourcesGetter:     resource.NewResourceGetter(factory),
		componentsGetter:    components.NewComponentsGetter(factory.KubernetesSharedInformerFactory()),
		resourceQuotaGetter: quotas.NewResourceQuotaGetter(factory.KubernetesSharedInformerFactory()),
		revisionGetter:      revisions.NewRevisionGetter(factory.KubernetesSharedInformerFactory()),
		routerOperator:      routers.NewRouterOperator(client.Kubernetes(), factory.KubernetesSharedInformerFactory()),
		gitVerifier:         git.NewGitVerifier(factory.KubernetesSharedInformerFactory()),
		registryGetter:      registries.NewRegistryGetter(factory.KubernetesSharedInformerFactory()),
	}
}

func (r *resourceHandler) handleGetNamespacedResources(request *restful.Request, response *restful.Response) {
	r.handleListNamespaceResources(request, response)
}

func (r *resourceHandler) handleListNamespaceResources(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	resource := request.PathParameter("resources")
	orderBy := params.GetStringValueWithDefault(request, params.OrderByParam, v1alpha2.CreateTime)
	limit, offset := params.ParsePaging(request)
	reverse := params.GetBoolValueWithDefault(request, params.ReverseParam, false)
	conditions, err := params.ParseConditions(request)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, err)
		return
	}

	result, err := r.resourcesGetter.ListResources(namespace, resource, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteAsJson(result)
}

func (r *resourceHandler) handleGetSystemHealthStatus(_ *restful.Request, response *restful.Response) {
	result, err := r.componentsGetter.GetSystemHealthStatus()

	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteAsJson(result)
}

func (r *resourceHandler) handleGetComponentStatus(request *restful.Request, response *restful.Response) {
	component := request.PathParameter("component")
	result, err := r.componentsGetter.GetComponentStatus(component)

	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteAsJson(result)
}

func (r *resourceHandler) handleGetComponents(_ *restful.Request, response *restful.Response) {
	result, err := r.componentsGetter.GetAllComponentsStatus()

	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteAsJson(result)
}

func (r *resourceHandler) handleGetClusterQuotas(_ *restful.Request, response *restful.Response) {
	result, err := r.resourceQuotaGetter.GetClusterQuota()
	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteAsJson(result)
}

func (r *resourceHandler) handleGetNamespaceQuotas(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	quota, err := r.resourceQuotaGetter.GetNamespaceQuota(namespace)

	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteAsJson(quota)
}

func (r *resourceHandler) handleGetDaemonSetRevision(request *restful.Request, response *restful.Response) {
	daemonset := request.PathParameter("daemonset")
	namespace := request.PathParameter("namespace")
	revision, err := strconv.Atoi(request.PathParameter("revision"))

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := r.revisionGetter.GetDaemonSetRevision(namespace, daemonset, revision)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(result)
}

func (r *resourceHandler) handleGetDeploymentRevision(request *restful.Request, response *restful.Response) {
	deploy := request.PathParameter("deployment")
	namespace := request.PathParameter("namespace")
	revision := request.PathParameter("revision")

	result, err := r.revisionGetter.GetDeploymentRevision(namespace, deploy, revision)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(result)
}

func (r *resourceHandler) handleGetStatefulSetRevision(request *restful.Request, response *restful.Response) {
	statefulset := request.PathParameter("statefulset")
	namespace := request.PathParameter("namespace")
	revision, err := strconv.Atoi(request.PathParameter("revision"))
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := r.revisionGetter.GetStatefulSetRevision(namespace, statefulset, revision)
	if err != nil {
		api.HandleInternalError(response, err)
		return
	}
	response.WriteAsJson(result)
}

// Get ingress controller service for specified namespace
func (r *resourceHandler) handleGetRouter(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	router, err := r.routerOperator.GetRouter(namespace)
	if err != nil {
		if k8serr.IsNotFound(err) {
			response.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		} else {
			api.HandleInternalError(response, err)
		}
		return
	}

	response.WriteAsJson(router)
}

// Create ingress controller and related services
func (r *resourceHandler) handleCreateRouter(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	newRouter := api.Router{}
	err := request.ReadEntity(&newRouter)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("wrong annotations, missing key or value")))
		return
	}

	routerType := v1.ServiceTypeNodePort
	if strings.Compare(strings.ToLower(newRouter.RouterType), "loadbalancer") == 0 {
		routerType = v1.ServiceTypeLoadBalancer
	}

	router, err := r.routerOperator.CreateRouter(namespace, routerType, newRouter.Annotations)
	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteAsJson(router)
}

// Delete ingress controller and services
func (r *resourceHandler) handleDeleteRouter(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	router, err := r.routerOperator.DeleteRouter(namespace)
	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteAsJson(router)
}

func (r *resourceHandler) handleUpdateRouter(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	newRouter := api.Router{}
	err := request.ReadEntity(&newRouter)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	var routerType = v1.ServiceTypeNodePort
	if strings.Compare(strings.ToLower(newRouter.RouterType), "loadbalancer") == 0 {
		routerType = v1.ServiceTypeLoadBalancer
	}
	router, err := r.routerOperator.UpdateRouter(namespace, routerType, newRouter.Annotations)

	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteAsJson(router)
}

func (r *resourceHandler) handleVerifyGitCredential(request *restful.Request, response *restful.Response) {

	var credential api.GitCredential
	err := request.ReadEntity(&credential)
	if err != nil {
		api.HandleBadRequest(response, err)
		return
	}

	err = r.gitVerifier.VerifyGitCredential(credential.RemoteUrl, credential.SecretRef.Namespace, credential.SecretRef.Name)
	if err != nil {
		api.HandleBadRequest(response, err)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func (r *resourceHandler) handleVerifyRegistryCredential(request *restful.Request, response *restful.Response) {
	var credential api.RegistryCredential
	err := request.ReadEntity(&credential)
	if err != nil {
		api.HandleBadRequest(response, err)
		return
	}

	err = r.registryGetter.VerifyRegistryCredential(credential)
	if err != nil {
		api.HandleBadRequest(response, err)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func (r *resourceHandler) handleGetRegistryEntry(request *restful.Request, response *restful.Response) {
	imageName := request.QueryParameter("image")
	namespace := request.QueryParameter("namespace")
	secretName := request.QueryParameter("secret")

	detail, err := r.registryGetter.GetEntry(namespace, secretName, imageName)
	if err != nil {
		api.HandleBadRequest(response, err)
		return
	}

	response.WriteAsJson(detail)
}

func (r *resourceHandler) handleGetNamespacedAbnormalWorkloads(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	result := api.Workloads{
		Namespace: namespace,
		Count:     make(map[string]int),
	}

	for _, workloadType := range []string{api.ResourceKindDeployment, api.ResourceKindStatefulSet, api.ResourceKindDaemonSet, api.ResourceKindJob, api.ResourceKindPersistentVolumeClaim} {
		var notReadyStatus string

		switch workloadType {
		case api.ResourceKindPersistentVolumeClaim:
			notReadyStatus = strings.Join([]string{v1alpha2.StatusPending, v1alpha2.StatusLost}, "|")
		case api.ResourceKindJob:
			notReadyStatus = v1alpha2.StatusFailed
		default:
			notReadyStatus = v1alpha2.StatusUpdating
		}

		res, err := r.resourcesGetter.ListResources(namespace, workloadType, &params.Conditions{Match: map[string]string{v1alpha2.Status: notReadyStatus}}, "", false, -1, 0)
		if err != nil {
			api.HandleInternalError(response, err)
		}

		result.Count[workloadType] = len(res.Items)
	}

	response.WriteAsJson(result)

}

func (r *resourceHandler) handleGetAbnormalWorkloads(request *restful.Request, response *restful.Response) {
	r.handleGetNamespacedAbnormalWorkloads(request, response)
}
