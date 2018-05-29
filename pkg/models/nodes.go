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

	"strings"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"

	ksutil "kubesphere.io/kubesphere/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"
	"strconv"
)

type ResultNodes struct {
	Nodes []ResultNode `json:"nodes"`
}
type ResultNode struct {
	NodeName     string       `json:"node_name"`
	PodsCount    string       `json:"pods_count"`
	PodsCapacity string       `json:"pods_capacity"`
	CPU          []CPUNode    `json:"cpu"`
	Memory       []MemoryNode `json:"memory"`
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
			cpu_utilization, _ := strconv.ParseFloat(metric.Find("value").ToString(), 64)
			cpuMetrics.TimeStamp = timestamp.ToString()
			cpuMetrics.TotalCPU = fmt.Sprintf("%.1f", total_cpu)
			cpuMetrics.CPUUtilization = fmt.Sprintf("%.3f", cpu_utilization/1000)
			cpuMetrics.UsedCPU = fmt.Sprintf("%.1f", total_cpu*cpu_utilization/1000)

			glog.Info("node " + nodeName + " has total cpu " + fmt.Sprintf("%.1f", total_cpu) + " CPU utilization " + fmt.Sprintf("%.3f", cpu_utilization/1000) + " at time" + timestamp.ToString())
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
	resultNode.PodsCount = strconv.Itoa(len(GetPodsForNode(nodeName,"")))
	resultNode.PodsCapacity = getPodsCapacity(nodeName)
	resultNode.CPU = nodeCPUMetrics
	resultNode.Memory = nodeMemMetrics

	return resultNode
}

func getPodsCapacity(nodeName string) string {
	var pods_capacity string
	cli := client.NewK8sClient()

	node, err := cli.CoreV1().Nodes().Get(nodeName,metav1.GetOptions{})

	if err != nil {
		glog.Error(err)
	} else {
		pods_capacity = node.Status.Capacity.Pods().String()
	}
	fmt.Println(pods_capacity)
	return pods_capacity
}