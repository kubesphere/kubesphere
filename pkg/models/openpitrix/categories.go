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

package openpitrix

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	typed_v1alpha1 "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	listers_v1alpha1 "kubesphere.io/kubesphere/pkg/client/listers/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

type CategoryInterface interface {
	CreateCategory(request *CreateCategoryRequest) (*CreateCategoryResponse, error)
	DeleteCategory(id string) error
	ModifyCategory(id string, request *ModifyCategoryRequest) error
	ListCategories(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	DescribeCategory(id string) (*Category, error)
}

type categoryOperator struct {
	ctgClient typed_v1alpha1.ApplicationV1alpha1Interface
	ctgLister listers_v1alpha1.HelmCategoryLister
}

func newCategoryOperator(ksFactory externalversions.SharedInformerFactory, ksClient versioned.Interface) CategoryInterface {
	c := &categoryOperator{
		ctgClient: ksClient.ApplicationV1alpha1(),
		ctgLister: ksFactory.Application().V1alpha1().HelmCategories().Lister(),
	}

	return c
}

func (c *categoryOperator) getCategoryByName(name string) (*v1alpha1.HelmCategory, error) {
	ctgs, err := c.ctgLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, ctg := range ctgs {
		if name == ctg.Spec.Name {
			return ctg, nil
		}
	}
	return nil, nil
}

func (c *categoryOperator) createCategory(name, desc string) (*v1alpha1.HelmCategory, error) {
	ctg := &v1alpha1.HelmCategory{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(v1alpha1.HelmCategoryIdPrefix),
		},
		Spec: v1alpha1.HelmCategorySpec{
			Description: desc,
			Name:        name,
		},
	}

	return c.ctgClient.HelmCategories().Create(context.TODO(), ctg, metav1.CreateOptions{})
}

func (c *categoryOperator) CreateCategory(request *CreateCategoryRequest) (*CreateCategoryResponse, error) {

	ctg, err := c.getCategoryByName(request.Name)
	if err != nil {
		return nil, err
	}

	if ctg != nil {
		return nil, errors.New("category %s exists", ctg.Spec.Name)
	}

	ctg, err = c.createCategory(request.Name, request.Description)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &CreateCategoryResponse{
		CategoryId: ctg.Name,
	}, nil
}

func (c *categoryOperator) DeleteCategory(id string) error {
	ctg, err := c.ctgClient.HelmCategories().Get(context.TODO(), id, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if ctg.Status.Total > 0 {
		return errors.New("category %s owns application", ctg.Spec.Name)
	}

	err = c.ctgClient.HelmCategories().Delete(context.TODO(), id, metav1.DeleteOptions{})

	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *categoryOperator) ModifyCategory(id string, request *ModifyCategoryRequest) error {

	ctg, err := c.ctgClient.HelmCategories().Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		return errors.New("category %s not found", id)
	}
	ctgCopy := ctg.DeepCopy()

	if request.Name != nil {
		ctgCopy.Spec.Name = *request.Name
	}

	if request.Description != nil {
		ctgCopy.Spec.Description = *request.Description
	}

	patch := client.MergeFrom(ctg)
	data, err := patch.Data(ctgCopy)
	if err != nil {
		klog.Error("create patch failed", err)
		return err
	}

	_, err = c.ctgClient.HelmCategories().Patch(context.TODO(), id, patch.Type(), data, metav1.PatchOptions{})

	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *categoryOperator) DescribeCategory(id string) (*Category, error) {
	var err error
	ctg := &v1alpha1.HelmCategory{}
	ctg, err = c.ctgClient.HelmCategories().Get(context.TODO(), id, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return convertCategory(ctg), nil
}

func (c *categoryOperator) ListCategories(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	ctgs, err := c.ctgLister.List(labels.Everything())

	if err != nil {
		return nil, err
	}

	sort.Sort(HelmCategoryList(ctgs))

	items := make([]interface{}, 0, limit)
	for i, j := offset, 0; i < len(ctgs) && j < limit; i, j = i+1, j+1 {
		items = append(items, convertCategory(ctgs[i]))
	}

	return &models.PageableResponse{Items: items, TotalCount: len(ctgs)}, nil
}
