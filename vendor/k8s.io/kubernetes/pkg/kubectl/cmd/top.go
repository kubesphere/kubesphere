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
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

var (
	supportedMetricsAPIVersions = []string{
		"v1beta1",
	}
	topLong = templates.LongDesc(i18n.T(`
		Display Resource (CPU/Memory/Storage) usage.

		The top command allows you to see the resource consumption for nodes or pods.

		This command requires Heapster to be correctly configured and working on the server. `))
)

func NewCmdTop(f cmdutil.Factory, out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top",
		Short: i18n.T("Display Resource (CPU/Memory/Storage) usage."),
		Long:  topLong,
		Run:   cmdutil.DefaultSubCommandRun(errOut),
	}

	// create subcommands
	cmd.AddCommand(NewCmdTopNode(f, nil, out))
	cmd.AddCommand(NewCmdTopPod(f, nil, out))

	return cmd
}

func SupportedMetricsAPIVersionAvailable(discoveredAPIGroups *metav1.APIGroupList) bool {
	for _, discoveredAPIGroup := range discoveredAPIGroups.Groups {
		if discoveredAPIGroup.Name != metricsapi.GroupName {
			continue
		}
		for _, version := range discoveredAPIGroup.Versions {
			for _, supportedVersion := range supportedMetricsAPIVersions {
				if version.Version == supportedVersion {
					return true
				}
			}
		}
	}
	return false
}
