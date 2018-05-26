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

type ResultNameSpaces struct {
	NameSpaces []ResultNameSpace `json:"namespaces"`
}
type ResultNameSpace struct {
	NameSpace string      `json:"namespace"`
	PodsCount string      `json:"pods_count"`
	Pods      []ResultPod `json:"pods"`
}
type ResultPod struct {
	PodName       string      `json:"pod_name"`
	CPURequest    string      `json:"cpu_request"`
	CPULimit      string      `json:"cpu_limit"`
	MemoryRequest string      `json:"mem_request"`
	MemoryLimit   string      `json:"mem_limit"`
	CPU           []CPUPod    `json:"cpu"`
	Memory        []MemoryPod `json:"memory"`
}
type CPUPod struct {
	TimeStamp      string `json:"timestamp"`
	UsedCPU        string `json:"used_cpu"`
	CPUUtilization string `json:"cpu_utilization"`
}

type MemoryPod struct {
	TimeStamp         string `json:"timestamp"`
	UsedMemory        string `json:"used_mem"`
	MemoryUtilization string `json:"mem_utilization"`
}

/*
Get all namespaces in default cluster
*/
func GetNameSpaces() []string {
	namespacesList := client.GetHeapsterMetrics("/namespaces")
	var namespaces []string
	dec := json.NewDecoder(strings.NewReader(namespacesList))
	err := dec.Decode(&namespaces)
	if err != nil {
		glog.Error(err)
	}
	return namespaces
}

/*
Get all pods under specified namespace in default cluster
*/
func GetPods(namespace string) []string {
	fmt.Println(namespace)
	podsList := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods")
	var pods []string
	dec := json.NewDecoder(strings.NewReader(podsList))
	err := dec.Decode(&pods)
	if err != nil {
		glog.Error(err)
	}
	return pods
}

func FormatNameSpaceMetrics(namespace string) ResultNameSpace {
	var resultNameSpace ResultNameSpace
	var resultPods []ResultPod
	var resultPod ResultPod

	pods := GetPods(namespace)

	resultNameSpace.NameSpace = namespace
	resultNameSpace.PodsCount = strconv.Itoa(len(pods))

	for _, pod := range pods {
		resultPod = FormatPodMetrics(namespace, pod)
		resultPods = append(resultPods, resultPod)
	}
	resultNameSpace.Pods = resultPods
	return resultNameSpace
}

func FormatPodMetrics(namespace, pod string) ResultPod {
	var resultPod ResultPod
	var podCPUMetrics []CPUPod
	var podMemMetrics []MemoryPod
	var cpuMetrics CPUPod
	var memMetrics MemoryPod

	resultPod.PodName = pod
	cpuRequest := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/metrics/cpu/request")
	cpuRequest = ksutil.JsonRawMessage(cpuRequest).Find("metrics").ToList()[0].Find("value").ToString()
	if cpuRequest != "" && cpuRequest != "0" {
		resultPod.CPURequest = cpuRequest
	} else {
		resultPod.CPURequest = "inf"
	}

	cpuLimit := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/metrics/cpu/limit")
	cpuLimit = ksutil.JsonRawMessage(cpuLimit).Find("metrics").ToList()[0].Find("value").ToString()
	if cpuLimit != "" && cpuLimit != "0" {
		resultPod.CPULimit = cpuLimit
	} else {
		resultPod.CPULimit = "inf"
	}
	memoryRequest := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/metrics/memory/request")
	resultPod.MemoryRequest = convertMemory(memoryRequest)

	memoryLimit := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/metrics/memory/limit")
	resultPod.MemoryLimit = convertMemory(memoryLimit)

	cpuUsageRate := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/metrics/cpu/usage_rate")
	if cpuUsageRate != "" {
		metrics := ksutil.JsonRawMessage(cpuUsageRate).Find("metrics").ToList()

		for _, metric := range metrics {
			timestamp := metric.Find("timestamp")
			cpu_utilization, _ := strconv.ParseFloat(metric.Find("value").ToString(), 64)
			cpuMetrics.TimeStamp = timestamp.ToString()
			cpuMetrics.CPUUtilization = fmt.Sprintf("%.3f", cpu_utilization/1000)
			if resultPod.CPULimit != "inf" {
				cpu_limit, _ := strconv.ParseFloat(resultPod.CPULimit, 64)
				cpuMetrics.UsedCPU = fmt.Sprintf("%.1f", cpu_limit*cpu_utilization/1000)
			} else {
				cpuMetrics.UsedCPU = "inf"
			}
			glog.Info("pod " + pod + " has limit cpu " + resultPod.CPULimit + " CPU utilization " + fmt.Sprintf("%.3f", cpu_utilization/1000) + " at time" + timestamp.ToString())
			podCPUMetrics = append(podCPUMetrics, cpuMetrics)
		}

	}

	resultPod.CPU = podCPUMetrics

	var used_mem_bytes float64

	memUsage := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/metrics/memory/usage")
	if memUsage != "" {
		metrics := ksutil.JsonRawMessage(memUsage).Find("metrics").ToList()

		for _, metric := range metrics {
			timestamp := metric.Find("timestamp")
			used_mem_bytes, _ = strconv.ParseFloat(metric.Find("value").ToString(), 64)
			used_mem := used_mem_bytes / 1024 / 1024
			memMetrics.TimeStamp = timestamp.ToString()
			memMetrics.UsedMemory = fmt.Sprintf("%.1f", used_mem)
			if resultPod.MemoryLimit != "inf" {
				mem_limit, _ := strconv.ParseFloat(resultPod.MemoryLimit, 64)
				memMetrics.MemoryUtilization = fmt.Sprintf("%.3f", used_mem/mem_limit)
			} else {
				memMetrics.MemoryUtilization = "inf"
			}

			glog.Info("pod " + pod + " has limit mem " + resultPod.MemoryLimit + " mem utilization " + memMetrics.MemoryUtilization + " at time" + timestamp.ToString())
			podMemMetrics = append(podMemMetrics, memMetrics)
		}
	}

	resultPod.Memory = podMemMetrics
	return resultPod
}

func convertMemory(memBytes string) string {
	var mem string

	if memBytes != "" {
		memMetric := ksutil.JsonRawMessage(memBytes).Find("metrics").ToList()[0].Find("value").ToString()

		if memMetric != "" && memMetric != "0" {

			memBytes, error := strconv.ParseFloat(memMetric, 64)
			if error == nil {
				mem = fmt.Sprintf("%.3f", memBytes/1024/1024)
			} else {
				mem = "inf"
			}
		} else {
			mem = "inf"
		}

	} else {
		mem = "inf"
	}
	return mem
}
