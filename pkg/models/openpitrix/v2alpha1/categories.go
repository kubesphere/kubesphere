/*
Copyright 2020 The KubeSphere Authors.
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

package v2alpha1

import (
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/openpitrix/category"
)

type CategoryInterface interface {
	ListCategories(q *query.Query) (*api.ListResult, error)
	DescribeCategory(id string) (*v1alpha1.HelmCategory, error)
}

type categoryOperator struct {
	ctgGetter resources.Interface
}

func newCategoryOperator(ksFactory externalversions.SharedInformerFactory) CategoryInterface {
	c := &categoryOperator{
		ctgGetter: category.New(ksFactory),
	}

	return c
}

func (c *categoryOperator) DescribeCategory(id string) (*v1alpha1.HelmCategory, error) {
	ret, err := c.ctgGetter.Get("", id)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	ctg := ret.(*v1alpha1.HelmCategory)
	return ctg, nil
}

func (c *categoryOperator) ListCategories(q *query.Query) (*api.ListResult, error) {

	result, err := c.ctgGetter.List("", q)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return result, nil
}
