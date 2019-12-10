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
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	"kubesphere.io/kubesphere/pkg/controller"
	controllerconfig "kubesphere.io/kubesphere/pkg/server/config"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/utils/term"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func NewControllerManagerCommand() *cobra.Command {
	s := options.NewKubeSphereControllerManagerOptions()

	cmd := &cobra.Command{
		Use:  "controller-manager",
		Long: `KubeSphere controller manager is a daemon that`,
		Run: func(cmd *cobra.Command, args []string) {

			err := controllerconfig.Load()
			if err != nil {
				klog.Fatal(err)
				os.Exit(1)
			}

			s = Complete(s)

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

func Complete(s *options.KubeSphereControllerManagerOptions) *options.KubeSphereControllerManagerOptions {
	conf := controllerconfig.Get()

	conf.Apply(&controllerconfig.Config{
		DevopsOptions:     s.DevopsOptions,
		KubernetesOptions: s.KubernetesOptions,
		S3Options:         s.S3Options,
		OpenPitrixOptions: s.OpenPitrixOptions,
	})

	out := &options.KubeSphereControllerManagerOptions{
		KubernetesOptions: conf.KubernetesOptions,
		DevopsOptions:     conf.DevopsOptions,
		S3Options:         conf.S3Options,
		OpenPitrixOptions: conf.OpenPitrixOptions,
		LeaderElection:    s.LeaderElection,
	}

	return out
}

func CreateClientSet(conf *controllerconfig.Config, stopCh <-chan struct{}) error {
	csop := &client.ClientSetOptions{}

	csop.SetKubernetesOptions(conf.KubernetesOptions).
		SetDevopsOptions(conf.DevopsOptions).
		SetS3Options(conf.S3Options).
		SetOpenPitrixOptions(conf.OpenPitrixOptions).
		SetKubeSphereOptions(conf.KubeSphereOptions)
	client.NewClientSetFactory(csop, stopCh)

	return nil
}

func Run(s *options.KubeSphereControllerManagerOptions, stopCh <-chan struct{}) error {
	err := CreateClientSet(controllerconfig.Get(), stopCh)
	if err != nil {
		klog.Error(err)
		return err
	}

	config := client.ClientSets().K8s().Config()

	run := func(ctx context.Context) {
		klog.V(0).Info("setting up manager")
		mgr, err := manager.New(config, manager.Options{})
		if err != nil {
			klog.Fatalf("unable to set up overall controller manager: %v", err)
		}

		klog.V(0).Info("setting up scheme")
		if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
			klog.Fatalf("unable add APIs to scheme: %v", err)
		}

		klog.V(0).Info("Setting up controllers")
		if err := controller.AddToManager(mgr); err != nil {
			klog.Fatalf("unable to register controllers to the manager: %v", err)
		}

		if err := AddControllers(mgr, config, stopCh); err != nil {
			klog.Fatalf("unable to register controllers to the manager: %v", err)
		}

		klog.V(0).Info("Starting the Cmd.")
		if err := mgr.Start(stopCh); err != nil {
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
		client.ClientSets().K8s().Kubernetes().CoreV1(),
		client.ClientSets().K8s().Kubernetes().CoordinationV1(),
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
