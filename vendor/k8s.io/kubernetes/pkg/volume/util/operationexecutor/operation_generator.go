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
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	expandcache "k8s.io/kubernetes/pkg/controller/volume/expand/cache"
	"k8s.io/kubernetes/pkg/features"
	kevents "k8s.io/kubernetes/pkg/kubelet/events"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/util/resizefs"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"
	volumetypes "k8s.io/kubernetes/pkg/volume/util/types"
	"k8s.io/kubernetes/pkg/volume/util/volumepathhandler"
)

var _ OperationGenerator = &operationGenerator{}

type operationGenerator struct {
	// Used to fetch objects from the API server like Node in the
	// VerifyControllerAttachedVolume operation.
	kubeClient clientset.Interface

	// volumePluginMgr is the volume plugin manager used to create volume
	// plugin objects.
	volumePluginMgr *volume.VolumePluginMgr

	// recorder is used to record events in the API server
	recorder record.EventRecorder

	// checkNodeCapabilitiesBeforeMount, if set, enables the CanMount check,
	// which verifies that the components (binaries, etc.) required to mount
	// the volume are available on the underlying node before attempting mount.
	checkNodeCapabilitiesBeforeMount bool

	// blkUtil provides volume path related operations for block volume
	blkUtil volumepathhandler.BlockVolumePathHandler
}

// NewOperationGenerator is returns instance of operationGenerator
func NewOperationGenerator(kubeClient clientset.Interface,
	volumePluginMgr *volume.VolumePluginMgr,
	recorder record.EventRecorder,
	checkNodeCapabilitiesBeforeMount bool,
	blkUtil volumepathhandler.BlockVolumePathHandler) OperationGenerator {

	return &operationGenerator{
		kubeClient:      kubeClient,
		volumePluginMgr: volumePluginMgr,
		recorder:        recorder,
		checkNodeCapabilitiesBeforeMount: checkNodeCapabilitiesBeforeMount,
		blkUtil: blkUtil,
	}
}

// OperationGenerator interface that extracts out the functions from operation_executor to make it dependency injectable
type OperationGenerator interface {
	// Generates the MountVolume function needed to perform the mount of a volume plugin
	GenerateMountVolumeFunc(waitForAttachTimeout time.Duration, volumeToMount VolumeToMount, actualStateOfWorldMounterUpdater ActualStateOfWorldMounterUpdater, isRemount bool) (volumetypes.GeneratedOperations, error)

	// Generates the UnmountVolume function needed to perform the unmount of a volume plugin
	GenerateUnmountVolumeFunc(volumeToUnmount MountedVolume, actualStateOfWorld ActualStateOfWorldMounterUpdater, podsDir string) (volumetypes.GeneratedOperations, error)

	// Generates the AttachVolume function needed to perform attach of a volume plugin
	GenerateAttachVolumeFunc(volumeToAttach VolumeToAttach, actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error)

	// Generates the DetachVolume function needed to perform the detach of a volume plugin
	GenerateDetachVolumeFunc(volumeToDetach AttachedVolume, verifySafeToDetach bool, actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error)

	// Generates the VolumesAreAttached function needed to verify if volume plugins are attached
	GenerateVolumesAreAttachedFunc(attachedVolumes []AttachedVolume, nodeName types.NodeName, actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error)

	// Generates the UnMountDevice function needed to perform the unmount of a device
	GenerateUnmountDeviceFunc(deviceToDetach AttachedVolume, actualStateOfWorld ActualStateOfWorldMounterUpdater, mounter mount.Interface) (volumetypes.GeneratedOperations, error)

	// Generates the function needed to check if the attach_detach controller has attached the volume plugin
	GenerateVerifyControllerAttachedVolumeFunc(volumeToMount VolumeToMount, nodeName types.NodeName, actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error)

	// Generates the MapVolume function needed to perform the map of a volume plugin
	GenerateMapVolumeFunc(waitForAttachTimeout time.Duration, volumeToMount VolumeToMount, actualStateOfWorldMounterUpdater ActualStateOfWorldMounterUpdater) (volumetypes.GeneratedOperations, error)

	// Generates the UnmapVolume function needed to perform the unmap of a volume plugin
	GenerateUnmapVolumeFunc(volumeToUnmount MountedVolume, actualStateOfWorld ActualStateOfWorldMounterUpdater) (volumetypes.GeneratedOperations, error)

	// Generates the UnmapDevice function needed to perform the unmap of a device
	GenerateUnmapDeviceFunc(deviceToDetach AttachedVolume, actualStateOfWorld ActualStateOfWorldMounterUpdater, mounter mount.Interface) (volumetypes.GeneratedOperations, error)

	// GetVolumePluginMgr returns volume plugin manager
	GetVolumePluginMgr() *volume.VolumePluginMgr

	GenerateBulkVolumeVerifyFunc(
		map[types.NodeName][]*volume.Spec,
		string,
		map[*volume.Spec]v1.UniqueVolumeName, ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error)

	GenerateExpandVolumeFunc(*expandcache.PVCWithResizeRequest, expandcache.VolumeResizeMap) (volumetypes.GeneratedOperations, error)
}

