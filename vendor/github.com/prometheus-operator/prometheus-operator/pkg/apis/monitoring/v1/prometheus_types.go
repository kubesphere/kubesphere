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
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

const (
	PrometheusesKind  = "Prometheus"
	PrometheusName    = "prometheuses"
	PrometheusKindKey = "prometheus"
)

// CommonPrometheusFields are the options available to both the Prometheus server and agent.
// +k8s:deepcopy-gen=true
type CommonPrometheusFields struct {
	// PodMetadata configures Labels and Annotations which are propagated to the prometheus pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`
	// ServiceMonitors to be selected for target discovery.
	//
	// If `spec.serviceMonitorSelector`, `spec.podMonitorSelector` and
	// `spec.probeSelector` are null, the Prometheus configuration is unmanaged.
	// The Prometheus operator will ensure that the Prometheus configuration's
	// Secret exists, but it is the responsibility of the user to provide the raw
	// gzipped Prometheus configuration under the `prometheus.yaml.gz` key.
	// This behavior is deprecated and will be removed in the next major version
	// of the custom resource definition. It is recommended to use
	// `spec.additionalScrapeConfigs` instead.
	ServiceMonitorSelector *metav1.LabelSelector `json:"serviceMonitorSelector,omitempty"`
	// Namespace's labels to match for ServiceMonitor discovery. If nil, only
	// check own namespace.
	ServiceMonitorNamespaceSelector *metav1.LabelSelector `json:"serviceMonitorNamespaceSelector,omitempty"`
	// *Experimental* PodMonitors to be selected for target discovery.
	//
	// If `spec.serviceMonitorSelector`, `spec.podMonitorSelector` and
	// `spec.probeSelector` are null, the Prometheus configuration is unmanaged.
	// The Prometheus operator will ensure that the Prometheus configuration's
	// Secret exists, but it is the responsibility of the user to provide the raw
	// gzipped Prometheus configuration under the `prometheus.yaml.gz` key.
	// This behavior is deprecated and will be removed in the next major version
	// of the custom resource definition. It is recommended to use
	// `spec.additionalScrapeConfigs` instead.
	PodMonitorSelector *metav1.LabelSelector `json:"podMonitorSelector,omitempty"`
	// Namespace's labels to match for PodMonitor discovery. If nil, only
	// check own namespace.
	PodMonitorNamespaceSelector *metav1.LabelSelector `json:"podMonitorNamespaceSelector,omitempty"`
	// *Experimental* Probes to be selected for target discovery.
	//
	// If `spec.serviceMonitorSelector`, `spec.podMonitorSelector` and
	// `spec.probeSelector` are null, the Prometheus configuration is unmanaged.
	// The Prometheus operator will ensure that the Prometheus configuration's
	// Secret exists, but it is the responsibility of the user to provide the raw
	// gzipped Prometheus configuration under the `prometheus.yaml.gz` key.
	// This behavior is deprecated and will be removed in the next major version
	// of the custom resource definition. It is recommended to use
	// `spec.additionalScrapeConfigs` instead.
	ProbeSelector *metav1.LabelSelector `json:"probeSelector,omitempty"`
	// *Experimental* Namespaces to be selected for Probe discovery. If nil, only check own namespace.
	ProbeNamespaceSelector *metav1.LabelSelector `json:"probeNamespaceSelector,omitempty"`
	// Version of Prometheus to be deployed.
	Version string `json:"version,omitempty"`
	// When a Prometheus deployment is paused, no actions except for deletion
	// will be performed on the underlying objects.
	Paused bool `json:"paused,omitempty"`
	// Image if specified has precedence over baseImage, tag and sha
	// combinations. Specifying the version is still necessary to ensure the
	// Prometheus Operator knows what version of Prometheus is being
	// configured.
	Image *string `json:"image,omitempty"`
	// Image pull policy for the 'prometheus', 'init-config-reloader' and 'config-reloader' containers.
	// See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details.
	// +kubebuilder:validation:Enum="";Always;Never;IfNotPresent
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling prometheus and alertmanager images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Number of replicas of each shard to deploy for a Prometheus deployment.
	// Number of replicas multiplied by shards is the total number of Pods
	// created.
	Replicas *int32 `json:"replicas,omitempty"`
	// EXPERIMENTAL: Number of shards to distribute targets onto. Number of
	// replicas multiplied by shards is the total number of Pods created. Note
	// that scaling down shards will not reshard data onto remaining instances,
	// it must be manually moved. Increasing shards will not reshard data
	// either but it will continue to be available from the same instances. To
	// query globally use Thanos sidecar and Thanos querier or remote write
	// data to a central location. Sharding is done on the content of the
	// `__address__` target meta-label.
	Shards *int32 `json:"shards,omitempty"`
	// Name of Prometheus external label used to denote replica name.
	// Defaults to the value of `prometheus_replica`. External label will
	// _not_ be added when value is set to empty string (`""`).
	ReplicaExternalLabelName *string `json:"replicaExternalLabelName,omitempty"`
	// Name of Prometheus external label used to denote Prometheus instance
	// name. Defaults to the value of `prometheus`. External label will
	// _not_ be added when value is set to empty string (`""`).
	PrometheusExternalLabelName *string `json:"prometheusExternalLabelName,omitempty"`
	// Log level for Prometheus to be configured with.
	//+kubebuilder:validation:Enum="";debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// Log format for Prometheus to be configured with.
	//+kubebuilder:validation:Enum="";logfmt;json
	LogFormat string `json:"logFormat,omitempty"`
	// Interval between consecutive scrapes. Default: `30s`
	// +kubebuilder:default:="30s"
	ScrapeInterval Duration `json:"scrapeInterval,omitempty"`
	// Number of seconds to wait for target to respond before erroring.
	ScrapeTimeout Duration `json:"scrapeTimeout,omitempty"`
	// The labels to add to any time series or alerts when communicating with
	// external systems (federation, remote storage, Alertmanager).
	ExternalLabels map[string]string `json:"externalLabels,omitempty"`
	// Enable Prometheus to be used as a receiver for the Prometheus remote write protocol. Defaults to the value of `false`.
	// WARNING: This is not considered an efficient way of ingesting samples.
	// Use it with caution for specific low-volume use cases.
	// It is not suitable for replacing the ingestion via scraping and turning
	// Prometheus into a push-based metrics collection system.
	// For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver
	// Only valid in Prometheus versions 2.33.0 and newer.
	EnableRemoteWriteReceiver bool `json:"enableRemoteWriteReceiver,omitempty"`
	// Enable access to Prometheus disabled features. By default, no features are enabled.
	// Enabling disabled features is entirely outside the scope of what the maintainers will
	// support and by doing so, you accept that this behaviour may break at any
	// time without notice.
	// For more information see https://prometheus.io/docs/prometheus/latest/disabled_features/
	EnableFeatures []string `json:"enableFeatures,omitempty"`
	// The external URL the Prometheus instances will be available under. This is
	// necessary to generate correct URLs. This is necessary if Prometheus is not
	// served from root of a DNS name.
	ExternalURL string `json:"externalUrl,omitempty"`
	// The route prefix Prometheus registers HTTP handlers for. This is useful,
	// if using ExternalURL and a proxy is rewriting HTTP routes of a request,
	// and the actual ExternalURL is still true, but the server serves requests
	// under a different route prefix. For example for use with `kubectl proxy`.
	RoutePrefix string `json:"routePrefix,omitempty"`
	// Storage spec to specify how storage shall be used.
	Storage *StorageSpec `json:"storage,omitempty"`
	// Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will
	// be appended to other volumes that are generated as a result of StorageSpec objects.
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the prometheus container,
	// that are generated as a result of StorageSpec objects.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// Defines the web command line flags when starting Prometheus.
	Web *PrometheusWebSpec `json:"web,omitempty"`
	// Define resources requests and limits for single Pods.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// Prometheus Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Secrets is a list of Secrets in the same namespace as the Prometheus
	// object, which shall be mounted into the Prometheus Pods.
	// Each Secret is added to the StatefulSet definition as a volume named `secret-<secret-name>`.
	// The Secrets are mounted into /etc/prometheus/secrets/<secret-name> in the 'prometheus' container.
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a list of ConfigMaps in the same namespace as the Prometheus
	// object, which shall be mounted into the Prometheus Pods.
	// Each ConfigMap is added to the StatefulSet definition as a volume named `configmap-<configmap-name>`.
	// The ConfigMaps are mounted into /etc/prometheus/configmaps/<configmap-name> in the 'prometheus' container.
	ConfigMaps []string `json:"configMaps,omitempty"`
	// If specified, the pod's scheduling constraints.
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// If specified, the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// If specified, the pod's topology spread constraints.
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// remoteWrite is the list of remote write configurations.
	RemoteWrite []RemoteWriteSpec `json:"remoteWrite,omitempty"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// ListenLocal makes the Prometheus server listen on loopback, so that it
	// does not bind against the Pod IP.
	ListenLocal bool `json:"listenLocal,omitempty"`
	// Containers allows injecting additional containers or modifying operator
	// generated containers. This can be used to allow adding an authentication
	// proxy to a Prometheus pod or to change the behavior of an operator
	// generated container. Containers described here modify an operator
	// generated container if they share the same name and modifications are
	// done via a strategic merge patch. The current container names are:
	// `prometheus`, `config-reloader`, and `thanos-sidecar`. Overriding
	// containers is entirely outside the scope of what the maintainers will
	// support and by doing so, you accept that this behaviour may break at any
	// time without notice.
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
	// fetch secrets for injection into the Prometheus configuration from external sources. Any errors
	// during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// InitContainers described here modify an operator
	// generated init containers if they share the same name and modifications are
	// done via a strategic merge patch. The current init container name is:
	// `init-config-reloader`. Overriding init containers is entirely outside the
	// scope of what the maintainers will support and by doing so, you accept that
	// this behaviour may break at any time without notice.
	InitContainers []v1.Container `json:"initContainers,omitempty"`
	// AdditionalScrapeConfigs allows specifying a key of a Secret containing
	// additional Prometheus scrape configurations. Scrape configurations
	// specified are appended to the configurations generated by the Prometheus
	// Operator. Job configurations specified must have the form as specified
	// in the official Prometheus documentation:
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config.
	// As scrape configs are appended, the user is responsible to make sure it
	// is valid. Note that using this feature may expose the possibility to
	// break upgrades of Prometheus. It is advised to review Prometheus release
	// notes to ensure that no incompatible scrape configs are going to break
	// Prometheus after the upgrade.
	AdditionalScrapeConfigs *v1.SecretKeySelector `json:"additionalScrapeConfigs,omitempty"`
	// APIServerConfig allows specifying a host and auth methods to access apiserver.
	// If left empty, Prometheus is assumed to run inside of the cluster
	// and will discover API servers automatically and use the pod's CA certificate
	// and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.
	APIServerConfig *APIServerConfig `json:"apiserverConfig,omitempty"`
	// Priority class assigned to the Pods
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// Port name used for the pods and governing service.
	// This defaults to web
	PortName string `json:"portName,omitempty"`
	// ArbitraryFSAccessThroughSMs configures whether configuration
	// based on a service monitor can access arbitrary files on the file system
	// of the Prometheus container e.g. bearer token files.
	ArbitraryFSAccessThroughSMs ArbitraryFSAccessThroughSMsConfig `json:"arbitraryFSAccessThroughSMs,omitempty"`
	// When true, Prometheus resolves label conflicts by renaming the labels in
	// the scraped data to "exported_<label value>" for all targets created
	// from service and pod monitors.
	// Otherwise the HonorLabels field of the service or pod monitor applies.
	OverrideHonorLabels bool `json:"overrideHonorLabels,omitempty"`
	// When true, Prometheus ignores the timestamps for all the targets created
	// from service and pod monitors.
	// Otherwise the HonorTimestamps field of the service or pod monitor applies.
	OverrideHonorTimestamps bool `json:"overrideHonorTimestamps,omitempty"`
	// IgnoreNamespaceSelectors if set to true will ignore NamespaceSelector
	// settings from all PodMonitor, ServiceMonitor and Probe objects. They will
	// only discover endpoints within the namespace of the PodMonitor,
	// ServiceMonitor and Probe objects.
	// Defaults to false.
	IgnoreNamespaceSelectors bool `json:"ignoreNamespaceSelectors,omitempty"`
	// EnforcedNamespaceLabel If set, a label will be added to
	//
	// 1. all user-metrics (created by `ServiceMonitor`, `PodMonitor` and `Probe` objects) and
	// 2. in all `PrometheusRule` objects (except the ones excluded in `prometheusRulesExcludedFromEnforce`) to
	//    * alerting & recording rules and
	//    * the metrics used in their expressions (`expr`).
	//
	// Label name is this field's value.
	// Label value is the namespace of the created object (mentioned above).
	EnforcedNamespaceLabel string `json:"enforcedNamespaceLabel,omitempty"`
	// EnforcedSampleLimit defines global limit on number of scraped samples
	// that will be accepted. This overrides any SampleLimit set per
	// ServiceMonitor or/and PodMonitor. It is meant to be used by admins to
	// enforce the SampleLimit to keep overall number of samples/series under
	// the desired limit.
	// Note that if SampleLimit is lower that value will be taken instead.
	EnforcedSampleLimit *uint64 `json:"enforcedSampleLimit,omitempty"`
	// EnforcedTargetLimit defines a global limit on the number of scraped
	// targets.  This overrides any TargetLimit set per ServiceMonitor or/and
	// PodMonitor.  It is meant to be used by admins to enforce the TargetLimit
	// to keep the overall number of targets under the desired limit.
	// Note that if TargetLimit is lower, that value will be taken instead,
	// except if either value is zero, in which case the non-zero value will be
	// used.  If both values are zero, no limit is enforced.
	EnforcedTargetLimit *uint64 `json:"enforcedTargetLimit,omitempty"`
	// Per-scrape limit on number of labels that will be accepted for a sample. If
	// more than this number of labels are present post metric-relabeling, the
	// entire scrape will be treated as failed. 0 means no limit.
	// Only valid in Prometheus versions 2.27.0 and newer.
	EnforcedLabelLimit *uint64 `json:"enforcedLabelLimit,omitempty"`
	// Per-scrape limit on length of labels name that will be accepted for a sample.
	// If a label name is longer than this number post metric-relabeling, the entire
	// scrape will be treated as failed. 0 means no limit.
	// Only valid in Prometheus versions 2.27.0 and newer.
	EnforcedLabelNameLengthLimit *uint64 `json:"enforcedLabelNameLengthLimit,omitempty"`
	// Per-scrape limit on length of labels value that will be accepted for a sample.
	// If a label value is longer than this number post metric-relabeling, the
	// entire scrape will be treated as failed. 0 means no limit.
	// Only valid in Prometheus versions 2.27.0 and newer.
	EnforcedLabelValueLengthLimit *uint64 `json:"enforcedLabelValueLengthLimit,omitempty"`
	// EnforcedBodySizeLimit defines the maximum size of uncompressed response body
	// that will be accepted by Prometheus. Targets responding with a body larger than this many bytes
	// will cause the scrape to fail. Example: 100MB.
	// If defined, the limit will apply to all service/pod monitors and probes.
	// This is an experimental feature, this behaviour could
	// change or be removed in the future.
	// Only valid in Prometheus versions 2.28.0 and newer.
	EnforcedBodySizeLimit ByteSize `json:"enforcedBodySizeLimit,omitempty"`
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
	// AdditionalArgs allows setting additional arguments for the Prometheus container.
	// It is intended for e.g. activating hidden flags which are not supported by
	// the dedicated configuration options yet. The arguments are passed as-is to the
	// Prometheus container which may cause issues if they are invalid or not supported
	// by the given Prometheus version.
	// In case of an argument conflict (e.g. an argument which is already set by the
	// operator itself) or when providing an invalid argument the reconciliation will
	// fail and an error will be logged.
	AdditionalArgs []Argument `json:"additionalArgs,omitempty"`
	// Enable compression of the write-ahead log using Snappy. This flag is
	// only available in versions of Prometheus >= 2.11.0.
	WALCompression *bool `json:"walCompression,omitempty"`
	// List of references to PodMonitor, ServiceMonitor, Probe and PrometheusRule objects
	// to be excluded from enforcing a namespace label of origin.
	// Applies only if enforcedNamespaceLabel set to true.
	ExcludedFromEnforcement []ObjectReference `json:"excludedFromEnforcement,omitempty"`
	// Use the host's network namespace if true.
	// Make sure to understand the security implications if you want to enable it.
	// When hostNetwork is enabled, this will set dnsPolicy to ClusterFirstWithHostNet automatically.
	HostNetwork bool `json:"hostNetwork,omitempty"`
	// PodTargetLabels are added to all Pod/ServiceMonitors' podTargetLabels
	PodTargetLabels []string `json:"podTargetLabels,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="prom"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="The version of Prometheus"
// +kubebuilder:printcolumn:name="Desired",type="integer",JSONPath=".spec.replicas",description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.availableReplicas",description="The number of ready replicas"
// +kubebuilder:printcolumn:name="Reconciled",type="string",JSONPath=".status.conditions[?(@.type == 'Reconciled')].status"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type == 'Available')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Paused",type="boolean",JSONPath=".status.paused",description="Whether the resource reconciliation is paused or not",priority=1
// +kubebuilder:subresource:status

// Prometheus defines a Prometheus deployment.
type Prometheus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Prometheus cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec PrometheusSpec `json:"spec"`
	// Most recent observed status of the Prometheus cluster. Read-only.
	// More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status PrometheusStatus `json:"status,omitempty"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *Prometheus) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// PrometheusList is a list of Prometheuses.
// +k8s:openapi-gen=true
type PrometheusList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Prometheuses
	Items []*Prometheus `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *PrometheusList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// PrometheusSpec is a specification of the desired behavior of the Prometheus cluster. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type PrometheusSpec struct {
	CommonPrometheusFields `json:",inline"`
	// Base image to use for a Prometheus deployment.
	// Deprecated: use 'image' instead
	BaseImage string `json:"baseImage,omitempty"`
	// Tag of Prometheus container image to be deployed. Defaults to the value of `version`.
	// Version is ignored if Tag is set.
	// Deprecated: use 'image' instead.  The image tag can be specified
	// as part of the image URL.
	Tag string `json:"tag,omitempty"`
	// SHA of Prometheus container image to be deployed. Defaults to the value of `version`.
	// Similar to a tag, but the SHA explicitly deploys an immutable container image.
	// Version and Tag are ignored if SHA is set.
	// Deprecated: use 'image' instead.  The image digest can be specified
	// as part of the image URL.
	SHA string `json:"sha,omitempty"`
	// Time duration Prometheus shall retain data for. Default is '24h' if
	// retentionSize is not set, and must match the regular expression `[0-9]+(ms|s|m|h|d|w|y)`
	// (milliseconds seconds minutes hours days weeks years).
	Retention Duration `json:"retention,omitempty"`
	// Maximum amount of disk space used by blocks.
	RetentionSize ByteSize `json:"retentionSize,omitempty"`
	// Disable prometheus compaction.
	DisableCompaction bool `json:"disableCompaction,omitempty"`
	// /--rules.*/ command-line arguments.
	Rules Rules `json:"rules,omitempty"`
	// PrometheusRulesExcludedFromEnforce - list of prometheus rules to be excluded from enforcing
	// of adding namespace labels. Works only if enforcedNamespaceLabel set to true.
	// Make sure both ruleNamespace and ruleName are set for each pair.
	// Deprecated: use excludedFromEnforcement instead.
	PrometheusRulesExcludedFromEnforce []PrometheusRuleExcludeConfig `json:"prometheusRulesExcludedFromEnforce,omitempty"`
	// QuerySpec defines the query command line flags when starting Prometheus.
	Query *QuerySpec `json:"query,omitempty"`
	// A selector to select which PrometheusRules to mount for loading alerting/recording
	// rules from. Until (excluding) Prometheus Operator v0.24.0 Prometheus
	// Operator will migrate any legacy rule ConfigMaps to PrometheusRule custom
	// resources selected by RuleSelector. Make sure it does not match any config
	// maps that you do not want to be migrated.
	RuleSelector *metav1.LabelSelector `json:"ruleSelector,omitempty"`
	// Namespaces to be selected for PrometheusRules discovery. If unspecified, only
	// the same namespace as the Prometheus object is in is used.
	RuleNamespaceSelector *metav1.LabelSelector `json:"ruleNamespaceSelector,omitempty"`
	// Define details regarding alerting.
	Alerting *AlertingSpec `json:"alerting,omitempty"`
	// remoteRead is the list of remote read configurations.
	RemoteRead []RemoteReadSpec `json:"remoteRead,omitempty"`
	// AdditionalAlertRelabelConfigs allows specifying a key of a Secret containing
	// additional Prometheus alert relabel configurations. Alert relabel configurations
	// specified are appended to the configurations generated by the Prometheus
	// Operator. Alert relabel configurations specified must have the form as specified
	// in the official Prometheus documentation:
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alert_relabel_configs.
	// As alert relabel configs are appended, the user is responsible to make sure it
	// is valid. Note that using this feature may expose the possibility to
	// break upgrades of Prometheus. It is advised to review Prometheus release
	// notes to ensure that no incompatible alert relabel configs are going to break
	// Prometheus after the upgrade.
	AdditionalAlertRelabelConfigs *v1.SecretKeySelector `json:"additionalAlertRelabelConfigs,omitempty"`
	// AdditionalAlertManagerConfigs allows specifying a key of a Secret containing
	// additional Prometheus AlertManager configurations. AlertManager configurations
	// specified are appended to the configurations generated by the Prometheus
	// Operator. Job configurations specified must have the form as specified
	// in the official Prometheus documentation:
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config.
	// As AlertManager configs are appended, the user is responsible to make sure it
	// is valid. Note that using this feature may expose the possibility to
	// break upgrades of Prometheus. It is advised to review Prometheus release
	// notes to ensure that no incompatible AlertManager configs are going to break
	// Prometheus after the upgrade.
	AdditionalAlertManagerConfigs *v1.SecretKeySelector `json:"additionalAlertManagerConfigs,omitempty"`
	// Thanos configuration allows configuring various aspects of a Prometheus
	// server in a Thanos environment.
	//
	// This section is experimental, it may change significantly without
	// deprecation notice in any release.
	//
	// This is experimental and may change significantly without backward
	// compatibility in any release.
	Thanos *ThanosSpec `json:"thanos,omitempty"`
	// QueryLogFile specifies the file to which PromQL queries are logged.
	// If the filename has an empty path, e.g. 'query.log', prometheus-operator will mount the file into an
	// emptyDir volume at `/var/log/prometheus`. If a full path is provided, e.g. /var/log/prometheus/query.log, you must mount a volume
	// in the specified directory and it must be writable. This is because the prometheus container runs with a read-only root filesystem for security reasons.
	// Alternatively, the location can be set to a stdout location such as `/dev/stdout` to log
	// query information to the default Prometheus log stream.
	// This is only available in versions of Prometheus >= 2.16.0.
	// For more details, see the Prometheus docs (https://prometheus.io/docs/guides/query-log/)
	QueryLogFile string `json:"queryLogFile,omitempty"`
	// AllowOverlappingBlocks enables vertical compaction and vertical query merge in Prometheus.
	// This is still experimental in Prometheus so it may change in any upcoming release.
	AllowOverlappingBlocks bool `json:"allowOverlappingBlocks,omitempty"`
	// Exemplars related settings that are runtime reloadable.
	// It requires to enable the exemplar storage feature to be effective.
	Exemplars *Exemplars `json:"exemplars,omitempty"`
	// Interval between consecutive evaluations. Default: `30s`
	// +kubebuilder:default:="30s"
	EvaluationInterval Duration `json:"evaluationInterval,omitempty"`
	// Enable access to prometheus web admin API. Defaults to the value of `false`.
	// WARNING: Enabling the admin APIs enables mutating endpoints, to delete data,
	// shutdown Prometheus, and more. Enabling this should be done with care and the
	// user is advised to add additional authentication authorization via a proxy to
	// ensure only clients authorized to perform these actions can do so.
	// For more information see https://prometheus.io/docs/prometheus/latest/querying/api/#tsdb-admin-apis
	EnableAdminAPI bool `json:"enableAdminAPI,omitempty"`
	// Defines the runtime reloadable configuration of the timeseries database
	// (TSDB).
	TSDB TSDBSpec `json:"tsdb,omitempty"`
}

// PrometheusStatus is the most recent observed status of the Prometheus cluster.
// More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type PrometheusStatus struct {
	// Represents whether any actions on the underlying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`
	// Total number of non-terminated pods targeted by this Prometheus deployment
	// (their labels match the selector).
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this Prometheus deployment
	// that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this Prometheus deployment.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this Prometheus deployment.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
	// The current state of the Prometheus deployment.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
	// The list has one entry per shard. Each entry provides a summary of the shard status.
	// +listType=map
	// +listMapKey=shardID
	// +optional
	ShardStatuses []ShardStatus `json:"shardStatuses,omitempty"`
}

