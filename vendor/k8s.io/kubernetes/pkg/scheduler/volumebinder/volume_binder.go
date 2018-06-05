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

package volumebinder

import (
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	storageinformers "k8s.io/client-go/informers/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/controller/volume/persistentvolume"
)

// VolumeBinder sets up the volume binding library and manages
// the volume binding operations with a queue.
type VolumeBinder struct {
	Binder    persistentvolume.SchedulerVolumeBinder
	BindQueue *workqueue.Type
}

// NewVolumeBinder sets up the volume binding library and binding queue
func NewVolumeBinder(
	client clientset.Interface,
	pvcInformer coreinformers.PersistentVolumeClaimInformer,
	pvInformer coreinformers.PersistentVolumeInformer,
	storageClassInformer storageinformers.StorageClassInformer) *VolumeBinder {

	return &VolumeBinder{
		Binder:    persistentvolume.NewVolumeBinder(client, pvcInformer, pvInformer, storageClassInformer),
		BindQueue: workqueue.NewNamed("podsToBind"),
	}
}

// NewFakeVolumeBinder sets up a fake volume binder and binding queue
func NewFakeVolumeBinder(config *persistentvolume.FakeVolumeBinderConfig) *VolumeBinder {
	return &VolumeBinder{
		Binder:    persistentvolume.NewFakeVolumeBinder(config),
		BindQueue: workqueue.NewNamed("podsToBind"),
	}
}

// Run starts a goroutine to handle the binding queue with the given function.
func (b *VolumeBinder) Run(bindWorkFunc func(), stopCh <-chan struct{}) {
	go wait.Until(bindWorkFunc, time.Second, stopCh)

	<-stopCh
	b.BindQueue.ShutDown()
}

// DeletePodBindings will delete the cached volume bindings for the given pod.
func (b *VolumeBinder) DeletePodBindings(pod *v1.Pod) {
	cache := b.Binder.GetBindingsCache()
	if cache != nil && pod != nil {
		cache.DeleteBindings(pod)
	}
}
