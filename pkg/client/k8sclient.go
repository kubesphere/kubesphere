/*
Copyright 2018 The KubeSphere Authors.
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

package client

import (
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"

	"kubesphere.io/kubesphere/pkg/options"
)

var k8sClient *kubernetes.Clientset

func NewK8sClient() *kubernetes.Clientset {
	if k8sClient != nil {
		return k8sClient
	}

	kubeConfig, err := options.ServerOptions.GetKubeConfig()
	if err != nil {
		glog.Error(err)
		panic(err)
	}

	k8sClient, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		glog.Error(err)
		panic(err)
	}
	return k8sClient
}
