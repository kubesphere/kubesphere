package filters

import (
	"fmt"
	"io"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
)

var NotSupportedVerbError = fmt.Errorf("not supported verb")

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

	var object client.Object
	if reqInfo.Verb == request.VerbCreate || reqInfo.Verb == request.VerbUpdate || reqInfo.Verb == request.VerbPatch {
		rawData, err := io.ReadAll(req.Request.Body)
		if err != nil {
			api.HandleError(w, req, err)
			return
		}

		object, err = d.CreateObjectFromRawData(gvr, rawData)
		if err != nil {
			api.HandleError(w, req, err)
			return
		}
	}

	var result interface{}

	switch reqInfo.Verb {
	case request.VerbGet:
		result, err = d.GetResource(req.Request.Context(), gvr, reqInfo.Namespace, reqInfo.Name)
	case request.VerbList:
		result, err = d.ListResources(req.Request.Context(), gvr, reqInfo.Namespace, query.ParseQueryParameter(req))
	case request.VerbCreate:
		err = d.CreateResource(req.Request.Context(), object)
	case request.VerbUpdate:
		err = d.UpdateResource(req.Request.Context(), object)
	case request.VerbDelete:
		err = d.DeleteResource(req.Request.Context(), gvr, reqInfo.Namespace, reqInfo.Name)
	case request.VerbPatch:
		err = d.PatchResource(req.Request.Context(), object)
	default:
		err = NotSupportedVerbError
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
