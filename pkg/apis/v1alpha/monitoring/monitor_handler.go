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
	"fmt"
	"encoding/json"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models/metrics"
	"kubesphere.io/kubesphere/pkg/client"
	"strings"
	"github.com/golang/glog"
	"github.com/bitly/go-simplejson"
	"regexp"
)

func (u MonitorResource) monitorWorkLoad(request *restful.Request, response *restful.Response) {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")

	promql := metrics.MakeWorkLoadRule(request)
	fmt.Println(promql)
	res := client.MakeRequestParams(request, promql)
	data := []byte(res)
	var dat metrics.CommonMetricsResult
	jsonErr := json.Unmarshal(data, &dat)
	if jsonErr != nil {
		glog.Errorln("json parse failed")
	}
	podMetrics := ""
	var podNames []string
	for _, x := range dat.Data.Result {
		podName := x.KubePodMetric.Pod
		podNames = append(podNames, podName)
	}

	podNamesRe2 := "^(" + strings.Join(podNames, "|") + ")$"
	fmt.Println(nsName + "\t" + podNamesRe2)
	promql = metrics.MakePodPromQL(request, []string{metricsName, nsName, "", "", podNamesRe2})
	fmt.Println(promql)
	podMetrics =  client.MakeRequestParams(request, promql)
	response.WriteEntity(podMetrics)
}

func (u MonitorResource) monitorPod(request *restful.Request, response *restful.Response) {
	podName := strings.Trim(request.PathParameter("pod_name"), " ")
	res := ""
	if podName != "" {
		// single pod single metric
		metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
		res = monitorPodSingleMertic(request, metricsName)
	}else {
		// multiple pod multiple metric
		res = monitorAllMetrics(request)
	}
	response.WriteEntity(res)
}

func monitorPodSingleMertic(request *restful.Request, metricsName string) string {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	nodeID := strings.Trim(request.PathParameter("node_id"), " ")
	podName := strings.Trim(request.PathParameter("pod_name"), " ")
	pod_re2 := strings.Trim(request.QueryParameter("pods_filter"), " ")
	params := []string{metricsName, nsName, nodeID, podName, pod_re2}
	//promql := metrics.MakePodPromQL(request, metricsName, nsName, nodeID, podName)
	promql := metrics.MakePodPromQL(request, params)
	return client.MakeRequestParams(request, promql)
}

func (u MonitorResource) monitorContainer(request *restful.Request, response *restful.Response) {
	promql := metrics.MakeContainerPromQL(request)
	res := client.MakeRequestParams(request, promql)
	response.WriteEntity(res)
}

func (u MonitorResource) monitorNameSpace(request *restful.Request, response *restful.Response) {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	res := ""
	if nsName != "" {
		// single
		metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
		res = monitorNameSpaceSingleMertic(request, metricsName)
	}else {
		// multiple
		res = monitorAllMetrics(request)
	}
	response.WriteEntity(res)
}

func monitorNameSpaceSingleMertic(request *restful.Request, metricsName string) string {
	recordingRule := metrics.MakeNameSpacePromQL(request, metricsName)
	return client.MakeRequestParams(request, recordingRule)
}

func parseFromJson(mertic string, metricsName string) *simplejson.Json {
	js, err := simplejson.NewJson([]byte(mertic))
	if err != nil {
		glog.Errorln(err)
	}
	js.Set("metric_name", metricsName)
	array, e := js.Get("data").Get("result").Array()
	if e != nil {
		glog.Errorln(err)
	}
	metricsLength := len(array)
	for i:=0; i<metricsLength; i++ {
		jstemp := js.Get("data").Get("result").GetIndex(i).Get("metric")
		_, isExist := jstemp.CheckGet("__name__")
		if isExist {
			jstemp.Del("__name__")
		}
	}
	//b, _ := js.Encode()
	//fmt.Println(string(b))
	return js
}

func collectNodeorClusterMetrics(request *restful.Request, metricsName string, ch chan <- *simplejson.Json) {
	mertic := monitorNodeorClusterSingleMertic(request, metricsName)
	fmt.Println(mertic)
	ch <- parseFromJson(mertic, metricsName)
}

func collectNameSpaceMetrics (request *restful.Request, metricsName string, ch chan <- *simplejson.Json) {
	mertic := monitorNameSpaceSingleMertic(request, metricsName)
	fmt.Println(mertic)
	ch <- parseFromJson(mertic, metricsName)
}

func collectPodMetrics (request *restful.Request, metricsName string, ch chan <- *simplejson.Json) {
	mertic := monitorPodSingleMertic(request, metricsName)
	fmt.Println(mertic)
	ch <- parseFromJson(mertic, metricsName)
}

