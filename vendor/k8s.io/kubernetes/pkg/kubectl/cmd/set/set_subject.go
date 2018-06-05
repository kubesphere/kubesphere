/*
Copyright 2017 The Kubernetes Authors.

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

package set

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/printers"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
)

var (
	subject_long = templates.LongDesc(`
	Update User, Group or ServiceAccount in a RoleBinding/ClusterRoleBinding.`)

	subject_example = templates.Examples(`
	# Update a ClusterRoleBinding for serviceaccount1
	kubectl set subject clusterrolebinding admin --serviceaccount=namespace:serviceaccount1

	# Update a RoleBinding for user1, user2, and group1
	kubectl set subject rolebinding admin --user=user1 --user=user2 --group=group1

	# Print the result (in yaml format) of updating rolebinding subjects from a local, without hitting the server
	kubectl create rolebinding admin --role=admin --user=admin -o yaml --dry-run | kubectl set subject --local -f - --user=foo -o yaml`)
)

type updateSubjects func(existings []rbac.Subject, targets []rbac.Subject) (bool, []rbac.Subject)

// SubjectOptions is the start of the data required to perform the operation. As new fields are added, add them here instead of
// referencing the cmd.Flags
type SubjectOptions struct {
	PrintFlags *printers.PrintFlags

	resource.FilenameOptions

	Infos             []*resource.Info
	Selector          string
	ContainerSelector string
	Output            string
	All               bool
	DryRun            bool
	Local             bool

	Users           []string
	Groups          []string
	ServiceAccounts []string

	namespace string

	PrintObj printers.ResourcePrinterFunc

	genericclioptions.IOStreams
}

func NewSubjectOptions(streams genericclioptions.IOStreams) *SubjectOptions {
	return &SubjectOptions{
		PrintFlags: printers.NewPrintFlags("subjects updated"),

		IOStreams: streams,
	}
}

func NewCmdSubject(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewSubjectOptions(streams)
	cmd := &cobra.Command{
		Use: "subject (-f FILENAME | TYPE NAME) [--user=username] [--group=groupname] [--serviceaccount=namespace:serviceaccountname] [--dry-run]",
		DisableFlagsInUseLine: true,
		Short:   i18n.T("Update User, Group or ServiceAccount in a RoleBinding/ClusterRoleBinding"),
		Long:    subject_long,
		Example: subject_example,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run(addSubjects))
		},
	}

	o.PrintFlags.AddFlags(cmd)

	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, "the resource to update the subjects")
	cmd.Flags().BoolVar(&o.All, "all", o.All, "Select all resources, including uninitialized ones, in the namespace of the specified resource types")
	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on, not including uninitialized ones, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&o.Local, "local", o.Local, "If true, set subject will NOT contact api-server but run locally.")
	cmdutil.AddDryRunFlag(cmd)
	cmd.Flags().StringArrayVar(&o.Users, "user", o.Users, "Usernames to bind to the role")
	cmd.Flags().StringArrayVar(&o.Groups, "group", o.Groups, "Groups to bind to the role")
	cmd.Flags().StringArrayVar(&o.ServiceAccounts, "serviceaccount", o.ServiceAccounts, "Service accounts to bind to the role")
	cmdutil.AddIncludeUninitializedFlag(cmd)
	return cmd
}

func (o *SubjectOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.Output = cmdutil.GetFlagString(cmd, "output")
	o.DryRun = cmdutil.GetDryRunFlag(cmd)

	if o.DryRun {
		o.PrintFlags.Complete("%s (dry run)")
	}
	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrintObj = printer.PrintObj

	var enforceNamespace bool
	o.namespace, enforceNamespace, err = f.DefaultNamespace()
	if err != nil {
		return err
	}

	includeUninitialized := cmdutil.ShouldIncludeUninitialized(cmd, false)
	builder := f.NewBuilder().
		WithScheme(legacyscheme.Scheme).
		LocalParam(o.Local).
		ContinueOnError().
		NamespaceParam(o.namespace).DefaultNamespace().
		FilenameParam(enforceNamespace, &o.FilenameOptions).
		IncludeUninitialized(includeUninitialized).
		Flatten()

	if o.Local {
		// if a --local flag was provided, and a resource was specified in the form
		// <resource>/<name>, fail immediately as --local cannot query the api server
		// for the specified resource.
		if len(args) > 0 {
			return resource.LocalResourceError
		}
	} else {
		builder = builder.
			LabelSelectorParam(o.Selector).
			ResourceTypeOrNameArgs(o.All, args...).
			Latest()
	}

	o.Infos, err = builder.Do().Infos()
	if err != nil {
		return err
	}

	return nil
}

func (o *SubjectOptions) Validate() error {
	if o.All && len(o.Selector) > 0 {
		return fmt.Errorf("cannot set --all and --selector at the same time")
	}
	if len(o.Users) == 0 && len(o.Groups) == 0 && len(o.ServiceAccounts) == 0 {
		return fmt.Errorf("you must specify at least one value of user, group or serviceaccount")
	}

	for _, sa := range o.ServiceAccounts {
		tokens := strings.Split(sa, ":")
		if len(tokens) != 2 || tokens[1] == "" {
			return fmt.Errorf("serviceaccount must be <namespace>:<name>")
		}

		for _, info := range o.Infos {
			_, ok := info.Object.(*rbac.ClusterRoleBinding)
			if ok && tokens[0] == "" {
				return fmt.Errorf("serviceaccount must be <namespace>:<name>, namespace must be specified")
			}
		}
	}

	return nil
}

func (o *SubjectOptions) Run(fn updateSubjects) error {
	patches := CalculatePatches(o.Infos, cmdutil.InternalVersionJSONEncoder(), func(info *resource.Info) ([]byte, error) {
		subjects := []rbac.Subject{}
		for _, user := range sets.NewString(o.Users...).List() {
			subject := rbac.Subject{
				Kind:     rbac.UserKind,
				APIGroup: rbac.GroupName,
				Name:     user,
			}
			subjects = append(subjects, subject)
		}
		for _, group := range sets.NewString(o.Groups...).List() {
			subject := rbac.Subject{
				Kind:     rbac.GroupKind,
				APIGroup: rbac.GroupName,
				Name:     group,
			}
			subjects = append(subjects, subject)
		}
		for _, sa := range sets.NewString(o.ServiceAccounts...).List() {
			tokens := strings.Split(sa, ":")
			namespace := tokens[0]
			name := tokens[1]
			if len(namespace) == 0 {
				namespace = o.namespace
			}
			subject := rbac.Subject{
				Kind:      rbac.ServiceAccountKind,
				Namespace: namespace,
				Name:      name,
			}
			subjects = append(subjects, subject)
		}

		transformed, err := updateSubjectForObject(info.Object, subjects, fn)
		if transformed && err == nil {
			// TODO: switch UpdatePodSpecForObject to work on v1.PodSpec
			return runtime.Encode(cmdutil.InternalVersionJSONEncoder(), cmdutil.AsDefaultVersionedOrOriginal(info.Object, info.Mapping))
		}
		return nil, err
	})

	allErrs := []error{}
	for _, patch := range patches {
		info := patch.Info
		if patch.Err != nil {
			allErrs = append(allErrs, fmt.Errorf("error: %s/%s %v\n", info.Mapping.Resource, info.Name, patch.Err))
			continue
		}

		//no changes
		if string(patch.Patch) == "{}" || len(patch.Patch) == 0 {
			allErrs = append(allErrs, fmt.Errorf("info: %s %q was not changed\n", info.Mapping.Resource, info.Name))
			continue
		}

		if o.Local || o.DryRun {
			if err := o.PrintObj(info.Object, o.Out); err != nil {
				return err
			}
			continue
		}

		obj, err := resource.NewHelper(info.Client, info.Mapping).Patch(info.Namespace, info.Name, types.StrategicMergePatchType, patch.Patch)
		if err != nil {
			allErrs = append(allErrs, fmt.Errorf("failed to patch subjects to rolebinding: %v\n", err))
			continue
		}
		info.Refresh(obj, true)

		return o.PrintObj(cmdutil.AsDefaultVersionedOrOriginal(info.Object, info.Mapping), o.Out)
	}
	return utilerrors.NewAggregate(allErrs)
}

//Note: the obj mutates in the function
func updateSubjectForObject(obj runtime.Object, subjects []rbac.Subject, fn updateSubjects) (bool, error) {
	switch t := obj.(type) {
	case *rbac.RoleBinding:
		transformed, result := fn(t.Subjects, subjects)
		t.Subjects = result
		return transformed, nil
	case *rbac.ClusterRoleBinding:
		transformed, result := fn(t.Subjects, subjects)
		t.Subjects = result
		return transformed, nil
	default:
		return false, fmt.Errorf("setting subjects is only supported for RoleBinding/ClusterRoleBinding")
	}
}

func addSubjects(existings []rbac.Subject, targets []rbac.Subject) (bool, []rbac.Subject) {
	transformed := false
	updated := existings
	for _, item := range targets {
		if !contain(existings, item) {
			updated = append(updated, item)
			transformed = true
		}
	}
	return transformed, updated
}

func contain(slice []rbac.Subject, item rbac.Subject) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
