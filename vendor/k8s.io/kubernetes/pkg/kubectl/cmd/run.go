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

package cmd

import (
	"fmt"
	"io"

	"k8s.io/client-go/dynamic"
	"k8s.io/kubernetes/pkg/printers"

	"github.com/docker/distribution/reference"
	"github.com/spf13/cobra"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	api "k8s.io/kubernetes/pkg/apis/core"
	coreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/util/interrupt"
	uexec "k8s.io/utils/exec"
)

var (
	runLong = templates.LongDesc(i18n.T(`
		Create and run a particular image, possibly replicated.

		Creates a deployment or job to manage the created container(s).`))

	runExample = templates.Examples(i18n.T(`
		# Start a single instance of nginx.
		kubectl run nginx --image=nginx

		# Start a single instance of hazelcast and let the container expose port 5701 .
		kubectl run hazelcast --image=hazelcast --port=5701

		# Start a single instance of hazelcast and set environment variables "DNS_DOMAIN=cluster" and "POD_NAMESPACE=default" in the container.
		kubectl run hazelcast --image=hazelcast --env="DNS_DOMAIN=cluster" --env="POD_NAMESPACE=default"

		# Start a single instance of hazelcast and set labels "app=hazelcast" and "env=prod" in the container.
		kubectl run hazelcast --image=nginx --labels="app=hazelcast,env=prod"

		# Start a replicated instance of nginx.
		kubectl run nginx --image=nginx --replicas=5

		# Dry run. Print the corresponding API objects without creating them.
		kubectl run nginx --image=nginx --dry-run

		# Start a single instance of nginx, but overload the spec of the deployment with a partial set of values parsed from JSON.
		kubectl run nginx --image=nginx --overrides='{ "apiVersion": "v1", "spec": { ... } }'

		# Start a pod of busybox and keep it in the foreground, don't restart it if it exits.
		kubectl run -i -t busybox --image=busybox --restart=Never

		# Start the nginx container using the default command, but use custom arguments (arg1 .. argN) for that command.
		kubectl run nginx --image=nginx -- <arg1> <arg2> ... <argN>

		# Start the nginx container using a different command and custom arguments.
		kubectl run nginx --image=nginx --command -- <cmd> <arg1> ... <argN>

		# Start the perl container to compute π to 2000 places and print it out.
		kubectl run pi --image=perl --restart=OnFailure -- perl -Mbignum=bpi -wle 'print bpi(2000)'

		# Start the cron job to compute π to 2000 places and print it out every 5 minutes.
		kubectl run pi --schedule="0/5 * * * ?" --image=perl --restart=OnFailure -- perl -Mbignum=bpi -wle 'print bpi(2000)'`))
)

type RunObject struct {
	Object  runtime.Object
	Mapping *meta.RESTMapping
}

type RunOptions struct {
	PrintFlags    *printers.PrintFlags
	DeleteFlags   *DeleteFlags
	DeleteOptions *DeleteOptions
	RecordFlags   *genericclioptions.RecordFlags

	DryRun bool

	PrintObj func(runtime.Object) error
	Recorder genericclioptions.Recorder

	DynamicClient dynamic.DynamicInterface

	ArgsLenAtDash  int
	Attach         bool
	Expose         bool
	Generator      string
	Image          string
	Interactive    bool
	LeaveStdinOpen bool
	Port           string
	Quiet          bool
	Schedule       string
	TTY            bool

	genericclioptions.IOStreams
}

func NewRunOptions(streams genericclioptions.IOStreams) *RunOptions {
	return &RunOptions{
		PrintFlags:  printers.NewPrintFlags("created"),
		DeleteFlags: NewDeleteFlags("to use to replace the resource."),
		RecordFlags: genericclioptions.NewRecordFlags(),

		Recorder: genericclioptions.NoopRecorder{},

		IOStreams: streams,
	}
}

