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
	"strconv"
	"strings"

	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

var promQLMeterTemplates = map[string]string{
	// cluster
	"meter_cluster_cpu_usage": `
round(
	(
		sum(
			avg_over_time(kube_pod_container_resource_requests{resource="cpu",unit="core"}[$step])
		) >=
		(
			avg_over_time(:node_cpu_utilisation:avg1m[$step]) * 
			sum(
				avg_over_time(node:node_num_cpu:sum[$step])
			)
		)
	)
	or
	(
		(
			avg_over_time(:node_cpu_utilisation:avg1m[$step]) * 
			sum(
				avg_over_time(node:node_num_cpu:sum[$step])
			)
		) >
		sum(
			avg_over_time(kube_pod_container_resource_requests{resource="cpu",unit="core"}[$step])
		)
	),
	0.001
)`,

	"meter_cluster_memory_usage": `
round(
	(
		sum(
			avg_over_time(kube_pod_container_resource_requests{resource="memory",unit="byte"}[$step])
		) >=
		(
			avg_over_time(:node_memory_utilisation:[$step]) * 
			sum(
				avg_over_time(node:node_memory_bytes_total:sum[$step])
			)
		)
	)
	or
	(
		(
			avg_over_time(:node_memory_utilisation:[$step]) * 
			sum(
				avg_over_time(node:node_memory_bytes_total:sum[$step])
			)
		) >
		sum(
			avg_over_time(kube_pod_container_resource_requests{resource="memory",unit="byte"}[$step])
		)
	),
	1
)`,

	// avg over time manually
	"meter_cluster_net_bytes_transmitted": `
round(
	sum(
		increase(
			node_network_transmit_bytes_total{
				job="node-exporter",
				device!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)"
			}[$step]
		)
	) / $factor,
	1
)`,

	// avg over time manually
	"meter_cluster_net_bytes_received": `
round(
	sum(
		increase(
			node_network_receive_bytes_total{
				job="node-exporter",
				device!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)"
			}[$step]
		)
	) / $factor,
	1
)`,

	"meter_cluster_pvc_bytes_total": `
sum(
	topk(1, avg_over_time(namespace:pvc_bytes_total:sum{}[$step])) by (persistentvolumeclaim)
)`,

	// node
	"meter_node_cpu_usage": `
round(
	(
		sum(
			avg_over_time(kube_pod_container_resource_requests{$nodeSelector, resource="cpu",unit="core"}[$step])
		) by (node) >=
		sum(
			avg_over_time(node:node_cpu_utilisation:avg1m{$nodeSelector}[$step]) * 
			avg_over_time(node:node_num_cpu:sum{$nodeSelector}[$step])
		) by (node)
	)
	or
	(
		sum(
			avg_over_time(node:node_cpu_utilisation:avg1m{$nodeSelector}[$step]) * 
			avg_over_time(node:node_num_cpu:sum{$nodeSelector}[$step])
		) by (node) >
		sum(
			avg_over_time(kube_pod_container_resource_requests{$nodeSelector, resource="cpu",unit="core"}[$step])
		) by (node)
	)
	or
	(
		sum(
			avg_over_time(node:node_cpu_utilisation:avg1m{$nodeSelector}[$step]) * 
			avg_over_time(node:node_num_cpu:sum{$nodeSelector}[$step])
		) by (node)
	),
	0.001
)`,

	"meter_node_memory_usage_wo_cache": `
round(
	(
		sum(
			avg_over_time(kube_pod_container_resource_requests{$nodeSelector, resource="memory",unit="byte"}[$step])
		) by (node) >=
		sum(
			avg_over_time(node:node_memory_bytes_total:sum{$nodeSelector}[$step]) -
			avg_over_time(node:node_memory_bytes_available:sum{$nodeSelector}[$step])
		) by (node)
	)
	or
	(
		sum(
			avg_over_time(node:node_memory_bytes_total:sum{$nodeSelector}[$step]) -
			avg_over_time(node:node_memory_bytes_available:sum{$nodeSelector}[$step])
		) by (node) >
		sum(
			avg_over_time(kube_pod_container_resource_requests{$nodeSelector, resource="memory",unit="byte"}[$step])
		) by (node)
	)
	or
	(
		sum(
			avg_over_time(node:node_memory_bytes_total:sum{$nodeSelector}[$step]) -
			avg_over_time(node:node_memory_bytes_available:sum{$nodeSelector}[$step])
		) by (node)
	),
	0.001
)`,

	// avg over time manually
	"meter_node_net_bytes_transmitted": `
round(
	sum by (node) (
		sum without (instance) (
			label_replace(
				increase(
					node_network_transmit_bytes_total{
						job="node-exporter",
						device!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)",
						$instanceSelector
					}[$step]
				),
				"node",
				"$1",
				"instance",
				"(.*)"
			)
		)
	) / $factor,
	1
)`,

	// avg over time manually
	"meter_node_net_bytes_received": `
round(
	sum by (node) (
		sum without (instance) (
			label_replace(
				increase(
					node_network_receive_bytes_total{
						job="node-exporter",
						device!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)",
						$instanceSelector
					}[$step]
				),
				"node", "
				$1",
				"instance",
				"(.*)"
			)
		)
	) / $factor,
	1
)`,

	"meter_node_pvc_bytes_total": `
sum(
	topk(
		1,
		avg_over_time(
			namespace:pvc_bytes_total:sum{$nodeSelector}[$step]
		)
	) by (persistentvolumeclaim, node)
) by (node)`,

	// workspace
	"meter_workspace_cpu_usage": `
round(
	(
		sum by (workspace) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job",
					namespace!="",
					resource="cpu",
					$1
				}[$step]
			)
		) >=
		sum by (workspace) (
			avg_over_time(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", $1}[$step])
		)
	)
	or
	(
		sum by (workspace) (
			avg_over_time(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", $1}[$step])
		) >
		sum by (workspace) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job",
					namespace!="",
					resource="cpu",
					$1
				}[$step]
			)
		)
	)
	or
	(
		sum by (workspace) (
			avg_over_time(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", $1}[$step])
		)
	),
	0.001
)`,

	"meter_workspace_memory_usage": `
round(
	(
		sum by (workspace) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job",
					namespace!="",
					resource="memory",
					$1
				}[$step]
			)
		) >=
		sum by (workspace) (
			avg_over_time(namespace:container_memory_usage_bytes:sum{namespace!="", $1}[$step])
		)
	)
	or
	(
		sum by (workspace) (
			avg_over_time(namespace:container_memory_usage_bytes:sum{namespace!="", $1}[$step])
		) >
		sum by (workspace) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job", namespace!="", resource="memory", $1
				}[$step]
			)
		)
	)
	or
	(
		sum by (workspace) (
			avg_over_time(namespace:container_memory_usage_bytes:sum{namespace!="", $1}[$step])
		)
	),
	1
)`,

	// avg over time manually
	"meter_workspace_net_bytes_transmitted": `
round(
	(
		sum by (workspace) (
			sum by (namespace) (
				increase(
					container_network_transmit_bytes_total{
						namespace!="",
						pod!="",
						interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)",
						job="kubelet"
					}[$step]
				)
			) * on (namespace) group_left(workspace)
			kube_namespace_labels{$1}
		) or on(workspace) max by(workspace) (kube_namespace_labels{$1} * 0)
	) / $factor, 
	1
)`,

	"meter_workspace_net_bytes_received": `
round(
	(
		sum by (workspace) (
			sum by (namespace) (
				increase(
					container_network_receive_bytes_total{
						namespace!="",
						pod!="",
						interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)",
						job="kubelet"
					}[$step]
				)
			) * on (namespace) group_left(workspace)
			kube_namespace_labels{$1}
		) or on(workspace) max by(workspace) (kube_namespace_labels{$1} * 0)
	) / $factor, 
	1
)`,

	"meter_workspace_pvc_bytes_total": `
sum (
	topk(
		1,
		avg_over_time(namespace:pvc_bytes_total:sum{$1}[$step])
	) by (persistentvolumeclaim, workspace)
) by (workspace)`,

	// namespace
	"meter_namespace_cpu_usage": `
round(
	(
		sum by (namespace) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job",
					namespace!="",
					resource="cpu",
					$1
				}[$step]
			)
		) >=
		sum by (namespace) (
			avg_over_time(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", $1}[$step])
		)
	)
	or
	(
		sum by (namespace) (
			avg_over_time(namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", $1}[$step])
		) >
		sum by (namespace) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{owner_kind!="Job", namespace!="", resource="cpu", $1}[$step]
			)
		)
	)
	or
	(
		sum by (namespace) (
			avg_over_time(
				namespace:container_cpu_usage_seconds_total:sum_rate{namespace!="", $1}[$step]
			)
		)
	),
	0.001
)`,

	"meter_namespace_memory_usage_wo_cache": `
round(
	(
		sum by (namespace) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job",
					namespace!="",
					resource="memory",
					$1
				}[$step]
			)
		) >=
		sum by (namespace) (
			avg_over_time(namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", $1}[$step])
		)
	)
	or
	(
		sum by (namespace) (
			avg_over_time(namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", $1}[$step])
		) >
		sum by (namespace) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job", namespace!="", resource="memory", $1
				}[$step]
			)
		)
	)
	or
	(
		sum by (namespace) (
			avg_over_time(
				namespace:container_memory_usage_bytes_wo_cache:sum{namespace!="", $1}[$step]
			)
		)
	),
	1
)`,

	"meter_namespace_net_bytes_transmitted": `
round(
	(
		sum by (namespace) (
			increase(
				container_network_transmit_bytes_total{
					namespace!="",
					pod!="",
					interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)",
					job="kubelet"
				}[$step]
			)
			* on (namespace) group_left(workspace)
			kube_namespace_labels{$1}
		)
		or on(namespace) max by(namespace) (kube_namespace_labels{$1} * 0)
	) / $factor, 
	1
)`,

	"meter_namespace_net_bytes_received": `
round(
	(
		sum by (namespace) (
			increase(
				container_network_receive_bytes_total{
					namespace!="",
					pod!="",
					interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)",
					job="kubelet"
				}[$step]
			)
			* on (namespace) group_left(workspace)
			kube_namespace_labels{$1}
		)
		or on(namespace) max by(namespace) (kube_namespace_labels{$1} * 0)
	) / $factor, 
	1
)`,

	"meter_namespace_pvc_bytes_total": `
sum (
	topk(
		1,
		avg_over_time(namespace:pvc_bytes_total:sum{$1}[$step])
	) by (persistentvolumeclaim, namespace)
) by (namespace)`,

	// application
	"meter_application_cpu_usage": `
round(
	(
		sum by (namespace, application) (
			label_replace(
				avg_over_time(
					namespace:kube_workload_resource_request:sum{workload!~"Job:.+", resource="cpu", $1}[$step]
				),
				"application",
				"$app",
				"",
				""
			)
		) >=
		sum by (namespace, application) (
			label_replace(
				avg_over_time(namespace:workload_cpu_usage:sum{$1}[$step]),
				"application",
				"$app",
				"",
				""
			)
		)
	)
	or
	(
		sum by (namespace, application) (
			label_replace(
				avg_over_time(namespace:workload_cpu_usage:sum{$1}[$step]),
				"application",
				"$app",
				"",
				""
			)
		) >
		sum by (namespace, application) (
			label_replace(
				avg_over_time(
					namespace:kube_workload_resource_request:sum{workload!~"Job:.+", resource="cpu", $1}[$step]
				),
				"application",
				"$app",
				"",
				""
			)
		)
	)
	or
	(
		sum by (namespace, application) (
			label_replace(
				avg_over_time(namespace:workload_cpu_usage:sum{$1}[$step]),
				"application",
				"$app",
				"",
				""
			)
		)
	),
	0.001
)`,

	"meter_application_memory_usage_wo_cache": `
round(
	(
		sum by (namespace, application) (
			label_replace(
				avg_over_time(
					namespace:kube_workload_resource_request:sum{workload!~"Job:.+", resource="memory", $1}[$step]
				),
				"application",
				"$app",
				"",
				""
			)
		) >=
		sum by (namespace, application) (
			label_replace(
				avg_over_time(namespace:workload_memory_usage_wo_cache:sum{$1}[$step]),
				"application",
				"$app",
				"",
				""
			)
		)
	)
	or
	(
		sum by (namespace, application) (
			label_replace(
				avg_over_time(namespace:workload_memory_usage_wo_cache:sum{$1}[$step]),
				"application",
				"$app",
				"",
				""
			)
		) >
		sum by (namespace, application) (
			label_replace(
				avg_over_time(
					namespace:kube_workload_resource_request:sum{workload!~"Job:.+", resource="memory", $1}[$step]
				),
				"application",
				"$app",
				"",
				""
			)
		)
	)
	or
	(
		sum by (namespace, application) (
			label_replace(
				avg_over_time(namespace:workload_memory_usage_wo_cache:sum{$1}[$step]),
				"application",
				"$app",
				"",
				""
			)
		)
	),
	1
)`,

	"meter_application_net_bytes_transmitted": `
round(
	sum by (namespace, application) (
		label_replace(
			increase(
				namespace:workload_net_bytes_transmitted:sum{$1}[$step]
			),
			"application",
			"$app",
			"",
			""
		)
	) / $factor,
	1
)`,

	"meter_application_net_bytes_received": `
round(
	sum by (namespace, application) (
		label_replace(
			increase(
				namespace:workload_net_bytes_received:sum{$1}[$step]
			),
			"application",
			"$app",
			"",
			""
		)
	) / $factor,
	1
)`,

	"meter_application_pvc_bytes_total": `
sum by (namespace, application) (
	label_replace(
		topk(1, avg_over_time(namespace:pvc_bytes_total:sum{$1}[$step])) by (persistentvolumeclaim),
		"application",
		"$app",
		"",
		""
	)
)`,

	// workload
	"meter_workload_cpu_usage": `
round(
	(
		sum by (namespace, workload) (
			avg_over_time(
				namespace:kube_workload_resource_request:sum{
					workload!~"Job:.+", resource="cpu", $1
				}[$step]
			)
		) >=
		sum by (namespace, workload) (
			avg_over_time(namespace:workload_cpu_usage:sum{$1}[$step])
		)
	)
	or
	(
		sum by (namespace, workload) (
			avg_over_time(namespace:workload_cpu_usage:sum{$1}[$step])
		) >
		sum by (namespace, workload) (
			avg_over_time(
				namespace:kube_workload_resource_request:sum{
					workload!~"Job:.+", resource="cpu", $1
				}[$step]
			)
		)
	)
	or
	(
		sum by (namespace, workload) (
			avg_over_time(namespace:workload_cpu_usage:sum{$1}[$step])
		)
	),
	0.001
)`,

	"meter_workload_memory_usage_wo_cache": `
round(
	(
		sum by (namespace, workload) (
			avg_over_time(
				namespace:kube_workload_resource_request:sum{
					workload!~"Job:.+", resource="memory", $1
				}[$step]
			)
		) >=
		sum by (namespace, workload) (
			avg_over_time(namespace:workload_memory_usage_wo_cache:sum{$1}[$step])
		)
	)
	or
	(
		sum by (namespace, workload) (
			avg_over_time(namespace:workload_memory_usage_wo_cache:sum{$1}[$step])
		) >
		sum by (namespace, workload) (
			avg_over_time(
				namespace:kube_workload_resource_request:sum{
					workload!~"Job:.+", resource="memory", $1
				}[$step]
			)
		)
	)
	or
	(
		sum by (namespace, workload) (
			avg_over_time(namespace:workload_memory_usage_wo_cache:sum{$1}[$step])
		)
	),
	1
)`,

	"meter_workload_net_bytes_transmitted": `
round(
	increase(
		namespace:workload_net_bytes_transmitted:sum{$1}[$step]
	) / $factor,
	1
)`,

	"meter_workload_net_bytes_received": `
round(
	increase(
		namespace:workload_net_bytes_received:sum{$1}[$step]
	) / $factor,
	1
)`,

	"meter_workload_pvc_bytes_total": `
sum by (namespace, workload) (
	topk(
		1,
		avg_over_time(namespace:pvc_bytes_total:sum{$1}[$step])
	) by (persistentvolumeclaim, namespace, workload)
)`,

	// service
	"meter_service_cpu_usage": `
round(
	sum by (namespace, service) (
		label_replace(
			sum by (namespace, pod) (
				avg_over_time(
					namespace:kube_pod_resource_request:sum{owner_kind!="Job", resource="cpu", $1}[$step]
				)
			) >=
			sum by (namespace, pod) (
				sum by (namespace, pod) (
					irate(
						container_cpu_usage_seconds_total{job="kubelet", pod!="", image!=""}[$step]
					)
				) * on (namespace, pod) group_left(owner_kind, owner_name)
				kube_pod_owner{} * on (namespace, pod) group_left(node)
				kube_pod_info{$1}
			),
			"service",
			"$svc",
			"",
			""
		)
	)
	or
	sum by (namespace, service) (
		label_replace(
			sum by (namespace, pod) (
				sum by (namespace, pod) (
					irate(
						container_cpu_usage_seconds_total{job="kubelet", pod!="", image!=""}[$step]
					)
				) * on (namespace, pod) group_left(owner_kind, owner_name)
				kube_pod_owner{} * on (namespace, pod) group_left(node)
				kube_pod_info{$1}
			) >
			sum by (namespace, pod) (
				avg_over_time(namespace:kube_pod_resource_request:sum{owner_kind!="Job", resource="cpu", $1}[$step])
			),
			"service",
			"$svc",
			"",
			""
		)
	)
	or
	sum by (namespace, service) (
		label_replace(
			sum by (namespace, pod) (
				sum by (namespace, pod) (
					irate(
						container_cpu_usage_seconds_total{job="kubelet", pod!="", image!=""}[$step]
					)
				) * on (namespace, pod) group_left(owner_kind, owner_name)
				kube_pod_owner{} * on (namespace, pod) group_left(node)
				kube_pod_info{$1}
			),
			"service",
			"$svc",
			"",
			""
		)
	),
	0.001
)`,

	"meter_service_memory_usage_wo_cache": `
round(
	(
		sum by (namespace, service) (
			label_replace(
				sum by (namespace, pod) (
					avg_over_time(
						namespace:kube_pod_resource_request:sum{owner_kind!="Job", resource="memory", $1}[$step]
					)
				) >=
				sum by (namespace, pod) (
					sum by (namespace, pod) (
						avg_over_time(
							container_memory_working_set_bytes{job="kubelet", pod!="", image!=""}[$step]
						)
					) * on (namespace, pod) group_left(owner_kind, owner_name)
					kube_pod_owner{} * on (namespace, pod) group_left(node)
					kube_pod_info{$1}
				),
				"service",
				"$svc",
				"",
				""
			)
		)
	)
	or
	(
		sum by (namespace, service) (
			label_replace(
				sum by (namespace, pod) (
					sum by (namespace, pod) (
						avg_over_time(
							container_memory_working_set_bytes{job="kubelet", pod!="", image!=""}[$step]
						)
					) * on (namespace, pod) group_left(owner_kind, owner_name)
					kube_pod_owner{} * on (namespace, pod) group_left(node)
					kube_pod_info{$1}
				) >
				sum by (namespace, pod) (
					avg_over_time(
						namespace:kube_pod_resource_request:sum{owner_kind!="Job", resource="memory", $1}[$step]
					)
				),
				"service",
				"$svc",
				"",
				""
			)
		)
	)
	or
	(
		sum by (namespace, service) (
			label_replace(
				sum by (namespace, pod) (
					sum by (namespace, pod) (
						avg_over_time(
							container_memory_working_set_bytes{job="kubelet", pod!="", image!=""}[$step]
						)
					) * on (namespace, pod) group_left(owner_kind, owner_name)
					kube_pod_owner{} * on (namespace, pod) group_left(node)
					kube_pod_info{$1}
				),
				"service",
				"$svc",
				"",
				""
			)
		)
	),
	1
)`,

	"meter_service_net_bytes_transmitted": `
round(
	sum by (namespace, service) (
		label_replace(
			sum by (namespace, pod) (
				sum by (namespace, pod) (
					increase(
						container_network_transmit_bytes_total{
							pod!="",
							interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", 
							job="kubelet"
						}[$step]
					)
				) * on (namespace, pod) group_left(owner_kind, owner_name)
				kube_pod_owner{} * on (namespace, pod) group_left(node)
				kube_pod_info{$1}
			),
			"service",
			"$svc",
			"",
			""
		)
	) / $factor,
	1
)`,

	"meter_service_net_bytes_received": `
round(
	sum by (namespace, service) (
		label_replace(
			sum by (namepace, pod) (
				sum by (namespace, pod) (
					increase(
						container_network_receive_bytes_total{
							pod!="",
							interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)",
							job="kubelet"
						}[$step]
					)
				) * on (namespace, pod) group_left(owner_kind, owner_name)
				kube_pod_owner{} * on (namespace, pod) group_left(node)
				kube_pod_info{$1}
			),
			"service",
			"$svc",
			"",
			""
		)
	) / $factor,
	1
)`,

	// pod
	"meter_pod_cpu_usage": `
round(
	(
		sum by (namespace, pod) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job",
					resource="cpu",
					$internalPodSelector
				}[$step]
			)
		)
		* on (namespace, pod) group_left(owner_kind, owner_name)
		kube_pod_owner{$1}
		* on (namespace, pod) group_left(node)
		kube_pod_info{$2} >=
		sum by (namespace, pod) (
			irate(container_cpu_usage_seconds_total{job="kubelet",pod!="",image!="", $internalPodSelector}[$step])
		)
		* on (namespace, pod) group_left(owner_kind, owner_name)
		kube_pod_owner{$1}
		* on (namespace, pod) group_left(node)
		kube_pod_info{$2}
	)
	or
	(
		sum by (namespace, pod) (
			irate(container_cpu_usage_seconds_total{job="kubelet",pod!="",image!="", $internalPodSelector}[$step])
		)
		* on (namespace, pod) group_left(owner_kind, owner_name)
		kube_pod_owner{$1}
		* on (namespace, pod) group_left(node)
		kube_pod_info{$2} >
		sum by (namespace, pod) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job",
					resource="cpu",
					$internalPodSelector
				}[$step]
			)
		)
		* on (namespace, pod) group_left(owner_kind, owner_name)
		kube_pod_owner{$1}
		* on (namespace, pod) group_left(node)
		kube_pod_info{$2}
	)
	or
	(
		sum by (namespace, pod) (
			irate(container_cpu_usage_seconds_total{job="kubelet",pod!="",image!="", $internalPodSelector}[$step])
		)
		* on (namespace, pod) group_left(owner_kind, owner_name)
		kube_pod_owner{$1}
		* on (namespace, pod) group_left(node)
		kube_pod_info{$2}
	),
	0.001
)`,

	"meter_pod_memory_usage_wo_cache": `
round(
	(
		sum by (namespace, pod) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job",
					resource="memory",
					$internalPodSelector
				}[$step]
			)
		)
		* on (namespace, pod) group_left(owner_kind, owner_name)
		kube_pod_owner{$1}
		* on (namespace, pod) group_left(node)
		kube_pod_info{$2} >=
		sum by (namespace, pod) (
			avg_over_time(container_memory_working_set_bytes{job="kubelet", pod!="", image!="", $internalPodSelector}[$step])
		)
		* on (namespace, pod) group_left(owner_kind, owner_name)
		kube_pod_owner{$1}
		* on (namespace, pod) group_left(node)
		kube_pod_info{$2}
	)
	or
	(
		sum by (namespace, pod) (
			avg_over_time(container_memory_working_set_bytes{job="kubelet", pod!="", image!="", $internalPodSelector}[$step])
		)
		* on (namespace, pod) group_left(owner_kind, owner_name)
		kube_pod_owner{$1}
		* on (namespace, pod) group_left(node)
		kube_pod_info{$2} >
		sum by (namespace, pod) (
			avg_over_time(
				namespace:kube_pod_resource_request:sum{
					owner_kind!="Job",
					resource="memory",
					$internalPodSelector
				}[$step]
			)
			* on (namespace, pod) group_left(owner_kind, owner_name)
			kube_pod_owner{$1}
			* on (namespace, pod) group_left(node)
			kube_pod_info{$2}
		)
	)
	or
	(
		sum by (namespace, pod) (
			avg_over_time(container_memory_working_set_bytes{job="kubelet", pod!="", image!="", $internalPodSelector}[$step])
		)
		* on (namespace, pod) group_left(owner_kind, owner_name)
		kube_pod_owner{$1}
		* on (namespace, pod) group_left(node)
		kube_pod_info{$2}
	),
	0.001
)`,

	"meter_pod_net_bytes_transmitted": `
sum by (namespace, pod) (
	increase(
		container_network_transmit_bytes_total{
			pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet", $internalPodSelector
		}[$step]
	) / $factor
)
* on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1}
* on (namespace, pod) group_left(node) kube_pod_info{$2}`,

	"meter_pod_net_bytes_received": `
sum by (namespace, pod) (
	increase(
		container_network_receive_bytes_total{
			pod!="", interface!~"^(cali.+|tunl.+|dummy.+|kube.+|flannel.+|cni.+|docker.+|veth.+|lo.*)", job="kubelet", $internalPodSelector
		}[$step]
	) / $factor
)
* on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1}
* on (namespace, pod) group_left(node) kube_pod_info{$2}`,

	"meter_pod_pvc_bytes_total": `
sum by (namespace, pod) (
	avg_over_time(namespace:pvc_bytes_total:sum{$internalPodSelector}[$step])
)
* on (namespace, pod) group_left(owner_kind, owner_name) kube_pod_owner{$1}
* on (namespace, pod) group_left(node) kube_pod_info{$2}`,
}

