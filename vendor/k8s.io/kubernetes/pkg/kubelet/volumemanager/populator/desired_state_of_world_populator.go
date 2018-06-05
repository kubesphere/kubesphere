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

/*
Package populator implements interfaces that monitor and keep the states of the
caches in sync with the "ground truth".
*/
package populator

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/kubelet/config"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/kubelet/pod"
	"k8s.io/kubernetes/pkg/kubelet/status"
	"k8s.io/kubernetes/pkg/kubelet/util/format"
	"k8s.io/kubernetes/pkg/kubelet/volumemanager/cache"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"
	volumetypes "k8s.io/kubernetes/pkg/volume/util/types"
)

// DesiredStateOfWorldPopulator periodically loops through the list of active
// pods and ensures that each one exists in the desired state of the world cache
// if it has volumes. It also verifies that the pods in the desired state of the
// world cache still exist, if not, it removes them.
type DesiredStateOfWorldPopulator interface {
	Run(sourcesReady config.SourcesReady, stopCh <-chan struct{})

	// ReprocessPod removes the specified pod from the list of processedPods
	// (if it exists) forcing it to be reprocessed. This is required to enable
	// remounting volumes on pod updates (volumes like Downward API volumes
	// depend on this behavior to ensure volume content is updated).
	ReprocessPod(podName volumetypes.UniquePodName)

	// HasAddedPods returns whether the populator has looped through the list
	// of active pods and added them to the desired state of the world cache,
	// at a time after sources are all ready, at least once. It does not
	// return true before sources are all ready because before then, there is
	// a chance many or all pods are missing from the list of active pods and
	// so few to none will have been added.
	HasAddedPods() bool
}

// NewDesiredStateOfWorldPopulator returns a new instance of
// DesiredStateOfWorldPopulator.
//
// kubeClient - used to fetch PV and PVC objects from the API server
// loopSleepDuration - the amount of time the populator loop sleeps between
//     successive executions
// podManager - the kubelet podManager that is the source of truth for the pods
//     that exist on this host
// desiredStateOfWorld - the cache to populate
func NewDesiredStateOfWorldPopulator(
	kubeClient clientset.Interface,
	loopSleepDuration time.Duration,
	getPodStatusRetryDuration time.Duration,
	podManager pod.Manager,
	podStatusProvider status.PodStatusProvider,
	desiredStateOfWorld cache.DesiredStateOfWorld,
	actualStateOfWorld cache.ActualStateOfWorld,
	kubeContainerRuntime kubecontainer.Runtime,
	keepTerminatedPodVolumes bool) DesiredStateOfWorldPopulator {
	return &desiredStateOfWorldPopulator{
		kubeClient:                kubeClient,
		loopSleepDuration:         loopSleepDuration,
		getPodStatusRetryDuration: getPodStatusRetryDuration,
		podManager:                podManager,
		podStatusProvider:         podStatusProvider,
		desiredStateOfWorld:       desiredStateOfWorld,
		actualStateOfWorld:        actualStateOfWorld,
		pods: processedPods{
			processedPods: make(map[volumetypes.UniquePodName]bool)},
		kubeContainerRuntime:     kubeContainerRuntime,
		keepTerminatedPodVolumes: keepTerminatedPodVolumes,
		hasAddedPods:             false,
		hasAddedPodsLock:         sync.RWMutex{},
	}
}

type desiredStateOfWorldPopulator struct {
	kubeClient                clientset.Interface
	loopSleepDuration         time.Duration
	getPodStatusRetryDuration time.Duration
	podManager                pod.Manager
	podStatusProvider         status.PodStatusProvider
	desiredStateOfWorld       cache.DesiredStateOfWorld
	actualStateOfWorld        cache.ActualStateOfWorld
	pods                      processedPods
	kubeContainerRuntime      kubecontainer.Runtime
	timeOfLastGetPodStatus    time.Time
	keepTerminatedPodVolumes  bool
	hasAddedPods              bool
	hasAddedPodsLock          sync.RWMutex
}

