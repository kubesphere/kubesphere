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

package groupbinding

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	v1alpha2 "kubesphere.io/api/iam/v1alpha2"
	fedv1beta1types "kubesphere.io/api/types/v1beta1"

	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/constants"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

func init() {
	v1alpha2.AddToScheme(scheme.Scheme)
}

type fixture struct {
	t *testing.T

	ksclient  *fake.Clientset
	k8sclient *k8sfake.Clientset
	// Objects to put in the store.
	groupBindingLister    []*v1alpha2.GroupBinding
	fedgroupBindingLister []*fedv1beta1types.FederatedGroupBinding
	userLister            []*v1alpha2.User
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

func newGroupBinding(name string, users []string) *v1alpha2.GroupBinding {
	return &v1alpha2.GroupBinding{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha2.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-binding", name),
		},
		GroupRef: v1alpha2.GroupRef{
			Name: name,
		},
		Users: users,
	}
}

func newUser(name string) *v1alpha2.User {
	return &v1alpha2.User{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha2.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha2.UserSpec{
			Email:       fmt.Sprintf("%s@kubesphere.io", name),
			Lang:        "zh-CN",
			Description: "fake user",
		},
	}
}

func newFederatedGroupBinding(groupBinding *iamv1alpha2.GroupBinding) *fedv1beta1types.FederatedGroupBinding {
	return &fedv1beta1types.FederatedGroupBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       fedv1beta1types.FederatedGroupBindingKind,
			APIVersion: fedv1beta1types.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: groupBinding.Name,
		},
		Spec: fedv1beta1types.FederatedGroupBindingSpec{
			Template: fedv1beta1types.GroupBindingTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels: groupBinding.Labels,
				},
				GroupRef: groupBinding.GroupRef,
				Users:    groupBinding.Users,
			},
			Placement: fedv1beta1types.GenericPlacementFields{
				ClusterSelector: &metav1.LabelSelector{},
			},
		},
	}
}

func (f *fixture) newController() (*Controller, ksinformers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
	f.ksclient = fake.NewSimpleClientset(f.objects...)
	f.k8sclient = k8sfake.NewSimpleClientset(f.kubeobjects...)

	ksinformers := ksinformers.NewSharedInformerFactory(f.ksclient, noResyncPeriodFunc())
	k8sinformers := kubeinformers.NewSharedInformerFactory(f.k8sclient, noResyncPeriodFunc())

	for _, groupBinding := range f.groupBindingLister {
		err := ksinformers.Iam().V1alpha2().GroupBindings().Informer().GetIndexer().Add(groupBinding)
		if err != nil {
			f.t.Errorf("add groupBinding:%s", err)
		}
	}

	for _, u := range f.userLister {
		err := ksinformers.Iam().V1alpha2().Users().Informer().GetIndexer().Add(u)
		if err != nil {
			f.t.Errorf("add groupBinding:%s", err)
		}
	}

	c := NewController(f.k8sclient, f.ksclient,
		ksinformers.Iam().V1alpha2().GroupBindings(),
		ksinformers.Types().V1beta1().FederatedGroupBindings(), true)
	c.Synced = []cache.InformerSynced{alwaysReady}
	c.recorder = &record.FakeRecorder{}

	return c, ksinformers, k8sinformers
}

func (f *fixture) run(userName string) {
	f.runController(userName, true, false)
}

func (f *fixture) runExpectError(userName string) {
	f.runController(userName, true, true)
}

