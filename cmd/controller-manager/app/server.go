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
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/cmd/controller-manager/app/options"
	"kubesphere.io/kubesphere/pkg/apis"
	controllerconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	"kubesphere.io/kubesphere/pkg/controller/namespace"
	"kubesphere.io/kubesphere/pkg/controller/network/nsnetworkpolicy"
	"kubesphere.io/kubesphere/pkg/controller/user"
	"kubesphere.io/kubesphere/pkg/controller/workspace"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/term"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func NewControllerManagerCommand() *cobra.Command {
	s := options.NewKubeSphereControllerManagerOptions()
	conf, err := controllerconfig.TryLoadFromDisk()
	if err == nil {
		// make sure LeaderElection is not nil
		s = &options.KubeSphereControllerManagerOptions{
			KubernetesOptions:   conf.KubernetesOptions,
			DevopsOptions:       conf.DevopsOptions,
			S3Options:           conf.S3Options,
			LdapOptions:         conf.LdapOptions,
			OpenPitrixOptions:   conf.OpenPitrixOptions,
			NetworkOptions:      conf.NetworkOptions,
			MultiClusterOptions: conf.MultiClusterOptions,
			ServiceMeshOptions:  conf.ServiceMeshOptions,
			LeaderElection:      s.LeaderElection,
			LeaderElect:         s.LeaderElect,
		}
	}

	cmd := &cobra.Command{
		Use:  "controller-manager",
		Long: `KubeSphere controller manager is a daemon that`,
		Run: func(cmd *cobra.Command, args []string) {
			if errs := s.Validate(); len(errs) != 0 {
				klog.Error(utilerrors.NewAggregate(errs))
				os.Exit(1)
			}

			if err = Run(s, signals.SetupSignalHandler()); err != nil {
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
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})
	return cmd
}

func Run(s *options.KubeSphereControllerManagerOptions, stopCh <-chan struct{}) error {
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
	if s.LdapOptions == nil || len(s.LdapOptions.Host) == 0 {
		return fmt.Errorf("ldap service address MUST not be empty")
	} else {
		if s.LdapOptions.Host == ldapclient.FAKE_HOST { // for debug only
			ldapClient = ldapclient.NewSimpleLdap()
		} else {
			ldapClient, err = ldapclient.NewLdapClient(s.LdapOptions, stopCh)
			if err != nil {
				return fmt.Errorf("failed to connect to ldap service, please check ldap status, error: %v", err)
			}
		}
	}

	var openpitrixClient openpitrix.Client
	if s.OpenPitrixOptions != nil && !s.OpenPitrixOptions.IsEmpty() {
		openpitrixClient, err = openpitrix.NewClient(s.OpenPitrixOptions)
		if err != nil {
			return fmt.Errorf("failed to connect to openpitrix, please check openpitrix status, error: %v", err)
		}
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
		kubernetesClient.Application(),
		kubernetesClient.Snapshot(),
		kubernetesClient.ApiExtensions())

	run := func(ctx context.Context) {
		klog.V(0).Info("setting up manager")
		// Use 8443 instead of 443 cause we need root permission to bind port 443
		mgr, err := manager.New(kubernetesClient.Config(), manager.Options{CertDir: s.WebhookCertDir, Port: 8443})
		if err != nil {
			klog.Fatalf("unable to set up overall controller manager: %v", err)
		}

		klog.V(0).Info("setting up scheme")
		if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
			klog.Fatalf("unable add APIs to scheme: %v", err)
		}

		klog.V(0).Info("Setting up controllers")
		err = workspace.Add(mgr)
		if err != nil {
			klog.Fatal("Unable to create workspace controller")
		}

		err = namespace.Add(mgr)
		if err != nil {
			klog.Fatal("Unable to create namespace controller")
		}

		// TODO(jeff): refactor config with CRD
		servicemeshEnabled := s.ServiceMeshOptions != nil && len(s.ServiceMeshOptions.IstioPilotHost) != 0
		if err = addControllers(mgr, kubernetesClient, informerFactory, devopsClient, s3Client, ldapClient, openpitrixClient,
			s.MultiClusterOptions.Enable, s.NetworkOptions.EnableNetworkPolicy, servicemeshEnabled, stopCh); err != nil {
			klog.Fatalf("unable to register controllers to the manager: %v", err)
		}

		// Start cache data after all informer is registered
		informerFactory.Start(stopCh)

		// Setup webhooks
		klog.Info("setting up webhook server")
		hookServer := mgr.GetWebhookServer()

		klog.Info("registering webhooks to the webhook server")
		hookServer.Register("/validate-email-iam-kubesphere-io-v1alpha2-user", &webhook.Admission{Handler: &user.EmailValidator{Client: mgr.GetClient()}})
		hookServer.Register("/validate-nsnp-kubesphere-io-v1alpha1-network", &webhook.Admission{Handler: &nsnetworkpolicy.NSNPValidator{Client: mgr.GetClient()}})

		klog.V(0).Info("Starting the controllers.")
		if err = mgr.Start(stopCh); err != nil {
			klog.Fatalf("unable to run the manager: %v", err)
		}

		select {}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-stopCh
		cancel()
	}()

	if !s.LeaderElect {
		run(ctx)
		return nil
	}

	id, err := os.Hostname()
	if err != nil {
		return err
	}

	// add a uniquifier so that two processes on the same host don't accidentally both become active
	id = id + "_" + string(uuid.NewUUID())

	lock, err := resourcelock.New(resourcelock.LeasesResourceLock,
		"kubesphere-system",
		"ks-controller-manager",
		kubernetesClient.Kubernetes().CoreV1(),
		kubernetesClient.Kubernetes().CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: id,
			EventRecorder: record.NewBroadcaster().NewRecorder(scheme.Scheme, v1.EventSource{
				Component: "ks-controller-manager",
			}),
		})

	if err != nil {
		klog.Fatalf("error creating lock: %v", err)
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: s.LeaderElection.LeaseDuration,
		RenewDeadline: s.LeaderElection.RenewDeadline,
		RetryPeriod:   s.LeaderElection.RetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: run,
			OnStoppedLeading: func() {
				klog.Errorf("leadership lost")
				os.Exit(0)
			},
		},
	})

	return nil
}
