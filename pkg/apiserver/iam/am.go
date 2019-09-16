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
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
	"sort"

	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

type RoleList struct {
	ClusterRoles []*v1.ClusterRole `json:"clusterRole" description:"cluster role list"`
	Roles        []*v1.Role        `json:"roles" description:"role list"`
}

func ListRoleUsers(req *restful.Request, resp *restful.Response) {
	roleName := req.PathParameter("role")
	namespace := req.PathParameter("namespace")

	users, err := iam.RoleUsers(namespace, roleName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(users)
}

func ListClusterRoles(req *restful.Request, resp *restful.Response) {
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := iam.ListClusterRoles(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)

}

func ListRoles(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := iam.ListRoles(namespace, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)

}

// List users by namespace
func ListNamespaceUsers(req *restful.Request, resp *restful.Response) {

	namespace := req.PathParameter("namespace")

	users, err := iam.NamespaceUsers(namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	// sort by time by default
	sort.Slice(users, func(i, j int) bool {
		return users[i].RoleBindTime.After(*users[j].RoleBindTime)
	})

	resp.WriteAsJson(users)
}

func ListUserRoles(req *restful.Request, resp *restful.Response) {

	username := req.PathParameter("user")

	roles, err := iam.GetUserRoles("", username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	_, clusterRoles, err := iam.GetUserClusterRoles(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	roleList := RoleList{}
	roleList.Roles = roles
	roleList.ClusterRoles = clusterRoles

	resp.WriteAsJson(roleList)
}

func RulesMapping(req *restful.Request, resp *restful.Response) {
	rules := policy.RoleRuleMapping
	resp.WriteAsJson(rules)
}

func ClusterRulesMapping(req *restful.Request, resp *restful.Response) {
	rules := policy.ClusterRoleRuleMapping
	resp.WriteAsJson(rules)
}

func ListClusterRoleRules(req *restful.Request, resp *restful.Response) {
	clusterRoleName := req.PathParameter("clusterrole")
	rules, err := iam.GetClusterRoleSimpleRules(clusterRoleName)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	resp.WriteAsJson(rules)
}

func ListClusterRoleUsers(req *restful.Request, resp *restful.Response) {
	clusterRoleName := req.PathParameter("clusterrole")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := iam.ListClusterRoleUsers(clusterRoleName, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		if k8serr.IsNotFound(err) {
			resp.WriteError(http.StatusNotFound, err)
		} else {
			resp.WriteError(http.StatusInternalServerError, err)
		}
		return
	}

	resp.WriteAsJson(result)
}

func ListRoleRules(req *restful.Request, resp *restful.Response) {
	namespaceName := req.PathParameter("namespace")
	roleName := req.PathParameter("role")

	rules, err := iam.GetRoleSimpleRules(namespaceName, roleName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(rules)
}