func (og *operationGenerator) GenerateVolumesAreAttachedFunc(
	attachedVolumes []AttachedVolume,
	nodeName types.NodeName,
	actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {

	// volumesPerPlugin maps from a volume plugin to a list of volume specs which belong
	// to this type of plugin
	volumesPerPlugin := make(map[string][]*volume.Spec)
	// volumeSpecMap maps from a volume spec to its unique volumeName which will be used
	// when calling MarkVolumeAsDetached
	volumeSpecMap := make(map[*volume.Spec]v1.UniqueVolumeName)
	// Iterate each volume spec and put them into a map index by the pluginName
	for _, volumeAttached := range attachedVolumes {
		if volumeAttached.VolumeSpec == nil {
			glog.Errorf("VerifyVolumesAreAttached.GenerateVolumesAreAttachedFunc: nil spec for volume %s", volumeAttached.VolumeName)
			continue
		}
		volumePlugin, err :=
			og.volumePluginMgr.FindPluginBySpec(volumeAttached.VolumeSpec)
		if err != nil || volumePlugin == nil {
			glog.Errorf(volumeAttached.GenerateErrorDetailed("VolumesAreAttached.FindPluginBySpec failed", err).Error())
		}
		volumeSpecList, pluginExists := volumesPerPlugin[volumePlugin.GetPluginName()]
		if !pluginExists {
			volumeSpecList = []*volume.Spec{}
		}
		volumeSpecList = append(volumeSpecList, volumeAttached.VolumeSpec)
		volumesPerPlugin[volumePlugin.GetPluginName()] = volumeSpecList
		volumeSpecMap[volumeAttached.VolumeSpec] = volumeAttached.VolumeName
	}

	volumesAreAttachedFunc := func() (error, error) {

		// For each volume plugin, pass the list of volume specs to VolumesAreAttached to check
		// whether the volumes are still attached.
		for pluginName, volumesSpecs := range volumesPerPlugin {
			attachableVolumePlugin, err :=
				og.volumePluginMgr.FindAttachablePluginByName(pluginName)
			if err != nil || attachableVolumePlugin == nil {
				glog.Errorf(
					"VolumeAreAttached.FindAttachablePluginBySpec failed for plugin %q with: %v",
					pluginName,
					err)
				continue
			}

			volumeAttacher, newAttacherErr := attachableVolumePlugin.NewAttacher()
			if newAttacherErr != nil {
				glog.Errorf(
					"VolumesAreAttached.NewAttacher failed for getting plugin %q with: %v",
					pluginName,
					newAttacherErr)
				continue
			}

			attached, areAttachedErr := volumeAttacher.VolumesAreAttached(volumesSpecs, nodeName)
			if areAttachedErr != nil {
				glog.Errorf(
					"VolumesAreAttached failed for checking on node %q with: %v",
					nodeName,
					areAttachedErr)
				continue
			}

			for spec, check := range attached {
				if !check {
					actualStateOfWorld.MarkVolumeAsDetached(volumeSpecMap[spec], nodeName)
					glog.V(1).Infof("VerifyVolumesAreAttached determined volume %q (spec.Name: %q) is no longer attached to node %q, therefore it was marked as detached.",
						volumeSpecMap[spec], spec.Name(), nodeName)
				}
			}
		}
		return nil, nil
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     volumesAreAttachedFunc,
		CompleteFunc:      util.OperationCompleteHook("<n/a>", "verify_volumes_are_attached_per_node"),
		EventRecorderFunc: nil, // nil because we do not want to generate event on error
	}, nil
}

func (og *operationGenerator) GenerateBulkVolumeVerifyFunc(
	pluginNodeVolumes map[types.NodeName][]*volume.Spec,
	pluginName string,
	volumeSpecMap map[*volume.Spec]v1.UniqueVolumeName,
	actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {

	bulkVolumeVerifyFunc := func() (error, error) {
		attachableVolumePlugin, err :=
			og.volumePluginMgr.FindAttachablePluginByName(pluginName)
		if err != nil || attachableVolumePlugin == nil {
			glog.Errorf(
				"BulkVerifyVolume.FindAttachablePluginBySpec failed for plugin %q with: %v",
				pluginName,
				err)
			return nil, nil
		}

		volumeAttacher, newAttacherErr := attachableVolumePlugin.NewAttacher()

		if newAttacherErr != nil {
			glog.Errorf(
				"BulkVerifyVolume.NewAttacher failed for getting plugin %q with: %v",
				attachableVolumePlugin,
				newAttacherErr)
			return nil, nil
		}
		bulkVolumeVerifier, ok := volumeAttacher.(volume.BulkVolumeVerifier)

		if !ok {
			glog.Errorf("BulkVerifyVolume failed to type assert attacher %q", bulkVolumeVerifier)
			return nil, nil
		}

		attached, bulkAttachErr := bulkVolumeVerifier.BulkVerifyVolumes(pluginNodeVolumes)
		if bulkAttachErr != nil {
			glog.Errorf("BulkVerifyVolume.BulkVerifyVolumes Error checking volumes are attached with %v", bulkAttachErr)
			return nil, nil
		}

		for nodeName, volumeSpecs := range pluginNodeVolumes {
			for _, volumeSpec := range volumeSpecs {
				nodeVolumeSpecs, nodeChecked := attached[nodeName]

				if !nodeChecked {
					glog.V(2).Infof("VerifyVolumesAreAttached.BulkVerifyVolumes failed for node %q and leaving volume %q as attached",
						nodeName,
						volumeSpec.Name())
					continue
				}

				check := nodeVolumeSpecs[volumeSpec]

				if !check {
					glog.V(2).Infof("VerifyVolumesAreAttached.BulkVerifyVolumes failed for node %q and volume %q",
						nodeName,
						volumeSpec.Name())
					actualStateOfWorld.MarkVolumeAsDetached(volumeSpecMap[volumeSpec], nodeName)
				}
			}
		}

		return nil, nil
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     bulkVolumeVerifyFunc,
		CompleteFunc:      util.OperationCompleteHook(pluginName, "verify_volumes_are_attached"),
		EventRecorderFunc: nil, // nil because we do not want to generate event on error
	}, nil

}

func (og *operationGenerator) GenerateAttachVolumeFunc(
	volumeToAttach VolumeToAttach,
	actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {
	// Get attacher plugin
	eventRecorderFunc := func(err *error) {
		if *err != nil {
			for _, pod := range volumeToAttach.ScheduledPods {
				og.recorder.Eventf(pod, v1.EventTypeWarning, kevents.FailedAttachVolume, (*err).Error())
			}
		}
	}

	attachableVolumePlugin, err :=
		og.volumePluginMgr.FindAttachablePluginBySpec(volumeToAttach.VolumeSpec)
	if err != nil || attachableVolumePlugin == nil {
		eventRecorderFunc(&err)
		return volumetypes.GeneratedOperations{}, volumeToAttach.GenerateErrorDetailed("AttachVolume.FindAttachablePluginBySpec failed", err)
	}

	volumeAttacher, newAttacherErr := attachableVolumePlugin.NewAttacher()
	if newAttacherErr != nil {
		eventRecorderFunc(&err)
		return volumetypes.GeneratedOperations{}, volumeToAttach.GenerateErrorDetailed("AttachVolume.NewAttacher failed", newAttacherErr)
	}

	attachVolumeFunc := func() (error, error) {
		// Execute attach
		devicePath, attachErr := volumeAttacher.Attach(
			volumeToAttach.VolumeSpec, volumeToAttach.NodeName)

		if attachErr != nil {
			if derr, ok := attachErr.(*util.DanglingAttachError); ok {
				addErr := actualStateOfWorld.MarkVolumeAsAttached(
					v1.UniqueVolumeName(""),
					volumeToAttach.VolumeSpec,
					derr.CurrentNode,
					derr.DevicePath)

				if addErr != nil {
					glog.Errorf("AttachVolume.MarkVolumeAsAttached failed to fix dangling volume error for volume %q with %s", volumeToAttach.VolumeName, addErr)
				}

			}
			// On failure, return error. Caller will log and retry.
			return volumeToAttach.GenerateError("AttachVolume.Attach failed", attachErr)
		}

		// Successful attach event is useful for user debugging
		simpleMsg, _ := volumeToAttach.GenerateMsg("AttachVolume.Attach succeeded", "")
		for _, pod := range volumeToAttach.ScheduledPods {
			og.recorder.Eventf(pod, v1.EventTypeNormal, kevents.SuccessfulAttachVolume, simpleMsg)
		}
		glog.Infof(volumeToAttach.GenerateMsgDetailed("AttachVolume.Attach succeeded", ""))

		// Update actual state of world
		addVolumeNodeErr := actualStateOfWorld.MarkVolumeAsAttached(
			v1.UniqueVolumeName(""), volumeToAttach.VolumeSpec, volumeToAttach.NodeName, devicePath)
		if addVolumeNodeErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToAttach.GenerateError("AttachVolume.MarkVolumeAsAttached failed", addVolumeNodeErr)
		}

		return nil, nil
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     attachVolumeFunc,
		EventRecorderFunc: eventRecorderFunc,
		CompleteFunc:      util.OperationCompleteHook(attachableVolumePlugin.GetPluginName(), "volume_attach"),
	}, nil
}

func (og *operationGenerator) GetVolumePluginMgr() *volume.VolumePluginMgr {
	return og.volumePluginMgr
}

func (og *operationGenerator) GenerateDetachVolumeFunc(
	volumeToDetach AttachedVolume,
	verifySafeToDetach bool,
	actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {
	var volumeName string
	var attachableVolumePlugin volume.AttachableVolumePlugin
	var pluginName string
	var err error

	if volumeToDetach.VolumeSpec != nil {
		// Get attacher plugin
		attachableVolumePlugin, err =
			og.volumePluginMgr.FindAttachablePluginBySpec(volumeToDetach.VolumeSpec)
		if err != nil || attachableVolumePlugin == nil {
			return volumetypes.GeneratedOperations{}, volumeToDetach.GenerateErrorDetailed("DetachVolume.FindAttachablePluginBySpec failed", err)
		}

		volumeName, err =
			attachableVolumePlugin.GetVolumeName(volumeToDetach.VolumeSpec)
		if err != nil {
			return volumetypes.GeneratedOperations{}, volumeToDetach.GenerateErrorDetailed("DetachVolume.GetVolumeName failed", err)
		}
	} else {
		// Get attacher plugin and the volumeName by splitting the volume unique name in case
		// there's no VolumeSpec: this happens only on attach/detach controller crash recovery
		// when a pod has been deleted during the controller downtime
		pluginName, volumeName, err = util.SplitUniqueName(volumeToDetach.VolumeName)
		if err != nil {
			return volumetypes.GeneratedOperations{}, volumeToDetach.GenerateErrorDetailed("DetachVolume.SplitUniqueName failed", err)
		}
		attachableVolumePlugin, err = og.volumePluginMgr.FindAttachablePluginByName(pluginName)
		if err != nil {
			return volumetypes.GeneratedOperations{}, volumeToDetach.GenerateErrorDetailed("DetachVolume.FindAttachablePluginBySpec failed", err)
		}
	}

	if pluginName == "" {
		pluginName = attachableVolumePlugin.GetPluginName()
	}

	volumeDetacher, err := attachableVolumePlugin.NewDetacher()
	if err != nil {
		return volumetypes.GeneratedOperations{}, volumeToDetach.GenerateErrorDetailed("DetachVolume.NewDetacher failed", err)
	}

	getVolumePluginMgrFunc := func() (error, error) {
		var err error
		if verifySafeToDetach {
			err = og.verifyVolumeIsSafeToDetach(volumeToDetach)
		}
		if err == nil {
			err = volumeDetacher.Detach(volumeName, volumeToDetach.NodeName)
		}
		if err != nil {
			// On failure, add volume back to ReportAsAttached list
			actualStateOfWorld.AddVolumeToReportAsAttached(
				volumeToDetach.VolumeName, volumeToDetach.NodeName)
			return volumeToDetach.GenerateError("DetachVolume.Detach failed", err)
		}

		glog.Infof(volumeToDetach.GenerateMsgDetailed("DetachVolume.Detach succeeded", ""))

		// Update actual state of world
		actualStateOfWorld.MarkVolumeAsDetached(
			volumeToDetach.VolumeName, volumeToDetach.NodeName)

		return nil, nil
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     getVolumePluginMgrFunc,
		CompleteFunc:      util.OperationCompleteHook(pluginName, "volume_detach"),
		EventRecorderFunc: nil, // nil because we do not want to generate event on error
	}, nil
}

func (og *operationGenerator) GenerateMountVolumeFunc(
	waitForAttachTimeout time.Duration,
	volumeToMount VolumeToMount,
	actualStateOfWorld ActualStateOfWorldMounterUpdater,
	isRemount bool) (volumetypes.GeneratedOperations, error) {
	// Get mounter plugin
	volumePlugin, err :=
		og.volumePluginMgr.FindPluginBySpec(volumeToMount.VolumeSpec)
	if err != nil || volumePlugin == nil {
		return volumetypes.GeneratedOperations{}, volumeToMount.GenerateErrorDetailed("MountVolume.FindPluginBySpec failed", err)
	}

	affinityErr := checkNodeAffinity(og, volumeToMount, volumePlugin)
	if affinityErr != nil {
		eventErr, detailedErr := volumeToMount.GenerateError("MountVolume.NodeAffinity check failed", affinityErr)
		og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeWarning, kevents.FailedMountVolume, eventErr.Error())
		return volumetypes.GeneratedOperations{}, detailedErr
	}

	volumeMounter, newMounterErr := volumePlugin.NewMounter(
		volumeToMount.VolumeSpec,
		volumeToMount.Pod,
		volume.VolumeOptions{})
	if newMounterErr != nil {
		eventErr, detailedErr := volumeToMount.GenerateError("MountVolume.NewMounter initialization failed", newMounterErr)
		og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeWarning, kevents.FailedMountVolume, eventErr.Error())
		return volumetypes.GeneratedOperations{}, detailedErr
	}

	mountCheckError := checkMountOptionSupport(og, volumeToMount, volumePlugin)

	if mountCheckError != nil {
		eventErr, detailedErr := volumeToMount.GenerateError("MountVolume.MountOptionSupport check failed", mountCheckError)
		og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeWarning, kevents.UnsupportedMountOption, eventErr.Error())
		return volumetypes.GeneratedOperations{}, detailedErr
	}

	// Get attacher, if possible
	attachableVolumePlugin, _ :=
		og.volumePluginMgr.FindAttachablePluginBySpec(volumeToMount.VolumeSpec)
	var volumeAttacher volume.Attacher
	if attachableVolumePlugin != nil {
		volumeAttacher, _ = attachableVolumePlugin.NewAttacher()
	}

	var fsGroup *int64
	if volumeToMount.Pod.Spec.SecurityContext != nil &&
		volumeToMount.Pod.Spec.SecurityContext.FSGroup != nil {
		fsGroup = volumeToMount.Pod.Spec.SecurityContext.FSGroup
	}

	mountVolumeFunc := func() (error, error) {
		if volumeAttacher != nil {
			// Wait for attachable volumes to finish attaching
			glog.Infof(volumeToMount.GenerateMsgDetailed("MountVolume.WaitForAttach entering", fmt.Sprintf("DevicePath %q", volumeToMount.DevicePath)))

			devicePath, err := volumeAttacher.WaitForAttach(
				volumeToMount.VolumeSpec, volumeToMount.DevicePath, volumeToMount.Pod, waitForAttachTimeout)
			if err != nil {
				// On failure, return error. Caller will log and retry.
				return volumeToMount.GenerateError("MountVolume.WaitForAttach failed", err)
			}

			glog.Infof(volumeToMount.GenerateMsgDetailed("MountVolume.WaitForAttach succeeded", fmt.Sprintf("DevicePath %q", devicePath)))

			deviceMountPath, err :=
				volumeAttacher.GetDeviceMountPath(volumeToMount.VolumeSpec)
			if err != nil {
				// On failure, return error. Caller will log and retry.
				return volumeToMount.GenerateError("MountVolume.GetDeviceMountPath failed", err)
			}

			// Mount device to global mount path
			err = volumeAttacher.MountDevice(
				volumeToMount.VolumeSpec,
				devicePath,
				deviceMountPath)
			if err != nil {
				// On failure, return error. Caller will log and retry.
				return volumeToMount.GenerateError("MountVolume.MountDevice failed", err)
			}

			glog.Infof(volumeToMount.GenerateMsgDetailed("MountVolume.MountDevice succeeded", fmt.Sprintf("device mount path %q", deviceMountPath)))

			// Update actual state of world to reflect volume is globally mounted
			markDeviceMountedErr := actualStateOfWorld.MarkDeviceAsMounted(
				volumeToMount.VolumeName, devicePath, deviceMountPath)
			if markDeviceMountedErr != nil {
				// On failure, return error. Caller will log and retry.
				return volumeToMount.GenerateError("MountVolume.MarkDeviceAsMounted failed", markDeviceMountedErr)
			}

			// resizeFileSystem will resize the file system if user has requested a resize of
			// underlying persistent volume and is allowed to do so.
			resizeSimpleError, resizeDetailedError := og.resizeFileSystem(volumeToMount, devicePath, deviceMountPath, volumePlugin.GetPluginName())

			if resizeSimpleError != nil || resizeDetailedError != nil {
				return resizeSimpleError, resizeDetailedError
			}

		}

		if og.checkNodeCapabilitiesBeforeMount {
			if canMountErr := volumeMounter.CanMount(); canMountErr != nil {
				err = fmt.Errorf(
					"Verify that your node machine has the required components before attempting to mount this volume type. %s",
					canMountErr)
				return volumeToMount.GenerateError("MountVolume.CanMount failed", err)
			}
		}

		// Execute mount
		mountErr := volumeMounter.SetUp(fsGroup)
		if mountErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError("MountVolume.SetUp failed", mountErr)
		}

		_, detailedMsg := volumeToMount.GenerateMsg("MountVolume.SetUp succeeded", "")
		verbosity := glog.Level(1)
		if isRemount {
			verbosity = glog.Level(4)
		}
		glog.V(verbosity).Infof(detailedMsg)

		// Update actual state of world
		markVolMountedErr := actualStateOfWorld.MarkVolumeAsMounted(
			volumeToMount.PodName,
			volumeToMount.Pod.UID,
			volumeToMount.VolumeName,
			volumeMounter,
			nil,
			volumeToMount.OuterVolumeSpecName,
			volumeToMount.VolumeGidValue,
			volumeToMount.VolumeSpec)
		if markVolMountedErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError("MountVolume.MarkVolumeAsMounted failed", markVolMountedErr)
		}

		return nil, nil
	}

	eventRecorderFunc := func(err *error) {
		if *err != nil {
			og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeWarning, kevents.FailedMountVolume, (*err).Error())
		}
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     mountVolumeFunc,
		EventRecorderFunc: eventRecorderFunc,
		CompleteFunc:      util.OperationCompleteHook(volumePlugin.GetPluginName(), "volume_mount"),
	}, nil
}

