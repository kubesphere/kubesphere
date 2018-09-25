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

var MetricsNames = []string{
	"cluster_cpu_utilisation",
	"cluster_memory_utilisation",
	"cluster_net_utilisation",
	"cluster_pod_count",

	"node_cpu_utilisation",
	"node_memory_utilisation",
	"node_memory_available",
	"node_memory_total",
	"node_net_utilisation",
	"node_net_bytes_transmitted",
	"node_net_bytes_received",
	"node_disk_read_iops",
	"node_disk_write_iops",
	"node_disk_read_throughput",
	"node_disk_write_throughput",
	"node_disk_capacity",
	"node_disk_available",
	"node_disk_utilization",

	"namespace_cpu_utilisation",
	"namespace_memory_utilisation",
	"namespace_memory_utilisation_wo_cache",
	"namespace_net_bytes_transmitted",
	"namespace_net_bytes_received",
	"namespace_pod_count",

	"pod_cpu_utilisation",
	"pod_memory_utilisation",
	"pod_memory_utilisation_wo_cache",
	"pod_net_bytes_transmitted",
	"pod_net_bytes_received",

	"workload_pod_cpu_utilisation",
	"workload_pod_memory_utilisation",
	"workload_pod_memory_utilisation_wo_cache",
	"workload_pod_net_bytes_transmitted",
	"workload_pod_net_bytes_received",
	//"container_cpu_utilisation",
	//"container_memory_utilisation_wo_cache",
	//"container_memory_utilisation",

	"tenant_cpu_utilisation",
	"tenant_memory_utilisation",
	"tenant_memory_utilisation_wo_cache",
	"tenant_net_bytes_transmitted",
	"tenant_net_bytes_received",
	"tenant_pod_count",
}

