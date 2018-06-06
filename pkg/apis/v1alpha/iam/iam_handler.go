/*
 Copyright 2018 The KubeSphere Authors.

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
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
	"net/http"
	"strings"
	"kubesphere.io/kubesphere/pkg/constants"
	"k8s.io/api/rbac/v1"
	"kubesphere.io/kubesphere/pkg/util/slice"
)

func Register(ws *restful.WebService) {
	//roles
	ws.Route(ws.GET("/users/{username}/roles").To(userRolesHandler).Filter(route.RouteLogging)).Produces(restful.MIME_JSON)
	//rules define
	ws.Route(ws.GET("/roles/rules").To(roleRulesHandler).Filter(route.RouteLogging)).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/clusterroles/rules").To(clusterRoleRulesHandler).Filter(route.RouteLogging)).Produces(restful.MIME_JSON)
	//user->rules
	ws.Route(ws.GET("/rules").To(usersRulesHandler).Filter(route.RouteLogging)).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/users/{username}/rules").To(userRulesHandler).Filter(route.RouteLogging)).Produces(restful.MIME_JSON)
	//role->rules
	ws.Route(ws.GET("/clusterroles/{name}/rules").To(clusterRoleRulesHandler).Filter(route.RouteLogging)).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{name}/rules").To(roleRulesHandler).Filter(route.RouteLogging)).Produces(restful.MIME_JSON)
	//role->users
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{name}/users").To(roleUsersHandler).Filter(route.RouteLogging)).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/clusterroles/{name}/users").To(clusterRoleUsersHandler).Filter(route.RouteLogging)).Produces(restful.MIME_JSON)
}

// username -> roles
func userRolesHandler(req *restful.Request, resp *restful.Response) {

	username := req.PathParameter("username")

	roles, err := models.GetRoles(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	clusterRoles, err := models.GetClusterRoles(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	roleList := roleList{}
	roleList.Roles = roles
	roleList.ClusterRoles = clusterRoles

	resp.WriteEntity(roleList)
}

// namespaces + role name -> users
func roleUsersHandler(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")
	namespace := req.PathParameter("namespace")

	roleBindings, err := models.GetRoleBindings(namespace, name)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	users := make([]string, 0)

	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind &&
				!strings.HasPrefix(subject.Name, "system") &&
				!slice.ContainsString(users, subject.Name) {
				users = append(users, subject.Name)
			}
		}
	}

	resp.WriteEntity(users)
}

// cluster role name -> users
func clusterRoleUsersHandler(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")

	roleBindings, err := models.GetClusterRoleBindings(name)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	users := make([]string, 0)

	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && !strings.HasPrefix(subject.Name, "system") &&
				!slice.ContainsString(users, subject.Name) {
				users = append(users, subject.Name)
			}
		}
	}

	resp.WriteEntity(users)
}

// username -> rules
func usersRulesHandler(req *restful.Request, resp *restful.Response) {
	users := strings.Split(req.QueryParameter("users"), ",")

	usersRules := make(map[string]userRuleList, 0)

	for _, username := range users {
		_, contains := usersRules[username]
		if username != "" && !contains {

			userRuleList := userRuleList{}

			clusterRules, err := getUserClusterRules(username)

			if err != nil {
				resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
				return
			}

			rules, err := getUserRules(username)

			if err != nil {
				resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
				return
			}

			userRuleList.ClusterRules = clusterRules
			userRuleList.Rules = rules

			usersRules[username] = userRuleList
		}
	}

	resp.WriteEntity(usersRules)
}

// username -> rules
func userRulesHandler(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("username")

	userRuleList := userRuleList{}

	clusterRules, err := getUserClusterRules(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	rules, err := getUserRules(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
		return
	}

	userRuleList.ClusterRules = clusterRules
	userRuleList.Rules = rules

	resp.WriteEntity(userRuleList)
}

// cluster role name -> rules
func clusterRoleRulesHandler(req *restful.Request, resp *restful.Response) {

	name := req.PathParameter("name")

	var rules []rule

	if name == "" {
		rules = clusterRoleRuleGroup
	} else {
		var err error
		rules, err = getClusterRoleRules(name)
		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
			return
		}
	}

	resp.WriteEntity(rules)
}

// role name -> rules
func roleRulesHandler(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")
	namespace := req.PathParameter("namespace")

	var rules []rule

	if namespace == "" && name == "" {
		rules = roleRuleGroup
	} else {
		var err error
		rules, err = getRoleRules(namespace, name)
		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
			return
		}
	}
	resp.WriteEntity(rules)
}
