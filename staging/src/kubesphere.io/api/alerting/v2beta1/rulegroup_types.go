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

package v2beta1

import (
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/model/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ResourceKindRuleGroup        = "RuleGroup"
	ResourceKindClusterRuleGroup = "ClusterRuleGroup"
	ResourceKindGlobalRuleGroup  = "GlobalRuleGroup"
)

// Duration is a valid time unit
// Supported units: y, w, d, h, m, s, ms Examples: `30s`, `1m`, `1h20m15s`
// +kubebuilder:validation:Pattern:="^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"
type Duration string

type Comparator string

type Severity string

type MatchType string

const (
	ComparatorLT Comparator = "<"
	ComparatorLE Comparator = "<="
	ComparatorGT Comparator = ">"
	ComparatorGE Comparator = ">="

	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"

	MatchEqual     = "="
	MatchNotEqual  = "!="
	MatchRegexp    = "=~"
	MatchNotRegexp = "!~"
)

type Rule struct {
	Alert string `json:"alert"`

	Expr intstr.IntOrString `json:"expr,omitempty"`

	For      Duration `json:"for,omitempty"`
	Severity Severity `json:"severity,omitempty"`

	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	Disable bool `json:"disable,omitempty"`
}

type NamespaceRule struct {
	Rule `json:",inline"`
	// If ExprBuilder is not nil, the configured Expr will be ignored
	ExprBuilder *NamespaceRuleExprBuilder `json:"exprBuilder,omitempty"`
}

type ClusterRule struct {
	Rule `json:",inline"`
	// If ExprBuilder is not nil, the configured Expr will be ignored
	ExprBuilder *ClusterRuleExprBuilder `json:"exprBuilder,omitempty"`
}

type GlobalRule struct {
	ClusterSelector   *MetricLabelSelector `json:"clusterSelector,omitempty"`
	NamespaceSelector *MetricLabelSelector `json:"namespaceSelector,omitempty"`
	Rule              `json:",inline"`
	// If ExprBuilder is not nil, the configured Expr will be ignored
	ExprBuilder *GlobalRuleExprBuilder `json:"exprBuilder,omitempty"`
}

// Only one of its members may be specified.
type MetricLabelSelector struct {
	InValues []string `json:"inValues,omitempty"`
	Matcher  *Matcher `json:"matcher,omitempty"`
}

func (s *MetricLabelSelector) ParseToMatcher(labelName string) *labels.Matcher {
	if s == nil {
		return nil
	}
	if len(s.InValues) == 1 {
		return &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  labelName,
			Value: s.InValues[0],
		}
	}
	if len(s.InValues) > 1 {
		return &labels.Matcher{
			Type:  labels.MatchRegexp,
			Name:  labelName,
			Value: fmt.Sprintf("(%s)", strings.Join(s.InValues, "|")),
		}
	}
	if s.Matcher != nil {
		var mtype labels.MatchType
		switch s.Matcher.Type {
		case MatchEqual:
			mtype = labels.MatchEqual
		case MatchNotEqual:
			mtype = labels.MatchNotEqual
		case MatchRegexp:
			mtype = labels.MatchRegexp
		case MatchNotRegexp:
			mtype = labels.MatchNotRegexp
		default:
			return nil
		}
		return &labels.Matcher{
			Type:  mtype,
			Name:  labelName,
			Value: s.Matcher.Value,
		}
	}
	return nil
}

func (s *MetricLabelSelector) Validate() error {
	if s.Matcher != nil {
		return s.Matcher.Validate()
	}
	return nil
}

type Matcher struct {
	Type  MatchType `json:"type"`
	Value string    `json:"value,omitempty"`
}

func (m *Matcher) Validate() error {
	var mtype labels.MatchType
	switch m.Type {
	case MatchEqual:
		mtype = labels.MatchEqual
	case MatchNotEqual:
		mtype = labels.MatchNotEqual
	case MatchRegexp:
		mtype = labels.MatchRegexp
	case MatchNotRegexp:
		mtype = labels.MatchNotRegexp
	default:
		return fmt.Errorf("unsupported match type [%s]", m.Type)
	}
	_, err := labels.NewMatcher(mtype, "name", m.Value)
	if err != nil {
		return fmt.Errorf("invalid matcher: %v", err)
	}
	return nil
}

