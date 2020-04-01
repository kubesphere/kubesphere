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
package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/api"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ldappool "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"net/http"
)

const groupName = "iam.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: groupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, k8sClient k8s.Client, factory informers.InformerFactory, ldapClient ldappool.Interface, cacheClient cache.Interface, options *authoptions.AuthenticationOptions) error {
	ws := runtime.NewWebService(GroupVersion)

	handler := newIAMHandler(k8sClient, factory, ldapClient, cacheClient, options)

	// implemented by create CRD object.
	//ws.Route(ws.POST("/users"))
	//ws.Route(ws.DELETE("/users/{user}"))
	//ws.Route(ws.PUT("/users/{user}"))
	//ws.Route(ws.GET("/users/{user}"))

	// TODO move to resources api
	//ws.Route(ws.GET("/users"))

	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(handler.ListRoles).
		Doc("Retrieve the roles that are assigned to the user in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/clusterroles").
		To(handler.ListClusterRoles).
		Doc("List all cluster roles.").
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// TODO merge
	//ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/users"))
	ws.Route(ws.GET("/namespaces/{namespace}/users").
		To(handler.ListNamespaceUsers).
		Doc("List all users in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/users").
		To(handler.ListClusterRoleUsers).
		Doc("List all users that are bound to the specified cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/roles").
		To(handler.ListWorkspaceRoles).
		Doc("List all workspace roles.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/members").
		To(handler.ListWorkspaceUsers).
		Doc("List all members in the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// TODO re-design
	ws.Route(ws.POST("/workspaces/{workspace}/members").
		To(handler.InviteUser).
		Doc("Invite a member to the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/members/{member}").
		To(handler.RemoveUser).
		Doc("Remove the specified member from the workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("member", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	c.Add(ws)
	return nil
}
