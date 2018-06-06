package models

import (
	"encoding/json"

	"strings"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"

	ksutil "kubesphere.io/kubesphere/pkg/util"

	"fmt"
	"strconv"

	"math"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/constants"
)

type ResultPod struct {
	PodName       string      `json:"pod_name"`
	NameSpace     string      `json:"namespace"`
	NodeName      string      `json:"node_name"`
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
	podsList := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods")
	var pods []string
	dec := json.NewDecoder(strings.NewReader(podsList))
	err := dec.Decode(&pods)
	if err != nil {
		glog.Error(err)
	}
	return pods
}

func GetPodsForNode(nodeName, namespace string) []string {
	var pods []string
	cli := client.NewK8sClient()
	podList, err := cli.CoreV1().Pods(namespace).List(v1.ListOptions{FieldSelector: "spec.nodeName=" + nodeName})
	if err != nil {
		glog.Error(err)
	} else {
		for _, pod := range podList.Items {
			pods = append(pods, pod.Name)
		}
	}
	return pods
}

func FormatPodsMetrics(nodeName, namespace string) constants.PageableResponse {
	var result constants.PageableResponse

	var resultPod ResultPod
	var pods []string
	if nodeName == "" {
		pods = GetPods(namespace)
	} else {
		pods = GetPodsForNode(nodeName, namespace)
	}

	var total_count int
	for i, pod := range pods {
		resultPod = FormatPodMetrics(namespace, pod)
		if nodeName != "" {
			resultPod.NodeName = nodeName
		} else {
			resultPod.NodeName = GetNodeNameForPod(pod, namespace)
		}
		result.Items = append(result.Items, resultPod)
		total_count = i
	}
	result.TotalCount = total_count + 1
	return result
}

func FormatPodMetrics(namespace, pod string) ResultPod {

	var resultPod ResultPod
	var podCPUMetrics []CPUPod
	var podMemMetrics []MemoryPod
	var cpuMetrics CPUPod
	var memMetrics MemoryPod

	resultPod.PodName = pod
	resultPod.NameSpace = namespace
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
	memoryRequest := ksutil.JsonRawMessage(client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/metrics/memory/request")).Find("metrics").ToList()[0].Find("value").ToString()
	resultPod.MemoryRequest = ConvertMemory(memoryRequest)

	memoryLimit := ksutil.JsonRawMessage(client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/metrics/memory/limit")).Find("metrics").ToList()[0].Find("value").ToString()
	resultPod.MemoryLimit = ConvertMemory(memoryLimit)

	cpuUsageRate := client.GetHeapsterMetrics("/namespaces/" + namespace + "/pods/" + pod + "/metrics/cpu/usage_rate")
	if cpuUsageRate != "" {
		metrics := ksutil.JsonRawMessage(cpuUsageRate).Find("metrics").ToList()

		for _, metric := range metrics {
			timestamp := metric.Find("timestamp")
			cpu_utilization, _ := strconv.ParseFloat(ConvertCPUUsageRate(metric.Find("value").ToString()), 64)
			cpuMetrics.TimeStamp = timestamp.ToString()
			cpuMetrics.CPUUtilization = fmt.Sprintf("%.3f", cpu_utilization)
			if resultPod.CPULimit != "inf" {
				cpu_limit, _ := strconv.ParseFloat(resultPod.CPULimit, 64)
				cpuMetrics.UsedCPU = fmt.Sprintf("%.1f", cpu_limit*cpu_utilization)
			} else {
				cpuMetrics.UsedCPU = "inf"
			}
			glog.Info("pod " + pod + " has limit cpu " + resultPod.CPULimit + " CPU utilization " + fmt.Sprintf("%.3f", cpu_utilization) + " at time" + timestamp.ToString())
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
			memMetrics.MemoryUtilization = fmt.Sprintf("%.3f", CalculateMemoryUsage(memoryRequest, memoryLimit, metric.Find("value").ToString()))

			glog.Info("pod " + pod + " has limit mem " + resultPod.MemoryLimit + " mem utilization " + memMetrics.MemoryUtilization + " at time" + timestamp.ToString())
			podMemMetrics = append(podMemMetrics, memMetrics)
		}
	}

	resultPod.Memory = podMemMetrics
	return resultPod
}

func ConvertMemory(memBytes string) string {
	var mem string

	if memBytes != "" {
		if memBytes != "" && memBytes != "0" {
			memBytes, error := strconv.ParseFloat(memBytes, 64)
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

func CalculateMemoryUsage(requestMem, limitMem, usedMem string) float64 {
	var requestMemInBytes, limitMemInBytes, usedMemInBytes, memUsage float64
	if requestMem != "" && requestMem != "0" && requestMem != "inf" {
		requestMemInBytes, _ = strconv.ParseFloat(requestMem, 64)
	} else {
		glog.Info("memory request is not set")
		requestMemInBytes = 0
	}
	if limitMem != "" && limitMem != "0" && limitMem != "inf" {
		limitMemInBytes, _ = strconv.ParseFloat(limitMem, 64)
	} else {
		glog.Info("memory limit is not set")
		limitMemInBytes = 0
	}
	if usedMem != "" && usedMem != "0" {
		usedMemInBytes, _ = strconv.ParseFloat(usedMem, 64)
	} else {
		usedMemInBytes = 0
	}

	if usedMemInBytes > 0 {
		if requestMemInBytes > 0 && limitMemInBytes > 0 {
			if usedMemInBytes > requestMemInBytes {
				glog.Info("used memory is higher than memory request")
				memUsage = usedMemInBytes / limitMemInBytes
			} else {
				memUsage = usedMemInBytes / requestMemInBytes
			}
		} else if requestMemInBytes > 0 && limitMemInBytes == 0 {
			if usedMemInBytes > requestMemInBytes {
				glog.Info("used memory is higher than memory request")
				memUsage = 0
			} else {
				memUsage = usedMemInBytes / requestMemInBytes
			}
		} else if requestMemInBytes == 0 && limitMemInBytes > 0 {
			if usedMemInBytes <= limitMemInBytes {
				memUsage = usedMemInBytes / limitMemInBytes
			}
		} else {
			memUsage = 0
		}
	} else {
		memUsage = 0
	}
	return memUsage
}

func ConvertCPUUsageRate(cpuUsageRate string) string {
	if cpuUsageRate != "" && cpuUsageRate != "0" {
		rate, _ := strconv.ParseFloat(cpuUsageRate, 64)
		rateBase := math.Pow10(strings.Count(cpuUsageRate, "") - 1)
		return fmt.Sprintf("%.3f", rate/rateBase)
	} else {
		return "0"
	}
}

func GetNodeNameForPod(podName, namespace string) string {
	var nodeName string
	cli := client.NewK8sClient()

	pod, err := cli.CoreV1().Pods(namespace).Get(podName, v1.GetOptions{})

	if err != nil {
		glog.Error(err)
	} else {
		nodeName = pod.Spec.NodeName
	}
	return nodeName
}
