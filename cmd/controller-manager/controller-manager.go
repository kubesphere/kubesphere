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

package main

import (
	"flag"
	"k8s.io/client-go/tools/clientcmd"
	"kubesphere.io/kubesphere/cmd/controller-manager/app"
	"kubesphere.io/kubesphere/pkg/apis"
	"kubesphere.io/kubesphere/pkg/controller"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	var metricsAddr, kubeConfigPath, masterURL string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()
	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("controller-manager")

	kubeConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfigPath)
	if err != nil {
		log.Error(err, "failed to build kubeconfig")
		os.Exit(1)
	}

	stopCh := signals.SetupSignalHandler()

	log.Info("setting up manager")
	mgr, err := manager.New(kubeConfig, manager.Options{})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	log.Info("setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable add APIs to scheme")
		os.Exit(1)
	}

	log.Info("Setting up controllers")
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "unable to register controllers to the manager")
		os.Exit(1)
	}

	if err := app.AddControllers(mgr, kubeConfig, stopCh); err != nil {
		log.Error(err, "unable to register controllers to the manager")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")
	if err := mgr.Start(stopCh); err != nil {
		log.Error(err, "unable to run the manager")
		os.Exit(1)
	}

}
