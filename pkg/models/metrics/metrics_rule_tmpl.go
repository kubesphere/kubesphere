/*
Copyright 2018 The KubeSphere Authors.
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

type MetricMap map[string]string

var recordingRuleTmplMap = MetricMap{
	//node:node_cpu_utilisation:avg1m{node=~"i-5xcldxos|i-6soe9zl1"}
	// node 后面都可以跟 {node=~""}
	"node_cpu_utilisation":    "node:node_cpu_utilisation:avg1m",
	"node_memory_utilisation": "node:node_memory_utilisation:",
	"node_memory_available":   "node:node_memory_bytes_available:sum",
	"node_memory_total":       "node:node_memory_bytes_total:sum",
	// cluster 后面都 不 可以跟 {cluster=~""}
	"cluster_cpu_utilisation":    ":node_cpu_utilisation:avg1m",
	"cluster_memory_utilisation": ":node_memory_utilisation:",
	// namespace  后面都可以跟 {namespace=~""}
	"namespace_cpu_utilisation":    "namespace:container_cpu_usage_seconds_total:sum_rate", // {namespace =~"kube-system|openpitrix-system|kubesphere-system"}
	"namespace_memory_utilisation": "namespace:container_memory_usage_bytes:sum",           //{namespace =~"kube-system|openpitrix-system|kubesphere-system"}
	// TODO
	"namespace_memory_utilisation_wo_cache": "namespace:container_memory_usage_bytes_wo_cache:sum", // {namespace =~"kube-system|openpitrix-system|kubesphere-system"}
}

var promqlTempMap = MetricMap{
	// pod
	"pod_cpu_utilisation": `sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name="$2", image!=""}[5m])) by (namespace, pod_name)`,
	"pod_memory_utilisation":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name="$2", image!=""}) by (namespace, pod_name)`,
	"pod_memory_utilisation_wo_cache": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name="$2", image!=""} - container_memory_cache{job="kubelet", namespace="monitoring", pod_name=~"$2",image!=""}) by (namespace, pod_name)`,

	"pod_cpu_utilisation_all": `sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name=~"$2", image!=""}[5m])) by (namespace, pod_name)`,
	"pod_memory_utilisation_all":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name=~"$2",  image!=""}) by (namespace, pod_name)`,
	"pod_memory_utilisation_wo_cache_all": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name=~"$2", image!=""} - container_memory_cache{job="kubelet", namespace="monitoring", pod_name=~"$2", image!=""}) by (namespace, pod_name)`,

	//"pod_cpu_utilisation_node":  `sum by (node, pod) (label_join(irate(container_cpu_usage_seconds_total{job="kubelet", image!=""}[5m]), "pod", " ", "pod_name") * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,
	"pod_cpu_utilisation_node":  `sum by (node, pod) (label_join(irate(container_cpu_usage_seconds_total{job="kubelet",pod_name=~"$2", image!=""}[5m]), "pod", " ", "pod_name") * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,
	"pod_memory_utilisation_node":  `sum by (node, pod) (label_join(container_memory_usage_bytes{job="kubelet",pod_name=~"$2", image!=""}, "pod", " ", "pod_name") * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,
	"pod_memory_utilisation_wo_cache_node": `sum by (node, pod) ((label_join(container_memory_usage_bytes{job="kubelet",pod_name=~"$2",  image!=""}, "pod", " ", "pod_name") - label_join(container_memory_cache{job="kubelet",pod_name=~"$2", image!=""}, "pod", " ", "pod_name")) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,

	// container
	"container_cpu_utilisation":            `sum(irate(container_cpu_usage_seconds_total{namespace="$1", pod_name="$2", container_name="$3"}[5m])) by (namespace, pod_name, container_name)`,
	//"container_cpu_utilisation_wo_podname": `sum(irate(container_cpu_usage_seconds_total{namespace="$1", container_name=~"$3"}[5m])) by (namespace, pod_name, container_name)`,
	"container_cpu_utilisation_all": `sum(irate(container_cpu_usage_seconds_total{namespace="$1", pod_name="$2", container_name!="POD"}[5m])) by (namespace, pod_name, container_name)`,
	//"container_cpu_utilisation_all_wo_podname": `sum(irate(container_cpu_usage_seconds_total{namespace="$1", container_name!="POD"}[5m])) by (namespace, pod_name, container_name)`,

	"container_memory_utilisation_wo_cache":     `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name="$3"} - ignoring(id, image, endpoint, instance, job, name, service) container_memory_cache{namespace="$1", pod_name="$2", container_name="$3"}`,
	"container_memory_utilisation_wo_cache_all": `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name=~"$3"} - ignoring(id, image, endpoint, instance, job, name, service) container_memory_cache{namespace="$1", pod_name="$2", container_name=~"$3"}`,
	"container_memory_utilisation":               `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name="$3"}`,
	"container_memory_utilisation_all":           `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name=~"$3"}`,
}
