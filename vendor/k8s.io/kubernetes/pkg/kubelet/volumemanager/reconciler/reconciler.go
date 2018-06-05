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

// Package reconciler implements interfaces that attempt to reconcile the
// desired state of the world with the actual state of the world by triggering
// relevant actions (attach, detach, mount, unmount).
package reconciler

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
	"k8s.io/kubernetes/pkg/kubelet/volumemanager/cache"
	utilfile "k8s.io/kubernetes/pkg/util/file"
	"k8s.io/kubernetes/pkg/util/goroutinemap/exponentialbackoff"
	"k8s.io/kubernetes/pkg/util/mount"
	utilstrings "k8s.io/kubernetes/pkg/util/strings"
	volumepkg "k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"
	"k8s.io/kubernetes/pkg/volume/util/nestedpendingoperations"
	"k8s.io/kubernetes/pkg/volume/util/operationexecutor"
	volumetypes "k8s.io/kubernetes/pkg/volume/util/types"
)

// Reconciler runs a periodic loop to reconcile the desired state of the world
// with the actual state of the world by triggering attach, detach, mount, and
// unmount operations.
// Note: This is distinct from the Reconciler implemented by the attach/detach
// controller. This reconciles state for the kubelet volume manager. That
// reconciles state for the attach/detach controller.
type Reconciler interface {
	// Starts running the reconciliation loop which executes periodically, checks
	// if volumes that should be mounted are mounted and volumes that should
	// be unmounted are unmounted. If not, it will trigger mount/unmount
	// operations to rectify.
	// If attach/detach management is enabled, the manager will also check if
	// volumes that should be attached are attached and volumes that should
	// be detached are detached and trigger attach/detach operations as needed.
	Run(stopCh <-chan struct{})

	// StatesHasBeenSynced returns true only after syncStates process starts to sync
	// states at least once after kubelet starts
	StatesHasBeenSynced() bool
}

// NewReconciler returns a new instance of Reconciler.
//
// controllerAttachDetachEnabled - if true, indicates that the attach/detach
//   controller is responsible for managing the attach/detach operations for
//   this node, and therefore the volume manager should not
// loopSleepDuration - the amount of time the reconciler loop sleeps between
//   successive executions
//   syncDuration - the amount of time the syncStates sleeps between
//   successive executions
// waitForAttachTimeout - the amount of time the Mount function will wait for
//   the volume to be attached
// nodeName - the Name for this node, used by Attach and Detach methods
// desiredStateOfWorld - cache containing the desired state of the world
// actualStateOfWorld - cache containing the actual state of the world
// populatorHasAddedPods - checker for whether the populator has finished
//   adding pods to the desiredStateOfWorld cache at least once after sources
//   are all ready (before sources are ready, pods are probably missing)
// operationExecutor - used to trigger attach/detach/mount/unmount operations
//   safely (prevents more than one operation from being triggered on the same
//   volume)
// mounter - mounter passed in from kubelet, passed down unmount path
// volumePluginMrg - volume plugin manager passed from kubelet
func NewReconciler(
	kubeClient clientset.Interface,
	controllerAttachDetachEnabled bool,
	loopSleepDuration time.Duration,
	syncDuration time.Duration,
	waitForAttachTimeout time.Duration,
	nodeName types.NodeName,
	desiredStateOfWorld cache.DesiredStateOfWorld,
	actualStateOfWorld cache.ActualStateOfWorld,
	populatorHasAddedPods func() bool,
	operationExecutor operationexecutor.OperationExecutor,
	mounter mount.Interface,
	volumePluginMgr *volumepkg.VolumePluginMgr,
	kubeletPodsDir string) Reconciler {
	return &reconciler{
		kubeClient:                    kubeClient,
		controllerAttachDetachEnabled: controllerAttachDetachEnabled,
		loopSleepDuration:             loopSleepDuration,
		syncDuration:                  syncDuration,
		waitForAttachTimeout:          waitForAttachTimeout,
		nodeName:                      nodeName,
		desiredStateOfWorld:           desiredStateOfWorld,
		actualStateOfWorld:            actualStateOfWorld,
		populatorHasAddedPods:         populatorHasAddedPods,
		operationExecutor:             operationExecutor,
		mounter:                       mounter,
		volumePluginMgr:               volumePluginMgr,
		kubeletPodsDir:                kubeletPodsDir,
		timeOfLastSync:                time.Time{},
	}
}

