/*
Copyright 2017 The Kubernetes Authors.

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

package pvcprotection

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/controller"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"
)

type reaction struct {
	verb      string
	resource  string
	reactorfn clienttesting.ReactionFunc
}

const (
	defaultNS       = "default"
	defaultPVCName  = "pvc1"
	defaultPodName  = "pod1"
	defaultNodeName = "node1"
)

func pod() *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPodName,
			Namespace: defaultNS,
		},
		Spec: v1.PodSpec{
			NodeName: defaultNodeName,
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	}
}

func unscheduled(pod *v1.Pod) *v1.Pod {
	pod.Spec.NodeName = ""
	return pod
}

func withPVC(pvcName string, pod *v1.Pod) *v1.Pod {
	volume := v1.Volume{
		Name: pvcName,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
	return pod
}

func withEmptyDir(pod *v1.Pod) *v1.Pod {
	volume := v1.Volume{
		Name: "emptyDir",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
	return pod
}

func withStatus(phase v1.PodPhase, pod *v1.Pod) *v1.Pod {
	pod.Status.Phase = phase
	return pod
}

func pvc() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPVCName,
			Namespace: defaultNS,
		},
	}
}

func withProtectionFinalizer(pvc *v1.PersistentVolumeClaim) *v1.PersistentVolumeClaim {
	pvc.Finalizers = append(pvc.Finalizers, volumeutil.PVCProtectionFinalizer)
	return pvc
}

func deleted(pvc *v1.PersistentVolumeClaim) *v1.PersistentVolumeClaim {
	pvc.DeletionTimestamp = &metav1.Time{}
	return pvc
}

func generateUpdateErrorFunc(t *testing.T, failures int) clienttesting.ReactionFunc {
	i := 0
	return func(action clienttesting.Action) (bool, runtime.Object, error) {
		i++
		if i <= failures {
			// Update fails
			update, ok := action.(clienttesting.UpdateAction)

			if !ok {
				t.Fatalf("Reactor got non-update action: %+v", action)
			}
			acc, _ := meta.Accessor(update.GetObject())
			return true, nil, apierrors.NewForbidden(update.GetResource().GroupResource(), acc.GetName(), errors.New("Mock error"))
		}
		// Update succeeds
		return false, nil, nil
	}
}

func TestPVCProtectionController(t *testing.T) {
	pvcVer := schema.GroupVersionResource{
		Group:    v1.GroupName,
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}

	tests := []struct {
		name string
		// Object to insert into fake kubeclient before the test starts.
		initialObjects []runtime.Object
		// Optional client reactors.
		reactors []reaction
		// PVC event to simulate. This PVC will be automatically added to
		// initalObjects.
		updatedPVC *v1.PersistentVolumeClaim
		// Pod event to simulate. This Pod will be automatically added to
		// initalObjects.
		updatedPod *v1.Pod
		// Pod event to similate. This Pod is *not* added to
		// initalObjects.
		deletedPod *v1.Pod
		// List of expected kubeclient actions that should happen during the
		// test.
		expectedActions                     []clienttesting.Action
		storageObjectInUseProtectionEnabled bool
	}{
		//
		// PVC events
		//
		{
			name:       "StorageObjectInUseProtection Enabled, PVC without finalizer -> finalizer is added",
			updatedPVC: pvc(),
			expectedActions: []clienttesting.Action{
				clienttesting.NewUpdateAction(pvcVer, defaultNS, withProtectionFinalizer(pvc())),
			},
			storageObjectInUseProtectionEnabled: true,
		},
		{
			name:                                "StorageObjectInUseProtection Disabled, PVC without finalizer -> finalizer is added",
			updatedPVC:                          pvc(),
			expectedActions:                     []clienttesting.Action{},
			storageObjectInUseProtectionEnabled: false,
		},
		{
			name:                                "PVC with finalizer -> no action",
			updatedPVC:                          withProtectionFinalizer(pvc()),
			expectedActions:                     []clienttesting.Action{},
			storageObjectInUseProtectionEnabled: true,
		},
		{
			name:       "saving PVC finalizer fails -> controller retries",
			updatedPVC: pvc(),
			reactors: []reaction{
				{
					verb:      "update",
					resource:  "persistentvolumeclaims",
					reactorfn: generateUpdateErrorFunc(t, 2 /* update fails twice*/),
				},
			},
			expectedActions: []clienttesting.Action{
				// This fails
				clienttesting.NewUpdateAction(pvcVer, defaultNS, withProtectionFinalizer(pvc())),
				// This fails too
				clienttesting.NewUpdateAction(pvcVer, defaultNS, withProtectionFinalizer(pvc())),
				// This succeeds
				clienttesting.NewUpdateAction(pvcVer, defaultNS, withProtectionFinalizer(pvc())),
			},
			storageObjectInUseProtectionEnabled: true,
		},
		{
			name:       "StorageObjectInUseProtection Enabled, deleted PVC with finalizer -> finalizer is removed",
			updatedPVC: deleted(withProtectionFinalizer(pvc())),
			expectedActions: []clienttesting.Action{
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
			},
			storageObjectInUseProtectionEnabled: true,
		},
		{
			name:       "StorageObjectInUseProtection Disabled, deleted PVC with finalizer -> finalizer is removed",
			updatedPVC: deleted(withProtectionFinalizer(pvc())),
			expectedActions: []clienttesting.Action{
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
			},
			storageObjectInUseProtectionEnabled: false,
		},
		{
			name:       "finalizer removal fails -> controller retries",
			updatedPVC: deleted(withProtectionFinalizer(pvc())),
			reactors: []reaction{
				{
					verb:      "update",
					resource:  "persistentvolumeclaims",
					reactorfn: generateUpdateErrorFunc(t, 2 /* update fails twice*/),
				},
			},
			expectedActions: []clienttesting.Action{
				// Fails
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
				// Fails too
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
				// Succeeds
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
			},
			storageObjectInUseProtectionEnabled: true,
		},
		{
			name: "deleted PVC with finalizer + pods with the PVC exists -> finalizer is not removed",
			initialObjects: []runtime.Object{
				withPVC(defaultPVCName, pod()),
			},
			updatedPVC:      deleted(withProtectionFinalizer(pvc())),
			expectedActions: []clienttesting.Action{},
		},
		{
			name: "deleted PVC with finalizer + pods with unrelated PVC and EmptyDir exists -> finalizer is removed",
			initialObjects: []runtime.Object{
				withEmptyDir(withPVC("unrelatedPVC", pod())),
			},
			updatedPVC: deleted(withProtectionFinalizer(pvc())),
			expectedActions: []clienttesting.Action{
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
			},
			storageObjectInUseProtectionEnabled: true,
		},
		{
			name: "deleted PVC with finalizer + pods with the PVC and is finished -> finalizer is removed",
			initialObjects: []runtime.Object{
				withStatus(v1.PodFailed, withPVC(defaultPVCName, pod())),
			},
			updatedPVC: deleted(withProtectionFinalizer(pvc())),
			expectedActions: []clienttesting.Action{
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
			},
			storageObjectInUseProtectionEnabled: true,
		},
		//
		// Pod events
		//
		{
			name: "updated running Pod -> no action",
			initialObjects: []runtime.Object{
				deleted(withProtectionFinalizer(pvc())),
			},
			updatedPod:                          withStatus(v1.PodRunning, withPVC(defaultPVCName, pod())),
			expectedActions:                     []clienttesting.Action{},
			storageObjectInUseProtectionEnabled: true,
		},
		{
			name: "updated finished Pod -> finalizer is removed",
			initialObjects: []runtime.Object{
				deleted(withProtectionFinalizer(pvc())),
			},
			updatedPod: withStatus(v1.PodSucceeded, withPVC(defaultPVCName, pod())),
			expectedActions: []clienttesting.Action{
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
			},
			storageObjectInUseProtectionEnabled: true,
		},
		{
			name: "updated unscheduled Pod -> finalizer is removed",
			initialObjects: []runtime.Object{
				deleted(withProtectionFinalizer(pvc())),
			},
			updatedPod: unscheduled(withPVC(defaultPVCName, pod())),
			expectedActions: []clienttesting.Action{
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
			},
			storageObjectInUseProtectionEnabled: true,
		},
		{
			name: "deleted running Pod -> finalizer is removed",
			initialObjects: []runtime.Object{
				deleted(withProtectionFinalizer(pvc())),
			},
			deletedPod: withStatus(v1.PodRunning, withPVC(defaultPVCName, pod())),
			expectedActions: []clienttesting.Action{
				clienttesting.NewUpdateAction(pvcVer, defaultNS, deleted(pvc())),
			},
			storageObjectInUseProtectionEnabled: true,
		},
	}

	for _, test := range tests {
		// Create client with initial data
		objs := test.initialObjects
		if test.updatedPVC != nil {
			objs = append(objs, test.updatedPVC)
		}
		if test.updatedPod != nil {
			objs = append(objs, test.updatedPod)
		}
		client := fake.NewSimpleClientset(objs...)

		// Create informers
		informers := informers.NewSharedInformerFactory(client, controller.NoResyncPeriodFunc())
		pvcInformer := informers.Core().V1().PersistentVolumeClaims()
		podInformer := informers.Core().V1().Pods()

		// Populate the informers with initial objects so the controller can
		// Get() and List() it.
		for _, obj := range objs {
			switch obj.(type) {
			case *v1.PersistentVolumeClaim:
				pvcInformer.Informer().GetStore().Add(obj)
			case *v1.Pod:
				podInformer.Informer().GetStore().Add(obj)
			default:
				t.Fatalf("Unknown initalObject type: %+v", obj)
			}
		}

		// Add reactor to inject test errors.
		for _, reactor := range test.reactors {
			client.Fake.PrependReactor(reactor.verb, reactor.resource, reactor.reactorfn)
		}

		// Create the controller
		ctrl := NewPVCProtectionController(pvcInformer, podInformer, client, test.storageObjectInUseProtectionEnabled)

		// Start the test by simulating an event
		if test.updatedPVC != nil {
			ctrl.pvcAddedUpdated(test.updatedPVC)
		}
		if test.updatedPod != nil {
			ctrl.podAddedDeletedUpdated(test.updatedPod, false)
		}
		if test.deletedPod != nil {
			ctrl.podAddedDeletedUpdated(test.deletedPod, true)
		}

		// Process the controller queue until we get expected results
		timeout := time.Now().Add(10 * time.Second)
		lastReportedActionCount := 0
		for {
			if time.Now().After(timeout) {
				t.Errorf("Test %q: timed out", test.name)
				break
			}
			if ctrl.queue.Len() > 0 {
				glog.V(5).Infof("Test %q: %d events queue, processing one", test.name, ctrl.queue.Len())
				ctrl.processNextWorkItem()
			}
			if ctrl.queue.Len() > 0 {
				// There is still some work in the queue, process it now
				continue
			}
			currentActionCount := len(client.Actions())
			if currentActionCount < len(test.expectedActions) {
				// Do not log evey wait, only when the action count changes.
				if lastReportedActionCount < currentActionCount {
					glog.V(5).Infof("Test %q: got %d actions out of %d, waiting for the rest", test.name, currentActionCount, len(test.expectedActions))
					lastReportedActionCount = currentActionCount
				}
				// The test expected more to happen, wait for the actions.
				// Most probably it's exponential backoff
				time.Sleep(10 * time.Millisecond)
				continue
			}
			break
		}
		actions := client.Actions()
		for i, action := range actions {
			if len(test.expectedActions) < i+1 {
				t.Errorf("Test %q: %d unexpected actions: %+v", test.name, len(actions)-len(test.expectedActions), spew.Sdump(actions[i:]))
				break
			}

			expectedAction := test.expectedActions[i]
			if !reflect.DeepEqual(expectedAction, action) {
				t.Errorf("Test %q: action %d\nExpected:\n%s\ngot:\n%s", test.name, i, spew.Sdump(expectedAction), spew.Sdump(action))
			}
		}

		if len(test.expectedActions) > len(actions) {
			t.Errorf("Test %q: %d additional expected actions", test.name, len(test.expectedActions)-len(actions))
			for _, a := range test.expectedActions[len(actions):] {
				t.Logf("    %+v", a)
			}
		}

	}
}
