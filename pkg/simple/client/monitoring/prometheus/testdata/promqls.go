/*
Copyright 2020 KubeSphere Authors

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

package testdata

var PromQLs = map[string]string{
	"cluster_cpu_utilisation":               `:node_cpu_utilisation:avg1m`,
	"node_cpu_utilisation":                  `node:node_cpu_utilisation:avg1m{node="i-2dazc1d6"}`,
	"node_cpu_total":                        `node:node_num_cpu:sum{node=~"i-2dazc1d6|i-ezjb7gsk"}`,
	"node_pod_quota":                        `max(kube_node_status_capacity{resource="pods",node=~"i-2dazc1d6|i-ezjb7gsk"}) by (node) unless on (node) (kube_node_status_condition{condition="Ready",status=~"unknown|false"} > 0)`,
	"workspace_cpu_usage":                   `round(sum by (workspace) (namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", workspace="system-workspace"}), 0.001)`,
	"workspace_memory_usage":                `sum by (workspace) (namespace:container_memory_usage_bytes:sum{namespace!="", workspace=~"system-workspace|demo", workspace!=""})`,
	"namespace_cpu_usage":                   `round(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", namespace="kube-system"}, 0.001)`,
	"namespace_memory_usage":                `namespace:container_memory_usage_bytes:sum{namespace!="", namespace=~"kube-system|default"}`,
	"namespace_memory_usage_wo_cache":       `namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", workspace="system-workspace", namespace=~"kube-system|default"}`,
	"workload_cpu_usage":                    `round(namespace:workload_cpu_usage:sum{namespace="default", workload=~"Deployment:(apiserver|coredns)"}, 0.001)`,
	"workload_deployment_replica_available": `label_join(sum (label_join(label_replace(kube_deployment_status_replicas_available{namespace="default", deployment!="", deployment=~"apiserver|coredns"}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"pod_cpu_usage":                         `round(sum by (namespace, pod) (irate(container_cpu_usage_seconds_total{job="kubelet", pod!="", image!=""}[5m])) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{owner_kind="ReplicaSet", owner_name=~"^elasticsearch-[^-]{1,10}$"} * on (namespace, pod) group_left(node) kube_pod_info{pod=~"elasticsearch-0", namespace="default"}, 0.001)`,
	"pod_memory_usage":                      `sum by (namespace, pod) (container_memory_usage_bytes{job="kubelet", pod!="", image!=""}) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{} * on (namespace, pod) group_left(node) kube_pod_info{pod="elasticsearch-12345", namespace="default"}`,
	"pod_memory_usage_wo_cache":             `sum by (namespace, pod) (container_memory_working_set_bytes{job="kubelet", pod!="", image!=""}) * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{} * on (namespace, pod) group_left(node) kube_pod_info{pod="elasticsearch-12345", node="i-2dazc1d6"}`,
	"container_cpu_usage":                   `round(sum by (namespace, pod, container) (irate(container_cpu_usage_seconds_total{job="kubelet", container!="POD", container!="", image!="", pod="elasticsearch-12345", namespace="default", container="syscall"}[5m])), 0.001)`,
	"container_memory_usage":                `sum by (namespace, pod, container) (container_memory_usage_bytes{job="kubelet", container!="POD", container!="", image!="", pod="elasticsearch-12345", namespace="default", container=~"syscall"})`,
	"pvc_inodes_available":                  `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes_free) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{namespace="default", persistentvolumeclaim="db-123"}`,
	"pvc_inodes_used":                       `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes_used) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{namespace="default", persistentvolumeclaim=~"db-123"}`,
	"pvc_inodes_total":                      `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{storageclass="default", persistentvolumeclaim=~"db-123"}`,
	"etcd_server_list":                      `label_replace(up{job="etcd"}, "node_ip", "$1", "instance", "(.*):.*")`,
}
