/*
Copyright 2020 KubeSphere Authors

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

package metricsserver

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	promlabels "github.com/prometheus/prometheus/model/labels"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	metricsV1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"

	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

// metricsServer implements monitoring interface backend by metrics-server
type metricsServer struct {
	metricsAPIAvailable bool
	metricsClient       metricsclient.Interface
	k8s                 kubernetes.Interface
}

var (
	supportedMetricsAPIs = map[string]bool{
		"v1beta1": true,
	}
)

const edgeNodeLabel = "node-role.kubernetes.io/edge"

func metricsAPISupported(discoveredAPIGroups *metav1.APIGroupList) bool {
	for _, discoveredAPIGroup := range discoveredAPIGroups.Groups {
		if discoveredAPIGroup.Name != metricsapi.GroupName {
			continue
		}
		for _, version := range discoveredAPIGroup.Versions {
			if _, found := supportedMetricsAPIs[version.Version]; found {
				return true
			}
		}
	}
	return false
}

func (m metricsServer) listEdgeNodes() (map[string]v1.Node, error) {
	nodes := make(map[string]v1.Node)

	nodeClient := m.k8s.CoreV1()

	nodeList, err := nodeClient.Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: edgeNodeLabel,
	})
	if err != nil {
		return nodes, err
	}

	for _, n := range nodeList.Items {
		nodes[n.Name] = n
	}

	return nodes, nil
}

func (m metricsServer) filterEdgeNodeNames(edgeNodes map[string]v1.Node, opts *monitoring.QueryOptions) map[string]bool {
	edgeNodeNamesFiltered := make(map[string]bool)

	regexMatcher, err := promlabels.NewMatcher(promlabels.MatchRegexp, "edgenodefilter", opts.ResourceFilter)

	if err != nil {
		klog.Errorf("Edge node filter regexp error %v\n", err)
		return edgeNodeNamesFiltered
	}

	for _, n := range edgeNodes {
		if regexMatcher.Matches(n.Name) {
			edgeNodeNamesFiltered[n.Name] = true
		}
	}

	return edgeNodeNamesFiltered
}

func (m metricsServer) parseEdgePods(opts *monitoring.QueryOptions) map[string]bool {

	edgePods := make(map[string]bool)

	var filters []string

	r, _ := regexp.Compile(`\s*\|\s*|\$`)
	if opts.ResourceFilter != "" {
		filters = r.Split(opts.ResourceFilter, -1)
	}

	if opts.NamespacedResourcesFilter != "" {
		filters = r.Split(opts.NamespacedResourcesFilter, -1)
	}

	for _, p := range filters {
		if p == "" || p == ".*" {
			continue
		}
		edgePods[p] = true
	}

	return edgePods
}

// node metrics of edge nodes
func (m metricsServer) getNodeMetricsFromMetricsAPI() (*metricsapi.NodeMetricsList, error) {
	var err error
	mc := m.metricsClient.MetricsV1beta1()
	nm := mc.NodeMetricses()
	versionedMetrics, err := nm.List(context.TODO(), metav1.ListOptions{LabelSelector: edgeNodeLabel})
	if err != nil {
		return nil, err
	}
	metrics := &metricsapi.NodeMetricsList{}
	err = metricsV1beta1.Convert_v1beta1_NodeMetricsList_To_metrics_NodeMetricsList(versionedMetrics, metrics, nil)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

// pods metrics of edge nodes
func (m metricsServer) getPodMetricsFromMetricsAPI(edgePods map[string]bool, opts *monitoring.QueryOptions) ([]metricsapi.PodMetrics, error) {
	mc := m.metricsClient.MetricsV1beta1()
	podName := opts.PodName
	ns := opts.NamespaceName

	// single pod request
	if ns != "" && podName != "" {
		pm := mc.PodMetricses(ns)
		versionedMetrics, err := pm.Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			klog.Error("Get pod metrics on edge node error:", err)
			return nil, err
		}
		metrics := &metricsapi.PodMetrics{}
		err = metricsV1beta1.Convert_v1beta1_PodMetrics_To_metrics_PodMetrics(versionedMetrics, metrics, nil)
		if err != nil {
			klog.Error("Convert pod metrics on edge node error:", err)
			return nil, err
		}
		return []metricsapi.PodMetrics{*metrics}, nil
	}

	combinedPodMetrics := []metricsapi.PodMetrics{}

	// handle cases with when edgePodName contains namespaceName
	if opts.NamespacedResourcesFilter != "" {
		for p := range edgePods {
			splitedPodName := strings.Split(p, "/")
			ns, p = strings.ReplaceAll(splitedPodName[0], " ", ""), strings.ReplaceAll(splitedPodName[1], " ", "")
			pm := mc.PodMetricses(ns)
			versionedMetrics, err := pm.Get(context.TODO(), p, metav1.GetOptions{})
			if err != nil {
				klog.Error("Get pod metrics on edge node error:", err)
				continue
			}
			metrics := &metricsapi.PodMetrics{}
			err = metricsV1beta1.Convert_v1beta1_PodMetrics_To_metrics_PodMetrics(versionedMetrics, metrics, nil)
			if err != nil {
				klog.Error("Convert pod metrics on edge node error:", err)
				continue
			}
			combinedPodMetrics = append(combinedPodMetrics, *metrics)
		}
		return combinedPodMetrics, nil
	}

	// use list request in other cases
	pm := mc.PodMetricses(ns)
	versionedMetricsList, err := pm.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Error("List pod metrics on edge node error:", err)
		return nil, err
	}
	podMetrics := &metricsapi.PodMetricsList{}
	err = metricsV1beta1.Convert_v1beta1_PodMetricsList_To_metrics_PodMetricsList(versionedMetricsList, podMetrics, nil)
	if err != nil {
		klog.Error("Convert pod metrics on edge node error:", err)
		return nil, err
	}
	for _, podMetric := range podMetrics.Items {
		if _, ok := edgePods[podMetric.Name]; !ok {
			continue
		}
		combinedPodMetrics = append(combinedPodMetrics, podMetric)
	}
	return combinedPodMetrics, nil

}

func NewMetricsClient(k kubernetes.Interface, options *k8s.KubernetesOptions) monitoring.Interface {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		klog.Error(err)
		return nil
	}

	discoveryClient := k.Discovery()
	apiGroups, err := discoveryClient.ServerGroups()
	if err != nil {
		klog.Error(err)
		return nil
	}

	metricsAPIAvailable := metricsAPISupported(apiGroups)

	if !metricsAPIAvailable {
		klog.Warningf("Metrics API not available.")
		return nil
	}

	metricsClient, err := metricsclient.NewForConfig(config)
	if err != nil {
		klog.Error(err)
		return nil
	}

	return NewMetricsServer(k, metricsAPIAvailable, metricsClient)
}

func NewMetricsServer(k kubernetes.Interface, a bool, m metricsclient.Interface) monitoring.Interface {
	var metricsServer metricsServer

	metricsServer.k8s = k
	metricsServer.metricsAPIAvailable = a
	metricsServer.metricsClient = m

	return metricsServer
}

func (m metricsServer) GetMetric(expr string, ts time.Time) monitoring.Metric {
	var parsedResp monitoring.Metric

	return parsedResp
}

func (m metricsServer) GetMetricOverTime(expr string, start, end time.Time, step time.Duration) monitoring.Metric {
	var parsedResp monitoring.Metric

	return parsedResp
}

// node metrics definition
const (
	metricsNodeCPUUsage           = "node_cpu_usage"
	metricsNodeCPUTotal           = "node_cpu_total"
	metricsNodeCPUUltilisation    = "node_cpu_utilisation"
	metricsNodeMemoryUsageWoCache = "node_memory_usage_wo_cache"
	metricsNodeMemoryTotal        = "node_memory_total"
	metricsNodeMemoryUltilisation = "node_memory_utilisation"
)

var edgeNodeMetrics = []string{metricsNodeCPUUsage, metricsNodeCPUTotal, metricsNodeCPUUltilisation, metricsNodeMemoryUsageWoCache, metricsNodeMemoryTotal, metricsNodeMemoryUltilisation}

// pod metrics definition
const (
	metricsPodCPUUsage    = "pod_cpu_usage"
	metricsPodMemoryUsage = "pod_memory_usage_wo_cache"
)

var (
	edgePodMetrics    = []string{metricsPodCPUUsage, metricsPodMemoryUsage}
	MeasuredResources = []v1.ResourceName{
		v1.ResourceCPU,
		v1.ResourceMemory,
	}
)

func (m metricsServer) parseErrorResp(metrics []string, err error) []monitoring.Metric {
	var res []monitoring.Metric

	for _, metric := range metrics {
		parsedResp := monitoring.Metric{MetricName: metric}
		parsedResp.Error = err.Error()
	}

	return res
}

func (m metricsServer) GetNamedMetrics(metrics []string, ts time.Time, o monitoring.QueryOption) []monitoring.Metric {
	var res []monitoring.Metric

	opts := monitoring.NewQueryOptions()
	o.Apply(opts)

	if !m.metricsAPIAvailable {
		klog.Warningf("Metrics API not available.")
		return m.parseErrorResp(metrics, errors.New("Metrics API not available."))
	}

	switch opts.Level {
	case monitoring.LevelNode:
		return m.GetNodeLevelNamedMetrics(metrics, ts, opts)
	case monitoring.LevelPod:
		return m.GetPodLevelNamedMetrics(metrics, ts, opts)
	default:
		return res
	}
}

func (m metricsServer) GetNodeLevelNamedMetrics(metrics []string, ts time.Time, opts *monitoring.QueryOptions) []monitoring.Metric {
	var res []monitoring.Metric

	edgeNodes, err := m.listEdgeNodes()
	if err != nil {
		klog.Errorf("List edge nodes error %v\n", err)
		return m.parseErrorResp(metrics, err)
	}

	edgeNodeNamesFiltered := m.filterEdgeNodeNames(edgeNodes, opts)
	if len(edgeNodeNamesFiltered) == 0 {
		klog.V(4).Infof("No edge node metrics is requested")
		return res
	}

	status := make(map[string]v1.NodeStatus)
	for n := range edgeNodeNamesFiltered {
		status[n] = edgeNodes[n].Status
	}

	metricsResult, err := m.getNodeMetricsFromMetricsAPI()
	if err != nil {
		klog.Errorf("Get edge node metrics error %v\n", err)
		return m.parseErrorResp(metrics, err)
	}

	metricsMap := make(map[string]bool)
	for _, m := range metrics {
		metricsMap[m] = true
	}

	nodeMetrics := make(map[string]*monitoring.MetricData)
	for _, enm := range edgeNodeMetrics {
		_, ok := metricsMap[enm]
		if ok {
			nodeMetrics[enm] = &monitoring.MetricData{MetricType: monitoring.MetricTypeVector}
		}
	}

	var usage v1.ResourceList
	var cap v1.ResourceList
	for _, m := range metricsResult.Items {
		_, ok := edgeNodeNamesFiltered[m.Name]
		if !ok {
			continue
		}

		m.Usage.DeepCopyInto(&usage)
		status[m.Name].Capacity.DeepCopyInto(&cap)

		metricValues := make(map[string]*monitoring.MetricValue)

		for _, enm := range edgeNodeMetrics {
			metricValues[enm] = &monitoring.MetricValue{
				Metadata: make(map[string]string),
			}
			metricValues[enm].Metadata["node"] = m.Name
			metricValues[enm].Metadata["role"] = "edge"
		}

		for _, addr := range status[m.Name].Addresses {
			if addr.Type == v1.NodeInternalIP {
				for _, enm := range edgeNodeMetrics {
					metricValues[enm].Metadata["host_ip"] = addr.Address
				}
				break
			}
		}

		for k, v := range metricsMap {
			switch k {
			case metricsNodeCPUUsage:
				if v {
					metricValues[metricsNodeCPUUsage].Sample = &monitoring.Point{float64(m.Timestamp.Unix()), float64(usage.Cpu().MilliValue()) / 1000}
				}
			case metricsNodeCPUTotal:
				if v {
					metricValues[metricsNodeCPUTotal].Sample = &monitoring.Point{float64(m.Timestamp.Unix()), float64(cap.Cpu().MilliValue()) / 1000}
				}
			case metricsNodeCPUUltilisation:
				if v {
					metricValues[metricsNodeCPUUltilisation].Sample = &monitoring.Point{float64(m.Timestamp.Unix()), float64(usage.Cpu().MilliValue()) / float64(cap.Cpu().MilliValue())}
				}
			case metricsNodeMemoryUsageWoCache:
				if v {
					metricValues[metricsNodeMemoryUsageWoCache].Sample = &monitoring.Point{float64(m.Timestamp.Unix()), float64(usage.Memory().Value())}
				}
			case metricsNodeMemoryTotal:
				if v {
					metricValues[metricsNodeMemoryTotal].Sample = &monitoring.Point{float64(m.Timestamp.Unix()), float64(cap.Memory().Value())}
				}
			case metricsNodeMemoryUltilisation:
				if v {
					metricValues[metricsNodeMemoryUltilisation].Sample = &monitoring.Point{float64(m.Timestamp.Unix()), float64(usage.Memory().Value()) / float64(cap.Memory().Value())}
				}
			}
		}

		for _, enm := range edgeNodeMetrics {
			_, ok = metricsMap[enm]
			if ok {
				nodeMetrics[enm].MetricValues = append(nodeMetrics[enm].MetricValues, *metricValues[enm])
			}
		}
	}

	for _, enm := range edgeNodeMetrics {
		_, ok := metricsMap[enm]
		if ok {
			res = append(res, monitoring.Metric{MetricName: enm, MetricData: *nodeMetrics[enm]})
		}
	}
	return res
}

func (m metricsServer) GetPodLevelNamedMetrics(metrics []string, ts time.Time, opts *monitoring.QueryOptions) []monitoring.Metric {
	var res []monitoring.Metric

	edgePods := m.parseEdgePods(opts)
	if len(edgePods) == 0 && opts.PodName == "" {
		klog.Errorf("Edge node filter regexp error: %v\n", errors.New("no edge node pods metrics is requested or resource filter invalid"))
		return res
	}

	podMetricsFromMetricsAPI, err := m.getPodMetricsFromMetricsAPI(edgePods, opts)
	if err != nil {
		klog.Errorf("Get pod metrics of edge nodes error %v\n", err)
		return m.parseErrorResp(metrics, err)
	}

	metricsMap := make(map[string]bool)
	for _, m := range metrics {
		metricsMap[m] = true
	}

	// init
	podMetrics := make(map[string]*monitoring.MetricData)
	for _, epm := range edgePodMetrics {
		_, ok := metricsMap[epm]
		if ok {
			podMetrics[epm] = &monitoring.MetricData{MetricType: monitoring.MetricTypeVector}
		}
	}

	for _, p := range podMetricsFromMetricsAPI {

		metricValues := make(map[string]*monitoring.MetricValue)

		for _, epm := range edgePodMetrics {
			metricValues[epm] = &monitoring.MetricValue{
				Metadata: make(map[string]string),
			}
			metricValues[epm].Metadata["pod"] = p.Name
			metricValues[epm].Metadata["namespace"] = p.Namespace
		}

		podMetricsUsge := make(v1.ResourceList)

		for _, res := range MeasuredResources {
			podMetricsUsge[res], _ = resource.ParseQuantity("0")
		}

		for _, podContainer := range p.Containers {
			for _, res := range MeasuredResources {
				quantity := podMetricsUsge[res]
				quantity.Add(podContainer.Usage[res])
				podMetricsUsge[res] = quantity
			}
		}

		for k, v := range metricsMap {
			switch k {
			case metricsPodCPUUsage:
				if v {
					cpuQuantity := podMetricsUsge[v1.ResourceCPU]
					metricValues[metricsPodCPUUsage].Sample = &monitoring.Point{float64(p.Timestamp.Unix()), float64(cpuQuantity.MilliValue()) / 1000}
				}
			case metricsPodMemoryUsage:
				if v {
					memoryQuantity := podMetricsUsge[v1.ResourceMemory]
					metricValues[metricsPodMemoryUsage].Sample = &monitoring.Point{float64(p.Timestamp.Unix()), float64(memoryQuantity.Value()) / (1024 * 1024)}
				}
			}

		}

		for _, epm := range edgePodMetrics {
			_, ok := metricsMap[epm]
			if ok {
				podMetrics[epm].MetricValues = append(podMetrics[epm].MetricValues, *metricValues[epm])
			}
		}
	}

	for _, epm := range edgePodMetrics {
		_, ok := metricsMap[epm]
		if ok {
			res = append(res, monitoring.Metric{MetricName: epm, MetricData: *podMetrics[epm]})
		}
	}
	return res
}

func (m metricsServer) GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, o monitoring.QueryOption) []monitoring.Metric {
	var res []monitoring.Metric

	opts := monitoring.NewQueryOptions()
	o.Apply(opts)

	if !m.metricsAPIAvailable {
		klog.Warningf("Metrics API not available.")
		return m.parseErrorResp(metrics, errors.New("Metrics API not available."))
	}

	switch opts.Level {
	case monitoring.LevelNode:
		return m.GetNodeLevelNamedMetricsOverTime(metrics, start, end, step, opts)
	case monitoring.LevelPod:
		return m.GetPodLevelNamedMetricsOverTime(metrics, start, end, step, opts)
	default:
		return res
	}

}

func (m metricsServer) GetNodeLevelNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opts *monitoring.QueryOptions) []monitoring.Metric {
	var res []monitoring.Metric
	edgeNodes, err := m.listEdgeNodes()
	if err != nil {
		klog.Errorf("List edge nodes error %v\n", err)
		return m.parseErrorResp(metrics, err)
	}

	edgeNodeNamesFiltered := m.filterEdgeNodeNames(edgeNodes, opts)
	if len(edgeNodeNamesFiltered) == 0 {
		klog.V(4).Infof("No edge node metrics is requested")
		return res
	}

	metricsResult, err := m.getNodeMetricsFromMetricsAPI()
	if err != nil {
		klog.Errorf("Get edge node metrics error %v\n", err)
		return m.parseErrorResp(metrics, err)
	}

	metricsMap := make(map[string]bool)
	for _, m := range metrics {
		metricsMap[m] = true
	}

	status := make(map[string]v1.NodeStatus)
	for n := range edgeNodeNamesFiltered {
		status[n] = edgeNodes[n].Status
	}

	nodeMetrics := make(map[string]*monitoring.MetricData)
	for _, enm := range edgeNodeMetrics {
		_, ok := metricsMap[enm]
		if ok {
			nodeMetrics[enm] = &monitoring.MetricData{MetricType: monitoring.MetricTypeMatrix}
		}
	}

	var usage v1.ResourceList
	var cap v1.ResourceList
	for _, m := range metricsResult.Items {
		_, ok := edgeNodeNamesFiltered[m.Name]
		if !ok {
			continue
		}

		m.Usage.DeepCopyInto(&usage)
		status[m.Name].Capacity.DeepCopyInto(&cap)

		metricValues := make(map[string]*monitoring.MetricValue)

		for _, enm := range edgeNodeMetrics {
			metricValues[enm] = &monitoring.MetricValue{
				Metadata: make(map[string]string),
			}
			metricValues[enm].Metadata["node"] = m.Name
			metricValues[enm].Metadata["role"] = "edge"
		}
		for _, addr := range status[m.Name].Addresses {
			if addr.Type == v1.NodeInternalIP {
				for _, enm := range edgeNodeMetrics {
					metricValues[enm].Metadata["host_ip"] = addr.Address
				}
				break
			}
		}

		for k, v := range metricsMap {
			switch k {
			case metricsNodeCPUUsage:
				if v {
					metricValues[metricsNodeCPUUsage].Series = append(metricValues[metricsNodeCPUUsage].Series, monitoring.Point{float64(m.Timestamp.Unix()), float64(usage.Cpu().MilliValue()) / 1000})
				}
			case metricsNodeCPUTotal:
				if v {
					metricValues[metricsNodeCPUTotal].Series = append(metricValues[metricsNodeCPUTotal].Series, monitoring.Point{float64(m.Timestamp.Unix()), float64(cap.Cpu().MilliValue()) / 1000})
				}
			case metricsNodeCPUUltilisation:
				if v {
					metricValues[metricsNodeCPUUltilisation].Series = append(metricValues[metricsNodeCPUUltilisation].Series, monitoring.Point{float64(m.Timestamp.Unix()), float64(usage.Cpu().MilliValue()) / float64(cap.Cpu().MilliValue())})
				}
			case metricsNodeMemoryUsageWoCache:
				if v {
					metricValues[metricsNodeMemoryUsageWoCache].Series = append(metricValues[metricsNodeMemoryUsageWoCache].Series, monitoring.Point{float64(m.Timestamp.Unix()), float64(usage.Memory().Value())})
				}
			case metricsNodeMemoryTotal:
				if v {
					metricValues[metricsNodeMemoryTotal].Series = append(metricValues[metricsNodeMemoryTotal].Series, monitoring.Point{float64(m.Timestamp.Unix()), float64(cap.Memory().Value())})
				}
			case metricsNodeMemoryUltilisation:
				if v {
					metricValues[metricsNodeMemoryUltilisation].Series = append(metricValues[metricsNodeMemoryUltilisation].Series, monitoring.Point{float64(m.Timestamp.Unix()), float64(usage.Memory().Value()) / float64(cap.Memory().Value())})
				}
			}
		}

		for _, enm := range edgeNodeMetrics {
			_, ok := metricsMap[enm]
			if ok {
				nodeMetrics[enm].MetricValues = append(nodeMetrics[enm].MetricValues, *metricValues[enm])
			}
		}
	}

	for _, enm := range edgeNodeMetrics {
		_, ok := metricsMap[enm]
		if ok {
			res = append(res, monitoring.Metric{MetricName: enm, MetricData: *nodeMetrics[enm]})
		}
	}
	return res
}

func (m metricsServer) GetPodLevelNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opts *monitoring.QueryOptions) []monitoring.Metric {
	var res []monitoring.Metric

	edgePods := m.parseEdgePods(opts)
	if len(edgePods) == 0 && opts.PodName == "" {
		klog.Errorf("Edge node filter regexp error: %v\n", errors.New("no edge node pods metrics is requested or resource filter invalid"))
		return res
	}

	podMetricsFromMetricsAPI, err := m.getPodMetricsFromMetricsAPI(edgePods, opts)
	if err != nil {
		klog.Errorf("Get pod metrics of edge nodes error %v\n", err)
		return m.parseErrorResp(metrics, err)
	}

	metricsMap := make(map[string]bool)
	for _, m := range metrics {
		metricsMap[m] = true
	}

	// init
	podMetrics := make(map[string]*monitoring.MetricData)
	for _, epm := range edgePodMetrics {
		_, ok := metricsMap[epm]
		if ok {
			podMetrics[epm] = &monitoring.MetricData{MetricType: monitoring.MetricTypeMatrix}
		}
	}

	for _, p := range podMetricsFromMetricsAPI {

		metricValues := make(map[string]*monitoring.MetricValue)

		for _, epm := range edgePodMetrics {
			metricValues[epm] = &monitoring.MetricValue{
				Metadata: make(map[string]string),
			}
			metricValues[epm].Metadata["pod"] = p.Name
			metricValues[epm].Metadata["namespace"] = p.Namespace
		}

		podMetricsUsge := make(v1.ResourceList)

		for _, res := range MeasuredResources {
			podMetricsUsge[res], _ = resource.ParseQuantity("0")
		}

		for _, podContainer := range p.Containers {
			for _, res := range MeasuredResources {
				quantity := podMetricsUsge[res]
				quantity.Add(podContainer.Usage[res])
				podMetricsUsge[res] = quantity
			}
		}

		for k, v := range metricsMap {
			switch k {
			case metricsPodCPUUsage:
				if v {
					cpuQuantity := podMetricsUsge[v1.ResourceCPU]
					metricValues[metricsPodCPUUsage].Series = append(metricValues[metricsPodCPUUsage].Series, monitoring.Point{float64(p.Timestamp.Unix()), float64(cpuQuantity.MilliValue()) / 1000})
				}
			case metricsPodMemoryUsage:
				if v {
					memoryQuantity := podMetricsUsge[v1.ResourceMemory]
					metricValues[metricsPodMemoryUsage].Series = append(metricValues[metricsPodMemoryUsage].Series, monitoring.Point{float64(p.Timestamp.Unix()), float64(memoryQuantity.Value()) / (1024 * 1024)})
				}
			}
		}

		for _, epm := range edgePodMetrics {
			_, ok := metricsMap[epm]
			if ok {
				podMetrics[epm].MetricValues = append(podMetrics[epm].MetricValues, *metricValues[epm])
			}
		}
	}

	for _, epm := range edgePodMetrics {
		_, ok := metricsMap[epm]
		if ok {
			res = append(res, monitoring.Metric{MetricName: epm, MetricData: *podMetrics[epm]})
		}
	}
	return res
}

func (m metricsServer) GetMetadata(namespace string) []monitoring.Metadata {
	var meta []monitoring.Metadata

	return meta
}

func (m metricsServer) GetMetricLabelSet(expr string, start, end time.Time) []map[string]string {
	var res []map[string]string

	return res
}

// meter
func (m metricsServer) GetNamedMeters(meters []string, time time.Time, opts []monitoring.QueryOption) []monitoring.Metric {
	return nil
}
func (m metricsServer) GetNamedMetersOverTime(metrics []string, start, end time.Time, step time.Duration, opts []monitoring.QueryOption) []monitoring.Metric {
	return nil
}
