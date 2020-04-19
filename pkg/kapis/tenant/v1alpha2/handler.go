package v1alpha2

import (
	"errors"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/tenant"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
)

type tenantHandler struct {
	tenant tenant.Interface
}

func newTenantHandler(_ k8s.Client, factory informers.InformerFactory) *tenantHandler {

	return &tenantHandler{
		tenant: tenant.New(factory),
	}
}

func (h *tenantHandler) ListWorkspaces(req *restful.Request, resp *restful.Response) {
	user, ok := request.UserFrom(req.Request.Context())
	queryParam := query.ParseQueryParameter(req)

	if !ok {
		err := errors.New("cannot obtain user info")
		klog.Errorln(err)
		api.HandleForbidden(resp, nil, err)
		return
	}

	result, err := h.tenant.ListWorkspaces(user, queryParam)

	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *tenantHandler) ListNamespaces(req *restful.Request, resp *restful.Response) {
	user, ok := request.UserFrom(req.Request.Context())
	queryParam := query.ParseQueryParameter(req)

	if !ok {
		err := errors.New("cannot obtain user info")
		klog.Errorln(err)
		api.HandleForbidden(resp, nil, err)
		return
	}

	workspace := req.PathParameter("workspace")

	result, err := h.tenant.ListNamespaces(user, workspace, queryParam)

	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
}
