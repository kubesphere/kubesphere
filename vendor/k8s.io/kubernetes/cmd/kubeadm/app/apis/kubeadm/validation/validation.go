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

package validation

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	bootstrapapi "k8s.io/client-go/tools/bootstrap/token/api"
	bootstraputil "k8s.io/client-go/tools/bootstrap/token/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	tokenutil "k8s.io/kubernetes/cmd/kubeadm/app/util/token"
	apivalidation "k8s.io/kubernetes/pkg/apis/core/validation"
	authzmodes "k8s.io/kubernetes/pkg/kubeapiserver/authorizer/modes"
	"k8s.io/kubernetes/pkg/kubelet/apis/kubeletconfig"
	kubeletscheme "k8s.io/kubernetes/pkg/kubelet/apis/kubeletconfig/scheme"
	kubeletvalidation "k8s.io/kubernetes/pkg/kubelet/apis/kubeletconfig/validation"
	"k8s.io/kubernetes/pkg/proxy/apis/kubeproxyconfig"
	kubeproxyscheme "k8s.io/kubernetes/pkg/proxy/apis/kubeproxyconfig/scheme"
	kubeproxyconfigv1alpha1 "k8s.io/kubernetes/pkg/proxy/apis/kubeproxyconfig/v1alpha1"
	proxyvalidation "k8s.io/kubernetes/pkg/proxy/apis/kubeproxyconfig/validation"
	"k8s.io/kubernetes/pkg/registry/core/service/ipallocator"
	"k8s.io/kubernetes/pkg/util/node"
)

// TODO: Break out the cloudprovider functionality out of core and only support the new flow
// described in https://github.com/kubernetes/community/pull/128
var cloudproviders = []string{
	"aws",
	"azure",
	"cloudstack",
	"gce",
	"external", // Support for out-of-tree cloud providers
	"openstack",
	"ovirt",
	"photon",
	"vsphere",
}

// Describes the authorization modes that are enforced by kubeadm
var requiredAuthzModes = []string{
	authzmodes.ModeRBAC,
	authzmodes.ModeNode,
}

// ValidateMasterConfiguration validates master configuration and collects all encountered errors
func ValidateMasterConfiguration(c *kubeadm.MasterConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateCloudProvider(c.CloudProvider, field.NewPath("cloudProvider"))...)
	allErrs = append(allErrs, ValidateAuthorizationModes(c.AuthorizationModes, field.NewPath("authorizationModes"))...)
	allErrs = append(allErrs, ValidateNetworking(&c.Networking, field.NewPath("networking"))...)
	allErrs = append(allErrs, ValidateCertSANs(c.APIServerCertSANs, field.NewPath("apiServerCertSANs"))...)
	allErrs = append(allErrs, ValidateCertSANs(c.Etcd.ServerCertSANs, field.NewPath("etcd").Child("serverCertSANs"))...)
	allErrs = append(allErrs, ValidateCertSANs(c.Etcd.PeerCertSANs, field.NewPath("etcd").Child("peerCertSANs"))...)
	allErrs = append(allErrs, ValidateAbsolutePath(c.CertificatesDir, field.NewPath("certificatesDir"))...)
	allErrs = append(allErrs, ValidateNodeName(c.NodeName, field.NewPath("nodeName"))...)
	allErrs = append(allErrs, ValidateToken(c.Token, field.NewPath("token"))...)
	allErrs = append(allErrs, ValidateTokenUsages(c.TokenUsages, field.NewPath("tokenUsages"))...)
	allErrs = append(allErrs, ValidateTokenGroups(c.TokenUsages, c.TokenGroups, field.NewPath("tokenGroups"))...)
	allErrs = append(allErrs, ValidateFeatureGates(c.FeatureGates, field.NewPath("featureGates"))...)
	allErrs = append(allErrs, ValidateAPIEndpoint(&c.API, field.NewPath("api"))...)
	allErrs = append(allErrs, ValidateProxy(c.KubeProxy.Config, field.NewPath("kubeProxy").Child("config"))...)
	if features.Enabled(c.FeatureGates, features.DynamicKubeletConfig) {
		allErrs = append(allErrs, ValidateKubeletConfiguration(&c.KubeletConfiguration, field.NewPath("kubeletConfiguration"))...)
	}
	return allErrs
}

