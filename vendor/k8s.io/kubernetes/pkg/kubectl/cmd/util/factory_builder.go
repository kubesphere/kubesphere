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

// this file contains factories with no other dependencies

package util

import (
	"os"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	scaleclient "k8s.io/client-go/scale"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/plugins"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

type ring2Factory struct {
	clientAccessFactory  ClientAccessFactory
	objectMappingFactory ObjectMappingFactory
}

func NewBuilderFactory(clientAccessFactory ClientAccessFactory, objectMappingFactory ObjectMappingFactory) BuilderFactory {
	f := &ring2Factory{
		clientAccessFactory:  clientAccessFactory,
		objectMappingFactory: objectMappingFactory,
	}

	return f
}

// NewBuilder returns a new resource builder for structured api objects.
func (f *ring2Factory) NewBuilder() *resource.Builder {
	mapper, mapperErr := f.objectMappingFactory.RESTMapper()

	categoryExpander := f.objectMappingFactory.CategoryExpander()
	return resource.NewBuilder(
		f.clientAccessFactory.ClientConfig,
		mapper,
		categoryExpander,
	).AddError(mapperErr)
}

// PluginLoader loads plugins from a path set by the KUBECTL_PLUGINS_PATH env var.
// If this env var is not set, it defaults to
//   "~/.kube/plugins", plus
//  "./kubectl/plugins" directory under the "data dir" directory specified by the XDG
// system directory structure spec for the given platform.
func (f *ring2Factory) PluginLoader() plugins.PluginLoader {
	if len(os.Getenv("KUBECTL_PLUGINS_PATH")) > 0 {
		return plugins.KubectlPluginsPathPluginLoader()
	}
	return plugins.TolerantMultiPluginLoader{
		plugins.XDGDataDirsPluginLoader(),
		plugins.UserDirPluginLoader(),
	}
}

func (f *ring2Factory) PluginRunner() plugins.PluginRunner {
	return &plugins.ExecPluginRunner{}
}

func (f *ring2Factory) ScaleClient() (scaleclient.ScalesGetter, error) {
	discoClient, err := f.clientAccessFactory.DiscoveryClient()
	if err != nil {
		return nil, err
	}
	restClient, err := f.clientAccessFactory.RESTClient()
	if err != nil {
		return nil, err
	}
	resolver := scaleclient.NewDiscoveryScaleKindResolver(discoClient)
	mapper, err := f.objectMappingFactory.RESTMapper()
	if err != nil {
		return nil, err
	}

	return scaleclient.New(restClient, mapper, dynamic.LegacyAPIPathResolverFunc, resolver), nil
}

func (f *ring2Factory) Scaler() (kubectl.Scaler, error) {
	scalesGetter, err := f.ScaleClient()
	if err != nil {
		return nil, err
	}

	return kubectl.NewScaler(scalesGetter), nil
}

func (f *ring2Factory) Reaper(mapping *meta.RESTMapping) (kubectl.Reaper, error) {
	clientset, clientsetErr := f.clientAccessFactory.ClientSet()
	if clientsetErr != nil {
		return nil, clientsetErr
	}
	scaler, err := f.ScaleClient()
	if err != nil {
		return nil, err
	}

	reaper, reaperErr := kubectl.ReaperFor(mapping.GroupVersionKind.GroupKind(), clientset, scaler)
	if kubectl.IsNoSuchReaperError(reaperErr) {
		return nil, reaperErr
	}
	return reaper, reaperErr
}
