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
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	apiextv1b1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"sigs.k8s.io/kubefed/pkg/apis/core/typeconfig"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	genericclient "sigs.k8s.io/kubefed/pkg/client/generic"
	ctlutil "sigs.k8s.io/kubefed/pkg/controller/util"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/enable"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/options"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

const (
	federatedGroupUsage = "The name of the API group to use for deleting the federated CRD type when the federated type config does not exist. Only used with --delete-crd."
	targetVersionUsage  = "The API version of the target type to use for deletion of the federated CRD type when the federated type config does not exist. Only used with --delete-crd."
)

var (
	disable_long = `
		Disables propagation of a Kubernetes API type.  This command
		can also optionally delete the API resources added by the enable
		command.

		Current context is assumed to be a Kubernetes cluster hosting
		the kubefed control plane. Please use the
		--host-cluster-context flag otherwise.`

	disable_example = `
		# Disable propagation of the kubernetes API type 'Deployment', named
		in FederatedTypeConfig as 'deployments.apps'
		kubefedctl disable deployments.apps

		# Disable propagation of the kubernetes API type 'Deployment', named
		in FederatedTypeConfig as 'deployments.apps', and delete the
		corresponding Federated API resource
		kubefedctl disable deployments.apps --delete-crd`
)

type disableType struct {
	options.GlobalSubcommandOptions
	options.CommonEnableOptions
	disableTypeOptions
}

type disableTypeOptions struct {
	deleteCRD           bool
	enableTypeDirective *enable.EnableTypeDirective
}

// Bind adds the disable specific arguments to the flagset passed in as an
// argument.
func (o *disableTypeOptions) Bind(flags *pflag.FlagSet) {
	flags.BoolVar(&o.deleteCRD, "delete-crd", false, "Whether to remove the API resource added by 'enable'.")
}

// NewCmdTypeDisable defines the `disable` command that
// disables federation of a Kubernetes API type.
func NewCmdTypeDisable(cmdOut io.Writer, config util.FedConfig) *cobra.Command {
	opts := &disableType{}

	cmd := &cobra.Command{
		Use:     "disable NAME",
		Short:   "Disables propagation of a Kubernetes API type",
		Long:    disable_long,
		Example: disable_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := opts.Complete(args)
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}

			err = opts.Run(cmdOut, config)
			if err != nil {
				klog.Fatalf("Error: %v", err)
			}
		},
	}

	flags := cmd.Flags()
	opts.GlobalSubcommandBind(flags)
	opts.CommonSubcommandBind(flags, federatedGroupUsage, targetVersionUsage)
	opts.Bind(flags)

	return cmd
}

// Complete ensures that options are valid and marshals them if necessary.
func (j *disableType) Complete(args []string) error {
	j.enableTypeDirective = enable.NewEnableTypeDirective()
	directive := j.enableTypeDirective

	if err := j.SetName(args); err != nil {
		return err
	}

	if !j.deleteCRD {
		if len(j.TargetVersion) > 0 {
			return errors.New("--version flag valid only with --delete-crd")
		} else if j.FederatedGroup != options.DefaultFederatedGroup {
			return errors.New("--kubefed-group flag valid only with --delete-crd")
		}
	}

	if len(j.TargetVersion) > 0 {
		directive.Spec.TargetVersion = j.TargetVersion
	}
	if len(j.FederatedGroup) > 0 {
		directive.Spec.FederatedGroup = j.FederatedGroup
	}

	return nil
}

// Run is the implementation of the `disable` command.
func (j *disableType) Run(cmdOut io.Writer, config util.FedConfig) error {
	hostConfig, err := config.HostConfig(j.HostClusterContext, j.Kubeconfig)
	if err != nil {
		return errors.Wrap(err, "Failed to get host cluster config")
	}

	// If . is specified, the target name is assumed as a group qualified name.
	// In such case, ignore the lookup to make sure deletion of a federatedtypeconfig
	// for which the corresponding target has been removed.
	name := j.TargetName
	if !strings.Contains(j.TargetName, ".") {
		apiResource, err := enable.LookupAPIResource(hostConfig, j.TargetName, "")
		if err != nil {
			return err
		}
		name = typeconfig.GroupQualifiedName(*apiResource)
	}

	typeConfigName := ctlutil.QualifiedName{
		Namespace: j.KubeFedNamespace,
		Name:      name,
	}
	j.enableTypeDirective.Name = typeConfigName.Name
	return DisableFederation(cmdOut, hostConfig, j.enableTypeDirective, typeConfigName, j.deleteCRD, j.DryRun, true)
}

