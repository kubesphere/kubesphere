/*
Copyright 2019 The KubeSphere Authors.

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

package manifest

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/api/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type manifestGetter struct {
	informers externalversions.SharedInformerFactory
}

func New(informers externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &manifestGetter{
		informers: informers,
	}
}

func (c manifestGetter) Get(namespace, name string) (runtime.Object, error) {
	manifest, err := c.informers.Application().V1alpha1().Manifests().Lister().Get(name)
	if err != nil {
		return nil, err
	}
	return c.transform(manifest), nil
}

func (c manifestGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	manifests, err := c.informers.Application().V1alpha1().Manifests().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, manifest := range manifests {
		result = append(result, manifest)
	}

	return v1alpha3.DefaultList(result, query, c.compare, c.filter, c.transform), nil
}

func (c manifestGetter) transform(obj runtime.Object) runtime.Object {
	in := obj.(*v1alpha1.Manifest)
	out := in.DeepCopy()
	return out
}

func (c manifestGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftManifest, ok := left.(*v1alpha1.Manifest)
	if !ok {
		return false
	}

	rightManifest, ok := right.(*v1alpha1.Manifest)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftManifest.ObjectMeta, rightManifest.ObjectMeta, field)
}

func (c manifestGetter) filter(object runtime.Object, filter query.Filter) bool {
	manifest, ok := object.(*v1alpha1.Manifest)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(manifest.ObjectMeta, filter)
}
