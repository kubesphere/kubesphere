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
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	ksclient  *fake.Clientset
	k8sclient *k8sfake.Clientset
	// Objects to put in the store.
	userLister []*iamv1alpha2.User
	// Actions expected to happen on the client.
	kubeactions []core.Action
	actions     []core.Action
	// Objects from here preloaded into NewSimpleFake.
	kubeobjects []runtime.Object
	objects     []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	return f
}

func newUser(name string) *iamv1alpha2.User {
	return &iamv1alpha2.User{
		TypeMeta: metav1.TypeMeta{APIVersion: iamv1alpha2.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: iamv1alpha2.UserSpec{
			Email:       fmt.Sprintf("%s@kubesphere.io", name),
			Lang:        "zh-CN",
			Description: "fake user",
		},
	}
}

func (f *fixture) newController() (*Controller, ksinformers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
	f.ksclient = fake.NewSimpleClientset(f.objects...)
	f.k8sclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	ldapClient := ldapclient.NewSimpleLdap()

	ksinformers := ksinformers.NewSharedInformerFactory(f.ksclient, noResyncPeriodFunc())
	k8sinformers := kubeinformers.NewSharedInformerFactory(f.k8sclient, noResyncPeriodFunc())

	for _, user := range f.userLister {
		err := ksinformers.Iam().V1alpha2().Users().Informer().GetIndexer().Add(user)
		if err != nil {
			f.t.Errorf("add user:%s", err)
		}
	}

	c := NewUserController(f.k8sclient, f.ksclient, nil,
		ksinformers.Iam().V1alpha2().Users(),
		nil, nil,
		ksinformers.Iam().V1alpha2().LoginRecords(),
		k8sinformers.Core().V1().ConfigMaps(),
		ldapClient, nil,
		options.NewAuthenticateOptions(), false)
	c.userSynced = alwaysReady
	c.recorder = &record.FakeRecorder{}

	return c, ksinformers, k8sinformers
}

func (f *fixture) run(userName string) {
	f.runController(userName, true, false)
}

func (f *fixture) runExpectError(userName string) {
	f.runController(userName, true, true)
}

func (f *fixture) runController(user string, startInformers bool, expectError bool) {
	c, i, k8sI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
		k8sI.Start(stopCh)
	}

	err := c.reconcile(user)
	if !expectError && err != nil {
		f.t.Errorf("error syncing user: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing user, got nil")
	}

	actions := filterInformerActions(f.ksclient.Actions())
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}

	k8sActions := filterInformerActions(f.k8sclient.Actions())
	for i, action := range k8sActions {
		if len(f.kubeactions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(k8sActions)-len(f.kubeactions), k8sActions[i:])
			break
		}

		expectedAction := f.kubeactions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.kubeactions) > len(k8sActions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.kubeactions)-len(k8sActions), f.kubeactions[len(k8sActions):])
	}
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()
		expUser := expObject.(*iamv1alpha2.User)
		user := object.(*iamv1alpha2.User)
		expUser.Status.LastTransitionTime = nil
		user.Status.LastTransitionTime = nil
		if !reflect.DeepEqual(expUser, user) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expPatch := e.GetPatch()
		patch := a.GetPatch()

		if !reflect.DeepEqual(expPatch, patch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expPatch, patch))
		}
	default:
		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	var ret []core.Action
	for _, action := range actions {
		if !action.Matches("update", "users") {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectUpdateUserStatusAction(user *iamv1alpha2.User) {
	expect := user.DeepCopy()
	expect.Finalizers = []string{"finalizers.kubesphere.io/users"}
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "users"}, "", expect)
	f.actions = append(f.actions, action)

	expect = expect.DeepCopy()
	expect.Status.State = iamv1alpha2.UserActive
	expect.Annotations = map[string]string{iamv1alpha2.PasswordEncryptedAnnotation: "true"}
	action = core.NewUpdateAction(schema.GroupVersionResource{Resource: "users"}, "", expect)
	f.actions = append(f.actions, action)
}

func getKey(user *iamv1alpha2.User, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(user)
	if err != nil {
		t.Errorf("Unexpected error getting key for user %v: %v", user.Name, err)
		return ""
	}
	return key
}

func TestDoNothing(t *testing.T) {
	f := newFixture(t)
	user := newUser("test")

	f.userLister = append(f.userLister, user)
	f.objects = append(f.objects, user)

	f.expectUpdateUserStatusAction(user)
	f.run(getKey(user, t))
}
