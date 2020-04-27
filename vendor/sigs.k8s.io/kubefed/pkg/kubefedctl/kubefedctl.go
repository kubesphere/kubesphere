/*
Copyright 2018 The Kubernetes Authors.

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

package kubefedctl

import (
	"flag"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/client-go/tools/clientcmd"
	apiserverflag "k8s.io/component-base/cli/flag"

	"sigs.k8s.io/kubefed/pkg/kubefedctl/enable"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/federate"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/orphaning"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

// NewKubeFedCtlCommand creates the `kubefedctl` command and its nested children.
func NewKubeFedCtlCommand(out io.Writer) *cobra.Command {
	// Parent command to which all subcommands are added.
	rootCmd := &cobra.Command{
		Use:   "kubefedctl",
		Short: "kubefedctl controls a Kubernetes Cluster Federation",
		Long:  "kubefedctl controls a Kubernetes Cluster Federation. Find more information at https://sigs.k8s.io/kubefed.",

		RunE: runHelp,
	}

	// Add the command line flags from other dependencies (e.g., klog), but do not
	// warn if they contain underscores.
	pflag.CommandLine.SetNormalizeFunc(apiserverflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	rootCmd.PersistentFlags().AddFlagSet(pflag.CommandLine)

	// From this point and forward we get warnings on flags that contain "_" separators
	rootCmd.SetGlobalNormalizationFunc(apiserverflag.WarnWordSepNormalizeFunc)

	// Prevent klog errors about logging before parsing.
	_ = flag.CommandLine.Parse(nil)

	fedConfig := util.NewFedConfig(clientcmd.NewDefaultPathOptions())
	rootCmd.AddCommand(enable.NewCmdTypeEnable(out, fedConfig))
	rootCmd.AddCommand(NewCmdTypeDisable(out, fedConfig))
	rootCmd.AddCommand(federate.NewCmdFederateResource(out, fedConfig))
	rootCmd.AddCommand(NewCmdJoin(out, fedConfig))
	rootCmd.AddCommand(NewCmdUnjoin(out, fedConfig))
	rootCmd.AddCommand(orphaning.NewCmdOrphaning(out, fedConfig))
	rootCmd.AddCommand(NewCmdVersion(out))

	return rootCmd
}

func runHelp(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
