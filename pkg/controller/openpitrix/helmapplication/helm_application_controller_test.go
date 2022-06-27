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

package helmapplication

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"kubesphere.io/api/application/v1alpha1"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
)

var _ = Describe("helmApplication", func() {

	const timeout = time.Second * 240
	const interval = time.Second * 1

	app := createApp()
	appVer := createAppVersion(app.GetHelmApplicationId(), "0.0.1")
	appVer2 := createAppVersion(app.GetHelmApplicationId(), "0.0.2")

	BeforeEach(func() {
		err := k8sClient.Create(context.Background(), app)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(context.Background(), appVer)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(context.Background(), appVer2)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Helm Application Controller", func() {
		It("Should success", func() {

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
				var localApp v1alpha1.HelmApplication
				appKey := types.NamespacedName{
					Name: app.Name,
				}
				k8sClient.Get(context.Background(), appKey, &localApp)
				return localApp.State() == v1alpha1.StateActive
			}, timeout, interval).Should(BeTrue())

			By("Mark workspace is deleted")
			Eventually(func() bool {
				var localApp v1alpha1.HelmApplication
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: app.Name}, &localApp)
				if err != nil {
					return false
				}
				appCopy := localApp.DeepCopy()
				appCopy.Annotations = map[string]string{}
				appCopy.Annotations[constants.DanglingAppCleanupKey] = constants.CleanupDanglingAppOngoing
				patchData := client.MergeFrom(&localApp)
				err = k8sClient.Patch(context.Background(), appCopy, patchData)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Draft app version are deleted")
			Eventually(func() bool {
				var ver v1alpha1.HelmApplicationVersion
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: appVer2.Name}, &ver)
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Active app version exists")
			Eventually(func() bool {
				var ver v1alpha1.HelmApplicationVersion
				err := k8sClient.Get(context.Background(), types.NamespacedName{Name: appVer.Name}, &ver)
				return err == nil
			}, timeout, interval).Should(BeTrue())

		})
	})
})

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

func createAppVersion(appId string, version string) *v1alpha1.HelmApplicationVersion {
	return &v1alpha1.HelmApplicationVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: idutils.GetUuid36(v1alpha1.HelmApplicationVersionIdPrefix),
			Labels: map[string]string{
				constants.ChartApplicationIdLabelKey: appId,
			},
		},
		Spec: v1alpha1.HelmApplicationVersionSpec{
			Metadata: &v1alpha1.Metadata{
				Version: version,
				Name:    "dummy-chart",
			},
		},
	}
}
