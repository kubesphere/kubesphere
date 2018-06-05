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

package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/printers"
)

type ClusterInfoDumpOptions struct {
	PrintFlags *printers.PrintFlags
	PrintObj   printers.ResourcePrinterFunc

	OutputDir     string
	AllNamespaces bool
	Namespaces    []string

	timeout       time.Duration
	clientset     internalclientset.Interface
	namespace     string
	logsForObject func(object, options runtime.Object, timeout time.Duration) (*restclient.Request, error)

	genericclioptions.IOStreams
}

// NewCmdCreateSecret groups subcommands to create various types of secrets
func NewCmdClusterInfoDump(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := &ClusterInfoDumpOptions{
		PrintFlags: printers.NewPrintFlags(""),

		IOStreams: ioStreams,
	}

	cmd := &cobra.Command{
		Use:     "dump",
		Short:   i18n.T("Dump lots of relevant info for debugging and diagnosis"),
		Long:    dumpLong,
		Example: dumpExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd))
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVar(&o.OutputDir, "output-directory", "", i18n.T("Where to output the files.  If empty or '-' uses stdout, otherwise creates a directory hierarchy in that directory"))
	cmd.Flags().StringSliceVar(&o.Namespaces, "namespaces", []string{}, "A comma separated list of namespaces to dump.")
	cmd.Flags().BoolVar(&o.AllNamespaces, "all-namespaces", false, "If true, dump all namespaces.  If true, --namespaces is ignored.")
	cmdutil.AddPodRunningTimeoutFlag(cmd, defaultPodLogsTimeout)
	return cmd
}

var (
	dumpLong = templates.LongDesc(i18n.T(`
    Dumps cluster info out suitable for debugging and diagnosing cluster problems.  By default, dumps everything to
    stdout. You can optionally specify a directory with --output-directory.  If you specify a directory, kubernetes will
    build a set of files in that directory.  By default only dumps things in the 'kube-system' namespace, but you can
    switch to a different namespace with the --namespaces flag, or specify --all-namespaces to dump all namespaces.

    The command also dumps the logs of all of the pods in the cluster, these logs are dumped into different directories
    based on namespace and pod name.`))

	dumpExample = templates.Examples(i18n.T(`
    # Dump current cluster state to stdout
    kubectl cluster-info dump

    # Dump current cluster state to /path/to/cluster-state
    kubectl cluster-info dump --output-directory=/path/to/cluster-state

    # Dump all namespaces to stdout
    kubectl cluster-info dump --all-namespaces

    # Dump a set of namespaces to /path/to/cluster-state
    kubectl cluster-info dump --namespaces default,kube-system --output-directory=/path/to/cluster-state`))
)

func setupOutputWriter(dir string, defaultWriter io.Writer, filename string) io.Writer {
	if len(dir) == 0 || dir == "-" {
		return defaultWriter
	}
	fullFile := path.Join(dir, filename)
	parent := path.Dir(fullFile)
	cmdutil.CheckErr(os.MkdirAll(parent, 0755))

	file, err := os.Create(path.Join(dir, filename))
	cmdutil.CheckErr(err)
	return file
}

func (o *ClusterInfoDumpOptions) Complete(f cmdutil.Factory, cmd *cobra.Command) error {
	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	jsonOutputFmt := "json"
	o.PrintFlags.OutputFormat = &jsonOutputFmt
	o.PrintObj = printer.PrintObj

	o.timeout, err = cmdutil.GetPodRunningTimeoutFlag(cmd)
	if err != nil {
		return err
	}
	o.clientset, err = f.ClientSet()
	if err != nil {
		return err
	}
	o.namespace, _, err = f.DefaultNamespace()
	if err != nil {
		return err
	}
	o.logsForObject = f.LogsForObject

	return nil
}

func (o *ClusterInfoDumpOptions) Run() error {
	nodes, err := o.clientset.Core().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if err := o.PrintObj(nodes, setupOutputWriter(o.OutputDir, o.Out, "nodes.json")); err != nil {
		return err
	}

	var namespaces []string
	if o.AllNamespaces {
		namespaceList, err := o.clientset.Core().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		for ix := range namespaceList.Items {
			namespaces = append(namespaces, namespaceList.Items[ix].Name)
		}
	} else {
		if len(o.Namespaces) == 0 {
			namespaces = []string{
				metav1.NamespaceSystem,
				o.namespace,
			}
		}
	}
	for _, namespace := range namespaces {
		// TODO: this is repetitive in the extreme.  Use reflection or
		// something to make this a for loop.
		events, err := o.clientset.Core().Events(namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if err := o.PrintObj(events, setupOutputWriter(o.OutputDir, o.Out, path.Join(namespace, "events.json"))); err != nil {
			return err
		}

		rcs, err := o.clientset.Core().ReplicationControllers(namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if err := o.PrintObj(rcs, setupOutputWriter(o.OutputDir, o.Out, path.Join(namespace, "replication-controllers.json"))); err != nil {
			return err
		}

		svcs, err := o.clientset.Core().Services(namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if err := o.PrintObj(svcs, setupOutputWriter(o.OutputDir, o.Out, path.Join(namespace, "services.json"))); err != nil {
			return err
		}

		sets, err := o.clientset.Extensions().DaemonSets(namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if err := o.PrintObj(sets, setupOutputWriter(o.OutputDir, o.Out, path.Join(namespace, "daemonsets.json"))); err != nil {
			return err
		}

		deps, err := o.clientset.Extensions().Deployments(namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if err := o.PrintObj(deps, setupOutputWriter(o.OutputDir, o.Out, path.Join(namespace, "deployments.json"))); err != nil {
			return err
		}

		rps, err := o.clientset.Extensions().ReplicaSets(namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if err := o.PrintObj(rps, setupOutputWriter(o.OutputDir, o.Out, path.Join(namespace, "replicasets.json"))); err != nil {
			return err
		}

		pods, err := o.clientset.Core().Pods(namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		if err := o.PrintObj(pods, setupOutputWriter(o.OutputDir, o.Out, path.Join(namespace, "pods.json"))); err != nil {
			return err
		}

		printContainer := func(writer io.Writer, container api.Container, pod *api.Pod) {
			writer.Write([]byte(fmt.Sprintf("==== START logs for container %s of pod %s/%s ====\n", container.Name, pod.Namespace, pod.Name)))
			defer writer.Write([]byte(fmt.Sprintf("==== END logs for container %s of pod %s/%s ====\n", container.Name, pod.Namespace, pod.Name)))

			request, err := o.logsForObject(pod, &api.PodLogOptions{Container: container.Name}, timeout)
			if err != nil {
				// Print error and return.
				writer.Write([]byte(fmt.Sprintf("Create log request error: %s\n", err.Error())))
				return
			}

			data, err := request.DoRaw()
			if err != nil {
				// Print error and return.
				writer.Write([]byte(fmt.Sprintf("Request log error: %s\n", err.Error())))
				return
			}
			writer.Write(data)
		}

		for ix := range pods.Items {
			pod := &pods.Items[ix]
			containers := pod.Spec.Containers
			writer := setupOutputWriter(o.OutputDir, o.Out, path.Join(namespace, pod.Name, "logs.txt"))

			for i := range containers {
				printContainer(writer, containers[i], pod)
			}
		}
	}
	if o.OutputDir != "-" {
		fmt.Fprintf(o.Out, "Cluster info dumped to %s\n", o.OutputDir)
	}
	return nil
}
