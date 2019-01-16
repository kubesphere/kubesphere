package metrics

import (
	"net/url"
	"strings"

	"k8s.io/api/core/v1"

	"kubesphere.io/kubesphere/pkg/client"
)

func GetNamespacesWithMetrics(namespaces []*v1.Namespace) []*v1.Namespace {
	var nsNameList []string
	for i := range namespaces {
		nsNameList = append(nsNameList, namespaces[i].Name)
	}
	nsFilter := "^(" + strings.Join(nsNameList, "|") + ")$"
	var timeRelateParams = make(url.Values)

	params := client.MonitoringRequestParams{
		ResourcesFilter: nsFilter,
		Params:          timeRelateParams,
		QueryType:       client.DefaultQueryType,
		MetricsFilter:   "namespace_cpu_usage|namespace_memory_usage_wo_cache|namespace_pod_count",
	}

	rawMetrics := MonitorAllMetrics(&params, MetricLevelNamespace)

	for _, result := range rawMetrics.Results {
		for _, data := range result.Data.Result {
			metricDescMap, ok := data["metric"].(map[string]interface{})
			if ok {
				if ns, exist := metricDescMap["namespace"]; exist {
					timeAndValue, ok := data["value"].([]interface{})
					if ok && len(timeAndValue) == 2 {
						for i := 0; i < len(namespaces); i++ {
							if namespaces[i].Name == ns {
								if namespaces[i].Annotations == nil {
									namespaces[i].Annotations = make(map[string]string, 0)
								}
								namespaces[i].Annotations[result.MetricName] = timeAndValue[1].(string)
							}
						}
					}
				}
			}
		}
	}

	return namespaces
}
