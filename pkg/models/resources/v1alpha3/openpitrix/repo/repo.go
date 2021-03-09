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

package repo

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"strings"
)

type reposGetter struct {
	ksInformer ksinformers.SharedInformerFactory
}

func New(ksinformer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &reposGetter{ksInformer: ksinformer}
}

func (d *reposGetter) Get(_, name string) (runtime.Object, error) {
	return d.ksInformer.Application().V1alpha1().HelmRepos().Lister().Get(name)
}

func (d *reposGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	var repos []*v1alpha1.HelmRepo
	var err error
	repos, err = d.ksInformer.Application().V1alpha1().HelmRepos().Lister().List(query.Selector())

	if err != nil {
		return nil, err
	}

	result := make([]runtime.Object, 0, len(repos))
	for _, user := range repos {
		result = append(result, user)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *reposGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	repo1, ok := left.(*v1alpha1.HelmRepo)
	if !ok {
		return false
	}

	repo2, ok := right.(*v1alpha1.HelmRepo)
	if !ok {
		return false
	}

	switch field {
	case query.FieldName:
		return strings.Compare(repo1.Spec.Name, repo2.Spec.Name) > 0
	default:
		return v1alpha3.DefaultObjectMetaCompare(repo1.ObjectMeta, repo2.ObjectMeta, field)
	}
}

func (d *reposGetter) filter(object runtime.Object, filter query.Filter) bool {
	repo, ok := object.(*v1alpha1.HelmRepo)

	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldName:
		return strings.Contains(repo.Spec.Name, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(repo.ObjectMeta, filter)
	}
}
