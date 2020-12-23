package resourceparse

import (
	"io"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"time"
)

func Parse(reader io.Reader, namespace, rlsName string, local bool) ([]*resource.Info, error) {
	if klog.V(2) {
		klog.Infof("parse resources, namespace: %s, release: %s", namespace, rlsName)
		start := time.Now()
		defer func() {
			klog.Infof("parse resources end, namespace: %s, release: %s, cost: %v", namespace, rlsName, time.Now().Sub(start))
		}()
	}

	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	builder := f.NewBuilder().Unstructured().NamespaceParam(namespace).ContinueOnError().Stream(reader, rlsName).Flatten()

	if local == true {
		builder = builder.Local()
	}
	r := builder.Do()
	infos, err := r.Infos()
	if err != nil {
		return nil, err
	}

	if local == false {
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
