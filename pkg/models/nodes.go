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

package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"
	v1beta2 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
	kubeclient "kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	ksutil "kubesphere.io/kubesphere/pkg/util"
)

const (
	//status: "False"
	OutOfDisk = "OutOfDisk"
	//status: "False"
	MemoryPressure = "MemoryPressure"
	//status: "False"
	DiskPressure = "DiskPressure"
	//status: "False"
	PIDPressure = "PIDPressure"
	//status: "True"
	KubeletReady = "Ready"
)

type ResultNode struct {
	NodeName      string       `json:"node_name"`
	NodeStatus    string       `json:"node_status"`
	PodsCount     string       `json:"pods_count"`
	PodsCapacity  string       `json:"pods_capacity"`
	UsedFS        string       `json:"used_fs"`
	TotalFS       string       `json:"total_fs"`
	FSUtilization string       `json:"fs_utilization"`
	CPU           []CPUNode    `json:"cpu"`
	Memory        []MemoryNode `json:"memory"`
}

type CPUNode struct {
	TimeStamp      string `json:"timestamp"`
	UsedCPU        string `json:"used_cpu"`
	TotalCPU       string `json:"total_cpu"`
	CPUUtilization string `json:"cpu_utilization"`
}

type MemoryNode struct {
	TimeStamp         string `json:"timestamp"`
	UsedMemory        string `json:"used_mem"`
	TotalMemory       string `json:"total_mem"`
	MemoryUtilization string `json:"mem_utilization"`
}

/*
Get all nodes in default cluster
*/
func GetNodes() []string {
	nodesList := client.GetHeapsterMetrics("/nodes")
	var nodes []string
	dec := json.NewDecoder(strings.NewReader(nodesList))
	err := dec.Decode(&nodes)
	if err != nil {
		glog.Error(err)
	}
	return nodes
}

/*
Format cpu/memory data for specified node
*/
func FormatNodeMetrics(nodeName string) ResultNode {
	var resultNode ResultNode
	var nodeCPUMetrics []CPUNode
	var nodeMemMetrics []MemoryNode
	var cpuMetrics CPUNode
	var memMetrics MemoryNode
	var total_cpu float64
	var total_mem float64

	cpuNodeAllocated := client.GetHeapsterMetrics("/nodes/" + nodeName + "/metrics/cpu/node_allocatable")
	if cpuNodeAllocated != "" {
		var err error
		total_cpu, err = strconv.ParseFloat(ksutil.JsonRawMessage(cpuNodeAllocated).Find("metrics").ToList()[0].Find("value").ToString(), 64)
		if err == nil {
			total_cpu = total_cpu / 1000
		}
	}

	cpuUsageRate := client.GetHeapsterMetrics("/nodes/" + nodeName + "/metrics/cpu/usage_rate")
	if cpuUsageRate != "" {
		metrics := ksutil.JsonRawMessage(cpuUsageRate).Find("metrics").ToList()

		for _, metric := range metrics {
			timestamp := metric.Find("timestamp")
			cpu_utilization, _ := strconv.ParseFloat(ConvertCPUUsageRate(metric.Find("value").ToString()), 64)
			cpuMetrics.TimeStamp = timestamp.ToString()
			cpuMetrics.TotalCPU = fmt.Sprintf("%.1f", total_cpu)
			cpuMetrics.CPUUtilization = fmt.Sprintf("%.3f", cpu_utilization)
			cpuMetrics.UsedCPU = fmt.Sprintf("%.1f", total_cpu*cpu_utilization)

			glog.Info("node " + nodeName + " has total cpu " + fmt.Sprintf("%.1f", total_cpu) + " CPU utilization " + fmt.Sprintf("%.3f", cpu_utilization) + " at time" + timestamp.ToString())
			nodeCPUMetrics = append(nodeCPUMetrics, cpuMetrics)
		}

	}

	memNodeAllocated := client.GetHeapsterMetrics("/nodes/" + nodeName + "/metrics/memory/node_allocatable")
	var total_mem_bytes, used_mem_bytes float64
	if memNodeAllocated != "" {
		var err error
		total_mem_bytes, err = strconv.ParseFloat(ksutil.JsonRawMessage(memNodeAllocated).Find("metrics").ToList()[0].Find("value").ToString(), 64)
		if err == nil {
			total_mem = total_mem_bytes / 1024 / 1024 / 1024
		}
	}

	memUsage := client.GetHeapsterMetrics("/nodes/" + nodeName + "/metrics/memory/usage")
	if memUsage != "" {
		metrics := ksutil.JsonRawMessage(memUsage).Find("metrics").ToList()

		for _, metric := range metrics {
			timestamp := metric.Find("timestamp")
			used_mem_bytes, _ = strconv.ParseFloat(metric.Find("value").ToString(), 64)
			used_mem := used_mem_bytes / 1024 / 1024 / 1024

			memMetrics.TimeStamp = timestamp.ToString()
			memMetrics.TotalMemory = fmt.Sprintf("%.1f", total_mem)
			memMetrics.UsedMemory = fmt.Sprintf("%.1f", used_mem)
			memMetrics.MemoryUtilization = fmt.Sprintf("%.3f", used_mem_bytes/total_mem_bytes)
			glog.Info("node " + nodeName + " has total mem " + fmt.Sprintf("%.1f", total_mem) + " mem utilization " + fmt.Sprintf("%.3f", used_mem_bytes/total_mem_bytes) + " at time" + timestamp.ToString())
			nodeMemMetrics = append(nodeMemMetrics, memMetrics)
		}
	}

	resultNode.NodeName = nodeName
	resultNode.PodsCount = strconv.Itoa(len(GetPodsForNode(nodeName, "")))
	nodeResObj := getNodeResObj(nodeName)
	resultNode.PodsCapacity = nodeResObj.Status.Capacity.Pods().String()
	resultNode.NodeStatus = getNodeStatus(nodeResObj)
	resultNode.UsedFS, resultNode.TotalFS, resultNode.FSUtilization = getNodeFileSystemStatus(nodeResObj)
	resultNode.CPU = nodeCPUMetrics
	resultNode.Memory = nodeMemMetrics

	return resultNode
}

