package prometheus

import (
	"github.com/google/go-cmp/cmp"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus/testdata"
	"testing"
)

func TestMakeExpr(t *testing.T) {
	tests := []struct {
		name string
		opt  monitoring.QueryOptions
	}{
		{"cluster_cpu_utilisation", monitoring.QueryOptions{Level: monitoring.LevelCluster}},
		{"node_cpu_utilisation", monitoring.QueryOptions{Level: monitoring.LevelNode, NodeName: "i-2dazc1d6"}},
		{"node_cpu_total", monitoring.QueryOptions{Level: monitoring.LevelNode, ResourceFilter: "i-2dazc1d6|i-ezjb7gsk"}},
		{"workspace_cpu_usage", monitoring.QueryOptions{Level: monitoring.LevelWorkspace, WorkspaceName: "system-workspace"}},
		{"workspace_memory_usage", monitoring.QueryOptions{Level: monitoring.LevelWorkspace, ResourceFilter: "system-workspace|demo"}},
		{"namespace_cpu_usage", monitoring.QueryOptions{Level: monitoring.LevelNamespace, NamespaceName: "kube-system"}},
		{"namespace_memory_usage", monitoring.QueryOptions{Level: monitoring.LevelNamespace, ResourceFilter: "kube-system|default"}},
		{"namespace_memory_usage_wo_cache", monitoring.QueryOptions{Level: monitoring.LevelNamespace, WorkspaceName: "system-workspace", ResourceFilter: "kube-system|default"}},
		{"workload_cpu_usage", monitoring.QueryOptions{Level: monitoring.LevelWorkload, WorkloadKind: "deployment", NamespaceName: "default", ResourceFilter: "apiserver|coredns"}},
		{"workload_deployment_replica_available", monitoring.QueryOptions{Level: monitoring.LevelWorkload, WorkloadKind: ".*", NamespaceName: "default", ResourceFilter: "apiserver|coredns"}},
		{"pod_cpu_usage", monitoring.QueryOptions{Level: monitoring.LevelPod, NamespaceName: "default", WorkloadKind: "deployment", WorkloadName: "elasticsearch", ResourceFilter: "elasticsearch-0"}},
		{"pod_memory_usage", monitoring.QueryOptions{Level: monitoring.LevelPod, NamespaceName: "default", PodName: "elasticsearch-12345"}},
		{"pod_memory_usage_wo_cache", monitoring.QueryOptions{Level: monitoring.LevelPod, NodeName: "i-2dazc1d6", PodName: "elasticsearch-12345"}},
		{"container_cpu_usage", monitoring.QueryOptions{Level: monitoring.LevelContainer, NamespaceName: "default", PodName: "elasticsearch-12345", ContainerName: "syscall"}},
		{"container_memory_usage", monitoring.QueryOptions{Level: monitoring.LevelContainer, NamespaceName: "default", PodName: "elasticsearch-12345", ResourceFilter: "syscall"}},
		{"pvc_inodes_available", monitoring.QueryOptions{Level: monitoring.LevelPVC, NamespaceName: "default", PersistentVolumeClaimName: "db-123"}},
		{"pvc_inodes_used", monitoring.QueryOptions{Level: monitoring.LevelPVC, NamespaceName: "default", ResourceFilter: "db-123"}},
		{"pvc_inodes_total", monitoring.QueryOptions{Level: monitoring.LevelPVC, StorageClassName: "default", ResourceFilter: "db-123"}},
		{"etcd_server_list", monitoring.QueryOptions{Level: monitoring.LevelComponent}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := testdata.PromQLs[tt.name]
			result := makeExpr(tt.name, tt.opt)
			if diff := cmp.Diff(result, expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}
