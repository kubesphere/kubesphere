// Copyright 2020 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	goflag "flag"
	cliflag "k8s.io/component-base/cli/flag"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	controllermanager "kubesphere.io/kubesphere/cmd/controller-manager/app"
	ksapiserver "kubesphere.io/kubesphere/cmd/ks-apiserver/app"
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

	commandFns := []func() *cobra.Command{
		apiserver,
		controllermanager,
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