func (og *operationGenerator) resizeFileSystem(volumeToMount VolumeToMount, devicePath, deviceMountPath, pluginName string) (simpleErr, detailedErr error) {
	if !utilfeature.DefaultFeatureGate.Enabled(features.ExpandPersistentVolumes) {
		glog.V(4).Infof("Resizing is not enabled for this volume %s", volumeToMount.VolumeName)
		return nil, nil
	}

	mounter := og.volumePluginMgr.Host.GetMounter(pluginName)
	// Get expander, if possible
	expandableVolumePlugin, _ :=
		og.volumePluginMgr.FindExpandablePluginBySpec(volumeToMount.VolumeSpec)

	if expandableVolumePlugin != nil &&
		expandableVolumePlugin.RequiresFSResize() &&
		volumeToMount.VolumeSpec.PersistentVolume != nil {
		pv := volumeToMount.VolumeSpec.PersistentVolume
		pvc, err := og.kubeClient.CoreV1().PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name, metav1.GetOptions{})
		if err != nil {
			// Return error rather than leave the file system un-resized, caller will log and retry
			return volumeToMount.GenerateError("MountVolume.resizeFileSystem get PVC failed", err)
		}

		pvcStatusCap := pvc.Status.Capacity[v1.ResourceStorage]
		pvSpecCap := pv.Spec.Capacity[v1.ResourceStorage]
		if pvcStatusCap.Cmp(pvSpecCap) < 0 {
			// File system resize was requested, proceed
			glog.V(4).Infof(volumeToMount.GenerateMsgDetailed("MountVolume.resizeFileSystem entering", fmt.Sprintf("DevicePath %q", volumeToMount.DevicePath)))

			if volumeToMount.VolumeSpec.ReadOnly {
				simpleMsg, detailedMsg := volumeToMount.GenerateMsg("MountVolume.resizeFileSystem failed", "requested read-only file system")
				glog.Warningf(detailedMsg)
				og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeWarning, kevents.FileSystemResizeFailed, simpleMsg)
				return nil, nil
			}

			diskFormatter := &mount.SafeFormatAndMount{
				Interface: mounter,
				Exec:      og.volumePluginMgr.Host.GetExec(expandableVolumePlugin.GetPluginName()),
			}

			resizer := resizefs.NewResizeFs(diskFormatter)
			resizeStatus, resizeErr := resizer.Resize(devicePath, deviceMountPath)

			if resizeErr != nil {
				return volumeToMount.GenerateError("MountVolume.resizeFileSystem failed", resizeErr)
			}

			if resizeStatus {
				simpleMsg, detailedMsg := volumeToMount.GenerateMsg("MountVolume.resizeFileSystem succeeded", "")
				og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeNormal, kevents.FileSystemResizeSuccess, simpleMsg)
				glog.Infof(detailedMsg)
			}

			// File system resize succeeded, now update the PVC's Capacity to match the PV's
			err = util.MarkFSResizeFinished(pvc, pv.Spec.Capacity, og.kubeClient)
			if err != nil {
				// On retry, resizeFileSystem will be called again but do nothing
				return volumeToMount.GenerateError("MountVolume.resizeFileSystem update PVC status failed", err)
			}
			return nil, nil
		}
	}
	return nil, nil
}