type NamespaceRuleExprBuilder struct {
	Workload *WorkloadExprBuilder `json:"workload,omitempty"`
}

type ClusterRuleExprBuilder struct {
	Node *NodeExprBuilder `json:"node,omitempty"`
}

// Only one of its members may be specified.
type GlobalRuleExprBuilder struct {
	Workload *ScopedWorkloadExprBuilder `json:"workload,omitempty"`
	Node     *ScopedNodeExprBuilder     `json:"node,omitempty"`
}

type WorkloadKind string

const (
	WorkloadDeployment  WorkloadKind = "deployment"
	WorkloadStatefulSet WorkloadKind = "statefulset"
	WorkloadDaemonSet   WorkloadKind = "daemonset"
)

type WorkloadExprBuilder struct {
	WorkloadKind  WorkloadKind `json:"kind"`
	WorkloadNames []string     `json:"names"`
	Comparator    Comparator   `json:"comparator"`

	MetricThreshold WorkloadMetricThreshold `json:"metricThreshold,omitempty"`
}

type ScopedWorkloadExprBuilder struct {
	WorkloadKind  WorkloadKind          `json:"kind"`
	WorkloadNames []ScopedWorkloadNames `json:"names"`
	Comparator    Comparator            `json:"comparator"`

	MetricThreshold WorkloadMetricThreshold `json:"metricThreshold,omitempty"`
}

// The cluster and namespace to which the workloads belongs must be specified.
type ScopedWorkloadNames struct {
	Cluster   string   `json:"cluster"`
	Namespace string   `json:"namespace"`
	Names     []string `json:"names"`
}

const (
	MetricWorkloadCpuUsage               = "namespace:workload_cpu_usage:sum"
	MetricWorkloadMemoryUsage            = "namespace:workload_memory_usage:sum"
	MetricWorkloadMemoryUsageWoCache     = "namespace:workload_memory_usage_wo_cache:sum"
	MetricWorkloadNetworkTransmittedRate = "namespace:workload_net_bytes_transmitted:sum_irate"
	MetricWorkloadNetworkReceivedRate    = "namespace:workload_net_bytes_received:sum_irate"
	MetricWorkloadPodUnavailableRatio    = "namespace:%s_unavailable_replicas:ratio" // "%s" must be one of "deployment", "statefulset" and "daemonset"
)

func (b *WorkloadExprBuilder) Build() string {
	if b == nil {
		return ""
	}
	if b.WorkloadKind == "" || len(b.WorkloadNames) == 0 || b.Comparator == "" {
		return ""
	}

	var (
		threshold float64
		metric    string
	)

	switch {
	case b.MetricThreshold.Cpu != nil:
		var cpu = b.MetricThreshold.Cpu
		if cpu.Usage != nil {
			metric = MetricWorkloadCpuUsage
			threshold = *cpu.Usage
		}
	case b.MetricThreshold.Memory != nil:
		var memory = b.MetricThreshold.Memory
		switch {
		case memory.Usage != nil:
			metric = MetricWorkloadMemoryUsage
			threshold = *memory.Usage
		case memory.UsageWoCache != nil:
			metric = MetricWorkloadMemoryUsageWoCache
			threshold = *memory.UsageWoCache
		}
	case b.MetricThreshold.Network != nil:
		var network = b.MetricThreshold.Network
		switch {
		case network.TransmittedRate != nil:
			metric = MetricWorkloadNetworkTransmittedRate
			threshold = *network.TransmittedRate
		case network.ReceivedRate != nil:
			metric = MetricWorkloadNetworkReceivedRate
			threshold = *network.ReceivedRate
		}
	case b.MetricThreshold.Replica != nil:
		var replica = b.MetricThreshold.Replica
		if replica.UnavailableRatio != nil {
			metric = fmt.Sprintf(MetricWorkloadPodUnavailableRatio, strings.ToLower(string(b.WorkloadKind)))
			threshold = *replica.UnavailableRatio
		}
	}

	if metric != "" {
		var filter string
		if len(b.WorkloadNames) == 1 {
			filter = fmt.Sprintf(`{workload="%s:%s"}`, b.WorkloadKind, b.WorkloadNames[0])
		} else {
			filter = fmt.Sprintf(`{workload=~"%s:(%s)"}`, b.WorkloadKind, strings.Join(b.WorkloadNames, "|"))
		}
		return metric + fmt.Sprintf("%s %s %v", filter, b.Comparator, threshold)
	}

	return ""
}

