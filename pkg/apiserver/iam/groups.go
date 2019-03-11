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

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
)

func CreateGroup(req *restful.Request, resp *restful.Response) {
	//var json map[string]interface{}

	var group models.Group

	err := req.ReadEntity(&group)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if !regexp.MustCompile("[a-z0-9]([-a-z0-9]*[a-z0-9])?").MatchString(group.Name) {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, fmt.Errorf("incalid group name %s", group))
		return
	}

	if group.Creator == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, fmt.Errorf("creator should not be null"))
		return
	}

	created, err := iam.CreateGroup(group)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(created)
}

func DeleteGroup(req *restful.Request, resp *restful.Response) {
	path := req.PathParameter("path")

	if path == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("group path must not be null")))
		return
	}

	err := iam.DeleteGroup(path)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)

}

func UpdateGroup(req *restful.Request, resp *restful.Response) {
	groupPathInPath := req.PathParameter("path")

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

func GroupDetail(req *restful.Request, resp *restful.Response) {

	path := req.PathParameter("path")

	conn, err := iam.NewConnection()

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	defer conn.Close()

	group, err := iam.GroupDetail(path, conn)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(group)

}

func GroupUsers(req *restful.Request, resp *restful.Response) {

	path := req.PathParameter("path")

	conn, err := iam.NewConnection()

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	defer conn.Close()

	group, err := iam.GroupDetail(path, conn)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	users := make([]*models.User, 0)

	modify := false

	for i := 0; i < len(group.Members); i++ {
		name := group.Members[i]
		user, err := iam.UserDetail(name, conn)

		if err != nil {
			if ldap.IsErrorWithCode(err, 32) {
				group.Members = append(group.Members[:i], group.Members[i+1:]...)
				i--
				modify = true
				continue
			} else {
				resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
				return
			}
		}

		clusterRoles, err := iam.GetClusterRoles(name)

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

		if group.Path == group.Name {
			workspaceRole := iam.GetWorkspaceRole(clusterRoles, group.Name)
			user.WorkspaceRole = workspaceRole
		}

		users = append(users, user)
	}

	if modify {
		go iam.UpdateGroup(group)
	}

	resp.WriteAsJson(users)

}

func CountHandler(req *restful.Request, resp *restful.Response) {
	count, err := iam.CountChild("")

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(map[string]int{"total_count": count})
}

func RootGroupList(req *restful.Request, resp *restful.Response) {

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

		conn, err := iam.NewConnection()

		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
			return
		}

		defer conn.Close()

		for _, v := range paths {
			path := strings.TrimSpace(v)
			group, err := iam.GroupDetail(path, conn)
			if err != nil {
				resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
				return
			}
			groups = append(groups, group)
		}

		resp.WriteAsJson(groups)
	}

}
