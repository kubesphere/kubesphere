// /*
// Copyright 2020 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */
//

package resourceparse

import (
	"io"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
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
