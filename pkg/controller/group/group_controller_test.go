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

package group

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	v1alpha2 "kubesphere.io/api/iam/v1alpha2"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"
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
	tenantv1alpha2.AddToScheme(scheme.Scheme)
}

type fixture struct {
	t *testing.T

	ksclient  *fake.Clientset
	k8sclient *k8sfake.Clientset
	// Objects to put in the store.
	groupLister    []*v1alpha2.Group
	fedgroupLister []*fedv1beta1types.FederatedGroup
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

func newGroup(name string) *v1alpha2.Group {
	return &v1alpha2.Group{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha2.SchemeGroupVersion.String(), Kind: "Group"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha2.GroupSpec{},
	}
}

func newUnmanagedGroup(name string) *v1alpha2.Group {
	return &v1alpha2.Group{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha2.SchemeGroupVersion.String(), Kind: "Group"},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Labels:     map[string]string{constants.KubefedManagedLabel: "false"},
			Finalizers: []string{"finalizers.kubesphere.io/groups"},
		},
		Spec: v1alpha2.GroupSpec{},
	}
}

func newFederatedGroup(group *v1alpha2.Group) *fedv1beta1types.FederatedGroup {
	return &fedv1beta1types.FederatedGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: group.Name,
		},
		Spec: fedv1beta1types.FederatedGroupSpec{
			Template: fedv1beta1types.GroupTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels: group.Labels,
				},
				Spec: group.Spec,
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

	for _, group := range f.groupLister {
		err := ksinformers.Iam().V1alpha2().Groups().Informer().GetIndexer().Add(group)
		if err != nil {
			f.t.Errorf("add group:%s", err)
		}
	}

	for _, group := range f.fedgroupLister {
		err := ksinformers.Types().V1beta1().FederatedGroups().Informer().GetIndexer().Add(group)
		if err != nil {
			f.t.Errorf("add federated group:%s", err)
		}
	}

	c := NewController(f.k8sclient, f.ksclient,
		ksinformers.Iam().V1alpha2().Groups(),
		ksinformers.Types().V1beta1().FederatedGroups(), true)
	c.recorder = &record.FakeRecorder{}

	return c, ksinformers, k8sinformers
}

func (f *fixture) run(userName string) {
	f.runController(userName, true, false)
}

func (f *fixture) runExpectError(userName string) {
	f.runController(userName, true, true)
}

func (f *fixture) runController(group string, startInformers bool, expectError bool) {
	c, i, k8sI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
		k8sI.Start(stopCh)
	}

	err := c.Handler(group)
	if !expectError && err != nil {
		f.t.Errorf("error syncing group: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing group, got nil")
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
	case core.DeleteCollectionActionImpl:
		e, _ := expected.(core.DeleteCollectionActionImpl)
		exp := e.GetListRestrictions()
		target := a.GetListRestrictions()
		if !reflect.DeepEqual(exp, target) {
			t.Errorf("Action %s %s has wrong Query\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(exp, target))
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
			(action.Matches("list", "groups") ||
				action.Matches("watch", "groups") ||
				action.Matches("list", "groups") ||
				action.Matches("list", "namespaces") ||
				action.Matches("get", "workspacetemplates") ||
				action.Matches("list", "federatedgroups") ||
				action.Matches("watch", "federatedgroups")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectUpdateGroupsFinalizerAction(group *v1alpha2.Group) {
	expect := group.DeepCopy()
	if expect.Labels == nil {
		expect.Labels = make(map[string]string, 0)
	}
	expect.Finalizers = []string{"finalizers.kubesphere.io/groups"}
	expect.Labels[constants.KubefedManagedLabel] = "false"
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "groups"}, "", expect)
	f.actions = append(f.actions, action)
}

func (f *fixture) expectUpdateWorkspaceRefAction(child *v1alpha2.Group, wsp *tenantv1alpha2.WorkspaceTemplate) {
	expect := child.DeepCopy()
	if expect.Labels == nil {
		expect.Labels = make(map[string]string, 0)
	}

	controllerutil.SetControllerReference(wsp, expect, scheme.Scheme)

	expect.Finalizers = []string{"finalizers.kubesphere.io/groups"}
	expect.Labels[constants.KubefedManagedLabel] = "false"
	updateAction := core.NewUpdateAction(schema.GroupVersionResource{Resource: "groups"}, "", expect)
	f.actions = append(f.actions, updateAction)
}

func (f *fixture) expectUpdateParentsRefAction(parent, child *v1alpha2.Group) {
	expect := child.DeepCopy()
	if expect.Labels == nil {
		expect.Labels = make(map[string]string, 0)
	}

	controllerutil.SetControllerReference(parent, expect, scheme.Scheme)

	expect.Finalizers = []string{"finalizers.kubesphere.io/groups"}
	expect.Labels[constants.KubefedManagedLabel] = "false"
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "groups"}, "", expect)
	f.actions = append(f.actions, action)
}