type reconciler struct {
	kubeClient                    clientset.Interface
	controllerAttachDetachEnabled bool
	loopSleepDuration             time.Duration
	syncDuration                  time.Duration
	waitForAttachTimeout          time.Duration
	nodeName                      types.NodeName
	desiredStateOfWorld           cache.DesiredStateOfWorld
	actualStateOfWorld            cache.ActualStateOfWorld
	populatorHasAddedPods         func() bool
	operationExecutor             operationexecutor.OperationExecutor
	mounter                       mount.Interface
	volumePluginMgr               *volumepkg.VolumePluginMgr
	kubeletPodsDir                string
	timeOfLastSync                time.Time
}

func (rc *reconciler) Run(stopCh <-chan struct{}) {
	wait.Until(rc.reconciliationLoopFunc(), rc.loopSleepDuration, stopCh)
}

func (rc *reconciler) reconciliationLoopFunc() func() {
	return func() {
		rc.reconcile()

		// Sync the state with the reality once after all existing pods are added to the desired state from all sources.
		// Otherwise, the reconstruct process may clean up pods' volumes that are still in use because
		// desired state of world does not contain a complete list of pods.
		if rc.populatorHasAddedPods() && !rc.StatesHasBeenSynced() {
			glog.Infof("Reconciler: start to sync state")
			rc.sync()
		}
	}
}

