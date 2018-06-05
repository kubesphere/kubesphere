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

package rollout

import (
	"io"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/printers"

	"github.com/spf13/cobra"
)

// UndoOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type UndoOptions struct {
	resource.FilenameOptions

	PrintFlags *printers.PrintFlags
	ToPrinter  func(string) (printers.ResourcePrinterFunc, error)

	Rollbackers []kubectl.Rollbacker
	Infos       []*resource.Info
	ToRevision  int64
	DryRun      bool

	Out io.Writer
}

var (
	undo_long = templates.LongDesc(`
		Rollback to a previous rollout.`)

	undo_example = templates.Examples(`
		# Rollback to the previous deployment
		kubectl rollout undo deployment/abc

		# Rollback to daemonset revision 3
		kubectl rollout undo daemonset/abc --to-revision=3

		# Rollback to the previous deployment with dry-run
		kubectl rollout undo --dry-run=true deployment/abc`)
)

func NewCmdRolloutUndo(f cmdutil.Factory, out io.Writer) *cobra.Command {
	o := &UndoOptions{
		PrintFlags: printers.NewPrintFlags(""),
	}

	validArgs := []string{"deployment", "daemonset", "statefulset"}
	argAliases := kubectl.ResourceAliases(validArgs)

	cmd := &cobra.Command{
		Use: "undo (TYPE NAME | TYPE/NAME) [flags]",
		DisableFlagsInUseLine: true,
		Short:   i18n.T("Undo a previous rollout"),
		Long:    undo_long,
		Example: undo_example,
		Run: func(cmd *cobra.Command, args []string) {
			allErrs := []error{}
			err := o.CompleteUndo(f, cmd, out, args)
			if err != nil {
				allErrs = append(allErrs, err)
			}
			err = o.RunUndo()
			if err != nil {
				allErrs = append(allErrs, err)
			}
			cmdutil.CheckErr(utilerrors.Flatten(utilerrors.NewAggregate(allErrs)))
		},
		ValidArgs:  validArgs,
		ArgAliases: argAliases,
	}

	cmd.Flags().Int64("to-revision", 0, "The revision to rollback to. Default to 0 (last revision).")
	usage := "identifying the resource to get from a server."
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	cmdutil.AddDryRunFlag(cmd)
	return cmd
}

func (o *UndoOptions) CompleteUndo(f cmdutil.Factory, cmd *cobra.Command, out io.Writer, args []string) error {
	if len(args) == 0 && cmdutil.IsFilenameSliceEmpty(o.Filenames) {
		return cmdutil.UsageErrorf(cmd, "Required resource not specified.")
	}

	o.ToRevision = cmdutil.GetFlagInt64(cmd, "to-revision")
	o.Out = out
	o.DryRun = cmdutil.GetDryRunFlag(cmd)

	cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	o.ToPrinter = func(operation string) (printers.ResourcePrinterFunc, error) {
		o.PrintFlags.NamePrintFlags.Operation = operation
		if o.DryRun {
			o.PrintFlags.Complete("%s (dry run)")
		}
		printer, err := o.PrintFlags.ToPrinter()
		if err != nil {
			return nil, err
		}

		return printer.PrintObj, nil
	}

	r := f.NewBuilder().
		WithScheme(legacyscheme.Scheme).
		NamespaceParam(cmdNamespace).DefaultNamespace().
		FilenameParam(enforceNamespace, &o.FilenameOptions).
		ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	err = r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}
		rollbacker, err := f.Rollbacker(info.ResourceMapping())
		if err != nil {
			return err
		}
		o.Infos = append(o.Infos, info)
		o.Rollbackers = append(o.Rollbackers, rollbacker)
		return nil
	})
	return err
}

func (o *UndoOptions) RunUndo() error {
	allErrs := []error{}
	for ix, info := range o.Infos {
		result, err := o.Rollbackers[ix].Rollback(info.Object, nil, o.ToRevision, o.DryRun)
		if err != nil {
			allErrs = append(allErrs, cmdutil.AddSourceToErr("undoing", info.Source, err))
			continue
		}
		printer, err := o.ToPrinter(result)
		if err != nil {
			allErrs = append(allErrs, err)
			continue
		}
		printer.PrintObj(cmdutil.AsDefaultVersionedOrOriginal(info.Object, info.Mapping), o.Out)
	}
	return utilerrors.NewAggregate(allErrs)
}