func monitorAllMetrics(request *restful.Request) string {
	metricsName := strings.Trim(request.QueryParameter("metrics_filter"), " ")
	if metricsName == "" {
		metricsName = ".*"
	}
	path := request.SelectedRoutePath()
	sourceType := path[strings.LastIndex(path,"/") + 1 : len(path) - 1]
	var ch = make(chan *simplejson.Json, 10)
	for _,k := range metrics.MetricsNames {
		bol, err := regexp.MatchString(metricsName, k)
		if !bol {
			continue
		}
		if err != nil {
			glog.Errorln("regex match failed")
			continue
		}
		if strings.HasPrefix(k, sourceType) {
			if sourceType == "node" || sourceType == "cluster"{
				go collectNodeorClusterMetrics(request, k, ch)
			}else if sourceType == "namespace" {
				go collectNameSpaceMetrics(request, k, ch)
			}else if sourceType == "pod" {
				go collectPodMetrics(request, k, ch)
			}
		}
	}

	// TODO 修改成 按照chan内容遍历
	var metricsArray []*simplejson.Json
	var tempjson *simplejson.Json
	for _,k := range metrics.MetricsNames {
		bol, err := regexp.MatchString(metricsName, k)
		if !bol {
			continue
		}
		if err != nil {
			glog.Errorln("regex match failed")
			continue
		}
		if strings.HasPrefix(k, sourceType) {
			tempjson = <- ch
			metricsArray = append(metricsArray, tempjson)
		}
	}
	js := simplejson.New()
	js.Set("metrics_level", sourceType)
	js.Set("results", metricsArray)
	jsByte, err := js.Encode()
	if err != nil {
		glog.Errorln(err)
	}
	return string(jsByte)
}

func (u MonitorResource) monitorNodeorCluster(request *restful.Request, response *restful.Response) {
	metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
	res := ""
	if metricsName != "" {
		// single
		res = monitorNodeorClusterSingleMertic(request, metricsName)
	}else {
		// multiple
		res = monitorAllMetrics(request)
	}
	response.WriteEntity(res)
}

func monitorNodeorClusterSingleMertic(request *restful.Request, metricsName string) string {
	recordingRule := metrics.MakeNodeorClusterRule(request, metricsName)
	return client.MakeRequestParams(request, recordingRule)
}

// MonitorResult is just a simple type
type MonitorResult struct {
	monitorResult string `json:"monitor_result" description:"response of metric query"`
}

type MonitorResource struct {
}

func Register(ws *restful.WebService, subPath string) {
	tags := []string{"monitoring apis"}
	u := MonitorResource{}

	ws.Route(ws.GET(subPath + "/clusters").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor cluster level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("cluster_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/nodes").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor nodes level metrics").
	//Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(false).DefaultValue("node_cpu_utilisation")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("node_cpu_utilisation")).
		Param(ws.QueryParameter("nodes_filter", "node re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/nodes/{node_id}").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor specific node level metrics").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("node_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces").To(u.monitorNameSpace).
		Filter(route.RouteLogging).
		Doc("monitor namespaces level metrics").
		Param(ws.QueryParameter("namespaces_filter", "namespaces re2 expression filter").DataType("string").Required(false).DefaultValue("")).
	//Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(false).DefaultValue("namespace_memory_utilisation")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("namespace_memory_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}").To(u.monitorNameSpace).
		Filter(route.RouteLogging).
		Doc("monitor specific namespace level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("namespace_memory_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}/pods").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor pods level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("")).
	//Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}/pods/{pod_name}").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor specific pod level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/nodes/{node_id}/pods").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor pods level metrics by nodeid").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("openpitrix.*")).
	//Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/nodes/{node_id}/pods/{pod_name}").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor specific pod level metrics by nodeid").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)


	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}/pods/{pod_name}/containers").To(u.monitorContainer).
		Filter(route.RouteLogging).
		Doc("monitor containers level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("containers_filter", "container re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
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
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/namespaces/{ns_name}/workload/{workload_kind}").To(u.monitorWorkLoad).
		Filter(route.RouteLogging).
		Doc("monitor specific workload level metrics").
		Param(ws.PathParameter("ns_name", "namespace").DataType("string").Required(true).DefaultValue("kube-system")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Param(ws.PathParameter("workload_kind", "workload kind").DataType("string").Required(true).DefaultValue("ReplicaSet")).
		Param(ws.QueryParameter("workload_name", "workload name").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(MonitorResult{}).
		Returns(200, "OK", MonitorResult{})).
		Produces(restful.MIME_JSON)
}

