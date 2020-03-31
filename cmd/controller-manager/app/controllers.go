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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/controller/application"
	"kubesphere.io/kubesphere/pkg/controller/destinationrule"
	"kubesphere.io/kubesphere/pkg/controller/job"
	"kubesphere.io/kubesphere/pkg/controller/s2ibinary"
	"kubesphere.io/kubesphere/pkg/controller/s2irun"
	"kubesphere.io/kubesphere/pkg/controller/storage/expansion"
	"kubesphere.io/kubesphere/pkg/controller/user"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func AddControllers(
	mgr manager.Manager,
	client k8s.Client,
	informerFactory informers.InformerFactory,
	stopCh <-chan struct{}) error {

	kubernetesInformer := informerFactory.KubernetesSharedInformerFactory()
	istioInformer := informerFactory.IstioSharedInformerFactory()
	kubesphereInformer := informerFactory.KubeSphereSharedInformerFactory()
	applicationInformer := informerFactory.ApplicationSharedInformerFactory()

	vsController := virtualservice.NewVirtualServiceController(kubernetesInformer.Core().V1().Services(),
		istioInformer.Networking().V1alpha3().VirtualServices(),
		istioInformer.Networking().V1alpha3().DestinationRules(),
		kubesphereInformer.Servicemesh().V1alpha2().Strategies(),
		client.Kubernetes(),
		client.Istio(),
		client.KubeSphere())

	drController := destinationrule.NewDestinationRuleController(kubernetesInformer.Apps().V1().Deployments(),
		istioInformer.Networking().V1alpha3().DestinationRules(),
		kubernetesInformer.Core().V1().Services(),
		kubesphereInformer.Servicemesh().V1alpha2().ServicePolicies(),
		client.Kubernetes(),
		client.Istio(),
		client.KubeSphere())

	apController := application.NewApplicationController(kubernetesInformer.Core().V1().Services(),
		kubernetesInformer.Apps().V1().Deployments(),
		kubernetesInformer.Apps().V1().StatefulSets(),
		kubesphereInformer.Servicemesh().V1alpha2().Strategies(),
		kubesphereInformer.Servicemesh().V1alpha2().ServicePolicies(),
		applicationInformer.App().V1beta1().Applications(),
		client.Kubernetes(),
		client.Application())

	jobController := job.NewJobController(kubernetesInformer.Batch().V1().Jobs(), client.Kubernetes())

	s2iBinaryController := s2ibinary.NewController(client.KubeSphere(),
		client.Kubernetes(),
		kubesphereInformer.Devops().V1alpha1().S2iBinaries())

	s2iRunController := s2irun.NewS2iRunController(client.KubeSphere(),
		client.Kubernetes(),
		kubesphereInformer.Devops().V1alpha1().S2iBinaries(),
		kubesphereInformer.Devops().V1alpha1().S2iRuns())

	volumeExpansionController := expansion.NewVolumeExpansionController(
		client.Kubernetes(),
		kubernetesInformer.Core().V1().PersistentVolumeClaims(),
		kubernetesInformer.Storage().V1().StorageClasses(),
		kubernetesInformer.Core().V1().Pods(),
		kubernetesInformer.Apps().V1().Deployments(),
		kubernetesInformer.Apps().V1().ReplicaSets(),
		kubernetesInformer.Apps().V1().StatefulSets())

	userController := user.NewController(
		client.Kubernetes(),
		client.KubeSphere(),
		kubesphereInformer.Iam().V1alpha2().Users())

	controllers := map[string]manager.Runnable{
		"virtualservice-controller":  vsController,
		"destinationrule-controller": drController,
		"application-controller":     apController,
		"job-controller":             jobController,
		"s2ibinary-controller":       s2iBinaryController,
		"s2irun-controller":          s2iRunController,
		"volumeexpansion-controller": volumeExpansionController,
		"user-controller":            userController,
	}

	for name, ctrl := range controllers {
		if err := mgr.Add(ctrl); err != nil {
			klog.Error(err, "add controller to manager failed", "name", name)
			return err
		}
	}

	return nil
}
