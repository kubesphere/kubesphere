/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package loginrecord

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/scheme"
)

func TestLoginRecordController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LoginRecord Controller Test Suite")
}

func newLoginRecord(username string) *iamv1beta1.LoginRecord {
	return &iamv1beta1.LoginRecord{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d", username, rand.Intn(1000000)),
			Labels: map[string]string{
				iamv1beta1.UserReferenceLabel: username,
			},
			CreationTimestamp: metav1.Now(),
		},
		Spec: iamv1beta1.LoginRecordSpec{
			Type:      iamv1beta1.Token,
			Provider:  "",
			Success:   true,
			Reason:    iamv1beta1.AuthenticatedSuccessfully,
			SourceIP:  "",
			UserAgent: "",
		},
	}
}

func newUser(username string) *iamv1beta1.User {
	return &iamv1beta1.User{
		ObjectMeta: metav1.ObjectMeta{Name: username},
	}
}

var _ = Describe("LoginRecord", func() {
	var user *iamv1beta1.User
	var loginRecord *iamv1beta1.LoginRecord
	var reconciler *Reconciler
	BeforeEach(func() {
		user = newUser("admin")
		loginRecord = newLoginRecord(user.Name)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(user, loginRecord).Build()
		reconciler = &Reconciler{}
		reconciler.Client = fakeClient
		reconciler.recorder = record.NewFakeRecorder(2)
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("LoginRecord Controller", func() {
		It("Should create successfully", func() {
			By("Expecting to reconcile successfully")
			_, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{
				Name: loginRecord.Name,
			}})
			Expect(err).Should(BeNil())
		})
	})
})
