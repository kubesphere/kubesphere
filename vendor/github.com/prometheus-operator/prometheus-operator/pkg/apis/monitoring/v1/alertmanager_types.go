// Copyright 2018 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	AlertmanagersKind   = "Alertmanager"
	AlertmanagerName    = "alertmanagers"
	AlertManagerKindKey = "alertmanager"
)

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="am"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="The version of Alertmanager"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.availableReplicas",description="The number of ready replicas"
// +kubebuilder:printcolumn:name="Reconciled",type="string",JSONPath=".status.conditions[?(@.type == 'Reconciled')].status"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type == 'Available')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Paused",type="boolean",JSONPath=".status.paused",description="Whether the resource reconciliation is paused or not",priority=1
// +kubebuilder:subresource:status

// Alertmanager describes an Alertmanager cluster.
type Alertmanager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Alertmanager cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec AlertmanagerSpec `json:"spec"`
	// Most recent observed status of the Alertmanager cluster. Read-only.
	// More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status AlertmanagerStatus `json:"status,omitempty"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *Alertmanager) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// AlertmanagerSpec is a specification of the desired behavior of the Alertmanager cluster. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type AlertmanagerSpec struct {
	// PodMetadata configures Labels and Annotations which are propagated to the alertmanager pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`
	// Image if specified has precedence over baseImage, tag and sha
	// combinations. Specifying the version is still necessary to ensure the
	// Prometheus Operator knows what version of Alertmanager is being
	// configured.
	Image *string `json:"image,omitempty"`
	// Image pull policy for the 'alertmanager', 'init-config-reloader' and 'config-reloader' containers.
	// See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details.
	// +kubebuilder:validation:Enum="";Always;Never;IfNotPresent
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Version the cluster should be on.
	Version string `json:"version,omitempty"`
	// Tag of Alertmanager container image to be deployed. Defaults to the value of `version`.
	// Version is ignored if Tag is set.
	// Deprecated: use 'image' instead.  The image tag can be specified
	// as part of the image URL.
	Tag string `json:"tag,omitempty"`
	// SHA of Alertmanager container image to be deployed. Defaults to the value of `version`.
	// Similar to a tag, but the SHA explicitly deploys an immutable container image.
	// Version and Tag are ignored if SHA is set.
	// Deprecated: use 'image' instead.  The image digest can be specified
	// as part of the image URL.
	SHA string `json:"sha,omitempty"`
	// Base image that is used to deploy pods, without tag.
	// Deprecated: use 'image' instead
	BaseImage string `json:"baseImage,omitempty"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling prometheus and alertmanager images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Secrets is a list of Secrets in the same namespace as the Alertmanager
	// object, which shall be mounted into the Alertmanager Pods.
	// Each Secret is added to the StatefulSet definition as a volume named `secret-<secret-name>`.
	// The Secrets are mounted into `/etc/alertmanager/secrets/<secret-name>` in the 'alertmanager' container.
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a list of ConfigMaps in the same namespace as the Alertmanager
	// object, which shall be mounted into the Alertmanager Pods.
	// Each ConfigMap is added to the StatefulSet definition as a volume named `configmap-<configmap-name>`.
	// The ConfigMaps are mounted into `/etc/alertmanager/configmaps/<configmap-name>` in the 'alertmanager' container.
	ConfigMaps []string `json:"configMaps,omitempty"`
	// ConfigSecret is the name of a Kubernetes Secret in the same namespace as the
	// Alertmanager object, which contains the configuration for this Alertmanager
	// instance. If empty, it defaults to `alertmanager-<alertmanager-name>`.
	//
	// The Alertmanager configuration should be available under the
	// `alertmanager.yaml` key. Additional keys from the original secret are
	// copied to the generated secret and mounted into the
	// `/etc/alertmanager/config` directory in the `alertmanager` container.
	//
	// If either the secret or the `alertmanager.yaml` key is missing, the
	// operator provisions a minimal Alertmanager configuration with one empty
	// receiver (effectively dropping alert notifications).
	ConfigSecret string `json:"configSecret,omitempty"`
	// Log level for Alertmanager to be configured with.
	//+kubebuilder:validation:Enum="";debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// Log format for Alertmanager to be configured with.
	//+kubebuilder:validation:Enum="";logfmt;json
	LogFormat string `json:"logFormat,omitempty"`
	// Size is the expected size of the alertmanager cluster. The controller will
	// eventually make the size of the running cluster equal to the expected
	// size.
	Replicas *int32 `json:"replicas,omitempty"`
	// Time duration Alertmanager shall retain data for. Default is '120h',
	// and must match the regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes hours).
	// +kubebuilder:default:="120h"
	Retention GoDuration `json:"retention,omitempty"`
	// Storage is the definition of how storage will be used by the Alertmanager
	// instances.
	Storage *StorageSpec `json:"storage,omitempty"`
	// Volumes allows configuration of additional volumes on the output StatefulSet definition.
	// Volumes specified will be appended to other volumes that are generated as a result of
	// StorageSpec objects.
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the alertmanager container,
	// that are generated as a result of StorageSpec objects.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// The external URL the Alertmanager instances will be available under. This is
	// necessary to generate correct URLs. This is necessary if Alertmanager is not
	// served from root of a DNS name.
	ExternalURL string `json:"externalUrl,omitempty"`
	// The route prefix Alertmanager registers HTTP handlers for. This is useful,
	// if using ExternalURL and a proxy is rewriting HTTP routes of a request,
	// and the actual ExternalURL is still true, but the server serves requests
	// under a different route prefix. For example for use with `kubectl proxy`.
	RoutePrefix string `json:"routePrefix,omitempty"`
	// If set to true all actions on the underlying managed objects are not
	// goint to be performed, except for delete actions.
	Paused bool `json:"paused,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Define resources requests and limits for single Pods.
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
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// Prometheus Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// ListenLocal makes the Alertmanager server listen on loopback, so that it
	// does not bind against the Pod IP. Note this is only for the Alertmanager
	// UI, not the gossip communication.
	ListenLocal bool `json:"listenLocal,omitempty"`
	// Containers allows injecting additional containers. This is meant to
	// allow adding an authentication proxy to an Alertmanager pod.
	// Containers described here modify an operator generated container if they
	// share the same name and modifications are done via a strategic merge
	// patch. The current container names are: `alertmanager` and
	// `config-reloader`. Overriding containers is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that
	// this behaviour may break at any time without notice.
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
	// fetch secrets for injection into the Alertmanager configuration from external sources. Any
	// errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// Using initContainers for any use case other then secret fetching is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that this behaviour may break
	// at any time without notice.
	InitContainers []v1.Container `json:"initContainers,omitempty"`
	// Priority class assigned to the Pods
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// AdditionalPeers allows injecting a set of additional Alertmanagers to peer with to form a highly available cluster.
	AdditionalPeers []string `json:"additionalPeers,omitempty"`
	// ClusterAdvertiseAddress is the explicit address to advertise in cluster.
	// Needs to be provided for non RFC1918 [1] (public) addresses.
	// [1] RFC1918: https://tools.ietf.org/html/rfc1918
	ClusterAdvertiseAddress string `json:"clusterAdvertiseAddress,omitempty"`
	// Interval between gossip attempts.
	ClusterGossipInterval GoDuration `json:"clusterGossipInterval,omitempty"`
	// Interval between pushpull attempts.
	ClusterPushpullInterval GoDuration `json:"clusterPushpullInterval,omitempty"`
	// Timeout for cluster peering.
	ClusterPeerTimeout GoDuration `json:"clusterPeerTimeout,omitempty"`
	// Port name used for the pods and governing service.
	// This defaults to web
	PortName string `json:"portName,omitempty"`
	// ForceEnableClusterMode ensures Alertmanager does not deactivate the cluster mode when running with a single replica.
	// Use case is e.g. spanning an Alertmanager cluster across Kubernetes clusters with a single replica in each.
	ForceEnableClusterMode bool `json:"forceEnableClusterMode,omitempty"`
	// AlertmanagerConfigs to be selected for to merge and configure Alertmanager with.
	AlertmanagerConfigSelector *metav1.LabelSelector `json:"alertmanagerConfigSelector,omitempty"`
	// The AlertmanagerConfigMatcherStrategy defines how AlertmanagerConfig objects match the alerts.
	// In the future more options may be added.
	AlertmanagerConfigMatcherStrategy AlertmanagerConfigMatcherStrategy `json:"alertmanagerConfigMatcherStrategy,omitempty"`
	// Namespaces to be selected for AlertmanagerConfig discovery. If nil, only
	// check own namespace.
	AlertmanagerConfigNamespaceSelector *metav1.LabelSelector `json:"alertmanagerConfigNamespaceSelector,omitempty"`
	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// This is an alpha field from kubernetes 1.22 until 1.24 which requires enabling the StatefulSetMinReadySeconds feature gate.
	// +optional
	MinReadySeconds *uint32 `json:"minReadySeconds,omitempty"`
	// Pods' hostAliases configuration
	// +listType=map
	// +listMapKey=ip
	HostAliases []HostAlias `json:"hostAliases,omitempty"`
	// Defines the web command line flags when starting Alertmanager.
	Web *AlertmanagerWebSpec `json:"web,omitempty"`
	// EXPERIMENTAL: alertmanagerConfiguration specifies the configuration of Alertmanager.
	// If defined, it takes precedence over the `configSecret` field.
	// This field may change in future releases.
	AlertmanagerConfiguration *AlertmanagerConfiguration `json:"alertmanagerConfiguration,omitempty"`
}

