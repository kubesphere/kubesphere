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

package controlplane

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmapiext "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha1"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	certphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	staticpodutil "k8s.io/kubernetes/cmd/kubeadm/app/util/staticpod"
	authzmodes "k8s.io/kubernetes/pkg/kubeapiserver/authorizer/modes"
	"k8s.io/kubernetes/pkg/master/reconcilers"
	utilpointer "k8s.io/kubernetes/pkg/util/pointer"
	"k8s.io/kubernetes/pkg/util/version"
)

// Static pod definitions in golang form are included below so that `kubeadm init` can get going.
const (
	DefaultCloudConfigPath = "/etc/kubernetes/cloud-config"

	deprecatedV19AdmissionControl = "NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
	defaultV19AdmissionControl    = "NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
)

// CreateInitStaticPodManifestFiles will write all static pod manifest files needed to bring up the control plane.
func CreateInitStaticPodManifestFiles(manifestDir string, cfg *kubeadmapi.MasterConfiguration) error {
	glog.V(1).Infoln("[controlplane] creating static pod files")
	return createStaticPodFiles(manifestDir, cfg, kubeadmconstants.KubeAPIServer, kubeadmconstants.KubeControllerManager, kubeadmconstants.KubeScheduler)
}

// CreateAPIServerStaticPodManifestFile will write APIserver static pod manifest file.
func CreateAPIServerStaticPodManifestFile(manifestDir string, cfg *kubeadmapi.MasterConfiguration) error {
	glog.V(1).Infoln("creating APIserver static pod files")
	return createStaticPodFiles(manifestDir, cfg, kubeadmconstants.KubeAPIServer)
}

// CreateControllerManagerStaticPodManifestFile will write  controller manager static pod manifest file.
func CreateControllerManagerStaticPodManifestFile(manifestDir string, cfg *kubeadmapi.MasterConfiguration) error {
	glog.V(1).Infoln("creating controller manager static pod files")
	return createStaticPodFiles(manifestDir, cfg, kubeadmconstants.KubeControllerManager)
}

// CreateSchedulerStaticPodManifestFile will write scheduler static pod manifest file.
func CreateSchedulerStaticPodManifestFile(manifestDir string, cfg *kubeadmapi.MasterConfiguration) error {
	glog.V(1).Infoln("creating scheduler static pod files")
	return createStaticPodFiles(manifestDir, cfg, kubeadmconstants.KubeScheduler)
}

