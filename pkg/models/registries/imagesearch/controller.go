/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package imagesearch

import (
	"context"
	"fmt"
	"sync"

	v1 "k8s.io/api/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"

	"kubesphere.io/kubesphere/pkg/constants"
)

var SharedImageSearchProviderController = NewController()

const (
	dockerHubRegisterProvider = "DockerHubRegistryProvider"
	harborRegisterProvider    = "HarborRegistryProvider"

	SecretTypeImageSearchProvider = "config.kubesphere.io/imagesearchprovider"
)

type Controller struct {
	imageSearchProviders      *sync.Map
	imageSearchProviderConfig *sync.Map
}

func NewController() *Controller {
	return &Controller{
		imageSearchProviders:      &sync.Map{},
		imageSearchProviderConfig: &sync.Map{}}
}

func (c *Controller) WatchConfigurationChanges(ctx context.Context, cache runtimecache.Cache) error {
	informer, err := cache.GetInformer(ctx, &v1.Secret{})
	if err != nil {
		return fmt.Errorf("get informer failed: %w", err)
	}

	c.initGenericProvider()

	_, err = informer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			return IsImageSearchProviderConfiguration(obj.(*v1.Secret))
		},
		Handler: &toolscache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.OnConfigurationChange(obj.(*v1.Secret))
			},
			UpdateFunc: func(old, new interface{}) {
				c.OnConfigurationChange(new.(*v1.Secret))
			},
			DeleteFunc: func(obj interface{}) {
				c.OnConfigurationDelete(obj.(*v1.Secret))
			},
		},
	})

	if err != nil {
		return fmt.Errorf("add event handler failed: %w", err)
	}

	return nil
}

func (c *Controller) GetDefaultProvider() SearchProvider {
	provider, _ := c.imageSearchProviders.Load(dockerHubRegisterProvider)
	return provider.(SearchProvider)
}

func (c *Controller) initGenericProvider() {
	dockerHubProvider, _ := searchProviderFactories[dockerHubRegisterProvider].Create(nil)
	c.imageSearchProviders.Store(dockerHubRegisterProvider, dockerHubProvider)

	harborProvider, _ := searchProviderFactories[harborRegisterProvider].Create(nil)
	c.imageSearchProviders.Store(harborRegisterProvider, harborProvider)
}

func IsImageSearchProviderConfiguration(secret *v1.Secret) bool {
	if secret.Namespace != constants.KubeSphereNamespace {
		return false
	}
	return secret.Type == SecretTypeImageSearchProvider
}

func (c *Controller) OnConfigurationDelete(secret *v1.Secret) {
	configuration, err := UnmarshalFrom(secret)
	if err != nil {
		klog.Errorf("failed to unmarshal secret data: %s", err)
		return
	}
	c.imageSearchProviders.Delete(configuration.Name)
	c.imageSearchProviderConfig.Delete(configuration.Name)
}

func (c *Controller) OnConfigurationChange(secret *v1.Secret) {
	configuration, err := UnmarshalFrom(secret)
	if err != nil {
		klog.Errorf("failed to unmarshal secret data: %s", err)
		return
	}

	if factory, ok := searchProviderFactories[configuration.Type]; ok {
		if provider, err := factory.Create(configuration.ProviderOptions); err != nil {
			klog.Error(fmt.Sprintf("failed to create image search provider %s: %s", configuration.Name, err))
		} else {
			c.imageSearchProviders.Store(configuration.Name, provider)
			c.imageSearchProviderConfig.Store(configuration.Name, configuration)
			klog.V(4).Infof("create image search provider %s successfully", configuration.Name)
		}
	} else {
		klog.Errorf("image search provider %s with type %s is not supported", configuration.Name, configuration.Type)
		return
	}

}

func (c *Controller) GetProvider(providerName string) (SearchProvider, bool) {
	if obj, ok := c.imageSearchProviders.Load(providerName); ok {
		if provider, ok := obj.(SearchProvider); ok {
			return provider, true
		}
	}
	return nil, false
}
