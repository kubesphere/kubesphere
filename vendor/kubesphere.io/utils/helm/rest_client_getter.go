package helm

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func NewClusterRESTClientGetter(kubeconfig []byte, namespace string) genericclioptions.RESTClientGetter {
	if len(kubeconfig) > 0 {
		return NewMemoryRESTClientGetter(kubeconfig, namespace)
	}
	flags := genericclioptions.NewConfigFlags(true)
	flags.Namespace = &namespace
	return flags
}

// MemoryRESTClientGetter is an implementation of the genericclioptions.RESTClientGetter.
type MemoryRESTClientGetter struct {
	kubeconfig []byte
	namespace  string
}

func NewMemoryRESTClientGetter(kubeconfig []byte, namespace string) genericclioptions.RESTClientGetter {
	return &MemoryRESTClientGetter{
		kubeconfig: kubeconfig,
		namespace:  namespace,
	}
}

func (c *MemoryRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	cfg, err := clientcmd.RESTConfigFromKubeConfig(c.kubeconfig)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *MemoryRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	discoveryClient, _ := discovery.NewDiscoveryClientForConfig(config)
	return memory.NewMemCacheClient(discoveryClient), nil
}

func (c *MemoryRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient, c.warning)
	return expander, nil
}

func (c *MemoryRESTClientGetter) warning(msg string) {
	klog.Warning(msg)
}

func (c *MemoryRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	overrides.Context.Namespace = c.namespace

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
}