// GetStaticPodSpecs returns all staticPodSpecs actualized to the context of the current MasterConfiguration
// NB. this methods holds the information about how kubeadm creates static pod manifests.
func GetStaticPodSpecs(cfg *kubeadmapi.MasterConfiguration, k8sVersion *version.Version) map[string]v1.Pod {
	// Get the required hostpath mounts
	mounts := getHostPathVolumesForTheControlPlane(cfg)

	// Prepare static pod specs
	staticPodSpecs := map[string]v1.Pod{
		kubeadmconstants.KubeAPIServer: staticpodutil.ComponentPod(v1.Container{
			Name:            kubeadmconstants.KubeAPIServer,
			Image:           images.GetCoreImage(kubeadmconstants.KubeAPIServer, cfg.GetControlPlaneImageRepository(), cfg.KubernetesVersion, cfg.UnifiedControlPlaneImage),
			ImagePullPolicy: cfg.ImagePullPolicy,
			Command:         getAPIServerCommand(cfg),
			VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(kubeadmconstants.KubeAPIServer)),
			LivenessProbe:   staticpodutil.ComponentProbe(cfg, kubeadmconstants.KubeAPIServer, int(cfg.API.BindPort), "/healthz", v1.URISchemeHTTPS),
			Resources:       staticpodutil.ComponentResources("250m"),
			Env:             getProxyEnvVars(),
		}, mounts.GetVolumes(kubeadmconstants.KubeAPIServer)),
		kubeadmconstants.KubeControllerManager: staticpodutil.ComponentPod(v1.Container{
			Name:            kubeadmconstants.KubeControllerManager,
			Image:           images.GetCoreImage(kubeadmconstants.KubeControllerManager, cfg.GetControlPlaneImageRepository(), cfg.KubernetesVersion, cfg.UnifiedControlPlaneImage),
			ImagePullPolicy: cfg.ImagePullPolicy,
			Command:         getControllerManagerCommand(cfg, k8sVersion),
			VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(kubeadmconstants.KubeControllerManager)),
			LivenessProbe:   staticpodutil.ComponentProbe(cfg, kubeadmconstants.KubeControllerManager, 10252, "/healthz", v1.URISchemeHTTP),
			Resources:       staticpodutil.ComponentResources("200m"),
			Env:             getProxyEnvVars(),
		}, mounts.GetVolumes(kubeadmconstants.KubeControllerManager)),
		kubeadmconstants.KubeScheduler: staticpodutil.ComponentPod(v1.Container{
			Name:            kubeadmconstants.KubeScheduler,
			Image:           images.GetCoreImage(kubeadmconstants.KubeScheduler, cfg.GetControlPlaneImageRepository(), cfg.KubernetesVersion, cfg.UnifiedControlPlaneImage),
			ImagePullPolicy: cfg.ImagePullPolicy,
			Command:         getSchedulerCommand(cfg),
			VolumeMounts:    staticpodutil.VolumeMountMapToSlice(mounts.GetVolumeMounts(kubeadmconstants.KubeScheduler)),
			LivenessProbe:   staticpodutil.ComponentProbe(cfg, kubeadmconstants.KubeScheduler, 10251, "/healthz", v1.URISchemeHTTP),
			Resources:       staticpodutil.ComponentResources("100m"),
			Env:             getProxyEnvVars(),
		}, mounts.GetVolumes(kubeadmconstants.KubeScheduler)),
	}

	// Some cloud providers need extra privileges for example to load node information from a config drive
	// TODO: when we fully to external cloud providers and the api server and controller manager do not need
	// to call out to cloud provider code, we can remove the support for the PrivilegedPods
	if cfg.PrivilegedPods {
		staticPodSpecs[kubeadmconstants.KubeAPIServer].Spec.Containers[0].SecurityContext = &v1.SecurityContext{
			Privileged: utilpointer.BoolPtr(true),
		}
		staticPodSpecs[kubeadmconstants.KubeControllerManager].Spec.Containers[0].SecurityContext = &v1.SecurityContext{
			Privileged: utilpointer.BoolPtr(true),
		}
	}

	return staticPodSpecs
}

// createStaticPodFiles creates all the requested static pod files.
func createStaticPodFiles(manifestDir string, cfg *kubeadmapi.MasterConfiguration, componentNames ...string) error {
	// TODO: Move the "pkg/util/version".Version object into the internal API instead of always parsing the string
	k8sVersion, err := version.ParseSemantic(cfg.KubernetesVersion)
	if err != nil {
		return err
	}

	// gets the StaticPodSpecs, actualized for the current MasterConfiguration
	glog.V(1).Infoln("[controlplane] getting StaticPodSpecs")
	specs := GetStaticPodSpecs(cfg, k8sVersion)

	// creates required static pod specs
	for _, componentName := range componentNames {
		// retrives the StaticPodSpec for given component
		spec, exists := specs[componentName]
		if !exists {
			return fmt.Errorf("couldn't retrive StaticPodSpec for %s", componentName)
		}

		// writes the StaticPodSpec to disk
		if err := staticpodutil.WriteStaticPodToDisk(componentName, manifestDir, spec); err != nil {
			return fmt.Errorf("failed to create static pod manifest file for %q: %v", componentName, err)
		}

		glog.Infof("[controlplane] wrote Static Pod manifest for component %s to %q\n", componentName, kubeadmconstants.GetStaticPodFilepath(componentName, manifestDir))
	}

	return nil
}

