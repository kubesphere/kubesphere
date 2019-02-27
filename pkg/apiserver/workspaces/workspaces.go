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
package workspaces

import (
	"net/http"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/metrics"

	"github.com/emicklei/go-restful"
	"k8s.io/api/core/v1"

	"fmt"
	"strings"

	"strconv"

	"regexp"

	"sort"

	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/workspaces"
)

const UserNameHeader = "X-Token-Username"

func V1Alpha2(ws *restful.WebService) {
	ws.Route(ws.GET("/workspaces").To(UserWorkspaceListHandler))
	ws.Route(ws.POST("/workspaces").To(WorkspaceCreateHandler))
	ws.Route(ws.DELETE("/workspaces/{name}").To(DeleteWorkspaceHandler))
	ws.Route(ws.GET("/workspaces/{name}").To(WorkspaceDetailHandler))
	ws.Route(ws.PUT("/workspaces/{name}").To(WorkspaceEditHandler))
	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").To(UserNamespaceListHandler))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{username}/namespaces").To(UserNamespaceListHandler))
	ws.Route(ws.POST("/workspaces/{name}/namespaces").To(NamespaceCreateHandler))
	ws.Route(ws.DELETE("/workspaces/{name}/namespaces/{namespace}").To(NamespaceDeleteHandler))
	ws.Route(ws.GET("/workspaces/{name}/namespaces/{namespace}").To(NamespaceCheckHandler))
	ws.Route(ws.GET("/namespaces/{namespace}").To(NamespaceCheckHandler))
	ws.Route(ws.GET("/workspaces/{name}/devops").To(DevOpsProjectHandler))
	ws.Route(ws.GET("/workspaces/{name}/members/{username}/devops").To(DevOpsProjectHandler))
	ws.Route(ws.POST("/workspaces/{name}/devops").To(DevOpsProjectCreateHandler))
	ws.Route(ws.DELETE("/workspaces/{name}/devops/{id}").To(DevOpsProjectDeleteHandler))

	ws.Route(ws.GET("/workspaces/{name}/members").To(MembersHandler))
	ws.Route(ws.GET("/workspaces/{name}/members/{member}").To(MemberHandler))
	ws.Route(ws.GET("/workspaces/{name}/roles").To(RolesHandler))
	// TODO /workspaces/{name}/roles/{role}
	ws.Route(ws.POST("/workspaces/{name}/members").To(MembersInviteHandler))
	ws.Route(ws.DELETE("/workspaces/{name}/members").To(MembersRemoveHandler))
}

func RolesHandler(req *restful.Request, resp *restful.Response) {

	name := req.PathParameter("name")

	workspace, err := workspaces.Detail(name)

	if errors.HandlerError(err, resp) {
		return
	}

	roles, err := workspaces.Roles(workspace)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(roles)
}

func MembersHandler(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("name")
	keyword := req.QueryParameter("keyword")

	users, err := workspaces.GetWorkspaceMembers(workspace, keyword)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(users)
}

func MemberHandler(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("name")
	username := req.PathParameter("member")

	user, err := iam.GetUser(username)
	if errors.HandlerError(err, resp) {
		return
	}

	namespaces, err := workspaces.Namespaces(workspace)

	if errors.HandlerError(err, resp) {
		return
	}

	user.WorkspaceRole = user.WorkspaceRoles[workspace]

	roles := make(map[string]string)

	for _, namespace := range namespaces {
		if role := user.Roles[namespace.Name]; role != "" {
			roles[namespace.Name] = role
		}
	}

	user.Roles = roles
	user.Rules = nil
	user.WorkspaceRules = nil
	user.WorkspaceRoles = nil
	user.ClusterRules = nil
	resp.WriteAsJson(user)
}

func MembersInviteHandler(req *restful.Request, resp *restful.Response) {
	var users []workspaces.UserInvite
	workspace := req.PathParameter("name")
	err := req.ReadEntity(&users)

	if errors.HandlerError(err, resp) {
		return
	}

	err = workspaces.Invite(workspace, users)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(errors.None)
}

