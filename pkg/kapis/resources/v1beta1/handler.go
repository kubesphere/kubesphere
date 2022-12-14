package v1beta1

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/components"
	v2 "kubesphere.io/kubesphere/pkg/models/registries/v2"
	"kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
)

func New(lister v1beta1.ResourceLister, componentsGetter components.ComponentsGetter) Handler {
	return Handler{
		lister:           lister,
		registryHelper:   v2.NewRegistryHelper(),
		componentsGetter: componentsGetter,
	}
}

type Handler struct {
	lister           v1beta1.ResourceLister
	registryHelper   v2.RegistryHelper
	componentsGetter components.ComponentsGetter
}

func (h *Handler) getResources(request *restful.Request, response *restful.Response) {
	group := request.PathParameter("group")
	version := request.PathParameter("version")
	resource := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	object, err := h.lister.GetResource(gvr, name, namespace)
	if err != nil {
		if err == v1beta1.ErrResourceNotSupported {
			api.HandleNotFound(response, request, err)
			return
		}
		klog.Error(err)
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(object)
}

func (h *Handler) listResources(request *restful.Request, response *restful.Response) {
	group := request.PathParameter("group")
	version := request.PathParameter("version")
	resource := request.PathParameter("resources")
	namespace := request.PathParameter("namespace")
	q := query.ParseQueryParameter(request)

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	objectList, err := h.lister.ListResources(gvr, namespace, q)
	if err != nil {
		if err == v1beta1.ErrResourceNotSupported {
			api.HandleNotFound(response, request, err)
			return
		}
		klog.Error(err)
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(objectList)
}