// getAPIServerCommand builds the right API server command from the given config object and version
func getAPIServerCommand(cfg *kubeadmapi.MasterConfiguration) []string {
	defaultArguments := map[string]string{
		"advertise-address":               cfg.API.AdvertiseAddress,
		"insecure-port":                   "0",
		"admission-control":               defaultV19AdmissionControl,
		"service-cluster-ip-range":        cfg.Networking.ServiceSubnet,
		"service-account-key-file":        filepath.Join(cfg.CertificatesDir, kubeadmconstants.ServiceAccountPublicKeyName),
		"client-ca-file":                  filepath.Join(cfg.CertificatesDir, kubeadmconstants.CACertName),
		"tls-cert-file":                   filepath.Join(cfg.CertificatesDir, kubeadmconstants.APIServerCertName),
		"tls-private-key-file":            filepath.Join(cfg.CertificatesDir, kubeadmconstants.APIServerKeyName),
		"kubelet-client-certificate":      filepath.Join(cfg.CertificatesDir, kubeadmconstants.APIServerKubeletClientCertName),
		"kubelet-client-key":              filepath.Join(cfg.CertificatesDir, kubeadmconstants.APIServerKubeletClientKeyName),
		"enable-bootstrap-token-auth":     "true",
		"secure-port":                     fmt.Sprintf("%d", cfg.API.BindPort),
		"allow-privileged":                "true",
		"kubelet-preferred-address-types": "InternalIP,ExternalIP,Hostname",
		// add options to configure the front proxy.  Without the generated client cert, this will never be useable
		// so add it unconditionally with recommended values
		"requestheader-username-headers":     "X-Remote-User",
		"requestheader-group-headers":        "X-Remote-Group",
		"requestheader-extra-headers-prefix": "X-Remote-Extra-",
		"requestheader-client-ca-file":       filepath.Join(cfg.CertificatesDir, kubeadmconstants.FrontProxyCACertName),
		"requestheader-allowed-names":        "front-proxy-client",
		"proxy-client-cert-file":             filepath.Join(cfg.CertificatesDir, kubeadmconstants.FrontProxyClientCertName),
		"proxy-client-key-file":              filepath.Join(cfg.CertificatesDir, kubeadmconstants.FrontProxyClientKeyName),
	}

	command := []string{"kube-apiserver"}

	if cfg.CloudProvider == "aws" || cfg.CloudProvider == "gce" {
		defaultArguments["admission-control"] = deprecatedV19AdmissionControl
	}

	// If the user set endpoints for an external etcd cluster
	if len(cfg.Etcd.Endpoints) > 0 {
		defaultArguments["etcd-servers"] = strings.Join(cfg.Etcd.Endpoints, ",")

		// Use any user supplied etcd certificates
		if cfg.Etcd.CAFile != "" {
			defaultArguments["etcd-cafile"] = cfg.Etcd.CAFile
		}
		if cfg.Etcd.CertFile != "" && cfg.Etcd.KeyFile != "" {
			defaultArguments["etcd-certfile"] = cfg.Etcd.CertFile
			defaultArguments["etcd-keyfile"] = cfg.Etcd.KeyFile
		}
	} else {
		// Default to etcd static pod on localhost
		defaultArguments["etcd-servers"] = "https://127.0.0.1:2379"
		defaultArguments["etcd-cafile"] = filepath.Join(cfg.CertificatesDir, kubeadmconstants.EtcdCACertName)
		defaultArguments["etcd-certfile"] = filepath.Join(cfg.CertificatesDir, kubeadmconstants.APIServerEtcdClientCertName)
		defaultArguments["etcd-keyfile"] = filepath.Join(cfg.CertificatesDir, kubeadmconstants.APIServerEtcdClientKeyName)

		// Warn for unused user supplied variables
		if cfg.Etcd.CAFile != "" {
			glog.Warningf("[controlplane] configuration for %s CAFile, %s, is unused without providing Endpoints for external %s\n", kubeadmconstants.Etcd, cfg.Etcd.CAFile, kubeadmconstants.Etcd)
		}
		if cfg.Etcd.CertFile != "" {
			glog.Warningf("[controlplane] configuration for %s CertFile, %s, is unused without providing Endpoints for external %s\n", kubeadmconstants.Etcd, cfg.Etcd.CertFile, kubeadmconstants.Etcd)
		}
		if cfg.Etcd.KeyFile != "" {
			glog.Warningf("[controlplane] configuration for %s KeyFile, %s, is unused without providing Endpoints for external %s\n", kubeadmconstants.Etcd, cfg.Etcd.KeyFile, kubeadmconstants.Etcd)
		}
	}

	if cfg.CloudProvider != "" {
		defaultArguments["cloud-provider"] = cfg.CloudProvider

		// Only append the --cloud-config option if there's a such file
		if _, err := os.Stat(DefaultCloudConfigPath); err == nil {
			defaultArguments["cloud-config"] = DefaultCloudConfigPath
		}
	}

	if features.Enabled(cfg.FeatureGates, features.HighAvailability) {
		defaultArguments["endpoint-reconciler-type"] = reconcilers.LeaseEndpointReconcilerType
	}

	if features.Enabled(cfg.FeatureGates, features.DynamicKubeletConfig) {
		defaultArguments["feature-gates"] = "DynamicKubeletConfig=true"
	}

	if features.Enabled(cfg.FeatureGates, features.Auditing) {
		defaultArguments["audit-policy-file"] = kubeadmconstants.GetStaticPodAuditPolicyFile()
		defaultArguments["audit-log-path"] = filepath.Join(kubeadmconstants.StaticPodAuditPolicyLogDir, kubeadmconstants.AuditPolicyLogFile)
		if cfg.AuditPolicyConfiguration.LogMaxAge == nil {
			defaultArguments["audit-log-maxage"] = fmt.Sprintf("%d", kubeadmapiext.DefaultAuditPolicyLogMaxAge)
		} else {
			defaultArguments["audit-log-maxage"] = fmt.Sprintf("%d", *cfg.AuditPolicyConfiguration.LogMaxAge)
		}
	}

	command = append(command, kubeadmutil.BuildArgumentListFromMap(defaultArguments, cfg.APIServerExtraArgs)...)
	command = append(command, getAuthzParameters(cfg.AuthorizationModes)...)

	return command
}