// AlertmanagerConfigMatcherStrategy defines the strategy used by AlertmanagerConfig objects to match alerts.
type AlertmanagerConfigMatcherStrategy struct {
	// If set to `OnNamespace`, the operator injects a label matcher matching the namespace of the AlertmanagerConfig object for all its routes and inhibition rules.
	// `None` will not add any additional matchers other than the ones specified in the AlertmanagerConfig.
	// Default is `OnNamespace`.
	// +kubebuilder:validation:Enum="OnNamespace";"None"
	// +kubebuilder:default:="OnNamespace"
	Type string `json:"type,omitempty"`
}

// AlertmanagerConfiguration defines the Alertmanager configuration.
// +k8s:openapi-gen=true
type AlertmanagerConfiguration struct {
	// The name of the AlertmanagerConfig resource which is used to generate the Alertmanager configuration.
	// It must be defined in the same namespace as the Alertmanager object.
	// The operator will not enforce a `namespace` label for routes and inhibition rules.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`
	// Defines the global parameters of the Alertmanager configuration.
	// +optional
	Global *AlertmanagerGlobalConfig `json:"global,omitempty"`
	// Custom notification templates.
	// +optional
	Templates []SecretOrConfigMap `json:"templates,omitempty"`
}

// AlertmanagerGlobalConfig configures parameters that are valid in all other configuration contexts.
// See https://prometheus.io/docs/alerting/latest/configuration/#configuration-file
type AlertmanagerGlobalConfig struct {
	// ResolveTimeout is the default value used by alertmanager if the alert does
	// not include EndsAt, after this time passes it can declare the alert as resolved if it has not been updated.
	// This has no impact on alerts from Prometheus, as they always include EndsAt.
	ResolveTimeout Duration `json:"resolveTimeout,omitempty"`

	// HTTP client configuration.
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// AlertmanagerStatus is the most recent observed status of the Alertmanager cluster. Read-only.
// More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type AlertmanagerStatus struct {
	// Represents whether any actions on the underlying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`
	// Total number of non-terminated pods targeted by this Alertmanager
	// object (their labels match the selector).
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this Alertmanager
	// object that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this Alertmanager cluster.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this Alertmanager object.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
	// The current state of the Alertmanager object.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
}

