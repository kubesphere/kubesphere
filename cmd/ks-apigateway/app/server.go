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
	"flag"
	"github.com/mholt/caddy/caddy/caddymain"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/spf13/cobra"
	apiserverconfig "kubesphere.io/kubesphere/pkg/server/config"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/utils/signals"

	"kubesphere.io/kubesphere/pkg/apigateway"
)

func NewAPIGatewayCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use: "ks-apigateway",
		Long: `The KubeSphere API Gateway, which is responsible 
for proxy request to the right backend. API Gateway also proxy 
Kubernetes API Server for KubeSphere authorization purpose.
`,
		RunE: func(cmd *cobra.Command, args []string) error {

			err := apiserverconfig.Load()
			if err != nil {
				return err
			}

			apigateway.RegisterPlugins()

			return Run(signals.SetupSignalHandler())
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	return cmd
}

func Run(stopCh <-chan struct{}) error {

	csop := &client.ClientSetOptions{}
	csop.SetKubernetesOptions(apiserverconfig.Get().KubernetesOptions)
	client.NewClientSetFactory(csop, stopCh)

	httpserver.RegisterDevDirective("authenticate", "jwt")
	httpserver.RegisterDevDirective("authentication", "jwt")
	httpserver.RegisterDevDirective("swagger", "jwt")
	caddymain.Run()

	return nil
}
