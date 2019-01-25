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
	MetricStatus             = "status"
	MetricStatusError        = "error"
	MetricStatusSuccess      = "success"
	ResultItemMetric         = "metric"
	ResultItemMetricResource = "resource"
	ResultItemValue          = "value"
	ResultItemValues         = "values"
	ResultSortTypeDesc       = "desc"
	ResultSortTypeAsce       = "asce"
)

const (
	MetricNameWorkloadCount     = "workload_count"
	MetricNameNamespacePodCount = "namespace_pod_count"

	MetricNameWorkspaceAllOrganizationCount = "workspace_all_organization_count"
	MetricNameWorkspaceAllAccountCount      = "workspace_all_account_count"
	MetricNameWorkspaceAllProjectCount      = "workspace_all_project_count"
	MetricNameWorkspaceAllDevopsCount       = "workspace_all_devops_project_count"
	MetricNameClusterAllProjectCount        = "cluster_namespace_count"

	MetricNameWorkspaceNamespaceCount = "workspace_namespace_count"
	MetricNameWorkspaceDevopsCount    = "workspace_devops_project_count"
	MetricNameWorkspaceMemberCount    = "workspace_member_count"
	MetricNameWorkspaceRoleCount      = "workspace_role_count"
	MetricNameComponentOnLine         = "component_online_count"
	MetricNameComponentLine           = "component_count"
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
	MetricLevelCluster          = "cluster"
	MetricLevelClusterWorkspace = "cluster_workspace"
	MetricLevelNode             = "node"
	MetricLevelWorkspace        = "workspace"
	MetricLevelNamespace        = "namespace"
	MetricLevelPod              = "pod"
	MetricLevelPodName          = "pod_name"
	MetricLevelContainer        = "container"
	MetricLevelContainerName    = "container_name"
	MetricLevelWorkload         = "workload"
)

const (
	ReplicaSet  = "ReplicaSet"
	StatefulSet = "StatefulSet"
	DaemonSet   = "DaemonSet"
	Deployment  = "Deployment"
)

const (
	NodeStatusRule                   = `kube_node_status_condition{condition="Ready"} > 0`
	PodInfoRule                      = `kube_pod_info{created_by_kind="$1",created_by_name=$2,namespace="$3"}`
	NamespaceLabelRule               = `kube_namespace_labels`
	WorkloadReplicaSetOwnerRule      = `kube_pod_owner{namespace="$1", owner_name!="<none>", owner_kind="ReplicaSet"}`
	WorkspaceNamespaceLabelRule      = `sum(kube_namespace_labels{label_kubesphere_io_workspace != ""}) by (label_kubesphere_io_workspace)`
	ExcludedVirtualNetworkInterfaces = `interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)"`
)

const (
	WorkspaceJoinedKey = "label_kubesphere_io_workspace"
)

type MetricMap map[string]string

