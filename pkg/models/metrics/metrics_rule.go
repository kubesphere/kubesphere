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
	"fmt"
	"strings"
	"github.com/emicklei/go-restful"
)

func MakeWorkLoadRule(request *restful.Request) string {
	// kube_pod_info{created_by_kind="DaemonSet",created_by_name="fluent-bit",endpoint="https-main",
	// host_ip="192.168.0.14",instance="10.244.114.187:8443",job="kube-state-metrics",
	// namespace="kube-system",node="i-k89a62il",pod="fluent-bit-l5vxr",
	// pod_ip="10.244.114.175",service="kube-state-metrics"}
	rule := `kube_pod_info{created_by_kind="$1",created_by_name=$2,namespace="$3"}`
	kind := strings.Trim(request.PathParameter("workload_kind"), " ")
	name := strings.Trim(request.QueryParameter("workload_name"), " ")
	namespace := strings.Trim(request.PathParameter("ns_name"), " ")
	if namespace == "" { namespace = ".*" }

	// kind alertnatives values: Deployment StatefulSet ReplicaSet Job DaemonSet
	kind = strings.ToLower(kind)
	if kind == "deployment"{
		kind = "ReplicaSet"
		if name != "" {
			name = "~\"" + name + ".*\""
		}else {
			name = "~\".*\""
		}
		rule = strings.Replace(rule, "$1", kind, -1)
		rule = strings.Replace(rule, "$2", name, -1)
		rule = strings.Replace(rule, "$3", namespace, -1)
		return rule

	}else if kind == "replicaset" {
		kind = "ReplicaSet"
	}else if kind == "statefulset" {
		kind = "StatefulSet"
	}else if kind == "daemonset" {
		kind = "DaemonSet"
	}else if kind == "job" {
		kind = "Job"
	}else {
		kind = ".*"
	}

	if name == "" {
		name = "~\".*\""
	}else  {
		name = "\"" + name + "\""
	}

	rule = strings.Replace(rule, "$1", kind, -1)
	rule = strings.Replace(rule, "$2", name, -1)
	rule = strings.Replace(rule, "$3", namespace, -1)
	return rule
}

func MakeContainerPromQL(request *restful.Request) string {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	poName := strings.Trim(request.PathParameter("pod_name"), " ")
	containerName := strings.Trim(request.PathParameter("container_name"), " ")
	// metricType container_cpu_utilisation  container_memory_utilisation container_memory_utilisation_wo_cache
	metricType := strings.Trim(request.QueryParameter("metrics_name"), " ")
	var promql = ""
	if containerName == "" {
		// all containers maybe use filter
		metricType += "_all"
		promql = RulePromQLTmplMap[metricType]
		promql = strings.Replace(promql, "$1", nsName, -1)
		promql = strings.Replace(promql, "$2", poName, -1)
		container_re2 := strings.Trim(request.QueryParameter("containers_filter"), " ")
		if container_re2 == "" {
			container_re2 = ".*"
		}
		promql = strings.Replace(promql, "$3", container_re2, -1)
		return promql
	}
	promql = RulePromQLTmplMap[metricType]

	promql = strings.Replace(promql, "$1", nsName, -1)
	promql = strings.Replace(promql, "$2", poName, -1)
	promql = strings.Replace(promql, "$3", containerName, -1)
	return promql
}

func MakePodPromQL(request *restful.Request, params []string) string {
	metricType := params[0]
	nsName := params[1]
	nodeID := params[2]
	podName := params[3]
	pod_re2 := params[4]
	var promql = ""
	if nsName != "" {
		// get pod metrics by namespace
		if podName != "" {
			// specific pod
			promql = RulePromQLTmplMap[metricType]
			promql = strings.Replace(promql, "$1", nsName, -1)
			promql = strings.Replace(promql, "$2", podName, -1)

		} else {
			// all pods
			metricType += "_all"
			promql = RulePromQLTmplMap[metricType]
			if pod_re2 == "" {
				pod_re2 = ".*"
			}
			promql = strings.Replace(promql, "$1", nsName, -1)
			promql = strings.Replace(promql, "$2", pod_re2, -1)
		}
	} else if nodeID != "" {
		// get pod metrics by nodeid
		metricType += "_node"
		promql = RulePromQLTmplMap[metricType]
		promql = strings.Replace(promql, "$3", nodeID, -1)
		if podName != "" {
			// specific pod
			promql = strings.Replace(promql, "$2", podName, -1)
		} else {
			// choose pod use re2 expression
			pod_re2 := strings.Trim(request.QueryParameter("pods_filter"), " ")
			if pod_re2 == "" {
				pod_re2 = ".*"
			}
			promql = strings.Replace(promql, "$2", pod_re2, -1)
		}
	}
	return promql
}

func MakeNameSpacePromQL(request *restful.Request, metricsName string) string {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	metricType := metricsName
	var recordingRule = RulePromQLTmplMap[metricType]
	ns_re2 := strings.Trim(request.QueryParameter("namespaces_filter"), " ")
	if nsName != "" {
		ns_re2 = nsName
	} else {
		if ns_re2 == "" {
			ns_re2 = ".*"
		}
	}
	recordingRule = strings.Replace(recordingRule, "$1", ns_re2, -1)
	fmt.Println(recordingRule)
	return recordingRule
}

func MakeNodeorClusterRule(request *restful.Request, metricsName string) string {
	node_id := request.PathParameter("node_id")
	var rule = RulePromQLTmplMap[metricsName]

	if strings.Contains(request.SelectedRoutePath(), "monitoring/cluster") {
		// cluster
		return rule
	} else {
		// node
		nodes_re2 := strings.Trim(request.QueryParameter("nodes_filter"), " ")
		if nodes_re2 == "" {
			nodes_re2 = ".*"
		}
		if strings.Contains(metricsName, "disk") && (!(strings.Contains(metricsName, "read") || strings.Contains(metricsName, "write"))) {
			// disk size promql
			node_filter := ""
			if node_id != "" {
				node_filter = "{" + "node" + "=" + "\"" + node_id + "\"" + "}"
			} else {
				node_filter = "{" + "node" + "=~" + "\"" + nodes_re2 + "\"" + "}"
			}
			rule = strings.Replace(rule, "$1", node_filter, -1)
		} else {
			// cpu, memory, network, disk_iops rules
			if node_id != "" {
				// specific node
				rule = rule + "{" + "node" + "=" + "\"" + node_id + "\"" + "}"
			} else {
				// all nodes or specific nodes filted with re2 syntax
				rule = rule + "{" + "node" + "=~" + "\"" + nodes_re2 + "\"" + "}"
			}
		}
	}
	return rule
}




