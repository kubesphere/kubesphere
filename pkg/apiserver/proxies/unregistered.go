package proxies

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/filters"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
	"kubesphere.io/kubesphere/pkg/utils/pathutil"
)

const (
	group     = "group"
	version   = "version"
	resources = "resources"
	namespace = "namespace"
	name      = "name"
)

type unregisteredMiddleware struct {
	registeredGv       map[string]bool
	resourceGetter     v1beta1.ResourceGetter
	parameterExtractor pathutil.PathParameterExtractor
}

func NewUnregisteredMiddleware(c *restful.Container, resourceGetter v1beta1.ResourceGetter) filters.Middleware {
	middleware := &unregisteredMiddleware{
		registeredGv:       make(map[string]bool, 0),
		resourceGetter:     resourceGetter,
		parameterExtractor: pathutil.NewKapisPathParameterExtractor(),
	}

	for _, ws := range c.RegisteredWebServices() {
		rootPath := ws.RootPath()
		if strings.HasPrefix(rootPath, "/kapis") {
			middleware.registeredGv[rootPath] = true
		}
	}

	return middleware
}

func (u *unregisteredMiddleware) Handle(w http.ResponseWriter, req *http.Request) bool {
	if req.Method != http.MethodGet {
		return false
	}

	requestPath := req.URL.Path
	if !strings.HasPrefix(requestPath, "/kapis") {
		return false
	}

	parameter := u.parameterExtractor.Extract(requestPath)

	gvr := schema.GroupVersionResource{
		Group:    parameter[group],
		Version:  parameter[version],
		Resource: parameter[resources],
	}

	if gvr.Group == "" ||
		gvr.Version == "" ||
		gvr.Resource == "" {
		return false
	}

	rootPath := fmt.Sprintf("/kapis/%s/%s", gvr.Group, gvr.Version)
	if u.registeredGv[rootPath] {
		return false
	}

	var (
		listReq bool
		q       *query.Query
	)
	restfulReq := restful.NewRequest(req)
	restfulResp := restful.NewResponse(w)
	if parameter[name] == "" {
		listReq = true
		q = query.ParseQueryParameter(restfulReq)
	}

	var (
		result interface{}
		err    error
	)
	if listReq {
		result, err = u.resourceGetter.ListResources(gvr, parameter[namespace], q)
	} else {
		result, err = u.resourceGetter.GetResource(gvr, parameter[name], parameter[namespace])
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
