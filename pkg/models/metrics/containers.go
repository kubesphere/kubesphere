package metrics

import (
	"encoding/json"

	"strings"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"

	"fmt"
)

type ContainerMetrics struct {
	NodeName      string                   `json:"node_name"`
	ContainerName string                   `json:"container_name"`
	CpuRequest    string                   `json:"cpu_request"`
	CpuLimit      string                   `json:"cpu_limit"`
	MemoryRequest string                   `json:"mem_request"`
	MemoryLimit   string                   `json:"mem_limit"`
	Cpu           []ContainerCpuMetrics    `json:"cpu"`
	Memory        []ContainerMemoryMetrics `json:"memory"`
}
type ContainerCpuMetrics struct {
	TimeStamp string `json:"timestamp"`
	UsedCpu   string `json:"used_cpu"`
}

type ContainerMemoryMetrics struct {
	TimeStamp  string `json:"timestamp"`
	UsedMemory string `json:"used_mem"`
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
	var resultContainer ContainerMetrics
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

func FormatContainerMetrics(namespace, podName, containerName string) ContainerMetrics {
	var resultContainer ContainerMetrics
	var containerCPUMetrics []ContainerCpuMetrics
	var containerMemMetrics []ContainerMemoryMetrics
	var cpuMetrics ContainerCpuMetrics
	var memMetrics ContainerMemoryMetrics

	resultContainer.ContainerName = containerName

	cpuRequest := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/cpu/request")
	cpuRequestMetrics, err := cpuRequest.GetObjectArray("metrics")
	if err == nil && len(cpuRequestMetrics) != 0 {
		requestCpu, _ := cpuRequestMetrics[0].GetFloat64("value")
		resultContainer.CpuRequest = FormatResourceLimit(requestCpu)
	} else {
		resultContainer.CpuRequest = Inf
	}

	cpuLimit := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/cpu/limit")
	cpuLimitMetrics, err := cpuLimit.GetObjectArray("metrics")
	if err == nil && len(cpuLimitMetrics) != 0 {
		limitCpu, _ := cpuLimitMetrics[0].GetFloat64("value")
		resultContainer.CpuLimit = FormatResourceLimit(limitCpu)
	} else {
		resultContainer.CpuLimit = Inf
	}

	memoryRequst := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/memory/request")
	memoryRequstMetrics, err := memoryRequst.GetObjectArray("metrics")
	if err == nil && len(memoryRequstMetrics) != 0 {
		requestMemory, _ := memoryRequstMetrics[0].GetFloat64("value")
		requestMemory = requestMemory / 1024 / 1024
		resultContainer.MemoryRequest = FormatResourceLimit(requestMemory)
	} else {
		resultContainer.MemoryRequest = Inf
	}

	memoryLimit := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/memory/limit")
	memoryLimitMetrics, err := memoryLimit.GetObjectArray("metrics")
	if err == nil && len(memoryLimitMetrics) != 0 {
		limitMemory, _ := memoryLimitMetrics[0].GetFloat64("value")
		limitMemory = limitMemory / 1024 / 1024
		resultContainer.MemoryLimit = FormatResourceLimit(limitMemory)
	} else {
		resultContainer.MemoryLimit = Inf
	}

	cpuUsageRate := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/cpu/usage_rate")
	cpuUsageRateMetrics, err := cpuUsageRate.GetObjectArray("metrics")
	if err != nil {
		glog.Error(err)
		containerCPUMetrics = make([]ContainerCpuMetrics, 0)
	} else {
		for _, metric := range cpuUsageRateMetrics {
			timestamp, _ := metric.GetString("timestamp")
			cpuMetrics.TimeStamp = timestamp
			usedCpu, _ := metric.GetFloat64("value")
			cpuMetrics.UsedCpu = fmt.Sprintf("%.1f", usedCpu)

			containerCPUMetrics = append(containerCPUMetrics, cpuMetrics)
		}
	}

	resultContainer.Cpu = containerCPUMetrics

	memoryUsage := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + podName + "/containers/" + containerName + "/metrics/memory/usage")
	memoryUsageMetrics, err := memoryUsage.GetObjectArray("metrics")
	if err != nil {
		glog.Error(err)
		containerMemMetrics = make([]ContainerMemoryMetrics, 0)
	} else {
		for _, metric := range memoryUsageMetrics {
			timestamp, _ := metric.GetString("timestamp")
			memMetrics.TimeStamp = timestamp
			usedMemoryBytes, _ := metric.GetFloat64("value")
			memMetrics.UsedMemory = fmt.Sprintf("%.1f", usedMemoryBytes/1024/1024)
			containerMemMetrics = append(containerMemMetrics, memMetrics)
		}
	}

	resultContainer.Memory = containerMemMetrics

	return resultContainer
}
