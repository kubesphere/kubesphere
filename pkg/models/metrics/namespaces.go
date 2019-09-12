/*

 Copyright 2019 The KubeSphere Authors.

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
package metrics

import (
	"net/url"
	"strings"

	"k8s.io/api/core/v1"
)

func GetNamespacesWithMetrics(namespaces []*v1.Namespace) []*v1.Namespace {
	var nsNameList []string
	for i := range namespaces {
		nsNameList = append(nsNameList, namespaces[i].Name)
	}
	nsFilter := "^(" + strings.Join(nsNameList, "|") + ")$"
	var timeRelateParams = make(url.Values)

	params := MonitoringRequestParams{
		ResourcesFilter: nsFilter,
		Params:          timeRelateParams,
		QueryType:       DefaultQueryType,
		MetricsFilter:   "namespace_cpu_usage|namespace_memory_usage_wo_cache|namespace_pod_count",
	}

	rawMetrics := GetNamespaceLevelMetrics(&params)

	for _, result := range rawMetrics.Results {
		for _, data := range result.Data.Result {
			metricDescMap, ok := data[ResultItemMetric].(map[string]interface{})
			if ok {
				if ns, exist := metricDescMap[ResultItemMetricResourceName]; exist {
					timeAndValue, ok := data[ResultItemValue].([]interface{})
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
