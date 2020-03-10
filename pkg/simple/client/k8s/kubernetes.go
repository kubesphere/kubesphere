package k8s

import (
	applicationclientset "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"strings"
)

type Client interface {
	Kubernetes() kubernetes.Interface
	KubeSphere() kubesphere.Interface
	Istio() istioclient.Interface
	Application() applicationclientset.Interface
	Discovery() discovery.DiscoveryInterface
	Master() string
	Config() *rest.Config
}

type kubernetesClient struct {
	// kubernetes client interface
	k8s kubernetes.Interface

	// discovery client
	discoveryClient *discovery.DiscoveryClient

	// generated clientset
	ks kubesphere.Interface

	application applicationclientset.Interface

	istio istioclient.Interface

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
		k8s:             kubernetes.NewForConfigOrDie(config),
		discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(config),
		ks:              kubesphere.NewForConfigOrDie(config),
		istio:           istioclient.NewForConfigOrDie(config),
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

	k.application, err = applicationclientset.NewForConfig(config)
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

func (k *kubernetesClient) Discovery() discovery.DiscoveryInterface {
	return k.discoveryClient
}

func (k *kubernetesClient) KubeSphere() kubesphere.Interface {
	return k.ks
}

func (k *kubernetesClient) Application() applicationclientset.Interface {
	return k.application
}

func (k *kubernetesClient) Istio() istioclient.Interface {
	return k.istio
}

// master address used to generate kubeconfig for downloading
func (k *kubernetesClient) Master() string {
	return k.master
}

func (k *kubernetesClient) Config() *rest.Config {
	return k.config
}
