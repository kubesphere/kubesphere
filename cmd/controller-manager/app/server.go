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
	v1 "k8s.io/api/core/v1"
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
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/term"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
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
			OpenPitrixOptions:   conf.OpenPitrixOptions,
			MultiClusterOptions: conf.MultiClusterOptions,
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

	openpitrixClient, err := openpitrix.NewClient(s.OpenPitrixOptions)
	if err != nil {
		klog.Errorf("Failed to create openpitrix client %v", err)
		return err
	}

	devopsClient, err := jenkins.NewDevopsClient(s.DevopsOptions)
	if err != nil {
		klog.Errorf("Failed to create devops client %v", err)
		return err
	}

	s3Client, err := s3.NewS3Client(s.S3Options)
	if err != nil {
		klog.Errorf("Failed to create s3 client %v", err)
		return err
	}

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.KubeSphere(),
		kubernetesClient.Istio(), kubernetesClient.Application(), kubernetesClient.Snapshot())

	run := func(ctx context.Context) {
		klog.V(0).Info("setting up manager")
		mgr, err := manager.New(kubernetesClient.Config(), manager.Options{})
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

		err = namespace.Add(mgr, openpitrixClient)
		if err != nil {
			klog.Fatal("Unable to create namespace controller")
		}

		if err := AddControllers(mgr, kubernetesClient, informerFactory, devopsClient, s3Client, stopCh); err != nil {
			klog.Fatalf("unable to register controllers to the manager: %v", err)
		}

		// Start cache data after all informer is registered
		informerFactory.Start(stopCh)

		// Setup webhooks
		klog.Info("setting up webhook server")
		hookServer := mgr.GetWebhookServer()

		klog.Info("registering webhooks to the webhook server")
		hookServer.Register("/mutating-encrypt-password-iam-kubesphere-io-v1alpha2-user", &webhook.Admission{Handler: &user.PasswordCipher{Client: mgr.GetClient()}})
		hookServer.Register("/validate-email-iam-kubesphere-io-v1alpha2-user", &webhook.Admission{Handler: &user.EmailValidator{Client: mgr.GetClient()}})
		hookServer.Register("/validate-service-nsnp-kubesphere-io-v1alpha1-network", &webhook.Admission{Handler: &nsnetworkpolicy.ServiceValidator{}})

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

	// TODO: change lockType to lease
	// once we finished moving to Kubernetes v1.16+, we
	// change lockType to lease
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
