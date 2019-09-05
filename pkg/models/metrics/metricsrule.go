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
	"strings"
)

// resources_filter = xxxx|xxxx
func MakeWorkloadPromQL(metricName, nsName, resources_filter, wkKind string) string {

	switch wkKind {
	case "deployment":
		wkKind = Deployment
	case "daemonset":
		wkKind = DaemonSet
	case "statefulset":
		wkKind = StatefulSet
	}

	if wkKind == "" {
		resources_filter = Any
	} else if resources_filter == "" {
		if strings.Contains(metricName, "pod") {
			resources_filter = wkKind + ":" + Any
		} else if strings.Contains(metricName, strings.ToLower(wkKind)) {
			resources_filter = Any
		}
	} else {
		var prefix string

		// The "workload_{deployment,statefulset,daemonset}_xxx" metric uses "deployment","statefulset" or "daemonset" label selectors
		// which match exactly a workload name
		// eg. kube_daemonset_status_number_unavailable{daemonset=~"^xxx$"}
		if strings.Contains(metricName, "deployment") || strings.Contains(metricName, "daemonset") || strings.Contains(metricName, "statefulset") {
			// to pass "resources_filter" to PromQL, we reformat it
			prefix = ""
		} else {
			// While workload_{cpu,memory,net}_xxx metrics uses "workload"
			// eg. namespace:workload_cpu_usage:sum{workload="Deployment:xxx"}
			prefix = wkKind + ":"
		}

		filters := strings.Split(resources_filter, "|")
		// reshape it to match PromQL re2 syntax
		resources_filter = ""
		for i, filter := range filters {

			resources_filter += "^" + prefix + filter + "$" // eg. ^Deployment:xxx$

			if i != len(filters)-1 {
				resources_filter += "|"
			}
		}
	}

	var promql = RulePromQLTmplMap[metricName]
	promql = strings.Replace(promql, "$2", nsName, -1)
	promql = strings.Replace(promql, "$3", resources_filter, -1)

	return promql
}

func MakeSpecificWorkloadRule(wkKind, wkName, namespace string) string {
	var rule = PodInfoRule
	if namespace == "" {
		namespace = ".*"
	}
	// alertnatives values: Deployment StatefulSet ReplicaSet DaemonSet
	wkKind = strings.ToLower(wkKind)

	switch wkKind {
	case "deployment":
		wkKind = ReplicaSet
		if wkName != "" {
			wkName = "~\"^" + wkName + `-(\\w)+$"`
		} else {
			wkName = "~\".*\""
		}
		rule = strings.Replace(rule, "$1", wkKind, -1)
		rule = strings.Replace(rule, "$2", wkName, -1)
		rule = strings.Replace(rule, "$3", namespace, -1)
		return rule
	case "replicaset":
		wkKind = ReplicaSet
	case "statefulset":
		wkKind = StatefulSet
	case "daemonset":
		wkKind = DaemonSet
	}

	if wkName == "" {
		wkName = "~\".*\""
	} else {
		wkName = "\"" + wkName + "\""
	}

	rule = strings.Replace(rule, "$1", wkKind, -1)
	rule = strings.Replace(rule, "$2", wkName, -1)
	rule = strings.Replace(rule, "$3", namespace, -1)
	return rule
}

func MakeAllWorkspacesPromQL(metricsName, nsFilter string) string {

	var promql = RulePromQLTmplMap[metricsName]
	nsFilter = "!~\"" + nsFilter + "\""
	promql = strings.Replace(promql, "$1", nsFilter, -1)
	return promql
}

func MakeSpecificWorkspacePromQL(metricsName, nsFilter string, workspace string) string {

	var promql = RulePromQLTmplMap[metricsName]

	nsFilter = "=~\"" + nsFilter + "\""
	workspace = "=~\"^(" + workspace + ")$\""

	promql = strings.Replace(promql, "$1", nsFilter, -1)
	promql = strings.Replace(promql, "$2", workspace, -1)
	return promql
}

func MakeContainerPromQL(nsName, nodeId, podName, containerName, metricName, containerFilter string) string {
	var promql string

	if nsName != "" {
		// get container metrics from namespace-pod
		promql = RulePromQLTmplMap[metricName]
		promql = strings.Replace(promql, "$1", nsName, -1)
	} else {
		// get container metrics from node-pod
		promql = RulePromQLTmplMap[metricName+"_node"]
		promql = strings.Replace(promql, "$1", nodeId, -1)
	}

	promql = strings.Replace(promql, "$2", podName, -1)

	if containerName == "" {

		if containerFilter == "" {
			containerFilter = ".*"
		}
		promql = strings.Replace(promql, "$3", containerFilter, -1)
	} else {
		promql = strings.Replace(promql, "$3", containerName, -1)
	}

	return promql
}

