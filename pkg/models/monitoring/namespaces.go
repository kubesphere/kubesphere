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

package monitoring

import "k8s.io/api/core/v1"

// TODO(wansir): Can we decouple this part from monitoring module, since the project structure has been changed
func GetNamespacesWithMetrics(namespaces []*v1.Namespace) []*v1.Namespace {
	//	var nsNameList []string
	//	for i := range namespaces {
	//		nsNameList = append(nsNameList, namespaces[i].Name)
	//	}
	//	nsFilter := "^(" + strings.Join(nsNameList, "|") + ")$"
	//
	//	now := time.Now()
	//	opt := &monitoring.QueryOptions{
	//		Level:           monitoring.MetricsLevelNamespace,
	//		ResourcesFilter: nsFilter,
	//		Start:           now,
	//		End:             now,
	//		MetricsFilter:   "namespace_cpu_usage|namespace_memory_usage_wo_cache|namespace_pod_count",
	//	}
	//
	//	gm, err := monitoring.Get(opt)
	//	if err != nil {
	//		klog.Error(err)
	//		return namespaces
	//	}
	//
	//	for _, m := range gm.Results {
	//		for _, v := range m.Data.MetricsValues {
	//			ns, exist := v.Metadata["namespace"]
	//			if !exist {
	//				continue
	//			}
	//
	//			for _, item := range namespaces {
	//				if item.Name == ns {
	//					if item.Annotations == nil {
	//						item.Annotations = make(map[string]string, 0)
	//					}
	//					item.Annotations[m.MetricsName] = strconv.FormatFloat(v.Sample[1], 'f', -1, 64)
	//				}
	//			}
	//		}
	//	}
	//
	return namespaces
}
