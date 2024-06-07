/*
Copyright 2020 KubeSphere Authors

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
