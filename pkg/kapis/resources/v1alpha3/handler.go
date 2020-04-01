package v1alpha3

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/components"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"net/http"
)

type Handler struct {
	namespacedResourceGetter *resource.ResourceGetter
	componentsGetter         components.ComponentsGetter
}

func New(factory informers.InformerFactory) *Handler {
	return &Handler{
		namespacedResourceGetter: resource.New(factory),
		componentsGetter:         components.NewComponentsGetter(factory.KubernetesSharedInformerFactory()),
	}
}

// handleListResources retrieves resources
func (h Handler) handleListResources(request *restful.Request, response *restful.Response) {
	query := query.ParseQueryParameter(request)
	resource := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")

	result, err := h.namespacedResourceGetter.List(resource, namespace, query)
	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}

func (h Handler) handleGetComponentStatus(request *restful.Request, response *restful.Response) {
	component := request.PathParameter("component")
	result, err := h.componentsGetter.GetComponentStatus(component)

	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (h Handler) handleGetSystemHealthStatus(request *restful.Request, response *restful.Response) {
	result, err := h.componentsGetter.GetSystemHealthStatus()

	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

// get all componentsHandler
func (h Handler) handleGetComponents(request *restful.Request, response *restful.Response) {

	result, err := h.componentsGetter.GetAllComponentsStatus()

	if err != nil {
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}
