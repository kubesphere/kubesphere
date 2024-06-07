package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type nullClient struct {
	kubernetes.Interface
}

func NewNullClient() Client {
	return &nullClient{}
}

func (n nullClient) Master() string {
	return ""
}

func (n nullClient) Config() *rest.Config {
	return nil
}
