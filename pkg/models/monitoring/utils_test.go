package monitoring

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"

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

func saveTestConfig(t *testing.T, conf *MeterConfig) {
	content, err := yaml.Marshal(conf)
	if err != nil {
		t.Fatalf("error marshal config. %v", err)
	}

	err = ioutil.WriteFile(meteringConfigName, content, 0640)
	if err != nil {
		t.Fatalf("error write configuration file, %v", err)
	}
}

func cleanTestConfig(t *testing.T) {

	if _, err := os.Stat(meteringConfigName); os.IsNotExist(err) {
		t.Log("file not exists, skipping")
		return
	}

	err := os.Remove(meteringConfigName)
	if err != nil {
		t.Fatalf("remove %s file failed", meteringConfigName)
	}

}

func TestGetCurrencyUnit(t *testing.T) {
	if getCurrencyUnit() != "" {
		t.Fatal("currency unit should be empty")
	}

	saveTestConfig(t, &MeterConfig{
		Billing: Billing{
			PriceInfo: PriceInfo{
				IngressNetworkTrafficPerMegabytesPerHour: 1,
				EgressNetworkTrafficPerMegabytesPerHour:  2,
				CpuPerCorePerHour:                        3,
				MemPerGigabytesPerHour:                   4,
				PvcPerGigabytesPerHour:                   5,
				CurrencyUnit:                             "CNY",
			},
		},
	})
	defer cleanTestConfig(t)

	if getCurrencyUnit() != "CNY" {
		t.Fatal("failed to get currency unit from config")
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

	saveTestConfig(t, &MeterConfig{
		Billing: Billing{
			PriceInfo: PriceInfo{
				IngressNetworkTrafficPerMegabytesPerHour: 1,
				EgressNetworkTrafficPerMegabytesPerHour:  2,
				CpuPerCorePerHour:                        3,
				MemPerGigabytesPerHour:                   4,
				PvcPerGigabytesPerHour:                   5,
				CurrencyUnit:                             "CNY",
			},
		},
	})
	defer cleanTestConfig(t)

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
		got := updateMetricStatData(test.metric, test.scalingMap)
		if diff := cmp.Diff(got, test.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
			return
		}
	}

}