func makeMeterExpr(meter string, o monitoring.QueryOptions) string {

	var tmpl string
	if tmpl = getMeterTemplate(meter); len(tmpl) == 0 {
		klog.Errorf("invalid meter %s", meter)
		return ""
	}
	tmpl = renderMeterTemplate(tmpl, o)

	switch o.Level {
	case monitoring.LevelCluster:
		return makeClusterMeterExpr(tmpl, o)
	case monitoring.LevelNode:
		return makeNodeMeterExpr(tmpl, o)
	case monitoring.LevelWorkspace:
		return makeWorkspaceMeterExpr(tmpl, o)
	case monitoring.LevelNamespace:
		return makeNamespaceMeterExpr(tmpl, o)
	case monitoring.LevelApplication:
		return makeApplicationMeterExpr(tmpl, o)
	case monitoring.LevelWorkload:
		return makeWorkloadMeterExpr(meter, tmpl, o)
	case monitoring.LevelService:
		return makeServiceMeterExpr(tmpl, o)
	case monitoring.LevelPod:
		return makePodMeterExpr(tmpl, o)
	default:
		return ""
	}

}

func getMeterTemplate(meter string) string {
	if tmpl, ok := promQLMeterTemplates[meter]; !ok {
		klog.Errorf("invalid meter %s", meter)
		return ""
	} else {
		return strings.Join(strings.Fields(strings.TrimSpace(tmpl)), " ")
	}
}

