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

package prometheus

import (
	"fmt"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"strings"
)

const (
	StatefulSet = "StatefulSet"
	DaemonSet   = "DaemonSet"
	Deployment  = "Deployment"
)

var promQLTemplates = map[string]string{
	//cluster
	"cluster_cpu_utilisation":            ":node_cpu_utilisation:avg1m",
	"cluster_cpu_usage":                  `round(:node_cpu_utilisation:avg1m * sum(node:node_num_cpu:sum), 0.001)`,
	"cluster_cpu_total":                  "sum(node:node_num_cpu:sum)",
	"cluster_memory_utilisation":         ":node_memory_utilisation:",
	"cluster_memory_available":           "sum(node:node_memory_bytes_available:sum)",
	"cluster_memory_total":               "sum(node:node_memory_bytes_total:sum)",
	"cluster_memory_usage_wo_cache":      "sum(node:node_memory_bytes_total:sum) - sum(node:node_memory_bytes_available:sum)",
	"cluster_net_utilisation":            ":node_net_utilisation:sum_irate",
	"cluster_net_bytes_transmitted":      "sum(node:node_net_bytes_transmitted:sum_irate)",
	"cluster_net_bytes_received":         "sum(node:node_net_bytes_received:sum_irate)",
	"cluster_disk_read_iops":             "sum(node:data_volume_iops_reads:sum)",
	"cluster_disk_write_iops":            "sum(node:data_volume_iops_writes:sum)",
	"cluster_disk_read_throughput":       "sum(node:data_volume_throughput_bytes_read:sum)",
	"cluster_disk_write_throughput":      "sum(node:data_volume_throughput_bytes_written:sum)",
	"cluster_disk_size_usage":            `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"} - node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
	"cluster_disk_size_utilisation":      `cluster:disk_utilization:ratio`,
	"cluster_disk_size_capacity":         `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
	"cluster_disk_size_available":        `sum(max(node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) by (device, instance))`,
	"cluster_disk_inode_total":           `sum(node:node_inodes_total:)`,
	"cluster_disk_inode_usage":           `sum(node:node_inodes_total:) - sum(node:node_inodes_free:)`,
	"cluster_disk_inode_utilisation":     `cluster:disk_inode_utilization:ratio`,
	"cluster_namespace_count":            `count(kube_namespace_labels)`,
	"cluster_pod_count":                  `cluster:pod:sum`,
	"cluster_pod_quota":                  `sum(max(kube_node_status_capacity{resource="pods"}) by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"cluster_pod_utilisation":            `cluster:pod_utilization:ratio`,
	"cluster_pod_running_count":          `cluster:pod_running:count`,
	"cluster_pod_succeeded_count":        `count(kube_pod_info unless on (pod) (kube_pod_status_phase{phase=~"Failed|Pending|Unknown|Running"} > 0) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0))`,
	"cluster_pod_abnormal_count":         `cluster:pod_abnormal:sum`,
	"cluster_node_online":                `sum(kube_node_status_condition{condition="Ready",status="true"})`,
	"cluster_node_offline":               `cluster:node_offline:sum`,
	"cluster_node_total":                 `sum(kube_node_status_condition{condition="Ready"})`,
	"cluster_cronjob_count":              `sum(kube_cronjob_labels)`,
	"cluster_pvc_count":                  `sum(kube_persistentvolumeclaim_info)`,
	"cluster_daemonset_count":            `sum(kube_daemonset_labels)`,
	"cluster_deployment_count":           `sum(kube_deployment_labels)`,
	"cluster_endpoint_count":             `sum(kube_endpoint_labels)`,
	"cluster_hpa_count":                  `sum(kube_hpa_labels)`,
	"cluster_job_count":                  `sum(kube_job_labels)`,
	"cluster_statefulset_count":          `sum(kube_statefulset_labels)`,
	"cluster_replicaset_count":           `count(kube_replicaset_labels)`,
	"cluster_service_count":              `sum(kube_service_info)`,
	"cluster_secret_count":               `sum(kube_secret_info)`,
	"cluster_pv_count":                   `sum(kube_persistentvolume_labels)`,
	"cluster_ingresses_extensions_count": `sum(kube_ingress_labels)`,
	"cluster_load1":                      `sum(node_load1{job="node-exporter"}) / sum(node:node_num_cpu:sum)`,
	"cluster_load5":                      `sum(node_load5{job="node-exporter"}) / sum(node:node_num_cpu:sum)`,
	"cluster_load15":                     `sum(node_load15{job="node-exporter"}) / sum(node:node_num_cpu:sum)`,
	"cluster_pod_abnormal_ratio":         `cluster:pod_abnormal:ratio`,
	"cluster_node_offline_ratio":         `cluster:node_offline:ratio`,

	//node
	"node_cpu_utilisation":        "node:node_cpu_utilisation:avg1m{$1}",
	"node_cpu_total":              "node:node_num_cpu:sum{$1}",
	"node_memory_utilisation":     "node:node_memory_utilisation:{$1}",
	"node_memory_available":       "node:node_memory_bytes_available:sum{$1}",
	"node_memory_total":           "node:node_memory_bytes_total:sum{$1}",
	"node_memory_usage_wo_cache":  "node:node_memory_bytes_total:sum{$1} - node:node_memory_bytes_available:sum{$1}",
	"node_net_utilisation":        "node:node_net_utilisation:sum_irate{$1}",
	"node_net_bytes_transmitted":  "node:node_net_bytes_transmitted:sum_irate{$1}",
	"node_net_bytes_received":     "node:node_net_bytes_received:sum_irate{$1}",
	"node_disk_read_iops":         "node:data_volume_iops_reads:sum{$1}",
	"node_disk_write_iops":        "node:data_volume_iops_writes:sum{$1}",
	"node_disk_read_throughput":   "node:data_volume_throughput_bytes_read:sum{$1}",
	"node_disk_write_throughput":  "node:data_volume_throughput_bytes_written:sum{$1}",
	"node_disk_size_capacity":     `sum(max(node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"} * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{$1}) by (device, node)) by (node)`,
	"node_disk_size_available":    `node:disk_space_available:{$1}`,
	"node_disk_size_usage":        `sum(max((node_filesystem_size_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"} - node_filesystem_avail_bytes{device=~"/dev/.*", device!~"/dev/loop\\d+", job="node-exporter"}) * on (namespace, pod) group_left(node) node_namespace_pod:kube_pod_info:{$1}) by (device, node)) by (node)`,
	"node_disk_size_utilisation":  `node:disk_space_utilization:ratio{$1}`,
	"node_disk_inode_total":       `node:node_inodes_total:{$1}`,
	"node_disk_inode_usage":       `node:node_inodes_total:{$1} - node:node_inodes_free:{$1}`,
	"node_disk_inode_utilisation": `node:disk_inode_utilization:ratio{$1}`,
	"node_pod_count":              `node:pod_count:sum{$1}`,
	"node_pod_quota":              `max(kube_node_status_capacity{resource="pods",$1}) by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)`,
	"node_pod_utilisation":        `node:pod_utilization:ratio{$1}`,
	"node_pod_running_count":      `node:pod_running:count{$1}`,
	"node_pod_succeeded_count":    `node:pod_succeeded:count{$1}`,
	"node_pod_abnormal_count":     `node:pod_abnormal:count{$1}`,
	"node_cpu_usage":              `round(node:node_cpu_utilisation:avg1m{$1} * node:node_num_cpu:sum{$1}, 0.001)`,
	"node_load1":                  `node:load1:ratio{$1}`,
	"node_load5":                  `node:load5:ratio{$1}`,
	"node_load15":                 `node:load15:ratio{$1}`,
	"node_pod_abnormal_ratio":     `node:pod_abnormal:ratio{$1}`,

	// workspace
	"workspace_cpu_usage":                  `round(sum by (workspace) (namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", $1}), 0.001)`,
	"workspace_memory_usage":               `sum by (workspace) (namespace:container_memory_usage_bytes:sum{namespace!="", $1})`,
	"workspace_memory_usage_wo_cache":      `sum by (workspace) (namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", $1})`,
	"workspace_net_bytes_transmitted":      `sum by (workspace) (sum by (namespace) (irate(container_network_transmit_bytes_total{namespace!="", pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m])) * on (namespace) group_left(workspace) kube_namespace_labels{$1}) or on(workspace) max by(workspace) (kube_namespace_labels{$1} * 0)`,
	"workspace_net_bytes_received":         `sum by (workspace) (sum by (namespace) (irate(container_network_receive_bytes_total{namespace!="", pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m])) * on (namespace) group_left(workspace) kube_namespace_labels{$1}) or on(workspace) max by(workspace) (kube_namespace_labels{$1} * 0)`,
	"workspace_pod_count":                  `sum by (workspace) (kube_pod_status_phase{phase!~"Failed|Succeeded", namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1})) or on(workspace) max by(workspace) (kube_namespace_labels{$1} * 0)`,
	"workspace_pod_running_count":          `sum by (workspace) (kube_pod_status_phase{phase="Running", namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1})) or on(workspace) max by(workspace) (kube_namespace_labels{$1} * 0)`,
	"workspace_pod_succeeded_count":        `sum by (workspace) (kube_pod_status_phase{phase="Succeeded", namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1})) or on(workspace) max by(workspace) (kube_namespace_labels{$1} * 0)`,
	"workspace_pod_abnormal_count":         `count by (workspace) ((kube_pod_info{node!=""} unless on (pod, namespace) (kube_pod_status_phase{job="kube-state-metrics", phase="Succeeded"}>0) unless on (pod, namespace) ((kube_pod_status_ready{job="kube-state-metrics", condition="true"}>0) and on (pod, namespace) (kube_pod_status_phase{job="kube-state-metrics", phase="Running"}>0)) unless on (pod, namespace) (kube_pod_container_status_waiting_reason{job="kube-state-metrics", reason="ContainerCreating"}>0)) * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_ingresses_extensions_count": `sum by (workspace) (kube_ingress_labels{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_cronjob_count":              `sum by (workspace) (kube_cronjob_labels{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_pvc_count":                  `sum by (workspace) (kube_persistentvolumeclaim_info{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_daemonset_count":            `sum by (workspace) (kube_daemonset_labels{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_deployment_count":           `sum by (workspace) (kube_deployment_labels{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_endpoint_count":             `sum by (workspace) (kube_endpoint_labels{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_hpa_count":                  `sum by (workspace) (kube_hpa_labels{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_job_count":                  `sum by (workspace) (kube_job_labels{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_statefulset_count":          `sum by (workspace) (kube_statefulset_labels{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_replicaset_count":           `count by (workspace) (kube_replicaset_labels{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_service_count":              `sum by (workspace) (kube_service_info{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_secret_count":               `sum by (workspace) (kube_secret_info{namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,
	"workspace_pod_abnormal_ratio":         `count by (workspace) ((kube_pod_info{node!=""} unless on (pod, namespace) (kube_pod_status_phase{job="kube-state-metrics", phase="Succeeded"}>0) unless on (pod, namespace) ((kube_pod_status_ready{job="kube-state-metrics", condition="true"}>0) and on (pod, namespace) (kube_pod_status_phase{job="kube-state-metrics", phase="Running"}>0)) unless on (pod, namespace) (kube_pod_container_status_waiting_reason{job="kube-state-metrics", reason="ContainerCreating"}>0)) * on (namespace) group_left(workspace) kube_namespace_labels{$1}) / sum by (workspace) (kube_pod_status_phase{phase!="Succeeded", namespace!=""} * on (namespace) group_left(workspace)(kube_namespace_labels{$1}))`,

	//namespace
	"namespace_cpu_usage":                  `round(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", $1}, 0.001)`,
	"namespace_memory_usage":               `namespace:container_memory_usage_bytes:sum{namespace!="", $1}`,
	"namespace_memory_usage_wo_cache":      `namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", $1}`,
	"namespace_net_bytes_transmitted":      `sum by (namespace) (irate(container_network_transmit_bytes_total{namespace!="", pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m]) * on (namespace) group_left(workspace) kube_namespace_labels{$1}) or on(namespace) max by(namespace) (kube_namespace_labels{$1} * 0)`,
	"namespace_net_bytes_received":         `sum by (namespace) (irate(container_network_receive_bytes_total{namespace!="", pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m]) * on (namespace) group_left(workspace) kube_namespace_labels{$1}) or on(namespace) max by(namespace) (kube_namespace_labels{$1} * 0)`,
	"namespace_pod_count":                  `sum by (namespace) (kube_pod_status_phase{phase!~"Failed|Succeeded", namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1}) or on(namespace) max by(namespace) (kube_namespace_labels{$1} * 0)`,
	"namespace_pod_running_count":          `sum by (namespace) (kube_pod_status_phase{phase="Running", namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1}) or on(namespace) max by(namespace) (kube_namespace_labels{$1} * 0)`,
	"namespace_pod_succeeded_count":        `sum by (namespace) (kube_pod_status_phase{phase="Succeeded", namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1}) or on(namespace) max by(namespace) (kube_namespace_labels{$1} * 0)`,
	"namespace_pod_abnormal_count":         `namespace:pod_abnormal:count{namespace!="", $1}`,
	"namespace_pod_abnormal_ratio":         `namespace:pod_abnormal:ratio{namespace!="", $1}`,
	"namespace_memory_limit_hard":          `min by (namespace) (kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", resource="limits.memory"} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_cpu_limit_hard":             `min by (namespace) (kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", resource="limits.cpu"} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_pod_count_hard":             `min by (namespace) (kube_resourcequota{resourcequota!="quota", type="hard", namespace!="", resource="count/pods"} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_cronjob_count":              `sum by (namespace) (kube_cronjob_labels{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_pvc_count":                  `sum by (namespace) (kube_persistentvolumeclaim_info{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_daemonset_count":            `sum by (namespace) (kube_daemonset_labels{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_deployment_count":           `sum by (namespace) (kube_deployment_labels{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_endpoint_count":             `sum by (namespace) (kube_endpoint_labels{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_hpa_count":                  `sum by (namespace) (kube_hpa_labels{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_job_count":                  `sum by (namespace) (kube_job_labels{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_statefulset_count":          `sum by (namespace) (kube_statefulset_labels{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_replicaset_count":           `count by (namespace) (kube_replicaset_labels{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_service_count":              `sum by (namespace) (kube_service_info{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_secret_count":               `sum by (namespace) (kube_secret_info{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_configmap_count":            `sum by (namespace) (kube_configmap_info{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_ingresses_extensions_count": `sum by (namespace) (kube_ingress_labels{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,
	"namespace_s2ibuilder_count":           `sum by (namespace) (s2i_s2ibuilder_created{namespace!=""} * on (namespace) group_left(workspace) kube_namespace_labels{$1})`,

	// workload
	"workload_cpu_usage":             `round(namespace:workload_cpu_usage:sum{$1}, 0.001)`,
	"workload_memory_usage":          `namespace:workload_memory_usage:sum{$1}`,
	"workload_memory_usage_wo_cache": `namespace:workload_memory_usage_wo_cache:sum{$1}`,
	"workload_net_bytes_transmitted": `namespace:workload_net_bytes_transmitted:sum_irate{$1}`,
	"workload_net_bytes_received":    `namespace:workload_net_bytes_received:sum_irate{$1}`,

	"workload_deployment_replica":                     `label_join(sum (label_join(label_replace(kube_deployment_spec_replicas{$2}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_deployment_replica_available":           `label_join(sum (label_join(label_replace(kube_deployment_status_replicas_available{$2}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_statefulset_replica":                    `label_join(sum (label_join(label_replace(kube_statefulset_replicas{$2}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_statefulset_replica_available":          `label_join(sum (label_join(label_replace(kube_statefulset_status_replicas_current{$2}, "owner_kind", "StatefulSet", "", ""), "workload", "", "statefulset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_daemonset_replica":                      `label_join(sum (label_join(label_replace(kube_daemonset_status_desired_number_scheduled{$2}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_daemonset_replica_available":            `label_join(sum (label_join(label_replace(kube_daemonset_status_number_available{$2}, "owner_kind", "DaemonSet", "", ""), "workload", "", "daemonset")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"workload_deployment_unavailable_replicas_ratio":  `namespace:deployment_unavailable_replicas:ratio{$1}`,
	"workload_daemonset_unavailable_replicas_ratio":   `namespace:daemonset_unavailable_replicas:ratio{$1}`,
	"workload_statefulset_unavailable_replicas_ratio": `namespace:statefulset_unavailable_replicas:ratio{$1}`,

	// pod
	"pod_cpu_usage":             `round(sum by (namespace, pod) (irate(container_cpu_usage_seconds_total{job="kubelet", pod!="", image!=""}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}, 0.001)`,
	"pod_memory_usage":          `sum by (namespace, pod) (container_memory_usage_bytes{job="kubelet", pod!="", image!=""}) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}`,
	"pod_memory_usage_wo_cache": `sum by (namespace, pod) (container_memory_working_set_bytes{job="kubelet", pod!="", image!=""}) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}`,
	"pod_net_bytes_transmitted": `sum by (namespace, pod) (irate(container_network_transmit_bytes_total{pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}`,
	"pod_net_bytes_received":    `sum by (namespace, pod) (irate(container_network_receive_bytes_total{pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet"}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1} * on (namespace, pod) group_left(node) kube_pod_info{$2}`,

	// container
	"container_cpu_usage":             `round(sum by (namespace, pod, container) (irate(container_cpu_usage_seconds_total{job="kubelet", container!="POD", container!="", image!="", $1}[5m])), 0.001)`,
	"container_memory_usage":          `sum by (namespace, pod, container) (container_memory_usage_bytes{job="kubelet", container!="POD", container!="", image!="", $1})`,
	"container_memory_usage_wo_cache": `sum by (namespace, pod, container) (container_memory_working_set_bytes{job="kubelet", container!="POD", container!="", image!="", $1})`,

	// pvc
	"pvc_inodes_available":   `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes_free) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{$1}`,
	"pvc_inodes_used":        `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes_used) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{$1}`,
	"pvc_inodes_total":       `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{$1}`,
	"pvc_inodes_utilisation": `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes_used / kubelet_volume_stats_inodes) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{$1}`,
	"pvc_bytes_available":    `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_available_bytes) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{$1}`,
	"pvc_bytes_used":         `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_used_bytes) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{$1}`,
	"pvc_bytes_total":        `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_capacity_bytes) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{$1}`,
	"pvc_bytes_utilisation":  `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_used_bytes / kubelet_volume_stats_capacity_bytes) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{$1}`,

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
	"apiserver_request_rate":              `apiserver:apiserver_request_total:sum_irate`,
	"apiserver_request_by_verb_rate":      `apiserver:apiserver_request_total:sum_verb_irate`,
	"apiserver_request_latencies":         `apiserver:apiserver_request_duration:avg`,
	"apiserver_request_by_verb_latencies": `apiserver:apiserver_request_duration:avg_by_verb`,

	"scheduler_up_sum":                          `scheduler:up:sum`,
	"scheduler_schedule_attempts":               `scheduler:scheduler_schedule_attempts:sum`,
	"scheduler_schedule_attempt_rate":           `scheduler:scheduler_schedule_attempts:sum_rate`,
	"scheduler_e2e_scheduling_latency":          `scheduler:scheduler_e2e_scheduling_duration:avg`,
	"scheduler_e2e_scheduling_latency_quantile": `scheduler:scheduler_e2e_scheduling_duration:histogram_quantile`,
}

func makeExpr(metric string, opts monitoring.QueryOptions) string {
	tmpl := promQLTemplates[metric]
	switch opts.Level {
	case monitoring.LevelCluster:
		return tmpl
	case monitoring.LevelNode:
		return makeNodeMetricExpr(tmpl, opts)
	case monitoring.LevelWorkspace:
		return makeWorkspaceMetricExpr(tmpl, opts)
	case monitoring.LevelNamespace:
		return makeNamespaceMetricExpr(tmpl, opts)
	case monitoring.LevelWorkload:
		return makeWorkloadMetricExpr(metric, tmpl, opts)
	case monitoring.LevelPod:
		return makePodMetricExpr(tmpl, opts)
	case monitoring.LevelContainer:
		return makeContainerMetricExpr(tmpl, opts)
	case monitoring.LevelPVC:
		return makePVCMetricExpr(tmpl, opts)
	case monitoring.LevelComponent:
		return tmpl
	default:
		return tmpl
	}
}

func makeNodeMetricExpr(tmpl string, o monitoring.QueryOptions) string {
	var nodeSelector string
	if o.NodeName != "" {
		nodeSelector = fmt.Sprintf(`node="%s"`, o.NodeName)
	} else {
		nodeSelector = fmt.Sprintf(`node=~"%s"`, o.ResourceFilter)
	}
	return strings.Replace(tmpl, "$1", nodeSelector, -1)
}

func makeWorkspaceMetricExpr(tmpl string, o monitoring.QueryOptions) string {
	var workspaceSelector string
	if o.WorkspaceName != "" {
		workspaceSelector = fmt.Sprintf(`workspace="%s"`, o.WorkspaceName)
	} else {
		workspaceSelector = fmt.Sprintf(`workspace=~"%s", workspace!=""`, o.ResourceFilter)
	}
	return strings.Replace(tmpl, "$1", workspaceSelector, -1)
}

func makeNamespaceMetricExpr(tmpl string, o monitoring.QueryOptions) string {
	var namespaceSelector string

	// For monitoring namespaces in the specific workspace
	// GET /workspaces/{workspace}/namespaces
	if o.WorkspaceName != "" {
		namespaceSelector = fmt.Sprintf(`workspace="%s", namespace=~"%s"`, o.WorkspaceName, o.ResourceFilter)
		return strings.Replace(tmpl, "$1", namespaceSelector, -1)
	}

	// For monitoring the specific namespaces
	// GET /namespaces/{namespace} or
	// GET /namespaces
	if o.NamespaceName != "" {
		namespaceSelector = fmt.Sprintf(`namespace="%s"`, o.NamespaceName)
	} else {
		namespaceSelector = fmt.Sprintf(`namespace=~"%s"`, o.ResourceFilter)
	}
	return strings.Replace(tmpl, "$1", namespaceSelector, -1)
}

func makeWorkloadMetricExpr(metric, tmpl string, o monitoring.QueryOptions) string {
	var kindSelector, workloadSelector string

	switch o.WorkloadKind {
	case "deployment":
		o.WorkloadKind = Deployment
	case "statefulset":
		o.WorkloadKind = StatefulSet
	case "daemonset":
		o.WorkloadKind = DaemonSet
	default:
		o.WorkloadKind = ".*"
	}
	workloadSelector = fmt.Sprintf(`namespace="%s", workload=~"%s:(%s)"`, o.NamespaceName, o.WorkloadKind, o.ResourceFilter)

	if strings.Contains(metric, "deployment") {
		kindSelector = fmt.Sprintf(`namespace="%s", deployment!="", deployment=~"%s"`, o.NamespaceName, o.ResourceFilter)
	}
	if strings.Contains(metric, "statefulset") {
		kindSelector = fmt.Sprintf(`namespace="%s", statefulset!="", statefulset=~"%s"`, o.NamespaceName, o.ResourceFilter)
	}
	if strings.Contains(metric, "daemonset") {
		kindSelector = fmt.Sprintf(`namespace="%s", daemonset!="", daemonset=~"%s"`, o.NamespaceName, o.ResourceFilter)
	}

	return strings.NewReplacer("$1", workloadSelector, "$2", kindSelector).Replace(tmpl)
}

func makePodMetricExpr(tmpl string, o monitoring.QueryOptions) string {
	var podSelector, workloadSelector string

	// For monitoriong pods of the specific workload
	// GET /namespaces/{namespace}/workloads/{kind}/{workload}/pods
	if o.WorkloadName != "" {
		switch o.WorkloadKind {
		case "deployment":
			workloadSelector = fmt.Sprintf(`owner_kind="ReplicaSet", owner_name=~"^%s-[^-]{1,10}$"`, o.WorkloadName)
		case "statefulset":
			workloadSelector = fmt.Sprintf(`owner_kind="StatefulSet", owner_name="%s"`, o.WorkloadName)
		case "daemonset":
			workloadSelector = fmt.Sprintf(`owner_kind="DaemonSet", owner_name="%s"`, o.WorkloadName)
		}
	}

	// For monitoring pods in the specific namespace
	// GET /namespaces/{namespace}/workloads/{kind}/{workload}/pods or
	// GET /namespaces/{namespace}/pods/{pod} or
	// GET /namespaces/{namespace}/pods
	if o.NamespaceName != "" {
		if o.PodName != "" {
			podSelector = fmt.Sprintf(`pod="%s", namespace="%s"`, o.PodName, o.NamespaceName)
		} else {
			podSelector = fmt.Sprintf(`pod=~"%s", namespace="%s"`, o.ResourceFilter, o.NamespaceName)
		}
	}

	// For monitoring pods on the specific node
	// GET /nodes/{node}/pods/{pod}
	if o.NodeName != "" {
		if o.PodName != "" {
			podSelector = fmt.Sprintf(`pod="%s", node="%s"`, o.PodName, o.NodeName)
		} else {
			podSelector = fmt.Sprintf(`pod=~"%s", node="%s"`, o.ResourceFilter, o.NodeName)
		}
	}
	return strings.NewReplacer("$1", workloadSelector, "$2", podSelector).Replace(tmpl)
}

func makeContainerMetricExpr(tmpl string, o monitoring.QueryOptions) string {
	var containerSelector string
	if o.ContainerName != "" {
		containerSelector = fmt.Sprintf(`pod="%s", namespace="%s", container="%s"`, o.PodName, o.NamespaceName, o.ContainerName)
	} else {
		containerSelector = fmt.Sprintf(`pod="%s", namespace="%s", container=~"%s"`, o.PodName, o.NamespaceName, o.ResourceFilter)
	}
	return strings.Replace(tmpl, "$1", containerSelector, -1)
}

func makePVCMetricExpr(tmpl string, o monitoring.QueryOptions) string {
	var pvcSelector string

	// For monitoring persistentvolumeclaims in the specific namespace
	// GET /namespaces/{namespace}/persistentvolumeclaims/{persistentvolumeclaim} or
	// GET /namespaces/{namespace}/persistentvolumeclaims
	if o.NamespaceName != "" {
		if o.PersistentVolumeClaimName != "" {
			pvcSelector = fmt.Sprintf(`namespace="%s", persistentvolumeclaim="%s"`, o.NamespaceName, o.PersistentVolumeClaimName)
		} else {
			pvcSelector = fmt.Sprintf(`namespace="%s", persistentvolumeclaim=~"%s"`, o.NamespaceName, o.ResourceFilter)
		}
		return strings.Replace(tmpl, "$1", pvcSelector, -1)
	}

	// For monitoring persistentvolumeclaims of the specific storageclass
	// GET /storageclasses/{storageclass}/persistentvolumeclaims
	if o.StorageClassName != "" {
		pvcSelector = fmt.Sprintf(`storageclass="%s", persistentvolumeclaim=~"%s"`, o.StorageClassName, o.ResourceFilter)
	}
	return strings.Replace(tmpl, "$1", pvcSelector, -1)
}
