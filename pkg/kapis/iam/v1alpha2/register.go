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
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"net/http"
)

const (
	GroupName = "iam.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(container *restful.Container, im im.IdentityManagementInterface, am am.AccessManagementInterface, options *authoptions.AuthenticationOptions) error {
	ws := runtime.NewWebService(GroupVersion)

	handler := newIAMHandler(im, am, options)

	// global resource
	ws.Route(ws.GET("/users").
		To(handler.ListUsers).
		Doc("List all users.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	// global resource
	ws.Route(ws.GET("/users/{user}").
		To(handler.DescribeUser).
		Doc("Retrieve user details.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.UserDetail{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	// global resource
	ws.Route(ws.GET("/globalroles").
		To(handler.ListGlobalRoles).
		Doc("List all cluster roles.").
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/clusterroles").
		To(handler.ListClusterRoles).
		Doc("List cluster roles.").
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/users").
		To(handler.ListWorkspaceUsers).
		Doc("List all members in the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/workspaceroles").
		To(handler.ListWorkspaceRoles).
		Doc("List all workspace roles.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/users").
		To(handler.ListNamespaceUsers).
		Doc("List all users in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	container.Add(ws)
	return nil
}
