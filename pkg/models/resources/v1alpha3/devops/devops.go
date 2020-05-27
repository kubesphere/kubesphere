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

package devops

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type devopsGetter struct {
	informers ksinformers.SharedInformerFactory
}

func New(ksinformer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &devopsGetter{informers: ksinformer}
}

func (n devopsGetter) Get(_, name string) (runtime.Object, error) {
	return n.informers.Devops().V1alpha3().DevOpsProjects().Lister().Get(name)
}

func (n devopsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	projects, err := n.informers.Devops().V1alpha3().DevOpsProjects().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, project := range projects {
		result = append(result, project)
	}

	return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n devopsGetter) filter(item runtime.Object, filter query.Filter) bool {
	devOpsProject, ok := item.(*devopsv1alpha3.DevOpsProject)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaFilter(devOpsProject.ObjectMeta, filter)
}

func (n devopsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftProject, ok := left.(*devopsv1alpha3.DevOpsProject)
	if !ok {
		return false
	}

	rightProject, ok := right.(*devopsv1alpha3.DevOpsProject)
	if !ok {
		return true
	}
	return v1alpha3.DefaultObjectMetaCompare(leftProject.ObjectMeta, rightProject.ObjectMeta, field)
}
