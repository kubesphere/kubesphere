// Copyright 2020 The prometheus-operator Authors
//
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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ThanosRulerKind    = "ThanosRuler"
	ThanosRulerName    = "thanosrulers"
	ThanosRulerKindKey = "thanosrulers"
)

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="ruler"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Paused",type="boolean",JSONPath=".status.paused",description="Whether the resource reconciliation is paused or not",priority=1

// ThanosRuler defines a ThanosRuler deployment.
type ThanosRuler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the ThanosRuler cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec ThanosRulerSpec `json:"spec"`
	// Most recent observed status of the ThanosRuler cluster. Read-only. Not
	// included when requesting from the apiserver, only from the ThanosRuler
	// Operator API itself. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status *ThanosRulerStatus `json:"status,omitempty"`
}

// ThanosRulerList is a list of ThanosRulers.
// +k8s:openapi-gen=true
type ThanosRulerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Prometheuses
	Items []*ThanosRuler `json:"items"`
}

// ThanosRulerSpec is a specification of the desired behavior of the ThanosRuler. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type ThanosRulerSpec struct {
	// Version of Thanos to be deployed.
	Version string `json:"version,omitempty"`
	// PodMetadata contains Labels and Annotations gets propagated to the thanos ruler pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`
	// Thanos container image URL.
	Image string `json:"image,omitempty"`
	// Image pull policy for the 'thanos', 'init-config-reloader' and 'config-reloader' containers.
	// See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details.
	// +kubebuilder:validation:Enum="";Always;Never;IfNotPresent
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling thanos images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// When a ThanosRuler deployment is paused, no actions except for deletion
	// will be performed on the underlying objects.
	Paused bool `json:"paused,omitempty"`
	// Number of thanos ruler instances to deploy.
	Replicas *int32 `json:"replicas,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Resources defines the resource requirements for single Pods.
	// If not provided, no requests/limits will be set
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// If specified, the pod's scheduling constraints.
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// If specified, the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// If specified, the pod's topology spread constraints.
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// Priority class assigned to the Pods
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// Thanos Ruler Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Storage spec to specify how storage shall be used.
	Storage *StorageSpec `json:"storage,omitempty"`
	// Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
	// be appended to other volumes that are generated as a result of StorageSpec objects.
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// ObjectStorageConfig configures object storage in Thanos.
	// Alternative to ObjectStorageConfigFile, and lower order priority.
	ObjectStorageConfig *v1.SecretKeySelector `json:"objectStorageConfig,omitempty"`
	// ObjectStorageConfigFile specifies the path of the object storage configuration file.
	// When used alongside with ObjectStorageConfig, ObjectStorageConfigFile takes precedence.
	ObjectStorageConfigFile *string `json:"objectStorageConfigFile,omitempty"`
	// ListenLocal makes the Thanos ruler listen on loopback, so that it
	// does not bind against the Pod IP.
	ListenLocal bool `json:"listenLocal,omitempty"`
	// QueryEndpoints defines Thanos querier endpoints from which to query metrics.
	// Maps to the --query flag of thanos ruler.
	QueryEndpoints []string `json:"queryEndpoints,omitempty"`
	// Define configuration for connecting to thanos query instances.
	// If this is defined, the QueryEndpoints field will be ignored.
	// Maps to the `query.config` CLI argument.
	// Only available with thanos v0.11.0 and higher.
	QueryConfig *v1.SecretKeySelector `json:"queryConfig,omitempty"`
	// Define URLs to send alerts to Alertmanager.  For Thanos v0.10.0 and higher,
	// AlertManagersConfig should be used instead.  Note: this field will be ignored
	// if AlertManagersConfig is specified.
	// Maps to the `alertmanagers.url` arg.
	AlertManagersURL []string `json:"alertmanagersUrl,omitempty"`
	// Define configuration for connecting to alertmanager.  Only available with thanos v0.10.0
	// and higher.  Maps to the `alertmanagers.config` arg.
	AlertManagersConfig *v1.SecretKeySelector `json:"alertmanagersConfig,omitempty"`
	// A label selector to select which PrometheusRules to mount for alerting and
	// recording.
	RuleSelector *metav1.LabelSelector `json:"ruleSelector,omitempty"`
	// Namespaces to be selected for Rules discovery. If unspecified, only
	// the same namespace as the ThanosRuler object is in is used.
	RuleNamespaceSelector *metav1.LabelSelector `json:"ruleNamespaceSelector,omitempty"`
	// EnforcedNamespaceLabel enforces adding a namespace label of origin for each alert
	// and metric that is user created. The label value will always be the namespace of the object that is
	// being created.
	EnforcedNamespaceLabel string `json:"enforcedNamespaceLabel,omitempty"`
	// List of references to PrometheusRule objects
	// to be excluded from enforcing a namespace label of origin.
	// Applies only if enforcedNamespaceLabel set to true.
	ExcludedFromEnforcement []ObjectReference `json:"excludedFromEnforcement,omitempty"`
	// PrometheusRulesExcludedFromEnforce - list of Prometheus rules to be excluded from enforcing
	// of adding namespace labels. Works only if enforcedNamespaceLabel set to true.
	// Make sure both ruleNamespace and ruleName are set for each pair
	// Deprecated: use excludedFromEnforcement instead.
	PrometheusRulesExcludedFromEnforce []PrometheusRuleExcludeConfig `json:"prometheusRulesExcludedFromEnforce,omitempty"`
	// Log level for ThanosRuler to be configured with.
	//+kubebuilder:validation:Enum="";debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// Log format for ThanosRuler to be configured with.
	//+kubebuilder:validation:Enum="";logfmt;json
	LogFormat string `json:"logFormat,omitempty"`
	// Port name used for the pods and governing service.
	// This defaults to web
	PortName string `json:"portName,omitempty"`
	// Interval between consecutive evaluations.
	// +kubebuilder:default:="15s"
	EvaluationInterval Duration `json:"evaluationInterval,omitempty"`
	// Time duration ThanosRuler shall retain data for. Default is '24h',
	// and must match the regular expression `[0-9]+(ms|s|m|h|d|w|y)` (milliseconds seconds minutes hours days weeks years).
	// +kubebuilder:default:="24h"
	Retention Duration `json:"retention,omitempty"`
	// Containers allows injecting additional containers or modifying operator generated
	// containers. This can be used to allow adding an authentication proxy to a ThanosRuler pod or
	// to change the behavior of an operator generated container. Containers described here modify
	// an operator generated container if they share the same name and modifications are done via a
	// strategic merge patch. The current container names are: `thanos-ruler` and `config-reloader`.
	// Overriding containers is entirely outside the scope of what the maintainers will support and by doing
	// so, you accept that this behaviour may break at any time without notice.
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
	// fetch secrets for injection into the ThanosRuler configuration from external sources. Any
	// errors during the execution of an initContainer will lead to a restart of the Pod.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// Using initContainers for any use case other then secret fetching is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that this behaviour may break
	// at any time without notice.
	InitContainers []v1.Container `json:"initContainers,omitempty"`
	// TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way.
	TracingConfig *v1.SecretKeySelector `json:"tracingConfig,omitempty"`
	// TracingConfig specifies the path of the tracing configuration file.
	// When used alongside with TracingConfig, TracingConfigFile takes precedence.
	TracingConfigFile string `json:"tracingConfigFile,omitempty"`
	// Labels configure the external label pairs to ThanosRuler. A default replica label
	// `thanos_ruler_replica` will be always added  as a label with the value of the pod's name and it will be dropped in the alerts.
	Labels map[string]string `json:"labels,omitempty"`
	// AlertDropLabels configure the label names which should be dropped in ThanosRuler alerts.
	// The replica label `thanos_ruler_replica` will always be dropped in alerts.
	AlertDropLabels []string `json:"alertDropLabels,omitempty"`
	// The external URL the Thanos Ruler instances will be available under. This is
	// necessary to generate correct URLs. This is necessary if Thanos Ruler is not
	// served from root of a DNS name.
	ExternalPrefix string `json:"externalPrefix,omitempty"`
	// The route prefix ThanosRuler registers HTTP handlers for. This allows thanos UI to be served on a sub-path.
	RoutePrefix string `json:"routePrefix,omitempty"`
	// GRPCServerTLSConfig configures the gRPC server from which Thanos Querier reads
	// recorded rule data.
	// Note: Currently only the CAFile, CertFile, and KeyFile fields are supported.
	// Maps to the '--grpc-server-tls-*' CLI args.
	GRPCServerTLSConfig *TLSConfig `json:"grpcServerTlsConfig,omitempty"`
	// The external Query URL the Thanos Ruler will set in the 'Source' field
	// of all alerts.
	// Maps to the '--alert.query-url' CLI arg.
	AlertQueryURL string `json:"alertQueryUrl,omitempty"`
	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// This is an alpha field from kubernetes 1.22 until 1.24 which requires enabling the StatefulSetMinReadySeconds feature gate.
	// +optional
	MinReadySeconds *uint32 `json:"minReadySeconds,omitempty"`
	// AlertRelabelConfigs configures alert relabeling in ThanosRuler.
	// Alert relabel configurations must have the form as specified in the official Prometheus documentation:
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs
	// Alternative to AlertRelabelConfigFile, and lower order priority.
	AlertRelabelConfigs *v1.SecretKeySelector `json:"alertRelabelConfigs,omitempty"`
	// AlertRelabelConfigFile specifies the path of the alert relabeling configuration file.
	// When used alongside with AlertRelabelConfigs, alertRelabelConfigFile takes precedence.
	AlertRelabelConfigFile *string `json:"alertRelabelConfigFile,omitempty"`
	// Pods' hostAliases configuration
	// +listType=map
	// +listMapKey=ip
	HostAliases []HostAlias `json:"hostAliases,omitempty"`
	// AdditionalArgs allows setting additional arguments for the ThanosRuler container.
	// It is intended for e.g. activating hidden flags which are not supported by
	// the dedicated configuration options yet. The arguments are passed as-is to the
	// ThanosRuler container which may cause issues if they are invalid or not supported
	// by the given ThanosRuler version.
	// In case of an argument conflict (e.g. an argument which is already set by the
	// operator itself) or when providing an invalid argument the reconciliation will
	// fail and an error will be logged.
	AdditionalArgs []Argument `json:"additionalArgs,omitempty"`
}

// ThanosRulerStatus is the most recent observed status of the ThanosRuler. Read-only. Not
// included when requesting from the apiserver, only from the Prometheus
// Operator API itself. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type ThanosRulerStatus struct {
	// Represents whether any actions on the underlying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`
	// Total number of non-terminated pods targeted by this ThanosRuler deployment
	// (their labels match the selector).
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this ThanosRuler deployment
	// that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this ThanosRuler deployment.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this ThanosRuler deployment.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ThanosRuler) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ThanosRulerList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
