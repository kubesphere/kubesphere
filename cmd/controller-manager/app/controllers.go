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
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha2"
	"kubesphere.io/kubesphere/pkg/controller/application"
	"kubesphere.io/kubesphere/pkg/controller/certificatesigningrequest"
	"kubesphere.io/kubesphere/pkg/controller/cluster"
	"kubesphere.io/kubesphere/pkg/controller/clusterrolebinding"
	"kubesphere.io/kubesphere/pkg/controller/destinationrule"
	"kubesphere.io/kubesphere/pkg/controller/devopscredential"
	"kubesphere.io/kubesphere/pkg/controller/devopsproject"
	"kubesphere.io/kubesphere/pkg/controller/globalrole"
	"kubesphere.io/kubesphere/pkg/controller/globalrolebinding"
	"kubesphere.io/kubesphere/pkg/controller/job"
	"kubesphere.io/kubesphere/pkg/controller/network/nsnetworkpolicy"
	"kubesphere.io/kubesphere/pkg/controller/network/provider"
	"kubesphere.io/kubesphere/pkg/controller/pipeline"
	"kubesphere.io/kubesphere/pkg/controller/s2ibinary"
	"kubesphere.io/kubesphere/pkg/controller/s2irun"
	"kubesphere.io/kubesphere/pkg/controller/storage/capability"
	"kubesphere.io/kubesphere/pkg/controller/storage/expansion"
	"kubesphere.io/kubesphere/pkg/controller/user"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice"
	"kubesphere.io/kubesphere/pkg/controller/workspacerole"
	"kubesphere.io/kubesphere/pkg/controller/workspacerolebinding"
	"kubesphere.io/kubesphere/pkg/controller/workspacetemplate"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/kubefed/pkg/controller/util"
)

