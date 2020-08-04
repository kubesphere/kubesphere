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
	istio "istio.io/client-go/pkg/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	application "sigs.k8s.io/application/pkg/client/clientset/versioned"
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

func (n nullClient) Snapshot() snapshotclient.Interface {
	return nil
}

func (n nullClient) ApiExtensions() apiextensionsclient.Interface {
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
