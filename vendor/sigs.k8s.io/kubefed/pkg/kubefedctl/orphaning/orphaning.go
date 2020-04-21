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
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"

	"sigs.k8s.io/kubefed/pkg/apis/core/typeconfig"
	ctlutil "sigs.k8s.io/kubefed/pkg/controller/util"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/enable"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/options"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

type orphanResource struct {
	options.GlobalSubcommandOptions
	typeName          string
	resourceName      string
	resourceNamespace string
}

// Bind adds the join specific arguments to the flagset passed in as an argument.
func (o *orphanResource) Bind(flags *pflag.FlagSet) error {
	flags.StringVarP(&o.resourceNamespace, "namespace", "n", "", "If present, the namespace scope for this CLI request")
	err := flags.MarkHidden("kubefed-namespace")
	if err != nil {
		return err
	}
	err = flags.MarkHidden("dry-run")
	if err != nil {
		return err
	}
	return nil
}

// NewCmdOrphaning the head of orphaning-deletion sub commands
func NewCmdOrphaning(cmdOut io.Writer, config util.FedConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orphaning-deletion",
		Short: "Manage orphaning delete policy",
		Long:  "Manage orphaning delete policy",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}
		},
	}
	cmd.AddCommand(newCmdEnableOrphaning(cmdOut, config))
	cmd.AddCommand(newCmdDisableOrphaning(cmdOut, config))
	cmd.AddCommand(newCmdStatusOrphaning(cmdOut, config))

	return cmd
}

//  Complete ensures that options are valid and marshals them if necessary.
func (o *orphanResource) Complete(args []string, config util.FedConfig) error {
	if len(args) == 0 {
		return errors.New("resource type is required")
	}

	o.typeName = args[0]

	if len(args) == 1 {
		return errors.New("resource name is required")
	}
	o.resourceName = args[1]

	if len(o.resourceNamespace) == 0 {
		var err error
		o.resourceNamespace, err = util.GetNamespace(o.HostClusterContext, o.Kubeconfig, config)
		return err
	}
	return nil
}

// Returns a Federated Resources Interface
func (o *orphanResource) GetResourceClient(config util.FedConfig, cmdOut io.Writer) (dynamic.ResourceInterface, error) {
	hostClientConfig := config.GetClientConfig(o.HostClusterContext, o.Kubeconfig)
	if err := o.SetHostClusterContextFromConfig(hostClientConfig); err != nil {
		return nil, err
	}
	hostConfig, err := hostClientConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to load configuration for cluster context %q in kubeconfig %q.`",
			o.HostClusterContext, o.Kubeconfig)
	}
	// Lookup kubernetes API availability
	apiResource, err := enable.LookupAPIResource(hostConfig, o.typeName, "")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to find targeted %s type", o.typeName)
	}
	klog.V(2).Infof("API Resource for %s/%s found", typeconfig.GroupQualifiedName(*apiResource), apiResource.Version)
	if !util.IsFederatedAPIResource(apiResource.Kind, apiResource.Group) {
		fmt.Fprintf(cmdOut, "Warning: %s/%s might not be a federated resource\n",
			typeconfig.GroupQualifiedName(*apiResource), apiResource.Version)
	}
	targetClient, err := ctlutil.NewResourceClient(hostConfig, apiResource)

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating client for %s", apiResource.Kind)
	}

	resourceClient := targetClient.Resources(o.resourceNamespace)
	return resourceClient, nil
}

// Returns the Federated resource where the orphaning-deletion will be managed
func (o *orphanResource) GetFederatedResource(resourceClient dynamic.ResourceInterface) (*unstructured.Unstructured, error) {
	resource, err := resourceClient.Get(o.resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to retrieve resource: %q",
			ctlutil.QualifiedName{Name: o.resourceName, Namespace: o.resourceNamespace})
	}
	return resource, nil
}
