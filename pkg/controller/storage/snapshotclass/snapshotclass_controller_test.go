/*

 Copyright 2021 The KubeSphere Authors.

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

package snapshotclass

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	snapshotV1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	snapFake "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/fake"
	snapInformersV1 "github.com/kubernetes-csi/external-snapshotter/client/v4/informers/externalversions"
	storageV1 "k8s.io/api/storage/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	k8sInformers "k8s.io/client-go/informers"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"

	ksfake "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
)

var (
	noReSyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T
	//nolint:unused
	snapshotSupported bool
	// Clients
	k8sClient           *k8sFake.Clientset
	snapshotClassClient *snapFake.Clientset
	//nolint:unused
	ksClient *ksfake.Clientset
	// Objects from here preload into NewSimpleFake.
	storageObjects       []runtime.Object // include StorageClass
	snapshotClassObjects []runtime.Object
	// Objects to put in the store.
	storageClassLister  []*storageV1.StorageClass
	snapshotClassLister []*snapshotV1.VolumeSnapshotClass
	// Actions expected to happen on the client.
	actions []core.Action
}

func newFixture(t *testing.T) *fixture {
	return &fixture{t: t}
}

func (f *fixture) newController() (*VolumeSnapshotClassController, k8sInformers.SharedInformerFactory, snapInformersV1.SharedInformerFactory) {
	f.k8sClient = k8sFake.NewSimpleClientset(f.storageObjects...)
	f.snapshotClassClient = snapFake.NewSimpleClientset(f.snapshotClassObjects...)

	k8sInformers := k8sInformers.NewSharedInformerFactory(f.k8sClient, noReSyncPeriodFunc())
	snapshotInformers := snapInformersV1.NewSharedInformerFactory(f.snapshotClassClient, noReSyncPeriodFunc())

	c := NewController(
		k8sInformers.Storage().V1().StorageClasses(),
		f.snapshotClassClient.SnapshotV1().VolumeSnapshotClasses(),
		snapshotInformers.Snapshot().V1().VolumeSnapshotClasses(),
	)

	for _, storageClass := range f.storageClassLister {
		_ = k8sInformers.Storage().V1().StorageClasses().Informer().GetIndexer().Add(storageClass)
	}
	for _, snapshotClass := range f.snapshotClassLister {
		_ = snapshotInformers.Snapshot().V1().VolumeSnapshotClasses().Informer().GetIndexer().Add(snapshotClass)
	}

	return c, k8sInformers, snapshotInformers
}

func (f *fixture) runController(scName string, startInformers bool, expectError bool) {
	c, k8sI, snapI := f.newController()

	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		k8sI.Start(stopCh)
		snapI.Start(stopCh)
	}

	err := c.syncHandler(scName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing, got nil")
	}

	var actions []core.Action
	actions = append(actions, f.snapshotClassClient.Actions()...)
	actions = append(actions, f.k8sClient.Actions()...)
	filerActions := filterInformerActions(actions)
	if len(filerActions) != len(f.actions) {
		f.t.Errorf("count of actions: differ (-got, +want): %s", cmp.Diff(filerActions, f.actions))
		return
	}
	for i, action := range filerActions {
		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}
}

func (f *fixture) run(scName string) {
	f.runController(scName, true, false)
}

func (f *fixture) expectCreateSnapshotClassAction(snapshotClass *snapshotV1.VolumeSnapshotClass) {
	f.actions = append(f.actions, core.NewCreateAction(
		schema.GroupVersionResource{Resource: "volumesnapshotclasses"}, snapshotClass.Namespace, snapshotClass))
}

func (f *fixture) expectDeleteSnapshotClassAction(snapshotClass *snapshotV1.VolumeSnapshotClass) {
	f.actions = append(f.actions, core.NewDeleteAction(
		schema.GroupVersionResource{Resource: "volumesnapshotclasses"}, snapshotClass.Namespace, snapshotClass.Name))
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	var ret []core.Action
	for _, action := range actions {
		if action.GetVerb() == "list" || action.GetVerb() == "watch" {
			continue
		}
		ret = append(ret, action)
	}
	return ret
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("\nExpected\n\t%#v\ngot\n\t%#v", expected, actual)
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
		if difference := cmp.Diff(object, expObject); len(difference) > 0 {
			t.Errorf("[CreateAction] %T differ (-got, +want): %s", expObject, difference)
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()
		if difference := cmp.Diff(object, expObject); len(difference) > 0 {
			t.Errorf("[UpdateAction] %T differ (-got, +want): %s", expObject, difference)
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
		if difference := cmp.Diff(e.Name, a.Name); len(difference) > 0 {
			t.Errorf("[UpdateAction] %T differ (-got, +want): %s", e.Name, difference)
		}
	default:
		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}

func newStorageClass(name string) *storageV1.StorageClass {
	isExpansion := true
	return &storageV1.StorageClass{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		AllowVolumeExpansion: &isExpansion,
	}
}

func newSnapshotClass(storageClass *storageV1.StorageClass) *snapshotV1.VolumeSnapshotClass {
	return &snapshotV1.VolumeSnapshotClass{
		ObjectMeta: v1.ObjectMeta{
			Name: storageClass.Name,
		},
		Driver:         storageClass.Provisioner,
		DeletionPolicy: snapshotV1.VolumeSnapshotContentDelete,
	}
}

func getKey(sc *storageV1.StorageClass, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sc)
	if err != nil {
		t.Errorf("Unexpected error getting key for %v: %v", sc.Name, err)
		return ""
	}
	return key
}

func TestCreateStorageClass(t *testing.T) {
	fixture := newFixture(t)
	storageClass := newStorageClass("csi-example")
	storageClass.Annotations = map[string]string{annotationAllowSnapshot: "true"}
	snapshotClass := newSnapshotClass(storageClass)

	// Objects exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)

	// Action expected
	fixture.expectCreateSnapshotClassAction(snapshotClass)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestDeleteStorageClass(t *testing.T) {
	storageClass := newStorageClass("csi-example")
	snapshotClass := newSnapshotClass(storageClass)

	fixture := newFixture(t)
	// Object exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass)
	fixture.snapshotClassObjects = append(fixture.snapshotClassObjects, snapshotClass)
	fixture.snapshotClassLister = append(fixture.snapshotClassLister, snapshotClass)

	// Action expected
	fixture.expectDeleteSnapshotClassAction(snapshotClass)

	// Run test
	fixture.run(getKey(storageClass, t))
}
