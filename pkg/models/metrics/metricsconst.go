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

const (
	ResultTypeVector         = "vector"
	ResultTypeMatrix         = "matrix"
	MetricStatusError        = "error"
	MetricStatusSuccess      = "success"
	ResultItemMetric         = "metric"
	ResultItemMetricResource = "resource"
	ResultItemValue          = "value"
)

const (
	MetricNameWorkloadCount     = "workload_count"
	MetricNameNamespacePodCount = "namespace_pod_count"

	MetricNameWorkspaceAllOrganizationCount = "workspace_all_organization_count"
	MetricNameWorkspaceAllAccountCount      = "workspace_all_account_count"
	MetricNameWorkspaceAllProjectCount      = "workspace_all_project_count"
	MetricNameWorkspaceAllDevopsCount       = "workspace_all_devops_project_count"

	MetricNameWorkspaceNamespaceCount = "workspace_namespace_count"
	MetricNameWorkspaceDevopsCount    = "workspace_devops_project_count"
	MetricNameWorkspaceMemberCount    = "workspace_member_count"
	MetricNameWorkspaceRoleCount      = "workspace_role_count"

	MetricNameClusterHealthyNodeCount   = "cluster_node_online"
	MetricNameClusterUnhealthyNodeCount = "cluster_node_offline"
	MetricNameClusterNodeCount          = "cluster_node_total"
)

const (
	WorkspaceResourceKindOrganization = "organization"
	WorkspaceResourceKindAccount      = "account"
	WorkspaceResourceKindNamespace    = "namespace"
	WorkspaceResourceKindDevops       = "devops"
	WorkspaceResourceKindMember       = "member"
	WorkspaceResourceKindRole         = "role"
)

const (
	MetricLevelCluster   = "cluster"
	MetricLevelNode      = "node"
	MetricLevelWorkspace = "workspace"
	MetricLevelNamespace = "namespace"
	MetricLevelPod       = "pod"
	MetricLevelContainer = "container"
	MetricLevelWorkload  = "workload"
)

type MetricMap map[string]string

var MetricsNames = []string{
	"cluster_cpu_utilisation",
	"cluster_cpu_usage",
	"cluster_cpu_total",
	"cluster_memory_utilisation",
	"cluster_pod_count",
	"cluster_memory_bytes_available",
	"cluster_memory_bytes_total",
	"cluster_memory_bytes_usage",
	"cluster_net_utilisation",
	"cluster_net_bytes_transmitted",
	"cluster_net_bytes_received",
	"cluster_disk_read_iops",
	"cluster_disk_write_iops",
	"cluster_disk_read_throughput",
	"cluster_disk_write_throughput",
	"cluster_disk_size_usage",
	"cluster_disk_size_utilisation",
	"cluster_disk_size_capacity",
	"cluster_disk_size_available",
	"cluster_node_online",
	"cluster_node_offline",
	"cluster_node_total",

	"node_cpu_utilisation",
	"node_cpu_total",
	"node_cpu_usage",
	"node_memory_utilisation",
	"node_memory_bytes_usage",
	"node_memory_bytes_available",
	"node_memory_bytes_total",
	"node_net_utilisation",
	"node_net_bytes_transmitted",
	"node_net_bytes_received",
	"node_disk_read_iops",
	"node_disk_write_iops",
	"node_disk_read_throughput",
	"node_disk_write_throughput",
	"node_disk_size_capacity",
	"node_disk_size_available",
	"node_disk_size_usage",
	"node_disk_size_utilisation",
	"node_pod_count",
	"node_pod_quota",

	"namespace_cpu_usage",
	"namespace_memory_usage",
	"namespace_memory_usage_wo_cache",
	"namespace_net_bytes_transmitted",
	"namespace_net_bytes_received",
	"namespace_pod_count",

	"pod_cpu_usage",
	"pod_memory_usage",
	"pod_memory_usage_wo_cache",
	"pod_net_bytes_transmitted",
	"pod_net_bytes_received",

	"workload_pod_cpu_usage",
	"workload_pod_memory_usage",
	"workload_pod_memory_usage_wo_cache",
	"workload_pod_net_bytes_transmitted",
	"workload_pod_net_bytes_received",
	//"container_cpu_usage",
	//"container_memory_usage_wo_cache",
	//"container_memory_usage",

	"workspace_cpu_usage",
	"workspace_memory_usage",
	"workspace_memory_usage_wo_cache",
	"workspace_net_bytes_transmitted",
	"workspace_net_bytes_received",
	"workspace_pod_count",
}

