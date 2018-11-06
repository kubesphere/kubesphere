package workspaces

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/api/core/v1"

	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/util/slice"

	"strconv"

	"regexp"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/metrics"
	"kubesphere.io/kubesphere/pkg/models/workspaces"
)

const UserNameHeader = "X-Token-Username"

func Register(ws *restful.WebService, subPath string) {

	ws.Route(ws.GET(subPath).To(UserWorkspaceListHandler))
	ws.Route(ws.POST(subPath).To(WorkspaceCreateHandler))
	ws.Route(ws.DELETE(subPath + "/{name}").To(DeleteWorkspaceHandler))
	ws.Route(ws.GET(subPath + "/{name}").To(WorkspaceDetailHandler))
	ws.Route(ws.PUT(subPath + "/{name}").To(WorkspaceEditHandler))
	ws.Route(ws.GET(subPath + "/{workspace}/namespaces").To(UserNamespaceListHandler))
	ws.Route(ws.GET(subPath + "/{workspace}/members/{username}/namespaces").To(UserNamespaceListHandler))
	ws.Route(ws.POST(subPath + "/{name}/namespaces").To(NamespaceCreateHandler))
	ws.Route(ws.DELETE(subPath + "/{name}/namespaces/{namespace}").To(NamespaceDeleteHandler))
	ws.Route(ws.GET(subPath + "/{name}/namespaces/{namespace}").To(NamespaceCheckHandler))
	ws.Route(ws.GET("/namespaces/{namespace}").To(NamespaceCheckHandler))
	ws.Route(ws.GET(subPath + "/{name}/devops").To(DevOpsProjectHandler))
	ws.Route(ws.GET(subPath + "/{name}/members/{username}/devops").To(DevOpsProjectHandler))
	ws.Route(ws.POST(subPath + "/{name}/devops").To(DevOpsProjectCreateHandler))
	ws.Route(ws.DELETE(subPath + "/{name}/devops/{id}").To(DevOpsProjectDeleteHandler))

	ws.Route(ws.GET(subPath + "/{name}/members").To(MembersHandler))
	ws.Route(ws.GET(subPath + "/{name}/members/{member}").To(MemberHandler))
	ws.Route(ws.GET(subPath + "/{name}/roles").To(RolesHandler))
	ws.Route(ws.GET(subPath + "/{name}/roles/{role}").To(RoleHandler))
	ws.Route(ws.POST(subPath + "/{name}/members").To(MembersInviteHandler))
	ws.Route(ws.DELETE(subPath + "/{name}/members").To(MembersRemoveHandler))
}

