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

package devopsproject

import (
	v1 "k8s.io/api/core/v1"
	devopsprojects "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/constants"
	fakeDevOps "kubesphere.io/kubesphere/pkg/simple/client/devops/fake"
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	devops "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
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
	devopsProjectLister []*devops.DevOpsProject
	namespaceLister     []*v1.Namespace
	actions             []core.Action
	kubeactions         []core.Action

	kubeobjects []runtime.Object
	// Objects from here preloaded into NewSimpleFake.
	objects []runtime.Object
	// Objects from here preloaded into devops
	initDevOpsProject   []string
	expectDevOpsProject []string
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	return f
}

func newDevOpsProject(name string, nsName string, withFinalizers bool, withStatus bool) *devopsprojects.DevOpsProject {
	project := &devopsprojects.DevOpsProject{
		TypeMeta: metav1.TypeMeta{APIVersion: devopsprojects.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if withFinalizers {
		project.Finalizers = []string{devopsprojects.DevOpsProjectFinalizerName}
	}
	if withStatus {
		project.Status = devops.DevOpsProjectStatus{AdminNamespace: nsName}
	}
	return project
}

func newNamespace(name string, projectName string, useGenerateName, withOwnerReference bool) *v1.Namespace {
	ns := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{constants.DevOpsProjectLabelKey: projectName},
		},
	}
	if useGenerateName {
		ns.ObjectMeta.Name = ""
		ns.ObjectMeta.GenerateName = projectName
	}
	if withOwnerReference {
		TRUE := true
		ns.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion:         devops.SchemeGroupVersion.String(),
				Kind:               devops.ResourceKindDevOpsProject,
				Name:               projectName,
				BlockOwnerDeletion: &TRUE,
				Controller:         &TRUE,
			},
		}
	}
	return ns
}

func newDeletingDevOpsProject(name string) *devopsprojects.DevOpsProject {
	now := metav1.Now()
	return &devopsprojects.DevOpsProject{
		TypeMeta: metav1.TypeMeta{APIVersion: devopsprojects.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			DeletionTimestamp: &now,
			Finalizers:        []string{devopsprojects.DevOpsProjectFinalizerName},
		},
	}
}

func (f *fixture) newController() (*Controller, informers.SharedInformerFactory, kubeinformers.SharedInformerFactory, *fakeDevOps.Devops) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())
	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())
	dI := fakeDevOps.New(f.initDevOpsProject...)

	c := NewController(f.kubeclient, f.client, dI,
		k8sI.Core().V1().Namespaces(),
		i.Devops().V1alpha3().DevOpsProjects(),
		i.Tenant().V1alpha1().Workspaces())

	c.devOpsProjectSynced = alwaysReady
	c.eventRecorder = &record.FakeRecorder{}

	for _, f := range f.devopsProjectLister {
		i.Devops().V1alpha3().DevOpsProjects().Informer().GetIndexer().Add(f)
	}

	for _, d := range f.namespaceLister {
		k8sI.Core().V1().Namespaces().Informer().GetIndexer().Add(d)
	}

	return c, i, k8sI, dI
}

func (f *fixture) run(fooName string) {
	f.runController(fooName, true, false)
}

func (f *fixture) runExpectError(fooName string) {
	f.runController(fooName, true, true)
}

func (f *fixture) runController(projectName string, startInformers bool, expectError bool) {
	c, i, k8sI, dI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
		k8sI.Start(stopCh)
	}

	err := c.syncHandler(projectName)
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
	k8sActions := filterInformerActions(f.kubeclient.Actions())
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

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}
	if len(dI.Projects) != len(f.expectDevOpsProject) {
		f.t.Errorf(" unexpected objects: %v", dI.Projects)
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
			(action.Matches("list", devopsprojects.ResourcePluralDevOpsProject) ||
				action.Matches("watch", devopsprojects.ResourcePluralDevOpsProject) ||
				action.Matches("list", "namespaces") ||
				action.Matches("watch", "namespaces") ||
				action.Matches("watch", "workspaces") ||
				action.Matches("list", "workspaces")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectUpdateDevOpsProjectAction(p *devopsprojects.DevOpsProject) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: devopsprojects.ResourcePluralDevOpsProject},
		p.Namespace, p)
	f.actions = append(f.actions, action)
}

func (f *fixture) expectUpdateNamespaceAction(p *v1.Namespace) {
	action := core.NewUpdateAction(schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}, p.Namespace, p)
	f.kubeactions = append(f.kubeactions, action)
}

func (f *fixture) expectCreateNamespaceAction(p *v1.Namespace) {
	action := core.NewCreateAction(schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}, p.Namespace, p)
	f.kubeactions = append(f.kubeactions, action)
}

