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
	"encoding/json"
	"regexp"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models/metrics"
)

func monitorWorkLoadSingleMetric(request *restful.Request, metricsName string) *simplejson.Json {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	podNamesRe2 := getPodNameRegexInWorkLoad(request)
	newpromql := metrics.MakePodPromQL(request, []string{metricsName, nsName, "", "", podNamesRe2})
	podMetrics := client.SendPrometheusRequest(request, newpromql)
	cleanedJson := reformatJson(podMetrics, metricsName)
	return cleanedJson
}

func getPodNameRegexInWorkLoad(request *restful.Request) string {
	promql := metrics.MakeWorkLoadRule(request)
	res := client.SendPrometheusRequest(request, promql)
	data := []byte(res)
	var dat metrics.CommonMetricsResult
	jsonErr := json.Unmarshal(data, &dat)
	if jsonErr != nil {
		glog.Errorln("json parse failed", jsonErr)
	}
	var podNames []string
	for _, x := range dat.Data.Result {
		podName := x.KubePodMetric.Pod
		podNames = append(podNames, podName)
	}
	podNamesRe2 := "^(" + strings.Join(podNames, "|") + ")$"
	return podNamesRe2
}

func (u MonitorResource) monitorPod(request *restful.Request, response *restful.Response) {
	podName := strings.Trim(request.PathParameter("pod_name"), " ")
	var res *simplejson.Json
	if podName != "" {
		// single pod single metric
		metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
		res = monitorPodSingleMetric(request, metricsName)
	} else {
		// multiple pod multiple metric
		res = monitorAllMetrics(request)
	}
	response.WriteAsJson(res)
}

func monitorPodSingleMetric(request *restful.Request, metricsName string) *simplejson.Json {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	nodeID := strings.Trim(request.PathParameter("node_id"), " ")
	podName := strings.Trim(request.PathParameter("pod_name"), " ")
	pod_re2 := strings.Trim(request.QueryParameter("pods_filter"), " ")
	params := []string{metricsName, nsName, nodeID, podName, pod_re2}
	promql := metrics.MakePodPromQL(request, params)
	if promql != "" {
		res := client.SendPrometheusRequest(request, promql)
		cleanedJson := reformatJson(res, metricsName)
		return cleanedJson
	}
	return nil
}

func (u MonitorResource) monitorContainer(request *restful.Request, response *restful.Response) {
	metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
	promql := metrics.MakeContainerPromQL(request)
	res := client.SendPrometheusRequest(request, promql)
	cleanedJson := reformatJson(res, metricsName)
	response.WriteAsJson(cleanedJson)
}

func (u MonitorResource) monitorWorkLoad(request *restful.Request, response *restful.Response) {
	res := monitorAllMetrics(request)
	response.WriteAsJson(res)
}

func (u MonitorResource) monitorNameSpace(request *restful.Request, response *restful.Response) {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	var res *simplejson.Json
	if nsName != "" {
		// single
		metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
		res = monitorNameSpaceSingleMetric(request, metricsName)
	} else {
		// multiple
		res = monitorAllMetrics(request)
	}
	response.WriteAsJson(res)
}

func monitorNameSpaceSingleMetric(request *restful.Request, metricsName string) *simplejson.Json {
	recordingRule := metrics.MakeNameSpacePromQL(request, metricsName)
	res := client.SendPrometheusRequest(request, recordingRule)
	cleanedJson := reformatJson(res, metricsName)
	return cleanedJson
}

func reformatJson(metric string, metricsName string) *simplejson.Json {
	js, err := simplejson.NewJson([]byte(metric))
	if err != nil {
		glog.Errorln(err)
	}
	array, _ := js.Get("data").Get("result").Array()
	metricsLength := len(array)
	for i := 0; i < metricsLength; i++ {
		jstemp := js.Get("data").Get("result").GetIndex(i).Get("metric")
		_, isExist := jstemp.CheckGet("__name__")
		if isExist {
			jstemp.Del("__name__")
		}
	}
	js.Set("metric_name", metricsName)

	return js
}

func collectNodeorClusterMetrics(request *restful.Request, metricsName string, ch chan<- *simplejson.Json) {
	metric := monitorNodeorClusterSingleMetric(request, metricsName)
	ch <- metric
}

func collectNameSpaceMetrics(request *restful.Request, metricsName string, ch chan<- *simplejson.Json) {
	metric := monitorNameSpaceSingleMetric(request, metricsName)
	ch <- metric
}

func collectWorkLoadMetrics(request *restful.Request, metricsName string, ch chan<- *simplejson.Json) {
	metricsName = strings.TrimLeft(metricsName, "workload_")
	metric := monitorWorkLoadSingleMetric(request, metricsName)
	ch <- metric
}

