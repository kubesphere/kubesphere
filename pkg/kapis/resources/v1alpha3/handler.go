package v1alpha3

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/components"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"net/http"
)

type Handler struct {
	namespacedResourceGetter *resource.NamespacedResourceGetter
	componentsGetter         components.ComponentsGetter
}

func New(client k8s.Client) *Handler {
	factory := informers.NewInformerFactories(client.Kubernetes(), client.KubeSphere(), client.S2i(), client.Application())

	return &Handler{
		namespacedResourceGetter: resource.New(factory),
		componentsGetter:         components.NewComponentsGetter(factory.KubernetesSharedInformerFactory()),
	}
}

func (h Handler) handleGetNamespacedResource(request *restful.Request, response *restful.Response) {
	resource := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")

	result, err := h.namespacedResourceGetter.Get(resource, namespace, name)
	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

// handleListNamedResource retrieves namespaced scope resources
func (h Handler) handleListNamespacedResource(request *restful.Request, response *restful.Response) {
	query := query.ParseQueryParameter(request)
	resource := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")

	result, err := h.namespacedResourceGetter.List(resource, namespace, query)
	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (h Handler) handleGetComponentStatus(request *restful.Request, response *restful.Response) {
	component := request.PathParameter("component")
	result, err := h.componentsGetter.GetComponentStatus(component)

	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (h Handler) handleGetSystemHealthStatus(request *restful.Request, response *restful.Response) {
	result, err := h.componentsGetter.GetSystemHealthStatus()

	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

// get all componentsHandler
func (h Handler) handleGetComponents(request *restful.Request, response *restful.Response) {

	result, err := h.componentsGetter.GetAllComponentsStatus()

	if err != nil {
		api.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}