func (rc *reconciler) reconcile() {
	// Unmounts are triggered before mounts so that a volume that was
	// referenced by a pod that was deleted and is now referenced by another
	// pod is unmounted from the first pod before being mounted to the new
	// pod.

	// Ensure volumes that should be unmounted are unmounted.
	for _, mountedVolume := range rc.actualStateOfWorld.GetMountedVolumes() {
		if !rc.desiredStateOfWorld.PodExistsInVolume(mountedVolume.PodName, mountedVolume.VolumeName) {
			// Volume is mounted, unmount it
			glog.V(5).Infof(mountedVolume.GenerateMsgDetailed("Starting operationExecutor.UnmountVolume", ""))
			err := rc.operationExecutor.UnmountVolume(
				mountedVolume.MountedVolume, rc.actualStateOfWorld, rc.kubeletPodsDir)
			if err != nil &&
				!nestedpendingoperations.IsAlreadyExists(err) &&
				!exponentialbackoff.IsExponentialBackoff(err) {
				// Ignore nestedpendingoperations.IsAlreadyExists and exponentialbackoff.IsExponentialBackoff errors, they are expected.
				// Log all other errors.
				glog.Errorf(mountedVolume.GenerateErrorDetailed(fmt.Sprintf("operationExecutor.UnmountVolume failed (controllerAttachDetachEnabled %v)", rc.controllerAttachDetachEnabled), err).Error())
			}
			if err == nil {
				glog.Infof(mountedVolume.GenerateMsgDetailed("operationExecutor.UnmountVolume started", ""))
			}
		}
	}

	// Ensure volumes that should be attached/mounted are attached/mounted.
	for _, volumeToMount := range rc.desiredStateOfWorld.GetVolumesToMount() {
		volMounted, devicePath, err := rc.actualStateOfWorld.PodExistsInVolume(volumeToMount.PodName, volumeToMount.VolumeName)
		volumeToMount.DevicePath = devicePath
		if cache.IsVolumeNotAttachedError(err) {
			if rc.controllerAttachDetachEnabled || !volumeToMount.PluginIsAttachable {
				// Volume is not attached (or doesn't implement attacher), kubelet attach is disabled, wait
				// for controller to finish attaching volume.
				glog.V(5).Infof(volumeToMount.GenerateMsgDetailed("Starting operationExecutor.VerifyControllerAttachedVolume", ""))
				err := rc.operationExecutor.VerifyControllerAttachedVolume(
					volumeToMount.VolumeToMount,
					rc.nodeName,
					rc.actualStateOfWorld)
				if err != nil &&
					!nestedpendingoperations.IsAlreadyExists(err) &&
					!exponentialbackoff.IsExponentialBackoff(err) {
					// Ignore nestedpendingoperations.IsAlreadyExists and exponentialbackoff.IsExponentialBackoff errors, they are expected.
					// Log all other errors.
					glog.Errorf(volumeToMount.GenerateErrorDetailed(fmt.Sprintf("operationExecutor.VerifyControllerAttachedVolume failed (controllerAttachDetachEnabled %v)", rc.controllerAttachDetachEnabled), err).Error())
				}
				if err == nil {
					glog.Infof(volumeToMount.GenerateMsgDetailed("operationExecutor.VerifyControllerAttachedVolume started", ""))
				}
			} else {
				// Volume is not attached to node, kubelet attach is enabled, volume implements an attacher,
				// so attach it
				volumeToAttach := operationexecutor.VolumeToAttach{
					VolumeName: volumeToMount.VolumeName,
					VolumeSpec: volumeToMount.VolumeSpec,
					NodeName:   rc.nodeName,
				}
				glog.V(5).Infof(volumeToAttach.GenerateMsgDetailed("Starting operationExecutor.AttachVolume", ""))
				err := rc.operationExecutor.AttachVolume(volumeToAttach, rc.actualStateOfWorld)
				if err != nil &&
					!nestedpendingoperations.IsAlreadyExists(err) &&
					!exponentialbackoff.IsExponentialBackoff(err) {
					// Ignore nestedpendingoperations.IsAlreadyExists and exponentialbackoff.IsExponentialBackoff errors, they are expected.
					// Log all other errors.
					glog.Errorf(volumeToMount.GenerateErrorDetailed(fmt.Sprintf("operationExecutor.AttachVolume failed (controllerAttachDetachEnabled %v)", rc.controllerAttachDetachEnabled), err).Error())
				}
				if err == nil {
					glog.Infof(volumeToMount.GenerateMsgDetailed("operationExecutor.AttachVolume started", ""))
				}
			}
		} else if !volMounted || cache.IsRemountRequiredError(err) {
			// Volume is not mounted, or is already mounted, but requires remounting
			remountingLogStr := ""
			isRemount := cache.IsRemountRequiredError(err)
			if isRemount {
				remountingLogStr = "Volume is already mounted to pod, but remount was requested."
			}
			glog.V(4).Infof(volumeToMount.GenerateMsgDetailed("Starting operationExecutor.MountVolume", remountingLogStr))
			err := rc.operationExecutor.MountVolume(
				rc.waitForAttachTimeout,
				volumeToMount.VolumeToMount,
				rc.actualStateOfWorld,
				isRemount)
			if err != nil &&
				!nestedpendingoperations.IsAlreadyExists(err) &&
				!exponentialbackoff.IsExponentialBackoff(err) {
				// Ignore nestedpendingoperations.IsAlreadyExists and exponentialbackoff.IsExponentialBackoff errors, they are expected.
				// Log all other errors.
				glog.Errorf(volumeToMount.GenerateErrorDetailed(fmt.Sprintf("operationExecutor.MountVolume failed (controllerAttachDetachEnabled %v)", rc.controllerAttachDetachEnabled), err).Error())
			}
			if err == nil {
				if remountingLogStr == "" {
					glog.V(1).Infof(volumeToMount.GenerateMsgDetailed("operationExecutor.MountVolume started", remountingLogStr))
				} else {
					glog.V(5).Infof(volumeToMount.GenerateMsgDetailed("operationExecutor.MountVolume started", remountingLogStr))
				}
			}
		}
	}

	// Ensure devices that should be detached/unmounted are detached/unmounted.
	for _, attachedVolume := range rc.actualStateOfWorld.GetUnmountedVolumes() {
		// Check IsOperationPending to avoid marking a volume as detached if it's in the process of mounting.
		if !rc.desiredStateOfWorld.VolumeExists(attachedVolume.VolumeName) &&
			!rc.operationExecutor.IsOperationPending(attachedVolume.VolumeName, nestedpendingoperations.EmptyUniquePodName) {
			if attachedVolume.GloballyMounted {
				// Volume is globally mounted to device, unmount it
				glog.V(5).Infof(attachedVolume.GenerateMsgDetailed("Starting operationExecutor.UnmountDevice", ""))
				err := rc.operationExecutor.UnmountDevice(
					attachedVolume.AttachedVolume, rc.actualStateOfWorld, rc.mounter)
				if err != nil &&
					!nestedpendingoperations.IsAlreadyExists(err) &&
					!exponentialbackoff.IsExponentialBackoff(err) {
					// Ignore nestedpendingoperations.IsAlreadyExists and exponentialbackoff.IsExponentialBackoff errors, they are expected.
					// Log all other errors.
					glog.Errorf(attachedVolume.GenerateErrorDetailed(fmt.Sprintf("operationExecutor.UnmountDevice failed (controllerAttachDetachEnabled %v)", rc.controllerAttachDetachEnabled), err).Error())
				}
				if err == nil {
					glog.Infof(attachedVolume.GenerateMsgDetailed("operationExecutor.UnmountDevice started", ""))
				}
			} else {
				// Volume is attached to node, detach it
				// Kubelet not responsible for detaching or this volume has a non-attachable volume plugin.
				if rc.controllerAttachDetachEnabled || !attachedVolume.PluginIsAttachable {
					rc.actualStateOfWorld.MarkVolumeAsDetached(attachedVolume.VolumeName, attachedVolume.NodeName)
					glog.Infof(attachedVolume.GenerateMsgDetailed("Volume detached", fmt.Sprintf("DevicePath %q", attachedVolume.DevicePath)))
				} else {
					// Only detach if kubelet detach is enabled
					glog.V(5).Infof(attachedVolume.GenerateMsgDetailed("Starting operationExecutor.DetachVolume", ""))
					err := rc.operationExecutor.DetachVolume(
						attachedVolume.AttachedVolume, false /* verifySafeToDetach */, rc.actualStateOfWorld)
					if err != nil &&
						!nestedpendingoperations.IsAlreadyExists(err) &&
						!exponentialbackoff.IsExponentialBackoff(err) {
						// Ignore nestedpendingoperations.IsAlreadyExists && exponentialbackoff.IsExponentialBackoff errors, they are expected.
						// Log all other errors.
						glog.Errorf(attachedVolume.GenerateErrorDetailed(fmt.Sprintf("operationExecutor.DetachVolume failed (controllerAttachDetachEnabled %v)", rc.controllerAttachDetachEnabled), err).Error())
					}
					if err == nil {
						glog.Infof(attachedVolume.GenerateMsgDetailed("operationExecutor.DetachVolume started", ""))
					}
				}
			}
		}
	}
}

