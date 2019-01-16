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
package groups

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/go-ldap/ldap"

	"kubesphere.io/kubesphere/pkg/errors"
	. "kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
)

func CreateHandler(req *restful.Request, resp *restful.Response) {
	//var json map[string]interface{}

	var group Group

	err := req.ReadEntity(&group)

	if err != nil {
		errors.HandleError(errors.New(errors.InvalidArgument, err.Error()), resp)
		return
	}

	if !regexp.MustCompile("[a-z0-9]([-a-z0-9]*[a-z0-9])?").MatchString(group.Name) {
		errors.HandleError(errors.New(errors.InvalidArgument, "invalid group name"), resp)
		return
	}

	if group.Creator == "" {
		errors.HandleError(errors.New(errors.InvalidArgument, "creator should not be null"), resp)
		return
	}

	created, err := iam.CreateGroup(group)

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteAsJson(created)
}

func DeleteHandler(req *restful.Request, resp *restful.Response) {
	path := req.PathParameter("path")

	if path == "" {
		resp.WriteError(http.StatusInternalServerError, fmt.Errorf("group path must not be null"))
		return
	}

	err := iam.DeleteGroup(path)
	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteAsJson(errors.None)

}

func EditHandler(req *restful.Request, resp *restful.Response) {
	groupPathInPath := req.PathParameter("path")

	var group Group

	req.ReadEntity(&group)

	if groupPathInPath != group.Path {
		errors.HandleError(errors.New(errors.InvalidArgument, fmt.Sprintf("the path of group (%s) does not match the path on the URL (%s)", group.Path, groupPathInPath)), resp)
		return
	}

	edited, err := iam.UpdateGroup(&group)

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteAsJson(edited)

}

func DetailHandler(req *restful.Request, resp *restful.Response) {

	path := req.PathParameter("path")

	conn, err := iam.NewConnection()

	if errors.HandleError(err, resp) {
		return
	}

	defer conn.Close()

	group, err := iam.GroupDetail(path, conn)

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteAsJson(group)

}

func UsersHandler(req *restful.Request, resp *restful.Response) {

	path := req.PathParameter("path")

	conn, err := iam.NewConnection()

	if errors.HandleError(err, resp) {
		return
	}

	defer conn.Close()

	group, err := iam.GroupDetail(path, conn)

	if errors.HandleError(err, resp) {
		return
	}

	users := make([]*User, 0)

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
				errors.HandleError(err, resp)
				return
			}
		}

		clusterRoles, err := iam.GetClusterRoles(name)

		for i := 0; i < len(clusterRoles); i++ {
			if clusterRoles[i].Annotations["rbac.authorization.k8s.io/clusterrole"] == "true" {
				user.ClusterRole = clusterRoles[i].Name
				break
			}
		}

		if group.Path == group.Name {

			workspaceRole := iam.GetWorkspaceRole(clusterRoles, group.Name)

			if errors.HandleError(err, resp) {
				return
			}

			user.WorkspaceRole = workspaceRole
		}

		users = append(users, user)
	}

	if modify {
		go iam.UpdateGroup(group)
	}

	resp.WriteEntity(users)

}

func CountHandler(req *restful.Request, resp *restful.Response) {
	count, err := iam.CountChild("")

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteEntity(map[string]int{"total_count": count})
}

func RootListHandler(req *restful.Request, resp *restful.Response) {

	array := req.QueryParameter("path")

	if array == "" {
		groups, err := iam.ChildList("")

		if errors.HandleError(err, resp) {
			return
		}

		resp.WriteAsJson(groups)
	} else {
		paths := strings.Split(array, ",")

		groups := make([]*Group, 0)

		conn, err := iam.NewConnection()

		if errors.HandleError(err, resp) {
			return
		}

		defer conn.Close()

		for _, v := range paths {
			path := strings.TrimSpace(v)
			group, err := iam.GroupDetail(path, conn)
			if errors.HandleError(err, resp) {
				return
			}
			groups = append(groups, group)
		}

		resp.WriteAsJson(groups)
	}

}
