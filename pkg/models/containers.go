package models

import (
	"encoding/json"

	"strings"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"

	ksutil "kubesphere.io/kubesphere/pkg/util"

	"fmt"
	"strconv"
)

type ResultNameSpaceForContainer struct {
	NameSpace string                  `json:"namespace"`
	PodsCount string                  `json:"pods_count"`
	Pods      []ResultPodForContainer `json:"pods"`
}

type ResultPodForContainer struct {
	PodName         string            `json:"pod_name"`
	ContainersCount string            `json:"containers_count"`
	Containers      []ResultContainer `json:"containers"`
}
type ResultContainer struct {
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

func FormatContainersMetrics(nodeName, namespace, podName string) ResultNameSpaceForContainer {
	var resultNameSpaceForContainer ResultNameSpaceForContainer
	var resultPodsForContainer []ResultPodForContainer

	var pods []string
	if nodeName == "" {
		pods = GetPods(namespace)
	} else {
		pods = GetPodsForNode(nodeName, namespace)
	}

	resultNameSpaceForContainer.NameSpace = namespace
	resultNameSpaceForContainer.PodsCount = strconv.Itoa(len(pods))

	if podName != "" {
		var resultPodForContainer ResultPodForContainer
		resultPodForContainer.PodName = podName
		resultPodForContainer = FormatPodMetricsWithContainers(namespace, podName)
		resultPodsForContainer = append(resultPodsForContainer, resultPodForContainer)
		resultNameSpaceForContainer.Pods = resultPodsForContainer
		return resultNameSpaceForContainer
	}
	for _, pod := range pods {
		var resultPodForContainer ResultPodForContainer
		resultPodForContainer.PodName = pod
		resultPodForContainer = FormatPodMetricsWithContainers(namespace, pod)
		resultPodsForContainer = append(resultPodsForContainer, resultPodForContainer)
	}
	resultNameSpaceForContainer.Pods = resultPodsForContainer
	return resultNameSpaceForContainer
}

func FormatPodMetricsWithContainers(namespace, pod string) ResultPodForContainer {

	var resultPod ResultPodForContainer
	var containers []string
	var resultContainers []ResultContainer

	resultPod.PodName = pod
	containers = GetContainers(namespace, pod)
	resultPod.ContainersCount = strconv.Itoa(len(containers))

	for _, container := range containers {
		var resultContainer ResultContainer
		var containerCPUMetrics []CPUContainer
		var containerMemMetrics []MemoryContainer
		var cpuMetrics CPUContainer
		var memMetrics MemoryContainer

		resultContainer.ContainerName = container
		cpuRequest := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/containers/" + container + "/metrics/cpu/request")
		cpuRequest = ksutil.JsonRawMessage(cpuRequest).Find("metrics").ToList()[0].Find("value").ToString()
		if cpuRequest != "" && cpuRequest != "0" {
			resultContainer.CPURequest = cpuRequest
		} else {
			resultContainer.CPURequest = "inf"
		}

		cpuLimit := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/containers/" + container + "/metrics/cpu/limit")
		cpuLimit = ksutil.JsonRawMessage(cpuLimit).Find("metrics").ToList()[0].Find("value").ToString()
		if cpuLimit != "" && cpuLimit != "0" {
			resultContainer.CPULimit = cpuLimit
		} else {
			resultContainer.CPULimit = "inf"
		}
		memoryRequest := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/containers/" + container + "/metrics/memory/request")
		resultContainer.MemoryRequest = ConvertMemory(memoryRequest)

		memoryLimit := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/containers/" + container + "/metrics/memory/limit")
		resultContainer.MemoryLimit = ConvertMemory(memoryLimit)

		cpuUsageRate := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/containers/" + container + "/metrics/cpu/usage_rate")
		if cpuUsageRate != "" {
			metrics := ksutil.JsonRawMessage(cpuUsageRate).Find("metrics").ToList()

			for _, metric := range metrics {
				timestamp := metric.Find("timestamp")
				cpu_utilization, _ := strconv.ParseFloat(metric.Find("value").ToString(), 64)
				cpuMetrics.TimeStamp = timestamp.ToString()
				cpuMetrics.CPUUtilization = fmt.Sprintf("%.3f", cpu_utilization/1000)
				if resultContainer.CPULimit != "inf" {
					cpu_limit, _ := strconv.ParseFloat(resultContainer.CPULimit, 64)
					cpuMetrics.UsedCPU = fmt.Sprintf("%.1f", cpu_limit*cpu_utilization/1000)
				} else {
					cpuMetrics.UsedCPU = "inf"
				}
				glog.Info("pod " + pod + " has limit cpu " + resultContainer.CPULimit + " CPU utilization " + fmt.Sprintf("%.3f", cpu_utilization/1000) + " at time" + timestamp.ToString())
				containerCPUMetrics = append(containerCPUMetrics, cpuMetrics)
			}

		}

		resultContainer.CPU = containerCPUMetrics

		var used_mem_bytes float64

		memUsage := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/containers/" + container + "/metrics/memory/usage")
		if memUsage != "" {
			metrics := ksutil.JsonRawMessage(memUsage).Find("metrics").ToList()

			for _, metric := range metrics {
				timestamp := metric.Find("timestamp")
				used_mem_bytes, _ = strconv.ParseFloat(metric.Find("value").ToString(), 64)
				used_mem := used_mem_bytes / 1024 / 1024
				memMetrics.TimeStamp = timestamp.ToString()
				memMetrics.UsedMemory = fmt.Sprintf("%.1f", used_mem)
				if resultContainer.MemoryLimit != "inf" {
					mem_limit, _ := strconv.ParseFloat(resultContainer.MemoryLimit, 64)
					memMetrics.MemoryUtilization = fmt.Sprintf("%.3f", used_mem/mem_limit)
				} else {
					memMetrics.MemoryUtilization = "inf"
				}

				glog.Info("pod " + pod + " has limit mem " + resultContainer.MemoryLimit + " mem utilization " + memMetrics.MemoryUtilization + " at time" + timestamp.ToString())
				containerMemMetrics = append(containerMemMetrics, memMetrics)
			}
		}

		resultContainer.Memory = containerMemMetrics
		resultContainers = append(resultContainers, resultContainer)
	}
	resultPod.Containers = resultContainers

	return resultPod
}