type processedPods struct {
	processedPods map[volumetypes.UniquePodName]bool
	sync.RWMutex
}

func (dswp *desiredStateOfWorldPopulator) Run(sourcesReady config.SourcesReady, stopCh <-chan struct{}) {
	// Wait for the completion of a loop that started after sources are all ready, then set hasAddedPods accordingly
	glog.Infof("Desired state populator starts to run")
	wait.PollUntil(dswp.loopSleepDuration, func() (bool, error) {
		done := sourcesReady.AllReady()
		dswp.populatorLoopFunc()()
		return done, nil
	}, stopCh)
	dswp.hasAddedPodsLock.Lock()
	dswp.hasAddedPods = true
	dswp.hasAddedPodsLock.Unlock()
	wait.Until(dswp.populatorLoopFunc(), dswp.loopSleepDuration, stopCh)
}

func (dswp *desiredStateOfWorldPopulator) ReprocessPod(
	podName volumetypes.UniquePodName) {
	dswp.deleteProcessedPod(podName)
}

func (dswp *desiredStateOfWorldPopulator) HasAddedPods() bool {
	dswp.hasAddedPodsLock.RLock()
	defer dswp.hasAddedPodsLock.RUnlock()
	return dswp.hasAddedPods
}

func (dswp *desiredStateOfWorldPopulator) populatorLoopFunc() func() {
	return func() {
		dswp.findAndAddNewPods()

		// findAndRemoveDeletedPods() calls out to the container runtime to
		// determine if the containers for a given pod are terminated. This is
		// an expensive operation, therefore we limit the rate that
		// findAndRemoveDeletedPods() is called independently of the main
		// populator loop.
		if time.Since(dswp.timeOfLastGetPodStatus) < dswp.getPodStatusRetryDuration {
			glog.V(5).Infof(
				"Skipping findAndRemoveDeletedPods(). Not permitted until %v (getPodStatusRetryDuration %v).",
				dswp.timeOfLastGetPodStatus.Add(dswp.getPodStatusRetryDuration),
				dswp.getPodStatusRetryDuration)

			return
		}

		dswp.findAndRemoveDeletedPods()
	}
}

func (dswp *desiredStateOfWorldPopulator) isPodTerminated(pod *v1.Pod) bool {
	podStatus, found := dswp.podStatusProvider.GetPodStatus(pod.UID)
	if !found {
		podStatus = pod.Status
	}
	return util.IsPodTerminated(pod, podStatus)
}

// Iterate through all pods and add to desired state of world if they don't
// exist but should
func (dswp *desiredStateOfWorldPopulator) findAndAddNewPods() {
	for _, pod := range dswp.podManager.GetPods() {
		if dswp.isPodTerminated(pod) {
			// Do not (re)add volumes for terminated pods
			continue
		}
		dswp.processPodVolumes(pod)
	}
}