func MembersRemoveHandler(req *restful.Request, resp *restful.Response) {
	query := req.QueryParameter("name")
	workspace := req.PathParameter("name")

	names := strings.Split(query, ",")

	err := workspaces.RemoveMembers(workspace, names)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(errors.None)
}

func NamespaceCheckHandler(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	exist, err := workspaces.NamespaceExistCheck(namespace)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(map[string]bool{"exist": exist})
}

func NamespaceDeleteHandler(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	workspace := req.PathParameter("name")

	err := workspaces.DeleteNamespace(workspace, namespace)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(errors.None)
}

func DevOpsProjectDeleteHandler(req *restful.Request, resp *restful.Response) {
	devops := req.PathParameter("id")
	workspace := req.PathParameter("name")
	force := req.QueryParameter("force")
	username := req.HeaderParameter(UserNameHeader)

	err := workspaces.UnBindDevopsProject(workspace, devops)

	if err != nil && force != "true" {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.New(errors.Internal, err.Error()))
		return
	}

	err = workspaces.DeleteDevopsProject(username, devops)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(errors.None)
}

func DevOpsProjectCreateHandler(req *restful.Request, resp *restful.Response) {

	workspace := req.PathParameter("name")
	username := req.HeaderParameter(UserNameHeader)

	var devops workspaces.DevopsProject

	err := req.ReadEntity(&devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, err.Error()))
		return
	}

	project, err := workspaces.CreateDevopsProject(username, workspace, devops)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(project)

}

func NamespaceCreateHandler(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("name")
	username := req.HeaderParameter(UserNameHeader)

	namespace := &v1.Namespace{}

	err := req.ReadEntity(namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, err.Error()))
		return
	}

	if namespace.Annotations == nil {
		namespace.Annotations = make(map[string]string, 0)
	}

	namespace.Annotations["creator"] = username
	namespace.Annotations["workspace"] = workspace

	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string, 0)
	}

	namespace.Labels["kubesphere.io/workspace"] = workspace

	namespace, err = workspaces.CreateNamespace(namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, err.Error()))
		return
	}

	resp.WriteAsJson(namespace)
}

func DevOpsProjectHandler(req *restful.Request, resp *restful.Response) {

	workspace := req.PathParameter("name")
	username := req.PathParameter("username")
	keyword := req.QueryParameter("keyword")

	if username == "" {
		username = req.HeaderParameter(UserNameHeader)
	}

	limit := 65535
	offset := 0
	orderBy := "createTime"
	reverse := true

	if groups := regexp.MustCompile(`^limit=(\d+),page=(\d+)$`).FindStringSubmatch(req.QueryParameter("paging")); len(groups) == 3 {
		limit, _ = strconv.Atoi(groups[1])
		page, _ := strconv.Atoi(groups[2])
		offset = (page - 1) * limit
	}

	if groups := regexp.MustCompile(`^(createTime|name)$`).FindStringSubmatch(req.QueryParameter("order")); len(groups) == 2 {
		orderBy = groups[1]
		reverse = false
	}

	if q := req.QueryParameter("reverse"); q != "" {
		b, err := strconv.ParseBool(q)
		if err == nil {
			reverse = b
		}
	}

	total, devOpsProjects, err := workspaces.ListDevopsProjectsByUser(username, workspace, keyword, orderBy, reverse, limit, offset)

	if errors.HandlerError(err, resp) {
		return
	}

	result := models.PageableResponse{}
	result.TotalCount = total
	result.Items = make([]interface{}, 0)
	for _, n := range devOpsProjects {
		result.Items = append(result.Items, n)
	}
	resp.WriteAsJson(result)
}

func WorkspaceCreateHandler(req *restful.Request, resp *restful.Response) {
	var workspace workspaces.Workspace
	username := req.HeaderParameter(UserNameHeader)
	err := req.ReadEntity(&workspace)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, err.Error()))
		return
	}
	if workspace.Name == "" || strings.Contains(workspace.Name, ":") {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, "invalid workspace name"))
		return
	}

	workspace.Path = workspace.Name
	workspace.Members = nil

	if workspace.Admin != "" {
		workspace.Creator = workspace.Admin
	} else {
		workspace.Creator = username
	}

	created, err := workspaces.Create(&workspace)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(created)

}

