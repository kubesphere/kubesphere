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
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/go-ldap/ldap"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
)

const (
	emailRegex = "^[a-z0-9]+([._\\-]*[a-z0-9])*@([a-z0-9]+[-a-z0-9]*[a-z0-9]+.){1,63}[a-z0-9]+$"
)

func CreateUser(req *restful.Request, resp *restful.Response) {
	var user models.User

	err := req.ReadEntity(&user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if user.Username == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid username")))
		return
	}

	if !regexp.MustCompile(emailRegex).MatchString(user.Email) {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid email")))
		return
	}

	if len(user.Password) < 6 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid password")))
		return
	}

	err = iam.CreateUser(user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func DeleteUser(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("name")

	operator := req.HeaderParameter(constants.UserNameHeader)

	if operator == username {
		resp.WriteHeaderAndEntity(http.StatusForbidden, errors.Wrap(fmt.Errorf("cannot delete yourself")))
		return
	}

	err := iam.DeleteUser(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func UpdateUser(req *restful.Request, resp *restful.Response) {

	usernameInPath := req.PathParameter("name")
	username := req.HeaderParameter(constants.UserNameHeader)
	var user models.User

	err := req.ReadEntity(&user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if usernameInPath != user.Username {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("the name of user (%s) does not match the name on the URL (%s)", user.Username, usernameInPath)))
		return
	}

	if !regexp.MustCompile(emailRegex).MatchString(user.Email) {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid email")))
		return
	}

	if user.Password != "" && len(user.Password) < 6 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("invalid password")))
		return
	}

	if username == user.Username && user.Password != "" {
		_, err = iam.Login(username, user.CurrentPassword, "")
		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("incorrect current password")))
			return
		}
	}

	err = iam.UpdateUser(user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func UserLoginLog(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("name")
	logs, err := iam.LoginLog(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	result := make([]map[string]string, 0)

	for _, v := range logs {
		item := strings.Split(v, ",")
		time := item[0]
		var ip string
		if len(item) > 1 {
			ip = item[1]
		}
		result = append(result, map[string]string{"login_time": time, "login_ip": ip})
	}

	resp.WriteAsJson(result)
}

func CurrentUserDetail(req *restful.Request, resp *restful.Response) {

	username := req.HeaderParameter(constants.UserNameHeader)

	conn, err := iam.NewConnection()

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	defer conn.Close()

	user, err := iam.UserDetail(username, conn)

	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
			resp.WriteHeaderAndEntity(http.StatusForbidden, errors.Wrap(err))
		} else {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		}
		return
	}

	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	clusterRules, err := iam.GetClusterRoleSimpleRules(clusterRoles)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	for i := 0; i < len(clusterRoles); i++ {
		if clusterRoles[i].Annotations["rbac.authorization.k8s.io/clusterrole"] == "true" {
			user.ClusterRole = clusterRoles[i].Name
			break
		}
	}

	user.ClusterRules = clusterRules

	resp.WriteAsJson(user)
}

func NamespacesListHandler(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("name")

	namespaces, err := iam.GetNamespaces(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(namespaces)
}

func UserDetail(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("name")

	conn, err := iam.NewConnection()

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	defer conn.Close()

	user, err := iam.UserDetail(username, conn)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	clusterRoles, err := iam.GetClusterRoles(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	clusterRules, err := iam.GetClusterRoleSimpleRules(clusterRoles)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	workspaceRoles := iam.GetWorkspaceRoles(clusterRoles)

	for i := 0; i < len(clusterRoles); i++ {
		if clusterRoles[i].Annotations["rbac.authorization.k8s.io/clusterrole"] == "true" {
			user.ClusterRole = clusterRoles[i].Name
			break
		}
	}

	user.ClusterRules = clusterRules

	user.WorkspaceRoles = workspaceRoles

	resp.WriteAsJson(user)
}

func UserList(req *restful.Request, resp *restful.Response) {

	limit, err := strconv.Atoi(req.QueryParameter("limit"))
	if err != nil {
		limit = 65535
	}
	offset, err := strconv.Atoi(req.QueryParameter("offset"))
	if err != nil {
		offset = 0
	}

	if check := req.QueryParameter("check"); check != "" {
		exist, err := iam.UserCreateCheck(check)
		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}

		resp.WriteAsJson(map[string]bool{"exist": exist})
		return
	}

	conn, err := iam.NewConnection()

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	defer conn.Close()

	if query := req.QueryParameter("name"); query != "" {
		names := strings.Split(query, ",")
		users := make([]*models.User, 0)
		for _, name := range names {
			user, err := iam.UserDetail(name, conn)
			if err != nil {
				if ldap.IsErrorWithCode(err, 32) {
					continue
				} else {
					resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
					return
				}
			}
			users = append(users, user)
		}

		resp.WriteAsJson(users)
		return
	}

	var total int
	var users []models.User

	if query := req.QueryParameter("search"); query != "" {
		total, users, err = iam.Search(query, limit, offset)
	} else if query := req.QueryParameter("keyword"); query != "" {
		total, users, err = iam.Search(query, limit, offset)
	} else {
		total, users, err = iam.UserList(limit, offset)
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	for i := 0; i < len(users); i++ {
		clusterRoles, err := iam.GetClusterRoles(users[i].Username)
		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}
		for j := 0; j < len(clusterRoles); j++ {
			if clusterRoles[j].Annotations["rbac.authorization.k8s.io/clusterrole"] == "true" {
				users[i].ClusterRole = clusterRoles[j].Name
				break
			}
		}
	}

	items := make([]interface{}, 0)

	for _, u := range users {
		items = append(items, u)
	}

	resp.WriteAsJson(models.PageableResponse{Items: items, TotalCount: total})
}
