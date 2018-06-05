/*
Copyright 2016 The Kubernetes Authors.

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

package operationexecutor

import (
	"strconv"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	expandcache "k8s.io/kubernetes/pkg/controller/volume/expand/cache"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/volume"
	volumetypes "k8s.io/kubernetes/pkg/volume/util/types"
)

const (
	numVolumesToMount                    = 2
	numAttachableVolumesToUnmount        = 2
	numNonAttachableVolumesToUnmount     = 2
	numDevicesToUnmount                  = 2
	numVolumesToAttach                   = 2
	numVolumesToDetach                   = 2
	numVolumesToVerifyAttached           = 2
	numVolumesToVerifyControllerAttached = 2
	numVolumesToMap                      = 2
	numAttachableVolumesToUnmap          = 2
	numNonAttachableVolumesToUnmap       = 2
	numDevicesToUnmap                    = 2
)

var _ OperationGenerator = &fakeOperationGenerator{}

func TestOperationExecutor_MountVolume_ConcurrentMountForNonAttachablePlugins(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	volumesToMount := make([]VolumeToMount, numVolumesToMount)
	secretName := "secret-volume"
	volumeName := v1.UniqueVolumeName(secretName)

	// Act
	for i := range volumesToMount {
		podName := "pod-" + strconv.Itoa((i + 1))
		pod := getTestPodWithSecret(podName, secretName)
		volumesToMount[i] = VolumeToMount{
			Pod:                pod,
			VolumeName:         volumeName,
			PluginIsAttachable: false, // this field determines whether the plugin is attachable
			ReportedInUse:      true,
		}
		oe.MountVolume(0 /* waitForAttachTimeOut */, volumesToMount[i], nil /* actualStateOfWorldMounterUpdater */, false /* isRemount */)
	}

	// Assert
	if !isOperationRunConcurrently(ch, quit, numVolumesToMount) {
		t.Fatalf("Unable to start mount operations in Concurrent for non-attachable volumes")
	}
}

func TestOperationExecutor_MountVolume_ConcurrentMountForAttachablePlugins(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	volumesToMount := make([]VolumeToMount, numVolumesToAttach)
	pdName := "pd-volume"
	volumeName := v1.UniqueVolumeName(pdName)
	// Act
	for i := range volumesToMount {
		podName := "pod-" + strconv.Itoa((i + 1))
		pod := getTestPodWithGCEPD(podName, pdName)
		volumesToMount[i] = VolumeToMount{
			Pod:                pod,
			VolumeName:         volumeName,
			PluginIsAttachable: true, // this field determines whether the plugin is attachable
			ReportedInUse:      true,
		}
		oe.MountVolume(0 /* waitForAttachTimeout */, volumesToMount[i], nil /* actualStateOfWorldMounterUpdater */, false /* isRemount */)
	}

	// Assert
	if !isOperationRunSerially(ch, quit) {
		t.Fatalf("Mount operations should not start concurrently for attachable volumes")
	}
}

func TestOperationExecutor_UnmountVolume_ConcurrentUnmountForAllPlugins(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	volumesToUnmount := make([]MountedVolume, numAttachableVolumesToUnmount+numNonAttachableVolumesToUnmount)
	pdName := "pd-volume"
	secretName := "secret-volume"

	// Act
	for i := 0; i < numNonAttachableVolumesToUnmount+numAttachableVolumesToUnmount; i++ {
		podName := "pod-" + strconv.Itoa(i+1)
		if i < numNonAttachableVolumesToUnmount {
			pod := getTestPodWithSecret(podName, secretName)
			volumesToUnmount[i] = MountedVolume{
				PodName:    volumetypes.UniquePodName(podName),
				VolumeName: v1.UniqueVolumeName(secretName),
				PodUID:     pod.UID,
			}
		} else {
			pod := getTestPodWithGCEPD(podName, pdName)
			volumesToUnmount[i] = MountedVolume{
				PodName:    volumetypes.UniquePodName(podName),
				VolumeName: v1.UniqueVolumeName(pdName),
				PodUID:     pod.UID,
			}
		}
		oe.UnmountVolume(volumesToUnmount[i], nil /* actualStateOfWorldMounterUpdater */, "" /*podsDir*/)
	}

	// Assert
	if !isOperationRunConcurrently(ch, quit, numNonAttachableVolumesToUnmount+numAttachableVolumesToUnmount) {
		t.Fatalf("Unable to start unmount operations concurrently for volume plugins")
	}
}

