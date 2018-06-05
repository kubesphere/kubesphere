/*
Copyright 2016 The Kubernetes Authors.

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

package v1alpha1

import (
	"net/url"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	kubeletscheme "k8s.io/kubernetes/pkg/kubelet/apis/kubeletconfig/scheme"
	kubeletconfigv1beta1 "k8s.io/kubernetes/pkg/kubelet/apis/kubeletconfig/v1beta1"
	kubeproxyscheme "k8s.io/kubernetes/pkg/proxy/apis/kubeproxyconfig/scheme"
	kubeproxyconfigv1alpha1 "k8s.io/kubernetes/pkg/proxy/apis/kubeproxyconfig/v1alpha1"
)

const (
	// DefaultServiceDNSDomain defines default cluster-internal domain name for Services and Pods
	DefaultServiceDNSDomain = "cluster.local"
	// DefaultServicesSubnet defines default service subnet range
	DefaultServicesSubnet = "10.96.0.0/12"
	// DefaultClusterDNSIP defines default DNS IP
	DefaultClusterDNSIP = "10.96.0.10"
	// DefaultKubernetesVersion defines default kubernetes version
	DefaultKubernetesVersion = "stable-1.10"
	// DefaultAPIBindPort defines default API port
	DefaultAPIBindPort = 6443
	// DefaultAuthorizationModes defines default authorization modes
	DefaultAuthorizationModes = "Node,RBAC"
	// DefaultCertificatesDir defines default certificate directory
	DefaultCertificatesDir = "/etc/kubernetes/pki"
	// DefaultImageRepository defines default image registry
	DefaultImageRepository = "k8s.gcr.io"
	// DefaultManifestsDir defines default manifests directory
	DefaultManifestsDir = "/etc/kubernetes/manifests"
	// DefaultCRISocket defines the default cri socket
	DefaultCRISocket = "/var/run/dockershim.sock"
	// DefaultClusterName defines the default cluster name
	DefaultClusterName = "kubernetes"

	// DefaultEtcdDataDir defines default location of etcd where static pods will save data to
	DefaultEtcdDataDir = "/var/lib/etcd"
	// DefaultEtcdClusterSize defines the default cluster size when using the etcd-operator
	DefaultEtcdClusterSize = 3
	// DefaultEtcdOperatorVersion defines the default version of the etcd-operator to use
	DefaultEtcdOperatorVersion = "v0.6.0"
	// DefaultEtcdCertDir represents the directory where PKI assets are stored for self-hosted etcd
	DefaultEtcdCertDir = "/etc/kubernetes/pki/etcd"
	// DefaultEtcdClusterServiceName is the default name of the service backing the etcd cluster
	DefaultEtcdClusterServiceName = "etcd-cluster"
	// DefaultProxyBindAddressv4 is the default bind address when the advertise address is v4
	DefaultProxyBindAddressv4 = "0.0.0.0"
	// DefaultProxyBindAddressv6 is the default bind address when the advertise address is v6
	DefaultProxyBindAddressv6 = "::"
	// KubeproxyKubeConfigFileName defines the file name for the kube-proxy's KubeConfig file
	KubeproxyKubeConfigFileName = "/var/lib/kube-proxy/kubeconfig.conf"

	// DefaultDiscoveryTimeout specifies the default discovery timeout for kubeadm (used unless one is specified in the NodeConfiguration)
	DefaultDiscoveryTimeout = 5 * time.Minute
)

var (
	// DefaultAuditPolicyLogMaxAge is defined as a var so its address can be taken
	// It is the number of days to store audit logs
	DefaultAuditPolicyLogMaxAge = int32(2)
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_MasterConfiguration assigns default values to Master node
func SetDefaults_MasterConfiguration(obj *MasterConfiguration) {
	if obj.KubernetesVersion == "" {
		obj.KubernetesVersion = DefaultKubernetesVersion
	}

	if obj.API.BindPort == 0 {
		obj.API.BindPort = DefaultAPIBindPort
	}

	if obj.Networking.ServiceSubnet == "" {
		obj.Networking.ServiceSubnet = DefaultServicesSubnet
	}

	if obj.Networking.DNSDomain == "" {
		obj.Networking.DNSDomain = DefaultServiceDNSDomain
	}

	if len(obj.AuthorizationModes) == 0 {
		obj.AuthorizationModes = strings.Split(DefaultAuthorizationModes, ",")
	}

	if obj.CertificatesDir == "" {
		obj.CertificatesDir = DefaultCertificatesDir
	}

	if obj.TokenTTL == nil {
		obj.TokenTTL = &metav1.Duration{
			Duration: constants.DefaultTokenDuration,
		}
	}

	if obj.CRISocket == "" {
		obj.CRISocket = DefaultCRISocket
	}

	if len(obj.TokenUsages) == 0 {
		obj.TokenUsages = constants.DefaultTokenUsages
	}

	if len(obj.TokenGroups) == 0 {
		obj.TokenGroups = constants.DefaultTokenGroups
	}

	if obj.ImageRepository == "" {
		obj.ImageRepository = DefaultImageRepository
	}

	if obj.Etcd.DataDir == "" {
		obj.Etcd.DataDir = DefaultEtcdDataDir
	}

	if obj.ClusterName == "" {
		obj.ClusterName = DefaultClusterName
	}

	SetDefaultsEtcdSelfHosted(obj)
	if features.Enabled(obj.FeatureGates, features.DynamicKubeletConfig) {
		SetDefaults_KubeletConfiguration(obj)
	}
	SetDefaults_ProxyConfiguration(obj)
	SetDefaults_AuditPolicyConfiguration(obj)
}

// SetDefaults_ProxyConfiguration assigns default values for the Proxy
func SetDefaults_ProxyConfiguration(obj *MasterConfiguration) {
	if obj.KubeProxy.Config == nil {
		obj.KubeProxy.Config = &kubeproxyconfigv1alpha1.KubeProxyConfiguration{}
	}
	if obj.KubeProxy.Config.ClusterCIDR == "" && obj.Networking.PodSubnet != "" {
		obj.KubeProxy.Config.ClusterCIDR = obj.Networking.PodSubnet
	}

	if obj.KubeProxy.Config.ClientConnection.KubeConfigFile == "" {
		obj.KubeProxy.Config.ClientConnection.KubeConfigFile = KubeproxyKubeConfigFileName
	}

	kubeproxyscheme.Scheme.Default(obj.KubeProxy.Config)
}

// SetDefaults_NodeConfiguration assigns default values to a regular node
func SetDefaults_NodeConfiguration(obj *NodeConfiguration) {
	if obj.CACertPath == "" {
		obj.CACertPath = DefaultCACertPath
	}
	if len(obj.TLSBootstrapToken) == 0 {
		obj.TLSBootstrapToken = obj.Token
	}
	if len(obj.DiscoveryToken) == 0 && len(obj.DiscoveryFile) == 0 {
		obj.DiscoveryToken = obj.Token
	}
	if obj.CRISocket == "" {
		obj.CRISocket = DefaultCRISocket
	}
	// Make sure file URLs become paths
	if len(obj.DiscoveryFile) != 0 {
		u, err := url.Parse(obj.DiscoveryFile)
		if err == nil && u.Scheme == "file" {
			obj.DiscoveryFile = u.Path
		}
	}
	if obj.DiscoveryTimeout == nil {
		obj.DiscoveryTimeout = &metav1.Duration{
			Duration: DefaultDiscoveryTimeout,
		}
	}
	if obj.ClusterName == "" {
		obj.ClusterName = DefaultClusterName
	}
}

// SetDefaultsEtcdSelfHosted sets defaults for self-hosted etcd if used
func SetDefaultsEtcdSelfHosted(obj *MasterConfiguration) {
	if obj.Etcd.SelfHosted != nil {
		if obj.Etcd.SelfHosted.ClusterServiceName == "" {
			obj.Etcd.SelfHosted.ClusterServiceName = DefaultEtcdClusterServiceName
		}

		if obj.Etcd.SelfHosted.EtcdVersion == "" {
			obj.Etcd.SelfHosted.EtcdVersion = constants.DefaultEtcdVersion
		}

		if obj.Etcd.SelfHosted.OperatorVersion == "" {
			obj.Etcd.SelfHosted.OperatorVersion = DefaultEtcdOperatorVersion
		}

		if obj.Etcd.SelfHosted.CertificatesDir == "" {
			obj.Etcd.SelfHosted.CertificatesDir = DefaultEtcdCertDir
		}
	}
}

// SetDefaults_KubeletConfiguration assigns default values to kubelet
func SetDefaults_KubeletConfiguration(obj *MasterConfiguration) {
	if obj.KubeletConfiguration.BaseConfig == nil {
		obj.KubeletConfiguration.BaseConfig = &kubeletconfigv1beta1.KubeletConfiguration{}
	}
	if obj.KubeletConfiguration.BaseConfig.StaticPodPath == "" {
		obj.KubeletConfiguration.BaseConfig.StaticPodPath = DefaultManifestsDir
	}
	if obj.KubeletConfiguration.BaseConfig.ClusterDNS == nil {
		dnsIP, err := constants.GetDNSIP(obj.Networking.ServiceSubnet)
		if err != nil {
			obj.KubeletConfiguration.BaseConfig.ClusterDNS = []string{DefaultClusterDNSIP}
		} else {
			obj.KubeletConfiguration.BaseConfig.ClusterDNS = []string{dnsIP.String()}
		}
	}
	if obj.KubeletConfiguration.BaseConfig.ClusterDomain == "" {
		obj.KubeletConfiguration.BaseConfig.ClusterDomain = DefaultServiceDNSDomain
	}
	if obj.KubeletConfiguration.BaseConfig.Authorization.Mode == "" {
		obj.KubeletConfiguration.BaseConfig.Authorization.Mode = kubeletconfigv1beta1.KubeletAuthorizationModeWebhook
	}
	if obj.KubeletConfiguration.BaseConfig.Authentication.X509.ClientCAFile == "" {
		obj.KubeletConfiguration.BaseConfig.Authentication.X509.ClientCAFile = DefaultCACertPath
	}

	scheme, _, _ := kubeletscheme.NewSchemeAndCodecs()
	if scheme != nil {
		scheme.Default(obj.KubeletConfiguration.BaseConfig)
	}
}

// SetDefaults_AuditPolicyConfiguration sets default values for the AuditPolicyConfiguration
func SetDefaults_AuditPolicyConfiguration(obj *MasterConfiguration) {
	if obj.AuditPolicyConfiguration.LogDir == "" {
		obj.AuditPolicyConfiguration.LogDir = constants.StaticPodAuditPolicyLogDir
	}
	if obj.AuditPolicyConfiguration.LogMaxAge == nil {
		obj.AuditPolicyConfiguration.LogMaxAge = &DefaultAuditPolicyLogMaxAge
	}
}