func (og *operationGenerator) GenerateUnmountVolumeFunc(
	volumeToUnmount MountedVolume,
	actualStateOfWorld ActualStateOfWorldMounterUpdater,
	podsDir string) (volumetypes.GeneratedOperations, error) {
	// Get mountable plugin
	volumePlugin, err :=
		og.volumePluginMgr.FindPluginByName(volumeToUnmount.PluginName)
	if err != nil || volumePlugin == nil {
		return volumetypes.GeneratedOperations{}, volumeToUnmount.GenerateErrorDetailed("UnmountVolume.FindPluginByName failed", err)
	}

	volumeUnmounter, newUnmounterErr := volumePlugin.NewUnmounter(
		volumeToUnmount.InnerVolumeSpecName, volumeToUnmount.PodUID)
	if newUnmounterErr != nil {
		return volumetypes.GeneratedOperations{}, volumeToUnmount.GenerateErrorDetailed("UnmountVolume.NewUnmounter failed", newUnmounterErr)
	}

	unmountVolumeFunc := func() (error, error) {
		mounter := og.volumePluginMgr.Host.GetMounter(volumeToUnmount.PluginName)

		// Remove all bind-mounts for subPaths
		podDir := path.Join(podsDir, string(volumeToUnmount.PodUID))
		if err := mounter.CleanSubPaths(podDir, volumeToUnmount.InnerVolumeSpecName); err != nil {
			return volumeToUnmount.GenerateError("error cleaning subPath mounts", err)
		}

		// Execute unmount
		unmountErr := volumeUnmounter.TearDown()
		if unmountErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToUnmount.GenerateError("UnmountVolume.TearDown failed", unmountErr)
		}

		glog.Infof(
			"UnmountVolume.TearDown succeeded for volume %q (OuterVolumeSpecName: %q) pod %q (UID: %q). InnerVolumeSpecName %q. PluginName %q, VolumeGidValue %q",
			volumeToUnmount.VolumeName,
			volumeToUnmount.OuterVolumeSpecName,
			volumeToUnmount.PodName,
			volumeToUnmount.PodUID,
			volumeToUnmount.InnerVolumeSpecName,
			volumeToUnmount.PluginName,
			volumeToUnmount.VolumeGidValue)

		// Update actual state of world
		markVolMountedErr := actualStateOfWorld.MarkVolumeAsUnmounted(
			volumeToUnmount.PodName, volumeToUnmount.VolumeName)
		if markVolMountedErr != nil {
			// On failure, just log and exit
			glog.Errorf(volumeToUnmount.GenerateErrorDetailed("UnmountVolume.MarkVolumeAsUnmounted failed", markVolMountedErr).Error())
		}

		return nil, nil
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     unmountVolumeFunc,
		CompleteFunc:      util.OperationCompleteHook(volumePlugin.GetPluginName(), "volume_unmount"),
		EventRecorderFunc: nil, // nil because we do not want to generate event on error
	}, nil
}

