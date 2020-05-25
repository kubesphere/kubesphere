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

package job

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/client-go/informers"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"testing"
	"time"
)

var (
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t             *testing.T
	kubeclient    *k8sfake.Clientset
	jobController *JobController
	jobLister     []*batchv1.Job

	kubeactions []core.Action
	actions     []core.Action

	kubeobjects []runtime.Object
	objects     []runtime.Object
}

func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "jobs") ||
				action.Matches("watch", "jobs")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func newJob(name string, spec batchv1.JobSpec) *batchv1.Job {
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: spec,
	}

	return job
}

func newFixture(t *testing.T) *fixture {

	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}

	return f
}

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

func (f *fixture) newController() (*JobController, informers.SharedInformerFactory) {
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)

	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())

	jobController := NewJobController(k8sI.Batch().V1().Jobs(), f.kubeclient)

	for _, job := range f.jobLister {
		_ = k8sI.Batch().V1().Jobs().Informer().GetIndexer().Add(job)
	}

	return jobController, k8sI
}

func (f *fixture) runController(jobName string, startInformers bool, expectError bool) {
	c, k8sI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		k8sI.Start(stopCh)
	}
	err := c.syncJob(jobName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing job: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing job, got nil")
	}

	actions := filterInformerActions(f.kubeclient.Actions())
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

func (f *fixture) expectAddAnnotationAction(job *batchv1.Job) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "jobs"}, job.Namespace, job)
	f.actions = append(f.actions, action)
}

func (f *fixture) run(jobName string) {
	f.runController(jobName, true, false)
}

func TestAddAnnotation(t *testing.T) {
	f := newFixture(t)
	job := newJob("test", batchv1.JobSpec{})

	f.jobLister = append(f.jobLister, job)
	f.objects = append(f.objects, job)

	f.kubeobjects = append(f.kubeobjects, job)

	f.expectAddAnnotationAction(job)
	f.run(getKey(job, t))
}

func getKey(job *batchv1.Job, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(job)
	if err != nil {
		t.Errorf("Unexpected error getting key for job %v: %v", job.Name, err)
		return ""
	}
	return key
}
