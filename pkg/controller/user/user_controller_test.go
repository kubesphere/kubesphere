/*
Copyright 2019 The KubeSphere Authors.

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

package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	runtimefakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/apis"

	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newUser(name string) *iamv1alpha2.User {
	return &iamv1alpha2.User{
		TypeMeta: metav1.TypeMeta{APIVersion: iamv1alpha2.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: iamv1alpha2.UserSpec{
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
		loginRecord := iamv1alpha2.LoginRecord{
			ObjectMeta: metav1.ObjectMeta{
				Name:   fmt.Sprintf("%s-%d", user.Name, i),
				Labels: map[string]string{iamv1alpha2.UserReferenceLabel: user.Name},
				// Ensure that the failed login record created after the user status change to active,
				// otherwise, the failed login attempts will not be counted.
				CreationTimestamp: metav1.NewTime(time.Now().Add(time.Minute)),
			},
			Spec: iamv1alpha2.LoginRecordSpec{
				Success: false,
			},
		}
		loginRecords = append(loginRecords, &loginRecord)
	}
	sch := scheme.Scheme
	if err := apis.AddToScheme(sch); err != nil {
		t.Fatalf("unable add APIs to scheme: %v", err)
	}

	client := runtimefakeclient.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(user).WithRuntimeObjects(loginRecords...).Build()
	ldap := ldapclient.NewSimpleLdap()
	c := &Reconciler{
		Recorder:              &record.FakeRecorder{},
		LdapClient:            ldap,
		Logger:                ctrl.Log.WithName("controllers").WithName(controllerName),
		Client:                client,
		AuthenticationOptions: authenticateOptions,
	}

	users := &iamv1alpha2.UserList{}
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
	user = updateEvent.Object.(*iamv1alpha2.User)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.Finalizers)

	updateEvent = <-w.ResultChan()
	// encrypt password
	assert.Equal(t, watch.Modified, updateEvent.Type)
	assert.NotNil(t, updateEvent.Object)
	user = updateEvent.Object.(*iamv1alpha2.User)
	assert.NotNil(t, user)
	assert.True(t, isEncrypted(user.Spec.EncryptedPassword))

	// becomes active after password encrypted
	updateEvent = <-w.ResultChan()
	user = updateEvent.Object.(*iamv1alpha2.User)
	assert.Equal(t, iamv1alpha2.UserActive, user.Status.State)

	// block user
	updateEvent = <-w.ResultChan()
	user = updateEvent.Object.(*iamv1alpha2.User)
	assert.Equal(t, iamv1alpha2.UserAuthLimitExceeded, user.Status.State)
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
	user = updateEvent.Object.(*iamv1alpha2.User)
	assert.Equal(t, iamv1alpha2.UserActive, user.Status.State)
}