// sync process tries to observe the real world by scanning all pods' volume directories from the disk.
// If the actual and desired state of worlds are not consistent with the observed world, it means that some
// mounted volumes are left out probably during kubelet restart. This process will reconstruct
// the volumes and update the actual and desired states. For the volumes that cannot support reconstruction,
// it will try to clean up the mount paths with operation executor.
func (rc *reconciler) sync() {
	defer rc.updateLastSyncTime()
	rc.syncStates()
}

func (rc *reconciler) updateLastSyncTime() {
	rc.timeOfLastSync = time.Now()
}

func (rc *reconciler) StatesHasBeenSynced() bool {
	return !rc.timeOfLastSync.IsZero()
}

type podVolume struct {
	podName        volumetypes.UniquePodName
	volumeSpecName string
	mountPath      string
	pluginName     string
	volumeMode     v1.PersistentVolumeMode
}

type reconstructedVolume struct {
	volumeName          v1.UniqueVolumeName
	podName             volumetypes.UniquePodName
	volumeSpec          *volumepkg.Spec
	outerVolumeSpecName string
	pod                 *v1.Pod
	attachablePlugin    volumepkg.AttachableVolumePlugin
	volumeGidValue      string
	devicePath          string
	reportedInUse       bool
	mounter             volumepkg.Mounter
	blockVolumeMapper   volumepkg.BlockVolumeMapper
}

