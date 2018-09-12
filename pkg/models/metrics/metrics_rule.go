package metrics

import (
	"github.com/emicklei/go-restful"
	"strings"
)

func MakeContainerPromQL(request *restful.Request) string {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	poName := strings.Trim(request.PathParameter("pod_name"), " ")
	containerName := strings.Trim(request.PathParameter("container_name"), " ")
	metricType := strings.Trim(request.HeaderParameter("metrics_name"), " ")
	// metricType container_cpu_utilisation  container_memory_utilisation container_memory_utilisation_wo_cache
	var promql = ""
	if containerName == "" {
		// all containers maybe use filter
		metricType += "_all"
		promql = promqlTempMap[metricType]
		promql = strings.Replace(promql, "$1", nsName, -1)
		promql = strings.Replace(promql, "$2", poName, -1)
		container_re2 := strings.Trim(request.HeaderParameter("container_re2"), " ")
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

	metricType := strings.Trim(request.HeaderParameter("metrics_name"), " ")
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
			pod_re2 := strings.Trim(request.HeaderParameter("pod_re2"), " ")
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
			pod_re2 := strings.Trim(request.HeaderParameter("pod_re2"), " ")
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
	metricType := request.HeaderParameter("metrics_name")
	var recordingRule = recordingRuleTmplMap[metricType]
	if nsName != "" {
		//specific namespace
		recordingRule = recordingRule + "{" + "namespace" + "=" + "\"" + nsName + "\"" + "}"
	}else {
		ns_re2 := strings.Trim(request.HeaderParameter("namespaces_re2"), " ")
		if ns_re2 != "" {
			recordingRule = recordingRule + "{" + "namespace" + "=~" + "\"" + ns_re2 + "\"" + "}"
		}
	}
	return recordingRule
}

func MakeRecordingRule(request *restful.Request) string {
	metricsName := request.HeaderParameter("metrics_name")
	node_id := request.PathParameter("node_id")
	var recordingRule = ""

	if strings.Contains(request.SelectedRoutePath(), "monitoring/cluster") {
		// cluster
		recordingRule = recordingRuleTmplMap[metricsName]
	}else {
		// node
		if node_id != "" {
			// specific node
			recordingRule = recordingRuleTmplMap[metricsName]
			recordingRule = recordingRule + "{" + "node" + "=" + "\"" + node_id + "\"" + "}"
		}else {
			// all nodes or specific nodes filted with re2 syntax
			nodes_re2 := strings.Trim(request.HeaderParameter("nodes_re2"), " ")
			recordingRule = recordingRuleTmplMap[metricsName]
			if nodes_re2 != "" {
				recordingRule = recordingRule + "{" + "node" + "=~" + "\"" + nodes_re2 + "\"" + "}"
			}
		}
	}
	return recordingRule
}
