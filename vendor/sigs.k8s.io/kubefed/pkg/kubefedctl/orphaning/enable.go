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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	ctlutil "sigs.k8s.io/kubefed/pkg/controller/util"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

var (
	orphaning_enable_long = `
		Prevents the removal of managed resources from member clusters when their managing federated 
		resource is removed. This is accomplished by adding 'kubefed.io/orphan: true' as an annotation to the 
		federated resource.

		Current context is assumed to be a Kubernetes cluster hosting
		the kubefed control plane. Please use the
		--host-cluster-context flag otherwise.`

	orphan_enable_example = `
		# Enable the orphaning mode for a federated resource of type FederatedDeployment and named foo 
		kubefedctl orphaning enable FederatedDeployment foo --host-cluster-context=cluster1`
)

// newCmdEnableOrphaning adds 'kubefed.io/orphan: true' as an annotation to the federated resource
func newCmdEnableOrphaning(cmdOut io.Writer, config util.FedConfig) *cobra.Command {
	opts := &orphanResource{}
	cmd := &cobra.Command{
		Use:     "enable <resource type> <resource name>",
		Short:   "Enable the orphaning (i.e. retention) of resources managed by a federated resource upon its removal.",
		Long:    orphaning_enable_long,
		Example: orphan_enable_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := opts.Complete(args, config)
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}

			err = opts.RunEnable(cmdOut, config)
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

// RunEnable implements the `enable` command.
func (o *orphanResource) RunEnable(cmdOut io.Writer, config util.FedConfig) error {
	resourceClient, err := o.GetResourceClient(config, cmdOut)
	if err != nil {
		return err
	}
	fedResource, err := o.GetFederatedResource(resourceClient)
	if err != nil {
		return err
	}
	if ctlutil.IsOrphaningEnabled(fedResource) {
		return nil
	}
	ctlutil.EnableOrphaning(fedResource)
	_, err = resourceClient.Update(fedResource, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "Failed to update resource %s %q", fedResource.GetKind(),
			ctlutil.QualifiedName{Name: fedResource.GetName(), Namespace: fedResource.GetNamespace()})
	}

	return nil
}
