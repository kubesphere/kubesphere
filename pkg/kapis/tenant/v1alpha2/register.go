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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/api"
	auditingv1alpha1 "kubesphere.io/kubesphere/pkg/api/auditing/v1alpha1"
	eventsv1alpha1 "kubesphere.io/kubesphere/pkg/api/events/v1alpha1"
	loggingv1alpha2 "kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"net/http"
)

const (
	GroupName = "tenant.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func AddToContainer(c *restful.Container, factory informers.InformerFactory, k8sclient kubernetes.Interface, ksclient kubesphere.Interface, evtsClient events.Client, loggingClient logging.Interface, auditingclient auditing.Client) error {
	mimePatch := []string{restful.MIME_JSON, runtime.MimeMergePatchJson, runtime.MimeJsonPatchJson}

	ws := runtime.NewWebService(GroupVersion)
	handler := newTenantHandler(factory, k8sclient, ksclient, evtsClient, loggingClient, auditingclient)

	ws.Route(ws.GET("/clusters").
		To(handler.ListClusters).
		Doc("List clusters available to users").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.POST("/workspaces").
		To(handler.CreateWorkspace).
		Reads(tenantv1alpha2.WorkspaceTemplate{}).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.WorkspaceTemplate{}).
		Doc("Create workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.DELETE("/workspaces/{workspace}").
		To(handler.DeleteWorkspace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Doc("Delete workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.PUT("/workspaces/{workspace}").
		To(handler.UpdateWorkspace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Reads(tenantv1alpha2.WorkspaceTemplate{}).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.WorkspaceTemplate{}).
		Doc("Update workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.PATCH("/workspaces/{workspace}").
		To(handler.PatchWorkspace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Consumes(mimePatch...).
		Reads(tenantv1alpha2.WorkspaceTemplate{}).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.WorkspaceTemplate{}).
		Doc("Update workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces").
		To(handler.ListWorkspaces).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}").
		To(handler.DescribeWorkspace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.WorkspaceTemplate{}).
		Doc("Describe workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/clusters").
		To(handler.ListWorkspaceClusters).
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Doc("List clusters authorized to the specified workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/namespaces").
		To(handler.ListNamespaces).
		Doc("List the namespaces for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/federatednamespaces").
		To(handler.ListFederatedNamespaces).
		Doc("List the federated namespaces for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/federatednamespaces").
		To(handler.ListFederatedNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the federated namespaces of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(handler.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the namespaces of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/devops").
		To(handler.ListDevOpsProjects).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the devops projects of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}/devops").
		To(handler.ListDevOpsProjects).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspacemember username")).
		Doc("List the devops projects of specified workspace for the workspace member").
		Reads(corev1.Namespace{}).
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/namespaces/{namespace}").
		To(handler.DescribeNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "project name")).
		Doc("Retrieve namespace details.").
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.DELETE("/workspaces/{workspace}/namespaces/{namespace}").
		To(handler.DeleteNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "project name")).
		Doc("Delete namespace.").
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.POST("/workspaces/{workspace}/namespaces").
		To(handler.CreateNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the namespaces of the specified workspace for the current user").
		Reads(corev1.Namespace{}).
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}/namespaces").
		To(handler.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspacemember username")).
		Doc("List the namespaces of the specified workspace for the workspace member").
		Reads(corev1.Namespace{}).
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.PUT("/workspaces/{workspace}/namespaces/{namespace}").
		To(handler.UpdateNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "project name")).
		Reads(corev1.Namespace{}).
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.PATCH("/workspaces/{workspace}/namespaces/{namespace}").
		To(handler.PatchNamespace).
		Consumes(mimePatch...).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "project name")).
		Reads(corev1.Namespace{}).
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/events").
		To(handler.Events).
		Doc("Query events against the cluster").
		Param(ws.QueryParameter("operation", "Operation type. This can be one of three types: `query` (for querying events), `statistics` (for retrieving statistical data), `histogram` (for displaying events count by time interval). Defaults to query.").DefaultValue("query")).
		Param(ws.QueryParameter("workspace_filter", "A comma-separated list of workspaces. This field restricts the query to specified workspaces. For example, the following filter matches the workspace my-ws and demo-ws: `my-ws,demo-ws`.")).
		Param(ws.QueryParameter("workspace_search", "A comma-separated list of keywords. Differing from **workspace_filter**, this field performs fuzzy matching on workspaces. For example, the following value limits the query to workspaces whose name contains the word my(My,MY,...) *OR* demo(Demo,DemO,...): `my,demo`.")).
		Param(ws.QueryParameter("involved_object_namespace_filter", "A comma-separated list of namespaces. This field restricts the query to specified `involvedObject.namespace`.")).
		Param(ws.QueryParameter("involved_object_namespace_search", "A comma-separated list of keywords. Differing from **involved_object_namespace_filter**, this field performs fuzzy matching on `involvedObject.namespace`")).
		Param(ws.QueryParameter("involved_object_name_filter", "A comma-separated list of names. This field restricts the query to specified `involvedObject.name`.")).
		Param(ws.QueryParameter("involved_object_name_search", "A comma-separated list of keywords. Differing from **involved_object_name_filter**, this field performs fuzzy matching on `involvedObject.name`.")).
		Param(ws.QueryParameter("involved_object_kind_filter", "A comma-separated list of kinds. This field restricts the query to specified `involvedObject.kind`.")).
		Param(ws.QueryParameter("reason_filter", "A comma-separated list of reasons. This field restricts the query to specified `reason`.")).
		Param(ws.QueryParameter("reason_search", "A comma-separated list of keywords. Differing from **reason_filter**, this field performs fuzzy matching on `reason`.")).
		Param(ws.QueryParameter("message_search", "A comma-separated list of keywords. This field performs fuzzy matching on `message`.")).
		Param(ws.QueryParameter("type_filter", "Type of event matching on `type`. This can be one of two types: `Warning`, `Normal`")).
		Param(ws.QueryParameter("start_time", "Start time of query (limits `lastTimestamp`). The format is a string representing seconds since the epoch, eg. 1136214245.")).
		Param(ws.QueryParameter("end_time", "End time of query (limits `lastTimestamp`). The format is a string representing seconds since the epoch, eg. 1136214245.")).
		Param(ws.QueryParameter("interval", "Time interval. It requires **operation** is set to `histogram`. The format is [0-9]+[smhdwMqy]. Defaults to 15m (i.e. 15 min).").DefaultValue("15m")).
		Param(ws.QueryParameter("sort", "Sort order. One of asc, desc. This field sorts events by `lastTimestamp`.").DataType("string").DefaultValue("desc")).
		Param(ws.QueryParameter("from", "The offset from the result set. This field returns query results from the specified offset. It requires **operation** is set to `query`. Defaults to 0 (i.e. from the beginning of the result set).").DataType("integer").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result set to return. It requires **operation** is set to `query`. Defaults to 10 (i.e. 10 event records).").DataType("integer").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.EventsQueryTag}).
		Writes(eventsv1alpha1.APIResponse{}).
		Returns(http.StatusOK, api.StatusOK, eventsv1alpha1.APIResponse{}))

	ws.Route(ws.GET("/logs").
		To(handler.QueryLogs).
		Doc("Query logs against the cluster.").
		Param(ws.QueryParameter("operation", "Operation type. This can be one of four types: query (for querying logs), statistics (for retrieving statistical data), histogram (for displaying log count by time interval) and export (for exporting logs). Defaults to query.").DefaultValue("query").DataType("string").Required(false)).
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
		Param(ws.QueryParameter("start_time", "Start time of query. Default to 0. The format is a string representing seconds since the epoch, eg. 1559664000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "End time of query. Default to now. The format is a string representing seconds since the epoch, eg. 1559664000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "Sort order. One of asc, desc. This field sorts logs by timestamp.").DataType("string").DefaultValue("desc").Required(false)).
		Param(ws.QueryParameter("from", "The offset from the result set. This field returns query results from the specified offset. It requires **operation** is set to query. Defaults to 0 (i.e. from the beginning of the result set).").DataType("integer").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result to return. It requires **operation** is set to query. Defaults to 10 (i.e. 10 log records).").DataType("integer").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.LogQueryTag}).
		Writes(loggingv1alpha2.APIResponse{}).
		Returns(http.StatusOK, api.StatusOK, loggingv1alpha2.APIResponse{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, "text/plain")

	ws.Route(ws.GET("/auditing/events").
		To(handler.Auditing).
		Doc("Query auditing events against the cluster").
		Param(ws.QueryParameter("operation", "Operation type. This can be one of three types: `query` (for querying events), `statistics` (for retrieving statistical data), `histogram` (for displaying events count by time interval). Defaults to query.").DefaultValue("query")).
		Param(ws.QueryParameter("workspace_filter", "A comma-separated list of workspaces. This field restricts the query to specified workspaces. For example, the following filter matches the workspace my-ws and demo-ws: `my-ws,demo-ws`.")).
		Param(ws.QueryParameter("workspace_search", "A comma-separated list of keywords. Differing from **workspace_filter**, this field performs fuzzy matching on workspaces. For example, the following value limits the query to workspaces whose name contains the word my(My,MY,...) *OR* demo(Demo,DemO,...): `my,demo`.")).
		Param(ws.QueryParameter("objectref_namespace_filter", "A comma-separated list of namespaces. This field restricts the query to specified `ObjectRef.Namespace`.")).
		Param(ws.QueryParameter("objectref_namespace_search", "A comma-separated list of keywords. Differing from **objectref_namespace_filter**, this field performs fuzzy matching on `ObjectRef.Namespace`.")).
		Param(ws.QueryParameter("objectref_name_filter", "A comma-separated list of names. This field restricts the query to specified `ObjectRef.Name`.")).
		Param(ws.QueryParameter("objectref_name_search", "A comma-separated list of keywords. Differing from **objectref_name_filter**, this field performs fuzzy matching on `ObjectRef.Name`.")).
		Param(ws.QueryParameter("level_filter", "A comma-separated list of levels. This know values are Metadata, Request, RequestResponse.")).
		Param(ws.QueryParameter("verb_filter", "A comma-separated list of verbs. This field restricts the query to specified verb. This field restricts the query to specified `Verb`.")).
		Param(ws.QueryParameter("user_filter", "A comma-separated list of user. This field restricts the query to specified user. For example, the following filter matches the user user1 and user2: `user1,user2`.")).
		Param(ws.QueryParameter("user_search", "A comma-separated list of keywords. Differing from **user_filter**, this field performs fuzzy matching on 'User.username'. For example, the following value limits the query to user whose name contains the word my(My,MY,...) *OR* demo(Demo,DemO,...): `my,demo`.")).
		Param(ws.QueryParameter("group_search", "A comma-separated list of keywords. This field performs fuzzy matching on 'User.Groups'. For example, the following value limits the query to group which contains the word my(My,MY,...) *OR* demo(Demo,DemO,...): `my,demo`.")).
		Param(ws.QueryParameter("source_ip_search", "A comma-separated list of keywords. This field performs fuzzy matching on 'SourceIPs'. For example, the following value limits the query to SourceIPs which contains 127.0 *OR* 192.168.: `127.0,192.168.`.")).
		Param(ws.QueryParameter("objectref_resource_filter", "A comma-separated list of resource. This field restricts the query to specified ip. This field restricts the query to specified `ObjectRef.Resource`.")).
		Param(ws.QueryParameter("objectref_subresource_filter", "A comma-separated list of subresource. This field restricts the query to specified subresource. This field restricts the query to specified `ObjectRef.Subresource`.")).
		Param(ws.QueryParameter("response_code_filter", "A comma-separated list of response status code. This field restricts the query to specified response status code. This field restricts the query to specified `ResponseStatus.code`.")).
		Param(ws.QueryParameter("response_status_filter", "A comma-separated list of response status. This field restricts the query to specified response status. This field restricts the query to specified `ResponseStatus.status`.")).
		Param(ws.QueryParameter("start_time", "Start time of query (limits `RequestReceivedTimestamp`). The format is a string representing seconds since the epoch, eg. 1136214245.")).
		Param(ws.QueryParameter("end_time", "End time of query (limits `RequestReceivedTimestamp`). The format is a string representing seconds since the epoch, eg. 1136214245.")).
		Param(ws.QueryParameter("interval", "Time interval. It requires **operation** is set to `histogram`. The format is [0-9]+[smhdwMqy]. Defaults to 15m (i.e. 15 min).").DefaultValue("15m")).
		Param(ws.QueryParameter("sort", "Sort order. One of asc, desc. This field sorts events by `RequestReceivedTimestamp`.").DataType("string").DefaultValue("desc")).
		Param(ws.QueryParameter("from", "The offset from the result set. This field returns query results from the specified offset. It requires **operation** is set to `query`. Defaults to 0 (i.e. from the beginning of the result set).").DataType("integer").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result set to return. It requires **operation** is set to `query`. Defaults to 10 (i.e. 10 event records).").DataType("integer").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AuditingQueryTag}).
		Writes(auditingv1alpha1.APIResponse{}).
		Returns(http.StatusOK, api.StatusOK, auditingv1alpha1.APIResponse{}))

	c.Add(ws)
	return nil
}
