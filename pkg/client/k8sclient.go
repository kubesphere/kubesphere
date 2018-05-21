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
	"kubesphere.io/kubesphere/pkg/options"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/golang/glog"
)

var k8sClient *kubernetes.Clientset

func getKubeConfig() (kubeConfig *rest.Config, err error) {

	kubeConfigFile := options.ServerOptions.GetKubeConfigFile()

	if len(kubeConfigFile) > 0 {

		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigFile)
		if err != nil {
			return nil, err
		}

	} else {

		kubeConfig, err = rest.InClusterConfig()
		if err != nil{
			return nil, err
		}
	}

	return kubeConfig, nil

}

func NewK8sClient() *kubernetes.Clientset {
	if k8sClient != nil {
		return k8sClient
	}

	kubeConfig, err := getKubeConfig()
	if err != nil {
		glog.Error(err)
		panic(err)
	}

	k8sClient, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil{
		glog.Error(err)
		panic(err)
	}
	return k8sClient
}
