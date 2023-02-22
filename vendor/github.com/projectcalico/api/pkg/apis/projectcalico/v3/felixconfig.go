// Copyright (c) 2017-2022 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/projectcalico/api/pkg/lib/numorstring"
)

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FelixConfigurationList contains a list of FelixConfiguration object.
type FelixConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []FelixConfiguration `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FelixConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec FelixConfigurationSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type IptablesBackend string

const (
	KindFelixConfiguration     = "FelixConfiguration"
	KindFelixConfigurationList = "FelixConfigurationList"
	IptablesBackendLegacy      = "Legacy"
	IptablesBackendNFTables    = "NFT"
	IptablesBackendAuto        = "Auto"
)

// +kubebuilder:validation:Enum=DoNothing;Enable;Disable
type AWSSrcDstCheckOption string

const (
	AWSSrcDstCheckOptionDoNothing AWSSrcDstCheckOption = "DoNothing"
	AWSSrcDstCheckOptionEnable                         = "Enable"
	AWSSrcDstCheckOptionDisable                        = "Disable"
)

// +kubebuilder:validation:Enum=Enabled;Disabled
type FloatingIPType string

const (
	FloatingIPsEnabled  FloatingIPType = "Enabled"
	FloatingIPsDisabled FloatingIPType = "Disabled"
)

// FelixConfigurationSpec contains the values of the Felix configuration.
type FelixConfigurationSpec struct {
	// UseInternalDataplaneDriver, if true, Felix will use its internal dataplane programming logic.  If false, it
	// will launch an external dataplane driver and communicate with it over protobuf.
	UseInternalDataplaneDriver *bool `json:"useInternalDataplaneDriver,omitempty"`
	// DataplaneDriver filename of the external dataplane driver to use.  Only used if UseInternalDataplaneDriver
	// is set to false.
	DataplaneDriver string `json:"dataplaneDriver,omitempty"`

	// DataplaneWatchdogTimeout is the readiness/liveness timeout used for Felix's (internal) dataplane driver.
	// Increase this value if you experience spurious non-ready or non-live events when Felix is under heavy load.
	// Decrease the value to get felix to report non-live or non-ready more quickly. [Default: 90s]
	//
	// Deprecated: replaced by the generic HealthTimeoutOverrides.
	DataplaneWatchdogTimeout *metav1.Duration `json:"dataplaneWatchdogTimeout,omitempty" configv1timescale:"seconds"`

	// IPv6Support controls whether Felix enables support for IPv6 (if supported by the in-use dataplane).
	IPv6Support *bool `json:"ipv6Support,omitempty" confignamev1:"Ipv6Support"`

	// RouteRefreshInterval is the period at which Felix re-checks the routes
	// in the dataplane to ensure that no other process has accidentally broken Calico's rules.
	// Set to 0 to disable route refresh. [Default: 90s]
	RouteRefreshInterval *metav1.Duration `json:"routeRefreshInterval,omitempty" configv1timescale:"seconds"`
	// InterfaceRefreshInterval is the period at which Felix rescans local interfaces to verify their state.
	// The rescan can be disabled by setting the interval to 0.
	InterfaceRefreshInterval *metav1.Duration `json:"interfaceRefreshInterval,omitempty" configv1timescale:"seconds"`
	// IptablesRefreshInterval is the period at which Felix re-checks the IP sets
	// in the dataplane to ensure that no other process has accidentally broken Calico's rules.
	// Set to 0 to disable IP sets refresh. Note: the default for this value is lower than the
	// other refresh intervals as a workaround for a Linux kernel bug that was fixed in kernel
	// version 4.11. If you are using v4.11 or greater you may want to set this to, a higher value
	// to reduce Felix CPU usage. [Default: 10s]
	IptablesRefreshInterval *metav1.Duration `json:"iptablesRefreshInterval,omitempty" configv1timescale:"seconds"`
	// IptablesPostWriteCheckInterval is the period after Felix has done a write
	// to the dataplane that it schedules an extra read back in order to check the write was not
	// clobbered by another process. This should only occur if another application on the system
	// doesn't respect the iptables lock. [Default: 1s]
	IptablesPostWriteCheckInterval *metav1.Duration `json:"iptablesPostWriteCheckInterval,omitempty" configv1timescale:"seconds" confignamev1:"IptablesPostWriteCheckIntervalSecs"`
	// IptablesLockFilePath is the location of the iptables lock file. You may need to change this
	// if the lock file is not in its standard location (for example if you have mapped it into Felix's
	// container at a different path). [Default: /run/xtables.lock]
	IptablesLockFilePath string `json:"iptablesLockFilePath,omitempty"`
	// IptablesLockTimeout is the time that Felix will wait for the iptables lock,
	// or 0, to disable. To use this feature, Felix must share the iptables lock file with all other
	// processes that also take the lock. When running Felix inside a container, this requires the
	// /run directory of the host to be mounted into the calico/node or calico/felix container.
	// [Default: 0s disabled]
	IptablesLockTimeout *metav1.Duration `json:"iptablesLockTimeout,omitempty" configv1timescale:"seconds" confignamev1:"IptablesLockTimeoutSecs"`
	// IptablesLockProbeInterval is the time that Felix will wait between
	// attempts to acquire the iptables lock if it is not available. Lower values make Felix more
	// responsive when the lock is contended, but use more CPU. [Default: 50ms]
	IptablesLockProbeInterval *metav1.Duration `json:"iptablesLockProbeInterval,omitempty" configv1timescale:"milliseconds" confignamev1:"IptablesLockProbeIntervalMillis"`
	// FeatureDetectOverride is used to override feature detection based on auto-detected platform
	// capabilities.  Values are specified in a comma separated list with no spaces, example;
	// "SNATFullyRandom=true,MASQFullyRandom=false,RestoreSupportsLock=".  "true" or "false" will
	// force the feature, empty or omitted values are auto-detected.
	FeatureDetectOverride string `json:"featureDetectOverride,omitempty" validate:"omitempty,keyValueList"`
	// FeatureGates is used to enable or disable tech-preview Calico features.
	// Values are specified in a comma separated list with no spaces, example;
	// "BPFConnectTimeLoadBalancingWorkaround=enabled,XyZ=false". This is
	// used to enable features that are not fully production ready.
	FeatureGates string `json:"featureGates,omitempty" validate:"omitempty,keyValueList"`
	// IpsetsRefreshInterval is the period at which Felix re-checks all iptables
	// state to ensure that no other process has accidentally broken Calico's rules. Set to 0 to
	// disable iptables refresh. [Default: 90s]
	IpsetsRefreshInterval *metav1.Duration `json:"ipsetsRefreshInterval,omitempty" configv1timescale:"seconds"`
	MaxIpsetSize          *int             `json:"maxIpsetSize,omitempty"`
	// IptablesBackend specifies which backend of iptables will be used. The default is Auto.
	IptablesBackend *IptablesBackend `json:"iptablesBackend,omitempty" validate:"omitempty,iptablesBackend"`

	// XDPRefreshInterval is the period at which Felix re-checks all XDP state to ensure that no
	// other process has accidentally broken Calico's BPF maps or attached programs. Set to 0 to
	// disable XDP refresh. [Default: 90s]
	XDPRefreshInterval *metav1.Duration `json:"xdpRefreshInterval,omitempty" configv1timescale:"seconds"`

	NetlinkTimeout *metav1.Duration `json:"netlinkTimeout,omitempty" configv1timescale:"seconds" confignamev1:"NetlinkTimeoutSecs"`

	// MetadataAddr is the IP address or domain name of the server that can answer VM queries for
	// cloud-init metadata. In OpenStack, this corresponds to the machine running nova-api (or in
	// Ubuntu, nova-api-metadata). A value of none (case insensitive) means that Felix should not
	// set up any NAT rule for the metadata path. [Default: 127.0.0.1]
	MetadataAddr string `json:"metadataAddr,omitempty"`
	// MetadataPort is the port of the metadata server. This, combined with global.MetadataAddr (if
	// not 'None'), is used to set up a NAT rule, from 169.254.169.254:80 to MetadataAddr:MetadataPort.
	// In most cases this should not need to be changed [Default: 8775].
	MetadataPort *int `json:"metadataPort,omitempty"`

	// OpenstackRegion is the name of the region that a particular Felix belongs to. In a multi-region
	// Calico/OpenStack deployment, this must be configured somehow for each Felix (here in the datamodel,
	// or in felix.cfg or the environment on each compute node), and must match the [calico]
	// openstack_region value configured in neutron.conf on each node. [Default: Empty]
	OpenstackRegion string `json:"openstackRegion,omitempty"`

	// InterfacePrefix is the interface name prefix that identifies workload endpoints and so distinguishes
	// them from host endpoint interfaces. Note: in environments other than bare metal, the orchestrators
	// configure this appropriately. For example our Kubernetes and Docker integrations set the 'cali' value,
	// and our OpenStack integration sets the 'tap' value. [Default: cali]
	InterfacePrefix string `json:"interfacePrefix,omitempty"`
	// InterfaceExclude is a comma-separated list of interfaces that Felix should exclude when monitoring for host
	// endpoints. The default value ensures that Felix ignores Kubernetes' IPVS dummy interface, which is used
	// internally by kube-proxy. If you want to exclude multiple interface names using a single value, the list
	// supports regular expressions. For regular expressions you must wrap the value with '/'. For example
	// having values '/^kube/,veth1' will exclude all interfaces that begin with 'kube' and also the interface
	// 'veth1'. [Default: kube-ipvs0]
	InterfaceExclude string `json:"interfaceExclude,omitempty"`

	// ChainInsertMode controls whether Felix hooks the kernel's top-level iptables chains by inserting a rule
	// at the top of the chain or by appending a rule at the bottom. insert is the safe default since it prevents
	// Calico's rules from being bypassed. If you switch to append mode, be sure that the other rules in the chains
	// signal acceptance by falling through to the Calico rules, otherwise the Calico policy will be bypassed.
	// [Default: insert]
	ChainInsertMode string `json:"chainInsertMode,omitempty"`
	// DefaultEndpointToHostAction controls what happens to traffic that goes from a workload endpoint to the host
	// itself (after the traffic hits the endpoint egress policy). By default Calico blocks traffic from workload
	// endpoints to the host itself with an iptables "DROP" action. If you want to allow some or all traffic from
	// endpoint to host, set this parameter to RETURN or ACCEPT. Use RETURN if you have your own rules in the iptables
	// "INPUT" chain; Calico will insert its rules at the top of that chain, then "RETURN" packets to the "INPUT" chain
	// once it has completed processing workload endpoint egress policy. Use ACCEPT to unconditionally accept packets
	// from workloads after processing workload endpoint egress policy. [Default: Drop]
	DefaultEndpointToHostAction string `json:"defaultEndpointToHostAction,omitempty" validate:"omitempty,dropAcceptReturn"`
	IptablesFilterAllowAction   string `json:"iptablesFilterAllowAction,omitempty" validate:"omitempty,acceptReturn"`
	IptablesMangleAllowAction   string `json:"iptablesMangleAllowAction,omitempty" validate:"omitempty,acceptReturn"`
	// LogPrefix is the log prefix that Felix uses when rendering LOG rules. [Default: calico-packet]
	LogPrefix string `json:"logPrefix,omitempty"`

	// LogFilePath is the full path to the Felix log. Set to none to disable file logging. [Default: /var/log/calico/felix.log]
	LogFilePath string `json:"logFilePath,omitempty"`

	// LogSeverityFile is the log severity above which logs are sent to the log file. [Default: Info]
	LogSeverityFile string `json:"logSeverityFile,omitempty" validate:"omitempty,logLevel"`
	// LogSeverityScreen is the log severity above which logs are sent to the stdout. [Default: Info]
	LogSeverityScreen string `json:"logSeverityScreen,omitempty" validate:"omitempty,logLevel"`
	// LogSeveritySys is the log severity above which logs are sent to the syslog. Set to None for no logging to syslog.
	// [Default: Info]
	LogSeveritySys string `json:"logSeveritySys,omitempty" validate:"omitempty,logLevel"`
	// LogDebugFilenameRegex controls which source code files have their Debug log output included in the logs.
	// Only logs from files with names that match the given regular expression are included.  The filter only applies
	// to Debug level logs.
	LogDebugFilenameRegex string `json:"logDebugFilenameRegex,omitempty" validate:"omitempty,regexp"`

	// IPIPEnabled overrides whether Felix should configure an IPIP interface on the host. Optional as Felix determines this based on the existing IP pools. [Default: nil (unset)]
	IPIPEnabled *bool `json:"ipipEnabled,omitempty" confignamev1:"IpInIpEnabled"`
	// IPIPMTU is the MTU to set on the tunnel device. See Configuring MTU [Default: 1440]
	IPIPMTU *int `json:"ipipMTU,omitempty" confignamev1:"IpInIpMtu"`

	// VXLANEnabled overrides whether Felix should create the VXLAN tunnel device for IPv4 VXLAN networking. Optional as Felix determines this based on the existing IP pools. [Default: nil (unset)]
	VXLANEnabled *bool `json:"vxlanEnabled,omitempty" confignamev1:"VXLANEnabled"`
	// VXLANMTU is the MTU to set on the IPv4 VXLAN tunnel device. See Configuring MTU [Default: 1410]
	VXLANMTU *int `json:"vxlanMTU,omitempty"`
	// VXLANMTUV6 is the MTU to set on the IPv6 VXLAN tunnel device. See Configuring MTU [Default: 1390]
	VXLANMTUV6 *int `json:"vxlanMTUV6,omitempty"`
	VXLANPort  *int `json:"vxlanPort,omitempty"`
	VXLANVNI   *int `json:"vxlanVNI,omitempty"`

	// AllowVXLANPacketsFromWorkloads controls whether Felix will add a rule to drop VXLAN encapsulated traffic
	// from workloads [Default: false]
	// +optional
	AllowVXLANPacketsFromWorkloads *bool `json:"allowVXLANPacketsFromWorkloads,omitempty"`
	// AllowIPIPPacketsFromWorkloads controls whether Felix will add a rule to drop IPIP encapsulated traffic
	// from workloads [Default: false]
	// +optional
	AllowIPIPPacketsFromWorkloads *bool `json:"allowIPIPPacketsFromWorkloads,omitempty"`

	// ReportingInterval is the interval at which Felix reports its status into the datastore or 0 to disable.
	// Must be non-zero in OpenStack deployments. [Default: 30s]
	ReportingInterval *metav1.Duration `json:"reportingInterval,omitempty" configv1timescale:"seconds" confignamev1:"ReportingIntervalSecs"`
	// ReportingTTL is the time-to-live setting for process-wide status reports. [Default: 90s]
	ReportingTTL *metav1.Duration `json:"reportingTTL,omitempty" configv1timescale:"seconds" confignamev1:"ReportingTTLSecs"`

	EndpointReportingEnabled *bool            `json:"endpointReportingEnabled,omitempty"`
	EndpointReportingDelay   *metav1.Duration `json:"endpointReportingDelay,omitempty" configv1timescale:"seconds" confignamev1:"EndpointReportingDelaySecs"`

	// IptablesMarkMask is the mask that Felix selects its IPTables Mark bits from. Should be a 32 bit hexadecimal
	// number with at least 8 bits set, none of which clash with any other mark bits in use on the system.
	// [Default: 0xff000000]
	IptablesMarkMask *uint32 `json:"iptablesMarkMask,omitempty"`

	DisableConntrackInvalidCheck *bool `json:"disableConntrackInvalidCheck,omitempty"`

	HealthEnabled *bool   `json:"healthEnabled,omitempty"`
	HealthHost    *string `json:"healthHost,omitempty"`
	HealthPort    *int    `json:"healthPort,omitempty"`
	// HealthTimeoutOverrides allows the internal watchdog timeouts of individual subcomponents to be
	// overriden.  This is useful for working around "false positive" liveness timeouts that can occur
	// in particularly stressful workloads or if CPU is constrained.  For a list of active
	// subcomponents, see Felix's logs.
	HealthTimeoutOverrides []HealthTimeoutOverride `json:"healthTimeoutOverrides,omitempty" validate:"omitempty,dive"`

	// PrometheusMetricsEnabled enables the Prometheus metrics server in Felix if set to true. [Default: false]
	PrometheusMetricsEnabled *bool `json:"prometheusMetricsEnabled,omitempty"`
	// PrometheusMetricsHost is the host that the Prometheus metrics server should bind to. [Default: empty]
	PrometheusMetricsHost string `json:"prometheusMetricsHost,omitempty" validate:"omitempty,prometheusHost"`
	// PrometheusMetricsPort is the TCP port that the Prometheus metrics server should bind to. [Default: 9091]
	PrometheusMetricsPort *int `json:"prometheusMetricsPort,omitempty"`
	// PrometheusGoMetricsEnabled disables Go runtime metrics collection, which the Prometheus client does by default, when
	// set to false. This reduces the number of metrics reported, reducing Prometheus load. [Default: true]
	PrometheusGoMetricsEnabled *bool `json:"prometheusGoMetricsEnabled,omitempty"`
	// PrometheusProcessMetricsEnabled disables process metrics collection, which the Prometheus client does by default, when
	// set to false. This reduces the number of metrics reported, reducing Prometheus load. [Default: true]
	PrometheusProcessMetricsEnabled *bool `json:"prometheusProcessMetricsEnabled,omitempty"`
	// PrometheusWireGuardMetricsEnabled disables wireguard metrics collection, which the Prometheus client does by default, when
	// set to false. This reduces the number of metrics reported, reducing Prometheus load. [Default: true]
	PrometheusWireGuardMetricsEnabled *bool `json:"prometheusWireGuardMetricsEnabled,omitempty"`

	// FailsafeInboundHostPorts is a list of UDP/TCP ports and CIDRs that Felix will allow incoming traffic to host endpoints
	// on irrespective of the security policy. This is useful to avoid accidentally cutting off a host with incorrect configuration.
	// For back-compatibility, if the protocol is not specified, it defaults to "tcp". If a CIDR is not specified, it will allow
	// traffic from all addresses. To disable all inbound host ports, use the value none. The default value allows ssh access
	// and DHCP.
	// [Default: tcp:22, udp:68, tcp:179, tcp:2379, tcp:2380, tcp:6443, tcp:6666, tcp:6667]
	FailsafeInboundHostPorts *[]ProtoPort `json:"failsafeInboundHostPorts,omitempty"`
	// FailsafeOutboundHostPorts is a list of UDP/TCP ports and CIDRs that Felix will allow outgoing traffic from host endpoints
	// to irrespective of the security policy. This is useful to avoid accidentally cutting off a host with incorrect configuration.
	// For back-compatibility, if the protocol is not specified, it defaults to "tcp". If a CIDR is not specified, it will allow
	// traffic from all addresses. To disable all outbound host ports, use the value none. The default value opens etcd's standard
	// ports to ensure that Felix does not get cut off from etcd as well as allowing DHCP and DNS.
	// [Default: tcp:179, tcp:2379, tcp:2380, tcp:6443, tcp:6666, tcp:6667, udp:53, udp:67]
	FailsafeOutboundHostPorts *[]ProtoPort `json:"failsafeOutboundHostPorts,omitempty"`

	// KubeNodePortRanges holds list of port ranges used for service node ports. Only used if felix detects kube-proxy running in ipvs mode.
	// Felix uses these ranges to separate host and workload traffic. [Default: 30000:32767].
	KubeNodePortRanges *[]numorstring.Port `json:"kubeNodePortRanges,omitempty" validate:"omitempty,dive"`

	// PolicySyncPathPrefix is used to by Felix to communicate policy changes to external services,
	// like Application layer policy. [Default: Empty]
	PolicySyncPathPrefix string `json:"policySyncPathPrefix,omitempty"`

	// UsageReportingEnabled reports anonymous Calico version number and cluster size to projectcalico.org. Logs warnings returned by the usage
	// server. For example, if a significant security vulnerability has been discovered in the version of Calico being used. [Default: true]
	UsageReportingEnabled *bool `json:"usageReportingEnabled,omitempty"`
	// UsageReportingInitialDelay controls the minimum delay before Felix makes a report. [Default: 300s]
	UsageReportingInitialDelay *metav1.Duration `json:"usageReportingInitialDelay,omitempty" configv1timescale:"seconds" confignamev1:"UsageReportingInitialDelaySecs"`
	// UsageReportingInterval controls the interval at which Felix makes reports. [Default: 86400s]
	UsageReportingInterval *metav1.Duration `json:"usageReportingInterval,omitempty" configv1timescale:"seconds" confignamev1:"UsageReportingIntervalSecs"`

	// NATPortRange specifies the range of ports that is used for port mapping when doing outgoing NAT. When unset the default behavior of the
	// network stack is used.
	NATPortRange *numorstring.Port `json:"natPortRange,omitempty"`

	// NATOutgoingAddress specifies an address to use when performing source NAT for traffic in a natOutgoing pool that
	// is leaving the network. By default the address used is an address on the interface the traffic is leaving on
	// (ie it uses the iptables MASQUERADE target)
	NATOutgoingAddress string `json:"natOutgoingAddress,omitempty"`

	// This is the IPv4 source address to use on programmed device routes. By default the source address is left blank,
	// leaving the kernel to choose the source address used.
	DeviceRouteSourceAddress string `json:"deviceRouteSourceAddress,omitempty"`

	// This is the IPv6 source address to use on programmed device routes. By default the source address is left blank,
	// leaving the kernel to choose the source address used.
	DeviceRouteSourceAddressIPv6 string `json:"deviceRouteSourceAddressIPv6,omitempty"`

	// This defines the route protocol added to programmed device routes, by default this will be RTPROT_BOOT
	// when left blank.
	DeviceRouteProtocol *int `json:"deviceRouteProtocol,omitempty"`
	// Whether or not to remove device routes that have not been programmed by Felix. Disabling this will allow external
	// applications to also add device routes. This is enabled by default which means we will remove externally added routes.
	RemoveExternalRoutes *bool `json:"removeExternalRoutes,omitempty"`

	// ExternalNodesCIDRList is a list of CIDR's of external-non-calico-nodes which may source tunnel traffic and have
	// the tunneled traffic be accepted at calico nodes.
	ExternalNodesCIDRList *[]string `json:"externalNodesList,omitempty"`

	DebugMemoryProfilePath          string           `json:"debugMemoryProfilePath,omitempty"`
	DebugDisableLogDropping         *bool            `json:"debugDisableLogDropping,omitempty"`
	DebugSimulateCalcGraphHangAfter *metav1.Duration `json:"debugSimulateCalcGraphHangAfter,omitempty" configv1timescale:"seconds"`
	DebugSimulateDataplaneHangAfter *metav1.Duration `json:"debugSimulateDataplaneHangAfter,omitempty" configv1timescale:"seconds"`

	IptablesNATOutgoingInterfaceFilter string `json:"iptablesNATOutgoingInterfaceFilter,omitempty" validate:"omitempty,ifaceFilter"`

	// SidecarAccelerationEnabled enables experimental sidecar acceleration [Default: false]
	SidecarAccelerationEnabled *bool `json:"sidecarAccelerationEnabled,omitempty"`

	// XDPEnabled enables XDP acceleration for suitable untracked incoming deny rules. [Default: true]
	XDPEnabled *bool `json:"xdpEnabled,omitempty" confignamev1:"XDPEnabled"`

	// GenericXDPEnabled enables Generic XDP so network cards that don't support XDP offload or driver
	// modes can use XDP. This is not recommended since it doesn't provide better performance than
	// iptables. [Default: false]
	GenericXDPEnabled *bool `json:"genericXDPEnabled,omitempty" confignamev1:"GenericXDPEnabled"`

	// BPFEnabled, if enabled Felix will use the BPF dataplane. [Default: false]
	BPFEnabled *bool `json:"bpfEnabled,omitempty" validate:"omitempty"`
	// BPFDisableUnprivileged, if enabled, Felix sets the kernel.unprivileged_bpf_disabled sysctl to disable
	// unprivileged use of BPF.  This ensures that unprivileged users cannot access Calico's BPF maps and
	// cannot insert their own BPF programs to interfere with Calico's. [Default: true]
	BPFDisableUnprivileged *bool `json:"bpfDisableUnprivileged,omitempty" validate:"omitempty"`
	// BPFLogLevel controls the log level of the BPF programs when in BPF dataplane mode.  One of "Off", "Info", or
	// "Debug".  The logs are emitted to the BPF trace pipe, accessible with the command `tc exec bpf debug`.
	// [Default: Off].
	// +optional
	BPFLogLevel string `json:"bpfLogLevel" validate:"omitempty,bpfLogLevel"`
	// BPFDataIfacePattern is a regular expression that controls which interfaces Felix should attach BPF programs to
	// in order to catch traffic to/from the network.  This needs to match the interfaces that Calico workload traffic
	// flows over as well as any interfaces that handle incoming traffic to nodeports and services from outside the
	// cluster.  It should not match the workload interfaces (usually named cali...).
	BPFDataIfacePattern string `json:"bpfDataIfacePattern,omitempty" validate:"omitempty,regexp"`
	// BPFL3IfacePattern is a regular expression that allows to list tunnel devices like wireguard or vxlan (i.e., L3 devices)
	// in addition to BPFDataIfacePattern. That is, tunnel interfaces not created by Calico, that Calico workload traffic flows
	// over as well as any interfaces that handle incoming traffic to nodeports and services from outside the cluster.
	BPFL3IfacePattern string `json:"bpfL3IfacePattern,omitempty" validate:"omitempty,regexp"`
	// BPFConnectTimeLoadBalancingEnabled when in BPF mode, controls whether Felix installs the connection-time load
	// balancer.  The connect-time load balancer is required for the host to be able to reach Kubernetes services
	// and it improves the performance of pod-to-service connections.  The only reason to disable it is for debugging
	// purposes.  [Default: true]
	BPFConnectTimeLoadBalancingEnabled *bool `json:"bpfConnectTimeLoadBalancingEnabled,omitempty" validate:"omitempty"`
	// BPFExternalServiceMode in BPF mode, controls how connections from outside the cluster to services (node ports
	// and cluster IPs) are forwarded to remote workloads.  If set to "Tunnel" then both request and response traffic
	// is tunneled to the remote node.  If set to "DSR", the request traffic is tunneled but the response traffic
	// is sent directly from the remote node.  In "DSR" mode, the remote node appears to use the IP of the ingress
	// node; this requires a permissive L2 network.  [Default: Tunnel]
	BPFExternalServiceMode string `json:"bpfExternalServiceMode,omitempty" validate:"omitempty,bpfServiceMode"`
	// BPFDSROptoutCIDRs is a list of CIDRs which are excluded from DSR. That is, clients
	// in those CIDRs will accesses nodeports as if BPFExternalServiceMode was set to
	// Tunnel.
	BPFDSROptoutCIDRs *[]string `json:"bpfDSROptoutCIDRs,omitempty" validate:"omitempty,cidrs"`
	// BPFExtToServiceConnmark in BPF mode, control a 32bit mark that is set on connections from an
	// external client to a local service. This mark allows us to control how packets of that
	// connection are routed within the host and how is routing interpreted by RPF check. [Default: 0]
	BPFExtToServiceConnmark *int `json:"bpfExtToServiceConnmark,omitempty" validate:"omitempty,gte=0,lte=4294967295"`
	// BPFKubeProxyIptablesCleanupEnabled, if enabled in BPF mode, Felix will proactively clean up the upstream
	// Kubernetes kube-proxy's iptables chains.  Should only be enabled if kube-proxy is not running.  [Default: true]
	BPFKubeProxyIptablesCleanupEnabled *bool `json:"bpfKubeProxyIptablesCleanupEnabled,omitempty" validate:"omitempty"`
	// BPFKubeProxyMinSyncPeriod, in BPF mode, controls the minimum time between updates to the dataplane for Felix's
	// embedded kube-proxy.  Lower values give reduced set-up latency.  Higher values reduce Felix CPU usage by
	// batching up more work.  [Default: 1s]
	BPFKubeProxyMinSyncPeriod *metav1.Duration `json:"bpfKubeProxyMinSyncPeriod,omitempty" validate:"omitempty" configv1timescale:"seconds"`
	// BPFKubeProxyEndpointSlicesEnabled in BPF mode, controls whether Felix's
	// embedded kube-proxy accepts EndpointSlices or not.
	BPFKubeProxyEndpointSlicesEnabled *bool `json:"bpfKubeProxyEndpointSlicesEnabled,omitempty" validate:"omitempty"`
	// BPFPSNATPorts sets the range from which we randomly pick a port if there is a source port
	// collision. This should be within the ephemeral range as defined by RFC 6056 (1024–65535) and
	// preferably outside the  ephemeral ranges used by common operating systems. Linux uses
	// 32768–60999, while others mostly use the IANA defined range 49152–65535. It is not necessarily
	// a problem if this range overlaps with the operating systems. Both ends of the range are
	// inclusive. [Default: 20000:29999]
	BPFPSNATPorts *numorstring.Port `json:"bpfPSNATPorts,omitempty"`
	// BPFMapSizeNATFrontend sets the size for nat front end map.
	// FrontendMap should be large enough to hold an entry for each nodeport,
	// external IP and each port in each service.
	BPFMapSizeNATFrontend *int `json:"bpfMapSizeNATFrontend,omitempty"`
	// BPFMapSizeNATBackend sets the size for nat back end map.
	// This is the total number of endpoints. This is mostly
	// more than the size of the number of services.
	BPFMapSizeNATBackend  *int `json:"bpfMapSizeNATBackend,omitempty"`
	BPFMapSizeNATAffinity *int `json:"bpfMapSizeNATAffinity,omitempty"`
	// BPFMapSizeRoute sets the size for the routes map.  The routes map should be large enough
	// to hold one entry per workload and a handful of entries per host (enough to cover its own IPs and
	// tunnel IPs).
	BPFMapSizeRoute *int `json:"bpfMapSizeRoute,omitempty"`
	// BPFMapSizeConntrack sets the size for the conntrack map.  This map must be large enough to hold
	// an entry for each active connection.  Warning: changing the size of the conntrack map can cause disruption.
	BPFMapSizeConntrack *int `json:"bpfMapSizeConntrack,omitempty"`
	// BPFMapSizeIPSets sets the size for ipsets map.  The IP sets map must be large enough to hold an entry
	// for each endpoint matched by every selector in the source/destination matches in network policy.  Selectors
	// such as "all()" can result in large numbers of entries (one entry per endpoint in that case).
	BPFMapSizeIPSets *int `json:"bpfMapSizeIPSets,omitempty"`
	// BPFMapSizeIfState sets the size for ifstate map.  The ifstate map must be large enough to hold an entry
	// for each device (host + workloads) on a host.
	BPFMapSizeIfState *int `json:"bpfMapSizeIfState,omitempty"`
	// BPFHostConntrackBypass Controls whether to bypass Linux conntrack in BPF mode for
	// workloads and services. [Default: true - bypass Linux conntrack]
	BPFHostConntrackBypass *bool `json:"bpfHostConntrackBypass,omitempty"`
	// BPFEnforceRPF enforce strict RPF on all host interfaces with BPF programs regardless of
	// what is the per-interfaces or global setting. Possible values are Disabled, Strict
	// or Loose. [Default: Strict]
	BPFEnforceRPF string `json:"bpfEnforceRPF,omitempty"`
	// BPFPolicyDebugEnabled when true, Felix records detailed information
	// about the BPF policy programs, which can be examined with the calico-bpf command-line tool.
	BPFPolicyDebugEnabled *bool `json:"bpfPolicyDebugEnabled,omitempty"`
	// RouteSource configures where Felix gets its routing information.
	// - WorkloadIPs: use workload endpoints to construct routes.
	// - CalicoIPAM: the default - use IPAM data to construct routes.
	RouteSource string `json:"routeSource,omitempty" validate:"omitempty,routeSource"`

	// Calico programs additional Linux route tables for various purposes.
	// RouteTableRanges specifies a set of table index ranges that Calico should use.
	// Deprecates`RouteTableRange`, overrides `RouteTableRange`.
	RouteTableRanges *RouteTableRanges `json:"routeTableRanges,omitempty" validate:"omitempty,dive"`

	// Deprecated in favor of RouteTableRanges.
	// Calico programs additional Linux route tables for various purposes.
	// RouteTableRange specifies the indices of the route tables that Calico should use.
	RouteTableRange *RouteTableRange `json:"routeTableRange,omitempty" validate:"omitempty"`

	// RouteSyncDisabled will disable all operations performed on the route table. Set to true to
	// run in network-policy mode only.
	RouteSyncDisabled *bool `json:"routeSyncDisabled,omitempty"`

	// WireguardEnabled controls whether Wireguard is enabled for IPv4 (encapsulating IPv4 traffic over an IPv4 underlay network). [Default: false]
	WireguardEnabled *bool `json:"wireguardEnabled,omitempty"`
	// WireguardEnabledV6 controls whether Wireguard is enabled for IPv6 (encapsulating IPv6 traffic over an IPv6 underlay network). [Default: false]
	WireguardEnabledV6 *bool `json:"wireguardEnabledV6,omitempty"`
	// WireguardListeningPort controls the listening port used by IPv4 Wireguard. [Default: 51820]
	WireguardListeningPort *int `json:"wireguardListeningPort,omitempty" validate:"omitempty,gt=0,lte=65535"`
	// WireguardListeningPortV6 controls the listening port used by IPv6 Wireguard. [Default: 51821]
	WireguardListeningPortV6 *int `json:"wireguardListeningPortV6,omitempty" validate:"omitempty,gt=0,lte=65535"`
	// WireguardRoutingRulePriority controls the priority value to use for the Wireguard routing rule. [Default: 99]
	WireguardRoutingRulePriority *int `json:"wireguardRoutingRulePriority,omitempty" validate:"omitempty,gt=0,lt=32766"`
	// WireguardInterfaceName specifies the name to use for the IPv4 Wireguard interface. [Default: wireguard.cali]
	WireguardInterfaceName string `json:"wireguardInterfaceName,omitempty" validate:"omitempty,interface"`
	// WireguardInterfaceNameV6 specifies the name to use for the IPv6 Wireguard interface. [Default: wg-v6.cali]
	WireguardInterfaceNameV6 string `json:"wireguardInterfaceNameV6,omitempty" validate:"omitempty,interface"`
	// WireguardMTU controls the MTU on the IPv4 Wireguard interface. See Configuring MTU [Default: 1440]
	WireguardMTU *int `json:"wireguardMTU,omitempty"`
	// WireguardMTUV6 controls the MTU on the IPv6 Wireguard interface. See Configuring MTU [Default: 1420]
	WireguardMTUV6 *int `json:"wireguardMTUV6,omitempty"`
	// WireguardHostEncryptionEnabled controls whether Wireguard host-to-host encryption is enabled. [Default: false]
	WireguardHostEncryptionEnabled *bool `json:"wireguardHostEncryptionEnabled,omitempty"`
	// WireguardKeepAlive controls Wireguard PersistentKeepalive option. Set 0 to disable. [Default: 0]
	WireguardPersistentKeepAlive *metav1.Duration `json:"wireguardKeepAlive,omitempty"`

	// Set source-destination-check on AWS EC2 instances. Accepted value must be one of "DoNothing", "Enable" or "Disable".
	// [Default: DoNothing]
	AWSSrcDstCheck *AWSSrcDstCheckOption `json:"awsSrcDstCheck,omitempty" validate:"omitempty,oneof=DoNothing Enable Disable"`

	// When service IP advertisement is enabled, prevent routing loops to service IPs that are
	// not in use, by dropping or rejecting packets that do not get DNAT'd by kube-proxy.
	// Unless set to "Disabled", in which case such routing loops continue to be allowed.
	// [Default: Drop]
	ServiceLoopPrevention string `json:"serviceLoopPrevention,omitempty" validate:"omitempty,oneof=Drop Reject Disabled"`

	// WorkloadSourceSpoofing controls whether pods can use the allowedSourcePrefixes annotation to send traffic with a source IP
	// address that is not theirs. This is disabled by default. When set to "Any", pods can request any prefix.
	WorkloadSourceSpoofing string `json:"workloadSourceSpoofing,omitempty" validate:"omitempty,oneof=Disabled Any"`

	// MTUIfacePattern is a regular expression that controls which interfaces Felix should scan in order
	// to calculate the host's MTU.
	// This should not match workload interfaces (usually named cali...).
	// +optional
	MTUIfacePattern string `json:"mtuIfacePattern,omitempty" validate:"omitempty,regexp"`

	// FloatingIPs configures whether or not Felix will program non-OpenStack floating IP addresses.  (OpenStack-derived
	// floating IPs are always programmed, regardless of this setting.)
	//
	// +optional
	FloatingIPs *FloatingIPType `json:"floatingIPs,omitempty" validate:"omitempty"`
}

type HealthTimeoutOverride struct {
	Name    string          `json:"name"`
	Timeout metav1.Duration `json:"timeout"`
}

type RouteTableRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type RouteTableIDRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type RouteTableRanges []RouteTableIDRange

func (r RouteTableRanges) NumDesignatedTables() int {
	var len int = 0
	for _, rng := range r {
		len += (rng.Max - rng.Min) + 1 // add one, since range is inclusive
	}

	return len
}

// ProtoPort is combination of protocol, port, and CIDR. Protocol and port must be specified.
type ProtoPort struct {
	Protocol string `json:"protocol"`
	Port     uint16 `json:"port"`
	// +optional
	Net string `json:"net"`
}

// New FelixConfiguration creates a new (zeroed) FelixConfiguration struct with the TypeMetadata
// initialized to the current version.
func NewFelixConfiguration() *FelixConfiguration {
	return &FelixConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindFelixConfiguration,
			APIVersion: GroupVersionCurrent,
		},
	}
}
