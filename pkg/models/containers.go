package models

import (
	"encoding/json"

	"strings"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"

	ksutil "kubesphere.io/kubesphere/pkg/util"

	"fmt"
	"strconv"
)

type ResultContainer struct {
	NodeName      string            `json:"node_name"`
	ContainerName string            `json:"container_name"`
	CPURequest    string            `json:"cpu_request"`
	CPULimit      string            `json:"cpu_limit"`
	MemoryRequest string            `json:"mem_request"`
	MemoryLimit   string            `json:"mem_limit"`
	CPU           []CPUContainer    `json:"cpu"`
	Memory        []MemoryContainer `json:"memory"`
}
type CPUContainer struct {
	TimeStamp      string `json:"timestamp"`
	UsedCPU        string `json:"used_cpu"`
	CPUUtilization string `json:"cpu_utilization"`
}

type MemoryContainer struct {
	TimeStamp         string `json:"timestamp"`
	UsedMemory        string `json:"used_mem"`
	MemoryUtilization string `json:"mem_utilization"`
}

/*
Get all containers under specified namespace in default cluster
*/
func GetContainers(namespace, podName string) []string {
	containersList := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + podName + "/containers")
	var containers []string
	dec := json.NewDecoder(strings.NewReader(containersList))
	err := dec.Decode(&containers)
	if err != nil {
		glog.Error(err)
	}
	return containers
}

func FormatContainersMetrics(nodeName, namespace, podName string) constants.PageableResponse {

	var result constants.PageableResponse
	var resultContainer ResultContainer
	var containers []string
	var total_count int
	containers = GetContainers(namespace, podName)

	for i, container := range containers {
		resultContainer = FormatContainerMetrics(namespace, podName, container)
		if nodeName != "" {
			resultContainer.NodeName = nodeName
		} else {
			resultContainer.NodeName = GetNodeNameForPod(podName, namespace)
		}
		result.Items = append(result.Items, resultContainer)
		total_count = i
	}
	result.TotalCount = total_count + 1

	return result
}

func FormatContainerMetrics(namespace, podName, containerName string) ResultContainer {
	var resultContainer ResultContainer
	var containerCPUMetrics []CPUContainer
	var containerMemMetrics []MemoryContainer
	var cpuMetrics CPUContainer
	var memMetrics MemoryContainer

	resultContainer.ContainerName = containerName
	cpuRequest := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/cpu/request")
	cpuRequest = ksutil.JsonRawMessage(cpuRequest).Find("metrics").ToList()[0].Find("value").ToString()
	if cpuRequest != "" && cpuRequest != "0" {
		resultContainer.CPURequest = cpuRequest
	} else {
		resultContainer.CPURequest = "inf"
	}

	cpuLimit := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/cpu/limit")
	cpuLimit = ksutil.JsonRawMessage(cpuLimit).Find("metrics").ToList()[0].Find("value").ToString()
	if cpuLimit != "" && cpuLimit != "0" {
		resultContainer.CPULimit = cpuLimit
	} else {
		resultContainer.CPULimit = "inf"
	}
	memoryRequest := ksutil.JsonRawMessage(client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/memory/request")).Find("metrics").ToList()[0].Find("value").ToString()
	resultContainer.MemoryRequest = ConvertMemory(memoryRequest)

	memoryLimit := ksutil.JsonRawMessage(client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/memory/limit")).Find("metrics").ToList()[0].Find("value").ToString()
	resultContainer.MemoryLimit = ConvertMemory(memoryLimit)

	cpuUsageRate := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/cpu/usage_rate")
	if cpuUsageRate != "" {
		metrics := ksutil.JsonRawMessage(cpuUsageRate).Find("metrics").ToList()

		for _, metric := range metrics {
			timestamp := metric.Find("timestamp")
			cpu_utilization, _ := strconv.ParseFloat(ConvertCPUUsageRate(metric.Find("value").ToString()), 64)
			cpuMetrics.TimeStamp = timestamp.ToString()
			cpuMetrics.CPUUtilization = fmt.Sprintf("%.3f", cpu_utilization)
			if resultContainer.CPULimit != "inf" {
				cpu_limit, _ := strconv.ParseFloat(resultContainer.CPULimit, 64)
				cpuMetrics.UsedCPU = fmt.Sprintf("%.1f", cpu_limit*cpu_utilization)
			} else {
				cpuMetrics.UsedCPU = "inf"
			}
			glog.Info("pod " + podName + " has limit cpu " + resultContainer.CPULimit + " CPU utilization " + fmt.Sprintf("%.3f", cpu_utilization) + " at time" + timestamp.ToString())
			containerCPUMetrics = append(containerCPUMetrics, cpuMetrics)
		}

	}

	resultContainer.CPU = containerCPUMetrics

	var used_mem_bytes float64

	memUsage := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/memory/usage")
	if memUsage != "" {
		metrics := ksutil.JsonRawMessage(memUsage).Find("metrics").ToList()

		for _, metric := range metrics {
			timestamp := metric.Find("timestamp")
			used_mem_bytes, _ = strconv.ParseFloat(metric.Find("value").ToString(), 64)
			used_mem := used_mem_bytes / 1024 / 1024
			memMetrics.TimeStamp = timestamp.ToString()
			memMetrics.UsedMemory = fmt.Sprintf("%.1f", used_mem)
			memMetrics.MemoryUtilization = fmt.Sprintf("%.3f", CalculateMemoryUsage(memoryRequest, memoryLimit, metric.Find("value").ToString()))
			glog.Info("pod " + podName + " has limit mem " + resultContainer.MemoryLimit + " mem utilization " + memMetrics.MemoryUtilization + " at time" + timestamp.ToString())
			containerMemMetrics = append(containerMemMetrics, memMetrics)
		}
	}

	resultContainer.Memory = containerMemMetrics

	return resultContainer
}