func DisableFederation(cmdOut io.Writer, config *rest.Config, enableTypeDirective *enable.EnableTypeDirective,
	typeConfigName ctlutil.QualifiedName, deleteCRD, dryRun, verifyStopped bool) error {
	client, err := genericclient.New(config)
	if err != nil {
		return errors.Wrap(err, "Failed to get kubefed clientset")
	}

	write := func(data string) {
		if cmdOut == nil {
			return
		}

		if _, err := cmdOut.Write([]byte(data)); err != nil {
			klog.Fatalf("Unexpected err: %v\n", err)
		}
	}

	typeConfig := &fedv1b1.FederatedTypeConfig{}
	ftcExists, err := checkFederatedTypeConfigExists(client, typeConfig, typeConfigName, write)
	if err != nil {
		return err
	}

	if dryRun {
		return nil
	}

	// Disable propagation and verify it is stopped before deleting the CRD
	// when no custom resources exist. This avoids spurious error messages in
	// the controller manager log as watches are terminated and cannot be
	// reestablished.
	if ftcExists {
		if deleteCRD {
			err = checkFederatedTypeCustomResourcesExist(config, typeConfig, write)
			if err != nil {
				return err
			}
		}
		if typeConfig.GetPropagationEnabled() {
			err = disablePropagation(client, typeConfig, typeConfigName, write)
			if err != nil {
				return err
			}
		}
		if verifyStopped {
			err = verifyPropagationControllerStopped(client, typeConfigName, write)
			if err != nil {
				return err
			}
		}
	}

	if deleteCRD {
		if !ftcExists {
			typeConfig, err = generatedFederatedTypeConfig(config, enableTypeDirective)
			if err != nil {
				return err
			}
		}
		err = deleteFederatedType(config, typeConfig, write)
		if err != nil {
			return err
		}
	}

	if ftcExists {
		err = deleteFederatedTypeConfig(client, typeConfig, typeConfigName, write)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkFederatedTypeConfigExists(client genericclient.Client, typeConfig *fedv1b1.FederatedTypeConfig, typeConfigName ctlutil.QualifiedName, write func(string)) (bool, error) {
	err := client.Get(context.TODO(), typeConfig, typeConfigName.Namespace, typeConfigName.Name)
	if err == nil {
		return true, nil
	}

	if apierrors.IsNotFound(err) {
		write(fmt.Sprintf("FederatedTypeConfig %q does not exist\n", typeConfigName))
		return false, nil
	}

	return false, errors.Wrapf(err, "Error retrieving FederatedTypeConfig %q", typeConfigName)
}

func disablePropagation(client genericclient.Client, typeConfig *fedv1b1.FederatedTypeConfig, typeConfigName ctlutil.QualifiedName, write func(string)) error {
	if typeConfig.GetPropagationEnabled() {
		typeConfig.Spec.Propagation = fedv1b1.PropagationDisabled
		err := client.Update(context.TODO(), typeConfig)
		if err != nil {
			return errors.Wrapf(err, "Error disabling propagation for FederatedTypeConfig %q", typeConfigName)
		}
		write(fmt.Sprintf("Disabled propagation for FederatedTypeConfig %q\n", typeConfigName))
	} else {
		write(fmt.Sprintf("Propagation already disabled for FederatedTypeConfig %q\n", typeConfigName))
	}
	return nil
}

func verifyPropagationControllerStopped(client genericclient.Client, typeConfigName ctlutil.QualifiedName, write func(string)) error {
	write(fmt.Sprintf("Verifying propagation controller is stopped for FederatedTypeConfig %q\n", typeConfigName))

	var typeConfig *fedv1b1.FederatedTypeConfig
	err := wait.PollImmediate(100*time.Millisecond, 10*time.Second, func() (bool, error) {
		typeConfig = &fedv1b1.FederatedTypeConfig{}
		err := client.Get(context.TODO(), typeConfig, typeConfigName.Namespace, typeConfigName.Name)
		if err != nil {
			klog.Errorf("Error retrieving FederatedTypeConfig %q: %v", typeConfigName, err)
			return false, nil
		}
		if typeConfig.Status.PropagationController == fedv1b1.ControllerStatusNotRunning {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return errors.Wrapf(err, "Unable to verify propagation controller for FederatedTypeConfig %q is stopped: %v", typeConfigName, err)
	}

	write(fmt.Sprintf("Propagation controller for FederatedTypeConfig %q is stopped\n", typeConfigName))
	return nil
}

func deleteFederatedTypeConfig(client genericclient.Client, typeConfig *fedv1b1.FederatedTypeConfig, typeConfigName ctlutil.QualifiedName, write func(string)) error {
	err := client.Delete(context.TODO(), typeConfig, typeConfig.Namespace, typeConfig.Name)
	if err != nil {
		return errors.Wrapf(err, "Error deleting FederatedTypeConfig %q", typeConfigName)
	}
	write(fmt.Sprintf("federatedtypeconfig %q deleted\n", typeConfigName))
	return nil
}

func generatedFederatedTypeConfig(config *rest.Config, enableTypeDirective *enable.EnableTypeDirective) (*fedv1b1.FederatedTypeConfig, error) {
	apiResource, err := enable.LookupAPIResource(config, enableTypeDirective.Name, enableTypeDirective.Spec.TargetVersion)
	if err != nil {
		return nil, err
	}
	typeConfig := enable.GenerateTypeConfigForTarget(*apiResource, enableTypeDirective).(*fedv1b1.FederatedTypeConfig)
	return typeConfig, nil
}

func deleteFederatedType(config *rest.Config, typeConfig typeconfig.Interface, write func(string)) error {
	err := checkFederatedTypeCustomResourcesExist(config, typeConfig, write)
	if err != nil {
		return err
	}

	crdName := typeconfig.GroupQualifiedName(typeConfig.GetFederatedType())
	err = deleteFederatedCRD(config, crdName, write)
	if err != nil {
		return err
	}

	return nil
}

func checkFederatedTypeCustomResourcesExist(config *rest.Config, typeConfig typeconfig.Interface, write func(string)) error {
	federatedTypeAPIResource := typeConfig.GetFederatedType()
	crdName := typeconfig.GroupQualifiedName(federatedTypeAPIResource)
	exists, err := customResourcesExist(config, &federatedTypeAPIResource)
	if err != nil {
		return err
	} else if exists {
		return errors.Errorf("Cannot delete CRD %q while resource instances exist. Please try kubefedctl disable again after removing the resource instances or without the '--delete-crd' option\n", crdName)
	}
	return nil
}

func customResourcesExist(config *rest.Config, resource *metav1.APIResource) (bool, error) {
	client, err := ctlutil.NewResourceClient(config, resource)
	if err != nil {
		return false, err
	}

	options := metav1.ListOptions{}
	objList, err := client.Resources("").List(options)
	if apierrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return len(objList.Items) != 0, nil
}

func deleteFederatedCRD(config *rest.Config, crdName string, write func(string)) error {
	client, err := apiextv1b1client.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "Error creating crd client")
	}

	err = client.CustomResourceDefinitions().Delete(crdName, nil)
	if apierrors.IsNotFound(err) {
		write(fmt.Sprintf("customresourcedefinition %q does not exist\n", crdName))
	} else if err != nil {
		return errors.Wrapf(err, "Error deleting crd %q", crdName)
	} else {
		write(fmt.Sprintf("customresourcedefinition %q deleted\n", crdName))
	}
	return nil
}
