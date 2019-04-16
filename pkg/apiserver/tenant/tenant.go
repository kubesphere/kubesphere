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
	"fmt"
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
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/models/tenant"
	"kubesphere.io/kubesphere/pkg/models/workspaces"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/elasticsearch"
	"kubesphere.io/kubesphere/pkg/simple/client/kubesphere"
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
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := tenant.ListDevopsProjects(workspace, username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}

func DeleteDevopsProject(req *restful.Request, resp *restful.Response) {
	devops := req.PathParameter("id")
	workspaceName := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	_, err := tenant.GetWorkspace(workspaceName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	err = kubesphere.Client().DeleteDevopsProject(username, devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	err = workspaces.UnBindDevopsProject(workspaceName, devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func CreateDevopsProject(req *restful.Request, resp *restful.Response) {

	workspaceName := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	var devops models.DevopsProject

	err := req.ReadEntity(&devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	glog.Infoln("create workspace", username, workspaceName, devops)
	project, err := workspaces.CreateDevopsProject(username, workspaceName, &devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
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

	rules, err := iam.GetUserDevopsSimpleRules(username, devops)

	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteAsJson(rules)
}

func LogQuery(req *restful.Request, resp *restful.Response) {

	username := req.HeaderParameter(constants.UserNameHeader)

	mapping, err := iam.GetUserWorkspaceRoleMap(username)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		glog.Errorln(err)
		return
	}

	workspaces := make([]string, 0)
	for workspaceName, role := range mapping {
		if role == fmt.Sprintf("workspace:%s:admin", workspaceName) {
			workspaces = append(workspaces, workspaceName)
		}
	}

	// regenerate the request for log query
	newUrl := net.FormatURL("http", "127.0.0.1", 80, "/kapis/logging.kubesphere.io/v1alpha2/cluster")
	values := req.Request.URL.Query()

	rules, err := iam.GetUserClusterRules(username)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		glog.Errorln(err)
		return
	}

	// if the user is not an admin
	if !iam.RulesMatchesRequired(rules, rbacv1.PolicyRule{Verbs: []string{"get"}, Resources: []string{"*"}, APIGroups: []string{"logging.kubesphere.io"}}) {
		// then the user can only view logs of workspaces he belongs to
		values.Set("workspaces", strings.Join(workspaces, ","))

		// if the user is not an admin, and belongs to no workspace
		// then no log visible
		if len(workspaces) == 0 {
			res := esclient.QueryResult{Status: http.StatusOK}
			resp.WriteAsJson(res)
			return
		}
	}

	newUrl.RawQuery = values.Encode()

	// forward the request to logging model
	newHttpRequest, _ := http.NewRequest(http.MethodGet, newUrl.String(), nil)
	logging.LoggingQueryCluster(restful.NewRequest(newHttpRequest), resp)
}
