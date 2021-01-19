/*
Copyright 2021 The KubeSphere Authors.

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

package porter

import (
	"context"
	"fmt"
	bgpconf "github.com/kubesphere/porter/api/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type bgpConfGetter struct {
	c cache.Cache
}

func NewBgpConfGetter(c cache.Cache) v1alpha3.Interface {
	return &bgpConfGetter{c}
}

func (d *bgpConfGetter) Get(namespace, name string) (runtime.Object, error) {
	conf := bgpconf.BgpConf{}
	err := d.c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func (d *bgpConfGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	return nil, fmt.Errorf("not support List method")
}