func addControllers(
	mgr manager.Manager,
	client k8s.Client,
	informerFactory informers.InformerFactory,
	devopsClient devops.Interface,
	s3Client s3.Interface,
	openpitrixClient openpitrix.Client,
	multiClusterEnabled bool,
	networkPolicyEnabled bool,
	serviceMeshEnabled bool,
	stopCh <-chan struct{}) error {

	kubernetesInformer := informerFactory.KubernetesSharedInformerFactory()
	istioInformer := informerFactory.IstioSharedInformerFactory()
	kubesphereInformer := informerFactory.KubeSphereSharedInformerFactory()
	applicationInformer := informerFactory.ApplicationSharedInformerFactory()

	var vsController, drController manager.Runnable
	if serviceMeshEnabled {
		vsController = virtualservice.NewVirtualServiceController(kubernetesInformer.Core().V1().Services(),
			istioInformer.Networking().V1alpha3().VirtualServices(),
			istioInformer.Networking().V1alpha3().DestinationRules(),
			kubesphereInformer.Servicemesh().V1alpha2().Strategies(),
			client.Kubernetes(),
			client.Istio(),
			client.KubeSphere())

		drController = destinationrule.NewDestinationRuleController(kubernetesInformer.Apps().V1().Deployments(),
			istioInformer.Networking().V1alpha3().DestinationRules(),
			kubernetesInformer.Core().V1().Services(),
			kubesphereInformer.Servicemesh().V1alpha2().ServicePolicies(),
			client.Kubernetes(),
			client.Istio(),
			client.KubeSphere())
	}

	apController := application.NewApplicationController(kubernetesInformer.Core().V1().Services(),
		kubernetesInformer.Apps().V1().Deployments(),
		kubernetesInformer.Apps().V1().StatefulSets(),
		kubesphereInformer.Servicemesh().V1alpha2().Strategies(),
		kubesphereInformer.Servicemesh().V1alpha2().ServicePolicies(),
		applicationInformer.App().V1beta1().Applications(),
		client.Kubernetes(),
		client.Application())

	jobController := job.NewJobController(kubernetesInformer.Batch().V1().Jobs(), client.Kubernetes())

	var s2iBinaryController, s2iRunController, devopsProjectController, devopsPipelineController, devopsCredentialController manager.Runnable
	if devopsClient != nil {
		s2iBinaryController = s2ibinary.NewController(client.Kubernetes(),
			client.KubeSphere(),
			kubesphereInformer.Devops().V1alpha1().S2iBinaries(),
			s3Client,
		)

		s2iRunController = s2irun.NewS2iRunController(client.Kubernetes(),
			client.KubeSphere(),
			kubesphereInformer.Devops().V1alpha1().S2iBinaries(),
			kubesphereInformer.Devops().V1alpha1().S2iRuns())

		devopsProjectController = devopsproject.NewController(client.Kubernetes(),
			client.KubeSphere(), devopsClient,
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
			informerFactory.KubeSphereSharedInformerFactory().Devops().V1alpha3().DevOpsProjects(),
		)

		devopsPipelineController = pipeline.NewController(client.Kubernetes(),
			client.KubeSphere(),
			devopsClient,
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
			informerFactory.KubeSphereSharedInformerFactory().Devops().V1alpha3().Pipelines())

		devopsCredentialController = devopscredential.NewController(client.Kubernetes(),
			devopsClient,
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Namespaces(),
			informerFactory.KubernetesSharedInformerFactory().Core().V1().Secrets())

	}

	storageCapabilityController := capability.NewController(
		client.Kubernetes(),
		client.KubeSphere(),
		kubernetesInformer.Storage().V1().StorageClasses(),
		informerFactory.SnapshotSharedInformerFactory().Snapshot().V1beta1().VolumeSnapshotClasses(),
		kubesphereInformer.Storage().V1alpha1().StorageClassCapabilities(),
		func(storageClassProvisioner string) string {
			return fmt.Sprintf(capability.CSIAddressFormat, storageClassProvisioner)
		},
	)

	volumeExpansionController := expansion.NewVolumeExpansionController(
		client.Kubernetes(),
		kubernetesInformer.Core().V1().PersistentVolumeClaims(),
		kubernetesInformer.Storage().V1().StorageClasses(),
		kubernetesInformer.Core().V1().Pods(),
		kubernetesInformer.Apps().V1().Deployments(),
		kubernetesInformer.Apps().V1().ReplicaSets(),
		kubernetesInformer.Apps().V1().StatefulSets())

	var fedUserCache, fedGlobalRoleBindingCache, fedGlobalRoleCache,
		fedWorkspaceCache, fedWorkspaceRoleCache, fedWorkspaceRoleBindingCache cache.Store
	var fedUserCacheController, fedGlobalRoleBindingCacheController, fedGlobalRoleCacheController,
		fedWorkspaceCacheController, fedWorkspaceRoleCacheController, fedWorkspaceRoleBindingCacheController cache.Controller

	if multiClusterEnabled {

		fedUserClient, err := util.NewResourceClient(client.Config(), &iamv1alpha2.FedUserResource)
		if err != nil {
			klog.Error(err)
			return err
		}
		fedGlobalRoleClient, err := util.NewResourceClient(client.Config(), &iamv1alpha2.FedGlobalRoleResource)
		if err != nil {
			klog.Error(err)
			return err
		}
		fedGlobalRoleBindingClient, err := util.NewResourceClient(client.Config(), &iamv1alpha2.FedGlobalRoleBindingResource)
		if err != nil {
			klog.Error(err)
			return err
		}
		fedWorkspaceClient, err := util.NewResourceClient(client.Config(), &tenantv1alpha2.FedWorkspaceResource)
		if err != nil {
			klog.Error(err)
			return err
		}
		fedWorkspaceRoleClient, err := util.NewResourceClient(client.Config(), &iamv1alpha2.FedWorkspaceRoleResource)
		if err != nil {
			klog.Error(err)
			return err
		}
		fedWorkspaceRoleBindingClient, err := util.NewResourceClient(client.Config(), &iamv1alpha2.FedWorkspaceRoleBindingResource)
		if err != nil {
			klog.Error(err)
			return err
		}

		fedUserCache, fedUserCacheController = util.NewResourceInformer(fedUserClient, "", &iamv1alpha2.FedUserResource, func(object runtime.Object) {})
		fedGlobalRoleCache, fedGlobalRoleCacheController = util.NewResourceInformer(fedGlobalRoleClient, "", &iamv1alpha2.FedGlobalRoleResource, func(object runtime.Object) {})
		fedGlobalRoleBindingCache, fedGlobalRoleBindingCacheController = util.NewResourceInformer(fedGlobalRoleBindingClient, "", &iamv1alpha2.FedGlobalRoleBindingResource, func(object runtime.Object) {})
		fedWorkspaceCache, fedWorkspaceCacheController = util.NewResourceInformer(fedWorkspaceClient, "", &tenantv1alpha2.FedWorkspaceResource, func(object runtime.Object) {})
		fedWorkspaceRoleCache, fedWorkspaceRoleCacheController = util.NewResourceInformer(fedWorkspaceRoleClient, "", &iamv1alpha2.FedWorkspaceRoleResource, func(object runtime.Object) {})
		fedWorkspaceRoleBindingCache, fedWorkspaceRoleBindingCacheController = util.NewResourceInformer(fedWorkspaceRoleBindingClient, "", &iamv1alpha2.FedWorkspaceRoleBindingResource, func(object runtime.Object) {})

		go fedUserCacheController.Run(stopCh)
		go fedGlobalRoleCacheController.Run(stopCh)
		go fedGlobalRoleBindingCacheController.Run(stopCh)
		go fedWorkspaceCacheController.Run(stopCh)
		go fedWorkspaceRoleCacheController.Run(stopCh)
		go fedWorkspaceRoleBindingCacheController.Run(stopCh)
	}

	userController := user.NewController(client.Kubernetes(), client.KubeSphere(), client.Config(),
		kubesphereInformer.Iam().V1alpha2().Users(),
		fedUserCache, fedUserCacheController, kubernetesInformer.Core().V1().ConfigMaps(), multiClusterEnabled)

	csrController := certificatesigningrequest.NewController(client.Kubernetes(), kubernetesInformer.Certificates().V1beta1().CertificateSigningRequests(),
		kubernetesInformer.Core().V1().ConfigMaps(), client.Config())

	clusterRoleBindingController := clusterrolebinding.NewController(client.Kubernetes(),
		kubernetesInformer.Rbac().V1().ClusterRoleBindings(), kubernetesInformer.Apps().V1().Deployments(),
		kubernetesInformer.Core().V1().Pods(), kubesphereInformer.Iam().V1alpha2().Users())

	globalRoleController := globalrole.NewController(client.Kubernetes(), client.KubeSphere(),
		kubesphereInformer.Iam().V1alpha2().GlobalRoles(), fedGlobalRoleCache, fedGlobalRoleCacheController)

	workspaceRoleController := workspacerole.NewController(client.Kubernetes(), client.KubeSphere(),
		kubesphereInformer.Iam().V1alpha2().WorkspaceRoles(), fedWorkspaceRoleCache, fedWorkspaceRoleCacheController)

	globalRoleBindingController := globalrolebinding.NewController(client.Kubernetes(), client.KubeSphere(),
		kubesphereInformer.Iam().V1alpha2().GlobalRoleBindings(), fedGlobalRoleBindingCache, fedGlobalRoleBindingCacheController, multiClusterEnabled)

	workspaceRoleBindingController := workspacerolebinding.NewController(client.Kubernetes(), client.KubeSphere(),
		kubesphereInformer.Iam().V1alpha2().WorkspaceRoleBindings(), fedWorkspaceRoleBindingCache, fedWorkspaceRoleBindingCacheController)

	workspaceTemplateController := workspacetemplate.NewController(client.Kubernetes(), client.KubeSphere(),
		kubesphereInformer.Tenant().V1alpha2().WorkspaceTemplates(), kubesphereInformer.Tenant().V1alpha1().Workspaces(),
		kubesphereInformer.Iam().V1alpha2().RoleBases(), kubesphereInformer.Iam().V1alpha2().WorkspaceRoles(),
		fedWorkspaceCache, fedWorkspaceCacheController, multiClusterEnabled)

	var clusterController manager.Runnable
	if multiClusterEnabled {
		clusterController = cluster.NewClusterController(
			client.Kubernetes(),
			client.Config(),
			kubesphereInformer.Cluster().V1alpha1().Clusters(),
			client.KubeSphere().ClusterV1alpha1().Clusters(),
			openpitrixClient)
	}

	var nsnpController manager.Runnable
	if networkPolicyEnabled {
		nsnpProvider, err := provider.NewNsNetworkPolicyProvider(client.Kubernetes(), kubernetesInformer.Networking().V1().NetworkPolicies())
		if err != nil {
			return err
		}

		nsnpController = nsnetworkpolicy.NewNSNetworkPolicyController(client.Kubernetes(),
			client.KubeSphere().NetworkV1alpha1(), kubesphereInformer.Network().V1alpha1().NamespaceNetworkPolicies(),
			kubernetesInformer.Core().V1().Services(), kubernetesInformer.Core().V1().Nodes(),
			kubesphereInformer.Tenant().V1alpha1().Workspaces(),
			kubernetesInformer.Core().V1().Namespaces(), nsnpProvider)
	}

	controllers := map[string]manager.Runnable{
		"virtualservice-controller":     vsController,
		"destinationrule-controller":    drController,
		"application-controller":        apController,
		"job-controller":                jobController,
		"s2ibinary-controller":          s2iBinaryController,
		"s2irun-controller":             s2iRunController,
		"volumeexpansion-controller":    volumeExpansionController,
		"user-controller":               userController,
		"cluster-controller":            clusterController,
		"nsnp-controller":               nsnpController,
		"csr-controller":                csrController,
		"clusterrolebinding-controller": clusterRoleBindingController,
		"globalrolebinding-controller":  globalRoleBindingController,
		"workspacetemplate-controller":  workspaceTemplateController,
	}

	if devopsClient != nil {
		controllers["pipeline-controller"] = devopsPipelineController
		controllers["devopsprojects-controller"] = devopsProjectController
		controllers["devopscredential-controller"] = devopsCredentialController
	}

	if storageCapabilityController.IsValidKubernetesVersion() {
		controllers["storagecapability-controller"] = storageCapabilityController
	}

	if multiClusterEnabled {
		controllers["globalrole-controller"] = globalRoleController
		controllers["workspacerole-controller"] = workspaceRoleController
		controllers["workspacerolebinding-controller"] = workspaceRoleBindingController
	}

	for name, ctrl := range controllers {
		if ctrl == nil {
			klog.V(4).Infof("%s is not going to run due to dependent component disabled.", name)
			continue
		}

		if err := mgr.Add(ctrl); err != nil {
			klog.Error(err, "add controller to manager failed", "name", name)
			return err
		}
	}

	return nil
}
