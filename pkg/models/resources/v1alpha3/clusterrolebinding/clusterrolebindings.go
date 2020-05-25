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

package clusterrolebinding

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type clusterrolebindingsGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &clusterrolebindingsGetter{sharedInformers: sharedInformers}
}

func (d *clusterrolebindingsGetter) Get(_, name string) (runtime.Object, error) {
	return d.sharedInformers.Rbac().V1().ClusterRoleBindings().Lister().Get(name)
}

func (d *clusterrolebindingsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	roleBindings, err := d.sharedInformers.Rbac().V1().ClusterRoleBindings().Lister().List(query.Selector())

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, roleBinding := range roleBindings {
		result = append(result, roleBinding)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *clusterrolebindingsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftRoleBinding, ok := left.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return false
	}

	rightRoleBinding, ok := right.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftRoleBinding.ObjectMeta, rightRoleBinding.ObjectMeta, field)
}

func (d *clusterrolebindingsGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*rbacv1.ClusterRoleBinding)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}
