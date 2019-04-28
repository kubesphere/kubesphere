/*

 Copyright 2019 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

*/
package tenant

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/net"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/logging"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/metrics"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/models/tenant"
	"kubesphere.io/kubesphere/pkg/models/workspaces"
	"kubesphere.io/kubesphere/pkg/params"

	"kubesphere.io/kubesphere/pkg/simple/client/elasticsearch"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"net/http"
	"strings"
)

func ListWorkspaceRules(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	rules, err := iam.GetUserWorkspaceSimpleRules(workspace, username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(rules)
}

func ListWorkspaces(req *restful.Request, resp *restful.Response) {
	username := req.HeaderParameter(constants.UserNameHeader)
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if orderBy == "" {
		orderBy = resources.CreateTime
		reverse = true
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := tenant.ListWorkspaces(username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}

func DescribeWorkspace(req *restful.Request, resp *restful.Response) {
	username := req.HeaderParameter(constants.UserNameHeader)
	workspaceName := req.PathParameter("workspace")

	result, err := tenant.DescribeWorkspace(username, workspaceName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}

func ListNamespaces(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	username := req.PathParameter("username")
	// /workspaces/{workspace}/members/{username}/namespaces
	if username == "" {
		// /workspaces/{workspace}/namespaces
		username = req.HeaderParameter(constants.UserNameHeader)
	}

	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	conditions.Match[constants.WorkspaceLabelKey] = workspace

	result, err := tenant.ListNamespaces(username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	namespaces := make([]*v1.Namespace, 0)

	for _, item := range result.Items {
		namespaces = append(namespaces, item.(*v1.Namespace).DeepCopy())
	}

	namespaces = metrics.GetNamespacesWithMetrics(namespaces)

	items := make([]interface{}, 0)

	for _, item := range namespaces {
		items = append(items, item)
	}

	result.Items = items

	resp.WriteAsJson(result)
}

func CreateNamespace(req *restful.Request, resp *restful.Response) {
	workspaceName := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)
	var namespace v1.Namespace
	err := req.ReadEntity(&namespace)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	workspace, err := tenant.GetWorkspace(workspaceName)

	err = checkResourceQuotas(workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusForbidden, errors.Wrap(err))
		return
	}

	if err != nil {
		if k8serr.IsNotFound(err) {
			resp.WriteHeaderAndEntity(http.StatusForbidden, errors.Wrap(err))
		} else {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		}
		return
	}

	created, err := tenant.CreateNamespace(workspaceName, &namespace, username)

	if err != nil {
		if k8serr.IsAlreadyExists(err) {
			resp.WriteHeaderAndEntity(http.StatusConflict, err)
		} else {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, err)
		}
		return
	}
	resp.WriteAsJson(created)
}

func DeleteNamespace(req *restful.Request, resp *restful.Response) {
	workspaceName := req.PathParameter("workspace")
	namespaceName := req.PathParameter("namespace")

	err := workspaces.DeleteNamespace(workspaceName, namespaceName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func checkResourceQuotas(wokrspace *v1alpha1.Workspace) error {
	return nil
}

func ListDevopsProjects(req *restful.Request, resp *restful.Response) {

	workspace := req.PathParameter("workspace")
	username := req.PathParameter(constants.UserNameHeader)
	if username == "" {
		username = req.HeaderParameter(constants.UserNameHeader)
	}
	orderBy := req.QueryParameter(params.OrderByParam)
	reverse := params.ParseReverse(req)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))

	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	result, err := tenant.ListDevopsProjects(workspace, username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(result)
}

func DeleteDevopsProject(req *restful.Request, resp *restful.Response) {
	projectId := req.PathParameter("id")
	workspaceName := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	_, err := tenant.GetWorkspace(workspaceName)

	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	err = tenant.DeleteDevOpsProject(projectId, username)

	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(errors.None)
}

func CreateDevopsProject(req *restful.Request, resp *restful.Response) {

	workspaceName := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	var devops devops.DevOpsProject

	err := req.ReadEntity(&devops)

	if err != nil {
		glog.Infof("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	glog.Infoln("create workspace", username, workspaceName, devops)
	project, err := tenant.CreateDevopsProject(username, workspaceName, &devops)

	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
}

func ListNamespaceRules(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	username := req.HeaderParameter(constants.UserNameHeader)

	rules, err := iam.GetUserNamespaceSimpleRules(namespace, username)

	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteAsJson(rules)
}

func ListDevopsRules(req *restful.Request, resp *restful.Response) {

	devops := req.PathParameter("devops")
	username := req.HeaderParameter(constants.UserNameHeader)

	rules, err := tenant.GetUserDevopsSimpleRules(username, devops)

	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(rules)
}

func LogQuery(req *restful.Request, resp *restful.Response) {

	username := req.HeaderParameter(constants.UserNameHeader)

	// regenerate the request for log query
	newUrl := net.FormatURL("http", "127.0.0.1", 80, "/kapis/logging.kubesphere.io/v1alpha2/cluster")
	values := req.Request.URL.Query()

	clusterRules, err := iam.GetUserClusterRules(username)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		glog.Errorln(err)
		return
	}

	hasClusterLogAccess := iam.RulesMatchesRequired(clusterRules, rbacv1.PolicyRule{Verbs: []string{"get"}, Resources: []string{"*"}, APIGroups: []string{"logging.kubesphere.io"}})
	// if the user is not a cluster admin
	if !hasClusterLogAccess {
		queryNamespaces := strings.Split(req.QueryParameter("namespaces"), ",")
		// then the user can only view logs of namespaces he belongs to
		namespaces := make([]string, 0)
		roles, err := iam.GetUserRoles("", username)
		if err != nil {
			resp.WriteError(http.StatusInternalServerError, err)
			glog.Errorln(err)
		}
		for _, role := range roles {
			if !sliceutil.HasString(namespaces, role.Namespace) && iam.RulesMatchesRequired(role.Rules, rbacv1.PolicyRule{Verbs: []string{"get"}, Resources: []string{"*"}, APIGroups: []string{"logging.kubesphere.io"}}) {
				namespaces = append(namespaces, role.Namespace)
			}
		}

		// if the user belongs to no namespace
		// then no log visible
		if len(namespaces) == 0 {
			res := esclient.QueryResult{Status: http.StatusOK}
			resp.WriteAsJson(res)
			return
		} else if len(queryNamespaces) == 1 && queryNamespaces[0] == "" {
			values.Set("namespaces", strings.Join(namespaces, ","))
		} else {
			inter := intersection(queryNamespaces, namespaces)
			if len(inter) == 0 {
				res := esclient.QueryResult{Status: http.StatusOK}
				resp.WriteAsJson(res)
				return
			}
			values.Set("namespaces", strings.Join(inter, ","))
		}
	}

	newUrl.RawQuery = values.Encode()

	// forward the request to logging model
	newHttpRequest, _ := http.NewRequest(http.MethodGet, newUrl.String(), nil)
	logging.LoggingQueryCluster(restful.NewRequest(newHttpRequest), resp)
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