func renderMeterTemplate(tmpl string, o monitoring.QueryOptions) string {
	if o.MeterOptions == nil {
		klog.Error("meter options not found")
		return ""
	}

	tmpl = replaceStepSelector(tmpl, o)
	tmpl = replacePVCSelector(tmpl, o)
	tmpl = replaceNodeSelector(tmpl, o)
	tmpl = replaceInstanceSelector(tmpl, o)
	tmpl = replaceAppSelector(tmpl, o)
	tmpl = replaceSvcSelector(tmpl, o)
	tmpl = replaceFactor(tmpl, o)

	return tmpl
}

func makeClusterMeterExpr(tmpl string, o monitoring.QueryOptions) string {
	return tmpl
}

func makeNodeMeterExpr(tmpl string, o monitoring.QueryOptions) string {
	return tmpl
}

func makeWorkspaceMeterExpr(tmpl string, o monitoring.QueryOptions) string {
	return makeWorkspaceMetricExpr(tmpl, o)
}

func makeNamespaceMeterExpr(tmpl string, o monitoring.QueryOptions) string {
	return makeNamespaceMetricExpr(tmpl, o)
}

func makeApplicationMeterExpr(tmpl string, o monitoring.QueryOptions) string {
	return strings.NewReplacer("$1", o.ResourceFilter).Replace(tmpl)
}

