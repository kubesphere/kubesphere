package v1alpha2

import (
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/tenant"
	apierr "kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"net/http"
)

type tenantHandler struct {
	tenant tenant.Interface
	am     am.AccessManagementInterface
}

func newTenantHandler(k8sClient k8s.Client, factory informers.InformerFactory, db *mysql.Database) *tenantHandler {

	return &tenantHandler{
		tenant: tenant.New(k8sClient.Kubernetes(), factory.KubernetesSharedInformerFactory(), factory.KubeSphereSharedInformerFactory(), db),
		am:     am.NewAMOperator(k8sClient.Kubernetes(), factory.KubernetesSharedInformerFactory()),
	}
}

func (h *tenantHandler) ListWorkspaces(req *restful.Request, resp *restful.Response) {
	username := req.HeaderParameter(constants.UserNameHeader)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, v1alpha2.CreateTime)
	limit, offset := params.ParsePaging(req)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, true)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.Errorln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	result, err := h.tenant.ListWorkspaces(username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(result)
}

func (h *tenantHandler) DescribeWorkspace(req *restful.Request, resp *restful.Response) {
	username := req.HeaderParameter(constants.UserNameHeader)
	workspaceName := req.PathParameter("workspace")

	result, err := h.tenant.DescribeWorkspace(username, workspaceName)

	if err != nil {
		if k8serr.IsNotFound(err) {
			api.HandleNotFound(resp, nil, err)
		} else {
			api.HandleInternalError(resp, nil, err)
		}
		return
	}

	resp.WriteAsJson(result)
}

func (h *tenantHandler) ListNamespaces(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	username := req.PathParameter("member")
	// /workspaces/{workspace}/members/{username}/namespaces
	if username == "" {
		// /workspaces/{workspace}/namespaces
		username = req.HeaderParameter(constants.UserNameHeader)
	}

	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, v1alpha2.CreateTime)
	limit, offset := params.ParsePaging(req)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, true)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	conditions.Match[constants.WorkspaceLabelKey] = workspace

	result, err := h.tenant.ListNamespaces(username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	namespaces := make([]*v1.Namespace, 0)

	for _, item := range result.Items {
		namespaces = append(namespaces, item.(*v1.Namespace).DeepCopy())
	}

	namespaces = monitoring.GetNamespacesWithMetrics(namespaces)

	items := make([]interface{}, 0)

	for _, item := range namespaces {
		items = append(items, item)
	}

	result.Items = items

	resp.WriteAsJson(result)
}

func (h *tenantHandler) CreateNamespace(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)
	var namespace v1.Namespace
	err := req.ReadEntity(&namespace)
	if err != nil {
		api.HandleNotFound(resp, nil, err)
		return
	}

	_, err = h.tenant.DescribeWorkspace("", workspace)

	if err != nil {
		if k8serr.IsNotFound(err) {
			api.HandleForbidden(resp, nil, err)
		} else {
			api.HandleInternalError(resp, nil, err)
		}
		return
	}

	created, err := h.tenant.CreateNamespace(workspace, &namespace, username)

	if err != nil {
		if k8serr.IsAlreadyExists(err) {
			resp.WriteHeaderAndEntity(http.StatusConflict, err)
		} else {
			api.HandleInternalError(resp, nil, err)
		}
		return
	}
	resp.WriteAsJson(created)
}

func (h *tenantHandler) DeleteNamespace(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	namespace := req.PathParameter("namespace")

	err := h.tenant.DeleteNamespace(workspace, namespace)

	if err != nil {
		if k8serr.IsNotFound(err) {
			api.HandleNotFound(resp, nil, err)
		} else {
			api.HandleInternalError(resp, nil, err)
		}
		return
	}

	resp.WriteAsJson(apierr.None)
}

func (h *tenantHandler) ListDevopsProjects(req *restful.Request, resp *restful.Response) {

	workspace := req.PathParameter("workspace")
	username := req.PathParameter("member")
	if username == "" {
		username = req.HeaderParameter(constants.UserNameHeader)
	}
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, v1alpha2.CreateTime)
	limit, offset := params.ParsePaging(req)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, true)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	conditions.Match["workspace"] = workspace

	result, err := h.tenant.ListDevopsProjects(username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *tenantHandler) GetDevOpsProjectsCount(req *restful.Request, resp *restful.Response) {
	username := req.HeaderParameter(constants.UserNameHeader)

	result, err := h.tenant.ListDevopsProjects(username, nil, "", false, 1, 0)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteEntity(struct {
		Count int `json:"count"`
	}{Count: result.TotalCount})
}
func (h *tenantHandler) DeleteDevopsProject(req *restful.Request, resp *restful.Response) {
	projectId := req.PathParameter("devops")
	workspace := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	_, err := h.tenant.DescribeWorkspace("", workspace)

	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	err = h.tenant.DeleteDevOpsProject(username, projectId)

	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(apierr.None)
}

func (h *tenantHandler) CreateDevopsProject(req *restful.Request, resp *restful.Response) {

}
