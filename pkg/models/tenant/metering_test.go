package tenant

import (
	"fmt"
	"testing"
	"time"

	"kubesphere.io/kubesphere/pkg/constants"

	"github.com/google/go-cmp/cmp"

	"kubesphere.io/kubesphere/pkg/models/metering"
	monitoringmodel "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

func TestIsRangeQuery(t *testing.T) {
	tests := []struct {
		options       QueryOptions
		expectedValue bool
	}{
		{
			options:       QueryOptions{},
			expectedValue: true,
		},
		{
			options:       QueryOptions{Time: time.Now()},
			expectedValue: false,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if tt.options.isRangeQuery() != tt.expectedValue {
				t.Fatal("error isRangeQuery")
			}
		})
	}
}

func TestShouldSort(t *testing.T) {
	tests := []struct {
		options       QueryOptions
		expectedValue bool
	}{
		{
			options: QueryOptions{
				Target:     "test",
				Identifier: "test",
			},
			expectedValue: true,
		},
		{
			options:       QueryOptions{},
			expectedValue: false,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if tt.options.shouldSort() != tt.expectedValue {
				t.Fatal("error shouldSort")
			}
		})
	}
}

func TestGetMetricPosMap(t *testing.T) {
	tests := []struct {
		metrics       []monitoring.Metric
		expectedValue map[string]int
	}{
		{
			metrics: []monitoring.Metric{
				{MetricName: "one"},
				{MetricName: "two"},
				{MetricName: "three"},
				{MetricName: "four"},
			},
			expectedValue: map[string]int{
				"one":   0,
				"two":   1,
				"three": 2,
				"four":  3,
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if diff := cmp.Diff(getMetricPosMap(tt.metrics), tt.expectedValue); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", tt.expectedValue, diff)
			}
		})
	}
}

func TestTransformMetricData(t *testing.T) {
	tests := []struct {
		metrics       monitoringmodel.Metrics
		expectedValue metering.PodsStats
	}{
		{
			metrics: monitoringmodel.Metrics{
				Results: []monitoring.Metric{
					{
						MetricName: "meter_pod_cpu_usage",
						MetricData: monitoring.MetricData{
							MetricValues: monitoring.MetricValues{
								{
									Metadata: map[string]string{
										"pod": "pod1",
									},
									SumValue: "10",
								},
							},
						},
					},
					{
						MetricName: "meter_pod_memory_usage_wo_cache",
						MetricData: monitoring.MetricData{
							MetricValues: monitoring.MetricValues{
								{
									Metadata: map[string]string{
										"pod": "pod1",
									},
									SumValue: "200",
								},
							},
						},
					},
					{
						MetricName: "meter_pod_net_bytes_transmitted",
						MetricData: monitoring.MetricData{
							MetricValues: monitoring.MetricValues{
								{
									Metadata: map[string]string{
										"pod": "pod1",
									},
									SumValue: "300",
								},
							},
						},
					},
					{
						MetricName: "meter_pod_net_bytes_received",
						MetricData: monitoring.MetricData{
							MetricValues: monitoring.MetricValues{
								{
									Metadata: map[string]string{
										"pod": "pod1",
									},
									SumValue: "300",
								},
							},
						},
					},
					{
						MetricName: "meter_pod_pvc_bytes_total",
						MetricData: monitoring.MetricData{
							MetricValues: monitoring.MetricValues{
								{
									Metadata: map[string]string{
										"pod": "pod1",
									},
									SumValue: "400",
								},
							},
						},
					},
				},
			},
			expectedValue: metering.PodsStats{
				"pod1": &metering.PodStatistic{
					CPUUsage:            10,
					MemoryUsageWoCache:  200,
					NetBytesReceived:    300,
					NetBytesTransmitted: 300,
					PVCBytesTotal:       400,
				},
			},
		},
	}

	var tOperator tenantOperator

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if diff := cmp.Diff(tOperator.transformMetricData(tt.metrics), tt.expectedValue); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", tt.expectedValue, diff)
			}
		})
	}

}

func TestGetAppNameFromLabels(t *testing.T) {
	var tOperator tenantOperator

	tests := []struct {
		labels        map[string]string
		expectedValue string
	}{
		{
			labels:        make(map[string]string),
			expectedValue: "",
		},
		{
			labels: map[string]string{
				constants.ApplicationName:    "app1",
				constants.ApplicationVersion: "v2",
			},
			expectedValue: "app1:v2",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if diff := cmp.Diff(tOperator.getAppNameFromLabels(tt.labels), tt.expectedValue); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", tt.expectedValue, diff)
			}
		})
	}
}
