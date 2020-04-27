/*
Copyright 2018 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"sigs.k8s.io/kubefed/pkg/kubefedctl/options"
)

const FederatedKindPrefix = "Federated"

// FedConfig provides a rest config based on the filesystem kubeconfig (via
// pathOptions) and context in order to talk to the host kubernetes cluster
// and the joining kubernetes cluster.
type FedConfig interface {
	HostConfig(context, kubeconfigPath string) (*rest.Config, error)
	ClusterConfig(context, kubeconfigPath string) (*rest.Config, error)
	GetClientConfig(ontext, kubeconfigPath string) clientcmd.ClientConfig
}

// fedConfig implements the FedConfig interface.
type fedConfig struct {
	pathOptions *clientcmd.PathOptions
}

// NewFedConfig creates a fedConfig for `kubefedctl` commands.
func NewFedConfig(pathOptions *clientcmd.PathOptions) FedConfig {
	return &fedConfig{
		pathOptions: pathOptions,
	}
}

// HostConfig provides a rest config to talk to the host kubernetes cluster
// based on the context and kubeconfig passed in.
func (a *fedConfig) HostConfig(context, kubeconfigPath string) (*rest.Config, error) {
	hostConfig := a.GetClientConfig(context, kubeconfigPath)
	hostClientConfig, err := hostConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return hostClientConfig, nil
}

// ClusterConfig provides a rest config to talk to the joining kubernetes
// cluster based on the context and kubeconfig passed in.
func (a *fedConfig) ClusterConfig(context, kubeconfigPath string) (*rest.Config, error) {
	clusterConfig := a.GetClientConfig(context, kubeconfigPath)
	clusterClientConfig, err := clusterConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return clusterClientConfig, nil
}

// getClientConfig is a helper method to create a client config from the
// context and kubeconfig passed as arguments.
func (a *fedConfig) GetClientConfig(context, kubeconfigPath string) clientcmd.ClientConfig {
	loadingRules := *a.pathOptions.LoadingRules
	loadingRules.Precedence = a.pathOptions.GetLoadingPrecedence()
	loadingRules.ExplicitPath = kubeconfigPath
	overrides := &clientcmd.ConfigOverrides{
		CurrentContext: context,
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&loadingRules, overrides)
}

// HostClientset provides a kubernetes API compliant clientset to
// communicate with the host cluster's kubernetes API server.
func HostClientset(config *rest.Config) (*kubeclient.Clientset, error) {
	return kubeclient.NewForConfig(config)
}

// ClusterClientset provides a kubernetes API compliant clientset to
// communicate with the joining cluster's kubernetes API server.
func ClusterClientset(config *rest.Config) (*kubeclient.Clientset, error) {
	return kubeclient.NewForConfig(config)
}

// ClusterServiceAccountName returns the name of a service account whose
// credentials are used by the host cluster to access the client cluster.
func ClusterServiceAccountName(joiningClusterName, hostClusterName string) string {
	return fmt.Sprintf("%s-%s", joiningClusterName, hostClusterName)
}

// RoleName returns the name of a Role or ClusterRole and its
// associated RoleBinding or ClusterRoleBinding that are used to allow
// the service account to access necessary resources on the cluster.
func RoleName(serviceAccountName string) string {
	return fmt.Sprintf("kubefed-controller-manager:%s", serviceAccountName)
}

// HealthCheckRoleName returns the name of a ClusterRole and its
// associated ClusterRoleBinding that is used to allow the service
// account to check the health of the cluster and list nodes.
func HealthCheckRoleName(serviceAccountName, namespace string) string {
	return fmt.Sprintf("kubefed-controller-manager:%s:healthcheck-%s", namespace, serviceAccountName)
}

// IsFederatedAPIResource checks if a resource with the given Kind and group is a Federated one
func IsFederatedAPIResource(kind, group string) bool {
	return strings.HasPrefix(kind, FederatedKindPrefix) && group == options.DefaultFederatedGroup
}

// GetNamespace returns namespace of the current context
func GetNamespace(hostClusterContext string, kubeconfig string, config FedConfig) (string, error) {
	clientConfig := config.GetClientConfig(hostClusterContext, kubeconfig)
	currentContext, err := options.CurrentContext(clientConfig)
	if err != nil {
		return "", err
	}

	ns, _, err := clientConfig.Namespace()
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get ClientConfig for host cluster context %q and kubeconfig %q",
			currentContext, kubeconfig)
	}

	if len(ns) == 0 {
		ns = "default"
	}
	return ns, nil
}
