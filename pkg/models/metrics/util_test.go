package metrics

import (
	"kubesphere.io/kubesphere/pkg/api/monitoring/v1alpha2"
	"reflect"
	"testing"
)

func TestSortBy(t *testing.T) {
	tests := []struct {
		description string
		rawMetrics  *Response
		sortMetrics string
		sortType    string
		expected    *Response
	}{
		{
			description: "sort a set of node metrics for node1, node2 and a node with blank name (this should be considered as abnormal).",
			rawMetrics: &Response{
				Results: []APIResponse{
					{
						MetricName: "node_cpu_usage",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "0.221"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "0.177"},
									},
									{
										Metric: map[string]string{"resource_name": ""},
										Value:  []interface{}{1578987135.334, ""},
									},
								},
							},
						},
					},
					{
						MetricName: "node_memory_total",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "8201043968"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "8201039872"},
									},
									{
										Metric: map[string]string{"resource_name": ""},
										Value:  []interface{}{1578987135.334, ""},
									},
								},
							},
						},
					},
					{
						MetricName: "node_pod_running_count",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "19"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "6"},
									},
									{
										Metric: map[string]string{"resource_name": ""},
										Value:  []interface{}{1578987135.334, ""},
									},
								},
							},
						},
					},
				},
			},
			sortMetrics: "node_cpu_usage",
			sortType:    "desc",
			expected: &Response{
				Results: []APIResponse{
					{
						MetricName: "node_cpu_usage",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "0.221"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "0.177"},
									},
								},
							},
						},
					},
					{
						MetricName: "node_memory_total",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "8201043968"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "8201039872"},
									},
								},
							},
						},
					},
					{
						MetricName: "node_pod_running_count",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "19"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "6"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			description: "sort a set of node metrics for node1, node2 and node3.",
			rawMetrics: &Response{
				Results: []APIResponse{
					{
						MetricName: "node_cpu_usage",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "0.221"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "0.177"},
									},
									{
										Metric: map[string]string{"resource_name": "node3"},
										Value:  []interface{}{1578987135.334, "0.194"},
									},
								},
							},
						},
					},
					{
						MetricName: "node_memory_total",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "8201043968"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "8201039872"},
									},
									{
										Metric: map[string]string{"resource_name": "node3"},
										Value:  []interface{}{1578987135.334, "8201056256"},
									},
								},
							},
						},
					},
					{
						MetricName: "node_pod_running_count",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "19"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "6"},
									},
									{
										Metric: map[string]string{"resource_name": "node3"},
										Value:  []interface{}{1578987135.334, "14"},
									},
								},
							},
						},
					},
				},
			},
			sortMetrics: "node_cpu_usage",
			sortType:    "desc",
			expected: &Response{
				Results: []APIResponse{
					{
						MetricName: "node_cpu_usage",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "0.221"},
									},
									{
										Metric: map[string]string{"resource_name": "node3"},
										Value:  []interface{}{1578987135.334, "0.194"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "0.177"},
									},
								},
							},
						},
					},
					{
						MetricName: "node_memory_total",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "8201043968"},
									},
									{
										Metric: map[string]string{"resource_name": "node3"},
										Value:  []interface{}{1578987135.334, "8201056256"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "8201039872"},
									},
								},
							},
						},
					},
					{
						MetricName: "node_pod_running_count",
						APIResponse: v1alpha2.APIResponse{
							Status: "success",
							Data: v1alpha2.QueryResult{
								ResultType: "vector",
								Result: []v1alpha2.QueryValue{
									{
										Metric: map[string]string{"resource_name": "node1"},
										Value:  []interface{}{1578987135.334, "19"},
									},
									{
										Metric: map[string]string{"resource_name": "node3"},
										Value:  []interface{}{1578987135.334, "14"},
									},
									{
										Metric: map[string]string{"resource_name": "node2"},
										Value:  []interface{}{1578987135.334, "6"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		res, _ := test.rawMetrics.SortBy(test.sortMetrics, test.sortType)
		if !reflect.DeepEqual(res, test.expected) {
			t.Errorf("got unexpected results: %v", res)
		}
	}
}
