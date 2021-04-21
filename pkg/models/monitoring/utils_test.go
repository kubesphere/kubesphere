package monitoring

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

func TestGetMaxPointValue(t *testing.T) {
	tests := []struct {
		actualPoints  []monitoring.Point
		expectedValue string
	}{
		{
			actualPoints: []monitoring.Point{
				{1.0, 2.0},
				{3.0, 4.0},
			},
			expectedValue: "4.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{2, 1},
				{4, 3.1},
			},
			expectedValue: "3.1000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{5, 100},
				{6, 100000.001},
			},
			expectedValue: "100000.0010000000",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			max := getMaxPointValue(tt.actualPoints)
			if max != tt.expectedValue {
				t.Fatal("max point value caculation is wrong.")
			}
		})
	}
}

func TestGetMinPointValue(t *testing.T) {
	tests := []struct {
		actualPoints  []monitoring.Point
		expectedValue string
	}{
		{
			actualPoints: []monitoring.Point{
				{1.0, 2.0},
				{3.0, 4.0},
			},
			expectedValue: "2.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{2, 1},
				{4, 3.1},
			},
			expectedValue: "1.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{5, 100},
				{6, 100000.001},
			},
			expectedValue: "100.0000000000",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			max := getMinPointValue(tt.actualPoints)
			if max != tt.expectedValue {
				t.Fatal("min point value caculation is wrong.")
			}
		})
	}
}

func TestGetSumPointValue(t *testing.T) {
	tests := []struct {
		actualPoints  []monitoring.Point
		expectedValue string
	}{
		{
			actualPoints: []monitoring.Point{
				{1.0, 2.0},
				{3.0, 4.0},
			},
			expectedValue: "6.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{2, 1},
				{4, 3.1},
			},
			expectedValue: "4.1000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{5, 100},
				{6, 100000.001},
			},
			expectedValue: "100100.0010000000",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			max := getSumPointValue(tt.actualPoints)
			if max != tt.expectedValue {
				t.Fatal("sum point value caculation is wrong.")
			}
		})
	}
}

func TestGetAvgPointValue(t *testing.T) {
	tests := []struct {
		actualPoints  []monitoring.Point
		expectedValue string
	}{
		{
			actualPoints: []monitoring.Point{
				{1.0, 2.0},
				{3.0, 4.0},
			},
			expectedValue: "3.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{2, 1},
				{4, 3.1},
			},
			expectedValue: "2.0500000000",
		},
		{
			actualPoints: []monitoring.Point{
				{5, 100},
				{6, 100000.001},
			},
			expectedValue: "50050.0005000000",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			max := getAvgPointValue(tt.actualPoints)
			if max != tt.expectedValue {
				t.Fatal("avg point value caculattion is wrong.")
			}
		})
	}
}

func TestGenerateFloatFormat(t *testing.T) {
	format := generateFloatFormat(10)
	if format != "%.10f" {
		t.Fatalf("get currency float format failed, %s", format)
	}
}

func TestGetResourceUnit(t *testing.T) {

	tests := []struct {
		meterName     string
		expectedValue string
	}{
		{
			meterName:     "no-exist",
			expectedValue: "",
		},
		{
			meterName:     "meter_cluster_cpu_usage",
			expectedValue: "cores",
		},
	}
	for _, tt := range tests {
		if getResourceUnit(tt.meterName) != tt.expectedValue {
			t.Fatal("get resource unit failed")
		}
	}

}

func TestSquashPoints(t *testing.T) {

	tests := []struct {
		input    []monitoring.Point
		factor   int
		expected []monitoring.Point
	}{
		{
			input: []monitoring.Point{
				{1, 1},
				{2, 2},
				{3, 3},
				{4, 4},
				{5, 5},
				{6, 6},
				{7, 7},
				{8, 8},
			},
			factor: 1,
			expected: []monitoring.Point{
				{1, 1},
				{2, 2},
				{3, 3},
				{4, 4},
				{5, 5},
				{6, 6},
				{7, 7},
				{8, 8},
			},
		},
		{
			input: []monitoring.Point{
				{1, 1},
				{2, 2},
				{3, 3},
				{4, 4},
				{5, 5},
				{6, 6},
				{7, 7},
				{8, 8},
			},
			factor: 2,
			expected: []monitoring.Point{
				{2, 3},
				{4, 7},
				{6, 11},
				{8, 15},
			},
		},
	}

	for _, tt := range tests {
		got := squashPoints(tt.input, tt.factor)
		if diff := cmp.Diff(got, tt.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", tt.expected, diff)
		}
	}
}