// syncStates scans the volume directories under the given pod directory.
// If the volume is not in desired state of world, this function will reconstruct
// the volume related information and put it in both the actual and desired state of worlds.
// For some volume plugins that cannot support reconstruction, it will clean up the existing
// mount points since the volume is no long needed (removed from desired state)
func (rc *reconciler) syncStates() {
	// Get volumes information by reading the pod's directory
	podVolumes, err := getVolumesFromPodDir(rc.kubeletPodsDir)
	if err != nil {
		glog.Errorf("Cannot get volumes from disk %v", err)
		return
	}
	volumesNeedUpdate := make(map[v1.UniqueVolumeName]*reconstructedVolume)
	volumeNeedReport := []v1.UniqueVolumeName{}
	for _, volume := range podVolumes {
		if rc.actualStateOfWorld.VolumeExistsWithSpecName(volume.podName, volume.volumeSpecName) {
			glog.V(4).Infof("Volume exists in actual state (volume.SpecName %s, pod.UID %s), skip cleaning up mounts", volume.volumeSpecName, volume.podName)
			// There is nothing to reconstruct
			continue
		}
		volumeInDSW := rc.desiredStateOfWorld.VolumeExistsWithSpecName(volume.podName, volume.volumeSpecName)

		reconstructedVolume, err := rc.reconstructVolume(volume)
		if err != nil {
			if volumeInDSW {
				// Some pod needs the volume, don't clean it up and hope that
				// reconcile() calls SetUp and reconstructs the volume in ASW.
				glog.V(4).Infof("Volume exists in desired state (volume.SpecName %s, pod.UID %s), skip cleaning up mounts", volume.volumeSpecName, volume.podName)
				continue
			}
			// No pod needs the volume.
			glog.Warningf("Could not construct volume information, cleanup the mounts. (pod.UID %s, volume.SpecName %s): %v", volume.podName, volume.volumeSpecName, err)
			rc.cleanupMounts(volume)
			continue
		}
		if volumeInDSW {
			// Some pod needs the volume. And it exists on disk. Some previous
			// kubelet must have created the directory, therefore it must have
			// reported the volume as in use. Mark the volume as in use also in
			// this new kubelet so reconcile() calls SetUp and re-mounts the
			// volume if it's necessary.
			volumeNeedReport = append(volumeNeedReport, reconstructedVolume.volumeName)
			glog.V(4).Infof("Volume exists in desired state (volume.SpecName %s, pod.UID %s), marking as InUse", volume.volumeSpecName, volume.podName)
			continue
		}
		// There is no pod that uses the volume.
		if rc.operationExecutor.IsOperationPending(reconstructedVolume.volumeName, nestedpendingoperations.EmptyUniquePodName) {
			glog.Warning("Volume is in pending operation, skip cleaning up mounts")
		}
		glog.V(2).Infof(
			"Reconciler sync states: could not find pod information in desired state, update it in actual state: %+v",
			reconstructedVolume)
		volumesNeedUpdate[reconstructedVolume.volumeName] = reconstructedVolume
	}

	if len(volumesNeedUpdate) > 0 {
		if err = rc.updateStates(volumesNeedUpdate); err != nil {
			glog.Errorf("Error occurred during reconstruct volume from disk: %v", err)
		}
	}
	if len(volumeNeedReport) > 0 {
		rc.desiredStateOfWorld.MarkVolumesReportedInUse(volumeNeedReport)
	}
}

