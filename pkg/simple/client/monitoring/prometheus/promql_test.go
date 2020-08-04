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
		opts monitoring.QueryOptions
	}{
		{
			name: "cluster_cpu_utilisation",
			opts: monitoring.QueryOptions{
				Level: monitoring.LevelCluster,
			},
		},
		{
			name: "node_cpu_utilisation",
			opts: monitoring.QueryOptions{
				Level:    monitoring.LevelNode,
				NodeName: "i-2dazc1d6",
			},
		},
		{
			name: "node_pod_quota",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelNode,
				ResourceFilter: "i-2dazc1d6|i-ezjb7gsk",
			},
		},
		{
			name: "node_cpu_total",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelNode,
				ResourceFilter: "i-2dazc1d6|i-ezjb7gsk",
			},
		},
		{
			name: "workspace_cpu_usage",
			opts: monitoring.QueryOptions{
				Level:         monitoring.LevelWorkspace,
				WorkspaceName: "system-workspace",
			},
		},
		{
			name: "workspace_memory_usage",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelWorkspace,
				ResourceFilter: "system-workspace|demo",
			},
		},
		{
			name: "namespace_cpu_usage",
			opts: monitoring.QueryOptions{
				Level:         monitoring.LevelNamespace,
				NamespaceName: "kube-system",
			},
		},
		{
			name: "namespace_memory_usage",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelNamespace,
				ResourceFilter: "kube-system|default",
			},
		},
		{
			name: "namespace_memory_usage_wo_cache",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelNamespace,
				WorkspaceName:  "system-workspace",
				ResourceFilter: "kube-system|default",
			},
		},
		{
			name: "workload_cpu_usage",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelWorkload,
				WorkloadKind:   "deployment",
				NamespaceName:  "default",
				ResourceFilter: "apiserver|coredns",
			},
		},
		{
			name: "workload_deployment_replica_available",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelWorkload,
				WorkloadKind:   ".*",
				NamespaceName:  "default",
				ResourceFilter: "apiserver|coredns",
			},
		},
		{
			name: "pod_cpu_usage",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelPod,
				NamespaceName:  "default",
				WorkloadKind:   "deployment",
				WorkloadName:   "elasticsearch",
				ResourceFilter: "elasticsearch-0",
			},
		},
		{
			name: "pod_memory_usage",
			opts: monitoring.QueryOptions{
				Level:         monitoring.LevelPod,
				NamespaceName: "default",
				PodName:       "elasticsearch-12345",
			},
		},
		{
			name: "pod_memory_usage_wo_cache",
			opts: monitoring.QueryOptions{
				Level:    monitoring.LevelPod,
				NodeName: "i-2dazc1d6",
				PodName:  "elasticsearch-12345",
			},
		},
		{
			name: "container_cpu_usage",
			opts: monitoring.QueryOptions{
				Level:         monitoring.LevelContainer,
				NamespaceName: "default",
				PodName:       "elasticsearch-12345",
				ContainerName: "syscall",
			},
		},
		{
			name: "container_memory_usage",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelContainer,
				NamespaceName:  "default",
				PodName:        "elasticsearch-12345",
				ResourceFilter: "syscall",
			},
		},
		{
			name: "pvc_inodes_available",
			opts: monitoring.QueryOptions{
				Level:                     monitoring.LevelPVC,
				NamespaceName:             "default",
				PersistentVolumeClaimName: "db-123",
			},
		},
		{
			name: "pvc_inodes_used",
			opts: monitoring.QueryOptions{
				Level:          monitoring.LevelPVC,
				NamespaceName:  "default",
				ResourceFilter: "db-123",
			},
		},
		{
			name: "pvc_inodes_total",
			opts: monitoring.QueryOptions{
				Level:            monitoring.LevelPVC,
				StorageClassName: "default",
				ResourceFilter:   "db-123",
			},
		},
		{
			name: "etcd_server_list",
			opts: monitoring.QueryOptions{
				Level: monitoring.LevelComponent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := testdata.PromQLs[tt.name]
			result := makeExpr(tt.name, tt.opts)
			if diff := cmp.Diff(result, expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}
