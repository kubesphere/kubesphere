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
	"k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/tenant"
	"kubesphere.io/kubesphere/pkg/models/workspaces"
	"kubesphere.io/kubesphere/pkg/params"
	"log"
	"net/http"
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

	conditions.Match["kubesphere.io/workspace"] = workspace

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
	workspace := req.PathParameter("workspace")
	force := req.QueryParameter("force")
	username := req.HeaderParameter(constants.UserNameHeader)

	err := workspaces.UnBindDevopsProject(workspace, devops)

	if err != nil && force != "true" {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	err = workspaces.DeleteDevopsProject(username, devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func CreateDevopsProject(req *restful.Request, resp *restful.Response) {

	workspace := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	var devops models.DevopsProject

	err := req.ReadEntity(&devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	log.Println("create workspace", username, workspace, devops)
	project, err := workspaces.CreateDevopsProject(username, workspace, devops)

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