func (rc *reconciler) cleanupMounts(volume podVolume) {
	glog.V(2).Infof("Reconciler sync states: could not find information (PID: %s) (Volume SpecName: %s) in desired state, clean up the mount points",
		volume.podName, volume.volumeSpecName)
	mountedVolume := operationexecutor.MountedVolume{
		PodName:             volume.podName,
		VolumeName:          v1.UniqueVolumeName(volume.volumeSpecName),
		InnerVolumeSpecName: volume.volumeSpecName,
		PluginName:          volume.pluginName,
		PodUID:              types.UID(volume.podName),
	}
	// TODO: Currently cleanupMounts only includes UnmountVolume operation. In the next PR, we will add
	// to unmount both volume and device in the same routine.
	err := rc.operationExecutor.UnmountVolume(mountedVolume, rc.actualStateOfWorld, rc.kubeletPodsDir)
	if err != nil {
		glog.Errorf(mountedVolume.GenerateErrorDetailed(fmt.Sprintf("volumeHandler.UnmountVolumeHandler for UnmountVolume failed"), err).Error())
		return
	}
}

// Reconstruct volume data structure by reading the pod's volume directories
func (rc *reconciler) reconstructVolume(volume podVolume) (*reconstructedVolume, error) {
	// plugin initializations
	plugin, err := rc.volumePluginMgr.FindPluginByName(volume.pluginName)
	if err != nil {
		return nil, err
	}
	attachablePlugin, err := rc.volumePluginMgr.FindAttachablePluginByName(volume.pluginName)
	if err != nil {
		return nil, err
	}

	// Create pod object
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: types.UID(volume.podName),
		},
	}
	mapperPlugin, err := rc.volumePluginMgr.FindMapperPluginByName(volume.pluginName)
	if err != nil {
		return nil, err
	}
	volumeSpec, err := rc.operationExecutor.ReconstructVolumeOperation(
		volume.volumeMode,
		plugin,
		mapperPlugin,
		pod.UID,
		volume.podName,
		volume.volumeSpecName,
		volume.mountPath,
		volume.pluginName)
	if err != nil {
		return nil, err
	}

	var uniqueVolumeName v1.UniqueVolumeName
	if attachablePlugin != nil {
		uniqueVolumeName, err = util.GetUniqueVolumeNameFromSpec(plugin, volumeSpec)
		if err != nil {
			return nil, err
		}
	} else {
		uniqueVolumeName = util.GetUniqueVolumeNameForNonAttachableVolume(volume.podName, plugin, volumeSpec)
	}
	// Check existence of mount point for filesystem volume or symbolic link for block volume
	isExist, checkErr := rc.operationExecutor.CheckVolumeExistenceOperation(volumeSpec, volume.mountPath, volumeSpec.Name(), rc.mounter, uniqueVolumeName, volume.podName, pod.UID, attachablePlugin)
	if checkErr != nil {
		return nil, checkErr
	}
	// If mount or symlink doesn't exist, volume reconstruction should be failed
	if !isExist {
		return nil, fmt.Errorf("Volume: %q is not mounted", uniqueVolumeName)
	}

	volumeMounter, newMounterErr := plugin.NewMounter(
		volumeSpec,
		pod,
		volumepkg.VolumeOptions{})
	if newMounterErr != nil {
		return nil, fmt.Errorf(
			"reconstructVolume.NewMounter failed for volume %q (spec.Name: %q) pod %q (UID: %q) with: %v",
			uniqueVolumeName,
			volumeSpec.Name(),
			volume.podName,
			pod.UID,
			newMounterErr)
	}

	// TODO: remove feature gate check after no longer needed
	var volumeMapper volumepkg.BlockVolumeMapper
	if utilfeature.DefaultFeatureGate.Enabled(features.BlockVolume) && volume.volumeMode == v1.PersistentVolumeBlock {
		var newMapperErr error
		if mapperPlugin != nil {
			volumeMapper, newMapperErr = mapperPlugin.NewBlockVolumeMapper(
				volumeSpec,
				pod,
				volumepkg.VolumeOptions{})
			if newMapperErr != nil {
				return nil, fmt.Errorf(
					"reconstructVolume.NewBlockVolumeMapper failed for volume %q (spec.Name: %q) pod %q (UID: %q) with: %v",
					uniqueVolumeName,
					volumeSpec.Name(),
					volume.podName,
					pod.UID,
					newMapperErr)
			}
		}
	}

	reconstructedVolume := &reconstructedVolume{
		volumeName: uniqueVolumeName,
		podName:    volume.podName,
		volumeSpec: volumeSpec,
		// volume.volumeSpecName is actually InnerVolumeSpecName. It will not be used
		// for volume cleanup.
		// TODO: in case pod is added back before reconciler starts to unmount, we can update this field from desired state information
		outerVolumeSpecName: volume.volumeSpecName,
		pod:                 pod,
		attachablePlugin:    attachablePlugin,
		volumeGidValue:      "",
		// devicePath is updated during updateStates() by checking node status's VolumesAttached data.
		// TODO: get device path directly from the volume mount path.
		devicePath:        "",
		mounter:           volumeMounter,
		blockVolumeMapper: volumeMapper,
	}
	return reconstructedVolume, nil
}

