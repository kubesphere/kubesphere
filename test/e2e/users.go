package e2e

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"kubesphere.io/client-go/rest"

	"kubesphere.io/kubesphere/test/e2e/framework"
)

var _ = Describe("User", func() {

	f := framework.NewKubeSphereFramework()

	userName := "test"
	var client *rest.RESTClient

	BeforeEach(func() {
		client = f.RestClient()
	})

	It("Create user", func() {
		By(fmt.Sprintf("Expecting to user %s does not exists", userName))

		user := &iamv1beta1.User{}
		err := client.Get().
			Prefix("/apis").
			Group("iam.kubesphere.io").
			Version("v1alpha2").
			Resource("users").
			Name(userName).
			Do(context.Background()).
			Into(user)

		Expect(err).ShouldNot(BeNil())

		user = &iamv1beta1.User{
			ObjectMeta: v1.ObjectMeta{Name: userName},
		}
		err = client.Post().
			Prefix("/apis").
			Group("iam.kubesphere.io").
			Version("v1alpha2").
			Resource("users").
			Body(user).
			Do(context.Background()).
			Into(user)

		Expect(err).Should(BeNil())
		Expect(user.ResourceVersion).ShouldNot(BeEmpty())

		framework.Logf("User created successfully")
	})
})