func (f *fixture) expectCreateFederatedGroupsAction(group *v1alpha2.Group) {
	federatedGroup := newFederatedGroup(group)

	controllerutil.SetControllerReference(group, federatedGroup, scheme.Scheme)

	actionCreate := core.NewCreateAction(schema.GroupVersionResource{Resource: "federatedgroups"}, "", federatedGroup)
	f.actions = append(f.actions, actionCreate)
}

func (f *fixture) expectUpdateFederatedGroupsAction(group *v1alpha2.Group) {
	g := newFederatedGroup(group)
	controllerutil.SetControllerReference(group, g, scheme.Scheme)
	actionCreate := core.NewUpdateAction(schema.GroupVersionResource{Group: "types.kubefed.io", Version: "v1beta1", Resource: "federatedgroups"}, "", g)
	f.actions = append(f.actions, actionCreate)
}

func (f *fixture) expectUpdateGroupsDeleteAction(group *v1alpha2.Group) {
	expect := group.DeepCopy()
	expect.Finalizers = []string{}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{v1alpha2.GroupReferenceLabel: group.Name}).String(),
	}

	actionDelete := core.NewDeleteCollectionAction(schema.GroupVersionResource{Resource: "groupbindings"}, "", listOptions)
	f.actions = append(f.actions, actionDelete)

	actionDelete = core.NewDeleteCollectionAction(schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "", listOptions)
	f.kubeactions = append(f.kubeactions, actionDelete)

	actionDelete = core.NewDeleteCollectionAction(schema.GroupVersionResource{Resource: "workspacerolebindings"}, "", listOptions)
	f.actions = append(f.actions, actionDelete)

	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "groups"}, "", expect)
	f.actions = append(f.actions, action)
}

func getKey(group *v1alpha2.Group, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(group)
	if err != nil {
		t.Errorf("Unexpected error getting key for group %v: %v", group.Name, err)
		return ""
	}
	return key
}

func TestDeletesGroup(t *testing.T) {
	f := newFixture(t)
	deletedGroup := newUnmanagedGroup("test")

	now := metav1.Now()
	deletedGroup.ObjectMeta.DeletionTimestamp = &now

	f.groupLister = append(f.groupLister, deletedGroup)
	f.objects = append(f.objects, deletedGroup)
	f.expectUpdateGroupsDeleteAction(deletedGroup)
	f.run(getKey(deletedGroup, t))
}

func TestDoNothing(t *testing.T) {
	f := newFixture(t)
	group := newGroup("test")

	f.groupLister = append(f.groupLister, group)
	f.objects = append(f.objects, group)

	f.expectUpdateGroupsFinalizerAction(group)
	f.run(getKey(group, t))
}

func TestGroupCreateWithParent(t *testing.T) {
	f := newFixture(t)
	parent := newGroup("parent")
	child := newGroup("child")
	child.Labels = map[string]string{v1alpha2.GroupParent: "parent"}

	f.groupLister = append(f.groupLister, parent, child)
	f.objects = append(f.objects, parent, child)

	f.expectUpdateParentsRefAction(parent, child)
	f.run(getKey(child, t))
}

func TestGroupCreateWithWorkspace(t *testing.T) {
	f := newFixture(t)
	child := newGroup("child")
	child.Labels = map[string]string{constants.WorkspaceLabelKey: "wsp"}

	f.groupLister = append(f.groupLister, child)
	f.objects = append(f.objects, child)

	wsp := tenantv1alpha2.WorkspaceTemplate{
		TypeMeta: metav1.TypeMeta{APIVersion: tenantv1alpha2.SchemeGroupVersion.String(), Kind: tenantv1alpha2.ResourceKindWorkspaceTemplate},
		ObjectMeta: metav1.ObjectMeta{
			Name: "wsp",
		},
	}
	f.objects = append(f.objects, &wsp)

	f.expectUpdateWorkspaceRefAction(child, &wsp)
	f.run(getKey(child, t))
}

func TestFederetedGroupCreate(t *testing.T) {
	f := newFixture(t)

	group := newUnmanagedGroup("test")

	f.groupLister = append(f.groupLister, group)
	f.objects = append(f.objects, group)

	f.expectCreateFederatedGroupsAction(group)
	f.run(getKey(group, t))
}

func TestFederetedGroupUpdate(t *testing.T) {
	f := newFixture(t)

	group := newUnmanagedGroup("test")

	federatedGroup := newFederatedGroup(group.DeepCopy())
	controllerutil.SetControllerReference(group, federatedGroup, scheme.Scheme)

	f.fedgroupLister = append(f.fedgroupLister, federatedGroup)
	f.objects = append(f.objects, federatedGroup)

	group.Labels["foo"] = "bar"
	f.groupLister = append(f.groupLister, group)
	f.objects = append(f.objects, group)

	f.expectUpdateFederatedGroupsAction(group.DeepCopy())
	f.run(getKey(group, t))
}
