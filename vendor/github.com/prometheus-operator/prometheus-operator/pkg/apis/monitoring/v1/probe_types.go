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
	ProbesKind   = "Probe"
	ProbeName    = "probes"
	ProbeKindKey = "probe"
)

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="prb"

// Probe defines monitoring for a set of static targets or ingresses.
type Probe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of desired Ingress selection for target discovery by Prometheus.
	Spec ProbeSpec `json:"spec"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *Probe) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// ProbeSpec contains specification parameters for a Probe.
// +k8s:openapi-gen=true
type ProbeSpec struct {
	// The job name assigned to scraped metrics by default.
	JobName string `json:"jobName,omitempty"`
	// Specification for the prober to use for probing targets.
	// The prober.URL parameter is required. Targets cannot be probed if left empty.
	ProberSpec ProberSpec `json:"prober,omitempty"`
	// The module to use for probing specifying how to probe the target.
	// Example module configuring in the blackbox exporter:
	// https://github.com/prometheus/blackbox_exporter/blob/master/example.yml
	Module string `json:"module,omitempty"`
	// Targets defines a set of static or dynamically discovered targets to probe.
	Targets ProbeTargets `json:"targets,omitempty"`
	// Interval at which targets are probed using the configured prober.
	// If not specified Prometheus' global scrape interval is used.
	Interval Duration `json:"interval,omitempty"`
	// Timeout for scraping metrics from the Prometheus exporter.
	// If not specified, the Prometheus global scrape interval is used.
	ScrapeTimeout Duration `json:"scrapeTimeout,omitempty"`
	// TLS configuration to use when scraping the endpoint.
	TLSConfig *ProbeTLSConfig `json:"tlsConfig,omitempty"`
	// Secret to mount to read bearer token for scraping targets. The secret
	// needs to be in the same namespace as the probe and accessible by
	// the Prometheus Operator.
	BearerTokenSecret v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`
	// BasicAuth allow an endpoint to authenticate over basic authentication.
	// More info: https://prometheus.io/docs/operating/configuration/#endpoint
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// OAuth2 for the URL. Only valid in Prometheus versions 2.27.0 and newer.
	OAuth2 *OAuth2 `json:"oauth2,omitempty"`
	// MetricRelabelConfigs to apply to samples before ingestion.
	MetricRelabelConfigs []*RelabelConfig `json:"metricRelabelings,omitempty"`
	// Authorization section for this endpoint
	Authorization *SafeAuthorization `json:"authorization,omitempty"`
	// SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
	SampleLimit uint64 `json:"sampleLimit,omitempty"`
	// TargetLimit defines a limit on the number of scraped targets that will be accepted.
	TargetLimit uint64 `json:"targetLimit,omitempty"`
	// Per-scrape limit on number of labels that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	LabelLimit uint64 `json:"labelLimit,omitempty"`
	// Per-scrape limit on length of labels name that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	LabelNameLengthLimit uint64 `json:"labelNameLengthLimit,omitempty"`
	// Per-scrape limit on length of labels value that will be accepted for a sample.
	// Only valid in Prometheus versions 2.27.0 and newer.
	LabelValueLengthLimit uint64 `json:"labelValueLengthLimit,omitempty"`
}

// ProbeTargets defines how to discover the probed targets.
// One of the `staticConfig` or `ingress` must be defined.
// If both are defined, `staticConfig` takes precedence.
// +k8s:openapi-gen=true
type ProbeTargets struct {
	// staticConfig defines the static list of targets to probe and the
	// relabeling configuration.
	// If `ingress` is also defined, `staticConfig` takes precedence.
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config.
	StaticConfig *ProbeTargetStaticConfig `json:"staticConfig,omitempty"`
	// ingress defines the Ingress objects to probe and the relabeling
	// configuration.
	// If `staticConfig` is also defined, `staticConfig` takes precedence.
	Ingress *ProbeTargetIngress `json:"ingress,omitempty"`
}

// Validate semantically validates the given ProbeTargets.
func (it *ProbeTargets) Validate() error {
	if it.StaticConfig == nil && it.Ingress == nil {
		return &ProbeTargetsValidationError{"at least one of .spec.targets.staticConfig and .spec.targets.ingress is required"}
	}

	return nil
}

// ProbeTargetsValidationError is returned by ProbeTargets.Validate()
// on semantically invalid configurations.
// +k8s:openapi-gen=false
type ProbeTargetsValidationError struct {
	err string
}

func (e *ProbeTargetsValidationError) Error() string {
	return e.err
}

// ProbeTargetStaticConfig defines the set of static targets considered for probing.
// +k8s:openapi-gen=true
type ProbeTargetStaticConfig struct {
	// The list of hosts to probe.
	Targets []string `json:"static,omitempty"`
	// Labels assigned to all metrics scraped from the targets.
	Labels map[string]string `json:"labels,omitempty"`
	// RelabelConfigs to apply to the label set of the targets before it gets
	// scraped.
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	RelabelConfigs []*RelabelConfig `json:"relabelingConfigs,omitempty"`
}

// ProbeTargetIngress defines the set of Ingress objects considered for probing.
// The operator configures a target for each host/path combination of each ingress object.
// +k8s:openapi-gen=true
type ProbeTargetIngress struct {
	// Selector to select the Ingress objects.
	Selector metav1.LabelSelector `json:"selector,omitempty"`
	// From which namespaces to select Ingress objects.
	NamespaceSelector NamespaceSelector `json:"namespaceSelector,omitempty"`
	// RelabelConfigs to apply to the label set of the target before it gets
	// scraped.
	// The original ingress address is available via the
	// `__tmp_prometheus_ingress_address` label. It can be used to customize the
	// probed URL.
	// The original scrape job's name is available via the `__tmp_prometheus_job_name` label.
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	RelabelConfigs []*RelabelConfig `json:"relabelingConfigs,omitempty"`
}

// ProberSpec contains specification parameters for the Prober used for probing.
// +k8s:openapi-gen=true
type ProberSpec struct {
	// Mandatory URL of the prober.
	URL string `json:"url"`
	// HTTP scheme to use for scraping.
	// Defaults to `http`.
	Scheme string `json:"scheme,omitempty"`
	// Path to collect metrics from.
	// Defaults to `/probe`.
	// +kubebuilder:default:="/probe"
	Path string `json:"path,omitempty"`
	// Optional ProxyURL.
	ProxyURL string `json:"proxyUrl,omitempty"`
}

// ProbeList is a list of Probes.
// +k8s:openapi-gen=true
type ProbeList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Probes
	Items []*Probe `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface.
func (l *ProbeList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// ProbeTLSConfig specifies TLS configuration parameters for the prober.
// +k8s:openapi-gen=true
type ProbeTLSConfig struct {
	SafeTLSConfig `json:",inline"`
}
