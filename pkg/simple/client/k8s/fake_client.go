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
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/clientset/versioned"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	application "sigs.k8s.io/application/pkg/client/clientset/versioned"
	applicationclientset "sigs.k8s.io/application/pkg/client/clientset/versioned"
)

type FakeClient struct {
	// kubernetes client interface
	K8sClient kubernetes.Interface

	// discovery client
	DiscoveryClient *discovery.DiscoveryClient

	// generated clientset
	KubeSphereClient kubesphere.Interface

	ApplicationClient applicationclientset.Interface

	IstioClient istioclient.Interface

	SnapshotClient snapshotclient.Interface

	ApiExtensionClient apiextensionsclient.Interface

	MasterURL string

	KubeConfig *rest.Config
}

func NewFakeClientSets(k8sClient kubernetes.Interface, discoveryClient *discovery.DiscoveryClient,
	kubeSphereClient kubesphere.Interface, applicationClient applicationclientset.Interface,
	istioClient istioclient.Interface, snapshotClient snapshotclient.Interface,
	apiextensionsclient apiextensionsclient.Interface, masterURL string, kubeConfig *rest.Config) Client {
	return &FakeClient{
		K8sClient:          k8sClient,
		DiscoveryClient:    discoveryClient,
		KubeSphereClient:   kubeSphereClient,
		ApplicationClient:  applicationClient,
		IstioClient:        istioClient,
		SnapshotClient:     snapshotClient,
		ApiExtensionClient: apiextensionsclient,
		MasterURL:          masterURL,
		KubeConfig:         kubeConfig,
	}
}

func (n *FakeClient) Kubernetes() kubernetes.Interface {
	return n.K8sClient
}

func (n *FakeClient) KubeSphere() kubesphere.Interface {
	return n.KubeSphereClient
}

func (n *FakeClient) Istio() istioclient.Interface {
	return n.IstioClient
}

func (n *FakeClient) Application() application.Interface {
	return n.ApplicationClient
}

func (n *FakeClient) Snapshot() snapshotclient.Interface {
	return nil
}

func (n *FakeClient) ApiExtensions() apiextensionsclient.Interface {
	return n.ApiExtensionClient
}

func (n *FakeClient) Discovery() discovery.DiscoveryInterface {
	return n.DiscoveryClient
}

func (n *FakeClient) Master() string {
	return n.MasterURL
}

func (n *FakeClient) Config() *rest.Config {
	return n.KubeConfig
}
