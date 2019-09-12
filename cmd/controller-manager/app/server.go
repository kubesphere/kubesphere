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
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/cmd/controller-manager/app/options"
	"kubesphere.io/kubesphere/pkg/apis"
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

			err = Complete(s)
			if err != nil {
				os.Exit(1)
			}

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

func Complete(s *options.KubeSphereControllerManagerOptions) error {
	conf := controllerconfig.Get()

	conf.Apply(&controllerconfig.Config{
		DevopsOptions:     s.DevopsOptions,
		KubernetesOptions: s.KubernetesOptions,
		S3Options:         s.S3Options,
	})

	s = &options.KubeSphereControllerManagerOptions{
		KubernetesOptions: conf.KubernetesOptions,
		DevopsOptions:     conf.DevopsOptions,
		S3Options:         conf.S3Options,
	}

	return nil
}

func CreateClientSet(s *options.KubeSphereControllerManagerOptions, stopCh <-chan struct{}) error {
	csop := &client.ClientSetOptions{}

	csop.SetKubernetesOptions(s.KubernetesOptions).
		SetDevopsOptions(s.DevopsOptions).
		SetS3Options(s.S3Options)
	client.NewClientSetFactory(csop, stopCh)

	return nil
}

func Run(s *options.KubeSphereControllerManagerOptions, stopCh <-chan struct{}) error {
	err := CreateClientSet(s, stopCh)
	if err != nil {
		klog.Error(err)
		return err
	}

	config := client.ClientSets().K8s().Config()

	klog.Info("setting up manager")
	mgr, err := manager.New(config, manager.Options{})
	if err != nil {
		klog.Error(err, "unable to set up overall controller manager")
		return err
	}

	klog.Info("setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		klog.Error(err, "unable add APIs to scheme")
		return err
	}

	klog.Info("Setting up controllers")
	if err := controller.AddToManager(mgr); err != nil {
		klog.Error(err, "unable to register controllers to the manager")
		return err
	}

	if err := AddControllers(mgr, config, stopCh); err != nil {
		klog.Error(err, "unable to register controllers to the manager")
		return err
	}

	klog.Info("Starting the Cmd.")
	if err := mgr.Start(stopCh); err != nil {
		klog.Error(err, "unable to run the manager")
		return err
	}

	return nil
}