// ValidateProxy validates proxy configuration and collects all encountered errors
func ValidateProxy(c *kubeproxyconfigv1alpha1.KubeProxyConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// Convert to the internal version
	internalcfg := &kubeproxyconfig.KubeProxyConfiguration{}
	err := kubeproxyscheme.Scheme.Convert(c, internalcfg, nil)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, "", err.Error()))
		return allErrs
	}
	return proxyvalidation.Validate(internalcfg)
}

// ValidateNodeConfiguration validates node configuration and collects all encountered errors
func ValidateNodeConfiguration(c *kubeadm.NodeConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateDiscovery(c)...)

	if !filepath.IsAbs(c.CACertPath) || !strings.HasSuffix(c.CACertPath, ".crt") {
		allErrs = append(allErrs, field.Invalid(field.NewPath("caCertPath"), c.CACertPath, "the ca certificate path must be an absolute path"))
	}
	return allErrs
}

// ValidateAuthorizationModes  validates authorization modes and collects all encountered errors
func ValidateAuthorizationModes(authzModes []string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	found := map[string]bool{}
	for _, authzMode := range authzModes {
		if !authzmodes.IsValidAuthorizationMode(authzMode) {
			allErrs = append(allErrs, field.Invalid(fldPath, authzMode, "invalid authorization mode"))
		}

		if found[authzMode] {
			allErrs = append(allErrs, field.Invalid(fldPath, authzMode, "duplicate authorization mode"))
			continue
		}
		found[authzMode] = true
	}
	for _, requiredMode := range requiredAuthzModes {
		if !found[requiredMode] {
			allErrs = append(allErrs, field.Required(fldPath, fmt.Sprintf("authorization mode %s must be enabled", requiredMode)))
		}
	}
	return allErrs
}

// ValidateDiscovery validates discovery related configuration and collects all encountered errors
func ValidateDiscovery(c *kubeadm.NodeConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(c.DiscoveryToken) != 0 {
		allErrs = append(allErrs, ValidateToken(c.DiscoveryToken, field.NewPath("discoveryToken"))...)
	}
	if len(c.DiscoveryFile) != 0 {
		allErrs = append(allErrs, ValidateDiscoveryFile(c.DiscoveryFile, field.NewPath("discoveryFile"))...)
	}
	allErrs = append(allErrs, ValidateArgSelection(c, field.NewPath("discovery"))...)
	allErrs = append(allErrs, ValidateToken(c.TLSBootstrapToken, field.NewPath("tlsBootstrapToken"))...)
	allErrs = append(allErrs, ValidateJoinDiscoveryTokenAPIServer(c.DiscoveryTokenAPIServers, field.NewPath("discoveryTokenAPIServers"))...)

	return allErrs
}

// ValidateArgSelection validates discovery related configuration and collects all encountered errors
func ValidateArgSelection(cfg *kubeadm.NodeConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(cfg.DiscoveryToken) != 0 && len(cfg.DiscoveryFile) != 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "discoveryToken and discoveryFile cannot both be set"))
	}
	if len(cfg.DiscoveryToken) == 0 && len(cfg.DiscoveryFile) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "discoveryToken or discoveryFile must be set"))
	}
	if len(cfg.DiscoveryTokenAPIServers) < 1 && len(cfg.DiscoveryToken) != 0 {
		allErrs = append(allErrs, field.Required(fldPath, "discoveryTokenAPIServers not set"))
	}

	if len(cfg.DiscoveryFile) != 0 && len(cfg.DiscoveryTokenCACertHashes) != 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "discoveryTokenCACertHashes cannot be used with discoveryFile"))
	}

	if len(cfg.DiscoveryFile) == 0 && len(cfg.DiscoveryToken) != 0 &&
		len(cfg.DiscoveryTokenCACertHashes) == 0 && !cfg.DiscoveryTokenUnsafeSkipCAVerification {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "using token-based discovery without discoveryTokenCACertHashes can be unsafe. set --discovery-token-unsafe-skip-ca-verification to continue"))
	}

	// TODO remove once we support multiple api servers
	if len(cfg.DiscoveryTokenAPIServers) > 1 {
		fmt.Println("[validation] WARNING: kubeadm doesn't fully support multiple API Servers yet")
	}
	return allErrs
}