func NewCmdRun(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewRunOptions(streams)

	cmd := &cobra.Command{
		Use: "run NAME --image=image [--env=\"key=value\"] [--port=port] [--replicas=replicas] [--dry-run=bool] [--overrides=inline-json] [--command] -- [COMMAND] [args...]",
		DisableFlagsInUseLine: true,
		Short:   i18n.T("Run a particular image on the cluster"),
		Long:    runLong,
		Example: runExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd))
			cmdutil.CheckErr(o.Run(f, cmd, args))
		},
	}

	o.DeleteFlags.AddFlags(cmd)
	o.PrintFlags.AddFlags(cmd)
	o.RecordFlags.AddFlags(cmd)

	addRunFlags(cmd)
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddPodRunningTimeoutFlag(cmd, defaultPodAttachTimeout)
	return cmd
}

func addRunFlags(cmd *cobra.Command) {
	cmdutil.AddDryRunFlag(cmd)
	cmd.Flags().String("generator", "", i18n.T("The name of the API generator to use, see http://kubernetes.io/docs/user-guide/kubectl-conventions/#generators for a list."))
	cmd.Flags().String("image", "", i18n.T("The image for the container to run."))
	cmd.MarkFlagRequired("image")
	cmd.Flags().String("image-pull-policy", "", i18n.T("The image pull policy for the container. If left empty, this value will not be specified by the client and defaulted by the server"))
	cmd.Flags().IntP("replicas", "r", 1, "Number of replicas to create for this container. Default is 1.")
	cmd.Flags().Bool("rm", false, "If true, delete resources created in this command for attached containers.")
	cmd.Flags().String("overrides", "", i18n.T("An inline JSON override for the generated object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field."))
	cmd.Flags().StringArray("env", []string{}, "Environment variables to set in the container")
	cmd.Flags().String("serviceaccount", "", "Service account to set in the pod spec")
	cmd.Flags().String("port", "", i18n.T("The port that this container exposes.  If --expose is true, this is also the port used by the service that is created."))
	cmd.Flags().Int("hostport", -1, "The host port mapping for the container port. To demonstrate a single-machine container.")
	cmd.Flags().StringP("labels", "l", "", "Comma separated labels to apply to the pod(s). Will override previous values.")
	cmd.Flags().BoolP("stdin", "i", false, "Keep stdin open on the container(s) in the pod, even if nothing is attached.")
	cmd.Flags().BoolP("tty", "t", false, "Allocated a TTY for each container in the pod.")
	cmd.Flags().Bool("attach", false, "If true, wait for the Pod to start running, and then attach to the Pod as if 'kubectl attach ...' were called.  Default false, unless '-i/--stdin' is set, in which case the default is true. With '--restart=Never' the exit code of the container process is returned.")
	cmd.Flags().Bool("leave-stdin-open", false, "If the pod is started in interactive mode or with stdin, leave stdin open after the first attach completes. By default, stdin will be closed after the first attach completes.")
	cmd.Flags().String("restart", "Always", i18n.T("The restart policy for this Pod.  Legal values [Always, OnFailure, Never].  If set to 'Always' a deployment is created, if set to 'OnFailure' a job is created, if set to 'Never', a regular pod is created. For the latter two --replicas must be 1.  Default 'Always', for CronJobs `Never`."))
	cmd.Flags().Bool("command", false, "If true and extra arguments are present, use them as the 'command' field in the container, rather than the 'args' field which is the default.")
	cmd.Flags().String("requests", "", i18n.T("The resource requirement requests for this container.  For example, 'cpu=100m,memory=256Mi'.  Note that server side components may assign requests depending on the server configuration, such as limit ranges."))
	cmd.Flags().String("limits", "", i18n.T("The resource requirement limits for this container.  For example, 'cpu=200m,memory=512Mi'.  Note that server side components may assign limits depending on the server configuration, such as limit ranges."))
	cmd.Flags().Bool("expose", false, "If true, a public, external service is created for the container(s) which are run")
	cmd.Flags().String("service-generator", "service/v2", i18n.T("The name of the generator to use for creating a service.  Only used if --expose is true"))
	cmd.Flags().String("service-overrides", "", i18n.T("An inline JSON override for the generated service object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field.  Only used if --expose is true."))
	cmd.Flags().Bool("quiet", false, "If true, suppress prompt messages.")
	cmd.Flags().String("schedule", "", i18n.T("A schedule in the Cron format the job should be run with."))
}

