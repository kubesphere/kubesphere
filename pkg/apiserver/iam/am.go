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
	"github.com/emicklei/go-restful"
	"k8s.io/api/rbac/v1"
	"net/http"
	"sort"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
)

type roleList struct {
	ClusterRoles []*v1.ClusterRole `json:"clusterRoles" protobuf:"bytes,2,rep,name=clusterRoles"`
	Roles        []*v1.Role        `json:"roles" protobuf:"bytes,2,rep,name=roles"`
}

func RoleRules(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	roleName := req.PathParameter("role")

	role, err := iam.GetRole(namespace, roleName)

	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	rules, err := iam.GetRoleSimpleRules([]*v1.Role{role}, namespace)

	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteAsJson(rules[namespace])
}

func RoleUsers(req *restful.Request, resp *restful.Response) {
	roleName := req.PathParameter("role")
	namespace := req.PathParameter("namespace")

	users, err := iam.RoleUsers(namespace, roleName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(users)
}

func NamespaceUsers(req *restful.Request, resp *restful.Response) {

	namespace := req.PathParameter("namespace")

	users, err := iam.NamespaceUsers(namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Username < users[j].Username
	})

	resp.WriteAsJson(users)
}

func UserRoles(req *restful.Request, resp *restful.Response) {

	username := req.PathParameter("username")

	roles, err := iam.GetRoles(username, "")

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	roleList := roleList{}
	roleList.Roles = roles
	roleList.ClusterRoles = clusterRoles

	resp.WriteAsJson(roleList)
}

func NamespaceRulesHandler(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	username := req.HeaderParameter(constants.UserNameHeader)

	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	roles, err := iam.GetRoles(username, namespace)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	for _, clusterRole := range clusterRoles {
		role := new(v1.Role)
		role.Name = clusterRole.Name
		role.Labels = clusterRole.Labels
		role.Namespace = namespace
		role.Annotations = clusterRole.Annotations
		role.Kind = "Role"
		role.Rules = clusterRole.Rules
		roles = append(roles, role)
	}

	rules, err := iam.GetRoleSimpleRules(roles, namespace)

	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteAsJson(rules[namespace])
}

func RulesMappingHandler(req *restful.Request, resp *restful.Response) {
	rules := policy.RoleRuleMapping
	resp.WriteAsJson(rules)
}

func ClusterRulesMappingHandler(req *restful.Request, resp *restful.Response) {
	rules := policy.ClusterRoleRuleMapping
	resp.WriteAsJson(rules)
}

func ClusterRoleRules(req *restful.Request, resp *restful.Response) {
	clusterRoleName := req.PathParameter("clusterrole")
	clusterRole, err := iam.GetClusterRole(clusterRoleName)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	rules, err := iam.GetClusterRoleSimpleRules([]*v1.ClusterRole{clusterRole})
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteAsJson(rules)
}

func ClusterRoleUsers(req *restful.Request, resp *restful.Response) {
	clusterRoleName := req.PathParameter("clusterrole")

	users, err := iam.ClusterRoleUsers(clusterRoleName)

	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteAsJson(users)
}
