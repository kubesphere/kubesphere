package k8s

import (
	application "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned"
	istio "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

type nullClient struct {
}

func NewNullClient() Client {
	return &nullClient{}
}

func (n nullClient) Kubernetes() kubernetes.Interface {
	return nil
}

func (n nullClient) KubeSphere() kubesphere.Interface {
	return nil
}

func (n nullClient) Istio() istio.Interface {
	return nil
}

func (n nullClient) Application() application.Interface {
	return nil
}

func (n nullClient) Discovery() discovery.DiscoveryInterface {
	return nil
}

func (n nullClient) Master() string {
	return ""
}

func (n nullClient) Config() *rest.Config {
	return nil
}
