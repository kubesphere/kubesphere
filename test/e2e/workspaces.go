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
	"time"

	. "github.com/onsi/ginkgo" //nolint:stylecheck
	. "github.com/onsi/gomega"
	"kubesphere.io/client-go/client"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/test/e2e/framework"
	"kubesphere.io/kubesphere/test/e2e/framework/iam"
	"kubesphere.io/kubesphere/test/e2e/framework/workspace"
)

const timeout = time.Second * 30
const interval = time.Second * 1

var _ = Describe("Worksspace", func() {

	f := framework.NewKubeSphereFramework("worksspace")

	var wsName = ""
	var gclient client.Client

	BeforeEach(func() {
		gclient = f.GenericClient("worksspace")
	})

	It("Should create workspace and initial workspace settings", func() {

		By("Expecting to create workspace thronght workspace template")
		wsName = f.TestWorkSpaceName()

		By("Expecting to create workspace successfully")
		Eventually(func() bool {
			wsp, err := workspace.GetWorkspace(gclient, wsName)
			return err == nil && wsp.Name == wsName
		}, timeout, interval).Should(BeTrue())

		By("Expecting initial workspace roles successfully")

		wspr := &iamv1alpha2.WorkspaceRoleList{}
		Eventually(func() bool {
			opts := iam.URLOptions

			err := gclient.List(context.TODO(), wspr, &opts)
			return err == nil && len(wspr.Items) > 0
		}, timeout, interval).Should(BeTrue())

		_, err := workspace.DeleteWorkspace(gclient, wsName)
		framework.ExpectNoError(err)

		framework.Logf("Workspace created successfully")
	})
})
