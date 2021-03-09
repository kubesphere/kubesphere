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
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"testing"
)

func TestOpenPitrixRepo(t *testing.T) {
	repoOperator := prepareRepoOperator()

	repo := v1alpha1.HelmRepo{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(v1alpha1.HelmRepoIdPrefix),

			Labels: map[string]string{
				constants.WorkspaceLabelKey: testWorkspace,
			},
		},
		Spec: v1alpha1.HelmRepoSpec{
			Name:       "test-repo",
			Url:        "https://charts.kubesphere.io/main",
			SyncPeriod: 0,
		},
	}

	// validate repo
	validateRes, err := repoOperator.ValidateRepo(repo.Spec.Url, &repo.Spec.Credential)
	if err != nil || validateRes.Ok == false {
		klog.Errorf("validate category failed, error: %s", err)
		t.Fail()
	}

	// validate the corrupt repo
	validateRes, err = repoOperator.ValidateRepo("http://www.baidu.com", &repo.Spec.Credential)
	if err == nil {
		klog.Errorf("validate category failed")
		t.Fail()
	}

	// create repo
	repoResp, err := repoOperator.CreateRepo(&repo)
	if err != nil {
		klog.Errorf("create category failed")
		t.Fail()
	}

	// add category to indexer
	repos, err := ksClient.ApplicationV1alpha1().HelmRepos().List(context.TODO(), metav1.ListOptions{})
	for _, repo := range repos.Items {
		err := fakeInformerFactory.KubeSphereSharedInformerFactory().Application().V1alpha1().HelmRepos().
			Informer().GetIndexer().Add(&repo)
		if err != nil {
			klog.Errorf("failed to add repo to indexer")
			t.FailNow()
		}
	}

	// list repo
	cond := &params.Conditions{Match: map[string]string{WorkspaceLabel: testWorkspace}}
	repoList, err := repoOperator.ListRepos(cond, "", false, 10, 0)
	if err != nil {
		klog.Errorf("list repo failed, err: %s", err)
		t.FailNow()
	}

	if len(repoList.Items) != 1 {
		klog.Errorf("list repo failed")
		t.FailNow()
	}

	// describe repo
	describeRepo, err := repoOperator.DescribeRepo(repoResp.RepoID)
	if err != nil {
		klog.Errorf("describe app failed, err: %s", err)
		t.FailNow()
	}
	_ = describeRepo

}

func prepareRepoOperator() RepoInterface {
	ksClient = fakeks.NewSimpleClientset()
	k8sClient = fakek8s.NewSimpleClientset()
	fakeInformerFactory = informers.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)

	return newRepoOperator(cachedReposData, fakeInformerFactory.KubeSphereSharedInformerFactory(), ksClient)
}
