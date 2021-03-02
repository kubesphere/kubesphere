package monitoring

import (
	"math"
	"os"

	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

const (
	METER_RESOURCE_TYPE_CPU = iota
	METER_RESOURCE_TYPE_MEM
	METER_RESOURCE_TYPE_NET_INGRESS
	METER_RESOURCE_TYPE_NET_EGRESS
	METER_RESOURCE_TYPE_PVC

	meteringConfig = "/etc/kubesphere/metering/ks-metering.yaml"
)

var meterResourceUnitMap = map[int]string{
	METER_RESOURCE_TYPE_CPU:         "cores",
	METER_RESOURCE_TYPE_MEM:         "bytes",
	METER_RESOURCE_TYPE_NET_INGRESS: "bytes",
	METER_RESOURCE_TYPE_NET_EGRESS:  "bytes",
	METER_RESOURCE_TYPE_PVC:         "bytes",
}

var MeterResourceMap = map[string]int{
	"meter_cluster_cpu_usage":                 METER_RESOURCE_TYPE_CPU,
	"meter_cluster_memory_usage":              METER_RESOURCE_TYPE_MEM,
	"meter_cluster_net_bytes_transmitted":     METER_RESOURCE_TYPE_NET_EGRESS,
	"meter_cluster_net_bytes_received":        METER_RESOURCE_TYPE_NET_INGRESS,
	"meter_cluster_pvc_bytes_total":           METER_RESOURCE_TYPE_PVC,
	"meter_node_cpu_usage":                    METER_RESOURCE_TYPE_CPU,
	"meter_node_memory_usage_wo_cache":        METER_RESOURCE_TYPE_MEM,
	"meter_node_net_bytes_transmitted":        METER_RESOURCE_TYPE_NET_EGRESS,
	"meter_node_net_bytes_received":           METER_RESOURCE_TYPE_NET_INGRESS,
	"meter_node_pvc_bytes_total":              METER_RESOURCE_TYPE_PVC,
	"meter_workspace_cpu_usage":               METER_RESOURCE_TYPE_CPU,
	"meter_workspace_memory_usage":            METER_RESOURCE_TYPE_MEM,
	"meter_workspace_net_bytes_transmitted":   METER_RESOURCE_TYPE_NET_EGRESS,
	"meter_workspace_net_bytes_received":      METER_RESOURCE_TYPE_NET_INGRESS,
	"meter_workspace_pvc_bytes_total":         METER_RESOURCE_TYPE_PVC,
	"meter_namespace_cpu_usage":               METER_RESOURCE_TYPE_CPU,
	"meter_namespace_memory_usage_wo_cache":   METER_RESOURCE_TYPE_MEM,
	"meter_namespace_net_bytes_transmitted":   METER_RESOURCE_TYPE_NET_EGRESS,
	"meter_namespace_net_bytes_received":      METER_RESOURCE_TYPE_NET_INGRESS,
	"meter_namespace_pvc_bytes_total":         METER_RESOURCE_TYPE_PVC,
	"meter_application_cpu_usage":             METER_RESOURCE_TYPE_CPU,
	"meter_application_memory_usage_wo_cache": METER_RESOURCE_TYPE_MEM,
	"meter_application_net_bytes_transmitted": METER_RESOURCE_TYPE_NET_EGRESS,
	"meter_application_net_bytes_received":    METER_RESOURCE_TYPE_NET_INGRESS,
	"meter_application_pvc_bytes_total":       METER_RESOURCE_TYPE_PVC,
	"meter_workload_cpu_usage":                METER_RESOURCE_TYPE_CPU,
	"meter_workload_memory_usage_wo_cache":    METER_RESOURCE_TYPE_MEM,
	"meter_workload_net_bytes_transmitted":    METER_RESOURCE_TYPE_NET_EGRESS,
	"meter_workload_net_bytes_received":       METER_RESOURCE_TYPE_NET_INGRESS,
	"meter_workload_pvc_bytes_total":          METER_RESOURCE_TYPE_PVC,
	"meter_service_cpu_usage":                 METER_RESOURCE_TYPE_CPU,
	"meter_service_memory_usage_wo_cache":     METER_RESOURCE_TYPE_MEM,
	"meter_service_net_bytes_transmitted":     METER_RESOURCE_TYPE_NET_EGRESS,
	"meter_service_net_bytes_received":        METER_RESOURCE_TYPE_NET_INGRESS,
	"meter_pod_cpu_usage":                     METER_RESOURCE_TYPE_CPU,
	"meter_pod_memory_usage_wo_cache":         METER_RESOURCE_TYPE_MEM,
	"meter_pod_net_bytes_transmitted":         METER_RESOURCE_TYPE_NET_EGRESS,
	"meter_pod_net_bytes_received":            METER_RESOURCE_TYPE_NET_INGRESS,
	"meter_pod_pvc_bytes_total":               METER_RESOURCE_TYPE_PVC,
}

type PriceInfo struct {
	CpuPerCorePerHour                         float64 `json:"cpuPerCorePerHour" yaml:"cpuPerCorePerHour"`
	MemPerGigabytesPerHour                    float64 `json:"memPerGigabytesPerHour" yaml:"memPerGigabytesPerHour"`
	IngressNetworkTrafficPerGiagabytesPerHour float64 `json:"ingressNetworkTrafficPerGiagabytesPerHour" yaml:"ingressNetworkTrafficPerGiagabytesPerHour"`
	EgressNetworkTrafficPerGigabytesPerHour   float64 `json:"egressNetworkTrafficPerGigabytesPerHour" yaml:"egressNetworkTrafficPerGigabytesPerHour"`
	PvcPerGigabytesPerHour                    float64 `json:"pvcPerGigabytesPerHour" yaml:"pvcPerGigabytesPerHour"`
	CurrencyUnit                              string  `json:"currencyUnit" yaml:"currencyUnit"`
}

type Billing struct {
	PriceInfo PriceInfo `json:"priceInfo" yaml:"priceInfo"`
}

type MeterConfig struct {
	Billing Billing `json:"billing" yaml:"billing"`
}

func (mc MeterConfig) GetPriceInfo() PriceInfo {
	return mc.Billing.PriceInfo
}

func LoadYaml() (*MeterConfig, error) {

	var meterConfig MeterConfig

	mf, err := os.Open(meteringConfig)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if err = yaml.NewYAMLOrJSONDecoder(mf, 1024).Decode(&meterConfig); err != nil {
		klog.Error(err)
		return nil, err
	}

	return &meterConfig, nil
}

func getMaxPointValue(points []monitoring.Point) float64 {
	var max float64
	for i, p := range points {
		if i == 0 {
			max = p.Value()
		}

		if p.Value() > max {
			max = p.Value()
		}
	}

	return max
}

func getMinPointValue(points []monitoring.Point) float64 {
	var min float64
	for i, p := range points {
		if i == 0 {
			min = p.Value()
		}

		if p.Value() < min {
			min = p.Value()
		}
	}

	return min
}

func getSumPointValue(points []monitoring.Point) float64 {
	avg := 0.0

	for _, p := range points {
		avg += p.Value()
	}

	return avg
}

func getAvgPointValue(points []monitoring.Point) float64 {
	return getSumPointValue(points) / float64(len(points))
}

func getCurrencyUnit() string {
	meterConfig, err := LoadYaml()
	if err != nil {
		klog.Error(err)
		return ""
	}

	return meterConfig.GetPriceInfo().CurrencyUnit
}

func getResourceUnit(meterName string) string {
	if resourceType, ok := MeterResourceMap[meterName]; !ok {
		klog.Errorf("invlaid meter %v", meterName)
		return ""
	} else {
		return meterResourceUnitMap[resourceType]
	}
}

func getFeeWithMeterName(meterName string, sum float64) float64 {

	meterConfig, err := LoadYaml()
	if err != nil {
		klog.Error(err)
		return -1
	}
	priceInfo := meterConfig.GetPriceInfo()

	if resourceType, ok := MeterResourceMap[meterName]; !ok {
		klog.Errorf("invlaid meter %v", meterName)
		return -1
	} else {
		switch resourceType {
		case METER_RESOURCE_TYPE_CPU:
			// unit: core, precision: 0.001
			sum = math.Round(sum*1000) / 1000
			return priceInfo.CpuPerCorePerHour * sum
		case METER_RESOURCE_TYPE_MEM:
			// unit: Gigabyte, precision: 0.1
			sum = math.Round(sum/1073741824*10) / 10
			return priceInfo.MemPerGigabytesPerHour * sum
		case METER_RESOURCE_TYPE_NET_INGRESS:
			// unit: Megabyte, precision: 1
			sum = math.Round(sum / 1048576)
			return priceInfo.IngressNetworkTrafficPerGiagabytesPerHour * sum
		case METER_RESOURCE_TYPE_NET_EGRESS:
			// unit: Megabyte, precision:
			sum = math.Round(sum / 1048576)
			return priceInfo.EgressNetworkTrafficPerGigabytesPerHour * sum
		case METER_RESOURCE_TYPE_PVC:
			// unit: Gigabyte, precision: 0.1
			sum = math.Round(sum/1073741824*10) / 10
			return priceInfo.PvcPerGigabytesPerHour * sum
		}

		return -1
	}
}

func updateMetricStatData(metric monitoring.Metric, scalingMap map[string]float64) monitoring.MetricData {
	metricName := metric.MetricName
	metricData := metric.MetricData
	for index, metricValue := range metricData.MetricValues {

		var points []monitoring.Point
		if metricData.MetricType == monitoring.MetricTypeMatrix {
			points = metricValue.Series
		} else {
			points = append(points, *metricValue.Sample)
		}

		var factor float64 = 1
		if scalingMap != nil {
			factor = scalingMap[metricName]
		}

		if len(points) == 1 {
			sample := points[0]
			sum := sample[1] * factor
			metricData.MetricValues[index].MinValue = sample[1]
			metricData.MetricValues[index].MaxValue = sample[1]
			metricData.MetricValues[index].AvgValue = sample[1]
			metricData.MetricValues[index].SumValue = sum
			metricData.MetricValues[index].Fee = getFeeWithMeterName(metricName, sum)
		} else {
			sum := getSumPointValue(points) * factor
			metricData.MetricValues[index].MinValue = getMinPointValue(points)
			metricData.MetricValues[index].MaxValue = getMaxPointValue(points)
			metricData.MetricValues[index].AvgValue = getAvgPointValue(points)
			metricData.MetricValues[index].SumValue = sum
			metricData.MetricValues[index].Fee = getFeeWithMeterName(metricName, sum)
		}
		metricData.MetricValues[index].CurrencyUnit = getCurrencyUnit()
		metricData.MetricValues[index].ResourceUnit = getResourceUnit(metricName)

	}
	return metricData
}
