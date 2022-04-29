/*
Copyright 2020 KubeSphere Authors

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

package k8s

import (
	"strings"

	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

type Client interface {
	Kubernetes() kubernetes.Interface
	KubeSphere() kubesphere.Interface
	Istio() istioclient.Interface
	Snapshot() snapshotclient.Interface
	ApiExtensions() apiextensionsclient.Interface
	Prometheus() promresourcesclient.Interface
	Master() string
	Config() *rest.Config
}

type kubernetesClient struct {
	// kubernetes client interface
	k8s kubernetes.Interface

	// generated clientset
	ks kubesphere.Interface

	istio istioclient.Interface

	snapshot snapshotclient.Interface

	apiextensions apiextensionsclient.Interface

	prometheus promresourcesclient.Interface

	master string

	config *rest.Config
}

// NewKubernetesClientOrDie creates KubernetesClient and panic if there is an error
func NewKubernetesClientOrDie(options *KubernetesOptions) Client {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		panic(err)
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	k := &kubernetesClient{
		k8s:           kubernetes.NewForConfigOrDie(config),
		ks:            kubesphere.NewForConfigOrDie(config),
		istio:         istioclient.NewForConfigOrDie(config),
		snapshot:      snapshotclient.NewForConfigOrDie(config),
		apiextensions: apiextensionsclient.NewForConfigOrDie(config),
		prometheus:    promresourcesclient.NewForConfigOrDie(config),
		master:        config.Host,
		config:        config,
	}

	if options.Master != "" {
		k.master = options.Master
	}
	// The https prefix is automatically added when using sa.
	// But it will not be set automatically when reading from kubeconfig
	// which may cause some problems in the client of other languages.
	if !strings.HasPrefix(k.master, "http://") && !strings.HasPrefix(k.master, "https://") {
		k.master = "https://" + k.master
	}
	return k
}

// NewKubernetesClient creates a KubernetesClient
func NewKubernetesClient(options *KubernetesOptions) (Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		return nil, err
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	var k kubernetesClient
	k.k8s, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.ks, err = kubesphere.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.istio, err = istioclient.NewForConfig(config)

	if err != nil {
		return nil, err
	}

	k.snapshot, err = snapshotclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.apiextensions, err = apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.prometheus, err = promresourcesclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.master = options.Master
	k.config = config

	return &k, nil
}

func (k *kubernetesClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *kubernetesClient) KubeSphere() kubesphere.Interface {
	return k.ks
}

func (k *kubernetesClient) Istio() istioclient.Interface {
	return k.istio
}

func (k *kubernetesClient) Snapshot() snapshotclient.Interface {
	return k.snapshot
}

func (k *kubernetesClient) ApiExtensions() apiextensionsclient.Interface {
	return k.apiextensions
}

func (k *kubernetesClient) Prometheus() promresourcesclient.Interface {
	return k.prometheus
}

// master address used to generate kubeconfig for downloading
func (k *kubernetesClient) Master() string {
	return k.master
}

func (k *kubernetesClient) Config() *rest.Config {
	return k.config
}