// Iterate through all pods in desired state of world, and remove if they no
// longer exist
func (dswp *desiredStateOfWorldPopulator) findAndRemoveDeletedPods() {
	var runningPods []*kubecontainer.Pod

	runningPodsFetched := false
	for _, volumeToMount := range dswp.desiredStateOfWorld.GetVolumesToMount() {
		pod, podExists := dswp.podManager.GetPodByUID(volumeToMount.Pod.UID)
		if podExists {
			// Skip running pods
			if !dswp.isPodTerminated(pod) {
				continue
			}
			if dswp.keepTerminatedPodVolumes {
				continue
			}
		}

		// Once a pod has been deleted from kubelet pod manager, do not delete
		// it immediately from volume manager. Instead, check the kubelet
		// containerRuntime to verify that all containers in the pod have been
		// terminated.
		if !runningPodsFetched {
			var getPodsErr error
			runningPods, getPodsErr = dswp.kubeContainerRuntime.GetPods(false)
			if getPodsErr != nil {
				glog.Errorf(
					"kubeContainerRuntime.findAndRemoveDeletedPods returned error %v.",
					getPodsErr)
				continue
			}

			runningPodsFetched = true
			dswp.timeOfLastGetPodStatus = time.Now()
		}

		runningContainers := false
		for _, runningPod := range runningPods {
			if runningPod.ID == volumeToMount.Pod.UID {
				if len(runningPod.Containers) > 0 {
					runningContainers = true
				}

				break
			}
		}

		if runningContainers {
			glog.V(4).Infof(
				"Pod %q has been removed from pod manager. However, it still has one or more containers in the non-exited state. Therefore, it will not be removed from volume manager.",
				format.Pod(volumeToMount.Pod))
			continue
		}

		if !dswp.actualStateOfWorld.VolumeExists(volumeToMount.VolumeName) && podExists {
			glog.V(4).Infof(volumeToMount.GenerateMsgDetailed("Actual state has not yet has this information skip removing volume from desired state", ""))
			continue
		}
		glog.V(4).Infof(volumeToMount.GenerateMsgDetailed("Removing volume from desired state", ""))

		dswp.desiredStateOfWorld.DeletePodFromVolume(
			volumeToMount.PodName, volumeToMount.VolumeName)
		dswp.deleteProcessedPod(volumeToMount.PodName)
	}
}

// processPodVolumes processes the volumes in the given pod and adds them to the
// desired state of the world.
func (dswp *desiredStateOfWorldPopulator) processPodVolumes(pod *v1.Pod) {
	if pod == nil {
		return
	}

	uniquePodName := util.GetUniquePodName(pod)
	if dswp.podPreviouslyProcessed(uniquePodName) {
		return
	}

	allVolumesAdded := true
	mountsMap, devicesMap := dswp.makeVolumeMap(pod.Spec.Containers)

	// Process volume spec for each volume defined in pod
	for _, podVolume := range pod.Spec.Volumes {
		volumeSpec, volumeGidValue, err :=
			dswp.createVolumeSpec(podVolume, pod.Name, pod.Namespace, mountsMap, devicesMap)
		if err != nil {
			glog.Errorf(
				"Error processing volume %q for pod %q: %v",
				podVolume.Name,
				format.Pod(pod),
				err)
			allVolumesAdded = false
			continue
		}

		// Add volume to desired state of world
		_, err = dswp.desiredStateOfWorld.AddPodToVolume(
			uniquePodName, pod, volumeSpec, podVolume.Name, volumeGidValue)
		if err != nil {
			glog.Errorf(
				"Failed to add volume %q (specName: %q) for pod %q to desiredStateOfWorld. err=%v",
				podVolume.Name,
				volumeSpec.Name(),
				uniquePodName,
				err)
			allVolumesAdded = false
		}

		glog.V(4).Infof(
			"Added volume %q (volSpec=%q) for pod %q to desired state.",
			podVolume.Name,
			volumeSpec.Name(),
			uniquePodName)
	}

	// some of the volume additions may have failed, should not mark this pod as fully processed
	if allVolumesAdded {
		dswp.markPodProcessed(uniquePodName)
		// New pod has been synced. Re-mount all volumes that need it
		// (e.g. DownwardAPI)
		dswp.actualStateOfWorld.MarkRemountRequired(uniquePodName)
	}

}

// podPreviouslyProcessed returns true if the volumes for this pod have already
// been processed by the populator
func (dswp *desiredStateOfWorldPopulator) podPreviouslyProcessed(
	podName volumetypes.UniquePodName) bool {
	dswp.pods.RLock()
	defer dswp.pods.RUnlock()

	_, exists := dswp.pods.processedPods[podName]
	return exists
}

// markPodProcessed records that the volumes for the specified pod have been
// processed by the populator
func (dswp *desiredStateOfWorldPopulator) markPodProcessed(
	podName volumetypes.UniquePodName) {
	dswp.pods.Lock()
	defer dswp.pods.Unlock()

	dswp.pods.processedPods[podName] = true
}

