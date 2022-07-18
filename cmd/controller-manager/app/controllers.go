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
	"time"

	"github.com/kubesphere/pvc-autoresizer/runners"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/kubefed/pkg/controller/util"

	"kubesphere.io/kubesphere/cmd/controller-manager/app/options"
	"kubesphere.io/kubesphere/pkg/controller/alerting"
	"kubesphere.io/kubesphere/pkg/controller/application"
	"kubesphere.io/kubesphere/pkg/controller/helm"
	"kubesphere.io/kubesphere/pkg/controller/namespace"
	"kubesphere.io/kubesphere/pkg/controller/openpitrix/helmapplication"
	"kubesphere.io/kubesphere/pkg/controller/openpitrix/helmcategory"
	"kubesphere.io/kubesphere/pkg/controller/openpitrix/helmrelease"
	"kubesphere.io/kubesphere/pkg/controller/openpitrix/helmrepo"
	"kubesphere.io/kubesphere/pkg/controller/quota"
	"kubesphere.io/kubesphere/pkg/controller/serviceaccount"
	"kubesphere.io/kubesphere/pkg/controller/user"
	"kubesphere.io/kubesphere/pkg/controller/workspace"
	"kubesphere.io/kubesphere/pkg/controller/workspacerole"
	"kubesphere.io/kubesphere/pkg/controller/workspacerolebinding"
	"kubesphere.io/kubesphere/pkg/controller/workspacetemplate"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"

	"kubesphere.io/kubesphere/pkg/controller/storage/snapshotclass"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/controller/certificatesigningrequest"
	"kubesphere.io/kubesphere/pkg/controller/cluster"
	"kubesphere.io/kubesphere/pkg/controller/clusterrolebinding"
	"kubesphere.io/kubesphere/pkg/controller/destinationrule"
	"kubesphere.io/kubesphere/pkg/controller/globalrole"
	"kubesphere.io/kubesphere/pkg/controller/globalrolebinding"
	"kubesphere.io/kubesphere/pkg/controller/group"
	"kubesphere.io/kubesphere/pkg/controller/groupbinding"
	"kubesphere.io/kubesphere/pkg/controller/job"
	"kubesphere.io/kubesphere/pkg/controller/loginrecord"
	"kubesphere.io/kubesphere/pkg/controller/network/ippool"
	"kubesphere.io/kubesphere/pkg/controller/network/nsnetworkpolicy"
	"kubesphere.io/kubesphere/pkg/controller/network/nsnetworkpolicy/provider"
	"kubesphere.io/kubesphere/pkg/controller/notification"
	"kubesphere.io/kubesphere/pkg/controller/storage/capability"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ippoolclient "kubesphere.io/kubesphere/pkg/simple/client/network/ippool"
)

var allControllers = []string{
	"user",
	"workspacetemplate",
	"workspace",
	"workspacerole",
	"workspacerolebinding",
	"namespace",

	"helmrepo",
	"helmcategory",
	"helmapplication",
	"helmapplicationversion",
	"helmrelease",
	"helm",

	"application",
	"serviceaccount",
	"resourcequota",

	"virtualservice",
	"destinationrule",
	"job",
	"storagecapability",
	"volumesnapshot",
	"pvcautoresizer",
	"workloadrestart",
	"loginrecord",
	"cluster",
	"nsnp",
	"ippool",
	"csr",

	"clusterrolebinding",

	"fedglobalrolecache",
	"globalrole",
	"fedglobalrolebindingcache",
	"globalrolebinding",

	"groupbinding",
	"group",

	"notification",
}