func (og *operationGenerator) GenerateUnmountDeviceFunc(
	deviceToDetach AttachedVolume,
	actualStateOfWorld ActualStateOfWorldMounterUpdater,
	mounter mount.Interface) (volumetypes.GeneratedOperations, error) {
	// Get attacher plugin
	attachableVolumePlugin, err :=
		og.volumePluginMgr.FindAttachablePluginByName(deviceToDetach.PluginName)
	if err != nil || attachableVolumePlugin == nil {
		return volumetypes.GeneratedOperations{}, deviceToDetach.GenerateErrorDetailed("UnmountDevice.FindAttachablePluginBySpec failed", err)
	}

	volumeDetacher, err := attachableVolumePlugin.NewDetacher()
	if err != nil {
		return volumetypes.GeneratedOperations{}, deviceToDetach.GenerateErrorDetailed("UnmountDevice.NewDetacher failed", err)
	}
	unmountDeviceFunc := func() (error, error) {
		deviceMountPath := deviceToDetach.DeviceMountPath
		refs, err := attachableVolumePlugin.GetDeviceMountRefs(deviceMountPath)

		if err != nil || mount.HasMountRefs(deviceMountPath, refs) {
			if err == nil {
				err = fmt.Errorf("The device mount path %q is still mounted by other references %v", deviceMountPath, refs)
			}
			return deviceToDetach.GenerateError("GetDeviceMountRefs check failed", err)
		}
		// Execute unmount
		unmountDeviceErr := volumeDetacher.UnmountDevice(deviceMountPath)
		if unmountDeviceErr != nil {
			// On failure, return error. Caller will log and retry.
			return deviceToDetach.GenerateError("UnmountDevice failed", unmountDeviceErr)
		}
		// Before logging that UnmountDevice succeeded and moving on,
		// use mounter.PathIsDevice to check if the path is a device,
		// if so use mounter.DeviceOpened to check if the device is in use anywhere
		// else on the system. Retry if it returns true.
		deviceOpened, deviceOpenedErr := isDeviceOpened(deviceToDetach, mounter)
		if deviceOpenedErr != nil {
			return nil, deviceOpenedErr
		}
		// The device is still in use elsewhere. Caller will log and retry.
		if deviceOpened {
			return deviceToDetach.GenerateError(
				"UnmountDevice failed",
				fmt.Errorf("the device is in use when it was no longer expected to be in use"))
		}

		glog.Infof(deviceToDetach.GenerateMsg("UnmountDevice succeeded", ""))

		// Update actual state of world
		markDeviceUnmountedErr := actualStateOfWorld.MarkDeviceAsUnmounted(
			deviceToDetach.VolumeName)
		if markDeviceUnmountedErr != nil {
			// On failure, return error. Caller will log and retry.
			return deviceToDetach.GenerateError("MarkDeviceAsUnmounted failed", markDeviceUnmountedErr)
		}

		return nil, nil
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     unmountDeviceFunc,
		CompleteFunc:      util.OperationCompleteHook(attachableVolumePlugin.GetPluginName(), "unmount_device"),
		EventRecorderFunc: nil, // nil because we do not want to generate event on error
	}, nil
}

// GenerateMapVolumeFunc marks volume as mounted based on following steps.
// If plugin is attachable, call WaitForAttach() and then mark the device
// as mounted. On next step, SetUpDevice is called without dependent of
// plugin type, but this method mainly is targeted for none attachable plugin.
// After setup is done, create symbolic links on both global map path and pod
// device map path. Once symbolic links are created, take fd lock by
// loopback for the device to avoid silent volume replacement. This lock
// will be realased once no one uses the device.
// If all steps are completed, the volume is marked as mounted.
func (og *operationGenerator) GenerateMapVolumeFunc(
	waitForAttachTimeout time.Duration,
	volumeToMount VolumeToMount,
	actualStateOfWorld ActualStateOfWorldMounterUpdater) (volumetypes.GeneratedOperations, error) {

	// Get block volume mapper plugin
	var blockVolumeMapper volume.BlockVolumeMapper
	blockVolumePlugin, err :=
		og.volumePluginMgr.FindMapperPluginBySpec(volumeToMount.VolumeSpec)
	if err != nil {
		return volumetypes.GeneratedOperations{}, volumeToMount.GenerateErrorDetailed("MapVolume.FindMapperPluginBySpec failed", err)
	}
	if blockVolumePlugin == nil {
		return volumetypes.GeneratedOperations{}, volumeToMount.GenerateErrorDetailed("MapVolume.FindMapperPluginBySpec failed to find BlockVolumeMapper plugin. Volume plugin is nil.", nil)
	}
	affinityErr := checkNodeAffinity(og, volumeToMount, blockVolumePlugin)
	if affinityErr != nil {
		eventErr, detailedErr := volumeToMount.GenerateError("MapVolume.NodeAffinity check failed", affinityErr)
		og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeWarning, kevents.FailedMountVolume, eventErr.Error())
		return volumetypes.GeneratedOperations{}, detailedErr
	}
	blockVolumeMapper, newMapperErr := blockVolumePlugin.NewBlockVolumeMapper(
		volumeToMount.VolumeSpec,
		volumeToMount.Pod,
		volume.VolumeOptions{})
	if newMapperErr != nil {
		eventErr, detailedErr := volumeToMount.GenerateError("MapVolume.NewBlockVolumeMapper initialization failed", newMapperErr)
		og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeWarning, kevents.FailedMapVolume, eventErr.Error())
		return volumetypes.GeneratedOperations{}, detailedErr
	}

	// Get attacher, if possible
	attachableVolumePlugin, _ :=
		og.volumePluginMgr.FindAttachablePluginBySpec(volumeToMount.VolumeSpec)
	var volumeAttacher volume.Attacher
	if attachableVolumePlugin != nil {
		volumeAttacher, _ = attachableVolumePlugin.NewAttacher()
	}

	mapVolumeFunc := func() (error, error) {
		var devicePath string
		// Set up global map path under the given plugin directory using symbolic link
		globalMapPath, err :=
			blockVolumeMapper.GetGlobalMapPath(volumeToMount.VolumeSpec)
		if err != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError("MapVolume.GetDeviceMountPath failed", err)
		}
		if volumeAttacher != nil {
			// Wait for attachable volumes to finish attaching
			glog.Infof(volumeToMount.GenerateMsgDetailed("MapVolume.WaitForAttach entering", fmt.Sprintf("DevicePath %q", volumeToMount.DevicePath)))

			devicePath, err = volumeAttacher.WaitForAttach(
				volumeToMount.VolumeSpec, volumeToMount.DevicePath, volumeToMount.Pod, waitForAttachTimeout)
			if err != nil {
				// On failure, return error. Caller will log and retry.
				return volumeToMount.GenerateError("MapVolume.WaitForAttach failed", err)
			}

			glog.Infof(volumeToMount.GenerateMsgDetailed("MapVolume.WaitForAttach succeeded", fmt.Sprintf("DevicePath %q", devicePath)))

		}
		// A plugin doesn't have attacher also needs to map device to global map path with SetUpDevice()
		pluginDevicePath, mapErr := blockVolumeMapper.SetUpDevice()
		if mapErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError("MapVolume.SetUp failed", mapErr)
		}
		// Update devicePath for none attachable plugin case
		if len(devicePath) == 0 {
			if len(pluginDevicePath) != 0 {
				devicePath = pluginDevicePath
			} else {
				return volumeToMount.GenerateError("MapVolume failed", fmt.Errorf("Device path of the volume is empty"))
			}
		}
		// Update actual state of world to reflect volume is globally mounted
		markDeviceMappedErr := actualStateOfWorld.MarkDeviceAsMounted(
			volumeToMount.VolumeName, devicePath, globalMapPath)
		if markDeviceMappedErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError("MapVolume.MarkDeviceAsMounted failed", markDeviceMappedErr)
		}

		mapErr = og.blkUtil.MapDevice(devicePath, globalMapPath, string(volumeToMount.Pod.UID))
		if mapErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError("MapVolume.MapDevice failed", mapErr)
		}

		// Device mapping for global map path succeeded
		simpleMsg, detailedMsg := volumeToMount.GenerateMsg("MapVolume.MapDevice succeeded", fmt.Sprintf("globalMapPath %q", globalMapPath))
		verbosity := glog.Level(4)
		og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeNormal, kevents.SuccessfulMountVolume, simpleMsg)
		glog.V(verbosity).Infof(detailedMsg)

		// Map device to pod device map path under the given pod directory using symbolic link
		volumeMapPath, volName := blockVolumeMapper.GetPodDeviceMapPath()
		mapErr = og.blkUtil.MapDevice(devicePath, volumeMapPath, volName)
		if mapErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError("MapVolume.MapDevice failed", mapErr)
		}

		// Take filedescriptor lock to keep a block device opened. Otherwise, there is a case
		// that the block device is silently removed and attached another device with same name.
		// Container runtime can't handler this problem. To avoid unexpected condition fd lock
		// for the block device is required.
		_, err = og.blkUtil.AttachFileDevice(devicePath)
		if err != nil {
			return volumeToMount.GenerateError("MapVolume.AttachFileDevice failed", err)
		}

		// Device mapping for pod device map path succeeded
		simpleMsg, detailedMsg = volumeToMount.GenerateMsg("MapVolume.MapDevice succeeded", fmt.Sprintf("volumeMapPath %q", volumeMapPath))
		verbosity = glog.Level(1)
		og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeNormal, kevents.SuccessfulMountVolume, simpleMsg)
		glog.V(verbosity).Infof(detailedMsg)

		// Update actual state of world
		markVolMountedErr := actualStateOfWorld.MarkVolumeAsMounted(
			volumeToMount.PodName,
			volumeToMount.Pod.UID,
			volumeToMount.VolumeName,
			nil,
			blockVolumeMapper,
			volumeToMount.OuterVolumeSpecName,
			volumeToMount.VolumeGidValue,
			volumeToMount.VolumeSpec)
		if markVolMountedErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError("MapVolume.MarkVolumeAsMounted failed", markVolMountedErr)
		}

		return nil, nil
	}

	eventRecorderFunc := func(err *error) {
		if *err != nil {
			og.recorder.Eventf(volumeToMount.Pod, v1.EventTypeWarning, kevents.FailedMapVolume, (*err).Error())
		}
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     mapVolumeFunc,
		EventRecorderFunc: eventRecorderFunc,
		CompleteFunc:      util.OperationCompleteHook(blockVolumePlugin.GetPluginName(), "map_volume"),
	}, nil
}

