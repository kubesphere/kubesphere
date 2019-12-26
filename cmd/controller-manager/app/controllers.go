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
package app

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubesphere.io/kubesphere/pkg/controller/application"
	"kubesphere.io/kubesphere/pkg/controller/destinationrule"
	"kubesphere.io/kubesphere/pkg/controller/job"
	"kubesphere.io/kubesphere/pkg/controller/s2ibinary"
	"kubesphere.io/kubesphere/pkg/controller/s2irun"
	"kubesphere.io/kubesphere/pkg/controller/storage/expansion"

	//"kubesphere.io/kubesphere/pkg/controller/job"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	applicationclientset "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned"
	applicationinformers "github.com/kubernetes-sigs/application/pkg/client/informers/externalversions"
	s2iclientset "github.com/kubesphere/s2ioperator/pkg/client/clientset/versioned"
	s2iinformers "github.com/kubesphere/s2ioperator/pkg/client/informers/externalversions"
	istioclientset "istio.io/client-go/pkg/clientset/versioned"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	kubesphereclientset "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	kubesphereinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
)

const defaultResync = 600 * time.Second

var log = logf.Log.WithName("controller-manager")

func AddControllers(mgr manager.Manager, cfg *rest.Config, stopCh <-chan struct{}) error {

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "building kubernetes client failed")
	}

	istioclient, err := istioclientset.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "create istio client failed")
		return err
	}

	applicationClient, err := applicationclientset.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "create application client failed")
		return err
	}
	s2iclient, err := s2iclientset.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "create s2i client failed")
		return err
	}
	kubesphereclient, err := kubesphereclientset.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "create kubesphere client failed")
		return err
	}

	informerFactory := informers.NewSharedInformerFactory(kubeClient, defaultResync)
	istioInformer := istioinformers.NewSharedInformerFactory(istioclient, defaultResync)
	applicationInformer := applicationinformers.NewSharedInformerFactory(applicationClient, defaultResync)
	s2iInformer := s2iinformers.NewSharedInformerFactory(s2iclient, defaultResync)

	kubesphereInformer := kubesphereinformers.NewSharedInformerFactory(kubesphereclient, defaultResync)

	vsController := virtualservice.NewVirtualServiceController(informerFactory.Core().V1().Services(),
		istioInformer.Networking().V1alpha3().VirtualServices(),
		istioInformer.Networking().V1alpha3().DestinationRules(),
		kubesphereInformer.Servicemesh().V1alpha2().Strategies(),
		kubeClient,
		istioclient,
		kubesphereclient)

	drController := destinationrule.NewDestinationRuleController(informerFactory.Apps().V1().Deployments(),
		istioInformer.Networking().V1alpha3().DestinationRules(),
		informerFactory.Core().V1().Services(),
		kubesphereInformer.Servicemesh().V1alpha2().ServicePolicies(),
		kubeClient,
		istioclient,
		kubesphereclient)

	apController := application.NewApplicationController(informerFactory.Core().V1().Services(),
		informerFactory.Apps().V1().Deployments(),
		informerFactory.Apps().V1().StatefulSets(),
		kubesphereInformer.Servicemesh().V1alpha2().Strategies(),
		kubesphereInformer.Servicemesh().V1alpha2().ServicePolicies(),
		applicationInformer.App().V1beta1().Applications(),
		kubeClient,
		applicationClient)

	jobController := job.NewJobController(informerFactory.Batch().V1().Jobs(), kubeClient)

	s2iBinaryController := s2ibinary.NewController(kubesphereclient,
		kubeClient,
		kubesphereInformer.Devops().V1alpha1().S2iBinaries())

	s2iRunController := s2irun.NewController(kubesphereclient, s2iclient, kubeClient,
		kubesphereInformer.Devops().V1alpha1().S2iBinaries(),
		s2iInformer.Devops().V1alpha1().S2iRuns())

	volumeExpansionController := expansion.NewVolumeExpansionController(
		kubeClient,
		informerFactory.Core().V1().PersistentVolumeClaims(),
		informerFactory.Storage().V1().StorageClasses(),
		informerFactory.Core().V1().Pods(),
		informerFactory.Apps().V1().Deployments(),
		informerFactory.Apps().V1().ReplicaSets(),
		informerFactory.Apps().V1().StatefulSets())

	kubesphereInformer.Start(stopCh)
	istioInformer.Start(stopCh)
	informerFactory.Start(stopCh)
	applicationInformer.Start(stopCh)
	s2iInformer.Start(stopCh)

	controllers := map[string]manager.Runnable{
		"virtualservice-controller":  vsController,
		"destinationrule-controller": drController,
		"application-controller":     apController,
		"job-controller":             jobController,
		"s2ibinary-controller":       s2iBinaryController,
		"s2irun-controller":          s2iRunController,
		"volumeexpansion-controller": volumeExpansionController,
	}

	for name, ctrl := range controllers {
		err = mgr.Add(ctrl)
		if err != nil {
			log.Error(err, "add controller to manager failed", "name", name)
			return err
		}
	}

	return nil
}