func (o *RunOptions) Complete(f cmdutil.Factory, cmd *cobra.Command) error {
	var err error

	o.RecordFlags.Complete(f.Command(cmd, false))
	o.Recorder, err = o.RecordFlags.ToRecorder()
	if err != nil {
		return err
	}

	o.DynamicClient, err = f.DynamicClient()
	if err != nil {
		return err
	}

	o.ArgsLenAtDash = cmd.ArgsLenAtDash()
	o.DryRun = cmdutil.GetFlagBool(cmd, "dry-run")
	o.Expose = cmdutil.GetFlagBool(cmd, "expose")
	o.Generator = cmdutil.GetFlagString(cmd, "generator")
	o.Image = cmdutil.GetFlagString(cmd, "image")
	o.Interactive = cmdutil.GetFlagBool(cmd, "stdin")
	o.LeaveStdinOpen = cmdutil.GetFlagBool(cmd, "leave-stdin-open")
	o.Port = cmdutil.GetFlagString(cmd, "port")
	o.Quiet = cmdutil.GetFlagBool(cmd, "quiet")
	o.Schedule = cmdutil.GetFlagString(cmd, "schedule")
	o.TTY = cmdutil.GetFlagBool(cmd, "tty")

	attachFlag := cmd.Flags().Lookup("attach")
	o.Attach = cmdutil.GetFlagBool(cmd, "attach")
	if !attachFlag.Changed && o.Interactive {
		o.Attach = true
	}

	if o.DryRun {
		o.PrintFlags.Complete("%s (dry run)")
	}
	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrintObj = func(obj runtime.Object) error {
		return printer.PrintObj(obj, o.Out)
	}

	deleteOpts := o.DeleteFlags.ToOptions(o.Out, o.ErrOut)
	deleteOpts.IgnoreNotFound = true
	deleteOpts.WaitForDeletion = false
	deleteOpts.GracePeriod = -1
	deleteOpts.Reaper = f.Reaper

	o.DeleteOptions = deleteOpts

	return nil
}