func TestGetFeeWithMeterName(t *testing.T) {

	priceInfo := meteringclient.PriceInfo{
		IngressNetworkTrafficPerMegabytesPerHour: 1,
		EgressNetworkTrafficPerMegabytesPerHour:  2,
		CpuPerCorePerHour:                        3,
		MemPerGigabytesPerHour:                   4,
		PvcPerGigabytesPerHour:                   5,
		CurrencyUnit:                             "CNY",
	}

	if getFeeWithMeterName("meter_cluster_cpu_usage", "1", priceInfo) != "3.000" {
		t.Error("failed to get fee with meter_cluster_cpu_usage")
		return
	}
	if getFeeWithMeterName("meter_cluster_memory_usage", "0", priceInfo) != "0.000" {
		t.Error("failed to get fee with meter_cluster_memory_usage")
		return
	}
	if getFeeWithMeterName("meter_cluster_net_bytes_transmitted", "0", priceInfo) != "0.000" {
		t.Error("failed to get fee with meter_cluster_net_bytes_transmitted")
		return
	}
	if getFeeWithMeterName("meter_cluster_net_bytes_received", "0", priceInfo) != "0.000" {
		t.Error("failed to get fee with meter_cluster_net_bytes_received")
		return
	}
	if getFeeWithMeterName("meter_cluster_pvc_bytes_total", "0", priceInfo) != "0.000" {
		t.Error("failed to get fee with meter_cluster_pvc_bytes_total")
		return
	}
}

func TestUpdateMetricStatData(t *testing.T) {

	priceInfo := meteringclient.PriceInfo{
		IngressNetworkTrafficPerMegabytesPerHour: 1,
		EgressNetworkTrafficPerMegabytesPerHour:  2,
		CpuPerCorePerHour:                        3,
		MemPerGigabytesPerHour:                   4,
		PvcPerGigabytesPerHour:                   5,
		CurrencyUnit:                             "CNY",
	}

	tests := []struct {
		metric     monitoring.Metric
		scalingMap map[string]float64
		expected   monitoring.MetricData
	}{
		{
			metric: monitoring.Metric{
				MetricName: "test",
				MetricData: monitoring.MetricData{
					MetricType: monitoring.MetricTypeMatrix,
					MetricValues: []monitoring.MetricValue{
						{
							Metadata: map[string]string{},
							Series: []monitoring.Point{
								{1, 1},
								{2, 2},
							},
						},
					},
				},
			},
			scalingMap: map[string]float64{
				"test": 1,
			},
			expected: monitoring.MetricData{
				MetricType: monitoring.MetricTypeMatrix,
				MetricValues: []monitoring.MetricValue{
					{
						Metadata: map[string]string{},
						Series: []monitoring.Point{
							{1, 1},
							{2, 2},
						},
						MinValue:     "1.0000000000",
						MaxValue:     "2.0000000000",
						AvgValue:     "1.5000000000",
						SumValue:     "3.0000000000",
						CurrencyUnit: "CNY",
					},
				},
			},
		},
		{
			metric: monitoring.Metric{
				MetricName: "test",
				MetricData: monitoring.MetricData{
					MetricType: monitoring.MetricTypeVector,
					MetricValues: []monitoring.MetricValue{
						{
							Metadata: map[string]string{},
							Sample:   &monitoring.Point{1, 2},
						},
					},
				},
			},
			scalingMap: nil,
			expected: monitoring.MetricData{
				MetricType: monitoring.MetricTypeVector,
				MetricValues: []monitoring.MetricValue{
					{
						Metadata:     map[string]string{},
						Sample:       &monitoring.Point{1, 2},
						MinValue:     "2.0000000000",
						MaxValue:     "2.0000000000",
						AvgValue:     "2.0000000000",
						SumValue:     "2.0000000000",
						CurrencyUnit: "CNY",
					},
				},
			},
		},
	}

	for _, test := range tests {
		got := updateMetricStatData(test.metric, test.scalingMap, priceInfo)
		if diff := cmp.Diff(got, test.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
			return
		}
	}

}