var ClusterMetricsNames = []string{
	"cluster_cpu_utilisation",
	"cluster_cpu_usage",
	"cluster_cpu_total",
	"cluster_memory_utilisation",
	"cluster_memory_available",
	"cluster_memory_total",
	"cluster_memory_usage_wo_cache",
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
	"cluster_disk_inode_total",
	"cluster_disk_inode_usage",
	"cluster_disk_inode_utilisation",

	"cluster_node_online",
	"cluster_node_offline",
	"cluster_node_total",

	"cluster_pod_count",
	"cluster_pod_quota",
	"cluster_pod_utilisation",
	"cluster_pod_running_count",
	"cluster_pod_succeeded_count",
	"cluster_pod_abnormal_count",
	"cluster_ingresses_extensions_count",
	"cluster_cronjob_count",
	"cluster_pvc_count",
	"cluster_daemonset_count",
	"cluster_deployment_count",
	"cluster_endpoint_count",
	"cluster_hpa_count",
	"cluster_job_count",
	"cluster_statefulset_count",
	"cluster_replicaset_count",
	"cluster_service_count",
	"cluster_secret_count",

	"cluster_namespace_count",

	"cluster_load1",
	"cluster_load5",
	"cluster_load15",
}
var NodeMetricsNames = []string{
	"node_cpu_utilisation",
	"node_cpu_total",
	"node_cpu_usage",
	"node_memory_utilisation",
	"node_memory_usage_wo_cache",
	"node_memory_available",
	"node_memory_total",

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

	"node_disk_inode_total",
	"node_disk_inode_usage",
	"node_disk_inode_utilisation",

	"node_pod_count",
	"node_pod_quota",
	"node_pod_utilisation",
	"node_pod_running_count",
	"node_pod_succeeded_count",
	"node_pod_abnormal_count",

	"node_load1",
	"node_load5",
	"node_load15",
}
var WorkspaceMetricsNames = []string{
	"workspace_cpu_usage",
	"workspace_memory_usage",
	"workspace_memory_usage_wo_cache",
	"workspace_net_bytes_transmitted",
	"workspace_net_bytes_received",
	"workspace_pod_count",
	"workspace_pod_running_count",
	"workspace_pod_succeeded_count",
	"workspace_pod_abnormal_count",
	"workspace_ingresses_extensions_count",

	"workspace_cronjob_count",
	"workspace_pvc_count",
	"workspace_daemonset_count",
	"workspace_deployment_count",
	"workspace_endpoint_count",
	"workspace_hpa_count",
	"workspace_job_count",
	"workspace_statefulset_count",
	"workspace_replicaset_count",
	"workspace_service_count",
	"workspace_secret_count",

	"workspace_all_project_count",
}
var NamespaceMetricsNames = []string{
	"namespace_cpu_usage",
	"namespace_memory_usage",
	"namespace_memory_usage_wo_cache",
	"namespace_net_bytes_transmitted",
	"namespace_net_bytes_received",
	"namespace_pod_count",
	"namespace_pod_running_count",
	"namespace_pod_succeeded_count",
	"namespace_pod_abnormal_count",

	"namespace_configmap_count_used",
	"namespace_jobs_batch_count_used",
	"namespace_roles_count_used",
	"namespace_memory_limit_used",
	"namespace_pvc_used",
	"namespace_memory_request_used",
	"namespace_pvc_count_used",
	"namespace_cronjobs_batch_count_used",
	"namespace_ingresses_extensions_count_used",
	"namespace_cpu_limit_used",
	"namespace_storage_request_used",
	"namespace_deployment_count_used",
	"namespace_pod_count_used",
	"namespace_statefulset_count_used",
	"namespace_daemonset_count_used",
	"namespace_secret_count_used",
	"namespace_service_count_used",
	"namespace_cpu_request_used",
	"namespace_service_loadbalancer_used",

	"namespace_configmap_count_hard",
	"namespace_jobs_batch_count_hard",
	"namespace_roles_count_hard",
	"namespace_memory_limit_hard",
	"namespace_pvc_hard",
	"namespace_memory_request_hard",
	"namespace_pvc_count_hard",
	"namespace_cronjobs_batch_count_hard",
	"namespace_ingresses_extensions_count_hard",
	"namespace_cpu_limit_hard",
	"namespace_storage_request_hard",
	"namespace_deployment_count_hard",
	"namespace_pod_count_hard",
	"namespace_statefulset_count_hard",
	"namespace_daemonset_count_hard",
	"namespace_secret_count_hard",
	"namespace_service_count_hard",
	"namespace_cpu_request_hard",
	"namespace_service_loadbalancer_hard",

	"namespace_cronjob_count",
	"namespace_pvc_count",
	"namespace_daemonset_count",
	"namespace_deployment_count",
	"namespace_endpoint_count",
	"namespace_hpa_count",
	"namespace_job_count",
	"namespace_statefulset_count",
	"namespace_replicaset_count",
	"namespace_service_count",
	"namespace_secret_count",

	"namespace_ingresses_extensions_count",
}

var PodMetricsNames = []string{
	"pod_cpu_usage",
	"pod_memory_usage",
	"pod_memory_usage_wo_cache",
	"pod_net_bytes_transmitted",
	"pod_net_bytes_received",
}

var WorkloadMetricsNames = []string{
	"workload_pod_cpu_usage",
	"workload_pod_memory_usage",
	"workload_pod_memory_usage_wo_cache",
	"workload_pod_net_bytes_transmitted",
	"workload_pod_net_bytes_received",

	"workload_deployment_replica",
	"workload_deployment_replica_available",
	"workload_statefulset_replica",
	"workload_statefulset_replica_available",
	"workload_daemonset_replica",
	"workload_daemonset_replica_available",
}

var ContainerMetricsNames = []string{
	"container_cpu_usage",
	"container_memory_usage",
	"container_memory_usage_wo_cache",
	//"container_net_bytes_transmitted",
	//"container_net_bytes_received",
}

