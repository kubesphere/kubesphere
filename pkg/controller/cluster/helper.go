package cluster

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func buildKubeconfigFromRestConfig(config *rest.Config) ([]byte, error) {
	apiConfig := api.NewConfig()

	apiConfig.Clusters["kubernetes"] = &api.Cluster{
		Server:                   config.Host,
		CertificateAuthorityData: config.CAData,
		CertificateAuthority:     config.CAFile,
	}

	apiConfig.AuthInfos["kubernetes-admin"] = &api.AuthInfo{
		ClientCertificate:     config.CertFile,
		ClientCertificateData: config.CertData,
		ClientKey:             config.KeyFile,
		ClientKeyData:         config.KeyData,
		TokenFile:             config.BearerTokenFile,
		Token:                 config.BearerToken,
		Username:              config.Username,
		Password:              config.Password,
	}

	apiConfig.Contexts["kubernetes-admin@kubernetes"] = &api.Context{
		Cluster:  "kubernetes",
		AuthInfo: "kubernetes-admin",
	}

	apiConfig.CurrentContext = "kubernetes-admin@kubernetes"

	return clientcmd.Write(*apiConfig)
}