var RulePromQLTmplMap = MetricMap{
	//cluster
	"cluster_cpu_utilisation":    ":node_cpu_utilisation:avg1m",
	"cluster_memory_utilisation": ":node_memory_utilisation:",
	// Cluster network utilisation (bytes received + bytes transmitted per second)
	"cluster_net_utilisation": ":node_net_utilisation:sum_irate",
	"cluster_pod_count":       `count(kube_pod_info{job="kube-state-metrics"})`,

	//node
	"node_cpu_utilisation":    "node:node_cpu_utilisation:avg1m",
	"node_memory_utilisation": "node:node_memory_utilisation:",
	"node_memory_available":   "node:node_memory_bytes_available:sum",
	"node_memory_total":       "node:node_memory_bytes_total:sum",
	// Node network utilisation (bytes received + bytes transmitted per second)
	"node_net_utilisation": "node:node_net_utilisation:sum_irate",
	// Node network bytes transmitted per second
	"node_net_bytes_transmitted": "node:node_net_bytes_transmitted:sum_irate",
	// Node network bytes received per second
	"node_net_bytes_received": "node:node_net_bytes_received:sum_irate",

	// node:data_volume_iops_reads:sum{node=~"i-5xcldxos|i-6soe9zl1"}
	"node_disk_read_iops": "node:data_volume_iops_reads:sum",
	// node:data_volume_iops_writes:sum{node=~"i-5xcldxos|i-6soe9zl1"}
	"node_disk_write_iops": "node:data_volume_iops_writes:sum",
	// node:data_volume_throughput_bytes_read:sum{node=~"i-5xcldxos|i-6soe9zl1"}
	"node_disk_read_throughput": "node:data_volume_throughput_bytes_read:sum",
	// node:data_volume_throughput_bytes_written:sum{node=~"i-5xcldxos|i-6soe9zl1"}
	"node_disk_write_throughput": "node:data_volume_throughput_bytes_written:sum",

	"node_disk_capacity":    `sum by (node) ((node_filesystem_avail{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1)`,
	"node_disk_available":   `sum by (node) ((node_filesystem_avail{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1)`,
	"node_disk_utilization": `sum by (node) (((node_filesystem_size{mountpoint="/", job="node-exporter"} - node_filesystem_avail{mountpoint="/", job="node-exporter"}) / node_filesystem_size{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1)`,

	//namespace
	"namespace_cpu_utilisation":             `namespace:container_cpu_usage_seconds_total:sum_rate{namespace=~"$1"}`,
	"namespace_memory_utilisation":          `namespace:container_memory_usage_bytes:sum{namespace=~"$1"}`,
	"namespace_memory_utilisation_wo_cache": `namespace:container_memory_usage_bytes_wo_cache:sum{namespace=~"$1"}`,
	"namespace_net_bytes_transmitted":       `sum by (namespace) (irate(container_network_transmit_bytes_total{namespace=~"$1", pod_name!="", interface="eth0", job="kubelet"}[2m]))`,
	"namespace_net_bytes_received":          `sum by (namespace) (irate(container_network_receive_bytes_total{namespace=~"$1", pod_name!="", interface="eth0", job="kubelet"}[2m]))`,
	// count(kube_pod_info) by (namespace)    namespace=~"monitoring|default|kube-system"
	"namespace_pod_count": `count(kube_pod_info{job="kube-state-metrics", namespace=~"$1"}) by (namespace)`,

	// pod
	"pod_cpu_utilisation":             `sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name="$2", image!=""}[5m])) by (namespace, pod_name)`,
	"pod_memory_utilisation":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name="$2", image!=""}) by (namespace, pod_name)`,
	"pod_memory_utilisation_wo_cache": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name="$2", image!=""} - container_memory_cache{job="kubelet", namespace="$1", pod_name="$2",image!=""}) by (namespace, pod_name)`,
	"pod_net_bytes_transmitted":       `sum by (namespace, pod_name) (irate(container_network_transmit_bytes_total{namespace="$1", pod_name!="", pod_name="$2", interface="eth0", job="kubelet"}[2m]))`,
	"pod_net_bytes_received":          `sum by (namespace, pod_name) (irate(container_network_receive_bytes_total{namespace="$1", pod_name!="", pod_name="$2", interface="eth0", job="kubelet"}[2m]))`,

	"pod_cpu_utilisation_all":             `sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name=~"$2", image!=""}[5m])) by (namespace, pod_name)`,
	"pod_memory_utilisation_all":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name=~"$2",  image!=""}) by (namespace, pod_name)`,
	"pod_memory_utilisation_wo_cache_all": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name=~"$2", image!=""} - container_memory_cache{job="kubelet", namespace="$1", pod_name=~"$2", image!=""}) by (namespace, pod_name)`,
	"pod_net_bytes_transmitted_all":       `sum by (namespace, pod_name) (irate(container_network_transmit_bytes_total{namespace="$1", pod_name!="", pod_name=~"$2", interface="eth0", job="kubelet"}[2m]))`,
	"pod_net_bytes_received_all":          `sum by (namespace, pod_name) (irate(container_network_receive_bytes_total{namespace="$1", pod_name!="", pod_name=~"$2", interface="eth0", job="kubelet"}[2m]))`,

	//"pod_cpu_utilisation_node":  `sum by (node, pod) (label_join(irate(container_cpu_usage_seconds_total{job="kubelet", image!=""}[5m]), "pod", " ", "pod_name") * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,
	"pod_cpu_utilisation_node":             `sum by (node, pod) (label_join(irate(container_cpu_usage_seconds_total{job="kubelet",pod_name=~"$2", image!=""}[5m]), "pod", " ", "pod_name") * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,
	"pod_memory_utilisation_node":          `sum by (node, pod) (label_join(container_memory_usage_bytes{job="kubelet",pod_name=~"$2", image!=""}, "pod", " ", "pod_name") * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,
	"pod_memory_utilisation_wo_cache_node": `sum by (node, pod) ((label_join(container_memory_usage_bytes{job="kubelet",pod_name=~"$2",  image!=""}, "pod", " ", "pod_name") - label_join(container_memory_cache{job="kubelet",pod_name=~"$2", image!=""}, "pod", " ", "pod_name")) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,

	// container
	"container_cpu_utilisation": `sum(irate(container_cpu_usage_seconds_total{namespace="$1", pod_name="$2", container_name="$3"}[5m])) by (namespace, pod_name, container_name)`,
	//"container_cpu_utilisation_wo_podname": `sum(irate(container_cpu_usage_seconds_total{namespace="$1", container_name=~"$3"}[5m])) by (namespace, pod_name, container_name)`,
	"container_cpu_utilisation_all": `sum(irate(container_cpu_usage_seconds_total{namespace="$1", pod_name="$2", container_name=~"$3", container_name!="POD"}[5m])) by (namespace, pod_name, container_name)`,
	//"container_cpu_utilisation_all_wo_podname": `sum(irate(container_cpu_usage_seconds_total{namespace="$1", container_name!="POD"}[5m])) by (namespace, pod_name, container_name)`,

	"container_memory_utilisation_wo_cache":     `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name="$3"} - ignoring(id, image, endpoint, instance, job, name, service) container_memory_cache{namespace="$1", pod_name="$2", container_name="$3"}`,
	"container_memory_utilisation_wo_cache_all": `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name=~"$3", container_name!="POD"} - ignoring(id, image, endpoint, instance, job, name, service) container_memory_cache{namespace="$1", pod_name="$2", container_name=~"$3", container_name!="POD"}`,
	"container_memory_utilisation":              `container_memory_usage_bytes{namespace="$1", pod_name="$2",  container_name="$3"}`,
	"container_memory_utilisation_all":          `container_memory_usage_bytes{namespace="$1", pod_name="$2",  container_name=~"$3", container_name!="POD"}`,

	// tenant
	"tenant_cpu_utilisation":             `sum(namespace:container_cpu_usage_seconds_total:sum_rate{namespace =~"$1"})`,
	"tenant_memory_utilisation":          `sum(namespace:container_memory_usage_bytes:sum{namespace =~"$1"})`,
	"tenant_memory_utilisation_wo_cache": `sum(namespace:container_memory_usage_bytes_wo_cache:sum{namespace =~"$1"})`,
	"tenant_net_bytes_transmitted":       `sum(sum by (namespace) (irate(container_network_transmit_bytes_total{namespace=~"$1", pod_name!="", interface="eth0", job="kubelet"}[2m])))`,
	"tenant_net_bytes_received":          `sum(sum by (namespace) (irate(container_network_receive_bytes_total{namespace=~"$1", pod_name!="", interface="eth0", job="kubelet"}[2m])))`,
	"tenant_pod_count":                   `sum(count(kube_pod_info{job="kube-state-metrics", namespace=~"$1"}) by (namespace))`,
}