// calcNodeCidrSize determines the size of the subnets used on each node, based
// on the pod subnet provided.  For IPv4, we assume that the pod subnet will
// be /16 and use /24. If the pod subnet cannot be parsed, the IPv4 value will
// be used (/24).
//
// For IPv6, the algorithm will do two three. First, the node CIDR will be set
// to a multiple of 8, using the available bits for easier readability by user.
// Second, the number of nodes will be 512 to 64K to attempt to maximize the
// number of nodes (see NOTE below). Third, pod networks of /113 and larger will
// be rejected, as the amount of bits available is too small.
//
// A special case is when the pod network size is /112, where /120 will be used,
// only allowing 256 nodes and 256 pods.
//
// If the pod network size is /113 or larger, the node CIDR will be set to the same
// size and this will be rejected later in validation.
//
// NOTE: Currently, the pod network must be /66 or larger. It is not reflected here,
// but a smaller value will fail later validation.
//
// NOTE: Currently, the design allows a maximum of 64K nodes. This algorithm splits
// the available bits to maximize the number used for nodes, but still have the node
// CIDR be a multiple of eight.
//
func calcNodeCidrSize(podSubnet string) string {
	maskSize := "24"
	if ip, podCidr, err := net.ParseCIDR(podSubnet); err == nil {
		if ip.To4() == nil {
			var nodeCidrSize int
			podNetSize, totalBits := podCidr.Mask.Size()
			switch {
			case podNetSize == 112:
				// Special case, allows 256 nodes, 256 pods/node
				nodeCidrSize = 120
			case podNetSize < 112:
				// Use multiple of 8 for node CIDR, with 512 to 64K nodes
				nodeCidrSize = totalBits - ((totalBits-podNetSize-1)/8-1)*8
			default:
				// Not enough bits, will fail later, when validate
				nodeCidrSize = podNetSize
			}
			maskSize = strconv.Itoa(nodeCidrSize)
		}
	}
	return maskSize
}