func TestOperationExecutor_UnmountDeviceConcurrently(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	attachedVolumes := make([]AttachedVolume, numDevicesToUnmount)
	pdName := "pd-volume"

	// Act
	for i := range attachedVolumes {
		attachedVolumes[i] = AttachedVolume{
			VolumeName: v1.UniqueVolumeName(pdName),
			NodeName:   "node-name",
		}
		oe.UnmountDevice(attachedVolumes[i], nil /* actualStateOfWorldMounterUpdater */, nil /* mount.Interface */)
	}

	// Assert
	if !isOperationRunSerially(ch, quit) {
		t.Fatalf("Unmount device operations should not start concurrently")
	}
}

func TestOperationExecutor_AttachVolumeConcurrently(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	volumesToAttach := make([]VolumeToAttach, numVolumesToAttach)
	pdName := "pd-volume"

	// Act
	for i := range volumesToAttach {
		volumesToAttach[i] = VolumeToAttach{
			VolumeName: v1.UniqueVolumeName(pdName),
			NodeName:   "node",
		}
		oe.AttachVolume(volumesToAttach[i], nil /* actualStateOfWorldAttacherUpdater */)
	}

	// Assert
	if !isOperationRunSerially(ch, quit) {
		t.Fatalf("Attach volume operations should not start concurrently")
	}
}

func TestOperationExecutor_DetachVolumeConcurrently(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	attachedVolumes := make([]AttachedVolume, numVolumesToDetach)
	pdName := "pd-volume"

	// Act
	for i := range attachedVolumes {
		attachedVolumes[i] = AttachedVolume{
			VolumeName: v1.UniqueVolumeName(pdName),
			NodeName:   "node",
		}
		oe.DetachVolume(attachedVolumes[i], true /* verifySafeToDetach */, nil /* actualStateOfWorldAttacherUpdater */)
	}

	// Assert
	if !isOperationRunSerially(ch, quit) {
		t.Fatalf("DetachVolume operations should not run concurrently")
	}
}

func TestOperationExecutor_VerifyVolumesAreAttachedConcurrently(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()

	// Act
	for i := 0; i < numVolumesToVerifyAttached; i++ {
		oe.VerifyVolumesAreAttachedPerNode(nil /* attachedVolumes */, "node-name", nil /* actualStateOfWorldAttacherUpdater */)
	}

	// Assert
	if !isOperationRunConcurrently(ch, quit, numVolumesToVerifyAttached) {
		t.Fatalf("VerifyVolumesAreAttached operation is not being run concurrently")
	}
}

func TestOperationExecutor_VerifyControllerAttachedVolumeConcurrently(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	volumesToMount := make([]VolumeToMount, numVolumesToVerifyControllerAttached)
	pdName := "pd-volume"

	// Act
	for i := range volumesToMount {
		volumesToMount[i] = VolumeToMount{
			VolumeName: v1.UniqueVolumeName(pdName),
		}
		oe.VerifyControllerAttachedVolume(volumesToMount[i], types.NodeName("node-name"), nil /* actualStateOfWorldMounterUpdater */)
	}

	// Assert
	if !isOperationRunSerially(ch, quit) {
		t.Fatalf("VerifyControllerAttachedVolume should not run concurrently")
	}
}

func TestOperationExecutor_MountVolume_ConcurrentMountForNonAttachablePlugins_VolumeMode_Block(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	volumesToMount := make([]VolumeToMount, numVolumesToMap)
	secretName := "secret-volume"
	volumeName := v1.UniqueVolumeName(secretName)
	volumeMode := v1.PersistentVolumeBlock
	tmpSpec := &volume.Spec{PersistentVolume: &v1.PersistentVolume{Spec: v1.PersistentVolumeSpec{VolumeMode: &volumeMode}}}

	// Act
	for i := range volumesToMount {
		podName := "pod-" + strconv.Itoa((i + 1))
		pod := getTestPodWithSecret(podName, secretName)
		volumesToMount[i] = VolumeToMount{
			Pod:                pod,
			VolumeName:         volumeName,
			PluginIsAttachable: false, // this field determines whether the plugin is attachable
			ReportedInUse:      true,
			VolumeSpec:         tmpSpec,
		}
		oe.MountVolume(0 /* waitForAttachTimeOut */, volumesToMount[i], nil /* actualStateOfWorldMounterUpdater */, false)
	}

	// Assert
	if !isOperationRunConcurrently(ch, quit, numVolumesToMap) {
		t.Fatalf("Unable to start map operations in Concurrent for non-attachable volumes")
	}
}