func (b *ScopedWorkloadExprBuilder) Build() string {
	// include the workload names into builded expr only.
	// the limited clusters and namespaces will be set to the clusterSelector and namespaceSelector separately.

	var names = make(map[string]struct{})
	for _, snames := range b.WorkloadNames {
		for _, name := range snames.Names {
			names[name] = struct{}{}
		}
	}

	var eb = WorkloadExprBuilder{
		WorkloadKind:    b.WorkloadKind,
		Comparator:      b.Comparator,
		MetricThreshold: b.MetricThreshold,
	}
	for name := range names {
		eb.WorkloadNames = append(eb.WorkloadNames, name)
	}

	return eb.Build()
}

// Only one of its members may be specified.
type WorkloadMetricThreshold struct {
	Cpu     *WorkloadCpuThreshold     `json:"cpu,omitempty"`
	Memory  *WorkloadMemoryThreshold  `json:"memory,omitempty"`
	Network *WorkloadNetworkThreshold `json:"network,omitempty"`
	Replica *WorkloadReplicaThreshold `json:"replica,omitempty"`
}

// Only one of its members may be specified.
type WorkloadCpuThreshold struct {
	// The unit is core
	Usage *float64 `json:"usage,omitempty"`
}

// Only one of its members may be specified.
type WorkloadMemoryThreshold struct {
	// The memory usage contains cache
	// The unit is bytes
	Usage *float64 `json:"usage,omitempty"`
	// The memory usage contains no cache
	// The unit is bytes
	UsageWoCache *float64 `json:"usageWoCache,omitempty"`
}

// Only one of its members may be specified.
type WorkloadNetworkThreshold struct {
	// The unit is bit/s
	TransmittedRate *float64 `json:"transmittedRate,omitempty"`
	// The unit is bit/s
	ReceivedRate *float64 `json:"receivedRate,omitempty"`
}

// Only one of its members may be specified.
type WorkloadReplicaThreshold struct {
	UnavailableRatio *float64 `json:"unavailableRatio,omitempty"`
}

type NodeExprBuilder struct {
	NodeNames  []string   `json:"names"`
	Comparator Comparator `json:"comparator"`

	MetricThreshold NodeMetricThreshold `json:"metricThreshold"`
}

type ScopedNodeExprBuilder struct {
	NodeNames  []ScopedNodeNames `json:"names"`
	Comparator Comparator        `json:"comparator"`

	MetricThreshold NodeMetricThreshold `json:"metricThreshold,omitempty"`
}

// The cluster to which the node belongs must be specified.
type ScopedNodeNames struct {
	Cluster string   `json:"cluster"`
	Names   []string `json:"names"`
}

const (
	MetricNodeCpuUtilization         = "node:node_cpu_utilisation:avg1m"
	MetricNodeCpuLoad1m              = "node:load1:ratio"
	MetricNodeCpuLoad5m              = "node:load5:ratio"
	MetricNodeCpuLoad15m             = "node:load15:ratio"
	MetricNodeMemoryUtilization      = "node:node_memory_utilisation:"
	MetricNodeMemoryAvailable        = "node:node_memory_bytes_available:sum"
	MetricNodeNetworkTransmittedRate = "node:node_net_bytes_transmitted:sum_irate"
	MetricNodeNetwrokReceivedRate    = "node:node_net_bytes_received:sum_irate"
	MetricNodeDiskSpaceUtilization   = "node:disk_space_utilization:ratio"
	MetricNodeDiskSpaceAvailable     = "node:disk_space_available:"
	MetricNodeDiskInodeUtilization   = "node:disk_inode_utilization:ratio"
	MetricNodeDiskIopsRead           = "node:data_volume_iops_reads:sum"
	MetricNodeDiskIopsWrite          = "node:data_volume_iops_writes:sum"
	MetricNodeDiskThroughputRead     = "node:data_volume_throughput_bytes_read:sum"
	MetricNodeDiskThroughputWrite    = "node:data_volume_throughput_bytes_write:sum"
	MetricNodePodUtilization         = "node:pod_utilization:ratio"
	MetricNodePodAbnormalRatio       = "node:pod_abnormal:ratio"
)

