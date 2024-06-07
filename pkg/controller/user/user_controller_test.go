/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	runtimefakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/scheme"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

func newUser(name string) *iamv1beta1.User {
	return &iamv1beta1.User{
		TypeMeta: metav1.TypeMeta{APIVersion: iamv1beta1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: iamv1beta1.UserSpec{
			Email:             fmt.Sprintf("%s@kubesphere.io", name),
			Lang:              "zh-CN",
			Description:       "fake user",
			EncryptedPassword: "test",
		},
	}
}

func TestDoNothing(t *testing.T) {
	authenticateOptions := authentication.NewOptions()
	authenticateOptions.AuthenticateRateLimiterMaxTries = 1
	authenticateOptions.AuthenticateRateLimiterDuration = 2 * time.Second
	user := newUser("test")
	loginRecords := make([]runtime.Object, 0)
	for i := 0; i < authenticateOptions.AuthenticateRateLimiterMaxTries+1; i++ {
		loginRecord := iamv1beta1.LoginRecord{
			ObjectMeta: metav1.ObjectMeta{
				Name:   fmt.Sprintf("%s-%d", user.Name, i),
				Labels: map[string]string{iamv1beta1.UserReferenceLabel: user.Name},
				// Ensure that the failed login record created after the user status change to active,
				// otherwise, the failed login attempts will not be counted.
				CreationTimestamp: metav1.NewTime(time.Now().Add(time.Minute)),
			},
			Spec: iamv1beta1.LoginRecordSpec{
				Success: false,
			},
		}
		loginRecords = append(loginRecords, &loginRecord)
	}

	client := runtimefakeclient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(user).WithRuntimeObjects(loginRecords...).Build()
	clusterClientSet, err := clusterclient.NewClusterClientSet(&informertest.FakeInformers{Scheme: scheme.Scheme})
	if err != nil {
		t.Fatal(err)
	}
	c := &Reconciler{
		recorder:              &record.FakeRecorder{},
		logger:                ctrl.Log.WithName("controllers").WithName(controllerName),
		Client:                client,
		authenticationOptions: authenticateOptions,
		clusterClient:         clusterClientSet,
	}

	users := &iamv1beta1.UserList{}
	w, err := client.Watch(context.Background(), users, &runtimeclient.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	result, err := c.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: user.Name},
	})
	if err != nil {
		t.Fatal(err)
	}

	// append finalizer
	updateEvent := <-w.ResultChan()
	assert.Equal(t, watch.Modified, updateEvent.Type)
	assert.NotNil(t, updateEvent.Object)
	user = updateEvent.Object.(*iamv1beta1.User)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.Finalizers)

	result, err = c.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: user.Name},
	})
	if err != nil {
		t.Fatal(err)
	}
	updateEvent = <-w.ResultChan()
	// encrypt password
	assert.Equal(t, watch.Modified, updateEvent.Type)
	assert.NotNil(t, updateEvent.Object)
	user = updateEvent.Object.(*iamv1beta1.User)
	assert.NotNil(t, user)
	assert.True(t, isEncrypted(user.Spec.EncryptedPassword))

	// becomes active after password encrypted
	updateEvent = <-w.ResultChan()
	user = updateEvent.Object.(*iamv1beta1.User)
	assert.Equal(t, iamv1beta1.UserActive, user.Status.State)

	// block user
	updateEvent = <-w.ResultChan()
	user = updateEvent.Object.(*iamv1beta1.User)
	assert.Equal(t, iamv1beta1.UserAuthLimitExceeded, user.Status.State)
	assert.True(t, result.Requeue)

	time.Sleep(result.RequeueAfter + time.Second)
	_, err = c.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: user.Name},
	})
	if err != nil {
		t.Fatal(err)
	}

	// unblock user
	updateEvent = <-w.ResultChan()
	user = updateEvent.Object.(*iamv1beta1.User)
	assert.Equal(t, iamv1beta1.UserActive, user.Status.State)
}
