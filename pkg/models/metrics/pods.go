package metrics

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	coreV1 "k8s.io/api/core/v1"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
)

type PodMetrics struct {
	PodName       string             `json:"pod_name"`
	NameSpace     string             `json:"namespace"`
	NodeName      string             `json:"node_name"`
	CPURequest    string             `json:"cpu_request"`
	CPULimit      string             `json:"cpu_limit"`
	MemoryRequest string             `json:"mem_request"`
	MemoryLimit   string             `json:"mem_limit"`
	CPU           []PodCpuMetrics    `json:"cpu"`
	Memory        []PodMemoryMetrics `json:"memory"`
}
type PodCpuMetrics struct {
	TimeStamp string `json:"timestamp"`
	UsedCpu   string `json:"used_cpu"`
}

type PodMemoryMetrics struct {
	TimeStamp  string `json:"timestamp"`
	UsedMemory string `json:"used_mem"`
}

const Inf = "inf"

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

func GetSinglePodMetrics(namespace string, podName string, ch chan PodMetrics) {
	podMetrics := FormatPodMetrics(namespace, podName)

	ch <- podMetrics
}

func GetPodsMetrics(podList *coreV1.PodList) []PodMetrics {
	items := make([]PodMetrics, 0)

	ch := make(chan PodMetrics)
	for _, pod := range podList.Items {
		go GetSinglePodMetrics(pod.Namespace, pod.Name, ch)
	}

	for range podList.Items {
		items = append(items, <-ch)
	}

	return items
}

func GetPodMetricsInNamespace(namespace string, labelSelector string) constants.PageableResponse {

	var podMetrics constants.PageableResponse
	k8sClient := client.NewK8sClient()
	var options v1.ListOptions
	if len(labelSelector) > 0 {
		options = v1.ListOptions{
			LabelSelector: labelSelector,
		}
	} else {
		options = v1.ListOptions{}
	}
	podList, _ := k8sClient.CoreV1().Pods(namespace).List(options)

	for _, podMetric := range GetPodsMetrics(podList) {
		podMetrics.Items = append(podMetrics.Items, podMetric)
	}
	podMetrics.TotalCount = len(podMetrics.Items)

	return podMetrics
}

func GetPodMetricsInDeployment(namespace string, deployment string) constants.PageableResponse {

	var podMetrics constants.PageableResponse

	k8sClient := client.NewK8sClient()

	deploy, err := k8sClient.ExtensionsV1beta1().Deployments(namespace).Get(deployment, v1.GetOptions{})
	if err != nil {
		glog.Error(err)
		podMetrics.Items = make([]interface{}, 0)
		podMetrics.TotalCount = len(podMetrics.Items)
		return podMetrics
	}

	labels := make([]string, 0)
	for k, v := range deploy.Spec.Selector.MatchLabels {
		labels = append(labels, k+"="+v)
	}

	options := v1.ListOptions{
		LabelSelector: strings.Join(labels, ","),
	}

	podList, err := k8sClient.CoreV1().Pods(namespace).List(options)
	for _, podMetric := range GetPodsMetrics(podList) {
		podMetrics.Items = append(podMetrics.Items, podMetric)
	}
	podMetrics.TotalCount = len(podMetrics.Items)

	return podMetrics

}

func GetPodMetricsInStatefulSet(namespace string, statefulSet string) constants.PageableResponse {

	var podMetrics constants.PageableResponse

	k8sClient := client.NewK8sClient()

	deploy, err := k8sClient.AppsV1().StatefulSets(namespace).Get(statefulSet, v1.GetOptions{})
	if err != nil {
		glog.Error(err)
		podMetrics.Items = make([]interface{}, 0)
		podMetrics.TotalCount = len(podMetrics.Items)
		return podMetrics
	}

	labels := make([]string, 0)
	for k, v := range deploy.Spec.Selector.MatchLabels {
		labels = append(labels, k+"="+v)
	}

	options := v1.ListOptions{
		LabelSelector: strings.Join(labels, ","),
	}

	podList, err := k8sClient.CoreV1().Pods(namespace).List(options)
	for _, podMetric := range GetPodsMetrics(podList) {
		podMetrics.Items = append(podMetrics.Items, podMetric)
	}
	podMetrics.TotalCount = len(podMetrics.Items)

	return podMetrics

}

func GetPodMetricsInDaemonset(namespace string, daemonset string) constants.PageableResponse {

	var podMetrics constants.PageableResponse

	k8sClient := client.NewK8sClient()

	deploy, err := k8sClient.ExtensionsV1beta1().DaemonSets(namespace).Get(daemonset, v1.GetOptions{})
	if err != nil {
		glog.Error(err)
		podMetrics.Items = make([]interface{}, 0)
		podMetrics.TotalCount = len(podMetrics.Items)
		return podMetrics
	}

	labels := make([]string, 0)
	for k, v := range deploy.Spec.Selector.MatchLabels {
		labels = append(labels, k+"="+v)
	}

	options := v1.ListOptions{
		LabelSelector: strings.Join(labels, ","),
	}

	podList, err := k8sClient.CoreV1().Pods(namespace).List(options)
	for _, podMetric := range GetPodsMetrics(podList) {
		podMetrics.Items = append(podMetrics.Items, podMetric)
	}
	podMetrics.TotalCount = len(podMetrics.Items)

	return podMetrics

}

func GetPodMetricsInNode(nodeName string) constants.PageableResponse {
	var podMetrics constants.PageableResponse
	k8sClient := client.NewK8sClient()
	options := v1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	}
	podList, _ := k8sClient.CoreV1().Pods("").List(options)
	for _, podMetric := range GetPodsMetrics(podList) {
		podMetrics.Items = append(podMetrics.Items, podMetric)
	}
	podMetrics.TotalCount = len(podMetrics.Items)

	return podMetrics
}

