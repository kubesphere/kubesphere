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

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	kubernetes.Interface
	Master() string
	Config() *rest.Config
}

type kubernetesClient struct {
	kubernetes.Interface
	master string
	config *rest.Config
}

// NewKubernetesClientOrDie creates KubernetesClient and panic if there is an error
func NewKubernetesClientOrDie(options *Options) Client {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		panic(err)
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	k := &kubernetesClient{
		Interface: kubernetes.NewForConfigOrDie(config),
		master:    config.Host,
		config:    config,
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
func NewKubernetesClient(options *Options) (Client, error) {
	config, err := clientcmd.BuildConfigFromFlags(options.Master, options.KubeConfig)
	if err != nil {
		return nil, err
	}
	config.QPS = options.QPS
	config.Burst = options.Burst
	var client kubernetesClient
	client.Interface, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.master = options.Master
	client.config = config
	return &client, nil
}

// Master address used to generate kubeconfig for downloading
func (k *kubernetesClient) Master() string {
	return k.master
}

func (k *kubernetesClient) Config() *rest.Config {
	return k.config
}
