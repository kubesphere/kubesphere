/*
Copyright 2018 The Knative Authors

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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DestinationRule
type DestinationRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DestinationRuleSpec `json:"spec"`
}

// DestinationRule defines policies that apply to traffic intended for a
// service after routing has occurred. These rules specify configuration
// for load balancing, connection pool size from the sidecar, and outlier
// detection settings to detect and evict unhealthy hosts from the load
// balancing pool. For example, a simple load balancing policy for the
// ratings service would look as follows:
//
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: bookinfo-ratings
// spec:
//   host: ratings.prod.svc.cluster.local
//   trafficPolicy:
//     loadBalancer:
//       simple: LEAST_CONN
//
//
// Version specific policies can be specified by defining a named
// subset and overriding the settings specified at the service level. The
// following rule uses a round robin load balancing policy for all traffic
// going to a subset named testversion that is composed of endpoints (e.g.,
// pods) with labels (version:v3).
//
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: bookinfo-ratings
// spec:
//   host: ratings.prod.svc.cluster.local
//   trafficPolicy:
//     loadBalancer:
//       simple: LEAST_CONN
//   subsets:
//   - name: testversion
//     labels:
//       version: v3
//     trafficPolicy:
//       loadBalancer:
//         simple: ROUND_ROBIN
//
//
// **Note:** Policies specified for subsets will not take effect until
// a route rule explicitly sends traffic to this subset.
//
// Traffic policies can be customized to specific ports as well. The
// following rule uses the least connection load balancing policy for all
// traffic to port 80, while uses a round robin load balancing setting for
// traffic to the port 9080.
//
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: bookinfo-ratings-port
// spec:
//   host: ratings.prod.svc.cluster.local
//   trafficPolicy: # Apply to all ports
//     portLevelSettings:
//     - port:
//         number: 80
//       loadBalancer:
//         simple: LEAST_CONN
//     - port:
//         number: 9080
//       loadBalancer:
//         simple: ROUND_ROBIN
//
type DestinationRuleSpec struct {
	// REQUIRED. The name of a service from the service registry. Service
	// names are looked up from the platform's service registry (e.g.,
	// Kubernetes services, Consul services, etc.) and from the hosts
	// declared by [ServiceEntries](#ServiceEntry). Rules defined for
	// services that do not exist in the service registry will be ignored.
	//
	// *Note for Kubernetes users*: When short names are used (e.g. "reviews"
	// instead of "reviews.default.svc.cluster.local"), Istio will interpret
	// the short name based on the namespace of the rule, not the service. A
	// rule in the "default" namespace containing a host "reviews will be
	// interpreted as "reviews.default.svc.cluster.local", irrespective of
	// the actual namespace associated with the reviews service. _To avoid
	// potential misconfigurations, it is recommended to always use fully
	// qualified domain names over short names._
	//
	// Note that the host field applies to both HTTP and TCP services.
	Host string `json:"host"`

	// Traffic policies to apply (load balancing policy, connection pool
	// sizes, outlier detection).
	TrafficPolicy *TrafficPolicy `json:"trafficPolicy,omitempty"`

	// One or more named sets that represent individual versions of a
	// service. Traffic policies can be overridden at subset level.
	Subsets []Subset `json:"subsets,omitempty"`
}

// Traffic policies to apply for a specific destination, across all
// destination ports. See DestinationRule for examples.
type TrafficPolicy struct {

	// Settings controlling the load balancer algorithms.
	LoadBalancer *LoadBalancerSettings `json:"loadBalancer,omitempty"`

	// Settings controlling the volume of connections to an upstream service
	ConnectionPool *ConnectionPoolSettings `json:"connectionPool,omitempty"`

	// Settings controlling eviction of unhealthy hosts from the load balancing pool
	OutlierDetection *OutlierDetection `json:"outlierDetection,omitempty"`

	// TLS related settings for connections to the upstream service.
	Tls *TLSSettings `json:"tls,omitempty"`

	// Traffic policies specific to individual ports. Note that port level
	// settings will override the destination-level settings. Traffic
	// settings specified at the destination-level will not be inherited when
	// overridden by port-level settings, i.e. default values will be applied
	// to fields omitted in port-level traffic policies.
	PortLevelSettings []PortTrafficPolicy `json:"portLevelSettings,omitempty"`
}

// Traffic policies that apply to specific ports of the service
type PortTrafficPolicy struct {
	// Specifies the port name or number of a port on the destination service
	// on which this policy is being applied.
	//
	// Names must comply with DNS label syntax (rfc1035) and therefore cannot
	// collide with numbers. If there are multiple ports on a service with
	// the same protocol the names should be of the form <protocol-name>-<DNS
	// label>.
	Port PortSelector `json:"port"`

	// Settings controlling the load balancer algorithms.
	LoadBalancer *LoadBalancerSettings `json:"loadBalancer,omitempty"`

	// Settings controlling the volume of connections to an upstream service
	ConnectionPool *ConnectionPoolSettings `json:"connectionPool,omitempty"`

	// Settings controlling eviction of unhealthy hosts from the load balancing pool
	OutlierDetection *OutlierDetection `json:"outlierDetection,omitempty"`

	// TLS related settings for connections to the upstream service.
	Tls *TLSSettings `json:"tls,omitempty"`
}

// A subset of endpoints of a service. Subsets can be used for scenarios
// like A/B testing, or routing to a specific version of a service. Refer
// to [VirtualService](#VirtualService) documentation for examples of using
// subsets in these scenarios. In addition, traffic policies defined at the
// service-level can be overridden at a subset-level. The following rule
// uses a round robin load balancing policy for all traffic going to a
// subset named testversion that is composed of endpoints (e.g., pods) with
// labels (version:v3).
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: bookinfo-ratings
// spec:
//   host: ratings.prod.svc.cluster.local
//   trafficPolicy:
//     loadBalancer:
//       simple: LEAST_CONN
//   subsets:
//   - name: testversion
//     labels:
//       version: v3
//     trafficPolicy:
//       loadBalancer:
//         simple: ROUND_ROBIN
//
// **Note:** Policies specified for subsets will not take effect until
// a route rule explicitly sends traffic to this subset.
type Subset struct {
	// REQUIRED. Name of the subset. The service name and the subset name can
	// be used for traffic splitting in a route rule.
	Name string `json:"name"`

	// REQUIRED. Labels apply a filter over the endpoints of a service in the
	// service registry. See route rules for examples of usage.
	Labels map[string]string `json:"labels"`

	// Traffic policies that apply to this subset. Subsets inherit the
	// traffic policies specified at the DestinationRule level. Settings
	// specified at the subset level will override the corresponding settings
	// specified at the DestinationRule level.
	TrafficPolicy *TrafficPolicy `json:"trafficPolicy,omitempty"`
}

// Load balancing policies to apply for a specific destination. See Envoy's
// load balancing
// [documentation](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/load_balancing.html)
// for more details.
//
// For example, the following rule uses a round robin load balancing policy
// for all traffic going to the ratings service.
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: bookinfo-ratings
// spec:
//   host: ratings.prod.svc.cluster.local
//   trafficPolicy:
//     loadBalancer:
//       simple: ROUND_ROBIN
//
// The following example sets up sticky sessions for the ratings service
// hashing-based load balancer for the same ratings service using the
// the User cookie as the hash key.
//
//  apiVersion: networking.istio.io/v1alpha3
//  kind: DestinationRule
//  metadata:
//    name: bookinfo-ratings
//  spec:
//    host: ratings.prod.svc.cluster.local
//    trafficPolicy:
//      loadBalancer:
//        consistentHash:
//          httpCookie:
//            name: user
//            ttl: 0s
type LoadBalancerSettings struct {
	// It is required to specify exactly one of the fields:
	// Simple or ConsistentHash
	Simple         SimpleLB          `json:"simple,omitempty"`
	ConsistentHash *ConsistentHashLB `json:"consistentHash,omitempty"`
}

// Standard load balancing algorithms that require no tuning.
type SimpleLB string

const (
	// Round Robin policy. Default
	SimpleLBRoundRobin SimpleLB = "ROUND_ROBIN"

	// The least request load balancer uses an O(1) algorithm which selects
	// two random healthy hosts and picks the host which has fewer active
	// requests.
	SimpleLBLeastConn SimpleLB = "LEAST_CONN"

	// The random load balancer selects a random healthy host. The random
	// load balancer generally performs better than round robin if no health
	// checking policy is configured.
	SimpleLBRandom SimpleLB = "RANDOM"

	// This option will forward the connection to the original IP address
	// requested by the caller without doing any form of load
	// balancing. This option must be used with care. It is meant for
	// advanced use cases. Refer to Original Destination load balancer in
	// Envoy for further details.
	SimpleLBPassthrough SimpleLB = "PASSTHROUGH"
)

// Consistent Hash-based load balancing can be used to provide soft
// session affinity based on HTTP headers, cookies or other
// properties. This load balancing policy is applicable only for HTTP
// connections. The affinity to a particular destination host will be
// lost when one or more hosts are added/removed from the destination
// service.
type ConsistentHashLB struct {

	// It is required to specify exactly one of the fields as hash key:
	// HttpHeaderName, HttpCookie, or UseSourceIP.
	// Hash based on a specific HTTP header.
	HttpHeaderName string `json:"httpHeaderName,omitempty"`

	// Hash based on HTTP cookie.
	HttpCookie *HTTPCookie `json:"httpCookie,omitempty"`

	// Hash based on the source IP address.
	UseSourceIp bool `json:"useSourceIp,omitempty"`

	// The minimum number of virtual nodes to use for the hash
	// ring. Defaults to 1024. Larger ring sizes result in more granular
	// load distributions. If the number of hosts in the load balancing
	// pool is larger than the ring size, each host will be assigned a
	// single virtual node.
	MinimumRingSize uint64 `json:"minimumRingSize,omitempty"`
}

// Describes a HTTP cookie that will be used as the hash key for the
// Consistent Hash load balancer. If the cookie is not present, it will
// be generated.
type HTTPCookie struct {
	// REQUIRED. Name of the cookie.
	Name string `json:"name"`

	// Path to set for the cookie.
	Path string `json:"path,omitempty"`

	// REQUIRED. Lifetime of the cookie.
	Ttl string `json:"ttl"`
}

// Connection pool settings for an upstream host. The settings apply to
// each individual host in the upstream service.  See Envoy's [circuit
// breaker](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/circuit_breaking)
// for more details. Connection pool settings can be applied at the TCP
// level as well as at HTTP level.
//
// For example, the following rule sets a limit of 100 connections to redis
// service called myredissrv with a connect timeout of 30ms
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: bookinfo-redis
// spec:
//   host: myredissrv.prod.svc.cluster.local
//   trafficPolicy:
//     connectionPool:
//       tcp:
//         maxConnections: 100
//         connectTimeout: 30ms
type ConnectionPoolSettings struct {

	// Settings common to both HTTP and TCP upstream connections.
	Tcp *TCPSettings `json:"tcp,omitempty"`

	// HTTP connection pool settings.
	Http *HTTPSettings `json:"http,omitempty"`
}

// Settings common to both HTTP and TCP upstream connections.
type TCPSettings struct {
	// Maximum number of HTTP1 /TCP connections to a destination host.
	MaxConnections int32 `json:"maxConnections,omitempty"`

	// TCP connection timeout.
	ConnectTimeout string `json:"connectTimeout,omitempty"`
}

// Settings applicable to HTTP1.1/HTTP2/GRPC connections.
type HTTPSettings struct {
	// Maximum number of pending HTTP requests to a destination. Default 1024.
	Http1MaxPendingRequests int32 `json:"http1MaxPendingRequests,omitempty"`

	// Maximum number of requests to a backend. Default 1024.
	Http2MaxRequests int32 `json:"http2MaxRequests,omitempty"`

	// Maximum number of requests per connection to a backend. Setting this
	// parameter to 1 disables keep alive.
	MaxRequestsPerConnection int32 `json:"maxRequestsPerConnection,omitempty"`

	// Maximum number of retries that can be outstanding to all hosts in a
	// cluster at a given time. Defaults to 3.
	MaxRetries int32 `json:"maxRetries,omitempty"`
}

// A Circuit breaker implementation that tracks the status of each
// individual host in the upstream service.  Applicable to both HTTP and
// TCP services.  For HTTP services, hosts that continually return 5xx
// errors for API calls are ejected from the pool for a pre-defined period
// of time. For TCP services, connection timeouts or connection
// failures to a given host counts as an error when measuring the
// consecutive errors metric. See Envoy's [outlier
// detection](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/outlier)
// for more details.
//
// The following rule sets a connection pool size of 100 connections and
// 1000 concurrent HTTP2 requests, with no more than 10 req/connection to
// "reviews" service. In addition, it configures upstream hosts to be
// scanned every 5 mins, such that any host that fails 7 consecutive times
// with 5XX error code will be ejected for 15 minutes.
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: reviews-cb-policy
// spec:
//   host: reviews.prod.svc.cluster.local
//   trafficPolicy:
//     connectionPool:
//       tcp:
//         maxConnections: 100
//       http:
//         http2MaxRequests: 1000
//         maxRequestsPerConnection: 10
//     outlierDetection:
//       consecutiveErrors: 7
//       interval: 5m
//       baseEjectionTime: 15m
type OutlierDetection struct {
	// Number of errors before a host is ejected from the connection
	// pool. Defaults to 5. When the upstream host is accessed over HTTP, a
	// 5xx return code qualifies as an error. When the upstream host is
	// accessed over an opaque TCP connection, connect timeouts and
	// connection error/failure events qualify as an error.
	ConsecutiveErrors int32 `json:"consecutiveErrors,omitempty"`

	// Time interval between ejection sweep analysis. format:
	// 1h/1m/1s/1ms. MUST BE >=1ms. Default is 10s.
	Interval string `json:"interval,omitempty"`

	// Minimum ejection duration. A host will remain ejected for a period
	// equal to the product of minimum ejection duration and the number of
	// times the host has been ejected. This technique allows the system to
	// automatically increase the ejection period for unhealthy upstream
	// servers. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 30s.
	BaseEjectionTime string `json:"baseEjectionTime,omitempty"`

	// Maximum % of hosts in the load balancing pool for the upstream
	// service that can be ejected. Defaults to 10%.
	MaxEjectionPercent int32 `json:"maxEjectionPercent,omitempty"`
}

// SSL/TLS related settings for upstream connections. See Envoy's [TLS
// context](https://www.envoyproxy.io/docs/envoy/latest/api-v1/cluster_manager/cluster_ssl.html#config-cluster-manager-cluster-ssl)
// for more details. These settings are common to both HTTP and TCP upstreams.
//
// For example, the following rule configures a client to use mutual TLS
// for connections to upstream database cluster.
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: db-mtls
// spec:
//   host: mydbserver.prod.svc.cluster.local
//   trafficPolicy:
//     tls:
//       mode: MUTUAL
//       clientCertificate: /etc/certs/myclientcert.pem
//       privateKey: /etc/certs/client_private_key.pem
//       caCertificates: /etc/certs/rootcacerts.pem
//
// The following rule configures a client to use TLS when talking to a
// foreign service whose domain matches *.foo.com.
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: tls-foo
// spec:
//   host: "*.foo.com"
//   trafficPolicy:
//     tls:
//       mode: SIMPLE
//
// The following rule configures a client to use Istio mutual TLS when talking
// to rating services.
//
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: ratings-istio-mtls
// spec:
//   host: ratings.prod.svc.cluster.local
//   trafficPolicy:
//     tls:
//       mode: ISTIO_MUTUAL
type TLSSettings struct {

	// REQUIRED: Indicates whether connections to this port should be secured
	// using TLS. The value of this field determines how TLS is enforced.
	Mode TLSmode `json:"mode"`

	// REQUIRED if mode is `MUTUAL`. The path to the file holding the
	// client-side TLS certificate to use.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	ClientCertificate string `json:"clientCertificate,omitempty"`

	// REQUIRED if mode is `MUTUAL`. The path to the file holding the
	// client's private key.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	PrivateKey string `json:"privateKey,omitempty"`

	// OPTIONAL: The path to the file containing certificate authority
	// certificates to use in verifying a presented server certificate. If
	// omitted, the proxy will not verify the server's certificate.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	CaCertificates string `json:"caCertificates,omitempty"`

	// A list of alternate names to verify the subject identity in the
	// certificate. If specified, the proxy will verify that the server
	// certificate's subject alt name matches one of the specified values.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	SubjectAltNames []string `json:"subjectAltNames,omitempty"`

	// SNI string to present to the server during TLS handshake.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	Sni string `json:"sni,omitempty"`
}

// TLS connection mode
type TLSmode string

const (
	// Do not setup a TLS connection to the upstream endpoint.
	TLSmodeDisable TLSmode = "DISABLE"

	// Originate a TLS connection to the upstream endpoint.
	TLSmodeSimple TLSmode = "SIMPLE"

	// Secure connections to the upstream using mutual TLS by presenting
	// client certificates for authentication.
	TLSmodeMutual TLSmode = "MUTUAL"

	// Secure connections to the upstream using mutual TLS by presenting
	// client certificates for authentication.
	// Compared to Mutual mode, this mode uses certificates generated
	// automatically by Istio for mTLS authentication. When this mode is
	// used, all other fields in `TLSSettings` should be empty.
	TLSmodeIstioMutual TLSmode = "ISTIO_MUTUAL"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DestinationRuleList is a list of DestinationRule resources
type DestinationRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []DestinationRule `json:"items"`
}