// markPodProcessed removes the specified pod from processedPods
func (dswp *desiredStateOfWorldPopulator) deleteProcessedPod(
	podName volumetypes.UniquePodName) {
	dswp.pods.Lock()
	defer dswp.pods.Unlock()

	delete(dswp.pods.processedPods, podName)
}

// createVolumeSpec creates and returns a mutatable volume.Spec object for the
// specified volume. It dereference any PVC to get PV objects, if needed.
// Returns an error if unable to obtain the volume at this time.
func (dswp *desiredStateOfWorldPopulator) createVolumeSpec(
	podVolume v1.Volume, podName string, podNamespace string, mountsMap map[string]bool, devicesMap map[string]bool) (*volume.Spec, string, error) {
	if pvcSource :=
		podVolume.VolumeSource.PersistentVolumeClaim; pvcSource != nil {
		glog.V(5).Infof(
			"Found PVC, ClaimName: %q/%q",
			podNamespace,
			pvcSource.ClaimName)

		// If podVolume is a PVC, fetch the real PV behind the claim
		pvName, pvcUID, err := dswp.getPVCExtractPV(
			podNamespace, pvcSource.ClaimName)
		if err != nil {
			return nil, "", fmt.Errorf(
				"error processing PVC %q/%q: %v",
				podNamespace,
				pvcSource.ClaimName,
				err)
		}

		glog.V(5).Infof(
			"Found bound PV for PVC (ClaimName %q/%q pvcUID %v): pvName=%q",
			podNamespace,
			pvcSource.ClaimName,
			pvcUID,
			pvName)

		// Fetch actual PV object
		volumeSpec, volumeGidValue, err :=
			dswp.getPVSpec(pvName, pvcSource.ReadOnly, pvcUID)
		if err != nil {
			return nil, "", fmt.Errorf(
				"error processing PVC %q/%q: %v",
				podNamespace,
				pvcSource.ClaimName,
				err)
		}

		glog.V(5).Infof(
			"Extracted volumeSpec (%v) from bound PV (pvName %q) and PVC (ClaimName %q/%q pvcUID %v)",
			volumeSpec.Name,
			pvName,
			podNamespace,
			pvcSource.ClaimName,
			pvcUID)

		// TODO: remove feature gate check after no longer needed
		if utilfeature.DefaultFeatureGate.Enabled(features.BlockVolume) {
			volumeMode, err := util.GetVolumeMode(volumeSpec)
			if err != nil {
				return nil, "", err
			}
			// Error if a container has volumeMounts but the volumeMode of PVC isn't Filesystem
			if mountsMap[podVolume.Name] && volumeMode != v1.PersistentVolumeFilesystem {
				return nil, "", fmt.Errorf(
					"Volume %q has volumeMode %q, but is specified in volumeMounts for pod %q/%q",
					podVolume.Name,
					volumeMode,
					podNamespace,
					podName)
			}
			// Error if a container has volumeDevices but the volumeMode of PVC isn't Block
			if devicesMap[podVolume.Name] && volumeMode != v1.PersistentVolumeBlock {
				return nil, "", fmt.Errorf(
					"Volume %q has volumeMode %q, but is specified in volumeDevices for pod %q/%q",
					podVolume.Name,
					volumeMode,
					podNamespace,
					podName)
			}
		}
		return volumeSpec, volumeGidValue, nil
	}

	// Do not return the original volume object, since the source could mutate it
	clonedPodVolume := podVolume.DeepCopy()

	return volume.NewSpecFromVolume(clonedPodVolume), "", nil
}