func DeleteWorkspaceHandler(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")

	if name == "" || strings.Contains(name, ":") {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, "invalid workspace name"))
		return
	}

	workspace, err := workspaces.Detail(name)

	if errors.HandlerError(err, resp) {
		return
	}

	err = workspaces.Delete(workspace)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(errors.None)
}
func WorkspaceEditHandler(req *restful.Request, resp *restful.Response) {
	var workspace workspaces.Workspace
	name := req.PathParameter("name")
	err := req.ReadEntity(&workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, err.Error()))
		return
	}

	if name != workspace.Name {
		resp.WriteError(http.StatusBadRequest, fmt.Errorf("the name of workspace (%s) does not match the name on the URL (%s)", workspace.Name, name))
		return
	}

	if workspace.Name == "" || strings.Contains(workspace.Name, ":") {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(errors.InvalidArgument, "invalid workspace name"))
		return
	}

	workspace.Path = workspace.Name

	workspace.Members = nil

	edited, err := workspaces.Edit(&workspace)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(edited)
}
func WorkspaceDetailHandler(req *restful.Request, resp *restful.Response) {

	name := req.PathParameter("name")

	workspace, err := workspaces.Detail(name)

	if errors.HandlerError(err, resp) {
		return
	}

	resp.WriteAsJson(workspace)
}

// List all workspaces for the current user
func UserWorkspaceListHandler(req *restful.Request, resp *restful.Response) {
	keyword := req.QueryParameter("keyword")
	username := req.HeaderParameter(UserNameHeader)

	ws, err := workspaces.ListWorkspaceByUser(username, keyword)

	if errors.HandlerError(err, resp) {
		return
	}

	sort.Slice(ws, func(i, j int) bool {
		t1, err := ws[i].GetCreateTime()
		if err != nil {
			return false
		}
		t2, err := ws[j].GetCreateTime()
		if err != nil {
			return true
		}
		return t1.After(t2)
	})

	resp.WriteAsJson(ws)
}

func UserNamespaceListHandler(req *restful.Request, resp *restful.Response) {
	withMetrics, err := strconv.ParseBool(req.QueryParameter("metrics"))

	if err != nil {
		withMetrics = false
	}

	username := req.PathParameter("username")
	keyword := req.QueryParameter("keyword")
	if username == "" {
		username = req.HeaderParameter(UserNameHeader)
	}
	limit := 65535
	offset := 0
	orderBy := "createTime"
	reverse := true

	if groups := regexp.MustCompile(`^limit=(\d+),page=(\d+)$`).FindStringSubmatch(req.QueryParameter("paging")); len(groups) == 3 {
		limit, _ = strconv.Atoi(groups[1])
		page, _ := strconv.Atoi(groups[2])
		if page < 0 {
			page = 1
		}
		offset = (page - 1) * limit
	}

	if groups := regexp.MustCompile(`^(createTime|name)$`).FindStringSubmatch(req.QueryParameter("order")); len(groups) == 2 {
		orderBy = groups[1]
		reverse = false
	}

	if q := req.QueryParameter("reverse"); q != "" {
		b, err := strconv.ParseBool(q)
		if err == nil {
			reverse = b
		}
	}

	workspaceName := req.PathParameter("workspace")

	total, namespaces, err := workspaces.ListNamespaceByUser(workspaceName, username, keyword, orderBy, reverse, limit, offset)

	if withMetrics {
		namespaces = metrics.GetNamespacesWithMetrics(namespaces)
	}

	if errors.HandlerError(err, resp) {
		return
	}

	result := models.PageableResponse{}
	result.TotalCount = total
	result.Items = make([]interface{}, 0)
	for _, n := range namespaces {
		result.Items = append(result.Items, n)
	}

	resp.WriteAsJson(result)
}