// updateDevicePath gets the node status to retrieve volume device path information.
func (rc *reconciler) updateDevicePath(volumesNeedUpdate map[v1.UniqueVolumeName]*reconstructedVolume) {
	node, fetchErr := rc.kubeClient.CoreV1().Nodes().Get(string(rc.nodeName), metav1.GetOptions{})
	if fetchErr != nil {
		glog.Errorf("updateStates in reconciler: could not get node status with error %v", fetchErr)
	} else {
		for _, attachedVolume := range node.Status.VolumesAttached {
			if volume, exists := volumesNeedUpdate[attachedVolume.Name]; exists {
				volume.devicePath = attachedVolume.DevicePath
				volumesNeedUpdate[attachedVolume.Name] = volume
				glog.V(4).Infof("Update devicePath from node status for volume (%q): %q", attachedVolume.Name, volume.devicePath)
			}
		}
	}
}

func getDeviceMountPath(volume *reconstructedVolume) (string, error) {
	volumeAttacher, err := volume.attachablePlugin.NewAttacher()
	if volumeAttacher == nil || err != nil {
		return "", err
	}
	deviceMountPath, err :=
		volumeAttacher.GetDeviceMountPath(volume.volumeSpec)
	if err != nil {
		return "", err
	}

	if volume.blockVolumeMapper != nil {
		deviceMountPath, err =
			volume.blockVolumeMapper.GetGlobalMapPath(volume.volumeSpec)
		if err != nil {
			return "", err
		}
	}
	return deviceMountPath, nil
}