// getPVCExtractPV fetches the PVC object with the given namespace and name from
// the API server, checks whether PVC is being deleted, extracts the name of the PV
// it is pointing to and returns it.
// An error is returned if the PVC object's phase is not "Bound".
func (dswp *desiredStateOfWorldPopulator) getPVCExtractPV(
	namespace string, claimName string) (string, types.UID, error) {
	pvc, err :=
		dswp.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(claimName, metav1.GetOptions{})
	if err != nil || pvc == nil {
		return "", "", fmt.Errorf(
			"failed to fetch PVC %s/%s from API server. err=%v",
			namespace,
			claimName,
			err)
	}

	if utilfeature.DefaultFeatureGate.Enabled(features.StorageObjectInUseProtection) {
		// Pods that uses a PVC that is being deleted must not be started.
		//
		// In case an old kubelet is running without this check or some kubelets
		// have this feature disabled, the worst that can happen is that such
		// pod is scheduled. This was the default behavior in 1.8 and earlier
		// and users should not be that surprised.
		// It should happen only in very rare case when scheduler schedules
		// a pod and user deletes a PVC that's used by it at the same time.
		if pvc.ObjectMeta.DeletionTimestamp != nil {
			return "", "", fmt.Errorf(
				"can't start pod because PVC %s/%s is being deleted",
				namespace,
				claimName)
		}
	}

	if pvc.Status.Phase != v1.ClaimBound || pvc.Spec.VolumeName == "" {

		return "", "", fmt.Errorf(
			"PVC %s/%s has non-bound phase (%q) or empty pvc.Spec.VolumeName (%q)",
			namespace,
			claimName,
			pvc.Status.Phase,
			pvc.Spec.VolumeName)
	}

	return pvc.Spec.VolumeName, pvc.UID, nil
}

// getPVSpec fetches the PV object with the given name from the API server
// and returns a volume.Spec representing it.
// An error is returned if the call to fetch the PV object fails.
func (dswp *desiredStateOfWorldPopulator) getPVSpec(
	name string,
	pvcReadOnly bool,
	expectedClaimUID types.UID) (*volume.Spec, string, error) {
	pv, err := dswp.kubeClient.CoreV1().PersistentVolumes().Get(name, metav1.GetOptions{})
	if err != nil || pv == nil {
		return nil, "", fmt.Errorf(
			"failed to fetch PV %q from API server. err=%v", name, err)
	}

	if pv.Spec.ClaimRef == nil {
		return nil, "", fmt.Errorf(
			"found PV object %q but it has a nil pv.Spec.ClaimRef indicating it is not yet bound to the claim",
			name)
	}

	if pv.Spec.ClaimRef.UID != expectedClaimUID {
		return nil, "", fmt.Errorf(
			"found PV object %q but its pv.Spec.ClaimRef.UID (%q) does not point to claim.UID (%q)",
			name,
			pv.Spec.ClaimRef.UID,
			expectedClaimUID)
	}

	volumeGidValue := getPVVolumeGidAnnotationValue(pv)
	return volume.NewSpecFromPersistentVolume(pv, pvcReadOnly), volumeGidValue, nil
}

func (dswp *desiredStateOfWorldPopulator) makeVolumeMap(containers []v1.Container) (map[string]bool, map[string]bool) {
	volumeDevicesMap := make(map[string]bool)
	volumeMountsMap := make(map[string]bool)

	for _, container := range containers {
		if container.VolumeMounts != nil {
			for _, mount := range container.VolumeMounts {
				volumeMountsMap[mount.Name] = true
			}
		}
		// TODO: remove feature gate check after no longer needed
		if utilfeature.DefaultFeatureGate.Enabled(features.BlockVolume) &&
			container.VolumeDevices != nil {
			for _, device := range container.VolumeDevices {
				volumeDevicesMap[device.Name] = true
			}
		}
	}

	return volumeMountsMap, volumeDevicesMap
}

func getPVVolumeGidAnnotationValue(pv *v1.PersistentVolume) string {
	if volumeGid, ok := pv.Annotations[util.VolumeGidAnnotationKey]; ok {
		return volumeGid
	}

	return ""
}
