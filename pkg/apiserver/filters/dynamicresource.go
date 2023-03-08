package filters

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
)

type DynamicResourceHandler struct {
	v1beta1.ResourceManager
	serviceErrorHandleFallback restful.ServiceErrorHandleFunction
}

func NewDynamicResourceHandle(serviceErrorHandleFallback restful.ServiceErrorHandleFunction, resourceGetter v1beta1.ResourceManager) *DynamicResourceHandler {
	return &DynamicResourceHandler{
		ResourceManager:            resourceGetter,
		serviceErrorHandleFallback: serviceErrorHandleFallback,
	}
}

func (d *DynamicResourceHandler) HandleServiceError(serviceError restful.ServiceError, req *restful.Request, w *restful.Response) {
	// only not found error will be handled
	if serviceError.Code != http.StatusNotFound {
		d.serviceErrorHandleFallback(serviceError, req, w)
		return
	}

	// TODO support write operation and workspace scope API
	if req.Request.Method != http.MethodGet {
		d.serviceErrorHandleFallback(serviceError, req, w)
		return
	}

	reqInfo, exist := request.RequestInfoFrom(req.Request.Context())
	if !exist {
		responsewriters.InternalError(w, req.Request, fmt.Errorf("no RequestInfo found in the context"))
		return
	}

	if reqInfo.IsKubernetesRequest {
		d.serviceErrorHandleFallback(serviceError, req, w)
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    reqInfo.APIGroup,
		Version:  reqInfo.APIVersion,
		Resource: reqInfo.Resource,
	}

	if gvr.Group == "" ||
		gvr.Version == "" ||
		gvr.Resource == "" {
		d.serviceErrorHandleFallback(serviceError, req, w)
		return
	}

	served, err := d.IsServed(gvr)
	if err != nil {
		responsewriters.InternalError(w, req.Request, err)
		return
	}

	if !served {
		d.serviceErrorHandleFallback(serviceError, req, w)
		return
	}

	var result interface{}
	if reqInfo.Verb == "list" {
		result, err = d.ListResources(req.Request.Context(), gvr, reqInfo.Namespace, query.ParseQueryParameter(req))
	} else {
		result, err = d.GetResource(req.Request.Context(), gvr, reqInfo.Name, reqInfo.Namespace)
	}

	if err != nil {
		if meta.IsNoMatchError(err) {
			d.serviceErrorHandleFallback(serviceError, req, w)
			return
		}
		api.HandleError(w, req, err)
		return
	}

	w.WriteAsJson(result)
}