var RulePromQLTmplMap = MetricMap{
	//cluster
	"cluster_cpu_utilisation":        ":node_cpu_utilisation:avg1m",
	"cluster_cpu_usage":              `sum (irate(container_cpu_usage_seconds_total{job="kubelet", image!=""}[5m]))`,
	"cluster_cpu_total":              "sum(node:node_num_cpu:sum)",
	"cluster_memory_utilisation":     ":node_memory_utilisation:",
	"cluster_pod_count":              `count(kube_pod_info unless on(pod) kube_pod_completion_time unless on(node) kube_node_labels{label_role="log"})`,
	"cluster_memory_bytes_available": "sum(node:node_memory_bytes_available:sum)",
	"cluster_memory_bytes_total":     "sum(node:node_memory_bytes_total:sum)",
	"cluster_memory_bytes_usage":     "sum(node:node_memory_bytes_total:sum) - sum(node:node_memory_bytes_available:sum)",
	"cluster_net_utilisation":        "sum(node:node_net_utilisation:sum_irate)",
	"cluster_net_bytes_transmitted":  "sum(node:node_net_bytes_transmitted:sum_irate)",
	"cluster_net_bytes_received":     "sum(node:node_net_bytes_received:sum_irate)",
	"cluster_disk_read_iops":         "sum(node:data_volume_iops_reads:sum)",
	"cluster_disk_write_iops":        "sum(node:data_volume_iops_writes:sum)",
	"cluster_disk_read_throughput":   "sum(node:data_volume_throughput_bytes_read:sum)",
	"cluster_disk_write_throughput":  "sum(node:data_volume_throughput_bytes_written:sum)",
	"cluster_disk_size_usage":        `sum(sum by (node) ((node_filesystem_size{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:)) - sum(sum by (node) ((node_filesystem_avail{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:))`,
	"cluster_disk_size_utilisation":  `(sum(sum by (node) ((node_filesystem_size{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:)) - sum(sum by (node) ((node_filesystem_avail{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:))) / sum(sum by (node) ((node_filesystem_size{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:))`,
	"cluster_disk_size_capacity":     `sum(sum by (node) ((node_filesystem_size{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:))`,
	"cluster_disk_size_available":    `sum(sum by (node) ((node_filesystem_avail{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:))`,

	//node
	"node_cpu_utilisation":        "node:node_cpu_utilisation:avg1m",
	"node_cpu_total":              "node:node_num_cpu:sum",
	"node_memory_utilisation":     "node:node_memory_utilisation:",
	"node_memory_bytes_available": "node:node_memory_bytes_available:sum",
	"node_memory_bytes_total":     "node:node_memory_bytes_total:sum",
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

	"node_disk_size_capacity":    `sum by (node) ((node_filesystem_size{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1)`,
	"node_disk_size_available":   `sum by (node) ((node_filesystem_avail{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1)`,
	"node_disk_size_usage":       `sum by (node) ((node_filesystem_size{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1)  -sum by (node) ((node_filesystem_avail{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1)`,
	"node_disk_size_utilisation": `sum by (node) (((node_filesystem_size{mountpoint="/", job="node-exporter"} - node_filesystem_avail{mountpoint="/", job="node-exporter"}) / node_filesystem_size{mountpoint="/", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1)`,
	"node_pod_count":             `count(kube_pod_info$1 unless on(pod) kube_pod_completion_time) by (node)`,
	// without log node: unless on(node) kube_node_labels{label_role="log"}
	"node_pod_quota":          `sum(kube_node_status_capacity_pods$1) by (node)`,
	"node_cpu_usage":          `sum by (node) (label_join(irate(container_cpu_usage_seconds_total{job="kubelet", image!=""}[5m]), "pod", " ", "pod_name") * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1)`,
	"node_memory_bytes_usage": "node:node_memory_bytes_total:sum$1 - node:node_memory_bytes_available:sum$1",

	//namespace
	"namespace_cpu_usage":             `namespace:container_cpu_usage_seconds_total:sum_rate{namespace=~"$1"}`,
	"namespace_memory_usage":          `namespace:container_memory_usage_bytes:sum{namespace=~"$1"}`,
	"namespace_memory_usage_wo_cache": `namespace:container_memory_usage_bytes_wo_cache:sum{namespace=~"$1"}`,
	"namespace_net_bytes_transmitted": `sum by (namespace) (irate(container_network_transmit_bytes_total{namespace=~"$1", pod_name!="", interface="eth0", job="kubelet"}[5m]))`,
	"namespace_net_bytes_received":    `sum by (namespace) (irate(container_network_receive_bytes_total{namespace=~"$1", pod_name!="", interface="eth0", job="kubelet"}[5m]))`,
	"namespace_pod_count":             `count(kube_pod_info{namespace=~"$1"} unless on(pod) kube_pod_completion_time) by (namespace)`,

	// pod
	"pod_cpu_usage":             `sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name="$2", image!=""}[5m])) by (namespace, pod_name)`,
	"pod_memory_usage":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name="$2", image!=""}) by (namespace, pod_name)`,
	"pod_memory_usage_wo_cache": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name="$2", image!=""} - container_memory_cache{job="kubelet", namespace="$1", pod_name="$2",image!=""}) by (namespace, pod_name)`,
	"pod_net_bytes_transmitted": `sum by (namespace, pod_name) (irate(container_network_transmit_bytes_total{namespace="$1", pod_name!="", pod_name="$2", interface="eth0", job="kubelet"}[5m]))`,
	"pod_net_bytes_received":    `sum by (namespace, pod_name) (irate(container_network_receive_bytes_total{namespace="$1", pod_name!="", pod_name="$2", interface="eth0", job="kubelet"}[5m]))`,

	"pod_cpu_usage_all":             `sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name=~"$2", image!=""}[5m])) by (namespace, pod_name)`,
	"pod_memory_usage_all":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name=~"$2",  image!=""}) by (namespace, pod_name)`,
	"pod_memory_usage_wo_cache_all": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name=~"$2", image!=""} - container_memory_cache{job="kubelet", namespace="$1", pod_name=~"$2", image!=""}) by (namespace, pod_name)`,
	"pod_net_bytes_transmitted_all": `sum by (namespace, pod_name) (irate(container_network_transmit_bytes_total{namespace="$1", pod_name!="", pod_name=~"$2", interface="eth0", job="kubelet"}[5m]))`,
	"pod_net_bytes_received_all":    `sum by (namespace, pod_name) (irate(container_network_receive_bytes_total{namespace="$1", pod_name!="", pod_name=~"$2", interface="eth0", job="kubelet"}[5m]))`,

	"pod_cpu_usage_node":             `sum by (node, pod) (label_join(irate(container_cpu_usage_seconds_total{job="kubelet",pod_name=~"$2", image!=""}[5m]), "pod", " ", "pod_name") * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,
	"pod_memory_usage_node":          `sum by (node, pod) (label_join(container_memory_usage_bytes{job="kubelet",pod_name=~"$2", image!=""}, "pod", " ", "pod_name") * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,
	"pod_memory_usage_wo_cache_node": `sum by (node, pod) ((label_join(container_memory_usage_bytes{job="kubelet",pod_name=~"$2",  image!=""}, "pod", " ", "pod_name") - label_join(container_memory_cache{job="kubelet",pod_name=~"$2", image!=""}, "pod", " ", "pod_name")) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{node=~"$3"})`,

	// container
	"container_cpu_usage":     `sum(irate(container_cpu_usage_seconds_total{namespace="$1", pod_name="$2", container_name="$3"}[5m])) by (namespace, pod_name, container_name)`,
	"container_cpu_usage_all": `sum(irate(container_cpu_usage_seconds_total{namespace="$1", pod_name="$2", container_name=~"$3", container_name!="POD"}[5m])) by (namespace, pod_name, container_name)`,

	"container_memory_usage_wo_cache":     `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name="$3"} - ignoring(id, image, endpoint, instance, job, name, service) container_memory_cache{namespace="$1", pod_name="$2", container_name="$3"}`,
	"container_memory_usage_wo_cache_all": `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name=~"$3", container_name!="POD"} - ignoring(id, image, endpoint, instance, job, name, service) container_memory_cache{namespace="$1", pod_name="$2", container_name=~"$3", container_name!="POD"}`,
	"container_memory_usage":              `container_memory_usage_bytes{namespace="$1", pod_name="$2",  container_name="$3"}`,
	"container_memory_usage_all":          `container_memory_usage_bytes{namespace="$1", pod_name="$2",  container_name=~"$3", container_name!="POD"}`,

	// enterprise
	"workspace_cpu_usage":             `sum(namespace:container_cpu_usage_seconds_total:sum_rate{namespace =~"$1"})`,
	"workspace_memory_usage":          `sum(namespace:container_memory_usage_bytes:sum{namespace =~"$1"})`,
	"workspace_memory_usage_wo_cache": `sum(namespace:container_memory_usage_bytes_wo_cache:sum{namespace =~"$1"})`,
	"workspace_net_bytes_transmitted": `sum(sum by (namespace) (irate(container_network_transmit_bytes_total{namespace=~"$1", pod_name!="", interface="eth0", job="kubelet"}[5m])))`,
	"workspace_net_bytes_received":    `sum(sum by (namespace) (irate(container_network_receive_bytes_total{namespace=~"$1", pod_name!="", interface="eth0", job="kubelet"}[5m])))`,
	"workspace_pod_count":             `sum(count(kube_pod_info{namespace=~"$1"} unless on(pod) kube_pod_completion_time) by (namespace))`,
}
