package k8s

import (
	applicationclientset "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned"
	s2i "github.com/kubesphere/s2ioperator/pkg/client/clientset/versioned"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"strings"
)

type KubernetesClient struct {
	// kubernetes client interface
	k8s *kubernetes.Clientset

	// discovery client
	discoveryClient *discovery.DiscoveryClient

	// generated clientset
	ks *kubesphere.Clientset

	s2i *s2i.Clientset

	application *applicationclientset.Clientset

	master string

	config *rest.Config
}

// NewKubernetesClientOrDie creates KubernetesClient and panic if there is an error
func NewKubernetesClientOrDie(options *KubernetesOptions) *KubernetesClient {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		panic(err)
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	k := &KubernetesClient{
		k8s:             kubernetes.NewForConfigOrDie(config),
		discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(config),
		ks:              kubesphere.NewForConfigOrDie(config),
		s2i:             s2i.NewForConfigOrDie(config),
		application:     applicationclientset.NewForConfigOrDie(config),
		master:          config.Host,
		config:          config,
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
func NewKubernetesClient(options *KubernetesOptions) (*KubernetesClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		return nil, err
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	var k KubernetesClient
	k.k8s, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.ks, err = kubesphere.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.s2i, err = s2i.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.application, err = applicationclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.master = options.Master
	k.config = config

	return &k, nil
}

func (k *KubernetesClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *KubernetesClient) Discovery() discovery.DiscoveryInterface {
	return k.discoveryClient
}

func (k *KubernetesClient) KubeSphere() kubesphere.Interface {
	return k.ks
}

func (k *KubernetesClient) S2i() s2i.Interface {
	return k.s2i
}

func (k *KubernetesClient) Application() applicationclientset.Interface {
	return k.application
}

// master address used to generate kubeconfig for downloading
func (k *KubernetesClient) Master() string {
	return k.master
}

func (k *KubernetesClient) Config() *rest.Config {
	return k.config
}
