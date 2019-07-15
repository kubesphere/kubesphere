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
	"kubesphere.io/kubesphere/pkg/apiserver/alerting"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/alert"
	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"net/http"
)

const (
	GroupName = "alerting.kubesphere.io"
	RespOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)
	tags := []string{"Alerting"}

	//tags := []string{"ResourceType"}

	ws.Route(ws.GET("/resourcetype").To(alerting.DescribeResourceTypes).
		Doc("Describe Resource Types").
		Param(ws.QueryParameter("rs_type_ids", "Specify resource type ids to query, comma-separated, eg. rst-2loEnEY6Oyzp,rst-2loEnEY6Oyzp.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_type_names", "Specify resource type names to query, comma-separated, eg. container,pod.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of rs_type_id, rs_type_name, rs_type_param, create_time, update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeResourceTypesResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeResourceTypesResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	//tags = []string{"Metric"}

	ws.Route(ws.GET("/metric").To(alerting.DescribeMetrics).
		Doc("Describe Metrics").
		Param(ws.QueryParameter("metric_ids", "Specify metric ids to query, comma-separated, eg. mt-RWXXoJkyJKEm,mt-vnAjqwNP5OPJ.").DataType("string").Required(false)).
		Param(ws.QueryParameter("metric_names", "Specify metric names to query, comma-separated, eg. node_memory_utilisation,node_cpu_utilisation.").DataType("string").Required(false)).
		Param(ws.QueryParameter("status", "Specify metric status to query, comma-separated, eg. active.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_type_ids", "Specify metric resource type ids to query, comma-separated, eg. rst-2loEnEY6Oyzp,rst-3m8ZmxVylG90.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of metric_id, metric_name, metric_param, status, create_time, update_time, rs_type_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeMetricsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeMetricsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	//tags = []string{"Policy"}

	ws.Route(ws.PATCH("/clusters/policy").To(alerting.ModifyPolicyByAlertCluster).
		Doc("Modify Policy By Alert Cluster level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/nodes/policy").To(alerting.ModifyPolicyByAlertNode).
		Doc("Modify Policy By Alert Node Level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/workspaces/policy").To(alerting.ModifyPolicyByAlertWorkspace).
		Doc("Modify Policy By Alert Workspace Level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/workspaces/{workspace}/policy").To(alerting.ModifyPolicyByAlertWorkspace).
		Doc("Modify Policy By Alert Workspace Level").
		Param(ws.PathParameter("workspace", "Specify workspace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/policy").To(alerting.ModifyPolicyByAlertNamespace).
		Doc("Modify Policy By Alert Namespace Level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/{namespace}/policy").To(alerting.ModifyPolicyByAlertNamespace).
		Doc("Modify Policy By Alert Namespace Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/{namespace}/workloads/policy").To(alerting.ModifyPolicyByAlertWorkload).
		Doc("Modify Policy By Alert Workload Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/{namespace}/pods/policy").To(alerting.ModifyPolicyByAlertPod).
		Doc("Modify Policy By Alert Pod Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/nodes/{node}/pods/policy").To(alerting.ModifyPolicyByAlertPod).
		Doc("Modify Policy By Alert Pod Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/{namespace}/pods/{pod}/containers/policy").To(alerting.ModifyPolicyByAlertContainer).
		Doc("Modify Policy By Alert Container Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/nodes/{node}/pods/{pod}/containers/policy").To(alerting.ModifyPolicyByAlertContainer).
		Doc("Modify Policy By Alert Container Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(alerting.PolicyByAlert{}).
		Writes(alerting.ModifyPolicyByAlertResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyPolicyByAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	//tags = []string{"Alert"}

	ws.Route(ws.POST("/clusters/alert").To(alerting.CreateAlertCluster).
		Doc("Create Alert Cluster level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/nodes/alert").To(alerting.CreateAlertNode).
		Doc("Create Alert Node Level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/workspaces/alert").To(alerting.CreateAlertWorkspace).
		Doc("Create Alert Workspace Level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/workspaces/{workspace}/alert").To(alerting.CreateAlertWorkspace).
		Doc("Create Alert Workspace Level").
		Param(ws.PathParameter("workspace", "Specify workspace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/namespaces/alert").To(alerting.CreateAlertNamespace).
		Doc("Create Alert Namespace Level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/namespaces/{namespace}/alert").To(alerting.CreateAlertNamespace).
		Doc("Create Alert Namespace Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/namespaces/{namespace}/workloads/alert").To(alerting.CreateAlertWorkload).
		Doc("Create Alert Workload Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/namespaces/{namespace}/pods/alert").To(alerting.CreateAlertPod).
		Doc("Create Alert Pod Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/nodes/{node}/pods/alert").To(alerting.CreateAlertPod).
		Doc("Create Alert Pod Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/namespaces/{namespace}/pods/{pod}/containers/alert").To(alerting.CreateAlertContainer).
		Doc("Create Alert Container Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/nodes/{node}/pods/{pod}/containers/alert").To(alerting.CreateAlertContainer).
		Doc("Create Alert Container Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.AlertInfo{}).
		Writes(pb.CreateAlertResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateAlertResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/clusters/alert").To(alerting.ModifyAlertByNameCluster).
		Doc("Modify Alert By Name Cluster level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/nodes/alert").To(alerting.ModifyAlertByNameNode).
		Doc("Modify Alert By Name Node Level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/workspaces/alert").To(alerting.ModifyAlertByNameWorkspace).
		Doc("Modify Alert By Name Workspace Level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/workspaces/{workspace}/alert").To(alerting.ModifyAlertByNameWorkspace).
		Doc("Modify Alert By Name Workspace Level").
		Param(ws.PathParameter("workspace", "Specify workspace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/alert").To(alerting.ModifyAlertByNameNamespace).
		Doc("Modify Alert By Name Namespace Level").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/{namespace}/alert").To(alerting.ModifyAlertByNameNamespace).
		Doc("Modify Alert By Name Namespace Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/{namespace}/workloads/alert").To(alerting.ModifyAlertByNameWorkload).
		Doc("Modify Alert By Name Workload Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/{namespace}/pods/alert").To(alerting.ModifyAlertByNamePod).
		Doc("Modify Alert By Name Pod Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/nodes/{node}/pods/alert").To(alerting.ModifyAlertByNamePod).
		Doc("Modify Alert By Name Pod Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/namespaces/{namespace}/pods/{pod}/containers/alert").To(alerting.ModifyAlertByNameContainer).
		Doc("Modify Alert By Name Container Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.PATCH("/nodes/{node}/pods/{pod}/containers/alert").To(alerting.ModifyAlertByNameContainer).
		Doc("Modify Alert By Name Container Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Alert{}).
		Writes(alerting.ModifyAlertByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.ModifyAlertByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/clusters/alert").To(alerting.DeleteAlertsByNameCluster).
		Doc("Delete Alerts By Name Cluster level").
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/nodes/alert").To(alerting.DeleteAlertsByNameNode).
		Doc("Delete Alerts By Name Node Level").
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/workspaces/alert").To(alerting.DeleteAlertsByNameWorkspace).
		Doc("Delete Alerts By Name Workspace Level").
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/workspaces/{workspace}/alert").To(alerting.DeleteAlertsByNameWorkspace).
		Doc("Delete Alerts By Name Workspace Level").
		Param(ws.PathParameter("workspace", "Specify workspace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/namespaces/alert").To(alerting.DeleteAlertsByNameNamespace).
		Doc("Delete Alerts By Name Namespace Level").
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/namespaces/{namespace}/alert").To(alerting.DeleteAlertsByNameNamespace).
		Doc("Delete Alerts By Name Namespace Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/namespaces/{namespace}/workloads/alert").To(alerting.DeleteAlertsByNameWorkload).
		Doc("Delete Alerts By Name Workload Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/namespaces/{namespace}/pods/alert").To(alerting.DeleteAlertsByNamePod).
		Doc("Delete Alerts By Name Pod Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/nodes/{node}/pods/alert").To(alerting.DeleteAlertsByNamePod).
		Doc("Delete Alerts By Name Pod Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/namespaces/{namespace}/pods/{pod}/containers/alert").To(alerting.DeleteAlertsByNameContainer).
		Doc("Delete Alerts By Name Container Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/nodes/{node}/pods/{pod}/containers/alert").To(alerting.DeleteAlertsByNameContainer).
		Doc("Delete Alerts By Name Container Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_names", "Specify alert names to delete, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(alerting.DeleteAlertsByNameResponse{}).
		Returns(http.StatusOK, RespOK, alerting.DeleteAlertsByNameResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/clusters/alert").To(alerting.DescribeAlertDetailsCluster).
		Doc("Describe Alert Details Cluster level").
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/alert").To(alerting.DescribeAlertDetailsNode).
		Doc("Describe Alert Details Node Level").
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/alert").To(alerting.DescribeAlertDetailsWorkspace).
		Doc("Describe Alert Details Workspace Level").
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/{workspace}/alert").To(alerting.DescribeAlertDetailsWorkspace).
		Doc("Describe Alert Details Workspace Level").
		Param(ws.PathParameter("workspace", "Specify workspace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/alert").To(alerting.DescribeAlertDetailsNamespace).
		Doc("Describe Alert Details Namespace Level").
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/alert").To(alerting.DescribeAlertDetailsNamespace).
		Doc("Describe Alert Details Namespace Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/alert").To(alerting.DescribeAlertDetailsWorkload).
		Doc("Describe Alert Details Workload Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/alert").To(alerting.DescribeAlertDetailsPod).
		Doc("Describe Alert Details Pod Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/alert").To(alerting.DescribeAlertDetailsPod).
		Doc("Describe Alert Details Pod Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/alert").To(alerting.DescribeAlertDetailsContainer).
		Doc("Describe Alert Details Container Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}/containers/alert").To(alerting.DescribeAlertDetailsContainer).
		Doc("Describe Alert Details Container Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertDetailsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertDetailsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/clusters/alert_status").To(alerting.DescribeAlertStatusCluster).
		Doc("Describe Alert Status Cluster level").
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/alert_status").To(alerting.DescribeAlertStatusNode).
		Doc("Describe Alert Status Node Level").
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/alert_status").To(alerting.DescribeAlertStatusWorkspace).
		Doc("Describe Alert Status Workspace Level").
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/{workspace}/alert_status").To(alerting.DescribeAlertStatusWorkspace).
		Doc("Describe Alert Status Workspace Level").
		Param(ws.PathParameter("workspace", "Specify workspace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/alert_status").To(alerting.DescribeAlertStatusNamespace).
		Doc("Describe Alert Status Namespace Level").
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/alert_status").To(alerting.DescribeAlertStatusNamespace).
		Doc("Describe Alert Status Namespace Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/alert_status").To(alerting.DescribeAlertStatusWorkload).
		Doc("Describe Alert Status Workload Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/alert_status").To(alerting.DescribeAlertStatusPod).
		Doc("Describe Alert Status Pod Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/alert_status").To(alerting.DescribeAlertStatusPod).
		Doc("Describe Alert Status Pod Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/alert_status").To(alerting.DescribeAlertStatusContainer).
		Doc("Describe Alert Status Container Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}/containers/alert_status").To(alerting.DescribeAlertStatusContainer).
		Doc("Describe Alert Status Container Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("alert_ids", "Specify alert ids to query, comma-separated, eg. al-WgBBMmRv3rMP,al-QVjOxnkD3mwW.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("disables", "Specify alert disabled status, comma-separated, eg. true,false.").DataType("string").Required(false)).
		Param(ws.QueryParameter("running_status", "Specify alert running status, comma-separated, eg. adding,running,deleting,updating,migrating.").DataType("string").Required(false)).
		Param(ws.QueryParameter("policy_ids", "Specify policy ids to query, comma-separated, eg. pl-zZmOyYqqx7vo,pl-zZPP12kqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("creators", "Specify creators to query, comma-separated, eg. admin,user1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rs_filter_ids", "Specify resource filter ids to query, comma-separated, eg. rf-ZyzVP265N3l5,rf-zZ416xNqx7vo.").DataType("string").Required(false)).
		Param(ws.QueryParameter("executor_ids", "Specify alert executor ids to query, comma-separated, eg. alerting-executor-5f8d9bb8b9-4rl95,alerting-executor-5f8d9bb8b9-jw728.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.alert_id, t1.alert_name, t1.disabled, t1.running_status, t1.alert_status, t1.create_time, t1.update_time, t1.policy_id, t1.rs_filter_id, t1.executor_id, t2.policy_id, t2.policy_name, t2.policy_description, t2.policy_config, t2.creator, t2.available_start_time, t2.available_end_time, t2.create_time, t2.update_time, t2.rs_type_id, t3.rs_filter_id, t3.rs_filter_name, t3.rs_filter_param, t3.status, t3.create_time, t3.update_time, t3.rs_type_id, t4.rs_type_id, t4.rs_type_name, t4.rs_type_param, t4.create_time, t4.update_time, t5.action_id, t5.action_name, t5.trigger_status, t5.trigger_action, t5.create_time, t5.update_time, t5.policy_id, t5.nf_address_list_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeAlertStatusResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeAlertStatusResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	//tags = []string{"History"}

	ws.Route(ws.GET("/clusters/history").To(alerting.DescribeHistoryDetailCluster).
		Doc("Describe History Detail Cluster level").
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/history").To(alerting.DescribeHistoryDetailNode).
		Doc("Describe History Detail Node Level").
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/history").To(alerting.DescribeHistoryDetailWorkspace).
		Doc("Describe History Detail Workspace Level").
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/{workspace}/history").To(alerting.DescribeHistoryDetailWorkspace).
		Doc("Describe History Detail Workspace Level").
		Param(ws.PathParameter("workspace", "Specify workspace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/history").To(alerting.DescribeHistoryDetailNamespace).
		Doc("Describe History Detail Namespace Level").
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/history").To(alerting.DescribeHistoryDetailNamespace).
		Doc("Describe History Detail Namespace Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/history").To(alerting.DescribeHistoryDetailWorkload).
		Doc("Describe History Detail Workload Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/history").To(alerting.DescribeHistoryDetailPod).
		Doc("Describe History Detail Pod Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/history").To(alerting.DescribeHistoryDetailPod).
		Doc("Describe History Detail Pod Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/history").To(alerting.DescribeHistoryDetailContainer).
		Doc("Describe History Detail Container Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}/containers/history").To(alerting.DescribeHistoryDetailContainer).
		Doc("Describe History Detail Container Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("search_word", "Specify search word").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_names", "Specify history names to query, comma-separated, eg. alert-trigger,alert-resume.").DataType("string").Required(false)).
		Param(ws.QueryParameter("alert_names", "Specify alert names to query, comma-separated, eg. alert-1,alert-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_names", "Specify rule names to query, comma-separated, eg. 内存利用率,CPU利用率.").DataType("string").Required(false)).
		Param(ws.QueryParameter("events", "Specify history events to query, comma-separated, eg. triggered,resumed,sent_success,sent_failed,commented.").DataType("string").Required(false)).
		Param(ws.QueryParameter("rule_ids", "Specify rule ids to query, comma-separated, eg. rl-nKEQK7kAGDYv,rl-RG3GJ8X8JQY1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resource_names", "Specify resource names to query, comma-separated, eg. master,node1.").DataType("string").Required(false)).
		Param(ws.QueryParameter("recent", "List most recent history after latest trigger event. One of true, false.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of t1.history_id, t1.history_name, t1.event, t1.content, t1.notification_id, t1.create_time, t1.update_time, t1.alert_id, t1.rule_id, t1.resource_name, t2.rule_id, t2.rule_name, t2.disabled, t2.monitor_periods, t2.severity, t2.metrics_type, t2.condition_type, t2.thresholds, t2.unit, t2.consecutive_count, t2.inhibit, t2.create_time, t2.update_time, t2.policy_id, t2.metric_id, t3.alert_id, t3.alert_name, t3.disabled, t3.running_status, t3.alert_status, t3.create_time, t3.update_time, t3.policy_id, t3.rs_filter_id, t3.executor_id, t4.rs_filter_id, t4.rs_filter_name, t4.rs_filter_param, t4.status, t4.create_time, t4.update_time, t4.rs_type_id, t5.metric_id, t5.metric_name, t5.metric_param, t5.status, t5.create_time, t5.update_time, t5.rs_type_id, t6.rs_type_id, t6.rs_type_name, t6.rs_type_param, t6.create_time, t6.update_time.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeHistoryDetailResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeHistoryDetailResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	//tags = []string{"Comment"}

	ws.Route(ws.POST("/comment").To(alerting.CreateComment).
		Doc("Create Comment").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Comment{}).
		Writes(pb.CreateCommentResponse{}).
		Returns(http.StatusOK, RespOK, pb.CreateCommentResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/comment").To(alerting.DescribeComments).
		Doc("Describe Comments").
		Param(ws.QueryParameter("comment_ids", "Specify comment ids to query, comma-separated, eg. cm-Dp7Z7VjvKnYL, cm-zyyGZZ640Op9.").DataType("string").Required(false)).
		Param(ws.QueryParameter("addressers", "Specify comment addresser names to query, comma-separated, eg. user1,tester2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("contents", "Specify comment contents to query, comma-separated.").DataType("string").Required(false)).
		Param(ws.QueryParameter("history_ids", "Specify history ids to query, comma-separated, eg. hs-zzz8NpjopLvj,hs-zzzlXz1ELM6y.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_key", "Sort key. One of comment_id, addresser, content, create_time, update_time, history_id.").DataType("string").Required(false)).
		Param(ws.QueryParameter("reverse", "Sort order, true-desc, false-asc.").DataType("bool").DefaultValue("false").Required(false)).
		Param(ws.QueryParameter("offset", "Beginning index of result to return. Use this option together with limit.").DataType("uint32").Required(false)).
		Param(ws.QueryParameter("limit", "Size of result to return.").DataType("uint32").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(pb.DescribeCommentsResponse{}).
		Returns(http.StatusOK, RespOK, pb.DescribeCommentsResponse{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	//tags = []string{"Resource"}

	ws.Route(ws.GET("/clusters/resource").To(alerting.DescribeResourcesCluster).
		Doc("Describe Resources Cluster level").
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/resource").To(alerting.DescribeResourcesNode).
		Doc("Describe Resources Node Level").
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/resource").To(alerting.DescribeResourcesWorkspace).
		Doc("Describe Resources Workspace Level").
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/{workspace}/resource").To(alerting.DescribeResourcesWorkspace).
		Doc("Describe Resources Workspace Level").
		Param(ws.PathParameter("workspace", "Specify workspace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/resource").To(alerting.DescribeResourcesNamespace).
		Doc("Describe Resources Namespace Level").
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/resource").To(alerting.DescribeResourcesNamespace).
		Doc("Describe Resources Namespace Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/resource").To(alerting.DescribeResourcesWorkload).
		Doc("Describe Resources Workload Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("workload_kind", "workload kind specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/resource").To(alerting.DescribeResourcesPod).
		Doc("Describe Resources Pod Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/resource").To(alerting.DescribeResourcesPod).
		Doc("Describe Resources Pod Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/resource").To(alerting.DescribeResourcesContainer).
		Doc("Describe Resources Container Level").
		Param(ws.PathParameter("namespace", "Specify namespace").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}/containers/resource").To(alerting.DescribeResourcesContainer).
		Doc("Describe Resources Container Level").
		Param(ws.PathParameter("node", "Specify node id").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("pod", "Specify pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("selector", "Specify selector, eg. [{\"app\": \"fluentbit-operator\"}]").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]string{}).
		Returns(http.StatusOK, RespOK, []string{})).
		Consumes(restful.MIME_JSON, constants.MIME_MERGEPATCH).
		Produces(restful.MIME_JSON)

	c.Add(ws)
	return nil
}