// GenerateUnmapVolumeFunc marks volume as unmonuted based on following steps.
// Remove symbolic links from pod device map path dir and  global map path dir.
// Once those cleanups are done, remove pod device map path dir.
// If all steps are completed, the volume is marked as unmounted.
func (og *operationGenerator) GenerateUnmapVolumeFunc(
	volumeToUnmount MountedVolume,
	actualStateOfWorld ActualStateOfWorldMounterUpdater) (volumetypes.GeneratedOperations, error) {

	// Get block volume unmapper plugin
	var blockVolumeUnmapper volume.BlockVolumeUnmapper
	blockVolumePlugin, err :=
		og.volumePluginMgr.FindMapperPluginByName(volumeToUnmount.PluginName)
	if err != nil {
		return volumetypes.GeneratedOperations{}, volumeToUnmount.GenerateErrorDetailed("UnmapVolume.FindMapperPluginByName failed", err)
	}
	if blockVolumePlugin == nil {
		return volumetypes.GeneratedOperations{}, volumeToUnmount.GenerateErrorDetailed("UnmapVolume.FindMapperPluginByName failed to find BlockVolumeMapper plugin. Volume plugin is nil.", nil)
	}
	blockVolumeUnmapper, newUnmapperErr := blockVolumePlugin.NewBlockVolumeUnmapper(
		volumeToUnmount.InnerVolumeSpecName, volumeToUnmount.PodUID)
	if newUnmapperErr != nil {
		return volumetypes.GeneratedOperations{}, volumeToUnmount.GenerateErrorDetailed("UnmapVolume.NewUnmapper failed", newUnmapperErr)
	}

	unmapVolumeFunc := func() (error, error) {
		// Try to unmap volumeName symlink under pod device map path dir
		// pods/{podUid}/volumeDevices/{escapeQualifiedPluginName}/{volumeName}
		podDeviceUnmapPath, volName := blockVolumeUnmapper.GetPodDeviceMapPath()
		unmapDeviceErr := og.blkUtil.UnmapDevice(podDeviceUnmapPath, volName)
		if unmapDeviceErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToUnmount.GenerateError("UnmapVolume.UnmapDevice on pod device map path failed", unmapDeviceErr)
		}
		// Try to unmap podUID symlink under global map path dir
		// plugins/kubernetes.io/{PluginName}/volumeDevices/{volumePluginDependentPath}/{podUID}
		globalUnmapPath := volumeToUnmount.DeviceMountPath
		unmapDeviceErr = og.blkUtil.UnmapDevice(globalUnmapPath, string(volumeToUnmount.PodUID))
		if unmapDeviceErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToUnmount.GenerateError("UnmapVolume.UnmapDevice on global map path failed", unmapDeviceErr)
		}

		glog.Infof(
			"UnmapVolume succeeded for volume %q (OuterVolumeSpecName: %q) pod %q (UID: %q). InnerVolumeSpecName %q. PluginName %q, VolumeGidValue %q",
			volumeToUnmount.VolumeName,
			volumeToUnmount.OuterVolumeSpecName,
			volumeToUnmount.PodName,
			volumeToUnmount.PodUID,
			volumeToUnmount.InnerVolumeSpecName,
			volumeToUnmount.PluginName,
			volumeToUnmount.VolumeGidValue)

		// Update actual state of world
		markVolUnmountedErr := actualStateOfWorld.MarkVolumeAsUnmounted(
			volumeToUnmount.PodName, volumeToUnmount.VolumeName)
		if markVolUnmountedErr != nil {
			// On failure, just log and exit
			glog.Errorf(volumeToUnmount.GenerateErrorDetailed("UnmapVolume.MarkVolumeAsUnmounted failed", markVolUnmountedErr).Error())
		}

		return nil, nil
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     unmapVolumeFunc,
		CompleteFunc:      util.OperationCompleteHook(blockVolumePlugin.GetPluginName(), "unmap_volume"),
		EventRecorderFunc: nil, // nil because we do not want to generate event on error
	}, nil
}