func MakePodPromQL(metricName, nsName, nodeID, podName, podFilter string) string {

	if podFilter == "" {
		podFilter = ".*"
	}

	var promql = ""
	if nsName != "" {
		// get pod metrics by namespace
		if podName != "" {
			// specific pod
			promql = RulePromQLTmplMap[metricName]
			promql = strings.Replace(promql, "$1", nsName, -1)
			promql = strings.Replace(promql, "$2", podName, -1)

		} else {
			// all pods
			metricName += "_all"
			promql = RulePromQLTmplMap[metricName]

			promql = strings.Replace(promql, "$1", nsName, -1)
			promql = strings.Replace(promql, "$2", podFilter, -1)
		}
	} else if nodeID != "" {
		// get pod metrics by nodeid
		metricName += "_node"
		promql = RulePromQLTmplMap[metricName]
		promql = strings.Replace(promql, "$3", nodeID, -1)
		if podName != "" {
			// specific pod
			promql = strings.Replace(promql, "$2", podName, -1)
		} else {
			promql = strings.Replace(promql, "$2", podFilter, -1)
		}
	}
	return promql
}

func MakePVCPromQL(metricName, nsName, pvcName, scName, pvcFilter string) string {
	if pvcFilter == "" {
		pvcFilter = ".*"
	}

	var promql = ""
	if nsName != "" {
		// get pvc metrics by namespace
		if pvcName != "" {
			// specific pvc
			promql = RulePromQLTmplMap[metricName]
			promql = strings.Replace(promql, "$1", nsName, -1)
			promql = strings.Replace(promql, "$2", pvcName, -1)
		} else {
			// all pvc in a specific namespace
			metricName += "_ns"
			promql = RulePromQLTmplMap[metricName]
			promql = strings.Replace(promql, "$1", nsName, -1)
			promql = strings.Replace(promql, "$2", pvcFilter, -1)
		}
	} else {
		if scName != "" {
			// all pvc in a specific storageclass
			metricName += "_sc"
			promql = RulePromQLTmplMap[metricName]
			promql = strings.Replace(promql, "$1", scName, -1)
		}
	}
	return promql
}

func MakeNamespacePromQL(nsName string, nsFilter string, metricsName string) string {
	var recordingRule = RulePromQLTmplMap[metricsName]

	if nsName != "" {
		nsFilter = nsName
	} else {
		if nsFilter == "" {
			nsFilter = ".*"
		}
	}
	recordingRule = strings.Replace(recordingRule, "$1", nsFilter, -1)
	return recordingRule
}

// cluster rule
func MakeClusterRule(metricsName string) string {
	var rule = RulePromQLTmplMap[metricsName]
	return rule
}

// node rule
func MakeNodeRule(nodeID string, nodesFilter string, metricsName string) string {
	var rule = RulePromQLTmplMap[metricsName]

	if nodesFilter == "" {
		nodesFilter = ".*"
	}
	if strings.Contains(metricsName, "disk_size") || strings.Contains(metricsName, "pod") || strings.Contains(metricsName, "usage") || strings.Contains(metricsName, "inode") || strings.Contains(metricsName, "load") {
		// disk size promql
		if nodeID != "" {
			nodesFilter = "{" + "node" + "=" + "\"" + nodeID + "\"" + "}"
		} else {
			nodesFilter = "{" + "node" + "=~" + "\"" + nodesFilter + "\"" + "}"
		}
		rule = strings.Replace(rule, "$1", nodesFilter, -1)
	} else {
		// cpu, memory, network, disk_iops rules
		if nodeID != "" {
			// specific node
			rule = rule + "{" + "node" + "=" + "\"" + nodeID + "\"" + "}"
		} else {
			// all nodes or specific nodes filted with re2 syntax
			rule = rule + "{" + "node" + "=~" + "\"" + nodesFilter + "\"" + "}"
		}
	}

	return rule
}

func MakeComponentRule(metricsName string) string {
	var rule = RulePromQLTmplMap[metricsName]
	return rule
}
