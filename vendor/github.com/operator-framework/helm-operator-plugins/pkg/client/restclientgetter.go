/*
Copyright 2020 The Operator-SDK Authors.

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

package client

import (
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	cached "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var _ genericclioptions.RESTClientGetter = &restClientGetter{}

func newRESTClientGetter(cfg *rest.Config, rm meta.RESTMapper, ns string) genericclioptions.RESTClientGetter {
	return &restClientGetter{
		restConfig:      cfg,
		restMapper:      rm,
		namespaceConfig: &namespaceClientConfig{ns},
	}
}

type restClientGetter struct {
	restConfig      *rest.Config
	restMapper      meta.RESTMapper
	namespaceConfig clientcmd.ClientConfig

	setupDiscoveryClient  sync.Once
	cachedDiscoveryClient discovery.CachedDiscoveryInterface
}

func (c *restClientGetter) ToRESTConfig() (*rest.Config, error) {
	return c.restConfig, nil
}

func (c *restClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	var (
		dc  discovery.DiscoveryInterface
		err error
	)
	c.setupDiscoveryClient.Do(func() {
		dc, err = discovery.NewDiscoveryClientForConfig(c.restConfig)
		if err != nil {
			return
		}
		c.cachedDiscoveryClient = cached.NewMemCacheClient(dc)
	})
	if err != nil {
		return nil, err
	}
	return c.cachedDiscoveryClient, nil
}

func (c *restClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	return c.restMapper, nil
}

func (c *restClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return c.namespaceConfig
}

var _ clientcmd.ClientConfig = &namespaceClientConfig{}

type namespaceClientConfig struct {
	namespace string
}

func (c namespaceClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return clientcmdapi.Config{}, nil
}

func (c namespaceClientConfig) ClientConfig() (*rest.Config, error) {
	return nil, nil
}

func (c namespaceClientConfig) Namespace() (string, bool, error) {
	return c.namespace, false, nil
}

func (c namespaceClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	return nil
}
