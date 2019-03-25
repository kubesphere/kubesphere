/*

 Copyright 2019 The KubeSphere Authors.

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
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/mitchellh/go-homedir"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigFile string
	k8sClient      *kubernetes.Clientset
	k8sClientOnce  sync.Once
	KubeConfig     *rest.Config
)

func init() {
	flag.StringVar(&kubeConfigFile, "kubeconfig", fmt.Sprintf("%s/.kube/config", os.Getenv("HOME")), "path to kubeconfig file")
}

func Client() *kubernetes.Clientset {

	k8sClientOnce.Do(func() {

		config, err := Config()

		if err != nil {
			log.Fatalln(err)
		}

		k8sClient = kubernetes.NewForConfigOrDie(config)

		KubeConfig = config
	})

	return k8sClient
}

func Config() (kubeConfig *rest.Config, err error) {

	if kubeConfigFile == "" {
		if env := os.Getenv("KUBECONFIG"); env != "" {
			kubeConfigFile = env
		} else {
			if home, err := homedir.Dir(); err == nil {
				kubeConfigFile = fmt.Sprintf("%s/.kube/config", home)
			}
		}
	}

	if _, err = os.Stat(kubeConfigFile); err == nil {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	} else {
		kubeConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	kubeConfig.QPS = 1e6
	kubeConfig.Burst = 1e6

	return kubeConfig, nil
}