func (rc *reconciler) updateStates(volumesNeedUpdate map[v1.UniqueVolumeName]*reconstructedVolume) error {
	// Get the node status to retrieve volume device path information.
	rc.updateDevicePath(volumesNeedUpdate)

	for _, volume := range volumesNeedUpdate {
		err := rc.actualStateOfWorld.MarkVolumeAsAttached(
			//TODO: the devicePath might not be correct for some volume plugins: see issue #54108
			volume.volumeName, volume.volumeSpec, "" /* nodeName */, volume.devicePath)
		if err != nil {
			glog.Errorf("Could not add volume information to actual state of world: %v", err)
			continue
		}
		err = rc.actualStateOfWorld.MarkVolumeAsMounted(
			volume.podName,
			types.UID(volume.podName),
			volume.volumeName,
			volume.mounter,
			volume.blockVolumeMapper,
			volume.outerVolumeSpecName,
			volume.volumeGidValue,
			volume.volumeSpec)
		if err != nil {
			glog.Errorf("Could not add pod to volume information to actual state of world: %v", err)
			continue
		}
		glog.V(4).Infof("Volume: %s (pod UID %s) is marked as mounted and added into the actual state", volume.volumeName, volume.podName)
		if volume.attachablePlugin != nil {
			deviceMountPath, err := getDeviceMountPath(volume)
			if err != nil {
				glog.Errorf("Could not find device mount path for volume %s", volume.volumeName)
				continue
			}
			err = rc.actualStateOfWorld.MarkDeviceAsMounted(volume.volumeName, volume.devicePath, deviceMountPath)
			if err != nil {
				glog.Errorf("Could not mark device is mounted to actual state of world: %v", err)
				continue
			}
			glog.V(4).Infof("Volume: %s (pod UID %s) is marked device as mounted and added into the actual state", volume.volumeName, volume.podName)
		}
	}
	return nil
}

// getVolumesFromPodDir scans through the volumes directories under the given pod directory.
// It returns a list of pod volume information including pod's uid, volume's plugin name, mount path,
// and volume spec name.
func getVolumesFromPodDir(podDir string) ([]podVolume, error) {
	podsDirInfo, err := ioutil.ReadDir(podDir)
	if err != nil {
		return nil, err
	}
	volumes := []podVolume{}
	for i := range podsDirInfo {
		if !podsDirInfo[i].IsDir() {
			continue
		}
		podName := podsDirInfo[i].Name()
		podDir := path.Join(podDir, podName)

		// Find filesystem volume information
		// ex. filesystem volume: /pods/{podUid}/volume/{escapeQualifiedPluginName}/{volumeName}
		volumesDirs := map[v1.PersistentVolumeMode]string{
			v1.PersistentVolumeFilesystem: path.Join(podDir, config.DefaultKubeletVolumesDirName),
		}
		// TODO: remove feature gate check after no longer needed
		if utilfeature.DefaultFeatureGate.Enabled(features.BlockVolume) {
			// Find block volume information
			// ex. block volume: /pods/{podUid}/volumeDevices/{escapeQualifiedPluginName}/{volumeName}
			volumesDirs[v1.PersistentVolumeBlock] = path.Join(podDir, config.DefaultKubeletVolumeDevicesDirName)
		}
		for volumeMode, volumesDir := range volumesDirs {
			var volumesDirInfo []os.FileInfo
			if volumesDirInfo, err = ioutil.ReadDir(volumesDir); err != nil {
				// Just skip the loop because given volumesDir doesn't exist depending on volumeMode
				continue
			}
			for _, volumeDir := range volumesDirInfo {
				pluginName := volumeDir.Name()
				volumePluginPath := path.Join(volumesDir, pluginName)
				volumePluginDirs, err := utilfile.ReadDirNoStat(volumePluginPath)
				if err != nil {
					glog.Errorf("Could not read volume plugin directory %q: %v", volumePluginPath, err)
					continue
				}
				unescapePluginName := utilstrings.UnescapeQualifiedNameForDisk(pluginName)
				for _, volumeName := range volumePluginDirs {
					mountPath := path.Join(volumePluginPath, volumeName)
					glog.V(5).Infof("podName: %v, mount path from volume plugin directory: %v, ", podName, mountPath)
					volumes = append(volumes, podVolume{
						podName:        volumetypes.UniquePodName(podName),
						volumeSpecName: volumeName,
						mountPath:      mountPath,
						pluginName:     unescapePluginName,
						volumeMode:     volumeMode,
					})
				}
			}
		}
	}
	glog.V(4).Infof("Get volumes from pod directory %q %+v", podDir, volumes)
	return volumes, nil
}
