/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha3

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/models"
)

const (
	GroupName = "tenant.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha3"}

func (h *handler) AddToContainer(c *restful.Container) error {

	ws := runtime.NewWebService(GroupVersion)

	ws.Route(ws.GET("/clusters").
		To(h.ListClusters).
		Deprecate().
		Doc("List clusters available to users").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/workspaces").
		To(h.ListWorkspaceTemplates).
		Deprecate().
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/workspaces/{workspace}").
		To(h.DescribeWorkspaceTemplate).
		Deprecate().
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, tenantv1beta1.WorkspaceTemplate{}).
		Doc("Describe workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/workspaces/{workspace}/clusters").
		To(h.ListWorkspaceClusters).
		Deprecate().
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Doc("List clusters authorized to the specified workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/namespaces").
		To(h.ListNamespaces).
		Deprecate().
		Doc("List the namespaces for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(h.ListNamespaces).
		Deprecate().
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the namespaces of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/workspaces/{workspace}/namespaces/{namespace}").
		To(h.DescribeNamespace).
		Deprecate().
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "project name")).
		Doc("Retrieve namespace details.").
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/workspacetemplates").
		To(h.ListWorkspaceTemplates).
		Deprecate().
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/workspacetemplates/{workspace}").
		To(h.DescribeWorkspaceTemplate).
		Deprecate().
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.WorkspaceTemplate{}).
		Doc("Describe workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/workspaces").
		To(h.ListWorkspaces).
		Deprecate().
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	ws.Route(ws.GET("/workspaces/{workspace}").
		To(h.GetWorkspace).
		Deprecate().
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha1.Workspace{}).
		Doc("Get workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagUserRelatedResources}))

	c.Add(ws)
	return nil
}
