package metricsserver

import (
	"fmt"
	"testing"
	"time"

	"io/ioutil"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	metricsV1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	fakemetricsclient "k8s.io/metrics/pkg/client/clientset/versioned/fake"

	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

// mergeResourceLists will merge resoure lists. When two lists have the same resourece, the value from
// the last list will be present in the result
func mergeResourceLists(resourceLists ...v1.ResourceList) v1.ResourceList {
	result := v1.ResourceList{}
	for _, rl := range resourceLists {
		for resource, quantity := range rl {
			result[resource] = quantity
		}
	}
	return result
}

func getResourceList(cpu, memory string) v1.ResourceList {
	res := v1.ResourceList{}
	if cpu != "" {
		res[v1.ResourceCPU] = resource.MustParse(cpu)
	}
	if memory != "" {
		res[v1.ResourceMemory] = resource.MustParse(memory)
	}
	return res
}

var nodeCapacity = mergeResourceLists(getResourceList("8", "8Gi"))
var node1 = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "edgenode-1",
		Labels: map[string]string{
			"node-role.kubernetes.io/edge": "",
		},
	},
	Status: v1.NodeStatus{
		Capacity: nodeCapacity,
	},
}

var node2 = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "edgenode-2",
		Labels: map[string]string{
			"node-role.kubernetes.io/edge": "",
		},
	},
	Status: v1.NodeStatus{
		Capacity: nodeCapacity,
	},
}

var pod1 = &v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "pod1",
		Namespace: "ns1",
	},
}

var pod2 = &v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "pod2",
		Namespace: "ns2",
	},
}

const (
	layout = "2006-01-02T15:04:05.000Z"
	str    = "2021-03-25T12:34:56.789Z"
)