func makeWorkloadMeterExpr(meter string, tmpl string, o monitoring.QueryOptions) string {
	return makeWorkloadMetricExpr(meter, tmpl, o)
}

func makeServiceMeterExpr(tmpl string, o monitoring.QueryOptions) string {
	return strings.Replace(tmpl, "$1", o.ResourceFilter, -1)
}

func makePodMeterExpr(tmpl string, o monitoring.QueryOptions) string {

	// here we support internal pod selector to accelerate metering pod filter operation, otherwise we will iterate
	// pod in cluster scope which required longer time
	var internalPodSelector string
	if o.PodName != "" {
		internalPodSelector += fmt.Sprintf(`pod="%s", `, o.PodName)
	}
	if o.ResourceFilter != "" {
		internalPodSelector += fmt.Sprintf(`pod=~"%s", `, o.ResourceFilter)
	}
	if o.NamespaceName != "" {
		internalPodSelector += fmt.Sprintf(`namespace="%s"`, o.NamespaceName)
	}

	tmpl = strings.NewReplacer("$internalPodSelector", internalPodSelector).Replace(tmpl)

	return makePodMetricExpr(tmpl, o)
}

func replacePVCSelector(tmpl string, o monitoring.QueryOptions) string {
	var filterConditions []string

	switch o.Level {
	case monitoring.LevelCluster:
		break
	case monitoring.LevelNode:
		if o.NodeName != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`node="%s"`, o.NodeName))
		} else if o.ResourceFilter != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`node=~"%s"`, o.ResourceFilter))
		}
		if o.PVCFilter != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`persistentvolumeclaim=~"%s"`, o.PVCFilter))
		}
		if o.StorageClassName != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`storageclass="%s"`, o.StorageClassName))
		}
	case monitoring.LevelWorkspace:
		if o.WorkspaceName != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`workspace="%s"`, o.WorkspaceName))
		}
		if o.PVCFilter != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`persistentvolumeclaim=~"%s"`, o.PVCFilter))
		}
		if o.StorageClassName != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`storageclass="%s"`, o.StorageClassName))
		}
	case monitoring.LevelNamespace:
		if o.NamespaceName != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`namespace="%s"`, o.NamespaceName))
		}
		if o.PVCFilter != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`persistentvolumeclaim=~"%s"`, o.PVCFilter))
		}
		if o.StorageClassName != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`storageclass="%s"`, o.StorageClassName))
		}
	case monitoring.LevelApplication:
		if o.NamespaceName != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`namespace="%s"`, o.NamespaceName))
		}
		// o.PVCFilter is required and is filled with application volumes else just empty string which means that
		// there won't be any match application volumes
		filterConditions = append(filterConditions, fmt.Sprintf(`persistentvolumeclaim=~"%s"`, o.PVCFilter))

		if o.StorageClassName != "" {
			filterConditions = append(filterConditions, fmt.Sprintf(`storageclass="%s"`, o.StorageClassName))
		}
	default:
		return tmpl
	}
	return strings.Replace(tmpl, "$pvc", strings.Join(filterConditions, ","), -1)
}