func (b *NodeExprBuilder) Build() string {
	if len(b.NodeNames) == 0 || b.Comparator == "" {
		return ""
	}

	var (
		threshold float64
		metric    string
	)

	switch {
	case b.MetricThreshold.Cpu != nil:
		var cpu = b.MetricThreshold.Cpu
		switch {
		case cpu.Utilization != nil:
			metric = MetricNodeCpuUtilization
			threshold = *cpu.Utilization
		case cpu.Load1m != nil:
			metric = MetricNodeCpuLoad1m
			threshold = *cpu.Load1m
		case cpu.Load5m != nil:
			metric = MetricNodeCpuLoad5m
			threshold = *cpu.Load5m
		case cpu.Load15m != nil:
			metric = MetricNodeCpuLoad15m
			threshold = *cpu.Load15m
		}
	case b.MetricThreshold.Memory != nil:
		var memory = b.MetricThreshold.Memory
		switch {
		case memory.Utilization != nil:
			metric = MetricNodeMemoryUtilization
			threshold = *memory.Utilization
		case memory.Available != nil:
			metric = MetricNodeMemoryAvailable
			threshold = *memory.Available
		}
	case b.MetricThreshold.Network != nil:
		var network = b.MetricThreshold.Network
		switch {
		case network.TransmittedRate != nil:
			metric = MetricNodeNetworkTransmittedRate
			threshold = *network.TransmittedRate
		case network.ReceivedRate != nil:
			metric = MetricNodeNetwrokReceivedRate
			threshold = *network.ReceivedRate
		}
	case b.MetricThreshold.Disk != nil:
		var disk = b.MetricThreshold.Disk
		switch {
		case disk.SpaceUtilization != nil:
			metric = MetricNodeDiskSpaceUtilization
			threshold = *disk.SpaceUtilization
		case disk.SpaceAvailable != nil:
			metric = MetricNodeDiskSpaceAvailable
			threshold = *disk.SpaceAvailable
		case disk.InodeUtilization != nil:
			metric = MetricNodeDiskInodeUtilization
			threshold = *disk.InodeUtilization
		case disk.IopsRead != nil:
			metric = MetricNodeDiskIopsRead
			threshold = *disk.IopsRead
		case disk.IopsWrite != nil:
			metric = MetricNodeDiskIopsWrite
			threshold = *disk.IopsWrite
		case disk.ThroughputRead != nil:
			metric = MetricNodeDiskThroughputRead
			threshold = *disk.ThroughputRead
		case disk.ThroughputWrite != nil:
			metric = MetricNodeDiskThroughputWrite
			threshold = *disk.ThroughputWrite
		}
	case b.MetricThreshold.Pod != nil:
		var pod = b.MetricThreshold.Pod
		switch {
		case pod.Utilization != nil:
			metric = MetricNodePodUtilization
			threshold = *pod.Utilization
		case pod.AbnormalRatio != nil:
			metric = MetricNodePodAbnormalRatio
			threshold = *pod.AbnormalRatio
		}
	}

	if metric != "" {
		var filter string
		if len(b.NodeNames) == 1 {
			filter = fmt.Sprintf(`{node="%s"}`, b.NodeNames[0])
		} else {
			filter = fmt.Sprintf(`{node=~"(%s)"}`, strings.Join(b.NodeNames, "|"))
		}
		return metric + fmt.Sprintf("%s %s %v", filter, b.Comparator, threshold)
	}

	return ""
}

func (b *ScopedNodeExprBuilder) Build() string {
	// include the node names into builded expr only.
	// the limited clusters will be set to the clusterSelector.

	var names = make(map[string]struct{})
	for _, snames := range b.NodeNames {
		for _, name := range snames.Names {
			names[name] = struct{}{}
		}
	}

	var eb = NodeExprBuilder{
		Comparator:      b.Comparator,
		MetricThreshold: b.MetricThreshold,
	}
	for name := range names {
		eb.NodeNames = append(eb.NodeNames, name)
	}

	return eb.Build()
}

// Only one of its members may be specified.
type NodeMetricThreshold struct {
	Cpu     *NodeCpuThreshold     `json:"cpu,omitempty"`
	Memory  *NodeMemoryThreshold  `json:"memory,omitempty"`
	Network *NodeNetworkThreshold `json:"network,omitempty"`
	Disk    *NodeDiskThreshold    `json:"disk,omitempty"`
	Pod     *NodePodThreshold     `json:"pod,omitempty"`
}

