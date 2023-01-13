package proxies

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/filters"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
)

type unregisteredMiddleware struct {
	registeredGv   sets.String
	resourceGetter v1beta1.ResourceGetter
}

func NewUnregisteredMiddleware(c *restful.Container, resourceGetter v1beta1.ResourceGetter) filters.Middleware {
	middleware := &unregisteredMiddleware{
		registeredGv:   sets.NewString(),
		resourceGetter: resourceGetter,
	}

	for _, ws := range c.RegisteredWebServices() {
		rootPath := ws.RootPath()
		if strings.HasPrefix(rootPath, "/kapis") {
			middleware.registeredGv.Insert(rootPath)
		}
	}

	return middleware
}

func (u *unregisteredMiddleware) Handle(w http.ResponseWriter, req *http.Request) bool {
	if req.Method != http.MethodGet {
		return false
	}

	reqInfo, exist := request.RequestInfoFrom(req.Context())
	if !exist {
		return false
	}

	if reqInfo.IsKubernetesRequest {
		return false
	}

	gvr := schema.GroupVersionResource{
		Group:    reqInfo.APIGroup,
		Version:  reqInfo.APIVersion,
		Resource: reqInfo.Resource,
	}

	if gvr.Group == "" ||
		gvr.Version == "" ||
		gvr.Resource == "" {
		return false
	}

	rootPath := fmt.Sprintf("/kapis/%s/%s", gvr.Group, gvr.Version)
	if u.registeredGv.Has(rootPath) {
		return true
	}

	var (
		listReq bool
		q       *query.Query
	)
	restfulReq := restful.NewRequest(req)
	restfulResp := restful.NewResponse(w)
	if reqInfo.Name == "" {
		listReq = true
		q = query.ParseQueryParameter(restfulReq)
	}

	var (
		result interface{}
		err    error
	)
	if listReq {
		result, err = u.resourceGetter.ListResources(gvr, reqInfo.Namespace, q)
	} else {
		result, err = u.resourceGetter.GetResource(gvr, reqInfo.Name, reqInfo.Namespace)
	}
	handleResponse(result, err, restfulResp, restfulReq)
	return true
}

func handleResponse(result interface{}, err error, resp *restful.Response, req *restful.Request) {
	resp.SetRequestAccepts(restful.MIME_JSON)
	if err != nil {
		if err == v1beta1.ErrResourceNotSupported {
			api.HandleBadRequest(resp, req, err)
			return
		}
		klog.Error(err)
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}
