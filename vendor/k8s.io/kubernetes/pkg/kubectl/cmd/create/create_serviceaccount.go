/*
Copyright 2016 The Kubernetes Authors.

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

package create

import (
	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
)

var (
	serviceAccountLong = templates.LongDesc(i18n.T(`
		Create a service account with the specified name.`))

	serviceAccountExample = templates.Examples(i18n.T(`
	  # Create a new service account named my-service-account
	  kubectl create serviceaccount my-service-account`))
)

type ServiceAccountOpts struct {
	CreateSubcommandOptions *CreateSubcommandOptions
}

// NewCmdCreateServiceAccount is a macro command to create a new service account
func NewCmdCreateServiceAccount(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	options := &ServiceAccountOpts{
		CreateSubcommandOptions: NewCreateSubcommandOptions(ioStreams),
	}

	cmd := &cobra.Command{
		Use: "serviceaccount NAME [--dry-run]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"sa"},
		Short:                 i18n.T("Create a service account with the specified name"),
		Long:                  serviceAccountLong,
		Example:               serviceAccountExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(options.Complete(cmd, args))
			cmdutil.CheckErr(options.Run(f))
		},
	}

	options.CreateSubcommandOptions.PrintFlags.AddFlags(cmd)

	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.ServiceAccountV1GeneratorName)
	return cmd
}

func (o *ServiceAccountOpts) Complete(cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	var generator kubectl.StructuredGenerator
	switch generatorName := cmdutil.GetFlagString(cmd, "generator"); generatorName {
	case cmdutil.ServiceAccountV1GeneratorName:
		generator = &kubectl.ServiceAccountGeneratorV1{Name: name}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}

	return o.CreateSubcommandOptions.Complete(cmd, args, generator)
}

// CreateServiceAccount implements the behavior to run the create service account command
func (o *ServiceAccountOpts) Run(f cmdutil.Factory) error {
	return RunCreateSubcommand(f, o.CreateSubcommandOptions)
}