var RulePromQLTmplMap = MetricMap{
	//cluster
	"cluster_cpu_utilisation":       ":node_cpu_utilisation:avg1m",
	"cluster_cpu_usage":             `:node_cpu_utilisation:avg1m * sum(node:node_num_cpu:sum)`,
	"cluster_cpu_total":             "sum(node:node_num_cpu:sum)",
	"cluster_memory_utilisation":    ":node_memory_utilisation:",
	"cluster_memory_available":      "sum(node:node_memory_bytes_available:sum)",
	"cluster_memory_total":          "sum(node:node_memory_bytes_total:sum)",
	"cluster_memory_usage_wo_cache": "sum(node:node_memory_bytes_total:sum) - sum(node:node_memory_bytes_available:sum)",

	"cluster_net_utilisation":       ":node_net_utilisation:sum_irate",
	"cluster_net_bytes_transmitted": "sum(node:node_net_bytes_transmitted:sum_irate)",
	"cluster_net_bytes_received":    "sum(node:node_net_bytes_received:sum_irate)",
	"cluster_disk_read_iops":        "sum(node:data_volume_iops_reads:sum)",
	"cluster_disk_write_iops":       "sum(node:data_volume_iops_writes:sum)",
	"cluster_disk_read_throughput":  "sum(node:data_volume_throughput_bytes_read:sum)",
	"cluster_disk_write_throughput": "sum(node:data_volume_throughput_bytes_written:sum)",

	"cluster_disk_size_usage":       `sum(max((node_filesystem_size{device=~"/dev/.+", job="node-exporter"} - node_filesystem_avail{device=~"/dev/.+", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:) by (node))`,
	"cluster_disk_size_utilisation": `1 - sum(max(node_filesystem_avail{device=~"/dev/.+", job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:) by (node)) / sum(max(node_filesystem_size{device=~"/dev/.+", job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:) by (node))`,
	"cluster_disk_size_capacity":    `sum(max(node_filesystem_size{device=~"/dev/.+", job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:) by (node))`,
	"cluster_disk_size_available":   `sum(max(node_filesystem_avail{device=~"/dev/.+", job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:) by (node))`,

	"cluster_disk_inode_total":       `sum(node:disk_inodes_total:)`,
	"cluster_disk_inode_usage":       `sum(node:disk_inodes_total:) - sum(node:disk_inodes_free:)`,
	"cluster_disk_inode_utilisation": `1 - sum(node:disk_inodes_free:) / sum(node:disk_inodes_total:)`,

	"cluster_namespace_count": `count(kube_namespace_annotations)`,

	// cluster_pod_count = cluster_pod_running_count + cluster_pod_succeeded_count + cluster_pod_abnormal_count
	"cluster_pod_count":           `sum((kube_pod_status_scheduled{condition="true"} > 0)  * on (namespace, pod) group_left(node) (sum by (node, namespace, pod) (kube_pod_info)) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"cluster_pod_quota":           `sum(kube_node_status_capacity_pods unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"cluster_pod_utilisation":     `sum(kube_pod_info unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)) / sum(kube_node_status_capacity_pods unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"cluster_pod_running_count":   `count(kube_pod_info unless on (pod) (kube_pod_status_phase{phase=~"Failed|Pending|Unknown|Succeeded"} > 0) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"cluster_pod_succeeded_count": `count(kube_pod_info unless on (pod) (kube_pod_status_phase{phase=~"Failed|Pending|Unknown|Running"} > 0) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"cluster_pod_abnormal_count":  `count(kube_pod_info unless on (pod) (kube_pod_status_phase{phase=~"Succeeded|Running"} > 0) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,

	"cluster_node_online":  `sum(kube_node_status_condition{condition="Ready",status="true"})`,
	"cluster_node_offline": `sum(kube_node_status_condition{condition="Ready",status=~"unknown|false"})`,
	"cluster_node_total":   `sum(kube_node_status_condition{condition="Ready"})`,

	"cluster_ingresses_extensions_count": `sum(kube_ingress_labels)`,

	"cluster_configmap_count_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/configmaps"}) by (resource, type)`,
	"cluster_jobs_batch_count_used":           `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/jobs.batch"}) by (resource, type)`,
	"cluster_roles_count_used":                `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/roles.rbac.authorization.k8s.io"}) by (resource, type)`,
	"cluster_memory_limit_used":               `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="limits.memory"}) by (resource, type)`,
	"cluster_pvc_used":                        `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="persistentvolumeclaims"}) by (resource, type)`,
	"cluster_memory_request_used":             `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="requests.memory"}) by (resource, type)`,
	"cluster_pvc_count_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/persistentvolumeclaims"}) by (resource, type)`,
	"cluster_cronjobs_batch_count_used":       `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/cronjobs.batch"}) by (resource, type)`,
	"cluster_ingresses_extensions_count_used": `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/ingresses.extensions"}) by (resource, type)`,
	"cluster_cpu_limit_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="limits.cpu"}) by (resource, type)`,
	"cluster_storage_request_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="requests.storage"}) by (resource, type)`,
	"cluster_deployment_count_used":           `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/deployments.apps"}) by (resource, type)`,
	"cluster_pod_count_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/pods"}) by (resource, type)`,
	"cluster_statefulset_count_used":          `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/statefulsets.apps"}) by (resource, type)`,
	"cluster_daemonset_count_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/daemonsets.apps"}) by (resource, type)`,
	"cluster_secret_count_used":               `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/secrets"}) by (resource, type)`,
	"cluster_service_count_used":              `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="count/services"}) by (resource, type)`,
	"cluster_cpu_request_used":                `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="requests.cpu"}) by (resource, type)`,
	"cluster_service_loadbalancer_used":       `sum(kube_resourcequota{resourcequota!="quota", type="used", resource="services.loadbalancers"}) by (resource, type)`,

	"cluster_cronjob_count":     `sum(kube_cronjob_labels)`,
	"cluster_pvc_count":         `sum(kube_persistentvolumeclaim_info)`,
	"cluster_daemonset_count":   `sum(kube_daemonset_labels)`,
	"cluster_deployment_count":  `sum(kube_deployment_labels)`,
	"cluster_endpoint_count":    `sum(kube_endpoint_labels)`,
	"cluster_hpa_count":         `sum(kube_hpa_labels)`,
	"cluster_job_count":         `sum(kube_job_labels)`,
	"cluster_statefulset_count": `sum(kube_statefulset_labels)`,
	"cluster_replicaset_count":  `count(kube_replicaset_created)`,
	"cluster_service_count":     `sum(kube_service_info)`,
	"cluster_secret_count":      `sum(kube_secret_info)`,
	"cluster_pv_count":          `sum(kube_persistentvolume_labels)`,

	"cluster_load1":  `sum(node_load1{job="node-exporter"}) / sum(node:node_num_cpu:sum)`,
	"cluster_load5":  `sum(node_load5{job="node-exporter"}) / sum(node:node_num_cpu:sum)`,
	"cluster_load15": `sum(node_load15{job="node-exporter"}) / sum(node:node_num_cpu:sum)`,

	//node
	"node_cpu_utilisation":       "node:node_cpu_utilisation:avg1m",
	"node_cpu_total":             "node:node_num_cpu:sum",
	"node_memory_utilisation":    "node:node_memory_utilisation:",
	"node_memory_available":      "node:node_memory_bytes_available:sum",
	"node_memory_total":          "node:node_memory_bytes_total:sum",
	"node_memory_usage_wo_cache": "node:node_memory_bytes_total:sum$1 - node:node_memory_bytes_available:sum$1",

	"node_net_utilisation":       "node:node_net_utilisation:sum_irate",
	"node_net_bytes_transmitted": "node:node_net_bytes_transmitted:sum_irate",
	"node_net_bytes_received":    "node:node_net_bytes_received:sum_irate",
	"node_disk_read_iops":        "node:data_volume_iops_reads:sum",
	"node_disk_write_iops":       "node:data_volume_iops_writes:sum",
	"node_disk_read_throughput":  "node:data_volume_throughput_bytes_read:sum",
	"node_disk_write_throughput": "node:data_volume_throughput_bytes_written:sum",

	"node_disk_size_capacity":    `max(node_filesystem_size{device=~"/dev/.+", job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1) by (node)`,
	"node_disk_size_available":   `max(node_filesystem_avail{device=~"/dev/.+", job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1) by (node)`,
	"node_disk_size_usage":       `max((node_filesystem_size{device=~"/dev/.+", job="node-exporter"} - node_filesystem_avail{device=~"/dev/.+", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1) by (node)`,
	"node_disk_size_utilisation": `max(((node_filesystem_size{device=~"/dev/.+", job="node-exporter"} - node_filesystem_avail{device=~"/dev/.+", job="node-exporter"}) / node_filesystem_size{device=~"/dev/.+", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1) by (node)`,

	"node_disk_inode_total":       `node:disk_inodes_total:$1`,
	"node_disk_inode_usage":       `node:disk_inodes_total:$1 - node:disk_inodes_free:$1`,
	"node_disk_inode_utilisation": `(1 - (node:disk_inodes_free:$1 / node:disk_inodes_total:$1))`,

	"node_pod_count":           `sum by (node) ((kube_pod_status_scheduled{condition="true"} > 0)  * on (namespace, pod) group_left(node) kube_pod_info$1 unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"node_pod_quota":           `sum(kube_node_status_capacity_pods$1) by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)`,
	"node_pod_utilisation":     `(sum(kube_pod_info$1) by (node) / sum(kube_node_status_capacity_pods$1) by (node)) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)`,
	"node_pod_running_count":   `count(kube_pod_info$1 unless on (pod) (kube_pod_status_phase{phase=~"Failed|Pending|Unknown|Succeeded"} > 0))  by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)`,
	"node_pod_succeeded_count": `count(kube_pod_info$1 unless on (pod) (kube_pod_status_phase{phase=~"Failed|Pending|Unknown|Running"} > 0))  by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)`,
	"node_pod_abnormal_count":  `count(kube_pod_info$1 unless on (pod) (kube_pod_status_phase{phase=~"Succeeded|Running"} > 0)) by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)`,

	// without log node: unless on(node) kube_node_labels{label_role="log"}
	"node_cpu_usage": `node:node_cpu_utilisation:avg1m$1 * node:node_num_cpu:sum$1`,

	"node_load1":  `sum by (node) (node_load1{job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1) / node:node_num_cpu:sum`,
	"node_load5":  `sum by (node) (node_load5{job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1) / node:node_num_cpu:sum`,
	"node_load15": `sum by (node) (node_load15{job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1) / node:node_num_cpu:sum`,

	//namespace
	"namespace_cpu_usage":             `namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", namespace=~"$1"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_memory_usage":          `namespace:container_memory_usage_bytes:sum{namespace!="", namespace=~"$1"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_memory_usage_wo_cache": `namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", namespace=~"$1"}* on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_net_bytes_transmitted": `sum by (namespace) (irate(container_network_transmit_bytes_total{namespace!="", namespace=~"$1", pod_name!="", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))* on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_net_bytes_received":    `sum by (namespace) (irate(container_network_receive_bytes_total{namespace!="", namespace=~"$1", pod_name!="", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m])) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_count":             `sum(kube_pod_status_phase{phase!~"Failed|Succeeded", namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_running_count":     `sum(kube_pod_status_phase{phase="Running", namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_succeeded_count":   `sum(kube_pod_status_phase{phase="Succeeded", namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_abnormal_count":    `sum(kube_pod_status_phase{phase=~"Failed|Pending|Unknown", namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,

	"namespace_configmap_count_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/configmaps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_jobs_batch_count_used":           `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/jobs.batch"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_roles_count_used":                `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/roles.rbac.authorization.k8s.io"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_memory_limit_used":               `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="limits.memory"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pvc_used":                        `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="persistentvolumeclaims"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_memory_request_used":             `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="requests.memory"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pvc_count_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/persistentvolumeclaims"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_cronjobs_batch_count_used":       `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/cronjobs.batch"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_ingresses_extensions_count_used": `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/ingresses.extensions"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_cpu_limit_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="limits.cpu"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_storage_request_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="requests.storage"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_deployment_count_used":           `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/deployments.apps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_count_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/pods"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_statefulset_count_used":          `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/statefulsets.apps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_daemonset_count_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/daemonsets.apps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_secret_count_used":               `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/secrets"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_service_count_used":              `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/services"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_cpu_request_used":                `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="requests.cpu"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_service_loadbalancer_used":       `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="services.loadbalancers"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,

	"namespace_configmap_count_hard":            `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/configmaps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_jobs_batch_count_hard":           `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/jobs.batch"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_roles_count_hard":                `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/roles.rbac.authorization.k8s.io"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_memory_limit_hard":               `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="limits.memory"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pvc_hard":                        `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="persistentvolumeclaims"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_memory_request_hard":             `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="requests.memory"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pvc_count_hard":                  `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/persistentvolumeclaims"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_cronjobs_batch_count_hard":       `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/cronjobs.batch"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_ingresses_extensions_count_hard": `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/ingresses.extensions"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_cpu_limit_hard":                  `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="limits.cpu"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_storage_request_hard":            `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="requests.storage"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_deployment_count_hard":           `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/deployments.apps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_count_hard":                  `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/pods"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_statefulset_count_hard":          `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/statefulsets.apps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_daemonset_count_hard":            `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/daemonsets.apps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_secret_count_hard":               `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/secrets"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_service_count_hard":              `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/services"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_cpu_request_hard":                `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="requests.cpu"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_service_loadbalancer_hard":       `sum(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="services.loadbalancers"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,

	"namespace_cronjob_count":     `sum(kube_cronjob_labels{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pvc_count":         `sum(kube_persistentvolumeclaim_info{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_daemonset_count":   `sum(kube_daemonset_labels{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_deployment_count":  `sum(kube_deployment_labels{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_endpoint_count":    `sum(kube_endpoint_labels{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_hpa_count":         `sum(kube_hpa_labels{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_job_count":         `sum(kube_job_labels{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_statefulset_count": `sum(kube_statefulset_labels{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_replicaset_count":  `count(kube_replicaset_created{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_service_count":     `sum(kube_service_info{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_secret_count":      `sum(kube_secret_info{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,

	"namespace_ingresses_extensions_count": `sum(kube_ingress_labels{namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,

	// pod
	"pod_cpu_usage":             `sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name!="", pod_name="$2", image!=""}[5m])) by (namespace, pod_name)`,
	"pod_memory_usage":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name!="", pod_name="$2", image!=""}) by (namespace, pod_name)`,
	"pod_memory_usage_wo_cache": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name!="", pod_name="$2", image!=""} - container_memory_cache{job="kubelet", namespace="$1", pod_name!="", pod_name="$2",image!=""}) by (namespace, pod_name)`,
	"pod_net_bytes_transmitted": `sum by (namespace, pod_name) (irate(container_network_transmit_bytes_total{namespace="$1", pod_name!="", pod_name="$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))`,
	"pod_net_bytes_received":    `sum by (namespace, pod_name) (irate(container_network_receive_bytes_total{namespace="$1", pod_name!="", pod_name="$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))`,

	"pod_cpu_usage_all":             `sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name!="", pod_name=~"$2", image!=""}[5m])) by (namespace, pod_name)`,
	"pod_memory_usage_all":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name!="", pod_name=~"$2",  image!=""}) by (namespace, pod_name)`,
	"pod_memory_usage_wo_cache_all": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name!="", pod_name=~"$2", image!=""} - container_memory_cache{job="kubelet", namespace="$1", pod_name!="", pod_name=~"$2", image!=""}) by (namespace, pod_name)`,
	"pod_net_bytes_transmitted_all": `sum by (namespace, pod_name) (irate(container_network_transmit_bytes_total{namespace="$1", pod_name!="", pod_name=~"$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))`,
	"pod_net_bytes_received_all":    `sum by (namespace, pod_name) (irate(container_network_receive_bytes_total{namespace="$1", pod_name!="", pod_name=~"$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))`,

	"pod_cpu_usage_node":             `sum by (node, pod_name) (irate(container_cpu_usage_seconds_total{job="kubelet",pod_name!="", pod_name=~"$2", image!=""}[5m]) * on (namespace, pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name"))`,
	"pod_memory_usage_node":          `sum by (node, pod_name) (container_memory_usage_bytes{job="kubelet",pod_name!="", pod_name=~"$2", image!=""} * on (namespace, pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name"))`,
	"pod_memory_usage_wo_cache_node": `sum by (node, pod_name) ((container_memory_usage_bytes{job="kubelet",pod_name!="", pod_name=~"$2", image!=""} - container_memory_cache{job="kubelet",pod_name!="", pod_name=~"$2", image!=""}) * on (namespace, pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name"))`,
	"pod_net_bytes_transmitted_node": `sum by (node, pod_name) (irate(container_network_transmit_bytes_total{pod_name!="", pod_name=~"$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name"))`,
	"pod_net_bytes_received_node":    `sum by (node, pod_name) (irate(container_network_receive_bytes_total{pod_name!="", pod_name=~"$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name"))`,

	// workload
	"workload_pod_cpu_usage":             `label_join(sum(label_replace(label_replace(label_replace(label_join(label_join(label_replace(sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$2", pod_name!="", pod_name=~"$3", image!=""}[5m])) by (namespace, pod_name) * on (pod_name) group_left(owner_kind) label_join(label_replace(kube_pod_owner{namespace="$2", pod=~".*"}, "owner_kind", "POD", "owner_kind", "<none>"), "pod_name", "", "pod", "_name") , "postfix", "-POD", "owner_kind", "POD"), "pod_name", "", "pod_name", "postfix"), "dist", "-", "owner_kind", "pod_name"), "pod_name", "$1", "dist", "ReplicaSet-(.+)-(.+)"), "workload", "$1", "pod_name", "(.+)-(.+)"), "owner_kind", "Deployment", "owner_kind", "ReplicaSet.*")) by (namespace, workload, owner_kind), "workload", ":", "owner_kind", "workload")`,
	"workload_pod_memory_usage":          `label_join(sum(label_replace(label_replace(label_replace(label_join(label_join(label_replace(sum(container_memory_usage_bytes{job="kubelet", namespace="$2", pod_name!="", pod_name=~"$3", image!=""}) by (namespace, pod_name) * on (pod_name) group_left(owner_kind) label_join(label_replace(kube_pod_owner{namespace="$2", pod=~".*"}, "owner_kind", "POD", "owner_kind", "<none>"), "pod_name", "", "pod", "_name") , "postfix", "-POD", "owner_kind", "POD"), "pod_name", "", "pod_name", "postfix"), "dist", "-", "owner_kind", "pod_name"), "pod_name", "$1", "dist", "ReplicaSet-(.+)-(.+)"), "workload", "$1", "pod_name", "(.+)-(.+)"), "owner_kind", "Deployment", "owner_kind", "ReplicaSet.*")) by (namespace, workload, owner_kind), "workload", ":", "owner_kind", "workload")`,
	"workload_pod_memory_usage_wo_cache": `label_join(sum(label_replace(label_replace(label_replace(label_join(label_join(label_replace(sum(container_memory_usage_bytes{job="kubelet", namespace="$2", pod_name!="", pod_name=~"$3", image!=""} - container_memory_cache{job="kubelet", namespace="$2", pod_name!="", pod_name=~"$3", image!=""}) by (namespace, pod_name) * on (pod_name) group_left(owner_kind) label_join(label_replace(kube_pod_owner{namespace="$2", pod=~".*"}, "owner_kind", "POD", "owner_kind", "<none>"), "pod_name", "", "pod", "_name") , "postfix", "-POD", "owner_kind", "POD"), "pod_name", "", "pod_name", "postfix"), "dist", "-", "owner_kind", "pod_name"), "pod_name", "$1", "dist", "ReplicaSet-(.+)-(.+)"), "workload", "$1", "pod_name", "(.+)-(.+)"), "owner_kind", "Deployment", "owner_kind", "ReplicaSet.*")) by (namespace, workload, owner_kind), "workload", ":", "owner_kind", "workload")`,
	"workload_pod_net_bytes_transmitted": `label_join(sum(label_replace(label_replace(label_replace(label_join(label_join(label_replace(sum(irate(container_network_transmit_bytes_total{namespace="$2", pod_name!="", pod_name=~"$3", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m])) by (namespace, pod_name) * on (pod_name) group_left(owner_kind) label_join(label_replace(kube_pod_owner{namespace="$2", pod=~".*"}, "owner_kind", "POD", "owner_kind", "<none>"), "pod_name", "", "pod", "_name") , "postfix", "-POD", "owner_kind", "POD"), "pod_name", "", "pod_name", "postfix"), "dist", "-", "owner_kind", "pod_name"), "pod_name", "$1", "dist", "ReplicaSet-(.+)-(.+)"), "workload", "$1", "pod_name", "(.+)-(.+)"), "owner_kind", "Deployment", "owner_kind", "ReplicaSet.*")) by (namespace, workload, owner_kind), "workload", ":", "owner_kind", "workload")`,
	"workload_pod_net_bytes_received":    `label_join(sum(label_replace(label_replace(label_replace(label_join(label_join(label_replace(sum(irate(container_network_receive_bytes_total{namespace="$2", pod_name!="", pod_name=~"$3", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m])) by (namespace, pod_name) * on (pod_name) group_left(owner_kind) label_join(label_replace(kube_pod_owner{namespace="$2", pod=~".*"}, "owner_kind", "POD", "owner_kind", "<none>"), "pod_name", "", "pod", "_name") , "postfix", "-POD", "owner_kind", "POD"), "pod_name", "", "pod_name", "postfix"), "dist", "-", "owner_kind", "pod_name"), "pod_name", "$1", "dist", "ReplicaSet-(.+)-(.+)"), "workload", "$1", "pod_name", "(.+)-(.+)"), "owner_kind", "Deployment", "owner_kind", "ReplicaSet.*")) by (namespace, workload, owner_kind), "workload", ":", "owner_kind", "workload")`,

	"workload_deployment_replica":            `label_join(sum (label_join(label_replace(kube_deployment_spec_replicas{namespace="$2", deployment=~"$3"}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_deployment_replica_available":  `label_join(sum (label_join(label_replace(kube_deployment_status_replicas_available{namespace="$2", deployment=~"$3"}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_statefulset_replica":           `label_join(sum (label_join(label_replace(kube_statefulset_replicas{namespace="$2", statefulset=~"$3"}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_statefulset_replica_available": `label_join(sum (label_join(label_replace(kube_statefulset_status_replicas_current{namespace="$2", statefulset=~"$3"}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_daemonset_replica":             `label_join(sum (label_join(label_replace(kube_daemonset_status_desired_number_scheduled{namespace="$2", daemonset=~"$3"}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_daemonset_replica_available":   `label_join(sum (label_join(label_replace(kube_daemonset_status_number_available{namespace="$2", daemonset=~"$3"}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,

	// container
	"container_cpu_usage":             `sum(irate(container_cpu_usage_seconds_total{namespace="$1", pod_name="$2", container_name!="POD", container_name=~"$3"}[5m])) by (namespace, pod_name, container_name)`,
	"container_memory_usage":          `sum(container_memory_usage_bytes{namespace="$1", pod_name="$2",  container_name!="POD", container_name=~"$3"}) by (namespace, pod_name, container_name)`,
	"container_memory_usage_wo_cache": `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name!="POD", container_name=~"$3"} - ignoring(id, image, endpoint, instance, job, name, service) container_memory_cache{namespace="$1", pod_name="$2", container_name!="POD", container_name=~"$3"}`,
	"container_net_bytes_transmitted": `sum(irate(container_network_transmit_bytes_total{job="kubelet", namespace="$1", pod_name="$2", container_name="POD", ` + ExcludedVirtualNetworkInterfaces + `}[5m])) by (namespace, pod_name, container_name)`,
	"container_net_bytes_received":    `sum(irate(container_network_receive_bytes_total{job="kubelet", namespace="$1", pod_name="$2", container_name="POD", ` + ExcludedVirtualNetworkInterfaces + `}[5m])) by (namespace, pod_name, container_name)`,

	"container_cpu_usage_node":             `sum by (node, pod_name, container_name) (irate(container_cpu_usage_seconds_total{job="kubelet", pod_name="$2", container_name!="POD", container_name!="", container_name=~"$3", image!=""}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name"))`,
	"container_memory_usage_node":          `sum by (node, pod_name, container_name) (container_memory_usage_bytes{job="kubelet", pod_name="$2", container_name!="POD", container_name!="", container_name=~"$3", image!=""} * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name"))`,
	"container_memory_usage_wo_cache_node": `sum by (node, pod_name, container_name) ((container_memory_usage_bytes{job="kubelet", pod_name="$2", container_name!="POD", container_name!="", container_name=~"$3", image!=""} - container_memory_cache{job="kubelet", pod_name="$2", container_name!="POD", container_name!="", container_name=~"$3", image!=""}) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name"))`,
	"container_net_bytes_transmitted_node": `sum by (node, pod_name, container_name) (irate(container_network_transmit_bytes_total{job="kubelet", ` + ExcludedVirtualNetworkInterfaces + `, pod_name="$2", container_name="POD", container_name!="", image!=""}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name"))`,
	"container_net_bytes_received_node":    `sum by (node, pod_name, container_name) (irate(container_network_receive_bytes_total{job="kubelet", ` + ExcludedVirtualNetworkInterfaces + `, pod_name="$2", container_name="POD", container_name!="", image!=""}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name"))`,

	// workspace
	"workspace_cpu_usage":             `sum(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", namespace$1})`,
	"workspace_memory_usage":          `sum(namespace:container_memory_usage_bytes:sum{namespace!="", namespace$1})`,
	"workspace_memory_usage_wo_cache": `sum(namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", namespace$1})`,
	"workspace_net_bytes_transmitted": `sum(sum by (namespace) (irate(container_network_transmit_bytes_total{namespace!="", namespace$1, pod_name!="", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m])))`,
	"workspace_net_bytes_received":    `sum(sum by (namespace) (irate(container_network_receive_bytes_total{namespace!="", namespace$1, pod_name!="", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m])))`,
	"workspace_pod_count":             `sum(kube_pod_status_phase{phase!~"Failed|Succeeded", namespace!="", namespace$1})`,
	"workspace_pod_running_count":     `sum(kube_pod_status_phase{phase="Running", namespace!="", namespace$1})`,
	"workspace_pod_succeeded_count":   `sum(kube_pod_status_phase{phase="Succeeded", namespace!="", namespace$1})`,
	"workspace_pod_abnormal_count":    `sum(kube_pod_status_phase{phase=~"Failed|Pending|Unknown", namespace!="", namespace$1})`,

	"workspace_configmap_count_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/configmaps"}) by (resource, type)`,
	"workspace_jobs_batch_count_used":           `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/jobs.batch"}) by (resource, type)`,
	"workspace_roles_count_used":                `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/roles.rbac.authorization.k8s.io"}) by (resource, type)`,
	"workspace_memory_limit_used":               `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="limits.memory"}) by (resource, type)`,
	"workspace_pvc_used":                        `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="persistentvolumeclaims"}) by (resource, type)`,
	"workspace_memory_request_used":             `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="requests.memory"}) by (resource, type)`,
	"workspace_pvc_count_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/persistentvolumeclaims"}) by (resource, type)`,
	"workspace_cronjobs_batch_count_used":       `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/cronjobs.batch"}) by (resource, type)`,
	"workspace_ingresses_extensions_count_used": `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/ingresses.extensions"}) by (resource, type)`,
	"workspace_cpu_limit_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="limits.cpu"}) by (resource, type)`,
	"workspace_storage_request_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="requests.storage"}) by (resource, type)`,
	"workspace_deployment_count_used":           `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/deployments.apps"}) by (resource, type)`,
	"workspace_pod_count_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/pods"}) by (resource, type)`,
	"workspace_statefulset_count_used":          `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/statefulsets.apps"}) by (resource, type)`,
	"workspace_daemonset_count_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/daemonsets.apps"}) by (resource, type)`,
	"workspace_secret_count_used":               `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/secrets"}) by (resource, type)`,
	"workspace_service_count_used":              `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/services"}) by (resource, type)`,
	"workspace_cpu_request_used":                `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="requests.cpu"}) by (resource, type)`,
	"workspace_service_loadbalancer_used":       `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="services.loadbalancers"}) by (resource, type)`,

	"workspace_ingresses_extensions_count": `sum(kube_ingress_labels{namespace!="", namespace$1})`,

	"workspace_cronjob_count":     `sum(kube_cronjob_labels{namespace!="", namespace$1})`,
	"workspace_pvc_count":         `sum(kube_persistentvolumeclaim_info{namespace!="", namespace$1})`,
	"workspace_daemonset_count":   `sum(kube_daemonset_labels{namespace!="", namespace$1})`,
	"workspace_deployment_count":  `sum(kube_deployment_labels{namespace!="", namespace$1})`,
	"workspace_endpoint_count":    `sum(kube_endpoint_labels{namespace!="", namespace$1})`,
	"workspace_hpa_count":         `sum(kube_hpa_labels{namespace!="", namespace$1})`,
	"workspace_job_count":         `sum(kube_job_labels{namespace!="", namespace$1})`,
	"workspace_statefulset_count": `sum(kube_statefulset_labels{namespace!="", namespace$1})`,
	"workspace_replicaset_count":  `count(kube_replicaset_created{namespace!="", namespace$1})`,
	"workspace_service_count":     `sum(kube_service_info{namespace!="", namespace$1})`,
	"workspace_secret_count":      `sum(kube_secret_info{namespace!="", namespace$1})`,

	"workspace_all_project_count": `count(kube_namespace_annotations)`,
}