func collectPodMetrics(request *restful.Request, metricsName string, ch chan<- *simplejson.Json) {
	metric := monitorPodSingleMetric(request, metricsName)
	ch <- metric
}

func monitorAllMetrics(request *restful.Request) *simplejson.Json {
	metricsName := strings.Trim(request.QueryParameter("metrics_filter"), " ")
	if metricsName == "" {
		metricsName = ".*"
	}
	path := request.SelectedRoutePath()
	sourceType := path[strings.LastIndex(path, "/")+1 : len(path)-1]
	if strings.Contains(path, "workload") {
		sourceType = "workload"
	}
	var ch = make(chan *simplejson.Json, 10)
	for _, k := range metrics.MetricsNames {
		bol, err := regexp.MatchString(metricsName, k)
		if !bol {
			continue
		}
		if err != nil {
			glog.Errorln("regex match failed", err)
			continue
		}
		if strings.HasPrefix(k, sourceType) {
			if sourceType == "node" || sourceType == "cluster" {
				go collectNodeorClusterMetrics(request, k, ch)
			} else if sourceType == "namespace" {
				go collectNameSpaceMetrics(request, k, ch)
			} else if sourceType == "pod" {
				go collectPodMetrics(request, k, ch)
			} else if sourceType == "workload" {
				go collectWorkLoadMetrics(request, k, ch)
			}
		}
	}
	var metricsArray []*simplejson.Json
	var tempjson *simplejson.Json
	for _, k := range metrics.MetricsNames {
		bol, err := regexp.MatchString(metricsName, k)
		if !bol {
			continue
		}
		if err != nil {
			glog.Errorln("regex match failed")
			continue
		}
		if strings.HasPrefix(k, sourceType) {
			tempjson = <-ch
			if tempjson != nil {
				metricsArray = append(metricsArray, tempjson)
			}
		}
	}

	js := simplejson.New()
	js.Set("metrics_level", sourceType)
	js.Set("results", metricsArray)
	return js
}

func (u MonitorResource) monitorNodeorCluster(request *restful.Request, response *restful.Response) {
	metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
	var res *simplejson.Json
	if metricsName != "" {
		// single
		res = monitorNodeorClusterSingleMetric(request, metricsName)
	} else {
		// multiple
		res = monitorAllMetrics(request)
	}
	response.WriteAsJson(res)
}

func monitorNodeorClusterSingleMetric(request *restful.Request, metricsName string) *simplejson.Json {
	recordingRule := metrics.MakeNodeorClusterRule(request, metricsName)
	res := client.SendPrometheusRequest(request, recordingRule)
	cleanedJson := reformatJson(res, metricsName)
	return cleanedJson
}

type MonitorResource struct {
}

func Register(ws *restful.WebService, subPath string) {
	tags := []string{"monitoring apis"}
	u := MonitorResource{}

	ws.Route(ws.GET(subPath+"/clusters").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor cluster level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("cluster_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/nodes").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor nodes level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("node_cpu_utilisation")).
		Param(ws.QueryParameter("nodes_filter", "node re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/nodes/{node_id}").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor specific node level metrics").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("node_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces").To(u.monitorNameSpace).
		Filter(route.RouteLogging).
		Doc("monitor namespaces level metrics").
		Param(ws.QueryParameter("namespaces_filter", "namespaces re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("namespace_memory_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}").To(u.monitorNameSpace).
		Filter(route.RouteLogging).
		Doc("monitor specific namespace level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("namespace_memory_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/pods").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor pods level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/pods/{pod_name}").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor specific pod level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/nodes/{node_id}/pods").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor pods level metrics by nodeid").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("openpitrix.*")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/nodes/{node_id}/pods/{pod_name}").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor specific pod level metrics by nodeid").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/pods/{pod_name}/containers").To(u.monitorContainer).
		Filter(route.RouteLogging).
		Doc("monitor containers level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("containers_filter", "container re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/pods/{pod_name}/containers/{container_name}").To(u.monitorContainer).
		Filter(route.RouteLogging).
		Doc("monitor specific container level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("container_name", "specific container").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/workloads/{workload_kind}").To(u.monitorWorkLoad).
		Filter(route.RouteLogging).
		Doc("monitor specific workload level metrics").
		Param(ws.PathParameter("ns_name", "namespace").DataType("string").Required(true).DefaultValue("kube-system")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false)).
		Param(ws.PathParameter("workload_kind", "workload kind").DataType("string").Required(true).DefaultValue("daemonset")).
		Param(ws.QueryParameter("workload_name", "workload name").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

}
