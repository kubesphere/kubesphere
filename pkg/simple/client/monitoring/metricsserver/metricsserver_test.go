package metricsserver

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/json-iterator/go"
	"io/ioutil"

	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
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
func mergeResourceLists(resourceLists ...corev1.ResourceList) corev1.ResourceList {
	result := corev1.ResourceList{}
	for _, rl := range resourceLists {
		for resource, quantity := range rl {
			result[resource] = quantity
		}
	}
	return result
}

func getResourceList(cpu, memory string) corev1.ResourceList {
	res := corev1.ResourceList{}
	if cpu != "" {
		res[corev1.ResourceCPU] = resource.MustParse(cpu)
	}
	if memory != "" {
		res[corev1.ResourceMemory] = resource.MustParse(memory)
	}
	return res
}

var nodeCapacity = mergeResourceLists(getResourceList("8", "8Gi"))
var node1 = &corev1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "edgenode-1",
		Labels: map[string]string{
			"node-role.kubernetes.io/edge": "",
		},
	},
	Status: corev1.NodeStatus{
		Capacity: nodeCapacity,
	},
}

var node2 = &corev1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "edgenode-2",
		Labels: map[string]string{
			"node-role.kubernetes.io/edge": "",
		},
	},
	Status: corev1.NodeStatus{
		Capacity: nodeCapacity,
	},
}

func TestGetNamedMetrics(t *testing.T) {
	tests := []struct {
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

	fakeK8sClient := fakek8s.NewSimpleClientset(node1, node2)
	informer := informers.NewSharedInformerFactory(fakeK8sClient, 0)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(node1)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(node2)

	fakeMetricsclient := &fakemetricsclient.Clientset{}
	layout := "2006-01-02T15:04:05.000Z"
	str := "2021-01-25T12:34:56.789Z"
	metricsTime, _ := time.Parse(layout, str)

	fakeMetricsclient.AddReactor("list", "nodes", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		metrics := &metricsV1beta1.NodeMetricsList{}
		nodeMetric1 := metricsV1beta1.NodeMetrics{
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
		nodeMetric2 := metricsV1beta1.NodeMetrics{
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
		metrics.Items = append(metrics.Items, nodeMetric1)
		metrics.Items = append(metrics.Items, nodeMetric2)

		return true, metrics, nil
	})

	for i, tt := range tests {
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
}

func TestGetNamedMetricsOverTime(t *testing.T) {
	tests := []struct {
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

	fakeK8sClient := fakek8s.NewSimpleClientset(node1, node2)
	informer := informers.NewSharedInformerFactory(fakeK8sClient, 0)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(node1)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(node2)

	fakeMetricsclient := &fakemetricsclient.Clientset{}
	layout := "2006-01-02T15:04:05.000Z"
	str := "2021-01-25T12:34:56.789Z"
	metricsTime, _ := time.Parse(layout, str)

	fakeMetricsclient.AddReactor("list", "nodes", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		metrics := &metricsV1beta1.NodeMetricsList{}
		nodeMetric1 := metricsV1beta1.NodeMetrics{
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
		nodeMetric2 := metricsV1beta1.NodeMetrics{
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
		metrics.Items = append(metrics.Items, nodeMetric1)
		metrics.Items = append(metrics.Items, nodeMetric2)

		return true, metrics, nil
	})

	for i, tt := range tests {
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
