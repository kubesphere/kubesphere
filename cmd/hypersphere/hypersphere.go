package main

import (
	goflag "flag"
	cliflag "k8s.io/component-base/cli/flag"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	controllermanager "kubesphere.io/kubesphere/cmd/controller-manager/app"
	ksapigateway "kubesphere.io/kubesphere/cmd/ks-apigateway/app"
	ksapiserver "kubesphere.io/kubesphere/cmd/ks-apiserver/app"
	ksaiam "kubesphere.io/kubesphere/cmd/ks-iam/app"
	"os"
)

func main() {
	hypersphereCommand, allCommandFns := NewHyperSphereCommand()

	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	basename := filepath.Base(os.Args[0])
	if err := commandFor(basename, hypersphereCommand, allCommandFns).Execute(); err != nil {
		os.Exit(1)
	}
}

func commandFor(basename string, defaultCommand *cobra.Command, commands []func() *cobra.Command) *cobra.Command {
	for _, commandFn := range commands {
		command := commandFn()
		if command.Name() == basename {
			return command
		}

		for _, alias := range command.Aliases {
			if alias == basename {
				return command
			}
		}
	}

	return defaultCommand
}

func NewHyperSphereCommand() (*cobra.Command, []func() *cobra.Command) {
	apiserver := func() *cobra.Command { return ksapiserver.NewAPIServerCommand() }
	controllermanager := func() *cobra.Command { return controllermanager.NewControllerManagerCommand() }
	iam := func() *cobra.Command { return ksaiam.NewAPIServerCommand() }
	apigateway := func() *cobra.Command { return ksapigateway.NewAPIGatewayCommand() }

	commandFns := []func() *cobra.Command{
		apiserver,
		controllermanager,
		iam,
		apigateway,
	}

	cmd := &cobra.Command{
		Use:   "hypersphere",
		Short: "Request a new project",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
	}

	for i := range commandFns {
		cmd.AddCommand(commandFns[i]())
	}

	return cmd, commandFns
}
