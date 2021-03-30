package e2e

import (
	"context"

	. "github.com/onsi/ginkgo" //nolint:stylecheck
	. "github.com/onsi/gomega" //nolint:stylecheck
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"kubesphere.io/client-go/client"

	"kubesphere.io/kubesphere/test/e2e/framework"
	"kubesphere.io/kubesphere/test/e2e/framework/workspace"
)

var _ = Describe("API Test", func() {

	f := framework.NewKubeSphereFramework("worksspace")
	var gclient client.Client

	BeforeEach(func() {
		gclient = f.GenericClient("worksspace")
	})

	It("Should list Kubernetes objects through ks proxy", func() {
		results := &corev1.NamespaceList{}

		Expect(gclient.List(context.TODO(), results)).Should(Succeed())

		if len(results.Items) < 1 {
			framework.Failf("Test Failed caused no Cluster role was found")
		}
		framework.Logf("Response is %v", results)
	})

	It("Should get Kubernetes objects through ks proxy", func() {
		deploy := &appsv1.Deployment{}

		Expect(gclient.Get(context.TODO(), client.ObjectKey{Namespace: "kubesphere-system", Name: "ks-apiserver"}, deploy)).Should(Succeed())
		Expect(deploy.Name).To(Equal("ks-apiserver"))
	})

	It("Should list Kubernetes objects through ksapi", func() {
		results := &corev1.NamespaceList{}

		opts := workspace.URLOptions

		err := gclient.List(context.TODO(), results, &opts)
		framework.ExpectNoError(err)
		framework.Logf("Response is %v", results)
	})

	It("Should list Kubernetes objects through ksapi by workspaces", func() {
		results := &corev1.NamespaceList{}

		opts := workspace.URLOptions

		err := gclient.List(context.TODO(), results, &opts, &client.WorkspaceOptions{Name: "kubesphere-system"})
		framework.ExpectNoError(err)
		framework.Logf("Response is %v", results)
	})
})