// AlertingSpec defines parameters for alerting configuration of Prometheus servers.
// +k8s:openapi-gen=true
type AlertingSpec struct {
	// AlertmanagerEndpoints Prometheus should fire alerts against.
	Alertmanagers []AlertmanagerEndpoints `json:"alertmanagers"`
}

// StorageSpec defines the configured storage for a group Prometheus servers.
// If no storage option is specified, then by default an [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) will be used.
// If multiple storage options are specified, priority will be given as follows: EmptyDir, Ephemeral, and lastly VolumeClaimTemplate.
// +k8s:openapi-gen=true
type StorageSpec struct {
	// Deprecated: subPath usage will be disabled by default in a future release, this option will become unnecessary.
	// DisableMountSubPath allows to remove any subPath usage in volume mounts.
	DisableMountSubPath bool `json:"disableMountSubPath,omitempty"`
	// EmptyDirVolumeSource to be used by the StatefulSet. If specified, used in place of any volumeClaimTemplate. More
	// info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir
	EmptyDir *v1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// EphemeralVolumeSource to be used by the StatefulSet.
	// This is a beta field in k8s 1.21, for lower versions, starting with k8s 1.19, it requires enabling the GenericEphemeralVolume feature gate.
	// More info: https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes
	Ephemeral *v1.EphemeralVolumeSource `json:"ephemeral,omitempty"`
	// A PVC spec to be used by the StatefulSet. The easiest way to use a volume that cannot be automatically provisioned
	// (for whatever reason) is to use a label selector alongside manually created PersistentVolumes.
	VolumeClaimTemplate EmbeddedPersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// QuerySpec defines the query command line flags when starting Prometheus.
// +k8s:openapi-gen=true
type QuerySpec struct {
	// The delta difference allowed for retrieving metrics during expression evaluations.
	LookbackDelta *string `json:"lookbackDelta,omitempty"`
	// Number of concurrent queries that can be run at once.
	// +kubebuilder:validation:Minimum:=1
	MaxConcurrency *int32 `json:"maxConcurrency,omitempty"`
	// Maximum number of samples a single query can load into memory. Note that queries will fail if they would load more samples than this into memory, so this also limits the number of samples a query can return.
	MaxSamples *int32 `json:"maxSamples,omitempty"`
	// Maximum time a query may take before being aborted.
	Timeout *Duration `json:"timeout,omitempty"`
}

// PrometheusWebSpec defines the web command line flags when starting Prometheus.
// +k8s:openapi-gen=true
type PrometheusWebSpec struct {
	WebConfigFileFields `json:",inline"`
	// The prometheus web page title
	PageTitle *string `json:"pageTitle,omitempty"`
	// Defines the maximum number of simultaneous connections
	// A zero value means that Prometheus doesn't accept any incoming connection.
	// +kubebuilder:validation:Minimum:=0
	MaxConnections *int32 `json:"maxConnections,omitempty"`
}

// ThanosSpec defines parameters for a Prometheus server within a Thanos deployment.
// +k8s:openapi-gen=true
type ThanosSpec struct {
	// Image if specified has precedence over baseImage, tag and sha
	// combinations. Specifying the version is still necessary to ensure the
	// Prometheus Operator knows what version of Thanos is being
	// configured.
	Image *string `json:"image,omitempty"`
	// Version describes the version of Thanos to use.
	Version *string `json:"version,omitempty"`
	// Tag of Thanos sidecar container image to be deployed. Defaults to the value of `version`.
	// Version is ignored if Tag is set.
	// Deprecated: use 'image' instead.  The image tag can be specified
	// as part of the image URL.
	Tag *string `json:"tag,omitempty"`
	// SHA of Thanos container image to be deployed. Defaults to the value of `version`.
	// Similar to a tag, but the SHA explicitly deploys an immutable container image.
	// Version and Tag are ignored if SHA is set.
	// Deprecated: use 'image' instead.  The image digest can be specified
	// as part of the image URL.
	SHA *string `json:"sha,omitempty"`
	// Thanos base image if other than default.
	// Deprecated: use 'image' instead
	BaseImage *string `json:"baseImage,omitempty"`
	// Resources defines the resource requirements for the Thanos sidecar.
	// If not provided, no requests/limits will be set
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// ObjectStorageConfig configures object storage in Thanos.
	// Alternative to ObjectStorageConfigFile, and lower order priority.
	ObjectStorageConfig *v1.SecretKeySelector `json:"objectStorageConfig,omitempty"`
	// ObjectStorageConfigFile specifies the path of the object storage configuration file.
	// When used alongside with ObjectStorageConfig, ObjectStorageConfigFile takes precedence.
	ObjectStorageConfigFile *string `json:"objectStorageConfigFile,omitempty"`
	// If true, the Thanos sidecar listens on the loopback interface
	// for the HTTP and gRPC endpoints.
	// It takes precedence over `grpcListenLocal` and `httpListenLocal`.
	// Deprecated: use `grpcListenLocal` and `httpListenLocal` instead.
	ListenLocal bool `json:"listenLocal,omitempty"`
	// If true, the Thanos sidecar listens on the loopback interface
	// for the gRPC endpoints.
	// It has no effect if `listenLocal` is true.
	GRPCListenLocal bool `json:"grpcListenLocal,omitempty"`
	// If true, the Thanos sidecar listens on the loopback interface
	// for the HTTP endpoints.
	// It has no effect if `listenLocal` is true.
	HTTPListenLocal bool `json:"httpListenLocal,omitempty"`
	// TracingConfig configures tracing in Thanos. This is an experimental feature, it may change in any upcoming release in a breaking way.
	TracingConfig *v1.SecretKeySelector `json:"tracingConfig,omitempty"`
	// TracingConfig specifies the path of the tracing configuration file.
	// When used alongside with TracingConfig, TracingConfigFile takes precedence.
	TracingConfigFile string `json:"tracingConfigFile,omitempty"`
	// GRPCServerTLSConfig configures the TLS parameters for the gRPC server
	// providing the StoreAPI.
	// Note: Currently only the CAFile, CertFile, and KeyFile fields are supported.
	// Maps to the '--grpc-server-tls-*' CLI args.
	GRPCServerTLSConfig *TLSConfig `json:"grpcServerTlsConfig,omitempty"`
	// LogLevel for Thanos sidecar to be configured with.
	//+kubebuilder:validation:Enum="";debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// LogFormat for Thanos sidecar to be configured with.
	//+kubebuilder:validation:Enum="";logfmt;json
	LogFormat string `json:"logFormat,omitempty"`
	// MinTime for Thanos sidecar to be configured with. Option can be a constant time in RFC3339 format or time duration relative to current time, such as -1d or 2h45m. Valid duration units are ms, s, m, h, d, w, y.
	MinTime string `json:"minTime,omitempty"`
	// ReadyTimeout is the maximum time Thanos sidecar will wait for Prometheus to start. Eg 10m
	ReadyTimeout Duration `json:"readyTimeout,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the thanos-sidecar container.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// AdditionalArgs allows setting additional arguments for the Thanos container.
	// The arguments are passed as-is to the Thanos container which may cause issues
	// if they are invalid or not supported the given Thanos version.
	// In case of an argument conflict (e.g. an argument which is already set by the
	// operator itself) or when providing an invalid argument the reconciliation will
	// fail and an error will be logged.
	AdditionalArgs []Argument `json:"additionalArgs,omitempty"`
}

// RemoteWriteSpec defines the configuration to write samples from Prometheus
// to a remote endpoint.
// +k8s:openapi-gen=true
type RemoteWriteSpec struct {
	// The URL of the endpoint to send samples to.
	URL string `json:"url"`
	// The name of the remote write queue, it must be unique if specified. The
	// name is used in metrics and logging in order to differentiate queues.
	// Only valid in Prometheus versions 2.15.0 and newer.
	Name string `json:"name,omitempty"`
	// Enables sending of exemplars over remote write. Note that
	// exemplar-storage itself must be enabled using the enableFeature option
	// for exemplars to be scraped in the first place.  Only valid in
	// Prometheus versions 2.27.0 and newer.
	SendExemplars *bool `json:"sendExemplars,omitempty"`
	// Timeout for requests to the remote write endpoint.
	RemoteTimeout Duration `json:"remoteTimeout,omitempty"`
	// Custom HTTP headers to be sent along with each remote write request.
	// Be aware that headers that are set by Prometheus itself can't be overwritten.
	// Only valid in Prometheus versions 2.25.0 and newer.
	Headers map[string]string `json:"headers,omitempty"`
	// The list of remote write relabel configurations.
	WriteRelabelConfigs []RelabelConfig `json:"writeRelabelConfigs,omitempty"`
	// OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// BasicAuth for the URL.
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// Bearer token for remote write.
	BearerToken string `json:"bearerToken,omitempty"`
	// File to read bearer token for remote write.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// Authorization section for remote write
	Authorization *Authorization `json:"authorization,omitempty"`
	// Sigv4 allows to configures AWS's Signature Verification 4
	Sigv4 *Sigv4 `json:"sigv4,omitempty"`
	// TLS Config to use for remote write.
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
	// Optional ProxyURL.
	ProxyURL string `json:"proxyUrl,omitempty"`
	// QueueConfig allows tuning of the remote write queue parameters.
	QueueConfig *QueueConfig `json:"queueConfig,omitempty"`
	// MetadataConfig configures the sending of series metadata to the remote storage.
	MetadataConfig *MetadataConfig `json:"metadataConfig,omitempty"`
}

// QueueConfig allows the tuning of remote write's queue_config parameters.
// This object is referenced in the RemoteWriteSpec object.
// +k8s:openapi-gen=true
type QueueConfig struct {
	// Capacity is the number of samples to buffer per shard before we start dropping them.
	Capacity int `json:"capacity,omitempty"`
	// MinShards is the minimum number of shards, i.e. amount of concurrency.
	MinShards int `json:"minShards,omitempty"`
	// MaxShards is the maximum number of shards, i.e. amount of concurrency.
	MaxShards int `json:"maxShards,omitempty"`
	// MaxSamplesPerSend is the maximum number of samples per send.
	MaxSamplesPerSend int `json:"maxSamplesPerSend,omitempty"`
	// BatchSendDeadline is the maximum time a sample will wait in buffer.
	BatchSendDeadline string `json:"batchSendDeadline,omitempty"`
	// MaxRetries is the maximum number of times to retry a batch on recoverable errors.
	MaxRetries int `json:"maxRetries,omitempty"`
	// MinBackoff is the initial retry delay. Gets doubled for every retry.
	MinBackoff string `json:"minBackoff,omitempty"`
	// MaxBackoff is the maximum retry delay.
	MaxBackoff string `json:"maxBackoff,omitempty"`
	// Retry upon receiving a 429 status code from the remote-write storage.
	// This is experimental feature and might change in the future.
	RetryOnRateLimit bool `json:"retryOnRateLimit,omitempty"`
}

// Sigv4 optionally configures AWS's Signature Verification 4 signing process to
// sign requests. Cannot be set at the same time as basic_auth or authorization.
// +k8s:openapi-gen=true
type Sigv4 struct {
	// Region is the AWS region. If blank, the region from the default credentials chain used.
	Region string `json:"region,omitempty"`
	// AccessKey is the AWS API key. If blank, the environment variable `AWS_ACCESS_KEY_ID` is used.
	AccessKey *v1.SecretKeySelector `json:"accessKey,omitempty"`
	// SecretKey is the AWS API secret. If blank, the environment variable `AWS_SECRET_ACCESS_KEY` is used.
	SecretKey *v1.SecretKeySelector `json:"secretKey,omitempty"`
	// Profile is the named AWS profile used to authenticate.
	Profile string `json:"profile,omitempty"`
	// RoleArn is the named AWS profile used to authenticate.
	RoleArn string `json:"roleArn,omitempty"`
}

// RemoteReadSpec defines the configuration for Prometheus to read back samples
// from a remote endpoint.
// +k8s:openapi-gen=true
type RemoteReadSpec struct {
	// The URL of the endpoint to query from.
	URL string `json:"url"`
	// The name of the remote read queue, it must be unique if specified. The name
	// is used in metrics and logging in order to differentiate read
	// configurations.  Only valid in Prometheus versions 2.15.0 and newer.
	Name string `json:"name,omitempty"`
	// An optional list of equality matchers which have to be present
	// in a selector to query the remote read endpoint.
	RequiredMatchers map[string]string `json:"requiredMatchers,omitempty"`
	// Timeout for requests to the remote read endpoint.
	RemoteTimeout Duration `json:"remoteTimeout,omitempty"`
	// Custom HTTP headers to be sent along with each remote read request.
	// Be aware that headers that are set by Prometheus itself can't be overwritten.
	// Only valid in Prometheus versions 2.26.0 and newer.
	Headers map[string]string `json:"headers,omitempty"`
	// Whether reads should be made for queries for time ranges that
	// the local storage should have complete data for.
	ReadRecent bool `json:"readRecent,omitempty"`
	// BasicAuth for the URL.
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// Bearer token for remote read.
	BearerToken string `json:"bearerToken,omitempty"`
	// File to read bearer token for remote read.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// Authorization section for remote read
	Authorization *Authorization `json:"authorization,omitempty"`
	// TLS Config to use for remote read.
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
	// Optional ProxyURL.
	ProxyURL string `json:"proxyUrl,omitempty"`
	// Whether to use the external labels as selectors for the remote read endpoint.
	// Requires Prometheus v2.34.0 and above.
	FilterExternalLabels *bool `json:"filterExternalLabels,omitempty"`
}

// RelabelConfig allows dynamic rewriting of the label set, being applied to samples before ingestion.
// It defines `<metric_relabel_configs>`-section of Prometheus configuration.
// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs
// +k8s:openapi-gen=true
type RelabelConfig struct {
	// The source labels select values from existing labels. Their content is concatenated
	// using the configured separator and matched against the configured regular expression
	// for the replace, keep, and drop actions.
	SourceLabels []LabelName `json:"sourceLabels,omitempty"`
	// Separator placed between concatenated source label values. default is ';'.
	Separator string `json:"separator,omitempty"`
	// Label to which the resulting value is written in a replace action.
	// It is mandatory for replace actions. Regex capture groups are available.
	TargetLabel string `json:"targetLabel,omitempty"`
	// Regular expression against which the extracted value is matched. Default is '(.*)'
	Regex string `json:"regex,omitempty"`
	// Modulus to take of the hash of the source label values.
	Modulus uint64 `json:"modulus,omitempty"`
	// Replacement value against which a regex replace is performed if the
	// regular expression matches. Regex capture groups are available. Default is '$1'
	Replacement string `json:"replacement,omitempty"`
	//Action to perform based on regex matching. Default is 'replace'.
	//uppercase and lowercase actions require Prometheus >= 2.36.
	//+kubebuilder:validation:Enum=replace;Replace;keep;Keep;drop;Drop;hashmod;HashMod;labelmap;LabelMap;labeldrop;LabelDrop;labelkeep;LabelKeep;lowercase;Lowercase;uppercase;Uppercase
	//+kubebuilder:default=replace
	Action string `json:"action,omitempty"`
}

// APIServerConfig defines a host and auth methods to access apiserver.
// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config
// +k8s:openapi-gen=true
type APIServerConfig struct {
	// Host of apiserver.
	// A valid string consisting of a hostname or IP followed by an optional port number
	Host string `json:"host"`
	// BasicAuth allow an endpoint to authenticate over basic authentication
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// Bearer token for accessing apiserver.
	BearerToken string `json:"bearerToken,omitempty"`
	// File to read bearer token for accessing apiserver.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// TLS Config to use for accessing apiserver.
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
	// Authorization section for accessing apiserver
	Authorization *Authorization `json:"authorization,omitempty"`
}

// AlertmanagerEndpoints defines a selection of a single Endpoints object
// containing alertmanager IPs to fire alerts against.
// +k8s:openapi-gen=true
type AlertmanagerEndpoints struct {
	// Namespace of Endpoints object.
	Namespace string `json:"namespace"`
	// Name of Endpoints object in Namespace.
	Name string `json:"name"`
	// Port the Alertmanager API is exposed on.
	Port intstr.IntOrString `json:"port"`
	// Scheme to use when firing alerts.
	Scheme string `json:"scheme,omitempty"`
	// Prefix for the HTTP path alerts are pushed to.
	PathPrefix string `json:"pathPrefix,omitempty"`
	// TLS Config to use for alertmanager connection.
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
	// BasicAuth allow an endpoint to authenticate over basic authentication
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// BearerTokenFile to read from filesystem to use when authenticating to
	// Alertmanager.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// Authorization section for this alertmanager endpoint
	Authorization *SafeAuthorization `json:"authorization,omitempty"`
	// Version of the Alertmanager API that Prometheus uses to send alerts. It
	// can be "v1" or "v2".
	APIVersion string `json:"apiVersion,omitempty"`
	// Timeout is a per-target Alertmanager timeout when pushing alerts.
	Timeout *Duration `json:"timeout,omitempty"`
	// Whether to enable HTTP2.
	EnableHttp2 *bool `json:"enableHttp2,omitempty"`
}

// /--rules.*/ command-line arguments
// +k8s:openapi-gen=true
type Rules struct {
	Alert RulesAlert `json:"alert,omitempty"`
}

// /--rules.alert.*/ command-line arguments
// +k8s:openapi-gen=true
type RulesAlert struct {
	// Max time to tolerate prometheus outage for restoring 'for' state of alert.
	ForOutageTolerance string `json:"forOutageTolerance,omitempty"`
	// Minimum duration between alert and restored 'for' state.
	// This is maintained only for alerts with configured 'for' time greater than grace period.
	ForGracePeriod string `json:"forGracePeriod,omitempty"`
	// Minimum amount of time to wait before resending an alert to Alertmanager.
	ResendDelay string `json:"resendDelay,omitempty"`
}

// MetadataConfig configures the sending of series metadata to the remote storage.
// +k8s:openapi-gen=true
type MetadataConfig struct {
	// Whether metric metadata is sent to the remote storage or not.
	Send bool `json:"send,omitempty"`
	// How frequently metric metadata is sent to the remote storage.
	SendInterval Duration `json:"sendInterval,omitempty"`
}

type ShardStatus struct {
	// Identifier of the shard.
	// +required
	ShardID string `json:"shardID"`
	// Total number of pods targeted by this shard.
	Replicas int32 `json:"replicas"`
	// Total number of non-terminated pods targeted by this shard
	// that have the desired spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this shard.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this shard.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}

type TSDBSpec struct {
	// Configures how old an out-of-order/out-of-bounds sample can be w.r.t.
	// the TSDB max time.
	// An out-of-order/out-of-bounds sample is ingested into the TSDB as long as
	// the timestamp of the sample is >= (TSDB.MaxTime - outOfOrderTimeWindow).
	// Out of order ingestion is an experimental feature and requires
	// Prometheus >= v2.39.0.
	OutOfOrderTimeWindow Duration `json:"outOfOrderTimeWindow,omitempty"`
}

type Exemplars struct {
	// Maximum number of exemplars stored in memory for all series.
	// If not set, Prometheus uses its default value.
	// A value of zero or less than zero disables the storage.
	MaxSize *int64 `json:"maxSize,omitempty"`
}

// SafeAuthorization specifies a subset of the Authorization struct, that is
// safe for use in Endpoints (no CredentialsFile field)
// +k8s:openapi-gen=true
type SafeAuthorization struct {
	// Set the authentication type. Defaults to Bearer, Basic will cause an
	// error
	Type string `json:"type,omitempty"`
	// The secret's key that contains the credentials of the request
	Credentials *v1.SecretKeySelector `json:"credentials,omitempty"`
}

// Validate semantically validates the given Authorization section.
func (c *SafeAuthorization) Validate() error {
	if c == nil {
		return nil
	}

	if strings.ToLower(strings.TrimSpace(c.Type)) == "basic" {
		return &AuthorizationValidationError{`Authorization type cannot be set to "basic", use "basic_auth" instead`}
	}
	if c.Credentials == nil {
		return &AuthorizationValidationError{"Authorization credentials are required"}
	}
	return nil
}

// Authorization contains optional `Authorization` header configuration.
// This section is only understood by versions of Prometheus >= 2.26.0.
type Authorization struct {
	SafeAuthorization `json:",inline"`
	// File to read a secret from, mutually exclusive with Credentials (from SafeAuthorization)
	CredentialsFile string `json:"credentialsFile,omitempty"`
}

// Validate semantically validates the given Authorization section.
func (c *Authorization) Validate() error {
	if c.Credentials != nil && c.CredentialsFile != "" {
		return &AuthorizationValidationError{"Authorization can not specify both Credentials and CredentialsFile"}
	}
	if strings.ToLower(strings.TrimSpace(c.Type)) == "basic" {
		return &AuthorizationValidationError{"Authorization type cannot be set to \"basic\", use \"basic_auth\" instead"}
	}
	return nil
}

// AuthorizationValidationError is returned by Authorization.Validate()
// on semantically invalid configurations.
// +k8s:openapi-gen=false
type AuthorizationValidationError struct {
	err string
}

func (e *AuthorizationValidationError) Error() string {
	return e.err
}
