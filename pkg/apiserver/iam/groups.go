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
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/go-ldap/ldap"

	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

func CreateGroup(req *restful.Request, resp *restful.Response) {
	var group models.Group

	err := req.ReadEntity(&group)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if !regexp.MustCompile("[a-z0-9]([-a-z0-9]*[a-z0-9])?").MatchString(group.Name) {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New(fmt.Sprintf("incalid group name %s", group)))
		return
	}

	created, err := iam.CreateGroup(&group)

	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultEntryAlreadyExists) {
			resp.WriteHeaderAndEntity(http.StatusConflict, errors.Wrap(err))
		} else {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		}
		return
	}

	resp.WriteAsJson(created)
}

func DeleteGroup(req *restful.Request, resp *restful.Response) {
	path := req.PathParameter("group")

	if path == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("group path must not be null")))
		return
	}

	err := iam.DeleteGroup(path)

	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
			resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
			return
		}
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)

}

func UpdateGroup(req *restful.Request, resp *restful.Response) {
	groupPathInPath := req.PathParameter("group")

	var group models.Group

	req.ReadEntity(&group)

	if groupPathInPath != group.Path {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("the path of group (%s) does not match the path on the URL (%s)", group.Path, groupPathInPath)))
		return
	}

	edited, err := iam.UpdateGroup(&group)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(edited)

}

func DescribeGroup(req *restful.Request, resp *restful.Response) {

	path := req.PathParameter("group")

	group, err := iam.DescribeGroup(path)

	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
			resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		} else {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		}
		return
	}

	resp.WriteAsJson(group)

}

func ListGroupUsers(req *restful.Request, resp *restful.Response) {

	path := req.PathParameter("group")

	group, err := iam.DescribeGroup(path)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	users := make([]*models.User, 0)

	modify := false

	for i := 0; i < len(group.Members); i++ {
		name := group.Members[i]
		user, err := iam.GetUserInfo(name)

		if err != nil {
			if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
				group.Members = append(group.Members[:i], group.Members[i+1:]...)
				i--
				modify = true
				continue
			} else {
				resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
				return
			}
		}

		users = append(users, user)
	}

	if modify {
		go iam.UpdateGroup(group)
	}

	resp.WriteAsJson(users)

}

func ListGroups(req *restful.Request, resp *restful.Response) {

	array := req.QueryParameter("path")

	if array == "" {
		groups, err := iam.ChildList("")

		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}

		resp.WriteAsJson(groups)
	} else {
		paths := strings.Split(array, ",")

		groups := make([]*models.Group, 0)

		for _, v := range paths {
			path := strings.TrimSpace(v)
			group, err := iam.DescribeGroup(path)
			if err != nil {
				resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
				return
			}
			groups = append(groups, group)
		}

		resp.WriteAsJson(groups)
	}

}
