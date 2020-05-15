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
	"kubesphere.io/kubesphere/pkg/api"
	eventsv1alpha1 "kubesphere.io/kubesphere/pkg/api/events/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"net/http"
)

const (
	GroupName = "tenant.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, factory informers.InformerFactory, evtsClient events.Client) error {
	ws := runtime.NewWebService(GroupVersion)
	handler := newTenantHandler(factory, evtsClient)

	ws.Route(ws.GET("/workspaces").
		To(handler.ListWorkspaces).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(handler.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the namespaces of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/events").
		To(handler.Events).
		Doc("Query events against the cluster").
		Param(ws.QueryParameter("operation", "Operation type. This can be one of four types: `query` (for querying events), `statistics` (for retrieving statistical data), `histogram` (for displaying events count by time interval). Defaults to query.").DefaultValue("query")).
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

	c.Add(ws)

	return nil
}