func replaceFactor(tmpl string, o monitoring.QueryOptions) string {
	stepStr := strconv.Itoa(int(o.MeterOptions.Step.Hours()))

	return strings.Replace(tmpl, "$factor", stepStr, -1)
}

func replaceStepSelector(tmpl string, o monitoring.QueryOptions) string {
	stepStr := strconv.Itoa(int(o.MeterOptions.Step.Hours())) + "h"

	return strings.Replace(tmpl, "$step", stepStr, -1)
}

func replaceNodeSelector(tmpl string, o monitoring.QueryOptions) string {

	var nodeSelector string
	if o.NodeName != "" {
		nodeSelector = fmt.Sprintf(`node="%s"`, o.NodeName)
	} else {
		nodeSelector = fmt.Sprintf(`node=~"%s"`, o.ResourceFilter)
	}
	return strings.Replace(tmpl, "$nodeSelector", nodeSelector, -1)
}

func replaceInstanceSelector(tmpl string, o monitoring.QueryOptions) string {
	var instanceSelector string
	if o.NodeName != "" {
		instanceSelector = fmt.Sprintf(`instance="%s"`, o.NodeName)
	} else {
		instanceSelector = fmt.Sprintf(`instance=~"%s"`, o.ResourceFilter)
	}
	return strings.Replace(tmpl, "$instanceSelector", instanceSelector, -1)
}

func replaceAppSelector(tmpl string, o monitoring.QueryOptions) string {
	return strings.Replace(tmpl, "$app", o.ApplicationName, -1)
}

func replaceSvcSelector(tmpl string, o monitoring.QueryOptions) string {
	return strings.Replace(tmpl, "$svc", o.ServiceName, -1)
}