func (o *RunOptions) Run(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	// Let kubectl run follow rules for `--`, see #13004 issue
	if len(args) == 0 || o.ArgsLenAtDash == 0 {
		return cmdutil.UsageErrorf(cmd, "NAME is required for run")
	}

	timeout, err := cmdutil.GetPodRunningTimeoutFlag(cmd)
	if err != nil {
		return cmdutil.UsageErrorf(cmd, "%v", err)
	}

	// validate image name
	imageName := o.Image
	if imageName == "" {
		return fmt.Errorf("--image is required")
	}
	validImageRef := reference.ReferenceRegexp.MatchString(imageName)
	if !validImageRef {
		return fmt.Errorf("Invalid image name %q: %v", imageName, reference.ErrReferenceInvalidFormat)
	}

	if o.TTY && !o.Interactive {
		return cmdutil.UsageErrorf(cmd, "-i/--stdin is required for containers with -t/--tty=true")
	}
	replicas := cmdutil.GetFlagInt(cmd, "replicas")
	if o.Interactive && replicas != 1 {
		return cmdutil.UsageErrorf(cmd, "-i/--stdin requires that replicas is 1, found %d", replicas)
	}
	if o.Expose && len(o.Port) == 0 {
		return cmdutil.UsageErrorf(cmd, "--port must be set when exposing a service")
	}

	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	restartPolicy, err := getRestartPolicy(cmd, o.Interactive)
	if err != nil {
		return err
	}
	if restartPolicy != api.RestartPolicyAlways && replicas != 1 {
		return cmdutil.UsageErrorf(cmd, "--restart=%s requires that --replicas=1, found %d", restartPolicy, replicas)
	}

	remove := cmdutil.GetFlagBool(cmd, "rm")
	if !o.Attach && remove {
		return cmdutil.UsageErrorf(cmd, "--rm should only be used for attached containers")
	}

	if o.Attach && o.DryRun {
		return cmdutil.UsageErrorf(cmd, "--dry-run can't be used with attached containers options (--attach, --stdin, or --tty)")
	}

	if err := verifyImagePullPolicy(cmd); err != nil {
		return err
	}

	clientset, err := f.ClientSet()
	if err != nil {
		return err
	}

	generatorName := o.Generator
	if len(o.Schedule) != 0 && len(generatorName) == 0 {
		generatorName = cmdutil.CronJobV1Beta1GeneratorName
	}
	if len(generatorName) == 0 {
		switch restartPolicy {
		case api.RestartPolicyAlways:
			generatorName = cmdutil.DeploymentAppsV1Beta1GeneratorName
		case api.RestartPolicyOnFailure:
			generatorName = cmdutil.JobV1GeneratorName
		case api.RestartPolicyNever:
			generatorName = cmdutil.RunPodV1GeneratorName
		}

		// Falling back because the generator was not provided and the default one could be unavailable.
		generatorNameTemp, err := cmdutil.FallbackGeneratorNameIfNecessary(generatorName, clientset.Discovery(), o.ErrOut)
		if err != nil {
			return err
		}
		if generatorNameTemp != generatorName {
			cmdutil.Warning(o.ErrOut, generatorName, generatorNameTemp)
		} else {
			generatorName = generatorNameTemp
		}
	}

	generators := f.Generators("run")
	generator, found := generators[generatorName]
	if !found {
		return cmdutil.UsageErrorf(cmd, "generator %q not found", generatorName)
	}
	names := generator.ParamNames()
	params := kubectl.MakeParams(cmd, names)
	params["name"] = args[0]
	if len(args) > 1 {
		params["args"] = args[1:]
	}

	params["env"] = cmdutil.GetFlagStringArray(cmd, "env")

	var createdObjects = []*RunObject{}
	runObject, err := o.createGeneratedObject(f, cmd, generator, names, params, cmdutil.GetFlagString(cmd, "overrides"), namespace)
	if err != nil {
		return err
	} else {
		createdObjects = append(createdObjects, runObject)
	}
	allErrs := []error{}
	if o.Expose {
		serviceGenerator := cmdutil.GetFlagString(cmd, "service-generator")
		if len(serviceGenerator) == 0 {
			return cmdutil.UsageErrorf(cmd, "No service generator specified")
		}
		serviceRunObject, err := o.generateService(f, cmd, serviceGenerator, params, namespace)
		if err != nil {
			allErrs = append(allErrs, err)
		} else {
			createdObjects = append(createdObjects, serviceRunObject)
		}
	}

	if o.Attach {
		if remove {
			defer o.removeCreatedObjects(f, createdObjects)
		}

		opts := &AttachOptions{
			StreamOptions: StreamOptions{
				In:    o.In,
				Out:   o.Out,
				Err:   o.ErrOut,
				Stdin: o.Interactive,
				TTY:   o.TTY,
				Quiet: o.Quiet,
			},
			GetPodTimeout: timeout,
			CommandName:   cmd.Parent().CommandPath() + " attach",

			Attach: &DefaultRemoteAttach{},
		}
		config, err := f.ClientConfig()
		if err != nil {
			return err
		}
		opts.Config = config

		clientset, err := f.ClientSet()
		if err != nil {
			return err
		}
		opts.PodClient = clientset.Core()

		attachablePod, err := f.AttachablePodForObject(runObject.Object, opts.GetPodTimeout)
		if err != nil {
			return err
		}
		err = handleAttachPod(f, clientset.Core(), attachablePod.Namespace, attachablePod.Name, opts)
		if err != nil {
			return err
		}

		var pod *api.Pod
		leaveStdinOpen := o.LeaveStdinOpen
		waitForExitCode := !leaveStdinOpen && restartPolicy == api.RestartPolicyNever
		if waitForExitCode {
			pod, err = waitForPod(clientset.Core(), attachablePod.Namespace, attachablePod.Name, kubectl.PodCompleted)
			if err != nil {
				return err
			}
		}

		// after removal is done, return successfully if we are not interested in the exit code
		if !waitForExitCode {
			return nil
		}

		switch pod.Status.Phase {
		case api.PodSucceeded:
			return nil
		case api.PodFailed:
			unknownRcErr := fmt.Errorf("pod %s/%s failed with unknown exit code", pod.Namespace, pod.Name)
			if len(pod.Status.ContainerStatuses) == 0 || pod.Status.ContainerStatuses[0].State.Terminated == nil {
				return unknownRcErr
			}
			// assume here that we have at most one status because kubectl-run only creates one container per pod
			rc := pod.Status.ContainerStatuses[0].State.Terminated.ExitCode
			if rc == 0 {
				return unknownRcErr
			}
			return uexec.CodeExitError{
				Err:  fmt.Errorf("pod %s/%s terminated (%s)\n%s", pod.Namespace, pod.Name, pod.Status.ContainerStatuses[0].State.Terminated.Reason, pod.Status.ContainerStatuses[0].State.Terminated.Message),
				Code: int(rc),
			}
		default:
			return fmt.Errorf("pod %s/%s left in phase %s", pod.Namespace, pod.Name, pod.Status.Phase)
		}

	}
	if runObject != nil {
		if err := o.PrintObj(runObject.Object); err != nil {
			return err
		}
	}

	return utilerrors.NewAggregate(allErrs)
}

