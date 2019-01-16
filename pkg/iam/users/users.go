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
package users

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
	. "kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
)

const (
	emailRegex = "^[a-z0-9]+([._\\-]*[a-z0-9])*@([a-z0-9]+[-a-z0-9]*[a-z0-9]+.){1,63}[a-z0-9]+$"
)

func CreateHandler(req *restful.Request, resp *restful.Response) {
	var user User

	err := req.ReadEntity(&user)

	if err != nil {
		errors.HandleError(errors.New(errors.InvalidArgument, err.Error()), resp)
		return
	}

	if user.Username == "" {
		errors.HandleError(errors.New(errors.InvalidArgument, "invalid username"), resp)
		return
	}

	if !regexp.MustCompile(emailRegex).MatchString(user.Email) {
		errors.HandleError(errors.New(errors.InvalidArgument, "invalid email"), resp)
		return
	}

	if len(user.Password) < 6 {
		errors.HandleError(errors.New(errors.InvalidArgument, "invalid password"), resp)
		return
	}

	err = iam.CreateUser(user)

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteAsJson(errors.None)
}

func DeleteHandler(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("name")

	operator := req.HeaderParameter(constants.UserNameHeader)

	if operator == username {
		errors.HandleError(errors.New(errors.Forbidden, "cannot delete yourself"), resp)
		return
	}

	err := iam.DeleteUser(username)

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteAsJson(errors.None)
}

func EditHandler(req *restful.Request, resp *restful.Response) {

	usernameInPath := req.PathParameter("name")
	username := req.HeaderParameter(constants.UserNameHeader)
	var user User

	err := req.ReadEntity(&user)

	if err != nil {
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	if usernameInPath != user.Username {
		errors.HandleError(errors.New(errors.InvalidArgument, fmt.Sprintf("the name of user (%s) does not match the name on the URL (%s)", user.Username, usernameInPath)), resp)
		return
	}

	if !regexp.MustCompile(emailRegex).MatchString(user.Email) {
		errors.HandleError(errors.New(errors.InvalidArgument, "invalid email"), resp)
		return
	}

	if user.Password != "" && len(user.Password) < 6 {
		errors.HandleError(errors.New(errors.InvalidArgument, "invalid password"), resp)
		return
	}

	if username == user.Username && user.Password != "" {
		_, err = iam.Login(username, user.CurrentPassword, "")
		if err != nil {
			errors.HandleError(errors.New(errors.InvalidArgument, "incorrect current password"), resp)
			return
		}
	}

	err = iam.UpdateUser(user)

	if errors.HandleError(err, resp) {
		return
	}
	resp.WriteAsJson(errors.None)
}

func LogHandler(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("name")
	logs, err := iam.LoginLog(username)

	if errors.HandleError(err, resp) {
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

	resp.WriteEntity(result)
}

func CurrentUserHandler(req *restful.Request, resp *restful.Response) {

	username := req.HeaderParameter(constants.UserNameHeader)

	conn, err := iam.NewConnection()

	if errors.HandleError(err, resp) {
		return
	}

	defer conn.Close()

	user, err := iam.UserDetail(username, conn)

	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
			errors.HandleError(errors.New(errors.Forbidden, err.Error()), resp)
		} else {
			errors.HandleError(err, resp)
		}
		return
	}

	clusterRoles, err := iam.GetClusterRoles(username)

	if errors.HandleError(err, resp) {
		return
	}

	clusterRules, err := iam.GetClusterRoleSimpleRules(clusterRoles)

	if errors.HandleError(err, resp) {
		return
	}

	for i := 0; i < len(clusterRoles); i++ {
		if clusterRoles[i].Annotations["rbac.authorization.k8s.io/clusterrole"] == "true" {
			user.ClusterRole = clusterRoles[i].Name
			break
		}
	}

	user.ClusterRules = clusterRules

	//user.Roles = roleMapping
	//user.Rules = rules

	//user.WorkspaceRoles = workspaceRoles
	//user.WorkspaceRules = workspaceRules

	resp.WriteEntity(user)
}

func NamespacesListHandler(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("name")

	namespaces, err := iam.GetNamespaces(username)

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteEntity(namespaces)
}

func DetailHandler(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("name")

	conn, err := iam.NewConnection()

	if errors.HandleError(err, resp) {
		return
	}

	defer conn.Close()

	user, err := iam.UserDetail(username, conn)

	if errors.HandleError(err, resp) {
		return
	}

	clusterRoles, err := iam.GetClusterRoles(username)

	if errors.HandleError(err, resp) {
		return
	}

	clusterRules, err := iam.GetClusterRoleSimpleRules(clusterRoles)

	if errors.HandleError(err, resp) {
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

	resp.WriteEntity(user)
}

func ListHandler(req *restful.Request, resp *restful.Response) {

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
		if errors.HandleError(err, resp) {
			return
		}

		resp.WriteEntity(map[string]bool{"exist": exist})
		return
	}

	conn, err := iam.NewConnection()

	if errors.HandleError(err, resp) {
		return
	}
	defer conn.Close()

	if query := req.QueryParameter("name"); query != "" {
		names := strings.Split(query, ",")
		users := make([]*User, 0)
		for _, name := range names {
			user, err := iam.UserDetail(name, conn)
			if err != nil {
				if ldap.IsErrorWithCode(err, 32) {
					continue
				} else {
					errors.HandleError(err, resp)
					return
				}
			}
			users = append(users, user)
		}

		resp.WriteEntity(users)
		return
	}

	var total int
	var users []User

	if query := req.QueryParameter("search"); query != "" {
		total, users, err = iam.Search(query, limit, offset)
	} else if query := req.QueryParameter("keyword"); query != "" {
		total, users, err = iam.Search(query, limit, offset)
	} else {
		total, users, err = iam.UserList(limit, offset)
	}

	if errors.HandleError(err, resp) {
		return
	}

	for i := 0; i < len(users); i++ {
		clusterRoles, err := iam.GetClusterRoles(users[i].Username)
		if errors.HandleError(err, resp) {
			return
		}
		for j := 0; j < len(clusterRoles); j++ {
			if clusterRoles[j].Annotations["rbac.authorization.k8s.io/clusterrole"] == "true" {
				users[i].ClusterRole = clusterRoles[j].Name
				break
			}
		}
	}

	resp.WriteEntity(map[string]interface{}{"items": users, "total_count": total})
}
