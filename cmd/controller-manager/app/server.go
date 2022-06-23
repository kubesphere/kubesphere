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

	"github.com/google/gops/agent"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"kubesphere.io/kubesphere/pkg/controller/network/webhooks"
	"kubesphere.io/kubesphere/pkg/controller/quota"
	"kubesphere.io/kubesphere/pkg/controller/user"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
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
			MonitoringOptions:     conf.MonitoringOptions,
			LeaderElection:        s.LeaderElection,
			LeaderElect:           s.LeaderElect,
			WebhookCertDir:        s.WebhookCertDir,
		}
	} else {
		klog.Fatal("Failed to load configuration from disk", err)
	}

	cmd := &cobra.Command{
		Use:  "controller-manager",
		Long: `KubeSphere controller manager is a daemon that embeds the control loops shipped with KubeSphere.`,
		Run: func(cmd *cobra.Command, args []string) {
			if errs := s.Validate(allControllers); len(errs) != 0 {
				klog.Error(utilerrors.NewAggregate(errs))
				os.Exit(1)
			}

			if s.GOPSEnabled {
				// Add agent to report additional information such as the current stack trace, Go version, memory stats, etc.
				// Bind to a random port on address 127.0.0.1
				if err := agent.Listen(agent.Options{}); err != nil {
					klog.Fatal(err)
				}
			}

			if err = Run(s, controllerconfig.WatchConfigChange(), signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags(allControllers)

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

func Run(s *options.KubeSphereControllerManagerOptions, configCh <-chan controllerconfig.Config, ctx context.Context) error {
	ictx, cancelFunc := context.WithCancel(context.TODO())
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		if err := run(s, ictx); err != nil {
			errCh <- err
		}
	}()

	// The ctx (signals.SetupSignalHandler()) is to control the entire program life cycle,
	// The ictx(internal context)  is created here to control the life cycle of the controller-manager(all controllers, sharedInformer, webhook etc.)
	// when config changed, stop server and renew context, start new server
	for {
		select {
		case <-ctx.Done():
			cancelFunc()
			return nil
		case cfg := <-configCh:
			cancelFunc()
			s.MergeConfig(&cfg)
			ictx, cancelFunc = context.WithCancel(context.TODO())
			go func() {
				if err := run(s, ictx); err != nil {
					errCh <- err
				}
			}()
		case err := <-errCh:
			cancelFunc()
			return err
		}
	}
}

func run(s *options.KubeSphereControllerManagerOptions, ctx context.Context) error {

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		klog.Errorf("Failed to create kubernetes clientset %v", err)
		return err
	}

	if s.S3Options != nil && len(s.S3Options.Endpoint) != 0 {
		_, err = s3.NewS3Client(s.S3Options)
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

	// TODO(jeff): refactor config with CRD
	// install all controllers
	if err = addAllControllers(mgr,
		kubernetesClient,
		informerFactory,
		s,
		ctx.Done()); err != nil {
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
	hookServer.Register("/persistentvolumeclaims", &webhook.Admission{Handler: &webhooks.AccessorHandler{C: mgr.GetClient()}})

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
