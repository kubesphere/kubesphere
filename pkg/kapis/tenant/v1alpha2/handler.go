package v1alpha2

import (
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/tenant"
	apierr "kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"net/http"
	"strings"
)

type tenantHandler struct {
	tenant tenant.Interface
	am     iam.AccessManagementInterface
}

func newTenantHandler(k8sClient k8s.Client, db *mysql.Database) *tenantHandler {
	factory := informers.NewInformerFactories(k8sClient.Kubernetes(), k8sClient.KubeSphere(), k8sClient.S2i(), k8sClient.Application())

	return &tenantHandler{
		tenant: tenant.New(k8sClient.Kubernetes(), factory.KubernetesSharedInformerFactory(), factory.KubeSphereSharedInformerFactory(), db),
		am:     iam.NewAMOperator(factory.KubernetesSharedInformerFactory()),
	}
}

func (h *tenantHandler) ListWorkspaceRules(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	rules, err := h.tenant.GetWorkspaceSimpleRules(workspace, username)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteEntity(rules)
}

func (h *tenantHandler) ListWorkspaces(req *restful.Request, resp *restful.Response) {
	username := req.HeaderParameter(constants.UserNameHeader)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, v1alpha2.CreateTime)
	limit, offset := params.ParsePaging(req)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, true)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.Errorln(err)
		api.HandleBadRequest(resp, err)
		return
	}

	result, err := h.tenant.ListWorkspaces(username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		api.HandleInternalError(resp, err)
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
			api.HandleNotFound(resp, err)
		} else {
			api.HandleInternalError(resp, err)
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
		api.HandleBadRequest(resp, err)
		return
	}

	conditions.Match[constants.WorkspaceLabelKey] = workspace

	result, err := h.tenant.ListNamespaces(username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		api.HandleInternalError(resp, err)
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
		api.HandleNotFound(resp, err)
		return
	}

	_, err = h.tenant.DescribeWorkspace("", workspace)

	if err != nil {
		if k8serr.IsNotFound(err) {
			api.HandleForbidden(resp, err)
		} else {
			api.HandleInternalError(resp, err)
		}
		return
	}

	created, err := h.tenant.CreateNamespace(workspace, &namespace, username)

	if err != nil {
		if k8serr.IsAlreadyExists(err) {
			resp.WriteHeaderAndEntity(http.StatusConflict, err)
		} else {
			api.HandleInternalError(resp, err)
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
			api.HandleNotFound(resp, err)
		} else {
			api.HandleInternalError(resp, err)
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
		api.HandleBadRequest(resp, err)
		return
	}
	conditions.Match["workspace"] = workspace

	result, err := h.tenant.ListDevopsProjects(username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *tenantHandler) GetDevOpsProjectsCount(req *restful.Request, resp *restful.Response) {
	username := req.HeaderParameter(constants.UserNameHeader)

	result, err := h.tenant.ListDevopsProjects(username, nil, "", false, 1, 0)
	if err != nil {
		api.HandleInternalError(resp, err)
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
		api.HandleInternalError(resp, err)
		return
	}

	err = h.tenant.DeleteDevOpsProject(username, projectId)

	if err != nil {
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteEntity(apierr.None)
}

func (h *tenantHandler) CreateDevopsProject(req *restful.Request, resp *restful.Response) {

}

func (h *tenantHandler) ListNamespaceRules(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	username := req.HeaderParameter(constants.UserNameHeader)

	rules, err := h.tenant.GetNamespaceSimpleRules(namespace, username)

	if err != nil {
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteAsJson(rules)
}

func (h *tenantHandler) ListDevopsRules(req *restful.Request, resp *restful.Response) {

	devops := req.PathParameter("devops")
	username := req.HeaderParameter(constants.UserNameHeader)

	rules, err := h.tenant.GetUserDevopsSimpleRules(username, devops)

	if err != nil {
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteAsJson(rules)
}

//TODO(wansir): We need move this part to logging module
//func (h *tenantHandler) LogQuery(req *restful.Request, resp *restful.Response) {
//	operation := req.QueryParameter("operation")
//	req, err := h.regenerateLoggingRequest(req)
//	switch {
//	case err != nil:
//		api.HandleInternalError(resp, err)
//	case req != nil:
//		loggingv1alpha2.Get(req, loggingv1alpha2.LevelCluster, h.k8s, h.lo, resp)
//	default:
//		if operation == "export" {
//			resp.Header().Set(restful.HEADER_ContentType, "text/plain")
//			resp.Header().Set("Content-Disposition", "attachment")
//			resp.Write(nil)
//		} else {
//			resp.WriteAsJson(v1alpha2.APIResponse{Logs: new(loggingclient.Logs)})
//		}
//	}
//}

// override namespace query conditions
//TODO(wansir): We need move this part to logging module
func (h *tenantHandler) regenerateLoggingRequest(req *restful.Request) (*restful.Request, error) {

	username := req.HeaderParameter(constants.UserNameHeader)

	// regenerate the request for log query
	newUrl := net.FormatURL("http", "127.0.0.1", 80, "/kapis/logging.kubesphere.io/v1alpha2/cluster")
	values := req.Request.URL.Query()

	clusterRoleRules, err := h.am.GetClusterPolicyRules(username)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	hasClusterLogAccess := iam.RulesMatchesRequired(clusterRoleRules, rbacv1.PolicyRule{Verbs: []string{"get"}, Resources: []string{"*"}, APIGroups: []string{"logging.kubesphere.io"}})
	// if the user is not a cluster admin
	if !hasClusterLogAccess {
		queryNamespaces := strings.Split(req.QueryParameter("namespaces"), ",")
		// then the user can only view logs of namespaces he belongs to
		namespaces := make([]string, 0)
		roles, err := h.am.GetRoles("", username)
		if err != nil {
			klog.Errorln(err)
			return nil, err
		}
		for _, role := range roles {
			if !sliceutil.HasString(namespaces, role.Namespace) && iam.RulesMatchesRequired(role.Rules, rbacv1.PolicyRule{Verbs: []string{"get"}, Resources: []string{"*"}, APIGroups: []string{"logging.kubesphere.io"}}) {
				namespaces = append(namespaces, role.Namespace)
			}
		}

		// if the user belongs to no namespace
		// then no log visible
		if len(namespaces) == 0 {
			return nil, nil
		} else if len(queryNamespaces) == 1 && queryNamespaces[0] == "" {
			values.Set("namespaces", strings.Join(namespaces, ","))
		} else {
			inter := intersection(queryNamespaces, namespaces)
			if len(inter) == 0 {
				return nil, nil
			}
			values.Set("namespaces", strings.Join(inter, ","))
		}
	}

	newUrl.RawQuery = values.Encode()

	// forward the request to logging model
	newHttpRequest, _ := http.NewRequest(http.MethodGet, newUrl.String(), nil)
	return restful.NewRequest(newHttpRequest), nil
}

func intersection(s1, s2 []string) (inter []string) {
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[e] = true
	}
	for _, e := range s2 {
		// If elements present in the hashmap then append intersection list.
		if hash[e] {
			inter = append(inter, e)
		}
	}
	//Remove dups from slice.
	inter = removeDups(inter)
	return
}

//Remove dups from slice.
func removeDups(elements []string) (nodups []string) {
	encountered := make(map[string]bool)
	for _, element := range elements {
		if !encountered[element] {
			nodups = append(nodups, element)
			encountered[element] = true
		}
	}
	return
}
