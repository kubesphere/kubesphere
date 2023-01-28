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

package category

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

type categoryGetter struct {
	c client.Reader
}

func New(c client.Reader) v1alpha3.Interface {
	return &categoryGetter{c: c}
}

func (c *categoryGetter) Get(_, name string) (runtime.Object, error) {
	category := &iamv1beta1.Category{}

	err := c.c.Get(context.Background(), client.ObjectKey{Name: name}, category)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (c *categoryGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	var (
		categoryList iamv1beta1.CategoryList
		result       []runtime.Object
	)
	labelsMap, err := labels.ConvertSelectorToLabelsMap(query.LabelSelector)
	if err != nil {
		return nil, err
	}
	err = c.c.List(context.Background(), &categoryList, client.MatchingLabels(labelsMap))
	if err != nil {
		return nil, err
	}

	for _, v := range categoryList.Items {
		result = append(result, &v)
	}
	return v1alpha3.DefaultList(result, query, c.compare, c.filter), nil
}

func (c *categoryGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftCategory, ok := left.(*iamv1beta1.Category)
	if !ok {
		return false
	}
	rightCategory, ok := right.(*iamv1beta1.Category)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftCategory.ObjectMeta, rightCategory.ObjectMeta, field)
}

func (c *categoryGetter) filter(object runtime.Object, filter query.Filter) bool {
	category, ok := object.(*iamv1beta1.Category)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(category.ObjectMeta, filter)
}