var (
	metricsTime, _ = time.Parse(layout, str)
	nodeMetric1    = metricsV1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name: "edgenode-1",
			Labels: map[string]string{
				"node-role.kubernetes.io/edge": "",
			},
		},
		Timestamp: metav1.Time{Time: metricsTime},
		Window:    metav1.Duration{Duration: time.Minute},
		Usage: v1.ResourceList{
			v1.ResourceCPU: *resource.NewMilliQuantity(
				int64(1000),
				resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(
				int64(1024*1024),
				resource.BinarySI),
		},
	}
	nodeMetric2 = metricsV1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name: "edgenode-2",
			Labels: map[string]string{
				"node-role.kubernetes.io/edge": "",
			},
		},
		Timestamp: metav1.Time{Time: metricsTime},
		Window:    metav1.Duration{Duration: time.Minute},
		Usage: v1.ResourceList{
			v1.ResourceCPU: *resource.NewMilliQuantity(
				int64(2000),
				resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(
				int64(2*1024*1024),
				resource.BinarySI),
		},
	}
	podMetric1 = metricsV1beta1.PodMetrics{
		TypeMeta: metav1.TypeMeta{
			Kind: "DaemonSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns1",
		},
		Timestamp: metav1.Time{Time: metricsTime},
		Window:    metav1.Duration{Duration: time.Minute},
		Containers: []metricsV1beta1.ContainerMetrics{
			metricsV1beta1.ContainerMetrics{
				Name: "containers-1",
				Usage: v1.ResourceList{
					v1.ResourceCPU: *resource.NewMilliQuantity(
						1,
						resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(
						int64(1024*1024),
						resource.DecimalSI),
				},
			},
		},
	}
	podMetric2 = metricsV1beta1.PodMetrics{
		TypeMeta: metav1.TypeMeta{
			Kind: "DaemonSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "ns2",
		},
		Timestamp: metav1.Time{Time: metricsTime},
		Window:    metav1.Duration{Duration: time.Minute},
		Containers: []metricsV1beta1.ContainerMetrics{
			metricsV1beta1.ContainerMetrics{
				Name: "containers-1",
				Usage: v1.ResourceList{
					v1.ResourceCPU: *resource.NewMilliQuantity(
						1,
						resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(
						int64(1024*1024),
						resource.DecimalSI),
				},
			},
			metricsV1beta1.ContainerMetrics{
				Name: "containers-2",
				Usage: v1.ResourceList{
					v1.ResourceCPU: *resource.NewMilliQuantity(
						1,
						resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(
						int64(1024*1024),
						resource.DecimalSI),
				},
			},
		},
	}
)

func TestGetNamedMetrics(t *testing.T) {
	nodeMetricsTests := []struct {
		metrics  []string
		filter   string
		expected string
	}{
		{
			metrics:  []string{"node_cpu_usage", "node_memory_usage_wo_cache"},
			filter:   ".*",
			expected: "metrics-vector-1.json",
		},
		{
			metrics:  []string{"node_cpu_usage", "node_cpu_utilisation"},
			filter:   "edgenode-2",
			expected: "metrics-vector-2.json",
		},
		{
			metrics:  []string{"node_memory_usage_wo_cache", "node_memory_utilisation"},
			filter:   "edgenode-1|edgenode-2",
			expected: "metrics-vector-3.json",
		},
	}
	podMetricsTests := []struct {
		metrics          []string
		filter           string
		namespacedFilter string
		expected         string
		podName          string
		namespaceName    string
	}{
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "pod1$",
			namespacedFilter: "",
			expected:         "metrics-vector-4.json",
			podName:          "",
			namespaceName:    "",
		},
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "pod1|pod2$",
			namespacedFilter: "",
			expected:         "metrics-vector-5.json",
			podName:          "",
			namespaceName:    "",
		},
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "",
			namespacedFilter: "ns2/pod2$",
			expected:         "metrics-vector-6.json",
			podName:          "",
			namespaceName:    "",
		},
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "",
			namespacedFilter: "ns1/pod1|ns2/pod2$",
			expected:         "metrics-vector-7.json",
			podName:          "",
			namespaceName:    "",
		},
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "",
			namespacedFilter: "",
			expected:         "metrics-vector-8.json",
			podName:          "pod1",
			namespaceName:    "ns1",
		},
	}

	fakeK8sClient := fakek8s.NewSimpleClientset(node1, node2, pod1, pod2)
	informer := informers.NewSharedInformerFactory(fakeK8sClient, 0)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(node1)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(node2)
	informer.Core().V1().Pods().Informer().GetIndexer().Add(pod1)
	informer.Core().V1().Pods().Informer().GetIndexer().Add(pod2)

	fakeMetricsclient := &fakemetricsclient.Clientset{}

	fakeMetricsclient.AddReactor("list", "nodes", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		metrics := &metricsV1beta1.NodeMetricsList{}

		metrics.Items = append(metrics.Items, nodeMetric1)
		metrics.Items = append(metrics.Items, nodeMetric2)

		return true, metrics, nil
	})

	fakeMetricsclient.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		metrics := &metricsV1beta1.PodMetricsList{}
		metrics.Items = append(metrics.Items, podMetric1)
		metrics.Items = append(metrics.Items, podMetric2)
		return true, metrics, nil
	})

	// test for node edge
	for i, tt := range nodeMetricsTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expected := make([]monitoring.Metric, 0)
			err := jsonFromFile(tt.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}

			client := NewMetricsServer(fakeK8sClient, true, fakeMetricsclient)
			result := client.GetNamedMetrics(tt.metrics, time.Now(), monitoring.NodeOption{ResourceFilter: tt.filter})
			if diff := cmp.Diff(result, expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}

	//test for pods on the node edges
	for i, tt := range podMetricsTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expected := make([]monitoring.Metric, 0)
			err := jsonFromFile(tt.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}
			fakeMetricsclient.ClearActions()
			if tt.podName == "pod1" || tt.filter == "pod1$" || tt.namespacedFilter == "ns1/pod1$" {
				fakeMetricsclient.PrependReactor("get", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
					return true, &podMetric1, nil
				})
			} else if tt.podName == "pod2" || tt.filter == "pod2$" || tt.namespacedFilter == "ns2/pod2$" {
				fakeMetricsclient.PrependReactor("get", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
					return true, &podMetric2, nil
				})
			} else {
				fakeMetricsclient.PrependReactor("get", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
					ns := action.GetNamespace()
					if ns == "ns1" {
						return true, &podMetric1, nil
					} else if ns == "ns2" {
						return true, &podMetric2, nil
					} else {
						return true, &metricsV1beta1.PodMetricsList{}, nil
					}
				})
			}

			client := NewMetricsServer(fakeK8sClient, true, fakeMetricsclient)
			result := client.GetNamedMetrics(
				tt.metrics,
				time.Now(),
				monitoring.PodOption{
					ResourceFilter:            tt.filter,
					NamespacedResourcesFilter: tt.namespacedFilter,
					PodName:                   tt.podName,
					NamespaceName:             tt.namespaceName,
				})
			if diff := cmp.Diff(sortedResults(result), expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func TestGetNamedMetricsOverTime(t *testing.T) {
	nodeMetricsTests := []struct {
		metrics  []string
		filter   string
		expected string
	}{
		{
			metrics:  []string{"node_cpu_usage", "node_memory_usage_wo_cache"},
			filter:   ".*",
			expected: "metrics-matrix-1.json",
		},
		{
			metrics:  []string{"node_cpu_usage", "node_cpu_utilisation"},
			filter:   "edgenode-2",
			expected: "metrics-matrix-2.json",
		},
		{
			metrics:  []string{"node_memory_usage_wo_cache", "node_memory_utilisation"},
			filter:   "edgenode-1|edgenode-2",
			expected: "metrics-matrix-3.json",
		},
	}

	podMetricsTests := []struct {
		metrics          []string
		filter           string
		namespacedFilter string
		expected         string
		podName          string
		namespaceName    string
	}{
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "pod1$",
			namespacedFilter: "",
			expected:         "metrics-matrix-4.json",
			podName:          "",
			namespaceName:    "",
		},
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "pod1|pod2$",
			namespacedFilter: "",
			expected:         "metrics-matrix-5.json",
			podName:          "",
			namespaceName:    "",
		},
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "",
			namespacedFilter: "ns2/pod2$",
			expected:         "metrics-matrix-6.json",
			podName:          "",
			namespaceName:    "",
		},
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "",
			namespacedFilter: "ns1/pod1|ns2/pod2$",
			expected:         "metrics-matrix-7.json",
			podName:          "",
			namespaceName:    "",
		},
		{
			metrics:          []string{"pod_cpu_usage", "pod_memory_usage_wo_cache"},
			filter:           "",
			namespacedFilter: "",
			expected:         "metrics-matrix-8.json",
			podName:          "pod1",
			namespaceName:    "ns1",
		},
	}

	fakeK8sClient := fakek8s.NewSimpleClientset(node1, node2, pod1, pod2)
	informer := informers.NewSharedInformerFactory(fakeK8sClient, 0)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(node1)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(node2)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(pod1)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(pod2)

	fakeMetricsclient := &fakemetricsclient.Clientset{}

	fakeMetricsclient.AddReactor("list", "nodes", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		metrics := &metricsV1beta1.NodeMetricsList{}
		metrics.Items = append(metrics.Items, nodeMetric1)
		metrics.Items = append(metrics.Items, nodeMetric2)

		return true, metrics, nil
	})

	fakeMetricsclient.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		metrics := &metricsV1beta1.PodMetricsList{}
		metrics.Items = append(metrics.Items, podMetric1)
		metrics.Items = append(metrics.Items, podMetric2)
		return true, metrics, nil
	})

	for i, tt := range nodeMetricsTests {
		fakeMetricsclient.Fake.ClearActions()
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expected := make([]monitoring.Metric, 0)
			err := jsonFromFile(tt.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}

			client := NewMetricsServer(fakeK8sClient, true, fakeMetricsclient)
			result := client.GetNamedMetricsOverTime(tt.metrics, time.Now().Add(-time.Minute*3), time.Now(), time.Minute, monitoring.NodeOption{ResourceFilter: tt.filter})
			if diff := cmp.Diff(result, expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}

	for i, tt := range podMetricsTests {

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expected := make([]monitoring.Metric, 0)
			err := jsonFromFile(tt.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}

			fakeMetricsclient.ClearActions()
			if tt.podName == "pod1" || tt.filter == "pod1$" || tt.namespacedFilter == "ns1/pod1$" {
				fakeMetricsclient.PrependReactor("get", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
					return true, &podMetric1, nil
				})
			} else if tt.podName == "pod2" || tt.filter == "pod2$" || tt.namespacedFilter == "ns2/pod2$" {
				fakeMetricsclient.PrependReactor("get", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
					return true, &podMetric2, nil
				})
			} else {
				fakeMetricsclient.PrependReactor("get", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
					ns := action.GetNamespace()
					if ns == "ns1" {
						return true, &podMetric1, nil
					} else if ns == "ns2" {
						return true, &podMetric2, nil
					} else {
						return true, &metricsV1beta1.PodMetricsList{}, nil
					}
				})
			}

			client := NewMetricsServer(fakeK8sClient, true, fakeMetricsclient)
			result := client.GetNamedMetricsOverTime(
				tt.metrics, time.Now().Add(-time.Minute*3),
				time.Now(),
				time.Minute,
				monitoring.PodOption{
					ResourceFilter:            tt.filter,
					NamespacedResourcesFilter: tt.namespacedFilter,
					PodName:                   tt.podName,
					NamespaceName:             tt.namespaceName,
				})
			if diff := cmp.Diff(sortedResults(result), expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func jsonFromFile(expectedFile string, expectedJsonPtr interface{}) error {
	json, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", expectedFile))
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(json, expectedJsonPtr)
	if err != nil {
		return err
	}

	return nil
}

func sortedResults(result []monitoring.Metric) []monitoring.Metric {

	for _, mr := range result {
		metricValues := mr.MetricData.MetricValues
		length := len(metricValues)
		for i, mv := range metricValues {
			podName, _ := mv.Metadata["pod"]
			if i == 0 && podName == "pod2" && length >= 2 {
				metricValues[0], metricValues[1] = metricValues[1], metricValues[0]
			}
			break

		}
	}

	return result

}
