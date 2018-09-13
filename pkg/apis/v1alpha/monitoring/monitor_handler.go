/*
Copyright 2018 The KubeSphere Authors.
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
package monitoring

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"net/http"
	"fmt"
	"encoding/json"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models/metrics"
	"kubesphere.io/kubesphere/pkg/client"
)

func validJson (resbody string) {
	data := []byte(resbody)
	//j := map[string]interface
	var dat map[string]interface{}
	jsonErr := json.Unmarshal(data, &dat)
	if jsonErr != nil {
		fmt.Errorf("json parse failed")
	}else {
		fmt.Println(dat)
	}
}

func (u MonitorResource) monitorPod(request *restful.Request, response *restful.Response) {
	promql := metrics.MakePodPromQL(request)
	res, err := client.MakeRequestParams(request, promql)
	//podName := strings.Trim(request.PathParameter("pod_name"), " ")
	//validJson(res)
	if err == nil {
		response.WriteEntity(res)
	} else {
		response.WriteErrorString(http.StatusNotFound, "request parse failed")
	}
}


// domain/v1/monitoring/container?ns=ns_name&pod=po_name&container=[default or regex]&metric_type=[cpu_usage memory_used]%time=t0&start=t1&end=t2&start=t3&step=5s&timeout=duration
// domain/v1/monitoring/namespace/ns_name/container?pod=po_name&container=[regex]&metric_type=[cpu_usage memory_used]%time=t0&start=t1&end=t2&start=t3&step=5s&timeout=duration
func (u MonitorResource) monitorContainer(request *restful.Request, response *restful.Response) {
	promql := metrics.MakeContainerPromQL(request)
	res, err := client.MakeRequestParams(request, promql)
	if err == nil {
		response.WriteEntity(res)
	} else {
		response.WriteErrorString(http.StatusNotFound, "request parse failed")
	}
}

//namespace: domain/v1/monitoring/ns/?namespace=[default or regex]&metric_type=[cpu_usage memory_used]%time=t0&start=t1&end=t2&start=t3&step=5s&timeout=duration
//api convention: domain/v1/monitoring/namespace/ns_name?metric_type=[cpu_usage memory_used]%time=t0&start=t1&end=t2&start=t3&step=5s&timeout=duration
func (u MonitorResource) monitorNameSpace(request *restful.Request, response *restful.Response) {
	recordingRule := metrics.MakeNameSpacePromQL(request)
	res, err := client.MakeRequestParams(request, recordingRule)
	if err == nil {
		response.WriteEntity(res)
	} else {
		response.WriteErrorString(http.StatusNotFound, "request parse failed")
	}
}

// metric_type=[cpu_usage memory_mount memory_used memory_available]&node=node_name%time=t0&start=t1&end=t2&start=t3&step=5s&timeout=duration
func (u MonitorResource) monitorNodeorCluster(request *restful.Request, response *restful.Response) {
	recordingRule := metrics.MakeRecordingRule(request)
	res, err := client.MakeRequestParams(request, recordingRule)
	if err == nil {
		response.WriteEntity(res)
	} else {
		response.WriteErrorString(http.StatusNotFound, "request parse failed")
	}
}

// MonitorResult is just a simple type
type MonitorResult struct {
	monitorResult string `json:"MonitorResult" description:"response of metric query"`
}

type MonitorResource struct {
}

func Register(ws *restful.WebService, subPath string) {
	tags := []string{"monitoring apis"}
	u := MonitorResource{}
	ws.Route(ws.GET(subPath + "/cluster").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor cluster level metrics").
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("cluster_cpu_utilization")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/nodes").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor nodes level metrics").
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("node_cpu_utilization")).
		Param(ws.QueryParameter("nodes_re2", "node re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/nodes/{node_id}").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor specific node level metrics").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("node_cpu_utilization")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces").To(u.monitorNameSpace).
		Filter(route.RouteLogging).
		Doc("monitor namespaces level metrics").
		Param(ws.QueryParameter("namespaces_re2", "namespaces re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("namespace_memory_utilization")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}").To(u.monitorNameSpace).
		Filter(route.RouteLogging).
		Doc("monitor specific namespace level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("namespace_memory_utilization")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}/pods").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor pods level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("pod_re2", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilization_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}/pods/{pod_name}").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor specific pod level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilization_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/nodes/{node_id}/pods").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor pods level metrics by nodeid").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.QueryParameter("pod_re2", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("openpitrix.*")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilization_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/nodes/{node_id}/pods/{pod_name}").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor specific pod level metrics by nodeid").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilization_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)


	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}/pods/{pod_name}/containers").To(u.monitorContainer).
		Filter(route.RouteLogging).
		Doc("monitor containers level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("container_re2", "container re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilization_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}/pods/{pod_name}/containers/{container_name}").To(u.monitorContainer).
		Filter(route.RouteLogging).
		Doc("monitor specific container level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("container_name", "specific container").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilization_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)
}

