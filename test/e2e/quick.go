package e2e

import (
	"context"

	"kubesphere.io/client-go/rest"

	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck
	. "github.com/onsi/gomega"    //nolint:stylecheck

	"kubesphere.io/kubesphere/test/e2e/framework"
)

var _ = Describe("API Test", func() {

	f := framework.NewKubeSphereFramework()
	var client *rest.RESTClient

	BeforeEach(func() {
		client = f.RestClient()
	})

	It("Should retrieve KubeSphere API version", func() {
		result := client.Get().AbsPath("/version").Do(context.Background())
		Expect(result.Error()).Should(BeNil())

		data, err := result.Raw()
		Expect(err).Should(BeNil())

		framework.Logf("Response is %v", string(data))
	})
})