// ValidateJoinDiscoveryTokenAPIServer validates discovery token for API server
func ValidateJoinDiscoveryTokenAPIServer(apiServers []string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for _, m := range apiServers {
		_, _, err := net.SplitHostPort(m)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, m, err.Error()))
		}
	}
	return allErrs
}

// ValidateDiscoveryFile validates location of a discovery file
func ValidateDiscoveryFile(discoveryFile string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	u, err := url.Parse(discoveryFile)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, discoveryFile, "not a valid HTTPS URL or a file on disk"))
		return allErrs
	}

	if u.Scheme == "" {
		// URIs with no scheme should be treated as files
		if _, err := os.Stat(discoveryFile); os.IsNotExist(err) {
			allErrs = append(allErrs, field.Invalid(fldPath, discoveryFile, "not a valid HTTPS URL or a file on disk"))
		}
		return allErrs
	}

	if u.Scheme != "https" {
		allErrs = append(allErrs, field.Invalid(fldPath, discoveryFile, "if a URL is used, the scheme must be https"))
	}
	return allErrs
}

// ValidateToken validates token
func ValidateToken(t string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	id, secret, err := tokenutil.ParseToken(t)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, t, err.Error()))
	}

	if len(id) == 0 || len(secret) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, t, "token must be of form '[a-z0-9]{6}.[a-z0-9]{16}'"))
	}
	return allErrs
}

// ValidateTokenGroups validates token groups
func ValidateTokenGroups(usages []string, groups []string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// adding groups only makes sense for authentication
	usagesSet := sets.NewString(usages...)
	usageAuthentication := strings.TrimPrefix(bootstrapapi.BootstrapTokenUsageAuthentication, bootstrapapi.BootstrapTokenUsagePrefix)
	if len(groups) > 0 && !usagesSet.Has(usageAuthentication) {
		allErrs = append(allErrs, field.Invalid(fldPath, groups, fmt.Sprintf("token groups cannot be specified unless --usages includes %q", usageAuthentication)))
	}

	// validate any extra group names
	for _, group := range groups {
		if err := bootstraputil.ValidateBootstrapGroupName(group); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, groups, err.Error()))
		}
	}

	return allErrs
}

// ValidateTokenUsages validates token usages
func ValidateTokenUsages(usages []string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// validate usages
	if err := bootstraputil.ValidateUsages(usages); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, usages, err.Error()))
	}

	return allErrs
}

// ValidateCertSANs validates alternative names
func ValidateCertSANs(altnames []string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for _, altname := range altnames {
		if len(validation.IsDNS1123Subdomain(altname)) != 0 && net.ParseIP(altname) == nil {
			allErrs = append(allErrs, field.Invalid(fldPath, altname, "altname is not a valid dns label or ip address"))
		}
	}
	return allErrs
}

// ValidateIPFromString validates ip address
func ValidateIPFromString(ipaddr string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if net.ParseIP(ipaddr) == nil {
		allErrs = append(allErrs, field.Invalid(fldPath, ipaddr, "ip address is not valid"))
	}
	return allErrs
}

// ValidateIPNetFromString validates network portion of ip address
func ValidateIPNetFromString(subnet string, minAddrs int64, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	_, svcSubnet, err := net.ParseCIDR(subnet)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, subnet, "couldn't parse subnet"))
		return allErrs
	}
	numAddresses := ipallocator.RangeSize(svcSubnet)
	if numAddresses < minAddrs {
		allErrs = append(allErrs, field.Invalid(fldPath, subnet, "subnet is too small"))
	}
	return allErrs
}

// ValidateNetworking validates networking configuration
func ValidateNetworking(c *kubeadm.Networking, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, apivalidation.ValidateDNS1123Subdomain(c.DNSDomain, field.NewPath("dnsDomain"))...)
	allErrs = append(allErrs, ValidateIPNetFromString(c.ServiceSubnet, constants.MinimumAddressesInServiceSubnet, field.NewPath("serviceSubnet"))...)
	if len(c.PodSubnet) != 0 {
		allErrs = append(allErrs, ValidateIPNetFromString(c.PodSubnet, constants.MinimumAddressesInServiceSubnet, field.NewPath("podSubnet"))...)
	}
	return allErrs
}

