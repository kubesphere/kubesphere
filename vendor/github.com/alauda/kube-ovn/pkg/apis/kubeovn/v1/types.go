package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ProtocolIPv4 = "IPv4"
	ProtocolIPv6 = "IPv6"
	ProtocolDual = "Dual"

	GWDistributedType = "distributed"
	GWCentralizedType = "centralized"
)

// Constants for condition
const (
	// Ready => controller considers this resource Ready
	Ready = "Ready"
	// Validated => Spec passed validating
	Validated = "Validated"
	// Error => last recorded error
	Error = "Error"

	ReasonInit = "Init"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

type IP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IPSpec `json:"spec"`
}

type IPSpec struct {
	PodName       string   `json:"podName"`
	Namespace     string   `json:"namespace"`
	Subnet        string   `json:"subnet"`
	AttachSubnets []string `json:"attachSubnets"`
	NodeName      string   `json:"nodeName"`
	IPAddress     string   `json:"ipAddress"`
	V4IPAddress   string   `json:"v4IpAddress"`
	V6IPAddress   string   `json:"v6IpAddress"`
	AttachIPs     []string `json:"attachIps"`
	MacAddress    string   `json:"macAddress"`
	AttachMacs    []string `json:"attachMacs"`
	ContainerID   string   `json:"containerID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []IP `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

type Subnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubnetSpec   `json:"spec"`
	Status SubnetStatus `json:"status,omitempty"`
}

type SubnetSpec struct {
	Default    bool     `json:"default"`
	Vpc        string   `json:"vpc,omitempty"`
	Protocol   string   `json:"protocol"`
	Namespaces []string `json:"namespaces,omitempty"`
	CIDRBlock  string   `json:"cidrBlock"`
	Gateway    string   `json:"gateway"`
	ExcludeIps []string `json:"excludeIps,omitempty"`
	Provider   string   `json:"provider,omitempty"`

	GatewayType string `json:"gatewayType"`
	GatewayNode string `json:"gatewayNode"`
	NatOutgoing bool   `json:"natOutgoing"`

	Private      bool     `json:"private"`
	AllowSubnets []string `json:"allowSubnets,omitempty"`

	Vlan            string `json:"vlan,omitempty"`
	UnderlayGateway bool   `json:"underlayGateway"`

	DisableInterConnection bool `json:"disableInterConnection"`
}

// ConditionType encodes information on the condition
type ConditionType string

// Condition describes the state of an object at a certain point.
// +k8s:deepcopy-gen=true
type SubnetCondition struct {
	// Type of condition.
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
	// Last time the condition was probed
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

type SubnetStatus struct {
	// Conditions represents the latest state of the object
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []SubnetCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	AvailableIPs    float64 `json:"availableIPs"`
	UsingIPs        float64 `json:"usingIPs"`
	V4AvailableIPs  float64 `json:"v4availableIPs"`
	V4UsingIPs      float64 `json:"v4usingIPs"`
	V6AvailableIPs  float64 `json:"v6availableIPs"`
	V6UsingIPs      float64 `json:"v6usingIPs"`
	ActivateGateway string  `json:"activateGateway"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Subnet `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

type Vlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VlanSpec   `json:"spec"`
	Status VlanStatus `json:"status"`
}

type VlanSpec struct {
	VlanId                int    `json:"vlanId"`
	ProviderInterfaceName string `json:"providerInterfaceName,omitempty"`
	LogicalInterfaceName  string `json:"logicalInterfaceName,omitempty"`
	Subnet                string `json:"subnet"`
}

type VlanStatus struct {
	// Conditions represents the latest state of the object
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []VlanCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// Condition describes the state of an object at a certain point.
// +k8s:deepcopy-gen=true
type VlanCondition struct {
	// Type of condition.
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
	// Last time the condition was probed
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Vlan `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

type Vpc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VpcSpec   `json:"spec"`
	Status VpcStatus `json:"status,omitempty"`
}

type VpcSpec struct {
	Namespaces   []string       `json:"namespaces,omitempty"`
	StaticRoutes []*StaticRoute `json:"staticRoutes,omitempty"`
}

type RoutePolicy string

const (
	PolicySrc RoutePolicy = "policySrc"
	PolicyDst RoutePolicy = "policyDst"
)

type StaticRoute struct {
	Policy    RoutePolicy `json:"policy,omitempty"`
	CIDR      string      `json:"cidr"`
	NextHopIP string      `json:"nextHopIP"`
}

type VpcStatus struct {
	// Conditions represents the latest state of the object
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []VpcCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	Standby                bool     `json:"standby"`
	Default                bool     `json:"default"`
	DefaultLogicalSwitch   string   `json:"defaultLogicalSwitch"`
	Router                 string   `json:"router"`
	TcpLoadBalancer        string   `json:"tcpLoadBalancer"`
	UdpLoadBalancer        string   `json:"udpLoadBalancer"`
	TcpSessionLoadBalancer string   `json:"tcpSessionLoadBalancer"`
	UdpSessionLoadBalancer string   `json:"udpSessionLoadBalancer"`
	Subnets                []string `json:"subnets"`
}

// Condition describes the state of an object at a certain point.
// +k8s:deepcopy-gen=true
type VpcCondition struct {
	// Type of condition.
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
	// Last time the condition was probed
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VpcList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Vpc `json:"items"`
}