func getKey(p *devopsprojects.DevOpsProject, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(p)
	if err != nil {
		t.Errorf("Unexpected error getting key for devopsprojects %v: %v", p.Name, err)
		return ""
	}
	return key
}

func TestDoNothing(t *testing.T) {
	f := newFixture(t)
	nsName := "test-123"
	projectName := "test"
	project := newDevOpsProject(projectName, nsName, true, true)
	ns := newNamespace(nsName, projectName, false, true)

	f.devopsProjectLister = append(f.devopsProjectLister, project)
	f.namespaceLister = append(f.namespaceLister, ns)
	f.objects = append(f.objects, project)
	f.initDevOpsProject = []string{ns.Name}
	f.expectDevOpsProject = []string{ns.Name}

	f.run(getKey(project, t))
}

func TestUpdateProjectFinalizers(t *testing.T) {
	f := newFixture(t)
	nsName := "test-123"
	projectName := "test"
	project := newDevOpsProject(projectName, nsName, false, true)
	ns := newNamespace(nsName, projectName, false, true)

	f.devopsProjectLister = append(f.devopsProjectLister, project)
	f.namespaceLister = append(f.namespaceLister, ns)
	f.objects = append(f.objects, project)
	f.kubeobjects = append(f.kubeobjects, ns)
	f.initDevOpsProject = []string{ns.Name}
	f.expectDevOpsProject = []string{ns.Name}
	expectUpdateProject := project.DeepCopy()
	expectUpdateProject.Finalizers = []string{devops.DevOpsProjectFinalizerName}
	f.expectUpdateDevOpsProjectAction(expectUpdateProject)
	f.run(getKey(project, t))
}

func TestUpdateProjectStatus(t *testing.T) {
	f := newFixture(t)
	nsName := "test-123"
	projectName := "test"
	project := newDevOpsProject(projectName, nsName, true, false)
	ns := newNamespace(nsName, projectName, false, true)

	f.devopsProjectLister = append(f.devopsProjectLister, project)
	f.namespaceLister = append(f.namespaceLister, ns)
	f.objects = append(f.objects, project)
	f.kubeobjects = append(f.kubeobjects, ns)
	f.initDevOpsProject = []string{ns.Name}
	f.expectDevOpsProject = []string{ns.Name}
	expectUpdateProject := project.DeepCopy()
	expectUpdateProject.Status.AdminNamespace = nsName
	f.expectUpdateDevOpsProjectAction(expectUpdateProject)
	f.run(getKey(project, t))
}

func TestUpdateNsOwnerReference(t *testing.T) {
	f := newFixture(t)
	nsName := "test-123"
	projectName := "test"
	project := newDevOpsProject(projectName, nsName, true, true)
	ns := newNamespace(nsName, projectName, false, false)

	f.devopsProjectLister = append(f.devopsProjectLister, project)
	f.namespaceLister = append(f.namespaceLister, ns)
	f.objects = append(f.objects, project)
	f.kubeobjects = append(f.kubeobjects, ns)
	f.initDevOpsProject = []string{ns.Name}
	f.expectDevOpsProject = []string{ns.Name}
	expectUpdateNs := newNamespace(nsName, projectName, false, true)

	f.expectUpdateNamespaceAction(expectUpdateNs)
	f.run(getKey(project, t))
}

func TestCreateDevOpsProjects(t *testing.T) {
	f := newFixture(t)
	project := newDevOpsProject("test", "", true, false)
	ns := newNamespace("test", "test", false, true)
	f.devopsProjectLister = append(f.devopsProjectLister, project)
	f.objects = append(f.objects, project)
	f.expectDevOpsProject = []string{""}
	expect := project.DeepCopy()
	expect.Status.AdminNamespace = "test"
	f.expectUpdateDevOpsProjectAction(expect)
	f.expectCreateNamespaceAction(ns)
	f.run(getKey(project, t))
}

func TestDeleteDevOpsProjects(t *testing.T) {
	f := newFixture(t)
	project := newDeletingDevOpsProject("test")

	f.devopsProjectLister = append(f.devopsProjectLister, project)
	f.objects = append(f.objects, project)
	f.initDevOpsProject = []string{project.Name}
	f.expectDevOpsProject = []string{project.Name}
	expectProject := project.DeepCopy()
	expectProject.Finalizers = []string{}
	f.expectUpdateDevOpsProjectAction(expectProject)
	f.run(getKey(project, t))
}

func TestDeleteDevOpsProjectsWithNull(t *testing.T) {
	f := newFixture(t)
	project := newDeletingDevOpsProject("test")

	f.devopsProjectLister = append(f.devopsProjectLister, project)
	f.objects = append(f.objects, project)
	f.expectDevOpsProject = []string{}
	expectProject := project.DeepCopy()
	expectProject.Finalizers = []string{}
	f.expectUpdateDevOpsProjectAction(expectProject)
	f.run(getKey(project, t))
}
