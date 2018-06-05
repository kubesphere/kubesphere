/*
Copyright 2014 The Kubernetes Authors.

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

package get

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"net/url"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/printers"
	"k8s.io/kubernetes/pkg/util/interrupt"
)

// GetOptions contains the input to the get command.
type GetOptions struct {
	PrintFlags             *PrintFlags
	ToPrinter              func(*meta.RESTMapping, bool) (printers.ResourcePrinterFunc, error)
	IsHumanReadablePrinter bool
	PrintWithOpenAPICols   bool

	CmdParent string

	resource.FilenameOptions

	Raw       string
	Watch     bool
	WatchOnly bool
	ChunkSize int64

	LabelSelector     string
	FieldSelector     string
	AllNamespaces     bool
	Namespace         string
	ExplicitNamespace bool

	ServerPrint bool

	NoHeaders      bool
	Sort           bool
	IgnoreNotFound bool
	Export         bool

	IncludeUninitialized bool

	genericclioptions.IOStreams
}

var (
	getLong = templates.LongDesc(`
		Display one or many resources

		Prints a table of the most important information about the specified resources.
		You can filter the list using a label selector and the --selector flag. If the
		desired resource type is namespaced you will only see results in your current
		namespace unless you pass --all-namespaces.

		Uninitialized objects are not shown unless --include-uninitialized is passed.

		By specifying the output as 'template' and providing a Go template as the value
		of the --template flag, you can filter the attributes of the fetched resources.`)

	getExample = templates.Examples(i18n.T(`
		# List all pods in ps output format.
		kubectl get pods

		# List all pods in ps output format with more information (such as node name).
		kubectl get pods -o wide

		# List a single replication controller with specified NAME in ps output format.
		kubectl get replicationcontroller web

		# List deployments in JSON output format, in the "v1" version of the "apps" API group:
		kubectl get deployments.v1.apps -o json

		# List a single pod in JSON output format.
		kubectl get -o json pod web-pod-13je7

		# List a pod identified by type and name specified in "pod.yaml" in JSON output format.
		kubectl get -f pod.yaml -o json

		# Return only the phase value of the specified pod.
		kubectl get -o template pod/web-pod-13je7 --template={{.status.phase}}

		# List all replication controllers and services together in ps output format.
		kubectl get rc,services

		# List one or more resources by their type and names.
		kubectl get rc/web service/frontend pods/web-pod-13je7`))
)

const (
	useOpenAPIPrintColumnFlagLabel = "use-openapi-print-columns"
	useServerPrintColumns          = "server-print"
)

// NewGetOptions returns a GetOptions with default chunk size 500.
func NewGetOptions(parent string, streams genericclioptions.IOStreams) *GetOptions {
	return &GetOptions{
		PrintFlags: NewGetPrintFlags(),
		CmdParent:  parent,

		IOStreams: streams,
		ChunkSize: 500,
	}
}

// NewCmdGet creates a command object for the generic "get" action, which
// retrieves one or more resources from a server.
func NewCmdGet(parent string, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewGetOptions(parent, streams)

	cmd := &cobra.Command{
		Use: "get [(-o|--output=)json|yaml|wide|custom-columns=...|custom-columns-file=...|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=...] (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags]",
		DisableFlagsInUseLine: true,
		Short:   i18n.T("Display one or many resources"),
		Long:    getLong + "\n\n" + cmdutil.SuggestApiResources(parent),
		Example: getExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd))
			cmdutil.CheckErr(o.Run(f, cmd, args))
		},
		SuggestFor: []string{"list", "ps"},
	}

	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Raw, "raw", o.Raw, "Raw URI to request from the server.  Uses the transport specified by the kubeconfig file.")
	cmd.Flags().BoolVarP(&o.Watch, "watch", "w", o.Watch, "After listing/getting the requested object, watch for changes. Uninitialized objects are excluded if no object name is provided.")
	cmd.Flags().BoolVar(&o.WatchOnly, "watch-only", o.WatchOnly, "Watch for changes to the requested object(s), without listing/getting first.")
	cmd.Flags().Int64Var(&o.ChunkSize, "chunk-size", o.ChunkSize, "Return large lists in chunks rather than all at once. Pass 0 to disable. This flag is beta and may change in the future.")
	cmd.Flags().BoolVar(&o.IgnoreNotFound, "ignore-not-found", o.IgnoreNotFound, "If the requested object does not exist the command will return exit code 0.")
	cmd.Flags().StringVarP(&o.LabelSelector, "selector", "l", o.LabelSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().StringVar(&o.FieldSelector, "field-selector", o.FieldSelector, "Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.")
	cmd.Flags().BoolVar(&o.AllNamespaces, "all-namespaces", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmdutil.AddIncludeUninitializedFlag(cmd)
	addOpenAPIPrintColumnFlags(cmd)
	addServerPrintColumnFlags(cmd)
	cmd.Flags().BoolVar(&o.Export, "export", o.Export, "If true, use 'export' for the resources.  Exported resources are stripped of cluster-specific information.")
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, "identifying the resource to get from a server.")
	return cmd
}

// Complete takes the command arguments and factory and infers any remaining options.
func (o *GetOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if len(o.Raw) > 0 {
		if len(args) > 0 {
			return fmt.Errorf("arguments may not be passed when --raw is specified")
		}
		return nil
	}

	o.ServerPrint = cmdutil.GetFlagBool(cmd, useServerPrintColumns)

	var err error
	o.Namespace, o.ExplicitNamespace, err = f.DefaultNamespace()
	if err != nil {
		return err
	}
	if o.AllNamespaces {
		o.ExplicitNamespace = false
	}

	isSorting, err := cmd.Flags().GetString("sort-by")
	if err != nil {
		return err
	}
	o.Sort = len(isSorting) > 0

	o.NoHeaders = cmdutil.GetFlagBool(cmd, "no-headers")

	// TODO (soltysh): currently we don't support sorting and custom columns
	// with server side print. So in these cases force the old behavior.
	outputOption := cmd.Flags().Lookup("output").Value.String()
	if o.Sort && outputOption == "custom-columns" {
		o.ServerPrint = false
	}

	// human readable printers have special conversion rules, so we determine if we're using one.
	if len(*o.PrintFlags.OutputFormat) == 0 || *o.PrintFlags.OutputFormat == "wide" {
		o.IsHumanReadablePrinter = true
	}

	o.IncludeUninitialized = cmdutil.ShouldIncludeUninitialized(cmd, false)
	o.PrintWithOpenAPICols = cmdutil.GetFlagBool(cmd, useOpenAPIPrintColumnFlagLabel)

	o.ToPrinter = func(mapping *meta.RESTMapping, withNamespace bool) (printers.ResourcePrinterFunc, error) {
		// make a new copy of current flags / opts before mutating
		printFlags := o.PrintFlags.Copy()

		if mapping != nil {
			if !cmdSpecifiesOutputFmt(cmd) && o.PrintWithOpenAPICols {
				if apiSchema, err := f.OpenAPISchema(); err == nil {
					printFlags.UseOpenAPIColumns(apiSchema, mapping)
				}
			}
			if resource.MultipleTypesRequested(args) {
				printFlags.EnsureWithKind(mapping.GroupVersionKind.GroupKind())
			}
		}
		if withNamespace {
			printFlags.EnsureWithNamespace()
		}

		printer, err := printFlags.ToPrinter()
		if err != nil {
			return nil, err
		}
		return printer.PrintObj, nil
	}

	switch {
	case o.Watch || o.WatchOnly:
		// include uninitialized objects when watching on a single object
		// unless explicitly set --include-uninitialized=false
		o.IncludeUninitialized = cmdutil.ShouldIncludeUninitialized(cmd, len(args) == 2)
	default:
		if len(args) == 0 && cmdutil.IsFilenameSliceEmpty(o.Filenames) {
			fmt.Fprintf(o.ErrOut, "You must specify the type of resource to get. %s\n\n", cmdutil.SuggestApiResources(o.CmdParent))
			fullCmdName := cmd.Parent().CommandPath()
			usageString := "Required resource not specified."
			if len(fullCmdName) > 0 && cmdutil.IsSiblingCommandExists(cmd, "explain") {
				usageString = fmt.Sprintf("%s\nUse \"%s explain <resource>\" for a detailed description of that resource (e.g. %[2]s explain pods).", usageString, fullCmdName)
			}

			return cmdutil.UsageErrorf(cmd, usageString)
		}
	}
	return nil
}

// Validate checks the set of flags provided by the user.
func (o *GetOptions) Validate(cmd *cobra.Command) error {
	if len(o.Raw) > 0 {
		if o.Watch || o.WatchOnly || len(o.LabelSelector) > 0 || o.Export {
			return fmt.Errorf("--raw may not be specified with other flags that filter the server request or alter the output")
		}
		if len(cmdutil.GetFlagString(cmd, "output")) > 0 {
			return cmdutil.UsageErrorf(cmd, "--raw and --output are mutually exclusive")
		}
		if _, err := url.ParseRequestURI(o.Raw); err != nil {
			return cmdutil.UsageErrorf(cmd, "--raw must be a valid URL path: %v", err)
		}
	}
	if cmdutil.GetFlagBool(cmd, "show-labels") {
		outputOption := cmd.Flags().Lookup("output").Value.String()
		if outputOption != "" && outputOption != "wide" {
			return fmt.Errorf("--show-labels option cannot be used with %s printer", outputOption)
		}
	}
	return nil
}

// Run performs the get operation.
// TODO: remove the need to pass these arguments, like other commands.
func (o *GetOptions) Run(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if len(o.Raw) > 0 {
		return o.raw(f)
	}
	if o.Watch || o.WatchOnly {
		return o.watch(f, cmd, args)
	}

	r := f.NewBuilder().
		Unstructured().
		NamespaceParam(o.Namespace).DefaultNamespace().AllNamespaces(o.AllNamespaces).
		FilenameParam(o.ExplicitNamespace, &o.FilenameOptions).
		LabelSelectorParam(o.LabelSelector).
		FieldSelectorParam(o.FieldSelector).
		ExportParam(o.Export).
		RequestChunksOf(o.ChunkSize).
		IncludeUninitialized(o.IncludeUninitialized).
		ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().
		Latest().
		Flatten().
		TransformRequests(func(req *rest.Request) {
			if o.ServerPrint && o.IsHumanReadablePrinter && !o.Sort {
				group := metav1beta1.GroupName
				version := metav1beta1.SchemeGroupVersion.Version

				tableParam := fmt.Sprintf("application/json;as=Table;v=%s;g=%s, application/json", version, group)
				req.SetHeader("Accept", tableParam)
			}
		}).
		Do()

	if o.IgnoreNotFound {
		r.IgnoreErrors(kapierrors.IsNotFound)
	}
	if err := r.Err(); err != nil {
		return err
	}

	if !o.IsHumanReadablePrinter {
		return o.printGeneric(r)
	}

	allErrs := []error{}
	errs := sets.NewString()
	infos, err := r.Infos()
	if err != nil {
		allErrs = append(allErrs, err)
	}

	objs := make([]runtime.Object, len(infos))
	for ix := range infos {
		if o.ServerPrint {
			table, err := o.decodeIntoTable(cmdutil.InternalVersionJSONEncoder(), infos[ix].Object)
			if err == nil {
				infos[ix].Object = table
			} else {
				// if we are unable to decode server response into a v1beta1.Table,
				// fallback to client-side printing with whatever info the server returned.
				glog.V(2).Infof("Unable to decode server response into a Table. Falling back to hardcoded types: %v", err)
			}
		}

		objs[ix] = infos[ix].Object
	}

	sorting, err := cmd.Flags().GetString("sort-by")
	if err != nil {
		return err
	}
	var sorter *kubectl.RuntimeSort
	if o.Sort && len(objs) > 1 {
		// TODO: questionable
		if sorter, err = kubectl.SortObjects(cmdutil.InternalVersionDecoder(), objs, sorting); err != nil {
			return err
		}
	}

	var printer printers.ResourcePrinter
	var lastMapping *meta.RESTMapping
	nonEmptyObjCount := 0
	w := printers.GetNewTabWriter(o.Out)
	for ix := range objs {
		var mapping *meta.RESTMapping
		var info *resource.Info
		if sorter != nil {
			info = infos[sorter.OriginalPosition(ix)]
			mapping = info.Mapping
		} else {
			info = infos[ix]
			mapping = info.Mapping
		}

		// if dealing with a table that has no rows, skip remaining steps
		// and avoid printing an unnecessary newline
		if table, isTable := info.Object.(*metav1beta1.Table); isTable {
			if len(table.Rows) == 0 {
				continue
			}
		}

		nonEmptyObjCount++

		printWithNamespace := o.AllNamespaces
		if mapping != nil && mapping.Scope.Name() == meta.RESTScopeNameRoot {
			printWithNamespace = false
		}

		if shouldGetNewPrinterForMapping(printer, lastMapping, mapping) {
			w.Flush()

			// TODO: this doesn't belong here
			// add linebreak between resource groups (if there is more than one)
			// skip linebreak above first resource group
			if lastMapping != nil && !o.NoHeaders {
				fmt.Fprintln(o.ErrOut)
			}

			printer, err = o.ToPrinter(mapping, printWithNamespace)
			if err != nil {
				if !errs.Has(err.Error()) {
					errs.Insert(err.Error())
					allErrs = append(allErrs, err)
				}
				continue
			}

			lastMapping = mapping
		}

		internalObj, err := legacyscheme.Scheme.ConvertToVersion(info.Object, info.Mapping.GroupVersionKind.GroupKind().WithVersion(runtime.APIVersionInternal).GroupVersion())
		if err != nil {
			// if there's an error, try to print what you have (mirrors old behavior).
			glog.V(1).Info(err)
			printer.PrintObj(info.Object, w)
		} else {
			printer.PrintObj(internalObj, w)
		}
	}
	w.Flush()
	if nonEmptyObjCount == 0 && !o.IgnoreNotFound {
		fmt.Fprintln(o.ErrOut, "No resources found.")
	}
	return utilerrors.NewAggregate(allErrs)
}

// raw makes a simple HTTP request to the provided path on the server using the default
// credentials.
func (o *GetOptions) raw(f cmdutil.Factory) error {
	restClient, err := f.RESTClient()
	if err != nil {
		return err
	}

	stream, err := restClient.Get().RequestURI(o.Raw).Stream()
	if err != nil {
		return err
	}
	defer stream.Close()

	_, err = io.Copy(o.Out, stream)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

// watch starts a client-side watch of one or more resources.
// TODO: remove the need for arguments here.
func (o *GetOptions) watch(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	r := f.NewBuilder().
		Unstructured().
		NamespaceParam(o.Namespace).DefaultNamespace().AllNamespaces(o.AllNamespaces).
		FilenameParam(o.ExplicitNamespace, &o.FilenameOptions).
		LabelSelectorParam(o.LabelSelector).
		FieldSelectorParam(o.FieldSelector).
		ExportParam(o.Export).
		RequestChunksOf(o.ChunkSize).
		IncludeUninitialized(o.IncludeUninitialized).
		ResourceTypeOrNameArgs(true, args...).
		SingleResourceType().
		Latest().
		Do()
	if err := r.Err(); err != nil {
		return err
	}
	infos, err := r.Infos()
	if err != nil {
		return err
	}
	if len(infos) > 1 {
		gvk := infos[0].Mapping.GroupVersionKind
		uniqueGVKs := 1

		// If requesting a resource count greater than a request's --chunk-size,
		// we will end up making multiple requests to the server, with each
		// request producing its own "Info" object. Although overall we are
		// dealing with a single resource type, we will end up with multiple
		// infos returned by the builder. To handle this case, only fail if we
		// have at least one info with a different GVK than the others.
		for _, info := range infos {
			if info.Mapping.GroupVersionKind != gvk {
				uniqueGVKs++
			}
		}

		if uniqueGVKs > 1 {
			return i18n.Errorf("watch is only supported on individual resources and resource collections - %d resources were found", uniqueGVKs)
		}
	}

	info := infos[0]
	mapping := info.ResourceMapping()
	printer, err := o.ToPrinter(mapping, o.AllNamespaces)
	if err != nil {
		return err
	}
	obj, err := r.Object()
	if err != nil {
		return err
	}

	// watching from resourceVersion 0, starts the watch at ~now and
	// will return an initial watch event.  Starting form ~now, rather
	// the rv of the object will insure that we start the watch from
	// inside the watch window, which the rv of the object might not be.
	rv := "0"
	isList := meta.IsListType(obj)
	if isList {
		// the resourceVersion of list objects is ~now but won't return
		// an initial watch event
		rv, err = meta.NewAccessor().ResourceVersion(obj)
		if err != nil {
			return err
		}
	}

	// print the current object
	if !o.WatchOnly {
		var objsToPrint []runtime.Object
		writer := printers.GetNewTabWriter(o.Out)

		if isList {
			objsToPrint, _ = meta.ExtractList(obj)
		} else {
			objsToPrint = append(objsToPrint, obj)
		}
		for _, objToPrint := range objsToPrint {
			if o.IsHumanReadablePrinter {
				// printing always takes the internal version, but the watch event uses externals
				internalGV := mapping.GroupVersionKind.GroupKind().WithVersion(runtime.APIVersionInternal).GroupVersion()
				objToPrint = attemptToConvertToInternal(objToPrint, legacyscheme.Scheme, internalGV)
			}
			if err := printer.PrintObj(objToPrint, writer); err != nil {
				return fmt.Errorf("unable to output the provided object: %v", err)
			}
		}
		writer.Flush()
	}

	// print watched changes
	w, err := r.Watch(rv)
	if err != nil {
		return err
	}

	first := true
	intr := interrupt.New(nil, w.Stop)
	intr.Run(func() error {
		_, err := watch.Until(0, w, func(e watch.Event) (bool, error) {
			if !isList && first {
				// drop the initial watch event in the single resource case
				first = false
				return false, nil
			}

			// printing always takes the internal version, but the watch event uses externals
			// TODO fix printing to use server-side or be version agnostic
			internalGV := mapping.GroupVersionKind.GroupKind().WithVersion(runtime.APIVersionInternal).GroupVersion()
			if err := printer.PrintObj(attemptToConvertToInternal(e.Object, legacyscheme.Scheme, internalGV), o.Out); err != nil {
				return false, err
			}
			return false, nil
		})
		return err
	})
	return nil
}

// attemptToConvertToInternal tries to convert to an internal type, but returns the original if it can't
func attemptToConvertToInternal(obj runtime.Object, converter runtime.ObjectConvertor, targetVersion schema.GroupVersion) runtime.Object {
	internalObject, err := converter.ConvertToVersion(obj, targetVersion)
	if err != nil {
		glog.V(1).Infof("Unable to convert %T to %v: err", obj, targetVersion, err)
		return obj
	}
	return internalObject
}

func (o *GetOptions) decodeIntoTable(encoder runtime.Encoder, obj runtime.Object) (runtime.Object, error) {
	if obj.GetObjectKind().GroupVersionKind().Kind != "Table" {
		return nil, fmt.Errorf("attempt to decode non-Table object into a v1beta1.Table")
	}

	b, err := runtime.Encode(encoder, obj)
	if err != nil {
		return nil, err
	}

	table := &metav1beta1.Table{}
	err = json.Unmarshal(b, table)
	if err != nil {
		return nil, err
	}

	for i := range table.Rows {
		row := &table.Rows[i]
		if row.Object.Raw == nil || row.Object.Object != nil {
			//if row already has Object.Object
			//we don't change it
			continue
		}

		converted, err := runtime.Decode(unstructured.UnstructuredJSONScheme, row.Object.Raw)
		if err != nil {
			//if error happens, we just continue
			continue
		}
		row.Object.Object = converted
	}

	return table, nil
}

func (o *GetOptions) printGeneric(r *resource.Result) error {
	// we flattened the data from the builder, so we have individual items, but now we'd like to either:
	// 1. if there is more than one item, combine them all into a single list
	// 2. if there is a single item and that item is a list, leave it as its specific list
	// 3. if there is a single item and it is not a list, leave it as a single item
	var errs []error
	singleItemImplied := false
	infos, err := r.IntoSingleItemImplied(&singleItemImplied).Infos()
	if err != nil {
		if singleItemImplied {
			return err
		}
		errs = append(errs, err)
	}

	if len(infos) == 0 && o.IgnoreNotFound {
		return utilerrors.Reduce(utilerrors.Flatten(utilerrors.NewAggregate(errs)))
	}

	printer, err := o.ToPrinter(nil, false)
	if err != nil {
		return err
	}

	var obj runtime.Object
	if !singleItemImplied || len(infos) > 1 {
		// we have more than one item, so coerce all items into a list.
		// we don't want an *unstructured.Unstructured list yet, as we
		// may be dealing with non-unstructured objects. Compose all items
		// into an api.List, and then decode using an unstructured scheme.
		list := api.List{
			TypeMeta: metav1.TypeMeta{
				Kind:       "List",
				APIVersion: "v1",
			},
			ListMeta: metav1.ListMeta{},
		}
		for _, info := range infos {
			list.Items = append(list.Items, info.Object)
		}

		listData, err := json.Marshal(list)
		if err != nil {
			return err
		}

		converted, err := runtime.Decode(unstructured.UnstructuredJSONScheme, listData)
		if err != nil {
			return err
		}

		obj = converted
	} else {
		obj = infos[0].Object
	}

	isList := meta.IsListType(obj)
	if isList {
		items, err := meta.ExtractList(obj)
		if err != nil {
			return err
		}

		// take the items and create a new list for display
		list := &unstructured.UnstructuredList{
			Object: map[string]interface{}{
				"kind":       "List",
				"apiVersion": "v1",
				"metadata":   map[string]interface{}{},
			},
		}
		if listMeta, err := meta.ListAccessor(obj); err == nil {
			list.Object["metadata"] = map[string]interface{}{
				"selfLink":        listMeta.GetSelfLink(),
				"resourceVersion": listMeta.GetResourceVersion(),
			}
		}

		for _, item := range items {
			list.Items = append(list.Items, *item.(*unstructured.Unstructured))
		}
		if err := printer.PrintObj(list, o.Out); err != nil {
			errs = append(errs, err)
		}
		return utilerrors.Reduce(utilerrors.Flatten(utilerrors.NewAggregate(errs)))
	}

	if printErr := printer.PrintObj(obj, o.Out); printErr != nil {
		errs = append(errs, printErr)
	}

	return utilerrors.Reduce(utilerrors.Flatten(utilerrors.NewAggregate(errs)))
}

func addOpenAPIPrintColumnFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(useOpenAPIPrintColumnFlagLabel, false, "If true, use x-kubernetes-print-column metadata (if present) from the OpenAPI schema for displaying a resource.")
}

func addServerPrintColumnFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(useServerPrintColumns, true, "If true, have the server return the appropriate table output. Supports extension APIs and CRDs.")
}

func shouldGetNewPrinterForMapping(printer printers.ResourcePrinter, lastMapping, mapping *meta.RESTMapping) bool {
	return printer == nil || lastMapping == nil || mapping == nil || mapping.Resource != lastMapping.Resource
}

func cmdSpecifiesOutputFmt(cmd *cobra.Command) bool {
	return cmdutil.GetFlagString(cmd, "output") != ""
}

// outputOptsForMappingFromOpenAPI looks for the output format metatadata in the
// openapi schema and modifies the passed print options for the mapping if found.
func updatePrintOptionsForOpenAPI(f cmdutil.Factory, mapping *meta.RESTMapping, printOpts *printers.PrintOptions) bool {

	// user has not specified any output format, check if OpenAPI has
	// default specification to print this resource type
	api, err := f.OpenAPISchema()
	if err != nil {
		// Error getting schema
		return false
	}
	// Found openapi metadata for this resource
	schema := api.LookupResource(mapping.GroupVersionKind)
	if schema == nil {
		// Schema not found, return empty columns
		return false
	}

	columns, found := openapi.GetPrintColumns(schema.GetExtensions())
	if !found {
		// Extension not found, return empty columns
		return false
	}

	return outputOptsFromStr(columns, printOpts)
}

// outputOptsFromStr parses the print-column metadata and generates printer.OutputOptions object.
func outputOptsFromStr(columnStr string, printOpts *printers.PrintOptions) bool {
	if columnStr == "" {
		return false
	}
	parts := strings.SplitN(columnStr, "=", 2)
	if len(parts) < 2 {
		return false
	}

	printOpts.OutputFormatType = parts[0]
	printOpts.OutputFormatArgument = parts[1]
	printOpts.AllowMissingKeys = true

	return true
}