func TestOperationExecutor_MountVolume_ConcurrentMountForAttachablePlugins_VolumeMode_Block(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	volumesToMount := make([]VolumeToMount, numVolumesToAttach)
	pdName := "pd-volume"
	volumeName := v1.UniqueVolumeName(pdName)
	volumeMode := v1.PersistentVolumeBlock
	tmpSpec := &volume.Spec{PersistentVolume: &v1.PersistentVolume{Spec: v1.PersistentVolumeSpec{VolumeMode: &volumeMode}}}

	// Act
	for i := range volumesToMount {
		podName := "pod-" + strconv.Itoa((i + 1))
		pod := getTestPodWithGCEPD(podName, pdName)
		volumesToMount[i] = VolumeToMount{
			Pod:                pod,
			VolumeName:         volumeName,
			PluginIsAttachable: true, // this field determines whether the plugin is attachable
			ReportedInUse:      true,
			VolumeSpec:         tmpSpec,
		}
		oe.MountVolume(0 /* waitForAttachTimeout */, volumesToMount[i], nil /* actualStateOfWorldMounterUpdater */, false)
	}

	// Assert
	if !isOperationRunSerially(ch, quit) {
		t.Fatalf("Map operations should not start concurrently for attachable volumes")
	}
}

func TestOperationExecutor_UnmountVolume_ConcurrentUnmountForAllPlugins_VolumeMode_Block(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	volumesToUnmount := make([]MountedVolume, numAttachableVolumesToUnmap+numNonAttachableVolumesToUnmap)
	pdName := "pd-volume"
	secretName := "secret-volume"
	volumeMode := v1.PersistentVolumeBlock
	tmpSpec := &volume.Spec{PersistentVolume: &v1.PersistentVolume{Spec: v1.PersistentVolumeSpec{VolumeMode: &volumeMode}}}

	// Act
	for i := 0; i < numNonAttachableVolumesToUnmap+numAttachableVolumesToUnmap; i++ {
		podName := "pod-" + strconv.Itoa(i+1)
		if i < numNonAttachableVolumesToUnmap {
			pod := getTestPodWithSecret(podName, secretName)
			volumesToUnmount[i] = MountedVolume{
				PodName:    volumetypes.UniquePodName(podName),
				VolumeName: v1.UniqueVolumeName(secretName),
				PodUID:     pod.UID,
				VolumeSpec: tmpSpec,
			}
		} else {
			pod := getTestPodWithGCEPD(podName, pdName)
			volumesToUnmount[i] = MountedVolume{
				PodName:    volumetypes.UniquePodName(podName),
				VolumeName: v1.UniqueVolumeName(pdName),
				PodUID:     pod.UID,
				VolumeSpec: tmpSpec,
			}
		}
		oe.UnmountVolume(volumesToUnmount[i], nil /* actualStateOfWorldMounterUpdater */, "" /* podsDir */)
	}

	// Assert
	if !isOperationRunConcurrently(ch, quit, numNonAttachableVolumesToUnmap+numAttachableVolumesToUnmap) {
		t.Fatalf("Unable to start unmap operations concurrently for volume plugins")
	}
}

func TestOperationExecutor_UnmountDeviceConcurrently_VolumeMode_Block(t *testing.T) {
	// Arrange
	ch, quit, oe := setup()
	attachedVolumes := make([]AttachedVolume, numDevicesToUnmap)
	pdName := "pd-volume"
	volumeMode := v1.PersistentVolumeBlock
	tmpSpec := &volume.Spec{PersistentVolume: &v1.PersistentVolume{Spec: v1.PersistentVolumeSpec{VolumeMode: &volumeMode}}}

	// Act
	for i := range attachedVolumes {
		attachedVolumes[i] = AttachedVolume{
			VolumeName: v1.UniqueVolumeName(pdName),
			NodeName:   "node-name",
			VolumeSpec: tmpSpec,
		}
		oe.UnmountDevice(attachedVolumes[i], nil /* actualStateOfWorldMounterUpdater */, nil /* mount.Interface */)
	}

	// Assert
	if !isOperationRunSerially(ch, quit) {
		t.Fatalf("Unmap device operations should not start concurrently")
	}
}

type fakeOperationGenerator struct {
	ch   chan interface{}
	quit chan interface{}
}

func newFakeOperationGenerator(ch chan interface{}, quit chan interface{}) OperationGenerator {
	return &fakeOperationGenerator{
		ch:   ch,
		quit: quit,
	}
}