// Only one of its members may be specified.
type NodeCpuThreshold struct {
	Utilization *float64 `json:"utilization,omitempty"`
	Load1m      *float64 `json:"load1m,omitempty"`
	Load5m      *float64 `json:"load5m,omitempty"`
	Load15m     *float64 `json:"load15m,omitempty"`
}

// Only one of its members may be specified.
type NodeMemoryThreshold struct {
	Utilization *float64 `json:"utilization,omitempty"`
	// The unit is bytes
	Available *float64 `json:"available,omitempty"`
}

// Only one of its members may be specified.
type NodePodThreshold struct {
	Utilization   *float64 `json:"utilization,omitempty"`
	AbnormalRatio *float64 `json:"abnormalRatio,omitempty"`
}

// Only one of its members may be specified.
type NodeDiskThreshold struct {
	SpaceUtilization *float64 `json:"spaceUtilization,omitempty"`
	// The unit is bytes
	SpaceAvailable   *float64 `json:"spaceAvailable,omitempty"`
	InodeUtilization *float64 `json:"inodeUtilization,omitempty"`
	// The unit is io/s
	IopsRead *float64 `json:"iopsRead,omitempty"`
	// The unit is io/s
	IopsWrite *float64 `json:"iopsWrite,omitempty"`
	// The unit is bytes/s
	ThroughputRead *float64 `json:"throughputRead,omitempty"`
	// The unit is bytes/s
	ThroughputWrite *float64 `json:"throughputWrite,omitempty"`
}

// Only one of its members may be specified.
type NodeNetworkThreshold struct {
	// The unit is bit/s
	TransmittedRate *float64 `json:"transmittedRate,omitempty"`
	// The unit is bit/s
	ReceivedRate *float64 `json:"receivedRate,omitempty"`
}

// RuleGroupSpec defines the desired state of RuleGroup
type RuleGroupSpec struct {
	Interval                string          `json:"interval,omitempty"`
	PartialResponseStrategy string          `json:"partial_response_strategy,omitempty"`
	Rules                   []NamespaceRule `json:"rules"`
}

// RuleGroupStatus defines the observed state of RuleGroup
type RuleGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:subresource:status
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true

// RuleGroup is the Schema for the RuleGroup API
type RuleGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RuleGroupSpec   `json:"spec,omitempty"`
	Status RuleGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RuleGroupList contains a list of RuleGroup
type RuleGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RuleGroup `json:"items"`
}

// ClusterRuleGroupSpec defines the desired state of ClusterRuleGroup
type ClusterRuleGroupSpec struct {
	Interval                string `json:"interval,omitempty"`
	PartialResponseStrategy string `json:"partial_response_strategy,omitempty"`

	Rules []ClusterRule `json:"rules"`
}

// ClusterRuleGroupStatus defines the observed state of ClusterRuleGroup
type ClusterRuleGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster

// ClusterRuleGroup is the Schema for the ClusterRuleGroup API
type ClusterRuleGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterRuleGroupSpec   `json:"spec,omitempty"`
	Status ClusterRuleGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterRuleGroupList contains a list of ClusterRuleGroup
type ClusterRuleGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterRuleGroup `json:"items"`
}

// GlobalRuleGroupSpec defines the desired state of GlobalRuleGroup
type GlobalRuleGroupSpec struct {
	Interval                string `json:"interval,omitempty"`
	PartialResponseStrategy string `json:"partial_response_strategy,omitempty"`

	Rules []GlobalRule `json:"rules"`
}

// GlobalRuleGroupStatus defines the observed state of GlobalRuleGroup
type GlobalRuleGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster

// GlobalRuleGroup is the Schema for the GlobalRuleGroup API
type GlobalRuleGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalRuleGroupSpec   `json:"spec,omitempty"`
	Status GlobalRuleGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GlobalRuleGroupList contains a list of GlobalRuleGroup
type GlobalRuleGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalRuleGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RuleGroup{}, &RuleGroupList{})
	SchemeBuilder.Register(&ClusterRuleGroup{}, &ClusterRuleGroupList{})
	SchemeBuilder.Register(&GlobalRuleGroup{}, &GlobalRuleGroupList{})
}
