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
	. "github.com/onsi/ginkgo" //nolint:stylecheck
	. "github.com/onsi/gomega" //nolint:stylecheck
	"kubesphere.io/kubesphere/test/e2e/framework"
	"kubesphere.io/kubesphere/test/e2e/framework/client"
	"kubesphere.io/kubesphere/test/e2e/framework/iam"
)

var _ = Describe("Worksspace", func() {
	f := framework.NewKubeSphereFramework("group")

	var wsName = ""
	var gclient client.Client

	BeforeEach(func() {
		gclient = f.GenericClient("group")
	})

	It("Should create group successfully", func() {

		By("Expecting to create workspace thronght workspace template")
		wsName = f.TestWorkSpaceName()

		group, err := iam.CreateGroup(gclient, iam.NewGroup("group1", wsName), wsName)
		framework.ExpectNoError(err)
		Eventually(func() bool {
			expGroup, err := iam.GetGroup(gclient, group.Name, wsName)
			return err == nil && expGroup.Name == group.Name
		}, timeout, interval).Should(BeTrue())
	})

})
