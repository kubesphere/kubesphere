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
	ResultTypeVector             = "vector"
	ResultTypeMatrix             = "matrix"
	MetricStatus                 = "status"
	MetricStatusError            = "error"
	MetricStatusSuccess          = "success"
	ResultItemMetric             = "metric"
	ResultItemMetricResource     = "resource"
	ResultItemMetricResourceName = "resource_name"
	ResultItemMetricNodeIp       = "node_ip"
	ResultItemMetricNodeName     = "node_name"
	ResultItemValue              = "value"
	ResultItemValues             = "values"
	ResultSortTypeDesc           = "desc"
	ResultSortTypeAsc            = "asc"
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
	MetricLevelPVC              = "persistentvolumeclaim"
	MetricLevelWorkload         = "workload"
	MetricLevelComponent        = "component"
)

const (
	ReplicaSet  = "ReplicaSet"
	StatefulSet = "StatefulSet"
	DaemonSet   = "DaemonSet"
	Deployment  = "Deployment"
	Any         = ".*"
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

// The metrics need to include extra info out of prometheus
// eg. add node name info to the etcd_server_list metric
const (
	EtcdServerList = "etcd_server_list"
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
	"cluster_ingresses_extensions_count",
	"cluster_namespace_count",

	"cluster_load1",
	"cluster_load5",
	"cluster_load15",

	// New in ks 2.0
	"cluster_pod_abnormal_ratio",
	"cluster_node_offline_ratio",
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

	// New in ks 2.0
	"node_pod_abnormal_ratio",
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

	// New in ks 2.0
	"workspace_pod_abnormal_ratio",
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

	// New in ks 2.0
	"namespace_pod_abnormal_ratio",
	"namespace_resourcequota_used_ratio",
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

	// New in ks 2.0
	"workload_deployment_unavailable_replicas_ratio",
	"workload_daemonset_unavailable_replicas_ratio",
	"workload_statefulset_unavailable_replicas_ratio",
}

var ContainerMetricsNames = []string{
	"container_cpu_usage",
	"container_memory_usage",
	"container_memory_usage_wo_cache",
	//"container_net_bytes_transmitted",
	//"container_net_bytes_received",
}

var PVCMetricsNames = []string{
	"pvc_inodes_available",
	"pvc_inodes_used",
	"pvc_inodes_total",
	"pvc_inodes_utilisation",
	"pvc_bytes_available",
	"pvc_bytes_used",
	"pvc_bytes_total",
	"pvc_bytes_utilisation",
}

var ComponentMetricsNames = []string{
	"etcd_server_list",
	"etcd_server_total",
	"etcd_server_up_total",
	"etcd_server_has_leader",
	"etcd_server_leader_changes",
	"etcd_server_proposals_failed_rate",
	"etcd_server_proposals_applied_rate",
	"etcd_server_proposals_committed_rate",
	"etcd_server_proposals_pending_count",
	"etcd_mvcc_db_size",
	"etcd_network_client_grpc_received_bytes",
	"etcd_network_client_grpc_sent_bytes",
	"etcd_grpc_call_rate",
	"etcd_grpc_call_failed_rate",
	"etcd_grpc_server_msg_received_rate",
	"etcd_grpc_server_msg_sent_rate",
	"etcd_disk_wal_fsync_duration",
	"etcd_disk_wal_fsync_duration_quantile",
	"etcd_disk_backend_commit_duration",
	"etcd_disk_backend_commit_duration_quantile",

	"apiserver_up_sum",
	"apiserver_request_rate",
	"apiserver_request_by_verb_rate",
	"apiserver_request_latencies",
	"apiserver_request_by_verb_latencies",

	"scheduler_up_sum",
	"scheduler_schedule_attempts",
	"scheduler_schedule_attempt_rate",
	"scheduler_e2e_scheduling_latency",
	"scheduler_e2e_scheduling_latency_quantile",

	"controller_manager_up_sum",

	"coredns_up_sum",
	"coredns_cache_hits",
	"coredns_cache_misses",
	"coredns_dns_request_rate",
	"coredns_dns_request_duration",
	"coredns_dns_request_duration_quantile",
	"coredns_dns_request_by_type_rate",
	"coredns_dns_request_by_rcode_rate",
	"coredns_panic_rate",
	"coredns_proxy_request_rate",
	"coredns_proxy_request_duration",
	"coredns_proxy_request_duration_quantile",

	"prometheus_up_sum",
	"prometheus_tsdb_head_samples_appended_rate",
}

var RulePromQLTmplMap = MetricMap{
	//cluster
	"cluster_cpu_utilisation":       ":node_cpu_utilisation:avg1m",
	"cluster_cpu_usage":             `round(:node_cpu_utilisation:avg1m * sum(node:node_num_cpu:sum), 0.001)`,
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

	"cluster_disk_size_usage":       `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"} - node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
	"cluster_disk_size_utilisation": `cluster:disk_utilization:ratio`,
	"cluster_disk_size_capacity":    `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
	"cluster_disk_size_available":   `sum(max(node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,

	"cluster_disk_inode_total":       `sum(node:node_inodes_total:)`,
	"cluster_disk_inode_usage":       `sum(node:node_inodes_total:) - sum(node:node_inodes_free:)`,
	"cluster_disk_inode_utilisation": `cluster:disk_inode_utilization:ratio`,

	"cluster_namespace_count": `count(kube_namespace_annotations)`,

	// cluster_pod_count = cluster_pod_running_count + cluster_pod_succeeded_count + cluster_pod_abnormal_count
	"cluster_pod_count":           `cluster:pod:sum`,
	"cluster_pod_quota":           `sum(max(kube_node_status_capacity_pods) by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"cluster_pod_utilisation":     `cluster:pod_utilization:ratio`,
	"cluster_pod_running_count":   `cluster:pod_running:count`,
	"cluster_pod_succeeded_count": `count(kube_pod_info unless on (pod) (kube_pod_status_phase{phase=~"Failed|Pending|Unknown|Running"} > 0) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"cluster_pod_abnormal_count":  `cluster:pod_abnormal:sum`,

	"cluster_node_online":  `sum(kube_node_status_condition{condition="Ready",status="true"})`,
	"cluster_node_offline": `cluster:node_offline:sum`,
	"cluster_node_total":   `sum(kube_node_status_condition{condition="Ready"})`,

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

	"cluster_cronjob_count":              `sum(kube_cronjob_labels)`,
	"cluster_pvc_count":                  `sum(kube_persistentvolumeclaim_info)`,
	"cluster_daemonset_count":            `sum(kube_daemonset_labels)`,
	"cluster_deployment_count":           `sum(kube_deployment_labels)`,
	"cluster_endpoint_count":             `sum(kube_endpoint_labels)`,
	"cluster_hpa_count":                  `sum(kube_hpa_labels)`,
	"cluster_job_count":                  `sum(kube_job_labels)`,
	"cluster_statefulset_count":          `sum(kube_statefulset_labels)`,
	"cluster_replicaset_count":           `count(kube_replicaset_created)`,
	"cluster_service_count":              `sum(kube_service_info)`,
	"cluster_secret_count":               `sum(kube_secret_info)`,
	"cluster_pv_count":                   `sum(kube_persistentvolume_labels)`,
	"cluster_ingresses_extensions_count": `sum(kube_ingress_labels)`,

	"cluster_load1":  `sum(node_load1{job="node-exporter"}) / sum(node:node_num_cpu:sum)`,
	"cluster_load5":  `sum(node_load5{job="node-exporter"}) / sum(node:node_num_cpu:sum)`,
	"cluster_load15": `sum(node_load15{job="node-exporter"}) / sum(node:node_num_cpu:sum)`,

	// cluster: New added in ks 2.0
	"cluster_pod_abnormal_ratio": `cluster:pod_abnormal:ratio`,
	"cluster_node_offline_ratio": `cluster:node_offline:ratio`,

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

	"node_disk_size_capacity":    `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1) by (device, node)) by (node)`,
	"node_disk_size_available":   `node:disk_space_available:$1`,
	"node_disk_size_usage":       `sum(max((node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"} - node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:$1) by (device, node)) by (node)`,
	"node_disk_size_utilisation": `node:disk_space_utilization:ratio$1`,

	"node_disk_inode_total":       `node:node_inodes_total:$1`,
	"node_disk_inode_usage":       `node:node_inodes_total:$1 - node:node_inodes_free:$1`,
	"node_disk_inode_utilisation": `node:disk_inode_utilization:ratio$1`,

	"node_pod_count":           `node:pod_count:sum$1`,
	"node_pod_quota":           `max(kube_node_status_capacity_pods$1) by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)`,
	"node_pod_utilisation":     `node:pod_utilization:ratio$1`,
	"node_pod_running_count":   `node:pod_running:count$1`,
	"node_pod_succeeded_count": `node:pod_succeeded:count$1`,
	"node_pod_abnormal_count":  `node:pod_abnormal:count$1`,

	// without log node: unless on(node) kube_node_labels{label_role="log"}
	"node_cpu_usage": `round(node:node_cpu_utilisation:avg1m$1 * node:node_num_cpu:sum$1, 0.001)`,

	"node_load1":  `node:load1:ratio$1`,
	"node_load5":  `node:load5:ratio$1`,
	"node_load15": `node:load15:ratio$1`,

	// New in ks 2.0
	"node_pod_abnormal_ratio": `node:pod_abnormal:ratio$1`,

	//namespace
	"namespace_cpu_usage":             `round(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", namespace=~"$1"}, 0.001)`,
	"namespace_memory_usage":          `namespace:container_memory_usage_bytes:sum{namespace!="", namespace=~"$1"}`,
	"namespace_memory_usage_wo_cache": `namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", namespace=~"$1"}`,
	"namespace_net_bytes_transmitted": `sum by (namespace) (irate(container_network_transmit_bytes_total{namespace!="", namespace=~"$1", pod_name!="", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))* on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_net_bytes_received":    `sum by (namespace) (irate(container_network_receive_bytes_total{namespace!="", namespace=~"$1", pod_name!="", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m])) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_count":             `sum(kube_pod_status_phase{phase!~"Failed|Succeeded", namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_running_count":     `sum(kube_pod_status_phase{phase="Running", namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_succeeded_count":   `sum(kube_pod_status_phase{phase="Succeeded", namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_abnormal_count":    `namespace:pod_abnormal:count{namespace!="", namespace=~"$1"}`,

	"namespace_roles_count_used":          `max(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="count/roles.rbac.authorization.k8s.io"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pvc_used":                  `max(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="persistentvolumeclaims"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_storage_request_used":      `max(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="requests.storage"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_service_loadbalancer_used": `max(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace=~"$1", resource="services.loadbalancers"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	// workarounds to calculate resource quota usage
	"namespace_deployment_count_used":           `count(kube_deployment_created{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_statefulset_count_used":          `count(kube_statefulset_created{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_daemonset_count_used":            `count(kube_daemonset_created{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_jobs_batch_count_used":           `count(kube_job_info{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_cronjobs_batch_count_used":       `count(kube_cronjob_created{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_pod_count_used":                  `sum(kube_pod_status_phase{phase!~"Failed|Succeeded", namespace!="", namespace=~"$1"}) by (namespace) * on (namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_service_count_used":              `count(kube_service_created{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_ingresses_extensions_count_used": `count(kube_ingress_created{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_pvc_count_used":                  `count(kube_persistentvolumeclaim_info{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_secret_count_used":               `count(kube_secret_created{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_configmap_count_used":            `count(kube_configmap_created{namespace="$1"}) by (namespace) * on(namespace) group_left(label_kubesphere_io_workspace) kube_namespace_labels`,
	"namespace_cpu_limit_used":                  `round(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", namespace="$1"}, 0.001)`,
	"namespace_cpu_request_used":                `round(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", namespace="$1"}, 0.001)`,
	"namespace_memory_limit_used":               `namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", namespace=~"$1"}`,
	"namespace_memory_request_used":             `namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", namespace=~"$1"}`,

	"namespace_configmap_count_hard":            `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/configmaps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_jobs_batch_count_hard":           `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/jobs.batch"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_roles_count_hard":                `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/roles.rbac.authorization.k8s.io"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_memory_limit_hard":               `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="limits.memory"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pvc_hard":                        `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="persistentvolumeclaims"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_memory_request_hard":             `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="requests.memory"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pvc_count_hard":                  `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/persistentvolumeclaims"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_cronjobs_batch_count_hard":       `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/cronjobs.batch"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_ingresses_extensions_count_hard": `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/ingresses.extensions"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_cpu_limit_hard":                  `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="limits.cpu"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_storage_request_hard":            `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="requests.storage"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_deployment_count_hard":           `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/deployments.apps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_pod_count_hard":                  `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/pods"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_statefulset_count_hard":          `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/statefulsets.apps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_daemonset_count_hard":            `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/daemonsets.apps"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_secret_count_hard":               `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/secrets"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_service_count_hard":              `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="count/services"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_cpu_request_hard":                `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="requests.cpu"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,
	"namespace_service_loadbalancer_hard":       `min(kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", namespace=~"$1", resource="services.loadbalancers"}) by (namespace, resource, type) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels)`,

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

	// New in ks 2.0
	"namespace_pod_abnormal_ratio":       `namespace:pod_abnormal:ratio{namespace!="", namespace=~"$1"}`,
	"namespace_resourcequota_used_ratio": `namespace:resourcequota_used:ratio{namespace!="", namespace=~"$1"}`,

	// pod
	"pod_cpu_usage":             `round(sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name!="", pod_name="$2", image!=""}[5m])) by (namespace, pod_name), 0.001)`,
	"pod_memory_usage":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name!="", pod_name="$2", image!=""}) by (namespace, pod_name)`,
	"pod_memory_usage_wo_cache": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name!="", pod_name="$2", image!=""} - container_memory_cache{job="kubelet", namespace="$1", pod_name!="", pod_name="$2",image!=""}) by (namespace, pod_name)`,
	"pod_net_bytes_transmitted": `sum by (namespace, pod_name) (irate(container_network_transmit_bytes_total{namespace="$1", pod_name!="", pod_name="$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))`,
	"pod_net_bytes_received":    `sum by (namespace, pod_name) (irate(container_network_receive_bytes_total{namespace="$1", pod_name!="", pod_name="$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))`,

	"pod_cpu_usage_all":             `round(sum(irate(container_cpu_usage_seconds_total{job="kubelet", namespace="$1", pod_name!="", pod_name=~"$2", image!=""}[5m])) by (namespace, pod_name), 0.001)`,
	"pod_memory_usage_all":          `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name!="", pod_name=~"$2",  image!=""}) by (namespace, pod_name)`,
	"pod_memory_usage_wo_cache_all": `sum(container_memory_usage_bytes{job="kubelet", namespace="$1", pod_name!="", pod_name=~"$2", image!=""} - container_memory_cache{job="kubelet", namespace="$1", pod_name!="", pod_name=~"$2", image!=""}) by (namespace, pod_name)`,
	"pod_net_bytes_transmitted_all": `sum by (namespace, pod_name) (irate(container_network_transmit_bytes_total{namespace="$1", pod_name!="", pod_name=~"$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))`,
	"pod_net_bytes_received_all":    `sum by (namespace, pod_name) (irate(container_network_receive_bytes_total{namespace="$1", pod_name!="", pod_name=~"$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]))`,

	"pod_cpu_usage_node":             `round(sum by (node, pod_name) (irate(container_cpu_usage_seconds_total{job="kubelet",pod_name!="", pod_name=~"$2", image!=""}[5m]) * on (namespace, pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name")), 0.001)`,
	"pod_memory_usage_node":          `sum by (node, pod_name) (container_memory_usage_bytes{job="kubelet",pod_name!="", pod_name=~"$2", image!=""} * on (namespace, pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name"))`,
	"pod_memory_usage_wo_cache_node": `sum by (node, pod_name) ((container_memory_usage_bytes{job="kubelet",pod_name!="", pod_name=~"$2", image!=""} - container_memory_cache{job="kubelet",pod_name!="", pod_name=~"$2", image!=""}) * on (namespace, pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name"))`,
	"pod_net_bytes_transmitted_node": `sum by (node, pod_name) (irate(container_network_transmit_bytes_total{pod_name!="", pod_name=~"$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name"))`,
	"pod_net_bytes_received_node":    `sum by (node, pod_name) (irate(container_network_receive_bytes_total{pod_name!="", pod_name=~"$2", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$3"}, "pod_name", "", "pod", "_name"))`,

	// workload
	// Join the "container_cpu_usage_seconds_total" metric with "kube_pod_owner" to calculate workload-level resource usage
	//
	// Note the name convention:
	// For hardware resource metrics, combine pod metric name with `workload_`
	// For k8s resource metrics, must specify the workload type in metric names
	"workload_pod_cpu_usage":             `round(namespace:workload_cpu_usage:sum{namespace="$2", workload=~"$3"}, 0.001)`,
	"workload_pod_memory_usage":          `namespace:workload_memory_usage:sum{namespace="$2", workload=~"$3"}`,
	"workload_pod_memory_usage_wo_cache": `namespace:workload_memory_usage_wo_cache:sum{namespace="$2", workload=~"$3"}`,
	"workload_pod_net_bytes_transmitted": `namespace:workload_net_bytes_transmitted:sum_irate{namespace="$2", workload=~"$3"}`,
	"workload_pod_net_bytes_received":    `namespace:workload_net_bytes_received:sum_irate{namespace="$2", workload=~"$3"}`,

	"workload_deployment_replica":            `label_join(sum (label_join(label_replace(kube_deployment_spec_replicas{namespace="$2", deployment=~"$3"}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_deployment_replica_available":  `label_join(sum (label_join(label_replace(kube_deployment_status_replicas_available{namespace="$2", deployment=~"$3"}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_statefulset_replica":           `label_join(sum (label_join(label_replace(kube_statefulset_replicas{namespace="$2", statefulset=~"$3"}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_statefulset_replica_available": `label_join(sum (label_join(label_replace(kube_statefulset_status_replicas_current{namespace="$2", statefulset=~"$3"}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_daemonset_replica":             `label_join(sum (label_join(label_replace(kube_daemonset_status_desired_number_scheduled{namespace="$2", daemonset=~"$3"}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_daemonset_replica_available":   `label_join(sum (label_join(label_replace(kube_daemonset_status_number_available{namespace="$2", daemonset=~"$3"}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,

	// New in ks 2.0
	"workload_deployment_unavailable_replicas_ratio":  `namespace:deployment_unavailable_replicas:ratio{namespace="$2", deployment=~"$3"}`,
	"workload_daemonset_unavailable_replicas_ratio":   `namespace:daemonset_unavailable_replicas:ratio{namespace="$2", daemonset=~"$3"}`,
	"workload_statefulset_unavailable_replicas_ratio": `namespace:statefulset_unavailable_replicas:ratio{namespace="$2", statefulset=~"$3"}`,

	// container
	"container_cpu_usage":             `round(sum(irate(container_cpu_usage_seconds_total{namespace="$1", pod_name="$2", container_name!="POD", container_name=~"$3"}[5m])) by (namespace, pod_name, container_name), 0.001)`,
	"container_memory_usage":          `sum(container_memory_usage_bytes{namespace="$1", pod_name="$2",  container_name!="POD", container_name=~"$3"}) by (namespace, pod_name, container_name)`,
	"container_memory_usage_wo_cache": `container_memory_usage_bytes{namespace="$1", pod_name="$2", container_name!="POD", container_name=~"$3"} - ignoring(id, image, endpoint, instance, job, name, service) container_memory_cache{namespace="$1", pod_name="$2", container_name!="POD", container_name=~"$3"}`,
	"container_net_bytes_transmitted": `sum(irate(container_network_transmit_bytes_total{job="kubelet", namespace="$1", pod_name="$2", container_name="POD", ` + ExcludedVirtualNetworkInterfaces + `}[5m])) by (namespace, pod_name, container_name)`,
	"container_net_bytes_received":    `sum(irate(container_network_receive_bytes_total{job="kubelet", namespace="$1", pod_name="$2", container_name="POD", ` + ExcludedVirtualNetworkInterfaces + `}[5m])) by (namespace, pod_name, container_name)`,

	"container_cpu_usage_node":             `round(sum by (node, pod_name, container_name) (irate(container_cpu_usage_seconds_total{job="kubelet", pod_name="$2", container_name!="POD", container_name!="", container_name=~"$3", image!=""}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name")), 0.001)`,
	"container_memory_usage_node":          `sum by (node, pod_name, container_name) (container_memory_usage_bytes{job="kubelet", pod_name="$2", container_name!="POD", container_name!="", container_name=~"$3", image!=""} * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name"))`,
	"container_memory_usage_wo_cache_node": `sum by (node, pod_name, container_name) ((container_memory_usage_bytes{job="kubelet", pod_name="$2", container_name!="POD", container_name!="", container_name=~"$3", image!=""} - container_memory_cache{job="kubelet", pod_name="$2", container_name!="POD", container_name!="", container_name=~"$3", image!=""}) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name"))`,
	"container_net_bytes_transmitted_node": `sum by (node, pod_name, container_name) (irate(container_network_transmit_bytes_total{job="kubelet", ` + ExcludedVirtualNetworkInterfaces + `, pod_name="$2", container_name="POD", container_name!="", image!=""}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name"))`,
	"container_net_bytes_received_node":    `sum by (node, pod_name, container_name) (irate(container_network_receive_bytes_total{job="kubelet", ` + ExcludedVirtualNetworkInterfaces + `, pod_name="$2", container_name="POD", container_name!="", image!=""}[5m]) * on (pod_name) group_left(node) label_join(node_namespace_pod:kube_pod_info:{node="$1"}, "pod_name", "", "pod", "_name"))`,

	// workspace
	"workspace_cpu_usage":             `round(sum(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", namespace$1, label_kubesphere_io_workspace$2}), 0.001)`,
	"workspace_memory_usage":          `sum(namespace:container_memory_usage_bytes:sum{namespace!="", namespace$1, label_kubesphere_io_workspace$2})`,
	"workspace_memory_usage_wo_cache": `sum(namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", namespace$1, label_kubesphere_io_workspace$2})`,
	"workspace_net_bytes_transmitted": `sum(sum by (namespace) (irate(container_network_transmit_bytes_total{namespace!="", namespace$1, pod_name!="", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m])))`,
	"workspace_net_bytes_received":    `sum(sum by (namespace) (irate(container_network_receive_bytes_total{namespace!="", namespace$1, pod_name!="", ` + ExcludedVirtualNetworkInterfaces + `, job="kubelet"}[5m])))`,
	"workspace_pod_count":             `sum(kube_pod_status_phase{phase!~"Failed|Succeeded", namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_pod_running_count":     `sum(kube_pod_status_phase{phase="Running", namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_pod_succeeded_count":   `sum(kube_pod_status_phase{phase="Succeeded", namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_pod_abnormal_count":    `count((kube_pod_info{node!="", namespace$1} unless on (pod, namespace) (kube_pod_status_phase{job="kube-state-metrics", phase="Succeeded"}>0) unless on (pod, namespace) ((kube_pod_status_ready{job="kube-state-metrics", condition="true"}>0) and on (pod, namespace) (kube_pod_status_phase{job="kube-state-metrics", phase="Running"}>0)) unless on (pod, namespace) (kube_pod_container_status_waiting_reason{job="kube-state-metrics", reason="ContainerCreating"}>0)) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,

	"workspace_configmap_count_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/configmaps"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_jobs_batch_count_used":           `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/jobs.batch"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_roles_count_used":                `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/roles.rbac.authorization.k8s.io"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_memory_limit_used":               `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="limits.memory"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_pvc_used":                        `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="persistentvolumeclaims"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_memory_request_used":             `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="requests.memory"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_pvc_count_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/persistentvolumeclaims"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_cronjobs_batch_count_used":       `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/cronjobs.batch"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_ingresses_extensions_count_used": `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/ingresses.extensions"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_cpu_limit_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="limits.cpu"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_storage_request_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="requests.storage"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_deployment_count_used":           `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/deployments.apps"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_pod_count_used":                  `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/pods"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_statefulset_count_used":          `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/statefulsets.apps"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_daemonset_count_used":            `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/daemonsets.apps"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_secret_count_used":               `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/secrets"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_service_count_used":              `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="count/services"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_cpu_request_used":                `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="requests.cpu"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,
	"workspace_service_loadbalancer_used":       `sum(kube_resourcequota{resourcequota!="quota", type="used", namespace!="", namespace$1, resource="services.loadbalancers"} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2})) by (resource, type)`,

	"workspace_ingresses_extensions_count": `sum(kube_ingress_labels{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,

	"workspace_cronjob_count":     `sum(kube_cronjob_labels{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_pvc_count":         `sum(kube_persistentvolumeclaim_info{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_daemonset_count":   `sum(kube_daemonset_labels{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_deployment_count":  `sum(kube_deployment_labels{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_endpoint_count":    `sum(kube_endpoint_labels{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_hpa_count":         `sum(kube_hpa_labels{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_job_count":         `sum(kube_job_labels{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_statefulset_count": `sum(kube_statefulset_labels{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_replicaset_count":  `count(kube_replicaset_created{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_service_count":     `sum(kube_service_info{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,
	"workspace_secret_count":      `sum(kube_secret_info{namespace!="", namespace$1} * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,

	"workspace_all_project_count": `count(kube_namespace_annotations)`,

	// New in ks 2.0
	"workspace_pod_abnormal_ratio": `count((kube_pod_info{node!="", namespace$1} unless on (pod, namespace) (kube_pod_status_phase{job="kube-state-metrics", phase="Succeeded"}>0) unless on (pod, namespace) ((kube_pod_status_ready{job="kube-state-metrics", condition="true"}>0) and on (pod, namespace) (kube_pod_status_phase{job="kube-state-metrics", phase="Running"}>0)) unless on (pod, namespace) (kube_pod_container_status_waiting_reason{job="kube-state-metrics", reason="ContainerCreating"}>0)) / sum(kube_pod_status_phase{phase!~"Succeeded", namespace!="", namespace$1}) * on (namespace) group_left(label_kubesphere_io_workspace)(kube_namespace_labels{label_kubesphere_io_workspace$2}))`,

	// PVC
	"pvc_inodes_available":      `max (kubelet_volume_stats_inodes_free{namespace="$1",persistentvolumeclaim="$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_inodes_used":           `max (kubelet_volume_stats_inodes_used{namespace="$1", persistentvolumeclaim="$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_inodes_total":          `max (kubelet_volume_stats_inodes{namespace="$1", persistentvolumeclaim="$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_inodes_utilisation":    `max (kubelet_volume_stats_inodes_used{namespace="$1", persistentvolumeclaim="$2"}/kubelet_volume_stats_inodes{namespace="$1", persistentvolumeclaim="$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_bytes_available":       `max (kubelet_volume_stats_available_bytes{namespace="$1", persistentvolumeclaim="$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_bytes_used":            `max (kubelet_volume_stats_used_bytes{namespace="$1", persistentvolumeclaim="$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_bytes_total":           `max (kubelet_volume_stats_capacity_bytes{namespace="$1", persistentvolumeclaim="$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_bytes_utilisation":     `max (kubelet_volume_stats_used_bytes{namespace="$1", persistentvolumeclaim="$2"}/kubelet_volume_stats_capacity_bytes{namespace="$1", persistentvolumeclaim="$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_inodes_available_ns":   `max (kubelet_volume_stats_inodes_free{namespace="$1",persistentvolumeclaim=~"$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_inodes_used_ns":        `max (kubelet_volume_stats_inodes_used{namespace="$1",persistentvolumeclaim=~"$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_inodes_total_ns":       `max (kubelet_volume_stats_inodes{namespace="$1",persistentvolumeclaim=~"$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_inodes_utilisation_ns": `max (kubelet_volume_stats_inodes_used{namespace="$1", persistentvolumeclaim=~"$2"}/kubelet_volume_stats_inodes{namespace="$1", persistentvolumeclaim=~"$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_bytes_available_ns":    `max (kubelet_volume_stats_available_bytes{namespace="$1",persistentvolumeclaim=~"$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_bytes_used_ns":         `max (kubelet_volume_stats_used_bytes{namespace="$1",persistentvolumeclaim=~"$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_bytes_total_ns":        `max (kubelet_volume_stats_capacity_bytes{namespace="$1",persistentvolumeclaim=~"$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_bytes_utilisation_ns":  `max (kubelet_volume_stats_used_bytes{namespace="$1", persistentvolumeclaim=~"$2"}/kubelet_volume_stats_capacity_bytes{namespace="$1", persistentvolumeclaim=~"$2"})by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info`,
	"pvc_inodes_available_sc":   `max (kubelet_volume_stats_inodes_free)by(namespace,persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info{storageclass="$1"}`,
	"pvc_inodes_used_sc":        `max (kubelet_volume_stats_inodes_used)by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info{storageclass="$1"}`,
	"pvc_inodes_total_sc":       `max (kubelet_volume_stats_inodes)by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info{storageclass="$1"}`,
	"pvc_inodes_utilisation_sc": `max (kubelet_volume_stats_inodes_used/kubelet_volume_stats_inodes)by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info{storageclass="$1"}`,
	"pvc_bytes_available_sc":    `max (kubelet_volume_stats_available_bytes)by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info{storageclass="$1"}`,
	"pvc_bytes_used_sc":         `max (kubelet_volume_stats_used_bytes)by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info{storageclass="$1"}`,
	"pvc_bytes_total_sc":        `max (kubelet_volume_stats_capacity_bytes)by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info{storageclass="$1"}`,
	"pvc_bytes_utilisation_sc":  `max (kubelet_volume_stats_used_bytes/kubelet_volume_stats_capacity_bytes)by(namespace, persistentvolumeclaim)*on(namespace, persistentvolumeclaim)group_left(storageclass)kube_persistentvolumeclaim_info{storageclass="$1"}`,

	// component
	"etcd_server_list":                           `label_replace(up{job="etcd"}, "node_ip", "$1", "instance", "(.*):.*")`,
	"etcd_server_total":                          `count(up{job="etcd"})`,
	"etcd_server_up_total":                       `etcd:up:sum`,
	"etcd_server_has_leader":                     `label_replace(etcd_server_has_leader, "node_ip", "$1", "instance", "(.*):.*")`,
	"etcd_server_leader_changes":                 `label_replace(etcd:etcd_server_leader_changes_seen:sum_changes, "node_ip", "$1", "node", "(.*)")`,
	"etcd_server_proposals_failed_rate":          `avg(etcd:etcd_server_proposals_failed:sum_irate)`,
	"etcd_server_proposals_applied_rate":         `avg(etcd:etcd_server_proposals_applied:sum_irate)`,
	"etcd_server_proposals_committed_rate":       `avg(etcd:etcd_server_proposals_committed:sum_irate)`,
	"etcd_server_proposals_pending_count":        `avg(etcd:etcd_server_proposals_pending:sum)`,
	"etcd_mvcc_db_size":                          `avg(etcd:etcd_debugging_mvcc_db_total_size:sum)`,
	"etcd_network_client_grpc_received_bytes":    `sum(etcd:etcd_network_client_grpc_received_bytes:sum_irate)`,
	"etcd_network_client_grpc_sent_bytes":        `sum(etcd:etcd_network_client_grpc_sent_bytes:sum_irate)`,
	"etcd_grpc_call_rate":                        `sum(etcd:grpc_server_started:sum_irate)`,
	"etcd_grpc_call_failed_rate":                 `sum(etcd:grpc_server_handled:sum_irate)`,
	"etcd_grpc_server_msg_received_rate":         `sum(etcd:grpc_server_msg_received:sum_irate)`,
	"etcd_grpc_server_msg_sent_rate":             `sum(etcd:grpc_server_msg_sent:sum_irate)`,
	"etcd_disk_wal_fsync_duration":               `avg(etcd:etcd_disk_wal_fsync_duration:avg)`,
	"etcd_disk_wal_fsync_duration_quantile":      `avg(etcd:etcd_disk_wal_fsync_duration:histogram_quantile) by (quantile)`,
	"etcd_disk_backend_commit_duration":          `avg(etcd:etcd_disk_backend_commit_duration:avg)`,
	"etcd_disk_backend_commit_duration_quantile": `avg(etcd:etcd_disk_backend_commit_duration:histogram_quantile) by (quantile)`,

	"apiserver_up_sum":                    `apiserver:up:sum`,
	"apiserver_request_rate":              `apiserver:apiserver_request_count:sum_irate`,
	"apiserver_request_by_verb_rate":      `apiserver:apiserver_request_count:sum_verb_irate`,
	"apiserver_request_latencies":         `apiserver:apiserver_request_latencies:avg`,
	"apiserver_request_by_verb_latencies": `apiserver:apiserver_request_latencies:avg_by_verb`,

	"scheduler_up_sum":                          `scheduler:up:sum`,
	"scheduler_schedule_attempts":               `scheduler:scheduler_schedule_attempts:sum`,
	"scheduler_schedule_attempt_rate":           `scheduler:scheduler_schedule_attempts:sum_rate`,
	"scheduler_e2e_scheduling_latency":          `scheduler:scheduler_e2e_scheduling_latency:avg`,
	"scheduler_e2e_scheduling_latency_quantile": `scheduler:scheduler_e2e_scheduling_latency:histogram_quantile`,

	"controller_manager_up_sum": `controller_manager:up:sum`,

	"coredns_up_sum":                          `coredns:up:sum`,
	"coredns_cache_hits":                      `coredns:coredns_cache_hits_total:sum_irate`,
	"coredns_cache_misses":                    `coredns:coredns_cache_misses:sum_irate`,
	"coredns_dns_request_rate":                `coredns:coredns_dns_request_count:sum_irate`,
	"coredns_dns_request_duration":            `coredns:coredns_dns_request_duration:avg`,
	"coredns_dns_request_duration_quantile":   `coredns:coredns_dns_request_duration:histogram_quantile`,
	"coredns_dns_request_by_type_rate":        `coredns:coredns_dns_request_type_count:sum_irate`,
	"coredns_dns_request_by_rcode_rate":       `coredns:coredns_dns_response_rcode_count:sum_irate`,
	"coredns_panic_rate":                      `coredns:coredns_panic_count:sum_irate`,
	"coredns_proxy_request_rate":              `coredns:coredns_proxy_request_count:sum_irate`,
	"coredns_proxy_request_duration":          `coredns:coredns_proxy_request_duration:avg`,
	"coredns_proxy_request_duration_quantile": `coredns:coredns_proxy_request_duration:histogram_quantile`,

	"prometheus_up_sum":                          `prometheus:up:sum`,
	"prometheus_tsdb_head_samples_appended_rate": `prometheus:prometheus_tsdb_head_samples_appended:sum_rate`,
}
