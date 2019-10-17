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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	devopsv1alpha2 "kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/tenant"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"

	"net/http"
)

const (
	GroupName = "tenant.kubesphere.io"
	RespOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {
	ok := "ok"
	ws := runtime.NewWebService(GroupVersion)

	ws.Route(ws.GET("/workspaces").
		To(tenant.ListWorkspaces).
		Returns(http.StatusOK, ok, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}").
		To(tenant.DescribeWorkspace).
		Doc("Describe the specified workspace").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, ok, v1alpha1.Workspace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/rules").
		To(tenant.ListWorkspaceRules).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the rules of the specified workspace for the current user").
		Returns(http.StatusOK, ok, models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/rules").
		To(tenant.ListNamespaceRules).
		Param(ws.PathParameter("namespace", "the name of the namespace")).
		Doc("List the rules of the specified namespace for the current user").
		Returns(http.StatusOK, ok, models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/devops/{devops}/rules").
		To(tenant.ListDevopsRules).
		Param(ws.PathParameter("devops", "devops project ID")).
		Doc("List the rules of the specified DevOps project for the current user").
		Returns(http.StatusOK, ok, models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(tenant.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the namespaces of the specified workspace for the current user").
		Returns(http.StatusOK, ok, []v1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{member}/namespaces").
		To(tenant.ListNamespacesByUsername).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("member", "workspace member's username")).
		Doc("List the namespaces for the workspace member").
		Returns(http.StatusOK, ok, []v1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.POST("/workspaces/{workspace}/namespaces").
		To(tenant.CreateNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create a namespace in the specified workspace").
		Returns(http.StatusOK, ok, []v1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/namespaces/{namespace}").
		To(tenant.DeleteNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "the name of the namespace")).
		Doc("Delete the specified namespace from the workspace").
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/devops").
		To(tenant.ListDevopsProjects).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.QueryParameter(params.PagingParam, "page").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(ws.QueryParameter(params.ConditionsParam, "query conditions").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Doc("List devops projects for the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{member}/devops").
		To(tenant.ListDevopsProjectsByUsername).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("member", "workspace member's username")).
		Param(ws.QueryParameter(params.PagingParam, "page").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(ws.QueryParameter(params.ConditionsParam, "query conditions").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Returns(http.StatusOK, ok, models.PageableResponse{}).
		Doc("List the devops projects for the workspace member").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/devopscount").
		To(tenant.GetDevOpsProjectsCount).
		Returns(http.StatusOK, ok, struct {
			Count uint32 `json:"count"`
		}{}).
		Doc("Get the devops projects count for the member").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.POST("/workspaces/{workspace}/devops").
		To(tenant.CreateDevopsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create a devops project in the specified workspace").
		Reads(devopsv1alpha2.DevOpsProject{}).
		Returns(http.StatusOK, RespOK, devopsv1alpha2.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/devops/{devops}").
		To(tenant.DeleteDevopsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("devops", "devops project ID")).
		Doc("Delete the specified devops project from the workspace").
		Returns(http.StatusOK, RespOK, devopsv1alpha2.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/logs").
		To(tenant.LogQuery).
		Doc("Query cluster-level logs in a multi-tenants environment").
		Param(ws.QueryParameter("operation", "Operation type. This can be one of four types: query (for querying logs), statistics (for retrieving statistical data), histogram (for displaying log count by time interval) and export (for exporting logs). Defaults to query.").DefaultValue("query").DataType("string").Required(false)).
		Param(ws.QueryParameter("workspaces", "A comma-separated list of workspaces. This field restricts the query to specified workspaces. For example, the following filter matches the workspace my-ws and demo-ws: `my-ws,demo-ws`").DataType("string").Required(false)).
		Param(ws.QueryParameter("workspace_query", "A comma-separated list of keywords. Differing from **workspaces**, this field performs fuzzy matching on workspaces. For example, the following value limits the query to workspaces whose name contains the word my(My,MY,...) *OR* demo(Demo,DemO,...): `my,demo`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("namespaces", "A comma-separated list of namespaces. This field restricts the query to specified namespaces. For example, the following filter matches the namespace my-ns and demo-ns: `my-ns,demo-ns`").DataType("string").Required(false)).
		Param(ws.QueryParameter("namespace_query", "A comma-separated list of keywords. Differing from **namespaces**, this field performs fuzzy matching on namespaces. For example, the following value limits the query to namespaces whose name contains the word my(My,MY,...) *OR* demo(Demo,DemO,...): `my,demo`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("workloads", "A comma-separated list of workloads. This field restricts the query to specified workloads. For example, the following filter matches the workload my-wl and demo-wl: `my-wl,demo-wl`").DataType("string").Required(false)).
		Param(ws.QueryParameter("workload_query", "A comma-separated list of keywords. Differing from **workloads**, this field performs fuzzy matching on workloads. For example, the following value limits the query to workloads whose name contains the word my(My,MY,...) *OR* demo(Demo,DemO,...): `my,demo`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pods", "A comma-separated list of pods. This field restricts the query to specified pods. For example, the following filter matches the pod my-po and demo-po: `my-po,demo-po`").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "A comma-separated list of keywords. Differing from **pods**, this field performs fuzzy matching on pods. For example, the following value limits the query to pods whose name contains the word my(My,MY,...) *OR* demo(Demo,DemO,...): `my,demo`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "A comma-separated list of containers. This field restricts the query to specified containers. For example, the following filter matches the container my-cont and demo-cont: `my-cont,demo-cont`").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "A comma-separated list of keywords. Differing from **containers**, this field performs fuzzy matching on containers. For example, the following value limits the query to containers whose name contains the word my(My,MY,...) *OR* demo(Demo,DemO,...): `my,demo`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "A comma-separated list of keywords. The query returns logs which contain at least one keyword. Case-insensitive matching. For example, if the field is set to `err,INFO`, the query returns any log containing err(ERR,Err,...) *OR* INFO(info,InFo,...).").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "Time interval. It requires **operation** is set to histogram. The format is [0-9]+[smhdwMqy]. Defaults to 15m (i.e. 15 min).").DefaultValue("15m").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "Start time of query. Default to 0. The format is a string representing milliseconds since the epoch, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "End time of query. Default to now. The format is a string representing milliseconds since the epoch, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "Sort order. One of acs, desc. This field sorts logs by timestamp.").DataType("string").DefaultValue("desc").Required(false)).
		Param(ws.QueryParameter("from", "The offset from the result set. This field returns query results from the specified offset. It requires **operation** is set to query. Defaults to 0 (i.e. from the beginning of the result set).").DataType("integer").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result to return. It requires **operation** is set to query. Defaults to 10 (i.e. 10 log records).").DataType("integer").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}).
		Writes(v1alpha2.Response{}).
		Returns(http.StatusOK, RespOK, v1alpha2.Response{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, "text/plain")

	c.Add(ws)
	return nil
}
