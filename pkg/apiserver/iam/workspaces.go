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
package iam

import (
	"fmt"
	"github.com/go-ldap/ldap"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/metrics"
	"kubesphere.io/kubesphere/pkg/models/workspaces"
	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	sliceutils "kubesphere.io/kubesphere/pkg/utils"
)

const UserNameHeader = "X-Token-Username"

func WorkspaceRoles(req *restful.Request, resp *restful.Response) {

	name := req.PathParameter("name")

	workspace, err := workspaces.Detail(name)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	roles, err := workspaces.Roles(workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(roles)
}

func WorkspaceMemberDetail(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("name")
	username := req.PathParameter("member")

	user, err := iam.GetUser(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	namespaces, err := workspaces.Namespaces(workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
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

func WorkspaceMemberInvite(req *restful.Request, resp *restful.Response) {
	var users []models.UserInvite
	workspace := req.PathParameter("name")
	err := req.ReadEntity(&users)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	err = workspaces.Invite(workspace, users)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func WorkspaceMemberRemove(req *restful.Request, resp *restful.Response) {
	query := req.QueryParameter("name")
	workspace := req.PathParameter("name")

	names := strings.Split(query, ",")

	err := workspaces.RemoveMembers(workspace, names)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func NamespaceCheckHandler(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	exist, err := workspaces.NamespaceExistCheck(namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(map[string]bool{"exist": exist})
}

func NamespaceDeleteHandler(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	workspace := req.PathParameter("name")

	err := workspaces.DeleteNamespace(workspace, namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
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

func DevOpsProjectCreateHandler(req *restful.Request, resp *restful.Response) {

	workspace := req.PathParameter("name")
	username := req.HeaderParameter(UserNameHeader)

	var devops models.DevopsProject

	err := req.ReadEntity(&devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	project, err := workspaces.CreateDevopsProject(username, workspace, devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
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
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
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
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid workspace name")))
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

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
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
	var workspace models.Workspace
	username := req.HeaderParameter(UserNameHeader)
	err := req.ReadEntity(&workspace)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	if workspace.Name == "" || strings.Contains(workspace.Name, ":") {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid workspace name")))
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

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(created)

}

func DeleteWorkspaceHandler(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")

	if name == "" || strings.Contains(name, ":") {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid workspace name")))
		return
	}

	workspace, err := workspaces.Detail(name)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	err = workspaces.Delete(workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}
func WorkspaceEditHandler(req *restful.Request, resp *restful.Response) {
	var workspace models.Workspace
	name := req.PathParameter("name")
	err := req.ReadEntity(&workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if name != workspace.Name {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("the name of workspace (%s) does not match the name on the URL (%s)", workspace.Name, name)))
		return
	}

	if workspace.Name == "" || strings.Contains(workspace.Name, ":") {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid workspace name")))
		return
	}

	workspace.Path = workspace.Name

	workspace.Members = nil

	edited, err := workspaces.Edit(&workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(edited)
}
func WorkspaceDetailHandler(req *restful.Request, resp *restful.Response) {

	name := req.PathParameter("name")

	workspace, err := workspaces.Detail(name)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(workspace)
}

// List all workspaces for the current user
func UserWorkspaceListHandler(req *restful.Request, resp *restful.Response) {
	keyword := req.QueryParameter("keyword")
	username := req.HeaderParameter(constants.UserNameHeader)

	ws, err := workspaces.ListWorkspaceByUser(username, keyword)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
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

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
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

func DevopsRulesHandler(req *restful.Request, resp *restful.Response) {
	//workspaceName := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)
	devopsName := req.PathParameter("devops")

	var rules []models.SimpleRule

	role, err := iam.GetDevopsRole(devopsName, username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	switch role {
	case "developer":
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"view", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	case "owner":
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"create", "edit", "view", "delete", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "credentials", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "devops", Actions: []string{"edit", "view", "delete"}},
		}
		break
	case "maintainer":
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"create", "edit", "view", "delete", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "credentials", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	case "reporter":
		fallthrough
	default:
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"view"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	}

	resp.WriteAsJson(rules)
}

func NamespacesRulesHandler(req *restful.Request, resp *restful.Response) {
	//workspaceName := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)
	namespaceName := req.PathParameter("namespace")

	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	roles, err := iam.GetRoles(username, namespaceName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	for _, clusterRole := range clusterRoles {
		role := new(rbac.Role)
		role.Name = clusterRole.Name
		role.Labels = clusterRole.Labels
		role.Namespace = namespaceName
		role.Annotations = clusterRole.Annotations
		role.Kind = "Role"
		role.Rules = clusterRole.Rules
		roles = append(roles, role)
	}

	rules, err := iam.GetRoleSimpleRules(roles, namespaceName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if rules[namespaceName] == nil {
		resp.WriteAsJson(make([]models.SimpleRule, 0))
	} else {
		resp.WriteAsJson(rules[namespaceName])
	}
}

func WorkspaceRulesHandler(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")

	username := req.HeaderParameter(constants.UserNameHeader)

	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	rules := iam.GetWorkspaceSimpleRules(clusterRoles, workspace)

	if rules[workspace] != nil {
		resp.WriteAsJson(rules[workspace])
	} else if rules["*"] != nil {
		resp.WriteAsJson(rules["*"])
	} else {
		resp.WriteAsJson(make([]models.SimpleRule, 0))
	}
}

func WorkspaceMemberList(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	limit, err := strconv.Atoi(req.QueryParameter("limit"))
	if err != nil {
		limit = 500
	}
	offset, err := strconv.Atoi(req.QueryParameter("offset"))
	if err != nil {
		offset = 0
	}

	conn, err := ldapclient.Client()

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	defer conn.Close()

	group, err := iam.GroupDetail(workspace, conn)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	keyword := ""

	if query := req.QueryParameter("keyword"); query != "" {
		keyword = query
	}

	users := make([]*models.User, 0)

	total := len(group.Members)

	members := sliceutils.RemoveString(group.Members, func(item string) bool {
		return keyword != "" && !strings.Contains(item, keyword)
	})

	for i := 0; i < len(members); i++ {
		username := members[i]

		if i < offset {
			continue
		}

		if len(users) == limit {
			break
		}

		user, err := iam.UserDetail(username, conn)

		if err != nil {
			if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
				group.Members = sliceutils.RemoveString(group.Members, func(item string) bool {
					return item == username
				})
				continue
			} else {
				resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
				return
			}
		}

		clusterRoles, err := iam.GetClusterRoles(username)

		for i := 0; i < len(clusterRoles); i++ {
			if clusterRoles[i].Annotations["rbac.authorization.k8s.io/clusterrole"] == "true" {
				user.ClusterRole = clusterRoles[i].Name
				break
			}
		}

		if group.Path == group.Name {

			workspaceRole := iam.GetWorkspaceRole(clusterRoles, group.Name)

			if err != nil {
				resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
				return
			}

			user.WorkspaceRole = workspaceRole
		}

		users = append(users, user)
	}

	if total > len(group.Members) {
		go iam.UpdateGroup(group)
	}
	if req.QueryParameter("limit") != "" {
		resp.WriteAsJson(map[string]interface{}{"items": users, "total_count": len(members)})
	} else {
		resp.WriteAsJson(users)
	}
}