func (fopg *fakeOperationGenerator) GenerateMountVolumeFunc(waitForAttachTimeout time.Duration, volumeToMount VolumeToMount, actualStateOfWorldMounterUpdater ActualStateOfWorldMounterUpdater, isRemount bool) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}
func (fopg *fakeOperationGenerator) GenerateUnmountVolumeFunc(volumeToUnmount MountedVolume, actualStateOfWorld ActualStateOfWorldMounterUpdater, podsDir string) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}
func (fopg *fakeOperationGenerator) GenerateAttachVolumeFunc(volumeToAttach VolumeToAttach, actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}
func (fopg *fakeOperationGenerator) GenerateDetachVolumeFunc(volumeToDetach AttachedVolume, verifySafeToDetach bool, actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}
func (fopg *fakeOperationGenerator) GenerateVolumesAreAttachedFunc(attachedVolumes []AttachedVolume, nodeName types.NodeName, actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}
func (fopg *fakeOperationGenerator) GenerateUnmountDeviceFunc(deviceToDetach AttachedVolume, actualStateOfWorld ActualStateOfWorldMounterUpdater, mounter mount.Interface) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}
func (fopg *fakeOperationGenerator) GenerateVerifyControllerAttachedVolumeFunc(volumeToMount VolumeToMount, nodeName types.NodeName, actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}

func (fopg *fakeOperationGenerator) GenerateExpandVolumeFunc(pvcWithResizeRequest *expandcache.PVCWithResizeRequest,
	resizeMap expandcache.VolumeResizeMap) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}

func (fopg *fakeOperationGenerator) GenerateBulkVolumeVerifyFunc(
	pluginNodeVolumes map[types.NodeName][]*volume.Spec,
	pluginNane string,
	volumeSpecMap map[*volume.Spec]v1.UniqueVolumeName,
	actualStateOfWorldAttacherUpdater ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}

func (fopg *fakeOperationGenerator) GenerateMapVolumeFunc(waitForAttachTimeout time.Duration, volumeToMount VolumeToMount, actualStateOfWorldMounterUpdater ActualStateOfWorldMounterUpdater) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}

func (fopg *fakeOperationGenerator) GenerateUnmapVolumeFunc(volumeToUnmount MountedVolume, actualStateOfWorld ActualStateOfWorldMounterUpdater) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}

func (fopg *fakeOperationGenerator) GenerateUnmapDeviceFunc(deviceToDetach AttachedVolume, actualStateOfWorld ActualStateOfWorldMounterUpdater, mounter mount.Interface) (volumetypes.GeneratedOperations, error) {
	opFunc := func() (error, error) {
		startOperationAndBlock(fopg.ch, fopg.quit)
		return nil, nil
	}
	return volumetypes.GeneratedOperations{
		OperationFunc: opFunc,
	}, nil
}

func (fopg *fakeOperationGenerator) GetVolumePluginMgr() *volume.VolumePluginMgr {
	return nil
}

func getTestPodWithSecret(podName, secretName string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			UID:  types.UID(podName),
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: secretName,
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name:  "secret-volume-test",
					Image: "k8s.gcr.io/mounttest:0.8",
					Args: []string{
						"--file_content=/etc/secret-volume/data-1",
						"--file_mode=/etc/secret-volume/data-1"},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      secretName,
							MountPath: "/data",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
}

func getTestPodWithGCEPD(podName, pdName string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			UID:  types.UID(podName + string(uuid.NewUUID())),
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: pdName,
					VolumeSource: v1.VolumeSource{
						GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
							PDName:   pdName,
							FSType:   "ext4",
							ReadOnly: false,
						},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name:  "pd-volume-test",
					Image: "k8s.gcr.io/mounttest:0.8",
					Args: []string{
						"--file_content=/etc/pd-volume/data-1",
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      pdName,
							MountPath: "/data",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
}

func isOperationRunSerially(ch <-chan interface{}, quit chan<- interface{}) bool {
	defer close(quit)
	numOperationsStarted := 0
loop:
	for {
		select {
		case <-ch:
			numOperationsStarted++
			if numOperationsStarted > 1 {
				return false
			}
		case <-time.After(5 * time.Second):
			break loop
		}
	}
	return true
}

func isOperationRunConcurrently(ch <-chan interface{}, quit chan<- interface{}, numOperationsToRun int) bool {
	defer close(quit)
	numOperationsStarted := 0
loop:
	for {
		select {
		case <-ch:
			numOperationsStarted++
			if numOperationsStarted == numOperationsToRun {
				return true
			}
		case <-time.After(5 * time.Second):
			break loop
		}
	}
	return false
}

func setup() (chan interface{}, chan interface{}, OperationExecutor) {
	ch, quit := make(chan interface{}), make(chan interface{})
	return ch, quit, NewOperationExecutor(newFakeOperationGenerator(ch, quit))
}

// This function starts by writing to ch and blocks on the quit channel
// until it is closed by the currently running test
func startOperationAndBlock(ch chan<- interface{}, quit <-chan interface{}) {
	ch <- nil
	<-quit
}
