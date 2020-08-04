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

package s2irun

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	s2i "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	client     *fake.Clientset
	kubeclient *k8sfake.Clientset
	// Objects to put in the store.
	s2ibinaryLister []*s2i.S2iBinary
	s2irunLister    []*s2i.S2iRun
	actions         []core.Action
	// Objects from here preloaded into NewSimpleFake.
	objects []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	return f
}

func newS2iBinary(name string) *s2i.S2iBinary {
	return &s2i.S2iBinary{
		TypeMeta: metav1.TypeMeta{APIVersion: s2i.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: s2i.S2iBinarySpec{},
	}
}

func newS2iBinaryWithCreateTime(name string, createTime metav1.Time) *s2i.S2iBinary {
	return &s2i.S2iBinary{
		TypeMeta: metav1.TypeMeta{APIVersion: s2i.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         metav1.NamespaceDefault,
			CreationTimestamp: createTime,
		},
		Spec: s2i.S2iBinarySpec{},
	}
}

func newS2iRun(name string, s2iBinaryName string) *s2i.S2iRun {
	return &s2i.S2iRun{
		TypeMeta: metav1.TypeMeta{APIVersion: s2i.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				s2i.S2iBinaryLabelKey: s2iBinaryName,
			},
		},
	}
}

func newDeletetingS2iRun(name string, s2iBinaryName string) *s2i.S2iRun {
	now := metav1.Now()
	return &s2i.S2iRun{
		TypeMeta: metav1.TypeMeta{APIVersion: s2i.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				s2i.S2iBinaryLabelKey: s2iBinaryName,
			},
			Finalizers:        []string{s2i.S2iBinaryFinalizerName},
			DeletionTimestamp: &now,
		},
	}
}

func (f *fixture) newController() (*Controller, informers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset()

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())

	c := NewS2iRunController(f.kubeclient, f.client,
		i.Devops().V1alpha1().S2iBinaries(), i.Devops().V1alpha1().S2iRuns())

	c.s2iBinarySynced = alwaysReady
	c.eventRecorder = &record.FakeRecorder{}

	for _, f := range f.s2ibinaryLister {
		i.Devops().V1alpha1().S2iBinaries().Informer().GetIndexer().Add(f)
	}
	for _, f := range f.s2irunLister {
		i.Devops().V1alpha1().S2iRuns().Informer().GetIndexer().Add(f)
	}

	return c, i
}

func (f *fixture) run(fooName string) {
	f.runController(fooName, true, false)
}

func (f *fixture) runExpectError(fooName string) {
	f.runController(fooName, true, true)
}

func (f *fixture) runController(s2iRunName string, startInformers bool, expectError bool) {
	c, i := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
	}

	err := c.syncHandler(s2iRunName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing foo: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing foo, got nil")
	}

	actions := filterInformerActions(f.client.Actions())
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

		if !reflect.DeepEqual(expObject, object) {
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
	case core.DeleteActionImpl:
		e, _ := expected.(core.DeleteActionImpl)

		expName := e.GetName()
		objectName := a.GetName()

		if !reflect.DeepEqual(expName, objectName) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expName, objectName))
		}
		expNamespace := e.GetNamespace()
		objectNamespace := a.GetNamespace()
		if !reflect.DeepEqual(expNamespace, objectNamespace) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expNamespace, objectNamespace))
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
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", s2i.ResourcePluralS2iRun) ||
				action.Matches("watch", s2i.ResourcePluralS2iRun) ||
				action.Matches("list", s2i.ResourcePluralS2iBinary) ||
				action.Matches("watch", s2i.ResourcePluralS2iBinary)) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectUpdateS2iRunAction(s2iRun *s2i.S2iRun) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: s2i.ResourcePluralS2iRun}, s2iRun.Namespace, s2iRun)
	f.actions = append(f.actions, action)
}

func (f *fixture) expectDeleteS2iBinaryAction(s2iBinary *s2i.S2iBinary) {
	action := core.NewDeleteAction(schema.GroupVersionResource{Resource: s2i.ResourcePluralS2iBinary}, s2iBinary.Namespace, s2iBinary.Name)
	f.actions = append(f.actions, action)
}

func getKey(s2i *s2i.S2iRun, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(s2i)
	if err != nil {
		t.Errorf("Unexpected error getting key for s2i %v: %v", s2i.Name, err)
		return ""
	}
	return key
}

func TestDoNothing(t *testing.T) {
	f := newFixture(t)
	s2iBinary := newS2iBinary("test")
	s2iRun := newS2iRun("test", s2iBinary.Name)

	f.s2ibinaryLister = append(f.s2ibinaryLister, s2iBinary)
	f.s2irunLister = append(f.s2irunLister, s2iRun)
	f.objects = append(f.objects, s2iBinary)
	f.objects = append(f.objects, s2iRun)

	f.expectUpdateS2iRunAction(s2iRun)
	f.run(getKey(s2iRun, t))
}

func TestDeleteS2iBinary(t *testing.T) {
	f := newFixture(t)
	s2iBinary := newS2iBinary("test")
	s2iRun := newDeletetingS2iRun("test", s2iBinary.Name)

	f.s2ibinaryLister = append(f.s2ibinaryLister, s2iBinary)
	f.s2irunLister = append(f.s2irunLister, s2iRun)
	f.objects = append(f.objects, s2iBinary)
	f.objects = append(f.objects, s2iRun)

	f.expectDeleteS2iBinaryAction(s2iBinary)
	f.expectUpdateS2iRunAction(s2iRun)
	f.run(getKey(s2iRun, t))
}

func TestDeleteOtherS2iBinary(t *testing.T) {
	f := newFixture(t)
	s2iBinary := newS2iBinary("test")
	s2iRun := newDeletetingS2iRun("test", s2iBinary.Name)
	otherS2iBinary := newS2iBinaryWithCreateTime("test2", metav1.NewTime(time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)))

	f.s2ibinaryLister = append(f.s2ibinaryLister, s2iBinary)
	f.s2ibinaryLister = append(f.s2ibinaryLister, otherS2iBinary)
	f.s2irunLister = append(f.s2irunLister, s2iRun)
	f.objects = append(f.objects, s2iBinary)
	f.objects = append(f.objects, s2iRun)
	f.objects = append(f.objects, otherS2iBinary)
	f.expectDeleteS2iBinaryAction(s2iBinary)
	f.expectDeleteS2iBinaryAction(otherS2iBinary)
	f.expectUpdateS2iRunAction(s2iRun)
	f.run(getKey(s2iRun, t))
}
