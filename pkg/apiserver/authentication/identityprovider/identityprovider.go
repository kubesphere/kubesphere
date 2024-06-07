/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package identityprovider

import (
	"context"
	"fmt"
	"sync"

	v1 "k8s.io/api/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
)

var SharedIdentityProviderController = NewController()

type Controller struct {
	identityProviders       *sync.Map
	identityProviderConfigs *sync.Map
}

func NewController() *Controller {
	return &Controller{identityProviders: &sync.Map{}, identityProviderConfigs: &sync.Map{}}
}

func (c *Controller) WatchConfigurationChanges(ctx context.Context, cache runtimecache.Cache) error {
	informer, err := cache.GetInformer(ctx, &v1.Secret{})
	if err != nil {
		return fmt.Errorf("get informer failed: %w", err)
	}

	_, err = informer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			return IsIdentityProviderConfiguration(obj.(*v1.Secret))
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

func (c *Controller) OnConfigurationDelete(secret *v1.Secret) {
	configuration, err := UnmarshalFrom(secret)
	if err != nil {
		klog.Errorf("failed to unmarshal secret data: %s", err)
		return
	}
	c.identityProviders.Delete(configuration.Name)
	c.identityProviderConfigs.Delete(configuration.Name)
}

func (c *Controller) OnConfigurationChange(secret *v1.Secret) {
	configuration, err := UnmarshalFrom(secret)
	if err != nil {
		klog.Errorf("failed to unmarshal secret data: %s", err)
		return
	}

	if genericProviderFactories[configuration.Type] == nil && oauthProviderFactories[configuration.Type] == nil {
		klog.Errorf("identity provider %s with type %s is not supported", configuration.Name, configuration.Type)
		return
	}

	if factory, ok := oauthProviderFactories[configuration.Type]; ok {
		if provider, err := factory.Create(configuration.ProviderOptions); err != nil {
			// don’t return errors, decoupling external dependencies
			klog.Error(fmt.Sprintf("failed to create identity provider %s: %s", configuration.Name, err))
		} else {
			c.identityProviders.Store(configuration.Name, provider)
			c.identityProviderConfigs.Store(configuration.Name, configuration)
			klog.Infof("create identity provider %s successfully", configuration.Name)
		}
	}
	if factory, ok := genericProviderFactories[configuration.Type]; ok {
		if provider, err := factory.Create(configuration.ProviderOptions); err != nil {
			klog.Error(fmt.Sprintf("failed to create identity provider %s: %s", configuration.Name, err))
		} else {
			c.identityProviders.Store(configuration.Name, provider)
			c.identityProviderConfigs.Store(configuration.Name, configuration)
			klog.V(4).Infof("create identity provider %s successfully", configuration.Name)
		}
	}

}

func (c *Controller) GetGenericProvider(providerName string) (GenericProvider, bool) {
	if obj, ok := c.identityProviders.Load(providerName); ok {
		if provider, ok := obj.(GenericProvider); ok {
			return provider, true
		}
	}
	return nil, false
}

func (c *Controller) GetOAuthProvider(providerName string) (OAuthProvider, bool) {
	if obj, ok := c.identityProviders.Load(providerName); ok {
		if provider, ok := obj.(OAuthProvider); ok {
			return provider, true
		}
	}
	return nil, false
}

func (c *Controller) ListConfigurations() []*Configuration {
	configurations := make([]*Configuration, 0)
	c.identityProviderConfigs.Range(func(key, value any) bool {
		if configuration, ok := value.(*Configuration); ok {
			configurations = append(configurations, configuration)
		}
		return true
	})
	return configurations
}
