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

	. "github.com/onsi/ginkgo" //nolint:stylecheck
	. "github.com/onsi/gomega" //nolint:stylecheck
	"k8s.io/apimachinery/pkg/util/wait"
	"kubesphere.io/client-go/client"

	"kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"kubesphere.io/kubesphere/test/e2e/constant"
	"kubesphere.io/kubesphere/test/e2e/framework"
	"kubesphere.io/kubesphere/test/e2e/framework/iam"
	"kubesphere.io/kubesphere/test/e2e/framework/resource"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	GroupName = "test-group"
	UserName  = "tester"
)

var _ = Describe("Groups", func() {
	f := framework.NewKubeSphereFramework("group")

	var workspace = ""
	var namespace = ""
	var group = ""
	var userClient client.Client

	adminClient := f.GenericClient("group")
	restClient := f.RestClient("group")

	Context("Grant Permissions by assign user to group", func() {
		It("Should create group and assign members successfully", func() {
			By("Expecting to create test workspace for Group tests")
			workspace = f.TestWorkSpaceName()
			namespace = f.CreateNamespace("group")
			g, err := iam.CreateGroup(adminClient, iam.NewGroup(GroupName, workspace), workspace)
			framework.ExpectNoError(err)
			group = g.Name
			Eventually(func() bool {
				expGroup, err := iam.GetGroup(adminClient, group, workspace)
				return err == nil && expGroup.Name == group
			}, timeout, interval).Should(BeTrue())

			By("Create user and wait until active")

			u, err := createUserWithWait(f, adminClient, UserName)
			framework.ExpectNoError(err)

			By("Assign user to Group")
			_, err = restClient.IamV1alpha2().Groups().CreateBinding(context.TODO(), workspace, group, UserName)
			framework.ExpectNoError(err)

			Eventually(func() bool {
				user, err := iam.GetUser(adminClient, UserName)
				return err == nil && stringutils.FindString(user.Spec.Groups, group) != -1
			}, timeout, interval).Should(BeTrue())

			By("Creating a new client with user authentication")
			userClient, err = iam.NewClient(f.GetScheme(), u.Name, constant.DefaultPassword)
			framework.ExpectNoError(err)
		})

		It(fmt.Sprintf("%s has no permissions to access namespace: %s", UserName, namespace), func() {
			err := CheckNamespaceAccess(f, userClient, namespace)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

		})

		It(fmt.Sprintf("%s should has full access namespace: %s", UserName, namespace), func() {

			rolename := fmt.Sprintf("%s-regular", workspace)
			By("Grant namespace permission by bind admin role to group")
			_, err := restClient.IamV1alpha2().RoleBindings().CreateRoleBinding(context.TODO(), namespace, "admin", group)
			framework.ExpectNoError(err)
			_, err = restClient.IamV1alpha2().RoleBindings().CreateWorkspaceRoleBinding(context.TODO(), workspace, rolename, group)
			framework.ExpectNoError(err)

			err = CheckNamespaceAccess(f, userClient, namespace)
			framework.ExpectNoError(err)
		})

	})
})

// Todo: The can-i API should be a better option, but ks-apiserver doesn't support it yet.
// So we will try to list objects in the namespace.
func CheckNamespaceAccess(f framework.KubeSphereFramework, c client.Client, namespace string) error {
	_, err := resource.ListPods(c, namespace)
	return err
}

// Create a user and wait until the user became active status.
func createUserWithWait(f framework.KubeSphereFramework, c client.Client, username string) (*v1alpha2.User, error) {
	u := iam.NewUser(username, "platform-regular")
	if _, err := iam.CreateUser(c, u); err != nil {
		return nil, err
	}

	err := wait.PollImmediate(framework.PollInterval, framework.DefaultSingleCallTimeout, func() (bool, error) {
		u, err := iam.GetUser(c, username)
		if err != nil {
			framework.Failf("Cannot retrieve User %q: %v", username, err)
			return false, err
		}
		if u == nil || u.Status.State == "" {
			return false, nil
		}
		return u.Status.State == v1alpha2.UserActive, nil
	})

	if err != nil {
		return nil, err
	}
	return u, nil
}