// GenerateUnmapDeviceFunc marks device as unmounted based on following steps.
// Check under globalMapPath dir if there isn't pod's symbolic links in it.
// If symbolick link isn't there, the device isn't referenced from Pods.
// Call plugin TearDownDevice to clean-up device connection, stored data under
// globalMapPath, these operations depend on plugin implementation.
// Once TearDownDevice is completed, remove globalMapPath dir.
// After globalMapPath is removed, fd lock by loopback for the device can
// be released safely because no one can consume the device at this point.
// At last, device open status will be checked just in case.
// If all steps are completed, the device is marked as unmounted.
func (og *operationGenerator) GenerateUnmapDeviceFunc(
	deviceToDetach AttachedVolume,
	actualStateOfWorld ActualStateOfWorldMounterUpdater,
	mounter mount.Interface) (volumetypes.GeneratedOperations, error) {

	blockVolumePlugin, err :=
		og.volumePluginMgr.FindMapperPluginByName(deviceToDetach.PluginName)
	if err != nil {
		return volumetypes.GeneratedOperations{}, deviceToDetach.GenerateErrorDetailed("UnmapDevice.FindMapperPluginByName failed", err)
	}
	if blockVolumePlugin == nil {
		return volumetypes.GeneratedOperations{}, deviceToDetach.GenerateErrorDetailed("UnmapDevice.FindMapperPluginByName failed to find BlockVolumeMapper plugin. Volume plugin is nil.", nil)
	}

	blockVolumeUnmapper, newUnmapperErr := blockVolumePlugin.NewBlockVolumeUnmapper(
		string(deviceToDetach.VolumeName),
		"" /* podUID */)
	if newUnmapperErr != nil {
		return volumetypes.GeneratedOperations{}, deviceToDetach.GenerateErrorDetailed("UnmapDevice.NewUnmapper failed", newUnmapperErr)
	}

	unmapDeviceFunc := func() (error, error) {
		// Search under globalMapPath dir if all symbolic links from pods have been removed already.
		// If symbolick links are there, pods may still refer the volume.
		globalMapPath := deviceToDetach.DeviceMountPath
		refs, err := og.blkUtil.GetDeviceSymlinkRefs(deviceToDetach.DevicePath, globalMapPath)
		if err != nil {
			return deviceToDetach.GenerateError("UnmapDevice.GetDeviceSymlinkRefs check failed", err)
		}
		if len(refs) > 0 {
			err = fmt.Errorf("The device %q is still referenced from other Pods %v", globalMapPath, refs)
			return deviceToDetach.GenerateError("UnmapDevice failed", err)
		}

		// Execute tear down device
		unmapErr := blockVolumeUnmapper.TearDownDevice(globalMapPath, deviceToDetach.DevicePath)
		if unmapErr != nil {
			// On failure, return error. Caller will log and retry.
			return deviceToDetach.GenerateError("UnmapDevice.TearDownDevice failed", unmapErr)
		}

		// Plugin finished TearDownDevice(). Now globalMapPath dir and plugin's stored data
		// on the dir are unnecessary, clean up it.
		removeMapPathErr := og.blkUtil.RemoveMapPath(globalMapPath)
		if removeMapPathErr != nil {
			// On failure, return error. Caller will log and retry.
			return deviceToDetach.GenerateError("UnmapDevice failed", removeMapPathErr)
		}

		// The block volume is not referenced from Pods. Release file descriptor lock.
		glog.V(4).Infof("UnmapDevice: deviceToDetach.DevicePath: %v", deviceToDetach.DevicePath)
		loopPath, err := og.blkUtil.GetLoopDevice(deviceToDetach.DevicePath)
		if err != nil {
			glog.Warningf(deviceToDetach.GenerateMsgDetailed("UnmapDevice: Couldn't find loopback device which takes file descriptor lock", fmt.Sprintf("device path: %q", deviceToDetach.DevicePath)))
		} else {
			err = og.blkUtil.RemoveLoopDevice(loopPath)
			if err != nil {
				return deviceToDetach.GenerateError("UnmapDevice.AttachFileDevice failed", err)
			}
		}

		// Before logging that UnmapDevice succeeded and moving on,
		// use mounter.PathIsDevice to check if the path is a device,
		// if so use mounter.DeviceOpened to check if the device is in use anywhere
		// else on the system. Retry if it returns true.
		deviceOpened, deviceOpenedErr := isDeviceOpened(deviceToDetach, mounter)
		if deviceOpenedErr != nil {
			return nil, deviceOpenedErr
		}
		// The device is still in use elsewhere. Caller will log and retry.
		if deviceOpened {
			return deviceToDetach.GenerateError(
				"UnmapDevice failed",
				fmt.Errorf("the device is in use when it was no longer expected to be in use"))
		}

		glog.Infof(deviceToDetach.GenerateMsgDetailed("UnmapDevice succeeded", ""))

		// Update actual state of world
		markDeviceUnmountedErr := actualStateOfWorld.MarkDeviceAsUnmounted(
			deviceToDetach.VolumeName)
		if markDeviceUnmountedErr != nil {
			// On failure, return error. Caller will log and retry.
			return deviceToDetach.GenerateError("MarkDeviceAsUnmounted failed", markDeviceUnmountedErr)
		}

		return nil, nil
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     unmapDeviceFunc,
		CompleteFunc:      util.OperationCompleteHook(blockVolumePlugin.GetPluginName(), "unmap_device"),
		EventRecorderFunc: nil, // nil because we do not want to generate event on error
	}, nil
}

func (og *operationGenerator) GenerateVerifyControllerAttachedVolumeFunc(
	volumeToMount VolumeToMount,
	nodeName types.NodeName,
	actualStateOfWorld ActualStateOfWorldAttacherUpdater) (volumetypes.GeneratedOperations, error) {
	volumePlugin, err :=
		og.volumePluginMgr.FindPluginBySpec(volumeToMount.VolumeSpec)
	if err != nil || volumePlugin == nil {
		return volumetypes.GeneratedOperations{}, volumeToMount.GenerateErrorDetailed("VerifyControllerAttachedVolume.FindPluginBySpec failed", err)
	}

	verifyControllerAttachedVolumeFunc := func() (error, error) {
		if !volumeToMount.PluginIsAttachable {
			// If the volume does not implement the attacher interface, it is
			// assumed to be attached and the actual state of the world is
			// updated accordingly.

			addVolumeNodeErr := actualStateOfWorld.MarkVolumeAsAttached(
				volumeToMount.VolumeName, volumeToMount.VolumeSpec, nodeName, "" /* devicePath */)
			if addVolumeNodeErr != nil {
				// On failure, return error. Caller will log and retry.
				return volumeToMount.GenerateError("VerifyControllerAttachedVolume.MarkVolumeAsAttachedByUniqueVolumeName failed", addVolumeNodeErr)
			}

			return nil, nil
		}

		if !volumeToMount.ReportedInUse {
			// If the given volume has not yet been added to the list of
			// VolumesInUse in the node's volume status, do not proceed, return
			// error. Caller will log and retry. The node status is updated
			// periodically by kubelet, so it may take as much as 10 seconds
			// before this clears.
			// Issue #28141 to enable on demand status updates.
			return volumeToMount.GenerateError("Volume has not been added to the list of VolumesInUse in the node's volume status", nil)
		}

		// Fetch current node object
		node, fetchErr := og.kubeClient.CoreV1().Nodes().Get(string(nodeName), metav1.GetOptions{})
		if fetchErr != nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError("VerifyControllerAttachedVolume failed fetching node from API server", fetchErr)
		}

		if node == nil {
			// On failure, return error. Caller will log and retry.
			return volumeToMount.GenerateError(
				"VerifyControllerAttachedVolume failed",
				fmt.Errorf("Node object retrieved from API server is nil"))
		}

		for _, attachedVolume := range node.Status.VolumesAttached {
			if attachedVolume.Name == volumeToMount.VolumeName {
				addVolumeNodeErr := actualStateOfWorld.MarkVolumeAsAttached(
					v1.UniqueVolumeName(""), volumeToMount.VolumeSpec, nodeName, attachedVolume.DevicePath)
				glog.Infof(volumeToMount.GenerateMsgDetailed("Controller attach succeeded", fmt.Sprintf("device path: %q", attachedVolume.DevicePath)))
				if addVolumeNodeErr != nil {
					// On failure, return error. Caller will log and retry.
					return volumeToMount.GenerateError("VerifyControllerAttachedVolume.MarkVolumeAsAttached failed", addVolumeNodeErr)
				}
				return nil, nil
			}
		}

		// Volume not attached, return error. Caller will log and retry.
		return volumeToMount.GenerateError("Volume not attached according to node status", nil)
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     verifyControllerAttachedVolumeFunc,
		CompleteFunc:      util.OperationCompleteHook(volumePlugin.GetPluginName(), "verify_controller_attached_volume"),
		EventRecorderFunc: nil, // nil because we do not want to generate event on error
	}, nil

}

