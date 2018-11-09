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

func MakeWorkloadRule(wkKind, wkName, namespace string) string {
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
			wkName = "~\"" + wkName + ".*\""
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

func MakeWorkspacePromQL(metricsName string, nsFilter string) string {
	promql := RulePromQLTmplMap[metricsName]
	promql = strings.Replace(promql, "$1", nsFilter, -1)
	return promql
}

func MakeContainerPromQL(nsName, podName, containerName, metricName, containerFilter string) string {
	var promql = ""
	if containerName == "" {
		// all containers maybe use filter
		metricName += "_all"
		promql = RulePromQLTmplMap[metricName]
		promql = strings.Replace(promql, "$1", nsName, -1)
		promql = strings.Replace(promql, "$2", podName, -1)

		if containerFilter == "" {
			containerFilter = ".*"
		}

		promql = strings.Replace(promql, "$3", containerFilter, -1)
		return promql
	}
	promql = RulePromQLTmplMap[metricName]

	promql = strings.Replace(promql, "$1", nsName, -1)
	promql = strings.Replace(promql, "$2", podName, -1)
	promql = strings.Replace(promql, "$3", containerName, -1)
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
	if strings.Contains(metricsName, "disk_size") || strings.Contains(metricsName, "pod") || strings.Contains(metricsName, "usage") || strings.Contains(metricsName, "inode") {
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
