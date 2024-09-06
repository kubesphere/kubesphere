/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package resourceparse

import (
	"io"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func Parse(reader io.Reader, namespace, rlsName string, local bool) ([]*resource.Info, error) {
	if klog.V(2).Enabled() {
		klog.Infof("parse resources, namespace: %s, release: %s", namespace, rlsName)
		start := time.Now()
		defer func() {
			klog.Infof("parse resources end, namespace: %s, release: %s, cost: %v", namespace, rlsName, time.Since(start))
		}()
	}

	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	builder := f.NewBuilder().Unstructured().NamespaceParam(namespace).ContinueOnError().Stream(reader, rlsName).Flatten()

	if local {
		builder = builder.Local()
	}
	r := builder.Do()
	infos, err := r.Infos()
	if err != nil {
		return nil, err
	}

	if !local {
		for i := range infos {
			infos[i].Namespace = namespace
			err := infos[i].Get()
			if err != nil {
				return nil, err
			}
		}
	}

	return infos, err
}
