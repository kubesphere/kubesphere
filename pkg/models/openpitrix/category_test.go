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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
	"testing"
)

func TestOpenPitrixCategory(t *testing.T) {
	ctgOperator := prepareCategoryOperator()

	ctgReq := &CreateCategoryRequest{
		Name: "test-ctg",
	}

	// create category
	ctgResp, err := ctgOperator.CreateCategory(ctgReq)
	if err != nil {
		klog.Errorf("create category failed")
		t.Fail()
	}

	// add category to indexer
	ctgs, err := ksClient.ApplicationV1alpha1().HelmCategories().List(context.TODO(), metav1.ListOptions{})
	for _, ctg := range ctgs.Items {
		err := fakeInformerFactory.KubeSphereSharedInformerFactory().Application().V1alpha1().HelmCategories().
			Informer().GetIndexer().Add(&ctg)
		if err != nil {
			klog.Errorf("failed to add category to indexer")
			t.FailNow()
		}
	}

	// describe category
	cond := &params.Conditions{}
	ctgList, err := ctgOperator.ListCategories(cond, "", false, 10, 0)
	if err != nil {
		klog.Errorf("list app failed, err: %s", err)
		t.FailNow()
	}

	if len(ctgList.Items) != 1 {
		klog.Errorf("list app failed")
		t.FailNow()
	}

	// describe category
	ctg, err := ctgOperator.DescribeCategory(ctgResp.CategoryId)
	if err != nil {
		klog.Errorf("describe app failed, err: %s", err)
		t.FailNow()
	}
	_ = ctg

}

func prepareCategoryOperator() CategoryInterface {
	ksClient = fakeks.NewSimpleClientset()
	k8sClient = fakek8s.NewSimpleClientset()
	fakeInformerFactory = informers.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)

	return newCategoryOperator(fakeInformerFactory.KubeSphereSharedInformerFactory(), ksClient)
}
