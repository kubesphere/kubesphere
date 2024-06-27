/*
Copyright 2022 KubeSphere Authors

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

package utils

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
)

func IsClusterReady(cluster *clusterv1alpha1.Cluster) bool {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == clusterv1alpha1.ClusterReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsClusterSchedulable(cluster *clusterv1alpha1.Cluster) bool {
	if !cluster.DeletionTimestamp.IsZero() {
		return false
	}

	if !IsClusterReady(cluster) {
		return false
	}

	for _, condition := range cluster.Status.Conditions {
		if condition.Type == clusterv1alpha1.ClusterSchedulable && condition.Status == corev1.ConditionFalse {
			return false
		}
	}
	return true
}

func IsHostCluster(cluster *clusterv1alpha1.Cluster) bool {
	if _, ok := cluster.Labels[clusterv1alpha1.HostCluster]; ok {
		return true
	}
	return false
}

func BuildKubeconfigFromRestConfig(config *rest.Config) ([]byte, error) {
	apiConfig := api.NewConfig()

	apiCluster := &api.Cluster{
		Server:                   config.Host,
		CertificateAuthorityData: config.CAData,
	}

	// generated kubeconfig will be used by cluster federation, CAFile is not
	// accepted by kubefed, so we need read CAFile
	if len(apiCluster.CertificateAuthorityData) == 0 && len(config.CAFile) != 0 {
		caData, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, err
		}
		apiCluster.CertificateAuthorityData = caData
	}

	apiConfig.Clusters["kubernetes"] = apiCluster

	apiConfig.AuthInfos["kubernetes-admin"] = &api.AuthInfo{
		ClientCertificateData: config.CertData,
		ClientKeyData:         config.KeyData,
		Token:                 config.BearerToken,
	}

	if config.BearerTokenFile != "" {
		newToken, _ := os.ReadFile(config.BearerToken)
		if len(newToken) > 0 {
			apiConfig.AuthInfos["kubernetes-admin"].Token = string(newToken)
		}
	}

	apiConfig.Contexts["kubernetes-admin@kubernetes"] = &api.Context{
		Cluster:  "kubernetes",
		AuthInfo: "kubernetes-admin",
	}

	apiConfig.CurrentContext = "kubernetes-admin@kubernetes"

	return clientcmd.Write(*apiConfig)
}
