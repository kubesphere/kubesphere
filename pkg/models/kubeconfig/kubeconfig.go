/*
Copyright 2019 The KubeSphere Authors.

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

package kubeconfig

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	ConfigTypeKubeConfig           = "kubeconfig"
	SecretTypeKubeConfig           = "config.kubesphere.io/" + ConfigTypeKubeConfig
	FileName                       = "config"
	DefaultClusterName             = "local"
	DefaultNamespace               = "default"
	InClusterCAFilePath            = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	PrivateKeyAnnotation           = "kubesphere.io/private-key"
	UserKubeConfigSecretNameFormat = "kubeconfig-%s"
)

type Interface interface {
	GetKubeConfig(ctx context.Context, username string) (string, error)
}

type operator struct {
	reader    runtimeclient.Reader
	masterURL string
}

func NewReadOnlyOperator(reader runtimeclient.Reader, masterURL string) Interface {
	return &operator{reader: reader, masterURL: masterURL}
}

// GetKubeConfig returns kubeconfig data for the specified user
func (o *operator) GetKubeConfig(ctx context.Context, username string) (string, error) {
	secretName := fmt.Sprintf(UserKubeConfigSecretNameFormat, username)

	secret := &corev1.Secret{}
	if err := o.reader.Get(ctx,
		types.NamespacedName{Namespace: constants.KubeSphereNamespace, Name: secretName}, secret); err != nil {
		return "", err
	}

	data := secret.Data[FileName]
	kubeconfig, err := clientcmd.Load(data)
	if err != nil {
		return "", err
	}

	masterURL := o.masterURL
	// server host override
	if cluster := kubeconfig.Clusters[DefaultClusterName]; cluster != nil && masterURL != "" {
		cluster.Server = masterURL
	}

	data, err = clientcmd.Write(*kubeconfig)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