func (o *RunOptions) removeCreatedObjects(f cmdutil.Factory, createdObjects []*RunObject) error {
	for _, obj := range createdObjects {
		namespace, err := metadataAccessor.Namespace(obj.Object)
		if err != nil {
			return err
		}
		var name string
		name, err = metadataAccessor.Name(obj.Object)
		if err != nil {
			return err
		}
		r := f.NewBuilder().
			WithScheme(legacyscheme.Scheme).
			ContinueOnError().
			NamespaceParam(namespace).DefaultNamespace().
			ResourceNames(obj.Mapping.Resource.Resource+"."+obj.Mapping.Resource.Group, name).
			Flatten().
			Do()
		// Note: we pass in "true" for the "quiet" parameter because
		// ReadResult will only print one thing based on the "quiet"
		// flag, and that's the "pod xxx deleted" message. If they
		// asked for us to remove the pod (via --rm) then telling them
		// its been deleted is unnecessary since that's what they asked
		// for. We should only print something if the "rm" fails.
		err = o.DeleteOptions.ReapResult(r, true, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// waitForPod watches the given pod until the exitCondition is true
func waitForPod(podClient coreclient.PodsGetter, ns, name string, exitCondition watch.ConditionFunc) (*api.Pod, error) {
	w, err := podClient.Pods(ns).Watch(metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	if err != nil {
		return nil, err
	}

	intr := interrupt.New(nil, w.Stop)
	var result *api.Pod
	err = intr.Run(func() error {
		ev, err := watch.Until(0, w, func(ev watch.Event) (bool, error) {
			return exitCondition(ev)
		})
		if ev != nil {
			result = ev.Object.(*api.Pod)
		}
		return err
	})

	// Fix generic not found error.
	if err != nil && errors.IsNotFound(err) {
		err = errors.NewNotFound(api.Resource("pods"), name)
	}

	return result, err
}

func handleAttachPod(f cmdutil.Factory, podClient coreclient.PodsGetter, ns, name string, opts *AttachOptions) error {
	pod, err := waitForPod(podClient, ns, name, kubectl.PodRunningAndReady)
	if err != nil && err != kubectl.ErrPodCompleted {
		return err
	}

	if pod.Status.Phase == api.PodSucceeded || pod.Status.Phase == api.PodFailed {
		return logOpts(f, pod, opts)
	}

	opts.PodClient = podClient
	opts.PodName = name
	opts.Namespace = ns

	// TODO: opts.Run sets opts.Err to nil, we need to find a better way
	stderr := opts.Err
	if err := opts.Run(); err != nil {
		fmt.Fprintf(stderr, "Error attaching, falling back to logs: %v\n", err)
		return logOpts(f, pod, opts)
	}
	return nil
}

// logOpts logs output from opts to the pods log.
func logOpts(f cmdutil.Factory, pod *api.Pod, opts *AttachOptions) error {
	ctrName, err := opts.GetContainerName(pod)
	if err != nil {
		return err
	}

	req, err := f.LogsForObject(pod, &api.PodLogOptions{Container: ctrName}, opts.GetPodTimeout)
	if err != nil {
		return err
	}

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	_, err = io.Copy(opts.Out, readCloser)
	if err != nil {
		return err
	}
	return nil
}

func getRestartPolicy(cmd *cobra.Command, interactive bool) (api.RestartPolicy, error) {
	restart := cmdutil.GetFlagString(cmd, "restart")
	if len(restart) == 0 {
		if interactive {
			return api.RestartPolicyOnFailure, nil
		} else {
			return api.RestartPolicyAlways, nil
		}
	}
	switch api.RestartPolicy(restart) {
	case api.RestartPolicyAlways:
		return api.RestartPolicyAlways, nil
	case api.RestartPolicyOnFailure:
		return api.RestartPolicyOnFailure, nil
	case api.RestartPolicyNever:
		return api.RestartPolicyNever, nil
	}
	return "", cmdutil.UsageErrorf(cmd, "invalid restart policy: %s", restart)
}

func verifyImagePullPolicy(cmd *cobra.Command) error {
	pullPolicy := cmdutil.GetFlagString(cmd, "image-pull-policy")
	switch api.PullPolicy(pullPolicy) {
	case api.PullAlways, api.PullIfNotPresent, api.PullNever:
		return nil
	case "":
		return nil
	}
	return cmdutil.UsageErrorf(cmd, "invalid image pull policy: %s", pullPolicy)
}

func (o *RunOptions) generateService(f cmdutil.Factory, cmd *cobra.Command, serviceGenerator string, paramsIn map[string]interface{}, namespace string) (*RunObject, error) {
	generators := f.Generators("expose")
	generator, found := generators[serviceGenerator]
	if !found {
		return nil, fmt.Errorf("missing service generator: %s", serviceGenerator)
	}
	names := generator.ParamNames()

	params := map[string]interface{}{}
	for key, value := range paramsIn {
		_, isString := value.(string)
		if isString {
			params[key] = value
		}
	}

	name, found := params["name"]
	if !found || len(name.(string)) == 0 {
		return nil, fmt.Errorf("name is a required parameter")
	}
	selector, found := params["labels"]
	if !found || len(selector.(string)) == 0 {
		selector = fmt.Sprintf("run=%s", name.(string))
	}
	params["selector"] = selector

	if defaultName, found := params["default-name"]; !found || len(defaultName.(string)) == 0 {
		params["default-name"] = name
	}

	runObject, err := o.createGeneratedObject(f, cmd, generator, names, params, cmdutil.GetFlagString(cmd, "service-overrides"), namespace)
	if err != nil {
		return nil, err
	}

	if err := o.PrintObj(runObject.Object); err != nil {
		return nil, err
	}
	// separate yaml objects
	if o.PrintFlags.OutputFormat != nil && *o.PrintFlags.OutputFormat == "yaml" {
		fmt.Fprintln(o.Out, "---")
	}

	return runObject, nil
}

func (o *RunOptions) createGeneratedObject(f cmdutil.Factory, cmd *cobra.Command, generator kubectl.Generator, names []kubectl.GeneratorParam, params map[string]interface{}, overrides, namespace string) (*RunObject, error) {
	err := kubectl.ValidateParams(names, params)
	if err != nil {
		return nil, err
	}

	// TODO: Validate flag usage against selected generator. More tricky since --expose was added.
	obj, err := generator.Generate(params)
	if err != nil {
		return nil, err
	}

	mapper, err := f.RESTMapper()
	if err != nil {
		return nil, err
	}
	// run has compiled knowledge of the thing is is creating
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return nil, err
	}
	mapping, err := mapper.RESTMapping(gvks[0].GroupKind(), gvks[0].Version)
	if err != nil {
		return nil, err
	}

	if len(overrides) > 0 {
		codec := runtime.NewCodec(scheme.DefaultJSONEncoder(), scheme.Codecs.UniversalDecoder(scheme.Registry.RegisteredGroupVersions()...))
		obj, err = cmdutil.Merge(codec, obj, overrides)
		if err != nil {
			return nil, err
		}
	}

	if err := o.Recorder.Record(obj); err != nil {
		glog.V(4).Infof("error recording current command: %v", err)
	}

	actualObj := obj
	if !o.DryRun {
		if err := kubectl.CreateOrUpdateAnnotation(cmdutil.GetFlagBool(cmd, cmdutil.ApplyAnnotationsFlag), obj, scheme.DefaultJSONEncoder()); err != nil {
			return nil, err
		}
		client, err := f.ClientForMapping(mapping)
		if err != nil {
			return nil, err
		}
		actualObj, err = resource.NewHelper(client, mapping).Create(namespace, false, obj)
		if err != nil {
			return nil, err
		}
	}
	actualObj = cmdutil.AsDefaultVersionedOrOriginal(actualObj, mapping)

	return &RunObject{
		Object:  actualObj,
		Mapping: mapping,
	}, nil
}
