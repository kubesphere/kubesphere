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

package helmcategory

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"time"
)

var _ = Describe("helmCategory", func() {

	const timeout = time.Second * 240
	const interval = time.Second * 1

	app := createApp()
	appVer := createAppVersion(app.GetHelmApplicationId())
	ctg := createCtg()

	BeforeEach(func() {
		err := k8sClient.Create(context.Background(), app)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(context.Background(), appVer)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(context.Background(), ctg)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Helm category Controller", func() {
		It("Should success", func() {
			key := types.NamespacedName{
				Name: v1alpha1.UncategorizedId,
			}

			By("Expecting category should exists")
			Eventually(func() bool {
				f := &v1alpha1.HelmCategory{}
				k8sClient.Get(context.Background(), key, f)
				return !f.CreationTimestamp.IsZero()
			}, timeout, interval).Should(BeTrue())

			By("Update helm app version status")
			Eventually(func() bool {
				k8sClient.Get(context.Background(), types.NamespacedName{Name: appVer.Name}, appVer)
				appVer.Status = v1alpha1.HelmApplicationVersionStatus{
					State: v1alpha1.StateActive,
				}
				err := k8sClient.Status().Update(context.Background(), appVer)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Wait for app status become active")
			Eventually(func() bool {
				appKey := types.NamespacedName{
					Name: app.Name,
				}
				k8sClient.Get(context.Background(), appKey, app)
				return app.State() == v1alpha1.StateActive
			}, timeout, interval).Should(BeTrue())

			By("Reconcile for `uncategorized` category")
			Eventually(func() bool {
				key := types.NamespacedName{Name: v1alpha1.UncategorizedId}
				ctg := v1alpha1.HelmCategory{}
				k8sClient.Get(context.Background(), key, &ctg)

				return ctg.Status.Total == 1
			}, timeout, interval).Should(BeTrue())
		})
	})
})

func createCtg() *v1alpha1.HelmCategory {
	return &v1alpha1.HelmCategory{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(v1alpha1.HelmCategoryIdPrefix),
		},
		Spec: v1alpha1.HelmCategorySpec{
			Name: "dummy-ctg",
		},
	}
}

func createApp() *v1alpha1.HelmApplication {
	return &v1alpha1.HelmApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(v1alpha1.HelmApplicationIdPrefix),
		},
		Spec: v1alpha1.HelmApplicationSpec{
			Name: "dummy-chart",
		},
	}
}

func createAppVersion(appId string) *v1alpha1.HelmApplicationVersion {
	return &v1alpha1.HelmApplicationVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(v1alpha1.HelmApplicationVersionIdPrefix),
			Labels: map[string]string{
				constants.ChartApplicationIdLabelKey: appId,
			},
		},
		Spec: v1alpha1.HelmApplicationVersionSpec{
			Metadata: &v1alpha1.Metadata{
				Version: "0.0.1",
				Name:    "dummy-chart",
			},
		},
	}
}