// AlertmanagerWebSpec defines the web command line flags when starting Alertmanager.
// +k8s:openapi-gen=true
type AlertmanagerWebSpec struct {
	WebConfigFileFields `json:",inline"`
}

// HTTPConfig defines a client HTTP configuration.
// See https://prometheus.io/docs/alerting/latest/configuration/#http_config
type HTTPConfig struct {
	// Authorization header configuration for the client.
	// This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.
	// +optional
	Authorization *SafeAuthorization `json:"authorization,omitempty"`
	// BasicAuth for the client.
	// This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// OAuth2 client credentials used to fetch a token for the targets.
	// +optional
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// The secret's key that contains the bearer token to be used by the client
	// for authentication.
	// The secret needs to be in the same namespace as the Alertmanager
	// object and accessible by the Prometheus Operator.
	// +optional
	BearerTokenSecret *v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`
	// TLS configuration for the client.
	// +optional
	TLSConfig *SafeTLSConfig `json:"tlsConfig,omitempty"`
	// Optional proxy URL.
	// +optional
	ProxyURL string `json:"proxyURL,omitempty"`
	// FollowRedirects specifies whether the client should follow HTTP 3xx redirects.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
}

// AlertmanagerList is a list of Alertmanagers.
// +k8s:openapi-gen=true
type AlertmanagerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Alertmanagers
	Items []Alertmanager `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}