// getControllerManagerCommand builds the right controller manager command from the given config object and version
func getControllerManagerCommand(cfg *kubeadmapi.MasterConfiguration, k8sVersion *version.Version) []string {
	defaultArguments := map[string]string{
		"address":                          "127.0.0.1",
		"leader-elect":                     "true",
		"kubeconfig":                       filepath.Join(kubeadmconstants.KubernetesDir, kubeadmconstants.ControllerManagerKubeConfigFileName),
		"root-ca-file":                     filepath.Join(cfg.CertificatesDir, kubeadmconstants.CACertName),
		"service-account-private-key-file": filepath.Join(cfg.CertificatesDir, kubeadmconstants.ServiceAccountPrivateKeyName),
		"cluster-signing-cert-file":        filepath.Join(cfg.CertificatesDir, kubeadmconstants.CACertName),
		"cluster-signing-key-file":         filepath.Join(cfg.CertificatesDir, kubeadmconstants.CAKeyName),
		"use-service-account-credentials":  "true",
		"controllers":                      "*,bootstrapsigner,tokencleaner",
	}

	// If using external CA, pass empty string to controller manager instead of ca.key/ca.crt path,
	// so that the csrsigning controller fails to start
	if res, _ := certphase.UsingExternalCA(cfg); res {
		defaultArguments["cluster-signing-key-file"] = ""
		defaultArguments["cluster-signing-cert-file"] = ""
	}

	if cfg.CloudProvider != "" {
		defaultArguments["cloud-provider"] = cfg.CloudProvider

		// Only append the --cloud-config option if there's a such file
		if _, err := os.Stat(DefaultCloudConfigPath); err == nil {
			defaultArguments["cloud-config"] = DefaultCloudConfigPath
		}
	}

	// Let the controller-manager allocate Node CIDRs for the Pod network.
	// Each node will get a subspace of the address CIDR provided with --pod-network-cidr.
	if cfg.Networking.PodSubnet != "" {
		maskSize := calcNodeCidrSize(cfg.Networking.PodSubnet)
		defaultArguments["allocate-node-cidrs"] = "true"
		defaultArguments["cluster-cidr"] = cfg.Networking.PodSubnet
		defaultArguments["node-cidr-mask-size"] = maskSize
	}

	command := []string{"kube-controller-manager"}
	command = append(command, kubeadmutil.BuildArgumentListFromMap(defaultArguments, cfg.ControllerManagerExtraArgs)...)

	return command
}

// getSchedulerCommand builds the right scheduler command from the given config object and version
func getSchedulerCommand(cfg *kubeadmapi.MasterConfiguration) []string {
	defaultArguments := map[string]string{
		"address":      "127.0.0.1",
		"leader-elect": "true",
		"kubeconfig":   filepath.Join(kubeadmconstants.KubernetesDir, kubeadmconstants.SchedulerKubeConfigFileName),
	}

	command := []string{"kube-scheduler"}
	command = append(command, kubeadmutil.BuildArgumentListFromMap(defaultArguments, cfg.SchedulerExtraArgs)...)
	return command
}

// getProxyEnvVars builds a list of environment variables to use in the control plane containers in order to use the right proxy
func getProxyEnvVars() []v1.EnvVar {
	envs := []v1.EnvVar{}
	for _, env := range os.Environ() {
		pos := strings.Index(env, "=")
		if pos == -1 {
			// malformed environment variable, skip it.
			continue
		}
		name := env[:pos]
		value := env[pos+1:]
		if strings.HasSuffix(strings.ToLower(name), "_proxy") && value != "" {
			envVar := v1.EnvVar{Name: name, Value: value}
			envs = append(envs, envVar)
		}
	}
	return envs
}

// getAuthzParameters gets the authorization-related parameters to the api server
// At this point, we can assume the list of authorization modes is valid (due to that it has been validated in the API machinery code already)
// If the list is empty; it's defaulted (mostly for unit testing)
func getAuthzParameters(modes []string) []string {
	command := []string{}
	strset := sets.NewString(modes...)

	if len(modes) == 0 {
		return []string{fmt.Sprintf("--authorization-mode=%s", kubeadmapiext.DefaultAuthorizationModes)}
	}

	if strset.Has(authzmodes.ModeABAC) {
		command = append(command, "--authorization-policy-file="+kubeadmconstants.AuthorizationPolicyPath)
	}
	if strset.Has(authzmodes.ModeWebhook) {
		command = append(command, "--authorization-webhook-config-file="+kubeadmconstants.AuthorizationWebhookConfigPath)
	}

	command = append(command, "--authorization-mode="+strings.Join(modes, ","))
	return command
}
