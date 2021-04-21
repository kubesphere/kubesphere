package monitoring

import (
	"fmt"
	"math/big"

	"k8s.io/klog"

	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

const (
	METER_RESOURCE_TYPE_CPU = iota
	METER_RESOURCE_TYPE_MEM
	METER_RESOURCE_TYPE_NET_INGRESS
	METER_RESOURCE_TYPE_NET_EGRESS
	METER_RESOURCE_TYPE_PVC

	meteringConfigDir  = "/etc/kubesphere/metering/"
	meteringConfigName = "ks-metering.yaml"

	meteringDefaultPrecision = 10
	meteringFeePrecision     = 3
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

func getMaxPointValue(points []monitoring.Point) string {
	var max *big.Float
	for i, p := range points {
		if i == 0 {
			max = new(big.Float).SetFloat64(p.Value())
		}

		pf := new(big.Float).SetFloat64(p.Value())
		if pf.Cmp(max) == 1 {
			max = pf
		}
	}

	return fmt.Sprintf(generateFloatFormat(meteringDefaultPrecision), max)
}

func getMinPointValue(points []monitoring.Point) string {
	var min *big.Float
	for i, p := range points {
		if i == 0 {
			min = new(big.Float).SetFloat64(p.Value())
		}

		pf := new(big.Float).SetFloat64(p.Value())
		if min.Cmp(pf) == 1 {
			min = pf
		}
	}

	return fmt.Sprintf(generateFloatFormat(meteringDefaultPrecision), min)
}

func getSumPointValue(points []monitoring.Point) string {
	sum := new(big.Float).SetFloat64(0)

	for _, p := range points {
		pf := new(big.Float).SetFloat64(p.Value())
		sum.Add(sum, pf)
	}

	return fmt.Sprintf(generateFloatFormat(meteringDefaultPrecision), sum)
}

func getAvgPointValue(points []monitoring.Point) string {
	sum, ok := new(big.Float).SetString(getSumPointValue(points))
	if !ok {
		klog.Error("failed to parse big.Float")
		return ""
	}

	length := new(big.Float).SetFloat64(float64(len(points)))

	return fmt.Sprintf(generateFloatFormat(meteringDefaultPrecision), sum.Quo(sum, length))
}

func generateFloatFormat(precision int) string {
	return "%." + fmt.Sprintf("%d", precision) + "f"
}

func getResourceUnit(meterName string) string {
	if resourceType, ok := MeterResourceMap[meterName]; !ok {
		klog.Errorf("invlaid meter %v", meterName)
		return ""
	} else {
		return meterResourceUnitMap[resourceType]
	}
}

func getFeeWithMeterName(meterName string, sum string, priceInfo meteringclient.PriceInfo) string {

	s, ok := new(big.Float).SetString(sum)
	if !ok {
		klog.Error("failed to parse string to float")
		return ""
	}

	if resourceType, ok := MeterResourceMap[meterName]; !ok {
		klog.Errorf("invlaid meter %v", meterName)
		return ""
	} else {
		switch resourceType {
		case METER_RESOURCE_TYPE_CPU:
			CpuPerCorePerHour := new(big.Float).SetFloat64(priceInfo.CpuPerCorePerHour)
			tmp := s.Mul(s, CpuPerCorePerHour)

			return fmt.Sprintf(generateFloatFormat(meteringFeePrecision), tmp)
		case METER_RESOURCE_TYPE_MEM:
			oneGiga := new(big.Float).SetInt64(1073741824)
			MemPerGigabytesPerHour := new(big.Float).SetFloat64(priceInfo.MemPerGigabytesPerHour)

			// transform unit from bytes to Gigabytes
			s.Quo(s, oneGiga)

			return fmt.Sprintf(generateFloatFormat(meteringFeePrecision), s.Mul(s, MemPerGigabytesPerHour))
		case METER_RESOURCE_TYPE_NET_INGRESS:
			oneMega := new(big.Float).SetInt64(1048576)
			IngressNetworkTrafficPerMegabytesPerHour := new(big.Float).SetFloat64(priceInfo.IngressNetworkTrafficPerMegabytesPerHour)

			// transform unit from bytes to Migabytes
			s.Quo(s, oneMega)

			return fmt.Sprintf(generateFloatFormat(meteringFeePrecision), s.Mul(s, IngressNetworkTrafficPerMegabytesPerHour))
		case METER_RESOURCE_TYPE_NET_EGRESS:
			oneMega := new(big.Float).SetInt64(1048576)
			EgressNetworkTrafficPerMegabytesPerHour := new(big.Float).SetPrec(meteringFeePrecision).SetFloat64(priceInfo.EgressNetworkTrafficPerMegabytesPerHour)

			// transform unit from bytes to Migabytes
			s.Quo(s, oneMega)

			return fmt.Sprintf(generateFloatFormat(meteringFeePrecision), s.Mul(s, EgressNetworkTrafficPerMegabytesPerHour))
		case METER_RESOURCE_TYPE_PVC:
			oneGiga := new(big.Float).SetInt64(1073741824)
			PvcPerGigabytesPerHour := new(big.Float).SetPrec(meteringFeePrecision).SetFloat64(priceInfo.PvcPerGigabytesPerHour)

			// transform unit from bytes to Gigabytes
			s.Quo(s, oneGiga)

			return fmt.Sprintf(generateFloatFormat(meteringFeePrecision), s.Mul(s, PvcPerGigabytesPerHour))
		}

		return ""
	}
}

func updateMetricStatData(metric monitoring.Metric, scalingMap map[string]float64, priceInfo meteringclient.PriceInfo) monitoring.MetricData {
	metricName := metric.MetricName
	metricData := metric.MetricData
	for index, metricValue := range metricData.MetricValues {

		// calulate min, max, avg value first, then squash points with factor
		if metricData.MetricType == monitoring.MetricTypeMatrix {
			metricData.MetricValues[index].MinValue = getMinPointValue(metricValue.Series)
			metricData.MetricValues[index].MaxValue = getMaxPointValue(metricValue.Series)
			metricData.MetricValues[index].AvgValue = getAvgPointValue(metricValue.Series)
		} else {
			metricData.MetricValues[index].MinValue = getMinPointValue([]monitoring.Point{*metricValue.Sample})
			metricData.MetricValues[index].MaxValue = getMaxPointValue([]monitoring.Point{*metricValue.Sample})
			metricData.MetricValues[index].AvgValue = getAvgPointValue([]monitoring.Point{*metricValue.Sample})
		}

		// squash points if step is more than one hour and calculate sum and fee
		var factor float64 = 1
		if scalingMap != nil {
			factor = scalingMap[metricName]
		}
		metricData.MetricValues[index].Series = squashPoints(metricData.MetricValues[index].Series, int(factor))

		if metricData.MetricType == monitoring.MetricTypeMatrix {
			sum := getSumPointValue(metricData.MetricValues[index].Series)
			metricData.MetricValues[index].SumValue = sum
			metricData.MetricValues[index].Fee = getFeeWithMeterName(metricName, sum, priceInfo)
		} else {
			sum := getSumPointValue([]monitoring.Point{*metricValue.Sample})
			metricData.MetricValues[index].SumValue = sum
			metricData.MetricValues[index].Fee = getFeeWithMeterName(metricName, sum, priceInfo)
		}

		metricData.MetricValues[index].CurrencyUnit = priceInfo.CurrencyUnit
		metricData.MetricValues[index].ResourceUnit = getResourceUnit(metricName)

	}
	return metricData
}

func squashPoints(input []monitoring.Point, factor int) (output []monitoring.Point) {

	if factor <= 0 {
		klog.Errorln("factor should be positive")
		return nil
	}

	for i := 0; i < len(input); i++ {

		if i%factor == 0 {
			output = append([]monitoring.Point{input[len(input)-1-i]}, output...)
		} else {
			output[0] = output[0].Add(input[len(input)-1-i])
		}
	}

	return output
}
