/*
Copyright 2019 The Kubernetes Authors.

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

package orphaning

import (
	"io"

	"github.com/spf13/cobra"

	"k8s.io/klog"

	ctlutil "sigs.k8s.io/kubefed/pkg/controller/util"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

const (
	Enabled  = "Enabled"
	Disabled = "Disabled"
)

var (
	orphaning_status_long = `
		Checks the status of "orphaning enable" ('kubefed.io/orphan: true') annotation on a federated resource. 
		Returns "Enabled" or "Disabled"

		Current context is assumed to be a Kubernetes cluster hosting the kubefed control plane. 
		Please use the --host-cluster-context flag otherwise.`

	orphaning_status_example = `
		# Checks the status of the orphaning mode of a federated resource of type FederatedDeployment and named foo 
		kubefedctl orphaning status FederatedDeployment foo --host-cluster-context=cluster1`
)

// newCmdStatusOrphaning checks status of orphaning deletion of the federated resource
func newCmdStatusOrphaning(cmdOut io.Writer, config util.FedConfig) *cobra.Command {
	opts := &orphanResource{}
	cmd := &cobra.Command{
		Use:     "status <resource type> <resource name>",
		Short:   "Get the orphaning deletion status of the federated resource",
		Long:    orphaning_status_long,
		Example: orphaning_status_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := opts.Complete(args, config)
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}

			err = opts.RunStatus(cmdOut, config)
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}
		},
	}

	flags := cmd.Flags()
	opts.GlobalSubcommandBind(flags)
	err := opts.Bind(flags)
	if err != nil {
		klog.Fatalf("Error: %v", err)
	}

	return cmd
}

// RunStatus implements the `status` command.
func (o *orphanResource) RunStatus(cmdOut io.Writer, config util.FedConfig) error {
	resourceClient, err := o.GetResourceClient(config, cmdOut)
	if err != nil {
		return err
	}
	fedResource, err := o.GetFederatedResource(resourceClient)
	if err != nil {
		return err
	}
	if ctlutil.IsOrphaningEnabled(fedResource) {
		_, err = cmdOut.Write([]byte(Enabled + "\n"))
		return err
	}
	_, err = cmdOut.Write([]byte(Disabled + "\n"))
	return err
}
