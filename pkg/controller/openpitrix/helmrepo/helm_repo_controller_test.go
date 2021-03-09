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

package helmrepo

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"time"
)

var repoUrl = "https://charts.kubesphere.io/main"

var _ = Describe("helmRepo", func() {

	const timeout = time.Second * 360
	const interval = time.Second * 1

	repo := createRepo()

	BeforeEach(func() {
		err := k8sClient.Create(context.Background(), repo)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Helm Repo Controller", func() {
		It("Should success", func() {
			key := types.NamespacedName{
				Name: repo.Name,
			}

			By("Expecting repo state is successful")
			Eventually(func() bool {
				repo := &v1alpha1.HelmRepo{}
				k8sClient.Get(context.Background(), key, repo)
				return repo.Status.State == v1alpha1.RepoStateSuccessful && len(repo.Status.Data) > 0
			}, timeout, interval).Should(BeTrue())
		})
	})
})

func createRepo() *v1alpha1.HelmRepo {
	return &v1alpha1.HelmRepo{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(v1alpha1.HelmRepoIdPrefix),
		},
		Spec: v1alpha1.HelmRepoSpec{
			Url: repoUrl,
		},
	}
}
