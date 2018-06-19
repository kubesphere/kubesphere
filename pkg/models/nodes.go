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
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

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

const GracePeriods = 900

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
	node, err := k8sclient.CoreV1().Nodes().Get(nodename, metav1.GetOptions{})
	if err != nil {
		glog.Fatal(err)
		return msg, err
	}

	if node.Spec.Unschedulable {
		glog.Info(node.Spec.Unschedulable)
		msg.Message = fmt.Sprintf("node %s have been drained", nodename)
		return msg, nil
	}

	data := []byte(" {\"spec\":{\"unschedulable\":true}}")
	nodestatus, err := k8sclient.CoreV1().Nodes().Patch(nodename, types.StrategicMergePatchType, data)
	glog.Info(nodestatus)

	if err != nil {
		glog.Fatal(err)
		return msg, err
	}
	msg.Message = "success"
	return msg, nil
}

func DrainStatus(nodename string) (msg constants.MessageResponse, err error) {

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
	// remove mirror pod static pod
	if len(podList.Items) > 0 {

		for _, pod := range podList.Items {

			if !containDaemonset(pod, *daemonsetList) {
				//static or mirror pod
				if isStaticPod(&pod) || isMirrorPod(&pod) {

					continue

				} else {

					pods = append(pods, pod)

				}

			}

		}
	}
	if len(pods) == 0 {

		msg.Message = fmt.Sprintf("success")
		return msg, nil

	} else {

		//create eviction
		getPodFn := func(namespace, name string) (*v1.Pod, error) {
			k8sclient := kubeclient.NewK8sClient()
			return k8sclient.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		}
		evicerr := evictPods(pods, 900, getPodFn)

		if evicerr == nil {

			msg.Message = fmt.Sprintf("success")
			return msg, nil

		} else {

			glog.Info(evicerr)
			msg.Message = evicerr.Error()
			return msg, nil
		}

	}

}

func getPodSource(pod *v1.Pod) (string, error) {
	if pod.Annotations != nil {
		if source, ok := pod.Annotations["kubernetes.io/config.source"]; ok {
			return source, nil
		}
	}
	return "", fmt.Errorf("cannot get source of pod %q", pod.UID)
}

func isStaticPod(pod *v1.Pod) bool {
	source, err := getPodSource(pod)
	return err == nil && source != "api"
}

func isMirrorPod(pod *v1.Pod) bool {
	_, ok := pod.Annotations[v1.MirrorPodAnnotationKey]
	return ok
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

func evictPod(pod v1.Pod, GracePeriodSeconds int) error {

	k8sclient := kubeclient.NewK8sClient()
	deleteOptions := &metav1.DeleteOptions{}
	if GracePeriodSeconds >= 0 {
		gracePeriodSeconds := int64(GracePeriodSeconds)
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	}

	var eviction policy.Eviction
	eviction.Kind = "Eviction"
	eviction.APIVersion = "policy/v1beta1"
	eviction.Namespace = pod.Namespace
	eviction.Name = pod.Name
	eviction.DeleteOptions = deleteOptions
	err := k8sclient.CoreV1().Pods(pod.Namespace).Evict(&eviction)
	if err != nil {
		return err
	}

	return nil
}

func evictPods(pods []v1.Pod, GracePeriodSeconds int, getPodFn func(namespace, name string) (*v1.Pod, error)) error {
	doneCh := make(chan bool, len(pods))
	errCh := make(chan error, 1)

	for _, pod := range pods {
		go func(pod v1.Pod, doneCh chan bool, errCh chan error) {
			var err error
			for {
				err = evictPod(pod, GracePeriodSeconds)
				if err == nil {
					break
				} else if apierrors.IsNotFound(err) {
					doneCh <- true
					glog.Info(fmt.Sprintf("pod %s evict", pod.Name))
					return
				} else if apierrors.IsTooManyRequests(err) {
					time.Sleep(5 * time.Second)
				} else {
					errCh <- fmt.Errorf("error when evicting pod %q: %v", pod.Name, err)
					return
				}
			}

			podArray := []v1.Pod{pod}
			_, err = waitForDelete(podArray, time.Second, time.Duration(math.MaxInt64), getPodFn)
			if err == nil {
				doneCh <- true
				glog.Info(fmt.Sprintf("pod %s delete", pod.Name))
			} else {
				errCh <- fmt.Errorf("error when waiting for pod %q terminating: %v", pod.Name, err)
			}
		}(pod, doneCh, errCh)
	}

	Timeout := GracePeriods * power(10, 9)
	doneCount := 0
	// 0 timeout means infinite, we use MaxInt64 to represent it.
	var globalTimeout time.Duration
	if Timeout == 0 {
		globalTimeout = time.Duration(math.MaxInt64)
	} else {
		globalTimeout = time.Duration(Timeout)
	}
	for {
		select {
		case err := <-errCh:
			return err
		case <-doneCh:
			doneCount++
			if doneCount == len(pods) {
				return nil
			}
		case <-time.After(globalTimeout):
			return fmt.Errorf("Drain did not complete within %v, please check node status in a few minutes", globalTimeout)
		}
	}
}

func waitForDelete(pods []v1.Pod, interval, timeout time.Duration, getPodFn func(string, string) (*v1.Pod, error)) ([]v1.Pod, error) {

	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		pendingPods := []v1.Pod{}
		for i, pod := range pods {
			p, err := getPodFn(pod.Namespace, pod.Name)
			if apierrors.IsNotFound(err) || (p != nil && p.ObjectMeta.UID != pod.ObjectMeta.UID) {
				continue
			} else if err != nil {
				return false, err
			} else {
				pendingPods = append(pendingPods, pods[i])
			}
		}
		pods = pendingPods
		if len(pendingPods) > 0 {
			return false, nil
		}
		return true, nil
	})
	return pods, err
}

func power(x int64, n int) int64 {

	var res int64 = 1
	for n != 0 {
		res *= x
		n--
	}

	return res

}
