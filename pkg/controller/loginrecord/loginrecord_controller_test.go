/*
Copyright 2020 The KubeSphere Authors.

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

package loginrecord

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clienttesting "k8s.io/client-go/testing"
	"kubesphere.io/kubesphere/pkg/apis"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"testing"
	"time"
)

func TestLoginRecordController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t,
		"LoginRecord Controller Test Suite",
		[]Reporter{printer.NewlineReporter{}})
}

func newLoginRecord(username string) *iamv1alpha2.LoginRecord {
	return &iamv1alpha2.LoginRecord{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d", username, rand.Intn(1000000)),
			Labels: map[string]string{
				iamv1alpha2.UserReferenceLabel: username,
			},
			CreationTimestamp: metav1.Now(),
		},
		Spec: iamv1alpha2.LoginRecordSpec{
			Type:      iamv1alpha2.Token,
			Provider:  "",
			Success:   true,
			Reason:    iamv1alpha2.AuthenticatedSuccessfully,
			SourceIP:  "",
			UserAgent: "",
		},
	}
}

func newUser(username string) *iamv1alpha2.User {
	return &iamv1alpha2.User{
		ObjectMeta: metav1.ObjectMeta{Name: username},
	}
}

var _ = Describe("LoginRecord", func() {
	var k8sClient *fakek8s.Clientset
	var ksClient *fakeks.Clientset
	var user *iamv1alpha2.User
	var loginRecord *iamv1alpha2.LoginRecord
	var controller *loginRecordController
	var informers externalversions.SharedInformerFactory
	BeforeEach(func() {
		user = newUser("admin")
		loginRecord = newLoginRecord(user.Name)
		k8sClient = fakek8s.NewSimpleClientset()
		ksClient = fakeks.NewSimpleClientset(loginRecord, user)
		informers = externalversions.NewSharedInformerFactory(ksClient, 0)
		loginRecordInformer := informers.Iam().V1alpha2().LoginRecords()
		userInformer := informers.Iam().V1alpha2().Users()
		err := loginRecordInformer.Informer().GetIndexer().Add(loginRecord)
		Expect(err).Should(BeNil())
		err = userInformer.Informer().GetIndexer().Add(user)
		Expect(err).Should(BeNil())
		err = apis.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())
		controller = NewLoginRecordController(k8sClient, ksClient, loginRecordInformer, userInformer, time.Hour, 1)
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("LoginRecord Controller", func() {
		It("Should create successfully", func() {

			By("Expecting to reconcile successfully")
			err := controller.reconcile(loginRecord.Name)
			Expect(err).Should(BeNil())

			By("Expecting to update user last login time successfully")
			err = controller.reconcile(loginRecord.Name)
			Expect(err).Should(BeNil())
			actions := ksClient.Actions()
			Expect(len(actions)).Should(Equal(1))
			newObject := user.DeepCopy()
			newObject.Status.LastLoginTime = &loginRecord.CreationTimestamp
			updateAction := clienttesting.NewUpdateAction(iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralUser), "", newObject)
			updateAction.Subresource = "status"
			Expect(actions[0]).Should(Equal(updateAction))
		})
	})
})
