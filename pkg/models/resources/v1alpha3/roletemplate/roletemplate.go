/*
 * Copyright 2022 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package roletemplate

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type roleTemplateGetter struct {
	c client.Reader
}

func New(c client.Reader) v1alpha3.Interface {
	return &roleTemplateGetter{
		c: c,
	}
}

func (r *roleTemplateGetter) Get(_, name string) (runtime.Object, error) {
	roleTemplate := &iamv1beta1.RoleTemplate{}
	err := r.c.Get(context.Background(), client.ObjectKey{Name: name}, roleTemplate)
	if err != nil {
		return nil, err
	}
	return roleTemplate, nil
}

func (r *roleTemplateGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	var roleTemplateList iamv1beta1.RoleTemplateList
	var result []runtime.Object
	labelsMap, err := labels.ConvertSelectorToLabelsMap(query.LabelSelector)
	if err != nil {
		return nil, err
	}
	err = r.c.List(context.Background(), &roleTemplateList, client.MatchingLabels(labelsMap))
	if err != nil {
		return nil, err
	}
	for _, v := range roleTemplateList.Items {
		result = append(result, &v)
	}
	return v1alpha3.DefaultList(result, query, r.compare, r.filter), nil
}

func (r *roleTemplateGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftRoleTemplate, ok := left.(*iamv1beta1.RoleTemplate)
	if !ok {
		return false
	}

	rightRoleTemplate, ok := right.(*iamv1beta1.RoleTemplate)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftRoleTemplate.ObjectMeta, rightRoleTemplate.ObjectMeta, field)
}

func (r *roleTemplateGetter) filter(object runtime.Object, filter query.Filter) bool {
	roleTemplate, ok := object.(*iamv1beta1.RoleTemplate)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(roleTemplate.ObjectMeta, filter)
}