func GetPodMetricsInNamespaceOfNode(namespace string, nodeName string) constants.PageableResponse {
	var podMetrics constants.PageableResponse
	k8sClient := client.NewK8sClient()
	options := v1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName + ",metadata.namespace=" + namespace,
	}
	podList, _ := k8sClient.CoreV1().Pods("").List(options)
	for _, podMetric := range GetPodsMetrics(podList) {
		podMetrics.Items = append(podMetrics.Items, podMetric)
	}
	podMetrics.TotalCount = len(podMetrics.Items)

	return podMetrics
}

func GetAllPodMetrics() constants.PageableResponse {
	var podMetrics constants.PageableResponse
	k8sClient := client.NewK8sClient()
	options := v1.ListOptions{}
	podList, _ := k8sClient.CoreV1().Pods("").List(options)
	for _, podMetric := range GetPodsMetrics(podList) {
		podMetrics.Items = append(podMetrics.Items, podMetric)
	}
	podMetrics.TotalCount = len(podMetrics.Items)

	return podMetrics
}

func FormatResourceLimit(limit float64) string {
	if limit <= 0 {
		return Inf
	} else {
		return fmt.Sprintf("%.1f", limit)
	}
}

func FormatPodMetrics(namespace string, pod string) PodMetrics {

	var resultPod PodMetrics
	var podCPUMetrics []PodCpuMetrics
	var podMemMetrics []PodMemoryMetrics
	var cpuMetrics PodCpuMetrics
	var memoryMetrics PodMemoryMetrics

	resultPod.PodName = pod
	resultPod.NameSpace = namespace
	cpuRequest := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + pod + "/metrics/cpu/request")
	cpuRequestMetrics, err := cpuRequest.GetObjectArray("metrics")
	if err != nil {
		glog.Error(err)
	} else {
		if len(cpuRequestMetrics) == 0 {
			resultPod.CPURequest = Inf
		} else {
			data, err := cpuRequestMetrics[0].GetNumber("value")
			if err != nil {
				glog.Error(err)
			}
			resultPod.CPURequest = data.String()
		}
	}

	cpuLimit := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + pod + "/metrics/cpu/limit")

	cpuLimitMetrics, err := cpuLimit.GetObjectArray("metrics")
	if len(cpuLimitMetrics) == 0 {
		resultPod.CPULimit = Inf
	} else {
		data, _ := cpuLimitMetrics[0].GetFloat64("value")
		resultPod.CPULimit = FormatResourceLimit(data)
	}

	memoryRequest := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + pod + "/metrics/memory/request")
	memoryRequestMetrics, err := memoryRequest.GetObjectArray("metrics")
	if err != nil {
		glog.Error(err)
	}

	if len(memoryRequestMetrics) == 0 {
		resultPod.MemoryRequest = Inf
	} else {
		data, _ := memoryRequestMetrics[0].GetNumber("value")
		memoryReq, _ := data.Float64()
		memoryReq = memoryReq / 1024 / 1024
		resultPod.MemoryRequest = FormatResourceLimit(memoryReq)
	}

	memoryLimit := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + pod + "/metrics/memory/limit")
	memoryLimitMetrics, err := memoryLimit.GetObjectArray("metrics")
	if err != nil || len(memoryLimitMetrics) == 0 {
		glog.Error(err)
		resultPod.MemoryLimit = Inf
	} else {
		data, _ := memoryLimitMetrics[0].GetNumber("value")
		memoryLim, _ := data.Float64()
		memoryLim = memoryLim / 1024 / 1024
		resultPod.MemoryLimit = FormatResourceLimit(memoryLim)
	}

	cpuUsageRate := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + pod + "/metrics/cpu/usage_rate")
	cpuUsageRateMetrics, err := cpuUsageRate.GetObjectArray("metrics")

	if err != nil {
		glog.Error(err)
		resultPod.CPU = make([]PodCpuMetrics, 0)
	} else {
		for _, cpuUsageRateMetric := range cpuUsageRateMetrics {
			timestamp, _ := cpuUsageRateMetric.GetString("timestamp")
			cpuUsageRate, _ := cpuUsageRateMetric.GetFloat64("value")
			cpuMetrics.TimeStamp = timestamp
			cpuMetrics.UsedCpu = fmt.Sprintf("%.1f", cpuUsageRate)

			podCPUMetrics = append(podCPUMetrics, cpuMetrics)
		}

		resultPod.CPU = podCPUMetrics

	}

	memUsage := client.GetHeapsterMetricsJson("/namespaces/" + namespace + "/pods/" + pod + "/metrics/memory/usage")
	memoryUsageMetrics, err := memUsage.GetObjectArray("metrics")
	if err != nil {
		glog.Error(err)
		resultPod.Memory = make([]PodMemoryMetrics, 0)
	} else {
		for _, memoryUsageMetric := range memoryUsageMetrics {
			timestamp, _ := memoryUsageMetric.GetString("timestamp")
			memoryMetrics.TimeStamp = timestamp
			usedMemoryBytes, err := memoryUsageMetric.GetFloat64("value")
			if err == nil {
				memoryMetrics.UsedMemory = fmt.Sprintf("%.1f", usedMemoryBytes/1024/1024)
			} else {
				memoryMetrics.UsedMemory = Inf
			}
			podMemMetrics = append(podMemMetrics, memoryMetrics)
		}
		resultPod.Memory = podMemMetrics
	}

	return resultPod
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
