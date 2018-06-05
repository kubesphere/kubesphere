/*
Copyright 2017 The Kubernetes Authors.

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

package dns

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/mholt/caddy/caddyfile"

	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/util/version"
)

const (
	// KubeDNSServiceAccountName describes the name of the ServiceAccount for the kube-dns addon
	KubeDNSServiceAccountName  = "kube-dns"
	kubeDNSStubDomain          = "stubDomains"
	kubeDNSUpstreamNameservers = "upstreamNameservers"
	kubeDNSFederation          = "federations"
)

// EnsureDNSAddon creates the kube-dns or CoreDNS addon
func EnsureDNSAddon(cfg *kubeadmapi.MasterConfiguration, client clientset.Interface) error {
	k8sVersion, err := version.ParseSemantic(cfg.KubernetesVersion)
	if err != nil {
		return fmt.Errorf("couldn't parse kubernetes version %q: %v", cfg.KubernetesVersion, err)
	}
	if features.Enabled(cfg.FeatureGates, features.CoreDNS) {
		return coreDNSAddon(cfg, client, k8sVersion)
	}
	return kubeDNSAddon(cfg, client, k8sVersion)
}

func kubeDNSAddon(cfg *kubeadmapi.MasterConfiguration, client clientset.Interface, k8sVersion *version.Version) error {
	if err := CreateServiceAccount(client); err != nil {
		return err
	}

	dnsip, err := kubeadmconstants.GetDNSIP(cfg.Networking.ServiceSubnet)
	if err != nil {
		return err
	}

	var dnsBindAddr, dnsProbeAddr string
	if dnsip.To4() == nil {
		dnsBindAddr = "::1"
		dnsProbeAddr = "[" + dnsBindAddr + "]"
	} else {
		dnsBindAddr = "127.0.0.1"
		dnsProbeAddr = dnsBindAddr
	}

	// Get the YAML manifest conditionally based on the k8s version
	kubeDNSDeploymentBytes := GetKubeDNSManifest(k8sVersion)
	dnsDeploymentBytes, err := kubeadmutil.ParseTemplate(kubeDNSDeploymentBytes,
		struct{ ImageRepository, Arch, Version, DNSBindAddr, DNSProbeAddr, DNSDomain, MasterTaintKey string }{
			ImageRepository: cfg.ImageRepository,
			Arch:            runtime.GOARCH,
			// Get the kube-dns version conditionally based on the k8s version
			Version:        GetDNSVersion(k8sVersion, kubeadmconstants.KubeDNS),
			DNSBindAddr:    dnsBindAddr,
			DNSProbeAddr:   dnsProbeAddr,
			DNSDomain:      cfg.Networking.DNSDomain,
			MasterTaintKey: kubeadmconstants.LabelNodeRoleMaster,
		})
	if err != nil {
		return fmt.Errorf("error when parsing kube-dns deployment template: %v", err)
	}

	dnsServiceBytes, err := kubeadmutil.ParseTemplate(KubeDNSService, struct{ DNSIP string }{
		DNSIP: dnsip.String(),
	})
	if err != nil {
		return fmt.Errorf("error when parsing kube-proxy configmap template: %v", err)
	}

	if err := createKubeDNSAddon(dnsDeploymentBytes, dnsServiceBytes, client); err != nil {
		return err
	}
	fmt.Println("[addons] Applied essential addon: kube-dns")
	return nil
}

// CreateServiceAccount creates the necessary serviceaccounts that kubeadm uses/might use, if they don't already exist.
func CreateServiceAccount(client clientset.Interface) error {

	return apiclient.CreateOrUpdateServiceAccount(client, &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeDNSServiceAccountName,
			Namespace: metav1.NamespaceSystem,
		},
	})
}

func createKubeDNSAddon(deploymentBytes, serviceBytes []byte, client clientset.Interface) error {
	kubednsDeployment := &apps.Deployment{}
	if err := kuberuntime.DecodeInto(legacyscheme.Codecs.UniversalDecoder(), deploymentBytes, kubednsDeployment); err != nil {
		return fmt.Errorf("unable to decode kube-dns deployment %v", err)
	}

	// Create the Deployment for kube-dns or update it in case it already exists
	if err := apiclient.CreateOrUpdateDeployment(client, kubednsDeployment); err != nil {
		return err
	}

	kubednsService := &v1.Service{}
	return createDNSService(kubednsService, serviceBytes, client)
}

func coreDNSAddon(cfg *kubeadmapi.MasterConfiguration, client clientset.Interface, k8sVersion *version.Version) error {
	// Get the YAML manifest conditionally based on the k8s version
	dnsDeploymentBytes := GetCoreDNSManifest(k8sVersion)
	coreDNSDeploymentBytes, err := kubeadmutil.ParseTemplate(dnsDeploymentBytes, struct{ MasterTaintKey, Version string }{
		MasterTaintKey: kubeadmconstants.LabelNodeRoleMaster,
		Version:        GetDNSVersion(k8sVersion, kubeadmconstants.CoreDNS),
	})
	if err != nil {
		return fmt.Errorf("error when parsing CoreDNS deployment template: %v", err)
	}

	// Get the kube-dns ConfigMap for translation to equivalent CoreDNS Config.
	kubeDNSConfigMap, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(kubeadmconstants.KubeDNS, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	stubDomain, err := translateStubDomainOfKubeDNSToProxyCoreDNS(kubeDNSStubDomain, kubeDNSConfigMap)
	if err != nil {
		return err
	}

	upstreamNameserver, err := translateUpstreamNameServerOfKubeDNSToUpstreamProxyCoreDNS(kubeDNSUpstreamNameservers, kubeDNSConfigMap)
	if err != nil {
		return err
	}
	coreDNSDomain := cfg.Networking.DNSDomain
	federations, err := translateFederationsofKubeDNSToCoreDNS(kubeDNSFederation, coreDNSDomain, kubeDNSConfigMap)
	if err != nil {
		return err
	}

	// Get the config file for CoreDNS
	coreDNSConfigMapBytes, err := kubeadmutil.ParseTemplate(CoreDNSConfigMap, struct{ DNSDomain, UpstreamNameserver, Federation, StubDomain string }{
		DNSDomain:          coreDNSDomain,
		UpstreamNameserver: upstreamNameserver,
		Federation:         federations,
		StubDomain:         stubDomain,
	})
	if err != nil {
		return fmt.Errorf("error when parsing CoreDNS configMap template: %v", err)
	}

	dnsip, err := kubeadmconstants.GetDNSIP(cfg.Networking.ServiceSubnet)
	if err != nil {
		return err
	}

	coreDNSServiceBytes, err := kubeadmutil.ParseTemplate(KubeDNSService, struct{ DNSIP string }{
		DNSIP: dnsip.String(),
	})

	if err != nil {
		return fmt.Errorf("error when parsing CoreDNS service template: %v", err)
	}

	if err := createCoreDNSAddon(coreDNSDeploymentBytes, coreDNSServiceBytes, coreDNSConfigMapBytes, client); err != nil {
		return err
	}
	fmt.Println("[addons] Applied essential addon: CoreDNS")
	return nil
}

func createCoreDNSAddon(deploymentBytes, serviceBytes, configBytes []byte, client clientset.Interface) error {
	coreDNSConfigMap := &v1.ConfigMap{}
	if err := kuberuntime.DecodeInto(legacyscheme.Codecs.UniversalDecoder(), configBytes, coreDNSConfigMap); err != nil {
		return fmt.Errorf("unable to decode CoreDNS configmap %v", err)
	}

	// Create the ConfigMap for CoreDNS or update it in case it already exists
	if err := apiclient.CreateOrUpdateConfigMap(client, coreDNSConfigMap); err != nil {
		return err
	}

	coreDNSClusterRoles := &rbac.ClusterRole{}
	if err := kuberuntime.DecodeInto(legacyscheme.Codecs.UniversalDecoder(), []byte(CoreDNSClusterRole), coreDNSClusterRoles); err != nil {
		return fmt.Errorf("unable to decode CoreDNS clusterroles %v", err)
	}

	// Create the Clusterroles for CoreDNS or update it in case it already exists
	if err := apiclient.CreateOrUpdateClusterRole(client, coreDNSClusterRoles); err != nil {
		return err
	}

	coreDNSClusterRolesBinding := &rbac.ClusterRoleBinding{}
	if err := kuberuntime.DecodeInto(legacyscheme.Codecs.UniversalDecoder(), []byte(CoreDNSClusterRoleBinding), coreDNSClusterRolesBinding); err != nil {
		return fmt.Errorf("unable to decode CoreDNS clusterrolebindings %v", err)
	}

	// Create the Clusterrolebindings for CoreDNS or update it in case it already exists
	if err := apiclient.CreateOrUpdateClusterRoleBinding(client, coreDNSClusterRolesBinding); err != nil {
		return err
	}

	coreDNSServiceAccount := &v1.ServiceAccount{}
	if err := kuberuntime.DecodeInto(legacyscheme.Codecs.UniversalDecoder(), []byte(CoreDNSServiceAccount), coreDNSServiceAccount); err != nil {
		return fmt.Errorf("unable to decode CoreDNS serviceaccount %v", err)
	}

	// Create the ConfigMap for CoreDNS or update it in case it already exists
	if err := apiclient.CreateOrUpdateServiceAccount(client, coreDNSServiceAccount); err != nil {
		return err
	}

	coreDNSDeployment := &apps.Deployment{}
	if err := kuberuntime.DecodeInto(legacyscheme.Codecs.UniversalDecoder(), deploymentBytes, coreDNSDeployment); err != nil {
		return fmt.Errorf("unable to decode CoreDNS deployment %v", err)
	}

	// Create the Deployment for CoreDNS or update it in case it already exists
	if err := apiclient.CreateOrUpdateDeployment(client, coreDNSDeployment); err != nil {
		return err
	}

	coreDNSService := &v1.Service{}
	return createDNSService(coreDNSService, serviceBytes, client)
}

func createDNSService(dnsService *v1.Service, serviceBytes []byte, client clientset.Interface) error {
	if err := kuberuntime.DecodeInto(legacyscheme.Codecs.UniversalDecoder(), serviceBytes, dnsService); err != nil {
		return fmt.Errorf("unable to decode the DNS service %v", err)
	}

	// Can't use a generic apiclient helper func here as we have to tolerate more than AlreadyExists.
	if _, err := client.CoreV1().Services(metav1.NamespaceSystem).Create(dnsService); err != nil {
		// Ignore if the Service is invalid with this error message:
		// 	Service "kube-dns" is invalid: spec.clusterIP: Invalid value: "10.96.0.10": provided IP is already allocated

		if !apierrors.IsAlreadyExists(err) && !apierrors.IsInvalid(err) {
			return fmt.Errorf("unable to create a new DNS service: %v", err)
		}

		if _, err := client.CoreV1().Services(metav1.NamespaceSystem).Update(dnsService); err != nil {
			return fmt.Errorf("unable to create/update the DNS service: %v", err)
		}
	}
	return nil
}

// translateStubDomainOfKubeDNSToProxyCoreDNS translates StubDomain Data in kube-dns ConfigMap
// in the form of Proxy for the CoreDNS Corefile.
func translateStubDomainOfKubeDNSToProxyCoreDNS(dataField string, kubeDNSConfigMap *v1.ConfigMap) (string, error) {
	if proxy, ok := kubeDNSConfigMap.Data[dataField]; ok {
		stubDomainData := make(map[string][]string)
		err := json.Unmarshal([]byte(proxy), &stubDomainData)
		if err != nil {
			return "", fmt.Errorf("failed to parse JSON from 'kube-dns ConfigMap: %v", err)
		}

		var proxyStanza []interface{}
		for domain, proxyIP := range stubDomainData {
			pStanza := map[string]interface{}{}
			pStanza["keys"] = []string{domain + ":53"}
			pStanza["body"] = [][]string{
				{"errors"},
				{"cache", "30"},
				append([]string{"proxy", "."}, proxyIP...),
			}
			proxyStanza = append(proxyStanza, pStanza)
		}
		stanzasBytes, err := json.Marshal(proxyStanza)
		if err != nil {
			return "", err
		}

		corefileStanza, err := caddyfile.FromJSON(stanzasBytes)
		if err != nil {
			return "", err
		}

		return prepCorefileFormat(string(corefileStanza), 4), nil
	}
	return "", nil
}

// translateUpstreamNameServerOfKubeDNSToUpstreamProxyCoreDNS translates UpstreamNameServer Data in kube-dns ConfigMap
// in the form of Proxy for the CoreDNS Corefile.
func translateUpstreamNameServerOfKubeDNSToUpstreamProxyCoreDNS(dataField string, kubeDNSConfigMap *v1.ConfigMap) (string, error) {
	if upstreamValues, ok := kubeDNSConfigMap.Data[dataField]; ok {
		var upstreamProxyIP []string

		err := json.Unmarshal([]byte(upstreamValues), &upstreamProxyIP)
		if err != nil {
			return "", fmt.Errorf("failed to parse JSON from 'kube-dns ConfigMap: %v", err)
		}

		coreDNSProxyStanzaList := strings.Join(upstreamProxyIP, " ")
		return coreDNSProxyStanzaList, nil
	}
	return "/etc/resolv.conf", nil
}

// translateFederationsofKubeDNSToCoreDNS translates Federations Data in kube-dns ConfigMap
// to Federation for CoreDNS Corefile.
func translateFederationsofKubeDNSToCoreDNS(dataField, coreDNSDomain string, kubeDNSConfigMap *v1.ConfigMap) (string, error) {
	if federation, ok := kubeDNSConfigMap.Data[dataField]; ok {
		var (
			federationStanza []interface{}
			body             [][]string
		)
		federationData := make(map[string]string)

		err := json.Unmarshal([]byte(federation), &federationData)
		if err != nil {
			return "", fmt.Errorf("failed to parse JSON from kube-dns ConfigMap: %v", err)
		}
		fStanza := map[string]interface{}{}

		for name, domain := range federationData {
			body = append(body, []string{name, domain})
		}
		federationStanza = append(federationStanza, fStanza)
		fStanza["keys"] = []string{"federation " + coreDNSDomain}
		fStanza["body"] = body
		stanzasBytes, err := json.Marshal(federationStanza)
		if err != nil {
			return "", err
		}

		corefileStanza, err := caddyfile.FromJSON(stanzasBytes)
		if err != nil {
			return "", err
		}

		return prepCorefileFormat(string(corefileStanza), 8), nil
	}
	return "", nil
}

// prepCorefileFormat indents the output of the Corefile caddytext and replaces tabs with spaces
// to neatly format the configmap, making it readable.
func prepCorefileFormat(s string, indentation int) string {
	r := []string{}
	for _, line := range strings.Split(s, "\n") {
		indented := strings.Repeat(" ", indentation) + line
		r = append(r, indented)
	}
	corefile := strings.Join(r, "\n")
	return "\n" + strings.Replace(corefile, "\t", "   ", -1)
}