func getNodeResObj(nodeName string) *v1.Node {
	cli := client.NewK8sClient()

	node, err := cli.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})

	if err != nil {
		glog.Error(err)
	} else {
		return node
	}
	return nil
}

func getNodeStatus(node *v1.Node) string {

	status := "Ready"
	conditions := node.Status.Conditions
	for _, cond := range conditions {
		if cond.Type == DiskPressure && cond.Status == "True" {
			status = "NotReady"
			break
		}
		if cond.Type == OutOfDisk && cond.Status == "True" {
			status = "NotReady"
			break
		}
		if cond.Type == MemoryPressure && cond.Status == "True" {
			status = "NotReady"
		}
		if cond.Type == PIDPressure && cond.Status == "True" {
			status = "NotReady"
			break
		}
		if cond.Type == KubeletReady && cond.Status == "False" {
			status = "NotReady"
			break
		}
	}
	return status
}

func getNodeFileSystemStatus(node *v1.Node) (string, string, string) {

	nodeMetricsAsStr := client.GetCAdvisorMetrics(node.Annotations["alpha.kubernetes.io/provided-node-ip"])
	if nodeMetricsAsStr != "" {
		usedBytesAsStr, _ := strconv.ParseFloat(ksutil.JsonRawMessage(nodeMetricsAsStr).Find("node").Find("fs").Find("usedBytes").ToString(), 64)
		capacityBytesAsStr, _ := strconv.ParseFloat(ksutil.JsonRawMessage(nodeMetricsAsStr).Find("node").Find("fs").Find("capacityBytes").ToString(), 64)
		return fmt.Sprintf("%.1f", usedBytesAsStr/1024/1024/1024), fmt.Sprintf("%.1f", capacityBytesAsStr/1024/1024/1024), fmt.Sprintf("%.3f", usedBytesAsStr/capacityBytesAsStr)
	}
	return "", "", ""
}

func DrainNode(nodename string) (msg constants.MessageResponse, err error) {

	k8sclient := kubeclient.NewK8sClient()
	var options metav1.ListOptions
	pods := make([]v1.Pod, 0)
	options.FieldSelector = "spec.nodeName=" + nodename
	podList, err := k8sclient.CoreV1().Pods("").List(options)
	if err != nil {

		glog.Fatal(err)
		return msg, err

	}
	options.FieldSelector = ""
	daemonsetList, err := k8sclient.AppsV1beta2().DaemonSets("").List(options)

	if err != nil {

		glog.Fatal(err)
		return msg, err

	}
	if len(podList.Items) > 0 {

		for _, pod := range podList.Items {

			if !containDaemonset(pod, *daemonsetList) {

				pods = append(pods, pod)
			}
		}
	}
	//create eviction
	var eviction policy.Eviction
	eviction.Kind = "Eviction"
	eviction.APIVersion = "policy/v1beta1"
	if len(pods) > 0 {

		for _, pod := range pods {

			eviction.Namespace = pod.Namespace
			eviction.Name = pod.Name
			err := k8sclient.CoreV1().Pods(pod.Namespace).Evict(&eviction)
			if err != nil {
				return msg, err
			}
		}

	}
	msg.Message = fmt.Sprintf("success")
	return msg, nil

}

func containDaemonset(pod v1.Pod, daemonsetList v1beta2.DaemonSetList) bool {

	flag := false
	for _, daemonset := range daemonsetList.Items {

		if strings.Contains(pod.Name, daemonset.Name) {

			flag = true
		}

	}
	return flag

}
