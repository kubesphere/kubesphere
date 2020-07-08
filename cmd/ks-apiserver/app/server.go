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
	kconfig "github.com/kiali/kiali/config"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"kubesphere.io/kubesphere/cmd/ks-apiserver/app/options"
	apiserverconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/utils/signals"
	"kubesphere.io/kubesphere/pkg/utils/term"

	tracing "kubesphere.io/kubesphere/pkg/kapis/servicemesh/metrics/v1alpha2"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	// Load configuration from file
	conf, err := apiserverconfig.TryLoadFromDisk()
	if err == nil {
		s = &options.ServerRunOptions{
			GenericServerRunOptions: s.GenericServerRunOptions,
			Config:                  conf,
		}
	}

	cmd := &cobra.Command{
		Use: "ks-apiserver",
		Long: `The KubeSphere API server validates and configures data for the API objects. 
The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if errs := s.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}

			return Run(s, signals.SetupSignalHandler())
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

func Run(s *options.ServerRunOptions, stopCh <-chan struct{}) error {

	initializeServicemeshConfig(s)

	apiserver, err := s.NewAPIServer(stopCh)
	if err != nil {
		return err
	}

	err = apiserver.PrepareRun(stopCh)
	if err != nil {
		return nil
	}

	return apiserver.Run(stopCh)
}

func initializeServicemeshConfig(s *options.ServerRunOptions) {
	// Initialize kiali config
	config := kconfig.NewConfig()

	// Config jaeger query endpoint address
	if s.ServiceMeshOptions != nil && len(s.ServiceMeshOptions.JaegerQueryHost) != 0 {
		tracing.JaegerQueryUrl = s.ServiceMeshOptions.JaegerQueryHost
	}

	// Exclude system namespaces
	config.API.Namespaces.Exclude = []string{"istio-system", "kubesphere*", "kube*"}
	config.InCluster = true

	// Set default prometheus service url
	config.ExternalServices.PrometheusServiceURL = s.ServiceMeshOptions.ServicemeshPrometheusHost
	config.ExternalServices.PrometheusCustomMetricsURL = config.ExternalServices.PrometheusServiceURL

	// Set istio pilot discovery service url
	config.ExternalServices.Istio.UrlServiceVersion = s.ServiceMeshOptions.IstioPilotHost

	kconfig.Set(config)
}
