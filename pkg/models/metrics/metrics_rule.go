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

package metrics

import (
	"github.com/emicklei/go-restful"
	"strings"
)

func MakeContainerPromQL(request *restful.Request) string {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	poName := strings.Trim(request.PathParameter("pod_name"), " ")
	containerName := strings.Trim(request.PathParameter("container_name"), " ")
	metricType := strings.Trim(request.QueryParameter("metrics_name"), " ")
	// metricType container_cpu_utilisation  container_memory_utilisation container_memory_utilisation_wo_cache
	var promql = ""
	if containerName == "" {
		// all containers maybe use filter
		metricType += "_all"
		promql = promqlTempMap[metricType]
		promql = strings.Replace(promql, "$1", nsName, -1)
		promql = strings.Replace(promql, "$2", poName, -1)
		container_re2 := strings.Trim(request.QueryParameter("container_re2"), " ")
		if container_re2 == "" {
			container_re2 = ".*"
		}
		promql = strings.Replace(promql, "$3", container_re2, -1)
		return promql
	}
	promql = promqlTempMap[metricType]

	promql = strings.Replace(promql, "$1", nsName, -1)
	promql = strings.Replace(promql, "$2", poName, -1)
	promql = strings.Replace(promql, "$3", containerName, -1)
	return promql
}

func MakePodPromQL(request *restful.Request) string {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	nodeID := strings.Trim(request.PathParameter("node_id"), " ")
	podName := strings.Trim(request.PathParameter("pod_name"), " ")

	metricType := strings.Trim(request.QueryParameter("metrics_name"), " ")
	var promql = ""
	if nsName != "" {
		// 通过namespace 获得 pod
		if podName != "" {
			// specific pod
			promql = promqlTempMap[metricType]
			promql = strings.Replace(promql, "$1", nsName, -1)
			promql = strings.Replace(promql, "$2", podName, -1)

		}else {
			// all pods
			metricType += "_all"
			promql = promqlTempMap[metricType]
			pod_re2 := strings.Trim(request.QueryParameter("pod_re2"), " ")
			if pod_re2 == "" {
				pod_re2 = ".*"
			}
			promql = strings.Replace(promql, "$1", nsName, -1)
			promql = strings.Replace(promql, "$2", pod_re2, -1)
		}
	}else if nodeID != "" {
		//通过 nodeid 获得 pod
		metricType += "_node"
		promql = promqlTempMap[metricType]
		promql = strings.Replace(promql, "$3", nodeID, -1)
		if podName  != "" {
			// 取指定的 pod，不适用 re2表达式
			promql = strings.Replace(promql, "$2", podName, -1)
		}else {
			//获取所有的pod，可以使用 re2
			pod_re2 := strings.Trim(request.QueryParameter("pod_re2"), " ")
			if pod_re2 == "" {
				pod_re2 = ".*"
			}
			promql = strings.Replace(promql, "$2", pod_re2, -1)
		}
	}else {

	}
	return promql
}

func MakeNameSpacePromQL(request *restful.Request) string {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	metricType := request.QueryParameter("metrics_name")
	var recordingRule = recordingRuleTmplMap[metricType]
	if nsName != "" {
		//specific namespace
		recordingRule = recordingRule + "{" + "namespace" + "=" + "\"" + nsName + "\"" + "}"
	}else {
		ns_re2 := strings.Trim(request.QueryParameter("namespaces_re2"), " ")
		if ns_re2 != "" {
			recordingRule = recordingRule + "{" + "namespace" + "=~" + "\"" + ns_re2 + "\"" + "}"
		}
	}
	return recordingRule
}

func MakeRecordingRule(request *restful.Request) string {
	metricsName := request.QueryParameter("metrics_name")
	node_id := request.PathParameter("node_id")
	var rule = recordingRuleTmplMap[metricsName]

	if strings.Contains(request.SelectedRoutePath(), "monitoring/cluster") {
		// cluster
		return rule
	}else {
		// node
		nodes_re2 := strings.Trim(request.QueryParameter("nodes_re2"), " ")
		if nodes_re2 == "" {
			nodes_re2 = ".*"
		}
		if strings.Contains(metricsName, "disk") {
			// disk metrics
			node_filter := ""
			if node_id != "" {
				node_filter = "{" + "node" + "=" + "\"" + node_id + "\"" + "}"
			}else {
				node_filter = "{" + "node" + "=~" + "\"" + nodes_re2 + "\"" + "}"
			}
			rule = strings.Replace(rule, "$1", node_filter, -1)
		}else {
			// cpu memory net metrics
			if node_id != "" {
				// specific node
				rule = rule + "{" + "node" + "=" + "\"" + node_id + "\"" + "}"
			}else {
				// all nodes or specific nodes filted with re2 syntax
				rule = rule + "{" + "node" + "=~" + "\"" + nodes_re2 + "\"" + "}"
			}
		}
	}
	return rule
}