// setup all available controllers one by one
func addAllControllers(mgr manager.Manager, client k8s.Client, informerFactory informers.InformerFactory,
	cmOptions *options.KubeSphereControllerManagerOptions,
	stopCh <-chan struct{}) error {
	var err error

	////////////////////////////////////
	// begin init necessary informers
	////////////////////////////////////
	kubernetesInformer := informerFactory.KubernetesSharedInformerFactory()
	istioInformer := informerFactory.IstioSharedInformerFactory()
	kubesphereInformer := informerFactory.KubeSphereSharedInformerFactory()
	////////////////////////////////////
	// end informers
	////////////////////////////////////

	////////////////////////////////////
	// begin init necessary clients
	////////////////////////////////////
	kubeconfigClient := kubeconfig.NewOperator(client.Kubernetes(),
		informerFactory.KubernetesSharedInformerFactory().Core().V1().ConfigMaps().Lister(),
		client.Config())

	var devopsClient devops.Interface
	if cmOptions.DevopsOptions != nil && len(cmOptions.DevopsOptions.Host) != 0 {
		devopsClient, err = jenkins.NewDevopsClient(cmOptions.DevopsOptions)
		if err != nil {
			return fmt.Errorf("failed to connect jenkins, please check jenkins status, error: %v", err)
		}
	}

	var ldapClient ldapclient.Interface
	// when there is no ldapOption, we set ldapClient as nil, which means we don't need to sync user info into ldap.
	if cmOptions.LdapOptions != nil && len(cmOptions.LdapOptions.Host) != 0 {
		if cmOptions.LdapOptions.Host == ldapclient.FAKE_HOST { // for debug only
			ldapClient = ldapclient.NewSimpleLdap()
		} else {
			ldapClient, err = ldapclient.NewLdapClient(cmOptions.LdapOptions, stopCh)
			if err != nil {
				return fmt.Errorf("failed to connect to ldap service, please check ldap status, error: %v", err)
			}
		}
	} else {
		klog.Warning("ks-controller-manager starts without ldap provided, it will not sync user into ldap")
	}
	////////////////////////////////////
	// end init clients
	////////////////////////////////////

	////////////////////////////////////////////////////////
	// begin init controller and add to manager one by one
	////////////////////////////////////////////////////////

	// "user" controller
	if cmOptions.IsControllerEnabled("user") {
		userController := &user.Reconciler{
			MultiClusterEnabled:     cmOptions.MultiClusterOptions.Enable,
			MaxConcurrentReconciles: 4,
			LdapClient:              ldapClient,
			DevopsClient:            devopsClient,
			KubeconfigClient:        kubeconfigClient,
			AuthenticationOptions:   cmOptions.AuthenticationOptions,
		}
		addControllerWithSetup(mgr, "user", userController)
	}

	// "workspacetemplate" controller
	if cmOptions.IsControllerEnabled("workspacetemplate") {
		workspaceTemplateReconciler := &workspacetemplate.Reconciler{MultiClusterEnabled: cmOptions.MultiClusterOptions.Enable}
		addControllerWithSetup(mgr, "workspacetemplate", workspaceTemplateReconciler)
	}

	// "workspace" controller
	if cmOptions.IsControllerEnabled("workspace") {
		workspaceReconciler := &workspace.Reconciler{}
		addControllerWithSetup(mgr, "workspace", workspaceReconciler)
	}

	// "workspacerole" controller
	if cmOptions.IsControllerEnabled("workspacerole") {
		workspaceRoleReconciler := &workspacerole.Reconciler{MultiClusterEnabled: cmOptions.MultiClusterOptions.Enable}
		addControllerWithSetup(mgr, "workspacerole", workspaceRoleReconciler)
	}

	// "workspacerolebinding" controller
	if cmOptions.IsControllerEnabled("workspacerolebinding") {
		workspaceRoleBindingReconciler := &workspacerolebinding.Reconciler{MultiClusterEnabled: cmOptions.MultiClusterOptions.Enable}
		addControllerWithSetup(mgr, "workspacerolebinding", workspaceRoleBindingReconciler)
	}

	// "namespace" controller
	if cmOptions.IsControllerEnabled("namespace") {
		namespaceReconciler := &namespace.Reconciler{GatewayOptions: cmOptions.GatewayOptions}
		addControllerWithSetup(mgr, "namespace", namespaceReconciler)
	}

	// "helmrepo" controller
	if cmOptions.IsControllerEnabled("helmrepo") {
		helmRepoReconciler := &helmrepo.ReconcileHelmRepo{}
		addControllerWithSetup(mgr, "helmrepo", helmRepoReconciler)
	}

	// "helmcategory" controller
	if cmOptions.IsControllerEnabled("helmcategory") {
		helmCategoryReconciler := &helmcategory.ReconcileHelmCategory{}
		addControllerWithSetup(mgr, "helmcategory", helmCategoryReconciler)
	}

	var opS3Client s3.Interface
	if !cmOptions.OpenPitrixOptions.AppStoreConfIsEmpty() {
		opS3Client, err = s3.NewS3Client(cmOptions.OpenPitrixOptions.S3Options)
		if err != nil {
			klog.Fatalf("failed to connect to s3, please check openpitrix s3 service status, error: %v", err)
		}

		// "helmapplication" controller
		if cmOptions.IsControllerEnabled("helmapplication") {
			reconcileHelmApp := (&helmapplication.ReconcileHelmApplication{})
			addControllerWithSetup(mgr, "helmapplication", reconcileHelmApp)
		}

		// "helmapplicationversion" controller
		if cmOptions.IsControllerEnabled("helmapplicationversion") {
			reconcileHelmAppVersion := (&helmapplication.ReconcileHelmApplicationVersion{})
			addControllerWithSetup(mgr, "helmapplicationversion", reconcileHelmAppVersion)
		}
	}

	// "helmrelease" controller
	if cmOptions.IsControllerEnabled("helmrelease") {
		reconcileHelmRelease := &helmrelease.ReconcileHelmRelease{
			// nil interface is valid value.
			StorageClient:      opS3Client,
			KsFactory:          informerFactory.KubeSphereSharedInformerFactory(),
			MultiClusterEnable: cmOptions.MultiClusterOptions.Enable,
			WaitTime:           cmOptions.OpenPitrixOptions.ReleaseControllerOptions.WaitTime,
			MaxConcurrent:      cmOptions.OpenPitrixOptions.ReleaseControllerOptions.MaxConcurrent,
			StopChan:           stopCh,
		}
		addControllerWithSetup(mgr, "helmrelease", reconcileHelmRelease)
	}

	// "helm" controller
	if cmOptions.IsControllerEnabled("helm") {
		if !cmOptions.GatewayOptions.IsEmpty() {
			helmReconciler := &helm.Reconciler{GatewayOptions: cmOptions.GatewayOptions}
			addControllerWithSetup(mgr, "helm", helmReconciler)
		}
	}

	// "application" controller
	if cmOptions.IsControllerEnabled("application") {
		selector, _ := labels.Parse(cmOptions.ApplicationSelector)
		applicationReconciler := &application.ApplicationReconciler{
			Scheme:              mgr.GetScheme(),
			Client:              mgr.GetClient(),
			Mapper:              mgr.GetRESTMapper(),
			ApplicationSelector: selector,
		}
		addControllerWithSetup(mgr, "application", applicationReconciler)
	}

	// "serviceaccount" controller
	if cmOptions.IsControllerEnabled("serviceaccount") {
		saReconciler := &serviceaccount.Reconciler{}
		addControllerWithSetup(mgr, "serviceaccount", saReconciler)
	}

	// "resourcequota" controller
	if cmOptions.IsControllerEnabled("resourcequota") {
		resourceQuotaReconciler := &quota.Reconciler{
			MaxConcurrentReconciles: quota.DefaultMaxConcurrentReconciles,
			ResyncPeriod:            quota.DefaultResyncPeriod,
			InformerFactory:         informerFactory.KubernetesSharedInformerFactory(),
		}
		addControllerWithSetup(mgr, "resourcequota", resourceQuotaReconciler)
	}

	serviceMeshEnabled := cmOptions.ServiceMeshOptions != nil && len(cmOptions.ServiceMeshOptions.IstioPilotHost) != 0
	if serviceMeshEnabled {
		// "virtualservice" controller
		if cmOptions.IsControllerEnabled("virtualservice") {
			vsController := virtualservice.NewVirtualServiceController(kubernetesInformer.Core().V1().Services(),
				istioInformer.Networking().V1alpha3().VirtualServices(),
				istioInformer.Networking().V1alpha3().DestinationRules(),
				kubesphereInformer.Servicemesh().V1alpha2().Strategies(),
				client.Kubernetes(),
				client.Istio(),
				client.KubeSphere())
			addController(mgr, "virtualservice", vsController)
		}

		// "destinationrule" controller
		if cmOptions.IsControllerEnabled("destinationrule") {
			drController := destinationrule.NewDestinationRuleController(kubernetesInformer.Apps().V1().Deployments(),
				istioInformer.Networking().V1alpha3().DestinationRules(),
				kubernetesInformer.Core().V1().Services(),
				kubesphereInformer.Servicemesh().V1alpha2().ServicePolicies(),
				client.Kubernetes(),
				client.Istio(),
				client.KubeSphere())
			addController(mgr, "destinationrule", drController)
		}
	}

	// "job" controller
	if cmOptions.IsControllerEnabled("job") {
		jobController := job.NewJobController(kubernetesInformer.Batch().V1().Jobs(), client.Kubernetes())
		addController(mgr, "job", jobController)
	}

	// "storagecapability" controller
	if cmOptions.IsControllerEnabled("storagecapability") {
		storageCapabilityController := capability.NewController(
			client.Kubernetes().StorageV1().StorageClasses(),
			kubernetesInformer.Storage().V1().StorageClasses(),
			kubernetesInformer.Storage().V1().CSIDrivers(),
		)
		addController(mgr, "storagecapability", storageCapabilityController)
	}

	// "volumesnapshot" controller
	if cmOptions.IsControllerEnabled("volumesnapshot") {
		volumeSnapshotController := snapshotclass.NewController(
			kubernetesInformer.Storage().V1().StorageClasses(),
			client.Snapshot().SnapshotV1().VolumeSnapshotClasses(),
			informerFactory.SnapshotSharedInformerFactory().Snapshot().V1().VolumeSnapshotClasses(),
		)
		addController(mgr, "volumesnapshot", volumeSnapshotController)
	}

	// "pvc-autoresizer"
	monitoringOptionsEnable := cmOptions.MonitoringOptions != nil && len(cmOptions.MonitoringOptions.Endpoint) != 0
	if monitoringOptionsEnable {
		if cmOptions.IsControllerEnabled("pvc-autoresizer") {
			if err := runners.SetupIndexer(mgr, false); err != nil {
				return err
			}
			promClient, err := runners.NewPrometheusClient(cmOptions.MonitoringOptions.Endpoint)
			if err != nil {
				return err
			}
			pvcAutoResizerController := runners.NewPVCAutoresizer(
				promClient,
				mgr.GetClient(),
				ctrl.Log.WithName("pvc-autoresizer"),
				1*time.Minute,
				mgr.GetEventRecorderFor("pvc-autoresizer"),
			)
			addController(mgr, "pvcautoresizer", pvcAutoResizerController)
		}
	}

	if cmOptions.IsControllerEnabled("pvc-workload-restarter") {
		restarter := runners.NewRestarter(
			mgr.GetClient(),
			ctrl.Log.WithName("pvc-workload-restarter"),
			1*time.Minute,
			mgr.GetEventRecorderFor("pvc-workload-restarter"),
		)
		addController(mgr, "pvcworkloadrestarter", restarter)
	}

	// "loginrecord" controller
	if cmOptions.IsControllerEnabled("loginrecord") {
		loginRecordController := loginrecord.NewLoginRecordController(
			client.Kubernetes(),
			client.KubeSphere(),
			kubesphereInformer.Iam().V1alpha2().LoginRecords(),
			kubesphereInformer.Iam().V1alpha2().Users(),
			cmOptions.AuthenticationOptions.LoginHistoryRetentionPeriod,
			cmOptions.AuthenticationOptions.LoginHistoryMaximumEntries)
		addController(mgr, "loginrecord", loginRecordController)
	}

	// "csr" controller
	if cmOptions.IsControllerEnabled("csr") {
		csrController := certificatesigningrequest.NewController(client.Kubernetes(),
			kubernetesInformer.Certificates().V1().CertificateSigningRequests(),
			kubernetesInformer.Core().V1().ConfigMaps(), client.Config())
		addController(mgr, "csr", csrController)
	}

	// "clusterrolebinding" controller
	if cmOptions.IsControllerEnabled("clusterrolebinding") {
		clusterRoleBindingController := clusterrolebinding.NewController(client.Kubernetes(),
			kubernetesInformer.Rbac().V1().ClusterRoleBindings(),
			kubernetesInformer.Apps().V1().Deployments(),
			kubernetesInformer.Core().V1().Pods(),
			kubesphereInformer.Iam().V1alpha2().Users(),
			cmOptions.AuthenticationOptions.KubectlImage)
		addController(mgr, "clusterrolebinding", clusterRoleBindingController)
	}

	// "fedglobalrolecache" controller
	var fedGlobalRoleCache cache.Store
	var fedGlobalRoleCacheController cache.Controller
	if cmOptions.IsControllerEnabled("fedglobalrolecache") {
		if cmOptions.MultiClusterOptions.Enable {
			fedGlobalRoleClient, err := util.NewResourceClient(client.Config(), &iamv1alpha2.FedGlobalRoleResource)
			if err != nil {
				klog.Fatalf("Unable to create FedGlobalRole controller: %v", err)
			}
			fedGlobalRoleCache, fedGlobalRoleCacheController = util.NewResourceInformer(fedGlobalRoleClient, "",
				&iamv1alpha2.FedGlobalRoleResource, func(object runtimeclient.Object) {})
			go fedGlobalRoleCacheController.Run(stopCh)
			addSuccessfullyControllers.Insert("fedglobalrolecache")
		}
	}

	// "globalrole" controller
	if cmOptions.IsControllerEnabled("globalrole") {
		if cmOptions.MultiClusterOptions.Enable {
			globalRoleController := globalrole.NewController(client.Kubernetes(), client.KubeSphere(),
				kubesphereInformer.Iam().V1alpha2().GlobalRoles(), fedGlobalRoleCache, fedGlobalRoleCacheController)
			addController(mgr, "globalrole", globalRoleController)
		}
	}

	// "fedglobalrolebindingcache" controller
	var fedGlobalRoleBindingCache cache.Store
	var fedGlobalRoleBindingCacheController cache.Controller
	if cmOptions.IsControllerEnabled("fedglobalrolebindingcache") {
		if cmOptions.MultiClusterOptions.Enable {
			fedGlobalRoleBindingClient, err := util.NewResourceClient(client.Config(), &iamv1alpha2.FedGlobalRoleBindingResource)
			if err != nil {
				klog.Fatalf("Unable to create FedGlobalRoleBinding controller: %v", err)
			}
			fedGlobalRoleBindingCache, fedGlobalRoleBindingCacheController = util.NewResourceInformer(fedGlobalRoleBindingClient, "",
				&iamv1alpha2.FedGlobalRoleBindingResource, func(object runtimeclient.Object) {})
			go fedGlobalRoleBindingCacheController.Run(stopCh)
			addSuccessfullyControllers.Insert("fedglobalrolebindingcache")
		}
	}

	// "globalrolebinding" controller
	if cmOptions.IsControllerEnabled("globalrolebinding") {
		globalRoleBindingController := globalrolebinding.NewController(client.Kubernetes(), client.KubeSphere(),
			kubesphereInformer.Iam().V1alpha2().GlobalRoleBindings(),
			fedGlobalRoleBindingCache, fedGlobalRoleBindingCacheController,
			cmOptions.MultiClusterOptions.Enable)
		addController(mgr, "globalrolebinding", globalRoleBindingController)
	}

	// "groupbinding" controller
	if cmOptions.IsControllerEnabled("groupbinding") {
		groupBindingController := groupbinding.NewController(client.Kubernetes(), client.KubeSphere(),
			kubesphereInformer.Iam().V1alpha2().GroupBindings(),
			kubesphereInformer.Types().V1beta1().FederatedGroupBindings(),
			cmOptions.MultiClusterOptions.Enable)
		addController(mgr, "groupbinding", groupBindingController)
	}

	// "group" controller
	if cmOptions.IsControllerEnabled("group") {
		groupController := group.NewController(client.Kubernetes(), client.KubeSphere(),
			kubesphereInformer.Iam().V1alpha2().Groups(),
			kubesphereInformer.Types().V1beta1().FederatedGroups(),
			cmOptions.MultiClusterOptions.Enable)
		addController(mgr, "group", groupController)
	}

	// "cluster" controller
	if cmOptions.IsControllerEnabled("cluster") {
		if cmOptions.MultiClusterOptions.Enable {
			clusterController := cluster.NewClusterController(
				client.Kubernetes(),
				client.KubeSphere(),
				client.Config(),
				kubesphereInformer.Cluster().V1alpha1().Clusters(),
				kubesphereInformer.Iam().V1alpha2().Users().Lister(),
				cmOptions.MultiClusterOptions.ClusterControllerResyncPeriod,
				cmOptions.MultiClusterOptions.HostClusterName,
			)
			addController(mgr, "cluster", clusterController)
		}
	}

	// "nsnp" controller
	if cmOptions.IsControllerEnabled("nsnp") {
		if cmOptions.NetworkOptions.EnableNetworkPolicy {
			nsnpProvider, err := provider.NewNsNetworkPolicyProvider(client.Kubernetes(), kubernetesInformer.Networking().V1().NetworkPolicies())
			if err != nil {
				klog.Fatalf("Unable to create NSNetworkPolicy controller: %v", err)
			}

			nsnpController := nsnetworkpolicy.NewNSNetworkPolicyController(client.Kubernetes(),
				client.KubeSphere().NetworkV1alpha1(),
				kubesphereInformer.Network().V1alpha1().NamespaceNetworkPolicies(),
				kubernetesInformer.Core().V1().Services(),
				kubernetesInformer.Core().V1().Nodes(),
				kubesphereInformer.Tenant().V1alpha1().Workspaces(),
				kubernetesInformer.Core().V1().Namespaces(), nsnpProvider, cmOptions.NetworkOptions.NSNPOptions)
			addController(mgr, "nsnp", nsnpController)
		}
	}

	// "ippool" controller
	if cmOptions.IsControllerEnabled("ippool") {
		ippoolProvider := ippoolclient.NewProvider(kubernetesInformer, client.KubeSphere(), client.Kubernetes(),
			cmOptions.NetworkOptions.IPPoolType, cmOptions.KubernetesOptions)
		if ippoolProvider != nil {
			ippoolController := ippool.NewIPPoolController(kubesphereInformer, kubernetesInformer, client.Kubernetes(),
				client.KubeSphere(), ippoolProvider)
			addController(mgr, "ippool", ippoolController)
		}
	}

	// "notification" controller
	if cmOptions.IsControllerEnabled("notification") {
		if cmOptions.MultiClusterOptions.Enable {
			notificationController, err := notification.NewController(client.Kubernetes(), mgr.GetClient(), mgr.GetCache())
			if err != nil {
				klog.Fatalf("Unable to create Notification controller: %v", err)
			}
			addController(mgr, "notification", notificationController)
		}
	}

	// controllers for alerting
	alertingOptionsEnable := cmOptions.AlertingOptions != nil && (cmOptions.AlertingOptions.PrometheusEndpoint != "" || cmOptions.AlertingOptions.ThanosRulerEndpoint != "")
	if alertingOptionsEnable {
		// "rulegroup" controller
		if cmOptions.IsControllerEnabled("rulegroup") {
			rulegroupReconciler := &alerting.RuleGroupReconciler{}
			addControllerWithSetup(mgr, "rulegroup", rulegroupReconciler)
		}
		// "clusterrulegroup" controller
		if cmOptions.IsControllerEnabled("clusterrulegroup") {
			clusterrulegroupReconciler := &alerting.ClusterRuleGroupReconciler{}
			addControllerWithSetup(mgr, "clusterrulegroup", clusterrulegroupReconciler)
		}
		// "globalrulegroup" controller
		if cmOptions.IsControllerEnabled("globalrulegroup") {
			globalrulegroupReconciler := &alerting.GlobalRuleGroupReconciler{}
			addControllerWithSetup(mgr, "globalrulegroup", globalrulegroupReconciler)
		}
	}

	// log all controllers process result
	for _, name := range allControllers {
		if cmOptions.IsControllerEnabled(name) {
			if addSuccessfullyControllers.Has(name) {
				klog.Infof("%s controller is enabled and added successfully.", name)
			} else {
				klog.Infof("%s controller is enabled but is not going to run due to its dependent component being disabled.", name)
			}
		} else {
			klog.Infof("%s controller is disabled by controller selectors.", name)
		}
	}

	return nil
}

var addSuccessfullyControllers = sets.NewString()

type setupableController interface {
	SetupWithManager(mgr ctrl.Manager) error
}

func addControllerWithSetup(mgr manager.Manager, name string, controller setupableController) {
	if err := controller.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create %v controller: %v", name, err)
	}
	addSuccessfullyControllers.Insert(name)
}

func addController(mgr manager.Manager, name string, controller manager.Runnable) {
	if err := mgr.Add(controller); err != nil {
		klog.Fatalf("Unable to create %v controller: %v", name, err)
	}
	addSuccessfullyControllers.Insert(name)
}