func (og *operationGenerator) verifyVolumeIsSafeToDetach(
	volumeToDetach AttachedVolume) error {
	// Fetch current node object
	node, fetchErr := og.kubeClient.CoreV1().Nodes().Get(string(volumeToDetach.NodeName), metav1.GetOptions{})
	if fetchErr != nil {
		if errors.IsNotFound(fetchErr) {
			glog.Warningf(volumeToDetach.GenerateMsgDetailed("Node not found on API server. DetachVolume will skip safe to detach check", ""))
			return nil
		}

		// On failure, return error. Caller will log and retry.
		return volumeToDetach.GenerateErrorDetailed("DetachVolume failed fetching node from API server", fetchErr)
	}

	if node == nil {
		// On failure, return error. Caller will log and retry.
		return volumeToDetach.GenerateErrorDetailed(
			"DetachVolume failed fetching node from API server",
			fmt.Errorf("node object retrieved from API server is nil"))
	}

	for _, inUseVolume := range node.Status.VolumesInUse {
		if inUseVolume == volumeToDetach.VolumeName {
			return volumeToDetach.GenerateErrorDetailed(
				"DetachVolume failed",
				fmt.Errorf("volume is still in use by node, according to Node status"))
		}
	}

	// Volume is not marked as in use by node
	glog.Infof(volumeToDetach.GenerateMsgDetailed("Verified volume is safe to detach", ""))
	return nil
}

func (og *operationGenerator) GenerateExpandVolumeFunc(
	pvcWithResizeRequest *expandcache.PVCWithResizeRequest,
	resizeMap expandcache.VolumeResizeMap) (volumetypes.GeneratedOperations, error) {

	volumeSpec := volume.NewSpecFromPersistentVolume(pvcWithResizeRequest.PersistentVolume, false)

	volumePlugin, err := og.volumePluginMgr.FindExpandablePluginBySpec(volumeSpec)

	if err != nil {
		return volumetypes.GeneratedOperations{}, fmt.Errorf("Error finding plugin for expanding volume: %q with error %v", pvcWithResizeRequest.QualifiedName(), err)
	}
	if volumePlugin == nil {
		return volumetypes.GeneratedOperations{}, fmt.Errorf("Can not find plugin for expanding volume: %q", pvcWithResizeRequest.QualifiedName())
	}

	expandVolumeFunc := func() (error, error) {
		newSize := pvcWithResizeRequest.ExpectedSize
		pvSize := pvcWithResizeRequest.PersistentVolume.Spec.Capacity[v1.ResourceStorage]
		if pvSize.Cmp(newSize) < 0 {
			updatedSize, expandErr := volumePlugin.ExpandVolumeDevice(
				volumeSpec,
				pvcWithResizeRequest.ExpectedSize,
				pvcWithResizeRequest.CurrentSize)

			if expandErr != nil {
				detailedErr := fmt.Errorf("Error expanding volume %q of plugin %s : %v", pvcWithResizeRequest.QualifiedName(), volumePlugin.GetPluginName(), expandErr)
				return detailedErr, detailedErr
			}
			glog.Infof("ExpandVolume succeeded for volume %s", pvcWithResizeRequest.QualifiedName())
			newSize = updatedSize
			// k8s doesn't have transactions, we can't guarantee that after updating PV - updating PVC will be
			// successful, that is why all PVCs for which pvc.Spec.Size > pvc.Status.Size must be reprocessed
			// until they reflect user requested size in pvc.Status.Size
			updateErr := resizeMap.UpdatePVSize(pvcWithResizeRequest, newSize)

			if updateErr != nil {
				detailedErr := fmt.Errorf("Error updating PV spec capacity for volume %q with : %v", pvcWithResizeRequest.QualifiedName(), updateErr)
				return detailedErr, detailedErr
			}
			glog.Infof("ExpandVolume.UpdatePV succeeded for volume %s", pvcWithResizeRequest.QualifiedName())
		}

		// No Cloudprovider resize needed, lets mark resizing as done
		// Rest of the volume expand controller code will assume PVC as *not* resized until pvc.Status.Size
		// reflects user requested size.
		if !volumePlugin.RequiresFSResize() {
			glog.V(4).Infof("Controller resizing done for PVC %s", pvcWithResizeRequest.QualifiedName())
			err := resizeMap.MarkAsResized(pvcWithResizeRequest, newSize)

			if err != nil {
				detailedErr := fmt.Errorf("Error marking pvc %s as resized : %v", pvcWithResizeRequest.QualifiedName(), err)
				return detailedErr, detailedErr
			}
			successMsg := fmt.Sprintf("ExpandVolume succeeded for volume %s", pvcWithResizeRequest.QualifiedName())
			og.recorder.Eventf(pvcWithResizeRequest.PVC, v1.EventTypeNormal, kevents.VolumeResizeSuccess, successMsg)
		} else {
			err := resizeMap.MarkForFSResize(pvcWithResizeRequest)
			if err != nil {
				detailedErr := fmt.Errorf("Error updating pvc %s condition for fs resize : %v", pvcWithResizeRequest.QualifiedName(), err)
				glog.Warning(detailedErr)
				return nil, nil
			}
		}
		return nil, nil
	}

	eventRecorderFunc := func(err *error) {
		if *err != nil {
			og.recorder.Eventf(pvcWithResizeRequest.PVC, v1.EventTypeWarning, kevents.VolumeResizeFailed, (*err).Error())
		}
	}

	return volumetypes.GeneratedOperations{
		OperationFunc:     expandVolumeFunc,
		EventRecorderFunc: eventRecorderFunc,
		CompleteFunc:      util.OperationCompleteHook(volumePlugin.GetPluginName(), "expand_volume"),
	}, nil
}

func checkMountOptionSupport(og *operationGenerator, volumeToMount VolumeToMount, plugin volume.VolumePlugin) error {
	mountOptions := util.MountOptionFromSpec(volumeToMount.VolumeSpec)

	if len(mountOptions) > 0 && !plugin.SupportsMountOption() {
		return fmt.Errorf("Mount options are not supported for this volume type")
	}
	return nil
}

// checkNodeAffinity looks at the PV node affinity, and checks if the node has the same corresponding labels
// This ensures that we don't mount a volume that doesn't belong to this node
func checkNodeAffinity(og *operationGenerator, volumeToMount VolumeToMount, plugin volume.VolumePlugin) error {
	if !utilfeature.DefaultFeatureGate.Enabled(features.PersistentLocalVolumes) {
		return nil
	}

	pv := volumeToMount.VolumeSpec.PersistentVolume
	if pv != nil {
		nodeLabels, err := og.volumePluginMgr.Host.GetNodeLabels()
		if err != nil {
			return err
		}
		err = util.CheckNodeAffinity(pv, nodeLabels)
		if err != nil {
			return err
		}
	}
	return nil
}

// isDeviceOpened checks the device status if the device is in use anywhere else on the system
func isDeviceOpened(deviceToDetach AttachedVolume, mounter mount.Interface) (bool, error) {
	isDevicePath, devicePathErr := mounter.PathIsDevice(deviceToDetach.DevicePath)
	var deviceOpened bool
	var deviceOpenedErr error
	if !isDevicePath && devicePathErr == nil ||
		(devicePathErr != nil && strings.Contains(devicePathErr.Error(), "does not exist")) {
		// not a device path or path doesn't exist
		//TODO: refer to #36092
		glog.V(3).Infof("The path isn't device path or doesn't exist. Skip checking device path: %s", deviceToDetach.DevicePath)
		deviceOpened = false
	} else {
		deviceOpened, deviceOpenedErr = mounter.DeviceOpened(deviceToDetach.DevicePath)
		if deviceOpenedErr != nil {
			return false, deviceToDetach.GenerateErrorDetailed("DeviceOpened failed", deviceOpenedErr)
		}
	}
	return deviceOpened, nil
}
