/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha3

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/models/tenant"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

type handler struct {
	tenant tenant.Interface
	auth   authorizer.Authorizer
	client runtimeclient.Client
}

func NewHandler(client runtimeclient.Client, k8sVersion *semver.Version, clusterClient clusterclient.Interface,
	am am.AccessManagementInterface, im im.IdentityManagementInterface, authorizer authorizer.Authorizer) rest.Handler {
	return &handler{
		tenant: tenant.New(client, k8sVersion, clusterClient, am, im, authorizer),
		client: client,
		auth:   authorizer,
	}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

func (h *handler) ListWorkspaceTemplates(req *restful.Request, resp *restful.Response) {
	authenticated, ok := request.UserFrom(req.Request.Context())
	queryParam := query.ParseQueryParameter(req)

	if !ok {
		err := fmt.Errorf("cannot obtain user info")
		klog.Errorln(err)
		api.HandleForbidden(resp, nil, err)
		return
	}

	result, err := h.tenant.ListWorkspaceTemplates(authenticated, queryParam)

	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	_ = resp.WriteEntity(result)
}

func (h *handler) ListNamespaces(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	queryParam := query.ParseQueryParameter(req)

	var workspaceMember user.Info
	if username := req.PathParameter("workspacemember"); username != "" {
		workspaceMember = &user.DefaultInfo{
			Name: username,
		}
	} else {
		requestUser, ok := request.UserFrom(req.Request.Context())
		if !ok {
			err := fmt.Errorf("cannot obtain user info")
			klog.Errorln(err)
			api.HandleForbidden(resp, nil, err)
			return
		}
		workspaceMember = requestUser
	}

	result, err := h.tenant.ListNamespaces(workspaceMember, workspace, queryParam)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	_ = resp.WriteEntity(result)
}

func (h *handler) DescribeWorkspaceTemplate(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	workspace, err := h.tenant.DescribeWorkspaceTemplate(workspaceName)
	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}
	_ = response.WriteEntity(workspace)
}

func (h *handler) ListWorkspaceClusters(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")

	result, err := h.tenant.ListWorkspaceClusters(workspaceName)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	_ = response.WriteEntity(result)
}

func (h *handler) DescribeNamespace(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	namespaceName := request.PathParameter("namespace")
	ns, err := h.tenant.DescribeNamespace(workspaceName, namespaceName)

	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	_ = response.WriteEntity(ns)
}

func (h *handler) ListClusters(r *restful.Request, response *restful.Response) {
	authenticated, ok := request.UserFrom(r.Request.Context())

	if !ok {
		_ = response.WriteEntity([]interface{}{})
		return
	}

	queryParam := query.ParseQueryParameter(r)
	result, err := h.tenant.ListClusters(authenticated, queryParam)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, r, err)
			return
		}
		api.HandleInternalError(response, r, err)
		return
	}

	_ = response.WriteEntity(result)
}

func (h *handler) ListWorkspaces(req *restful.Request, resp *restful.Response) {
	queryParam := query.ParseQueryParameter(req)
	authenticated, ok := request.UserFrom(req.Request.Context())
	if !ok {
		err := fmt.Errorf("cannot obtain user info")
		klog.Errorln(err)
		api.HandleForbidden(resp, nil, err)
		return
	}

	result, err := h.tenant.ListWorkspaces(authenticated, queryParam)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	_ = resp.WriteEntity(result)
}

func (h *handler) GetWorkspace(request *restful.Request, response *restful.Response) {
	workspace, err := h.tenant.GetWorkspace(request.PathParameter("workspace"))
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	_ = response.WriteEntity(workspace)
}
