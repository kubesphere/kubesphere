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

package v1alpha3

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/kapis/tenant/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	monitoringclient "kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

const (
	GroupName = "tenant.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha3"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func AddToContainer(c *restful.Container, factory informers.InformerFactory, k8sclient kubernetes.Interface,
	ksclient kubesphere.Interface, evtsClient events.Client, loggingClient logging.Client,
	auditingclient auditing.Client, am am.AccessManagementInterface, im im.IdentityManagementInterface, authorizer authorizer.Authorizer,
	monitoringclient monitoringclient.Interface, cache cache.Cache, meteringOptions *meteringclient.Options, opClient openpitrix.Interface) error {
	mimePatch := []string{restful.MIME_JSON, runtime.MimeMergePatchJson, runtime.MimeJsonPatchJson}

	ws := runtime.NewWebService(GroupVersion)
	v1alpha2Handler := v1alpha2.NewTenantHandler(factory, k8sclient, ksclient, evtsClient, loggingClient, auditingclient, am, im, authorizer, monitoringclient, resourcev1alpha3.NewResourceGetter(factory, cache), meteringOptions, opClient)
	handler := newTenantHandler(factory, k8sclient, ksclient, evtsClient, loggingClient, auditingclient, am, im, authorizer, monitoringclient, resourcev1alpha3.NewResourceGetter(factory, cache), meteringOptions, opClient)

	ws.Route(ws.POST("/workspacetemplates").
		To(v1alpha2Handler.CreateWorkspaceTemplate).
		Reads(tenantv1alpha2.WorkspaceTemplate{}).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.WorkspaceTemplate{}).
		Doc("Create workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	ws.Route(ws.DELETE("/workspacetemplates/{workspace}").
		To(v1alpha2Handler.DeleteWorkspaceTemplate).
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Doc("Delete workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	ws.Route(ws.PUT("/workspacetemplates/{workspace}").
		To(v1alpha2Handler.UpdateWorkspaceTemplate).
		Param(ws.PathParameter("workspace", "workspace name")).
		Reads(tenantv1alpha2.WorkspaceTemplate{}).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.WorkspaceTemplate{}).
		Doc("Update workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	ws.Route(ws.PATCH("/workspacetemplates/{workspace}").
		To(v1alpha2Handler.PatchWorkspaceTemplate).
		Param(ws.PathParameter("workspace", "workspace name")).
		Consumes(mimePatch...).
		Reads(tenantv1alpha2.WorkspaceTemplate{}).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.WorkspaceTemplate{}).
		Doc("Update workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	ws.Route(ws.GET("/workspacetemplates").
		To(v1alpha2Handler.ListWorkspaceTemplates).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	ws.Route(ws.GET("/workspacetemplates/{workspace}").
		To(v1alpha2Handler.DescribeWorkspaceTemplate).
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.WorkspaceTemplate{}).
		Doc("Describe workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	ws.Route(ws.GET("/workspaces").
		To(handler.ListWorkspaces).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	ws.Route(ws.GET("/workspaces/{workspace}").
		To(handler.GetWorkspace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha1.Workspace{}).
		Doc("Get workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	c.Add(ws)
	return nil
}
