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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
	"net/mail"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/go-ldap/ldap"
	rbacv1 "k8s.io/api/rbac/v1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

func CreateUser(req *restful.Request, resp *restful.Response) {
	var user models.User

	err := req.ReadEntity(&user)

	if err != nil {
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if user.Username == "" {
		err = fmt.Errorf("invalid username: %s", user.Username)
		klog.Info(err, user.Username)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	// Parses a single RFC 5322 address, e.g. "Barry Gibbs <bg@example.com>"
	if _, err = mail.ParseAddress(user.Email); err != nil {
		err = fmt.Errorf("invalid email: %s", user.Email)
		klog.Info(err, user.Email)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if len(user.Password) < 6 {
		err = fmt.Errorf("invalid password")
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	created, err := iam.CreateUser(&user)

	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultEntryAlreadyExists) {
			klog.Info(err)
			resp.WriteHeaderAndEntity(http.StatusConflict, errors.Wrap(err))
			return
		}
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(created)
}

func DeleteUser(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")

	operator := req.HeaderParameter(constants.UserNameHeader)

	if operator == username {
		err := fmt.Errorf("cannot delete yourself")
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusForbidden, errors.Wrap(err))
		return
	}

	err := iam.DeleteUser(username)

	if err != nil {
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func UpdateUser(req *restful.Request, resp *restful.Response) {

	usernameInPath := req.PathParameter("user")
	usernameInHeader := req.HeaderParameter(constants.UserNameHeader)
	var user models.User

	err := req.ReadEntity(&user)

	if err != nil {
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if usernameInPath != user.Username {
		err = fmt.Errorf("the name of user (%s) does not match the name on the URL (%s)", user.Username, usernameInPath)
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if _, err = mail.ParseAddress(user.Email); err != nil {
		err = fmt.Errorf("invalid email: %s", user.Email)
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if user.Password != "" && len(user.Password) < 6 {
		err = fmt.Errorf("invalid password")
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	// change password by self
	if usernameInHeader == user.Username && user.Password != "" {
		isUserManager, err := isUserManager(usernameInHeader)
		if err != nil {
			klog.Error(err)
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}
		if !isUserManager {
			_, err = iam.Login(usernameInHeader, user.CurrentPassword, "")
		}
		if err != nil {
			err = fmt.Errorf("incorrect current password")
			klog.Info(err)
			resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
			return
		}
	}

	if usernameInHeader == user.Username {
		// change cluster role by self is not permitted
		user.ClusterRole = ""
	}

	result, err := iam.UpdateUser(&user)

	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultEntryAlreadyExists) {
			klog.Info(err)
			resp.WriteHeaderAndEntity(http.StatusConflict, errors.Wrap(err))
			return
		}

		klog.Error(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}

func isUserManager(username string) (bool, error) {
	rules, err := iam.GetUserClusterRules(username)
	if err != nil {
		return false, err
	}
	if iam.RulesMatchesRequired(rules, rbacv1.PolicyRule{Verbs: []string{"update"}, Resources: []string{"users"}, APIGroups: []string{"iam.kubesphere.io"}}) {
		return true, nil
	}
	return false, nil
}

func UserLoginLogs(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")
	logs, err := iam.LoginLog(username)

	if err != nil {
		klog.Error(err)
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

func DescribeUser(req *restful.Request, resp *restful.Response) {

	username := req.PathParameter("user")

	user, err := iam.DescribeUser(username)

	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
			klog.Info(err)
			resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		} else {
			klog.Error(err)
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		}
		return
	}

	clusterRole, err := iam.GetUserClusterRole(username)

	if err != nil {
		klog.Error(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	user.ClusterRole = clusterRole.Name

	clusterRules, err := iam.GetUserClusterSimpleRules(username)

	if err != nil {
		klog.Error(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	result := struct {
		*models.User
		ClusterRules []models.SimpleRule `json:"cluster_rules"`
	}{
		User:         user,
		ClusterRules: clusterRules,
	}

	resp.WriteAsJson(result)
}

func Precheck(req *restful.Request, resp *restful.Response) {

	check := req.QueryParameter("check")

	exist, err := iam.UserCreateCheck(check)

	if err != nil {
		klog.Error(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(map[string]bool{"exist": exist})
}

func ListUsers(req *restful.Request, resp *restful.Response) {

	if check := req.QueryParameter("check"); check != "" {
		Precheck(req, resp)
		return
	}

	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	reverse := params.ParseReverse(req)

	if err != nil {
		klog.Info(err)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	users, err := iam.ListUsers(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Error(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(users)
}
