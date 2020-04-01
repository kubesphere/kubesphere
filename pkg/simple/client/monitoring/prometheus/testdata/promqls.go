package testdata

var PromQLs = map[string]string{
	"cluster_cpu_utilisation":               `:node_cpu_utilisation:avg1m`,
	"node_cpu_utilisation":                  `node:node_cpu_utilisation:avg1m{node="i-2dazc1d6"}`,
	"node_cpu_total":                        `node:node_num_cpu:sum{node=~"i-2dazc1d6|i-ezjb7gsk"}`,
	"workspace_cpu_usage":                   `round(sum by (label_kubesphere_io_workspace) (namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", label_kubesphere_io_workspace="system-workspace"}), 0.001)`,
	"workspace_memory_usage":                `sum by (label_kubesphere_io_workspace) (namespace:container_memory_usage_bytes:sum{namespace!="", label_kubesphere_io_workspace=~"system-workspace|demo", label_kubesphere_io_workspace!=""})`,
	"namespace_cpu_usage":                   `round(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", namespace="kube-system"}, 0.001)`,
	"namespace_memory_usage":                `namespace:container_memory_usage_bytes:sum{namespace!="", namespace=~"kube-system|default"}`,
	"namespace_memory_usage_wo_cache":       `namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", label_kubesphere_io_workspace="system-workspace", namespace=~"kube-system|default"}`,
	"workload_cpu_usage":                    `round(namespace:workload_cpu_usage:sum{namespace="default", workload=~"Deployment:apiserver|coredns"}, 0.001)`,
	"workload_deployment_replica_available": `label_join(sum (label_join(label_replace(kube_deployment_status_replicas_available{namespace="default"}, "owner_kind", "Deployment", "", ""), "workload", "", "deployment")) by (namespace, owner_kind, workload), "workload", ":", "owner_kind", "workload")`,
	"pod_cpu_usage":                         `round(label_join(sum by (namespace, pod_name) (irate(container_cpu_usage_seconds_total{job="kubelet", pod_name!="", image!=""}[5m])), "pod", "", "pod_name") * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{owner_kind="ReplicaSet", owner_name=~"^deployment-[^-]{1,10}$"} * on (namespace, pod) group_left(node) kube_pod_info{pod=~"elasticsearch-0", namespace="default"}, 0.001)`,
	"pod_memory_usage":                      `label_join(sum by (namespace, pod_name) (container_memory_usage_bytes{job="kubelet", pod_name!="", image!=""}), "pod", "", "pod_name") * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{} * on (namespace, pod) group_left(node) kube_pod_info{pod="elasticsearch-12345", namespace="default"}`,
	"pod_memory_usage_wo_cache":             `label_join(sum by (namespace, pod_name) (container_memory_working_set_bytes{job="kubelet", pod_name!="", image!=""}), "pod", "", "pod_name") * on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{} * on (namespace, pod) group_left(node) kube_pod_info{pod="elasticsearch-12345", node="i-2dazc1d6"}`,
	"container_cpu_usage":                   `round(sum by (namespace, pod_name, container_name) (irate(container_cpu_usage_seconds_total{job="kubelet", container_name!="POD", container_name!="", image!="", pod_name="elasticsearch-12345", namespace="default", container_name="syscall"}[5m])), 0.001)`,
	"container_memory_usage":                `sum by (namespace, pod_name, container_name) (container_memory_usage_bytes{job="kubelet", container_name!="POD", container_name!="", image!="", pod_name="elasticsearch-12345", namespace="default", container_name=~"syscall"})`,
	"pvc_inodes_available":                  `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes_free) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{namespace="default", persistentvolumeclaim="db-123"}`,
	"pvc_inodes_used":                       `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes_used) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{namespace="default", persistentvolumeclaim=~"db-123"}`,
	"pvc_inodes_total":                      `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_inodes) * on (namespace, persistentvolumeclaim) group_left (storageclass) kube_persistentvolumeclaim_info{storageclass="default", persistentvolumeclaim=~"db-123"}`,
	"etcd_server_list":                      `label_replace(up{job="etcd"}, "node_ip", "$1", "instance", "(.*):.*")`,
}
