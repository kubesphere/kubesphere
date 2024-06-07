/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package options

import (
	"time"

	corev1 "k8s.io/api/core/v1"

	"kubesphere.io/utils/helm"
	"kubesphere.io/utils/s3"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/models/composedapp"
	"kubesphere.io/kubesphere/pkg/models/terminal"
	"kubesphere.io/kubesphere/pkg/multicluster"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
)

type Options struct {
	KubernetesOptions     *k8s.Options
	AuthenticationOptions *authentication.Options
	MultiClusterOptions   *multicluster.Options
	TelemetryOptions      *TelemetryOptions
	TerminalOptions       *terminal.Options
	ComposedAppOptions    *composedapp.Options
	HelmExecutorOptions   *HelmExecutorOptions
	ExtensionOptions      *ExtensionOptions
	KubeSphereOptions     *KubeSphereOptions
	S3Options             *s3.Options
}

type HelmExecutorOptions struct {
	Image               string                `json:"image,omitempty" yaml:"image,omitempty" mapstructure:"image,omitempty"`
	Timeout             time.Duration         `json:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout,omitempty"`
	HistoryMax          uint                  `json:"historyMax,omitempty" yaml:"historyMax,omitempty" mapstructure:"historyMax,omitempty"`
	JobTTLAfterFinished time.Duration         `json:"jobTTLAfterFinished,omitempty" yaml:"jobTTLAfterFinished,omitempty" mapstructure:"jobTTLAfterFinished,omitempty"`
	Resources           *ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty" mapstructure:"resources,omitempty"`
}

type ResourceRequirements struct {
	Limits   map[corev1.ResourceName]string `json:"limits,omitempty" yaml:"limits,omitempty" mapstructure:"limits,omitempty"`
	Requests map[corev1.ResourceName]string `json:"requests,omitempty" yaml:"requests,omitempty" mapstructure:"requests,omitempty"`
}

func NewHelmExecutorOptions() *HelmExecutorOptions {
	return &HelmExecutorOptions{
		Image:               helm.DefaultKubectlImage,
		Timeout:             helm.MinimumTimeout,
		HistoryMax:          2,
		JobTTLAfterFinished: 0,
	}
}

type ExtensionIngressOptions struct {
	IngressClassName string `json:"ingressClassName,omitempty" yaml:"ingressClassName,omitempty" mapstructure:"ingressClassName,omitempty"`
	DomainSuffix     string `json:"domainSuffix,omitempty" yaml:"domainSuffix,omitempty" mapstructure:"domainSuffix,omitempty"`
	HTTPPort         uint   `json:"httpPort,omitempty" yaml:"httpPort,omitempty" mapstructure:"httpPort,omitempty"`
	HTTPSPort        uint   `json:"httpsPort,omitempty" yaml:"httpsPort,omitempty" mapstructure:"httpsPort,omitempty"`
}

type ExtensionOptions struct {
	ImageRegistry string                   `json:"imageRegistry,omitempty" yaml:"imageRegistry,omitempty" mapstructure:"imageRegistry,omitempty"`
	NodeSelector  map[string]string        `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty" mapstructure:"nodeSelector,omitempty"`
	Ingress       *ExtensionIngressOptions `json:"ingress,omitempty" yaml:"ingress,omitempty" mapstructure:"ingress,omitempty"`
}

func NewExtensionOptions() *ExtensionOptions {
	return &ExtensionOptions{}
}

type KubeSphereOptions struct {
	TLS bool `json:"tls,omitempty" yaml:"tls,omitempty" mapstructure:"tls,omitempty"`
}

func NewKubeSphereOptions() *KubeSphereOptions {
	return &KubeSphereOptions{
		TLS: false,
	}
}

// TelemetryOptions is the config data for telemetry.
type TelemetryOptions struct {
	// KSCloudURL for kubesphere cloud
	KSCloudURL string `json:"ksCloudURL,omitempty" yaml:"ksCloudURL,omitempty" mapstructure:"ksCloudURL"`

	// collect period
	Period *time.Duration `json:"period,omitempty" yaml:"period,omitempty" mapstructure:"period"`
}

func NewTelemetryOptions() *TelemetryOptions {
	return &TelemetryOptions{}
}