func RoleHandler(req *restful.Request, resp *restful.Response) {
	workspaceName := req.PathParameter("name")
	roleName := req.PathParameter("role")

	if !slice.ContainsString(workspaces.WorkSpaceRoles, roleName, nil) {
		resp.WriteHeaderAndEntity(http.StatusNotFound, constants.MessageResponse{Message: fmt.Sprintf("role %s not found", roleName)})
		return
	}

	role, rules, err := iam.WorkspaceRoleRules(workspaceName, roleName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	users, err := iam.WorkspaceRoleUsers(workspaceName, roleName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(map[string]interface{}{"role": role, "rules": rules, "users": users})
}

func RolesHandler(req *restful.Request, resp *restful.Response) {

	name := req.PathParameter("name")

	workspace, err := workspaces.Detail(name)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	roles, err := workspaces.Roles(workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(roles)
}

func MembersHandler(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("name")

	users, err := workspaces.GetWorkspaceMembers(workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(users)
}

func MemberHandler(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("name")
	username := req.PathParameter("member")

	user, err := iam.GetUser(username)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	namespaces, err := workspaces.Namespaces(workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
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
	resp.WriteEntity(user)
}

func MembersInviteHandler(req *restful.Request, resp *restful.Response) {
	var users []workspaces.UserInvite
	workspace := req.PathParameter("name")
	err := req.ReadEntity(&users)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	err = workspaces.Invite(workspace, users)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteHeaderAndEntity(http.StatusOK, constants.MessageResponse{Message: "success"})
}

func MembersRemoveHandler(req *restful.Request, resp *restful.Response) {
	query := req.QueryParameter("name")
	workspace := req.PathParameter("name")

	names := strings.Split(query, ",")

	err := workspaces.RemoveMembers(workspace, names)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteHeaderAndEntity(http.StatusOK, constants.MessageResponse{Message: "success"})
}

func NamespaceCheckHandler(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	exist, err := workspaces.NamespaceExistCheck(namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(map[string]bool{"exist": exist})
}

func NamespaceDeleteHandler(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	workspace := req.PathParameter("name")

	err := workspaces.DeleteNamespace(workspace, namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteHeaderAndEntity(http.StatusOK, constants.MessageResponse{Message: "success"})
}

func DevOpsProjectDeleteHandler(req *restful.Request, resp *restful.Response) {
	devops := req.PathParameter("id")
	workspace := req.PathParameter("name")
	force := req.QueryParameter("force")
	username := req.HeaderParameter(UserNameHeader)

	err := workspaces.UnBindDevopsProject(workspace, devops)

	if err != nil && force != "true" {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	err = workspaces.DeleteDevopsProject(username, devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(constants.MessageResponse{Message: "success"})
}

func DevOpsProjectCreateHandler(req *restful.Request, resp *restful.Response) {

	workspace := req.PathParameter("name")
	username := req.HeaderParameter(UserNameHeader)

	var devops workspaces.DevopsProject

	err := req.ReadEntity(&devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, constants.MessageResponse{Message: err.Error()})
		return
	}

	project, err := workspaces.CreateDevopsProject(username, workspace, devops)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(project)

}

func NamespaceCreateHandler(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("name")
	username := req.HeaderParameter(UserNameHeader)

	namespace := &v1.Namespace{}

	err := req.ReadEntity(namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, constants.MessageResponse{Message: err.Error()})
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
		resp.WriteHeaderAndEntity(http.StatusBadRequest, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(namespace)
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

	total, devOpsProjects, err := workspaces.ListDevopsProjectsByUser(username, workspace, keyword, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	result := constants.PageableResponse{}
	result.TotalCount = total
	result.Items = make([]interface{}, 0)
	for _, n := range devOpsProjects {
		result.Items = append(result.Items, n)
	}
	resp.WriteEntity(result)
}

func WorkspaceCreateHandler(req *restful.Request, resp *restful.Response) {
	var workspace workspaces.Workspace
	username := req.HeaderParameter(UserNameHeader)
	err := req.ReadEntity(&workspace)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, constants.MessageResponse{Message: err.Error()})
		return
	}
	if workspace.Name == "" || strings.Contains(workspace.Name, ":") {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, constants.MessageResponse{Message: "invalid workspace name"})
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
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(created)

}

func DeleteWorkspaceHandler(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")

	if name == "" || strings.Contains(name, ":") {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, constants.MessageResponse{Message: "invalid workspace name"})
		return
	}

	workspace, err := workspaces.Detail(name)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	err = workspaces.Delete(workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(constants.MessageResponse{Message: "success"})
}
func WorkspaceEditHandler(req *restful.Request, resp *restful.Response) {
	var workspace workspaces.Workspace
	name := req.PathParameter("name")
	err := req.ReadEntity(&workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, constants.MessageResponse{Message: err.Error()})
		return
	}

	if name != workspace.Name {
		resp.WriteError(http.StatusBadRequest, fmt.Errorf("the name of workspace (%s) does not match the name on the URL (%s)", workspace.Name, name))
		return
	}

	if workspace.Name == "" || strings.Contains(workspace.Name, ":") {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, constants.MessageResponse{Message: "invalid workspace name"})
		return
	}

	workspace.Path = workspace.Name

	workspace.Members = nil

	edited, err := workspaces.Edit(&workspace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(edited)
}
func WorkspaceDetailHandler(req *restful.Request, resp *restful.Response) {

	name := req.PathParameter("name")

	workspace, err := workspaces.Detail(name)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(workspace)
}

// List all workspaces for the current user
func UserWorkspaceListHandler(req *restful.Request, resp *restful.Response) {
	keyword := req.QueryParameter("keyword")
	username := req.HeaderParameter(UserNameHeader)

	list, err := workspaces.ListWorkspaceByUser(username, keyword)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	resp.WriteEntity(list)
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
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	result := constants.PageableResponse{}
	result.TotalCount = total
	result.Items = make([]interface{}, 0)
	for _, n := range namespaces {
		result.Items = append(result.Items, n)
	}

	resp.WriteEntity(result)
}
