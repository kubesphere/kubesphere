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
	"context"
	"fmt"
	"os"

	"kubesphere.io/kubesphere/pkg/models/kubeconfig"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"kubesphere.io/kubesphere/cmd/controller-manager/app/options"
	"kubesphere.io/kubesphere/pkg/apis"
	controllerconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/controller/application"
	"kubesphere.io/kubesphere/pkg/controller/helm"
	"kubesphere.io/kubesphere/pkg/controller/namespace"
	"kubesphere.io/kubesphere/pkg/controller/network/webhooks"
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
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/metrics"
	"kubesphere.io/kubesphere/pkg/utils/term"
	"kubesphere.io/kubesphere/pkg/version"
)

func NewControllerManagerCommand() *cobra.Command {
	s := options.NewKubeSphereControllerManagerOptions()
	conf, err := controllerconfig.TryLoadFromDisk()
	if err == nil {
		// make sure LeaderElection is not nil
		s = &options.KubeSphereControllerManagerOptions{
			KubernetesOptions:     conf.KubernetesOptions,
			DevopsOptions:         conf.DevopsOptions,
			S3Options:             conf.S3Options,
			AuthenticationOptions: conf.AuthenticationOptions,
			LdapOptions:           conf.LdapOptions,
			OpenPitrixOptions:     conf.OpenPitrixOptions,
			NetworkOptions:        conf.NetworkOptions,
			MultiClusterOptions:   conf.MultiClusterOptions,
			ServiceMeshOptions:    conf.ServiceMeshOptions,
			GatewayOptions:        conf.GatewayOptions,
			LeaderElection:        s.LeaderElection,
			LeaderElect:           s.LeaderElect,
			WebhookCertDir:        s.WebhookCertDir,
		}
	} else {
		klog.Fatal("Failed to load configuration from disk", err)
	}

	cmd := &cobra.Command{
		Use:  "controller-manager",
		Long: `KubeSphere controller manager is a daemon that`,
		Run: func(cmd *cobra.Command, args []string) {
			if errs := s.Validate(); len(errs) != 0 {
				klog.Error(utilerrors.NewAggregate(errs))
				os.Exit(1)
			}

			if err = run(s, signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()

	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of KubeSphere controller-manager",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Get())
		},
	}

	cmd.AddCommand(versionCmd)

	return cmd
}

func run(s *options.KubeSphereControllerManagerOptions, ctx context.Context) error {

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		klog.Errorf("Failed to create kubernetes clientset %v", err)
		return err
	}

	var devopsClient devops.Interface
	if s.DevopsOptions != nil && len(s.DevopsOptions.Host) != 0 {
		devopsClient, err = jenkins.NewDevopsClient(s.DevopsOptions)
		if err != nil {
			return fmt.Errorf("failed to connect jenkins, please check jenkins status, error: %v", err)
		}
	}

	var ldapClient ldapclient.Interface
	// when there is no ldapOption, we set ldapClient as nil, which means we don't need to sync user info into ldap.
	if s.LdapOptions != nil && len(s.LdapOptions.Host) != 0 {
		if s.LdapOptions.Host == ldapclient.FAKE_HOST { // for debug only
			ldapClient = ldapclient.NewSimpleLdap()
		} else {
			ldapClient, err = ldapclient.NewLdapClient(s.LdapOptions, ctx.Done())
			if err != nil {
				return fmt.Errorf("failed to connect to ldap service, please check ldap status, error: %v", err)
			}
		}
	} else {
		klog.Warning("ks-controller-manager starts without ldap provided, it will not sync user into ldap")
	}

	var s3Client s3.Interface
	if s.S3Options != nil && len(s.S3Options.Endpoint) != 0 {
		s3Client, err = s3.NewS3Client(s.S3Options)
		if err != nil {
			return fmt.Errorf("failed to connect to s3, please check s3 service status, error: %v", err)
		}
	}

	informerFactory := informers.NewInformerFactories(
		kubernetesClient.Kubernetes(),
		kubernetesClient.KubeSphere(),
		kubernetesClient.Istio(),
		kubernetesClient.Snapshot(),
		kubernetesClient.ApiExtensions(),
		kubernetesClient.Prometheus())

	mgrOptions := manager.Options{
		CertDir: s.WebhookCertDir,
		Port:    8443,
	}

	if s.LeaderElect {
		mgrOptions = manager.Options{
			CertDir:                 s.WebhookCertDir,
			Port:                    8443,
			LeaderElection:          s.LeaderElect,
			LeaderElectionNamespace: "kubesphere-system",
			LeaderElectionID:        "ks-controller-manager-leader-election",
			LeaseDuration:           &s.LeaderElection.LeaseDuration,
			RetryPeriod:             &s.LeaderElection.RetryPeriod,
			RenewDeadline:           &s.LeaderElection.RenewDeadline,
		}
	}

	klog.V(0).Info("setting up manager")
	ctrl.SetLogger(klogr.New())
	// Use 8443 instead of 443 cause we need root permission to bind port 443
	mgr, err := manager.New(kubernetesClient.Config(), mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}

	if err = apis.AddToScheme(mgr.GetScheme()); err != nil {
		klog.Fatalf("unable add APIs to scheme: %v", err)
	}

	// register common meta types into schemas.
	metav1.AddToGroupVersion(mgr.GetScheme(), metav1.SchemeGroupVersion)

	kubeconfigClient := kubeconfig.NewOperator(kubernetesClient.Kubernetes(),
		informerFactory.KubernetesSharedInformerFactory().Core().V1().ConfigMaps().Lister(),
		kubernetesClient.Config())
	userController := user.Reconciler{
		MultiClusterEnabled:     s.MultiClusterOptions.Enable,
		MaxConcurrentReconciles: 4,
		LdapClient:              ldapClient,
		DevopsClient:            devopsClient,
		KubeconfigClient:        kubeconfigClient,
		AuthenticationOptions:   s.AuthenticationOptions,
	}

	if err = userController.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create user controller: %v", err)
	}

	workspaceTemplateReconciler := &workspacetemplate.Reconciler{MultiClusterEnabled: s.MultiClusterOptions.Enable}
	if err = workspaceTemplateReconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create workspace template controller: %v", err)
	}

	workspaceReconciler := &workspace.Reconciler{}
	if err = workspaceReconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create workspace controller: %v", err)
	}

	workspaceRoleReconciler := &workspacerole.Reconciler{MultiClusterEnabled: s.MultiClusterOptions.Enable}
	if err = workspaceRoleReconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create workspace role controller: %v", err)
	}

	workspaceRoleBindingReconciler := &workspacerolebinding.Reconciler{MultiClusterEnabled: s.MultiClusterOptions.Enable}
	if err = workspaceRoleBindingReconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create workspace role binding controller: %v", err)
	}

	namespaceReconciler := &namespace.Reconciler{}
	if err = namespaceReconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create namespace controller: %v", err)
	}

	err = helmrepo.Add(mgr)
	if err != nil {
		klog.Fatal("Unable to create helm repo controller")
	}

	err = helmcategory.Add(mgr)
	if err != nil {
		klog.Fatal("Unable to create helm category controller")
	}

	var opS3Client s3.Interface
	if !s.OpenPitrixOptions.AppStoreConfIsEmpty() {
		opS3Client, err = s3.NewS3Client(s.OpenPitrixOptions.S3Options)
		if err != nil {
			klog.Fatalf("failed to connect to s3, please check openpitrix s3 service status, error: %v", err)
		}
		err = (&helmapplication.ReconcileHelmApplication{}).SetupWithManager(mgr)
		if err != nil {
			klog.Fatalf("Unable to create helm application controller, error: %s", err)
		}

		err = (&helmapplication.ReconcileHelmApplicationVersion{}).SetupWithManager(mgr)
		if err != nil {
			klog.Fatalf("Unable to create helm application version controller, error: %s ", err)
		}
	}

	err = (&helmrelease.ReconcileHelmRelease{
		// nil interface is valid value.
		StorageClient:      opS3Client,
		KsFactory:          informerFactory.KubeSphereSharedInformerFactory(),
		MultiClusterEnable: s.MultiClusterOptions.Enable,
		WaitTime:           s.OpenPitrixOptions.ReleaseControllerOptions.WaitTime,
		MaxConcurrent:      s.OpenPitrixOptions.ReleaseControllerOptions.MaxConcurrent,
		StopChan:           ctx.Done(),
	}).SetupWithManager(mgr)

	if err != nil {
		klog.Fatalf("Unable to create helm release controller, error: %s", err)
	}

	selector, _ := labels.Parse(s.ApplicationSelector)
	applicationReconciler := &application.ApplicationReconciler{
		Scheme:              mgr.GetScheme(),
		Client:              mgr.GetClient(),
		Mapper:              mgr.GetRESTMapper(),
		ApplicationSelector: selector,
	}
	if err = applicationReconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create application controller: %v", err)
	}

	saReconciler := &serviceaccount.Reconciler{}
	if err = saReconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create ServiceAccount controller: %v", err)
	}

	resourceQuotaReconciler := quota.Reconciler{}
	if err := resourceQuotaReconciler.SetupWithManager(mgr, quota.DefaultMaxConcurrentReconciles, quota.DefaultResyncPeriod, informerFactory.KubernetesSharedInformerFactory()); err != nil {
		klog.Fatalf("Unable to create ResourceQuota controller: %v", err)
	}

	if !s.GatewayOptions.IsEmpty() {
		helmReconciler := helm.Reconciler{GatewayOptions: s.GatewayOptions}
		if err := helmReconciler.SetupWithManager(mgr); err != nil {
			klog.Fatalf("Unable to create helm controller: %v", err)
		}
	}

	// TODO(jeff): refactor config with CRD
	servicemeshEnabled := s.ServiceMeshOptions != nil && len(s.ServiceMeshOptions.IstioPilotHost) != 0
	if err = addControllers(mgr,
		kubernetesClient,
		informerFactory,
		devopsClient,
		s3Client,
		ldapClient,
		s.KubernetesOptions,
		s.AuthenticationOptions,
		s.MultiClusterOptions,
		s.NetworkOptions,
		servicemeshEnabled,
		s.AuthenticationOptions.KubectlImage, ctx.Done()); err != nil {
		klog.Fatalf("unable to register controllers to the manager: %v", err)
	}

	// Start cache data after all informer is registered
	klog.V(0).Info("Starting cache resource from apiserver...")
	informerFactory.Start(ctx.Done())

	// Setup webhooks
	klog.V(2).Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()

	klog.V(2).Info("registering webhooks to the webhook server")
	hookServer.Register("/validate-email-iam-kubesphere-io-v1alpha2", &webhook.Admission{Handler: &user.EmailValidator{Client: mgr.GetClient()}})
	hookServer.Register("/validate-network-kubesphere-io-v1alpha1", &webhook.Admission{Handler: &webhooks.ValidatingHandler{C: mgr.GetClient()}})
	hookServer.Register("/mutate-network-kubesphere-io-v1alpha1", &webhook.Admission{Handler: &webhooks.MutatingHandler{C: mgr.GetClient()}})

	resourceQuotaAdmission, err := quota.NewResourceQuotaAdmission(mgr.GetClient(), mgr.GetScheme())
	if err != nil {
		klog.Fatalf("unable to create resource quota admission: %v", err)
	}
	hookServer.Register("/validate-quota-kubesphere-io-v1alpha2", &webhook.Admission{Handler: resourceQuotaAdmission})

	klog.V(2).Info("registering metrics to the webhook server")
	// Add an extra metric endpoint, so we can use the the same metric definition with ks-apiserver
	// /kapis/metrics is independent of controller-manager's built-in /metrics
	mgr.AddMetricsExtraHandler("/kapis/metrics", metrics.Handler())

	klog.V(0).Info("Starting the controllers.")
	if err = mgr.Start(ctx); err != nil {
		klog.Fatalf("unable to run the manager: %v", err)
	}

	return nil
}