func (f *fixture) runController(groupBinding string, startInformers bool, expectError bool) {
	c, i, k8sI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
		k8sI.Start(stopCh)
	}

	err := c.reconcile(groupBinding)
	if !expectError && err != nil {
		f.t.Errorf("error syncing groupBinding: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing groupBinding, got nil")
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
		expUser := expObject.(*v1alpha2.GroupBinding)
		groupBinding := object.(*v1alpha2.GroupBinding)

		if !reflect.DeepEqual(expUser, groupBinding) {
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
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "groupbindings") ||
				action.Matches("watch", "groupbindings") ||
				action.Matches("list", "federatedgroupbindings") ||
				action.Matches("list", "users") ||
				action.Matches("watch", "users") ||
				action.Matches("get", "users")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectUpdateGroupsFinalizerAction(groupBinding *v1alpha2.GroupBinding) {
	expect := groupBinding.DeepCopy()
	expect.Finalizers = []string{"finalizers.kubesphere.io/groupsbindings"}
	expect.Labels = map[string]string{constants.KubefedManagedLabel: "false"}
	action := core.NewUpdateAction(schema.GroupVersionResource{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "groupbindings"}, "", expect)
	f.actions = append(f.actions, action)
}

func (f *fixture) expectUpdateGroupsDeleteAction(groupBinding *v1alpha2.GroupBinding) {
	expect := groupBinding.DeepCopy()
	expect.Finalizers = []string{}
	action := core.NewUpdateAction(schema.GroupVersionResource{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "groupbindings"}, "", expect)
	f.actions = append(f.actions, action)
}

func (f *fixture) expectPatchUserAction(user *v1alpha2.User, groups []string) {
	newUser := user.DeepCopy()
	newUser.Spec.Groups = groups
	patch := client.MergeFrom(user)
	patchData, _ := patch.Data(newUser)

	f.actions = append(f.actions, core.NewPatchAction(schema.GroupVersionResource{Group: "iam.kubesphere.io", Resource: "users", Version: "v1alpha2"}, user.Namespace, user.Name, patch.Type(), patchData))
}

func (f *fixture) expectCreateFederatedGroupBindingsAction(groupBinding *v1alpha2.GroupBinding) {
	b := newFederatedGroupBinding(groupBinding)

	controllerutil.SetControllerReference(groupBinding, b, scheme.Scheme)

	actionCreate := core.NewCreateAction(schema.GroupVersionResource{Group: "types.kubefed.io", Version: "v1beta1", Resource: "federatedgroupbindings"}, "", b)
	f.actions = append(f.actions, actionCreate)
}

func getKey(groupBinding *v1alpha2.GroupBinding, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(groupBinding)
	if err != nil {
		t.Errorf("Unexpected error getting key for groupBinding %v: %v", groupBinding.Name, err)
		return ""
	}
	return key
}

func TestCreatesGroupBinding(t *testing.T) {
	f := newFixture(t)

	users := []string{"user1"}
	groupbinding := newGroupBinding("test", users)
	groupbinding.ObjectMeta.Finalizers = append(groupbinding.ObjectMeta.Finalizers, finalizer)
	groupbinding.Labels = map[string]string{constants.KubefedManagedLabel: "false"}
	f.groupBindingLister = append(f.groupBindingLister, groupbinding)
	f.objects = append(f.objects, groupbinding)

	user := newUser("user1")
	f.userLister = append(f.userLister, user)

	f.objects = append(f.objects, user)

	excepctGroups := []string{"test"}

	f.expectPatchUserAction(user, excepctGroups)
	f.expectCreateFederatedGroupBindingsAction(groupbinding)

	f.run(getKey(groupbinding, t))
}

func TestDeletesGroupBinding(t *testing.T) {
	f := newFixture(t)

	users := []string{"user1"}
	groupbinding := newGroupBinding("test", users)
	deletedGroup := groupbinding.DeepCopy()
	deletedGroup.Finalizers = append(groupbinding.ObjectMeta.Finalizers, finalizer)

	now := metav1.Now()
	deletedGroup.ObjectMeta.DeletionTimestamp = &now

	f.groupBindingLister = append(f.groupBindingLister, deletedGroup)
	f.objects = append(f.objects, deletedGroup)

	user := newUser("user1")
	user.Spec.Groups = []string{"test"}
	f.userLister = append(f.userLister, user)
	f.objects = append(f.objects, user)

	f.expectPatchUserAction(user, nil)
	f.expectUpdateGroupsDeleteAction(deletedGroup)

	f.run(getKey(deletedGroup, t))
}

func TestDoNothing(t *testing.T) {
	f := newFixture(t)

	users := []string{"user1"}
	groupBinding := newGroupBinding("test", users)

	f.groupBindingLister = append(f.groupBindingLister, groupBinding)
	f.objects = append(f.objects, groupBinding)

	f.expectUpdateGroupsFinalizerAction(groupBinding)
	f.run(getKey(groupBinding, t))

}