// ValidateAbsolutePath validates whether provided path is absolute or not
func ValidateAbsolutePath(path string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if !filepath.IsAbs(path) {
		allErrs = append(allErrs, field.Invalid(fldPath, path, "path is not absolute"))
	}
	return allErrs
}

// ValidateNodeName validates the name of a node
func ValidateNodeName(nodename string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if node.GetHostname(nodename) != nodename {
		allErrs = append(allErrs, field.Invalid(fldPath, nodename, "nodename is not valid, must be lower case"))
	}
	return allErrs
}

// ValidateCloudProvider validates if cloud provider is supported
func ValidateCloudProvider(provider string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(provider) == 0 {
		return allErrs
	}
	for _, supported := range cloudproviders {
		if provider == supported {
			return allErrs
		}
	}
	allErrs = append(allErrs, field.Invalid(fldPath, provider, "cloudprovider not supported"))
	return allErrs
}

// ValidateMixedArguments validates passed arguments
func ValidateMixedArguments(flag *pflag.FlagSet) error {
	// If --config isn't set, we have nothing to validate
	if !flag.Changed("config") {
		return nil
	}

	mixedInvalidFlags := []string{}
	flag.Visit(func(f *pflag.Flag) {
		if f.Name == "config" || f.Name == "ignore-preflight-errors" || strings.HasPrefix(f.Name, "skip-") || f.Name == "dry-run" || f.Name == "kubeconfig" || f.Name == "v" {
			// "--skip-*" flags or other whitelisted flags can be set with --config
			return
		}
		mixedInvalidFlags = append(mixedInvalidFlags, f.Name)
	})

	if len(mixedInvalidFlags) != 0 {
		return fmt.Errorf("can not mix '--config' with arguments %v", mixedInvalidFlags)
	}
	return nil
}

// ValidateFeatureGates validates provided feature gates
func ValidateFeatureGates(featureGates map[string]bool, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	validFeatures := features.Keys(features.InitFeatureGates)

	// check valid feature names are provided
	for k := range featureGates {
		if !features.Supports(features.InitFeatureGates, k) {
			allErrs = append(allErrs, field.Invalid(fldPath, featureGates,
				fmt.Sprintf("%s is not a valid feature name. Valid features are: %s", k, validFeatures)))
		}
	}

	return allErrs
}

// ValidateAPIEndpoint validates API server's endpoint
func ValidateAPIEndpoint(c *kubeadm.API, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	endpoint, err := kubeadmutil.GetMasterEndpoint(c)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, endpoint, err.Error()))
	}
	return allErrs
}

// ValidateIgnorePreflightErrors validates duplicates in ignore-preflight-errors flag.
func ValidateIgnorePreflightErrors(ignorePreflightErrors []string, skipPreflightChecks bool) (sets.String, error) {
	ignoreErrors := sets.NewString()
	allErrs := field.ErrorList{}

	for _, item := range ignorePreflightErrors {
		ignoreErrors.Insert(strings.ToLower(item)) // parameters are case insensitive
	}

	// TODO: remove once deprecated flag --skip-preflight-checks is removed.
	if skipPreflightChecks {
		ignoreErrors.Insert("all")
	}

	if ignoreErrors.Has("all") && ignoreErrors.Len() > 1 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("ignore-preflight-errors"), strings.Join(ignoreErrors.List(), ","), "don't specify individual checks if 'all' is used"))
	}

	return ignoreErrors, allErrs.ToAggregate()
}

// ValidateKubeletConfiguration validates kubelet configuration and collects all encountered errors
func ValidateKubeletConfiguration(c *kubeadm.KubeletConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	scheme, _, err := kubeletscheme.NewSchemeAndCodecs()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, "kubeletConfiguration", err.Error()))
		return allErrs
	}

	// Convert versioned config to internal config
	internalcfg := &kubeletconfig.KubeletConfiguration{}
	err = scheme.Convert(c.BaseConfig, internalcfg, nil)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, "kubeletConfiguration", err.Error()))
		return allErrs
	}

	err = kubeletvalidation.ValidateKubeletConfiguration(internalcfg)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, "kubeletConfiguration", err.Error()))
	}

	return allErrs
}
