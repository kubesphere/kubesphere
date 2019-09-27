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
