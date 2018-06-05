/*
Copyright 2015 The Kubernetes Authors.

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

package fc

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/util/mount"
	utilstrings "k8s.io/kubernetes/pkg/util/strings"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"
	"k8s.io/kubernetes/pkg/volume/util/volumepathhandler"
)

// This is the primary entrypoint for volume plugins.
func ProbeVolumePlugins() []volume.VolumePlugin {
	return []volume.VolumePlugin{&fcPlugin{nil}}
}

type fcPlugin struct {
	host volume.VolumeHost
}

var _ volume.VolumePlugin = &fcPlugin{}
var _ volume.PersistentVolumePlugin = &fcPlugin{}
var _ volume.BlockVolumePlugin = &fcPlugin{}

const (
	fcPluginName = "kubernetes.io/fc"
)

func (plugin *fcPlugin) Init(host volume.VolumeHost) error {
	plugin.host = host
	return nil
}

func (plugin *fcPlugin) GetPluginName() string {
	return fcPluginName
}

func (plugin *fcPlugin) GetVolumeName(spec *volume.Spec) (string, error) {
	volumeSource, _, err := getVolumeSource(spec)
	if err != nil {
		return "", err
	}

	// API server validates these parameters beforehand but attach/detach
	// controller creates volumespec without validation. They may be nil
	// or zero length. We should check again to avoid unexpected conditions.
	if len(volumeSource.TargetWWNs) != 0 && volumeSource.Lun != nil {
		// TargetWWNs are the FibreChannel target worldwide names
		return fmt.Sprintf("%v:%v", volumeSource.TargetWWNs, *volumeSource.Lun), nil
	} else if len(volumeSource.WWIDs) != 0 {
		// WWIDs are the FibreChannel World Wide Identifiers
		return fmt.Sprintf("%v", volumeSource.WWIDs), nil
	}

	return "", err
}

func (plugin *fcPlugin) CanSupport(spec *volume.Spec) bool {
	return (spec.Volume != nil && spec.Volume.FC != nil) || (spec.PersistentVolume != nil && spec.PersistentVolume.Spec.FC != nil)
}

func (plugin *fcPlugin) RequiresRemount() bool {
	return false
}

func (plugin *fcPlugin) SupportsMountOption() bool {
	return false
}

func (plugin *fcPlugin) SupportsBulkVolumeVerification() bool {
	return false
}

func (plugin *fcPlugin) GetAccessModes() []v1.PersistentVolumeAccessMode {
	return []v1.PersistentVolumeAccessMode{
		v1.ReadWriteOnce,
		v1.ReadOnlyMany,
	}
}

func (plugin *fcPlugin) NewMounter(spec *volume.Spec, pod *v1.Pod, _ volume.VolumeOptions) (volume.Mounter, error) {
	// Inject real implementations here, test through the internal function.
	return plugin.newMounterInternal(spec, pod.UID, &FCUtil{}, plugin.host.GetMounter(plugin.GetPluginName()), plugin.host.GetExec(plugin.GetPluginName()))
}

func (plugin *fcPlugin) newMounterInternal(spec *volume.Spec, podUID types.UID, manager diskManager, mounter mount.Interface, exec mount.Exec) (volume.Mounter, error) {
	// fc volumes used directly in a pod have a ReadOnly flag set by the pod author.
	// fc volumes used as a PersistentVolume gets the ReadOnly flag indirectly through the persistent-claim volume used to mount the PV
	fc, readOnly, err := getVolumeSource(spec)
	if err != nil {
		return nil, err
	}

	wwns, lun, wwids, err := getWwnsLunWwids(fc)
	if err != nil {
		return nil, fmt.Errorf("fc: no fc disk information found. failed to make a new mounter")
	}
	fcDisk := &fcDisk{
		podUID:  podUID,
		volName: spec.Name(),
		wwns:    wwns,
		lun:     lun,
		wwids:   wwids,
		manager: manager,
		io:      &osIOHandler{},
		plugin:  plugin,
	}
	// TODO: remove feature gate check after no longer needed
	if utilfeature.DefaultFeatureGate.Enabled(features.BlockVolume) {
		volumeMode, err := util.GetVolumeMode(spec)
		if err != nil {
			return nil, err
		}
		glog.V(5).Infof("fc: newMounterInternal volumeMode %s", volumeMode)
		return &fcDiskMounter{
			fcDisk:     fcDisk,
			fsType:     fc.FSType,
			volumeMode: volumeMode,
			readOnly:   readOnly,
			mounter:    &mount.SafeFormatAndMount{Interface: mounter, Exec: exec},
			deviceUtil: util.NewDeviceHandler(util.NewIOHandler()),
		}, nil
	}
	return &fcDiskMounter{
		fcDisk:     fcDisk,
		fsType:     fc.FSType,
		readOnly:   readOnly,
		mounter:    &mount.SafeFormatAndMount{Interface: mounter, Exec: exec},
		deviceUtil: util.NewDeviceHandler(util.NewIOHandler()),
	}, nil

}

func (plugin *fcPlugin) NewBlockVolumeMapper(spec *volume.Spec, pod *v1.Pod, _ volume.VolumeOptions) (volume.BlockVolumeMapper, error) {
	// If this called via GenerateUnmapDeviceFunc(), pod is nil.
	// Pass empty string as dummy uid since uid isn't used in the case.
	var uid types.UID
	if pod != nil {
		uid = pod.UID
	}
	return plugin.newBlockVolumeMapperInternal(spec, uid, &FCUtil{}, plugin.host.GetMounter(plugin.GetPluginName()), plugin.host.GetExec(plugin.GetPluginName()))
}

func (plugin *fcPlugin) newBlockVolumeMapperInternal(spec *volume.Spec, podUID types.UID, manager diskManager, mounter mount.Interface, exec mount.Exec) (volume.BlockVolumeMapper, error) {
	fc, readOnly, err := getVolumeSource(spec)
	if err != nil {
		return nil, err
	}

	wwns, lun, wwids, err := getWwnsLunWwids(fc)
	if err != nil {
		return nil, fmt.Errorf("fc: no fc disk information found. failed to make a new mapper")
	}

	return &fcDiskMapper{
		fcDisk: &fcDisk{
			podUID:  podUID,
			volName: spec.Name(),
			wwns:    wwns,
			lun:     lun,
			wwids:   wwids,
			manager: manager,
			io:      &osIOHandler{},
			plugin:  plugin},
		readOnly:   readOnly,
		mounter:    &mount.SafeFormatAndMount{Interface: mounter, Exec: exec},
		deviceUtil: util.NewDeviceHandler(util.NewIOHandler()),
	}, nil
}

func (plugin *fcPlugin) NewUnmounter(volName string, podUID types.UID) (volume.Unmounter, error) {
	// Inject real implementations here, test through the internal function.
	return plugin.newUnmounterInternal(volName, podUID, &FCUtil{}, plugin.host.GetMounter(plugin.GetPluginName()))
}

func (plugin *fcPlugin) newUnmounterInternal(volName string, podUID types.UID, manager diskManager, mounter mount.Interface) (volume.Unmounter, error) {
	return &fcDiskUnmounter{
		fcDisk: &fcDisk{
			podUID:  podUID,
			volName: volName,
			manager: manager,
			plugin:  plugin,
			io:      &osIOHandler{},
		},
		mounter:    mounter,
		deviceUtil: util.NewDeviceHandler(util.NewIOHandler()),
	}, nil
}

func (plugin *fcPlugin) NewBlockVolumeUnmapper(volName string, podUID types.UID) (volume.BlockVolumeUnmapper, error) {
	return plugin.newUnmapperInternal(volName, podUID, &FCUtil{})
}

func (plugin *fcPlugin) newUnmapperInternal(volName string, podUID types.UID, manager diskManager) (volume.BlockVolumeUnmapper, error) {
	return &fcDiskUnmapper{
		fcDisk: &fcDisk{
			podUID:  podUID,
			volName: volName,
			manager: manager,
			plugin:  plugin,
			io:      &osIOHandler{},
		},
		deviceUtil: util.NewDeviceHandler(util.NewIOHandler()),
	}, nil
}

func (plugin *fcPlugin) ConstructVolumeSpec(volumeName, mountPath string) (*volume.Spec, error) {
	// Find globalPDPath from pod volume directory(mountPath)
	// examples:
	//   mountPath:     pods/{podUid}/volumes/kubernetes.io~fc/{volumeName}
	//   globalPDPath : plugins/kubernetes.io/fc/50060e801049cfd1-lun-0
	var globalPDPath string
	mounter := plugin.host.GetMounter(plugin.GetPluginName())
	paths, err := mount.GetMountRefs(mounter, mountPath)
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		if strings.Contains(path, plugin.host.GetPluginDir(fcPluginName)) {
			globalPDPath = path
			break
		}
	}
	// Couldn't fetch globalPDPath
	if len(globalPDPath) == 0 {
		return nil, fmt.Errorf("couldn't fetch globalPDPath. failed to obtain volume spec")
	}
	arr := strings.Split(globalPDPath, "/")
	if len(arr) < 1 {
		return nil, fmt.Errorf("failed to retrieve volume plugin information from globalPDPath: %v", globalPDPath)
	}
	volumeInfo := arr[len(arr)-1]
	// Create volume from wwn+lun or wwid
	var fcVolume *v1.Volume
	if strings.Contains(volumeInfo, "-lun-") {
		wwnLun := strings.Split(volumeInfo, "-lun-")
		if len(wwnLun) < 2 {
			return nil, fmt.Errorf("failed to retrieve TargetWWN and Lun. volumeInfo is invalid: %v", volumeInfo)
		}
		lun, err := strconv.Atoi(wwnLun[1])
		if err != nil {
			return nil, err
		}
		lun32 := int32(lun)
		fcVolume = &v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				FC: &v1.FCVolumeSource{TargetWWNs: []string{wwnLun[0]}, Lun: &lun32},
			},
		}
		glog.V(5).Infof("ConstructVolumeSpec: TargetWWNs: %v, Lun: %v",
			fcVolume.VolumeSource.FC.TargetWWNs, *fcVolume.VolumeSource.FC.Lun)
	} else {
		fcVolume = &v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				FC: &v1.FCVolumeSource{WWIDs: []string{volumeInfo}},
			},
		}
		glog.V(5).Infof("ConstructVolumeSpec: WWIDs: %v", fcVolume.VolumeSource.FC.WWIDs)
	}
	return volume.NewSpecFromVolume(fcVolume), nil
}

// ConstructBlockVolumeSpec creates a new volume.Spec with following steps.
//   - Searches a file whose name is {pod uuid} under volume plugin directory.
//   - If a file is found, then retreives volumePluginDependentPath from globalMapPathUUID.
//   - Once volumePluginDependentPath is obtained, store volume information to VolumeSource
// examples:
//   mapPath: pods/{podUid}/{DefaultKubeletVolumeDevicesDirName}/{escapeQualifiedPluginName}/{volumeName}
//   globalMapPathUUID : plugins/kubernetes.io/{PluginName}/{DefaultKubeletVolumeDevicesDirName}/{volumePluginDependentPath}/{pod uuid}
func (plugin *fcPlugin) ConstructBlockVolumeSpec(podUID types.UID, volumeName, mapPath string) (*volume.Spec, error) {
	pluginDir := plugin.host.GetVolumeDevicePluginDir(fcPluginName)
	blkutil := volumepathhandler.NewBlockVolumePathHandler()
	globalMapPathUUID, err := blkutil.FindGlobalMapPathUUIDFromPod(pluginDir, mapPath, podUID)
	if err != nil {
		return nil, err
	}
	glog.V(5).Infof("globalMapPathUUID: %v, err: %v", globalMapPathUUID, err)

	// Retrieve volumePluginDependentPath from globalMapPathUUID
	// globalMapPathUUID examples:
	//   wwn+lun: plugins/kubernetes.io/fc/volumeDevices/50060e801049cfd1-lun-0/{pod uuid}
	//   wwid: plugins/kubernetes.io/fc/volumeDevices/3600508b400105e210000900000490000/{pod uuid}
	arr := strings.Split(globalMapPathUUID, "/")
	if len(arr) < 2 {
		return nil, fmt.Errorf("Fail to retrieve volume plugin information from globalMapPathUUID: %v", globalMapPathUUID)
	}
	l := len(arr) - 2
	volumeInfo := arr[l]

	// Create volume from wwn+lun or wwid
	var fcPV *v1.PersistentVolume
	if strings.Contains(volumeInfo, "-lun-") {
		wwnLun := strings.Split(volumeInfo, "-lun-")
		lun, err := strconv.Atoi(wwnLun[1])
		if err != nil {
			return nil, err
		}
		lun32 := int32(lun)
		fcPV = createPersistentVolumeFromFCVolumeSource(volumeName,
			v1.FCVolumeSource{TargetWWNs: []string{wwnLun[0]}, Lun: &lun32})
		glog.V(5).Infof("ConstructBlockVolumeSpec: TargetWWNs: %v, Lun: %v",
			fcPV.Spec.PersistentVolumeSource.FC.TargetWWNs,
			*fcPV.Spec.PersistentVolumeSource.FC.Lun)
	} else {
		fcPV = createPersistentVolumeFromFCVolumeSource(volumeName,
			v1.FCVolumeSource{WWIDs: []string{volumeInfo}})
		glog.V(5).Infof("ConstructBlockVolumeSpec: WWIDs: %v", fcPV.Spec.PersistentVolumeSource.FC.WWIDs)
	}
	return volume.NewSpecFromPersistentVolume(fcPV, false), nil
}

type fcDisk struct {
	volName string
	podUID  types.UID
	portal  string
	wwns    []string
	lun     string
	wwids   []string
	plugin  *fcPlugin
	// Utility interface that provides API calls to the provider to attach/detach disks.
	manager diskManager
	// io handler interface
	io ioHandler
	volume.MetricsNil
}

func (fc *fcDisk) GetPath() string {
	name := fcPluginName
	// safe to use PodVolumeDir now: volume teardown occurs before pod is cleaned up
	return fc.plugin.host.GetPodVolumeDir(fc.podUID, utilstrings.EscapeQualifiedNameForDisk(name), fc.volName)
}

func (fc *fcDisk) fcGlobalMapPath(spec *volume.Spec) (string, error) {
	mounter, err := volumeSpecToMounter(spec, fc.plugin.host)
	if err != nil {
		glog.Warningf("failed to get fc mounter: %v", err)
		return "", err
	}
	return fc.manager.MakeGlobalVDPDName(*mounter.fcDisk), nil
}

func (fc *fcDisk) fcPodDeviceMapPath() (string, string) {
	name := fcPluginName
	return fc.plugin.host.GetPodVolumeDeviceDir(fc.podUID, utilstrings.EscapeQualifiedNameForDisk(name)), fc.volName
}

type fcDiskMounter struct {
	*fcDisk
	readOnly   bool
	fsType     string
	volumeMode v1.PersistentVolumeMode
	mounter    *mount.SafeFormatAndMount
	deviceUtil util.DeviceUtil
}

var _ volume.Mounter = &fcDiskMounter{}

func (b *fcDiskMounter) GetAttributes() volume.Attributes {
	return volume.Attributes{
		ReadOnly:        b.readOnly,
		Managed:         !b.readOnly,
		SupportsSELinux: true,
	}
}

// Checks prior to mount operations to verify that the required components (binaries, etc.)
// to mount the volume are available on the underlying node.
// If not, it returns an error
func (b *fcDiskMounter) CanMount() error {
	return nil
}

func (b *fcDiskMounter) SetUp(fsGroup *int64) error {
	return b.SetUpAt(b.GetPath(), fsGroup)
}

func (b *fcDiskMounter) SetUpAt(dir string, fsGroup *int64) error {
	// diskSetUp checks mountpoints and prevent repeated calls
	err := diskSetUp(b.manager, *b, dir, b.mounter, fsGroup)
	if err != nil {
		glog.Errorf("fc: failed to setup")
	}
	return err
}

type fcDiskUnmounter struct {
	*fcDisk
	mounter    mount.Interface
	deviceUtil util.DeviceUtil
}

var _ volume.Unmounter = &fcDiskUnmounter{}

// Unmounts the bind mount, and detaches the disk only if the disk
// resource was the last reference to that disk on the kubelet.
func (c *fcDiskUnmounter) TearDown() error {
	return c.TearDownAt(c.GetPath())
}

func (c *fcDiskUnmounter) TearDownAt(dir string) error {
	return util.UnmountPath(dir, c.mounter)
}

// Block Volumes Support
type fcDiskMapper struct {
	*fcDisk
	readOnly   bool
	mounter    mount.Interface
	deviceUtil util.DeviceUtil
}

var _ volume.BlockVolumeMapper = &fcDiskMapper{}

func (b *fcDiskMapper) SetUpDevice() (string, error) {
	return "", nil
}

type fcDiskUnmapper struct {
	*fcDisk
	deviceUtil util.DeviceUtil
}

var _ volume.BlockVolumeUnmapper = &fcDiskUnmapper{}

func (c *fcDiskUnmapper) TearDownDevice(mapPath, devicePath string) error {
	err := c.manager.DetachBlockFCDisk(*c, mapPath, devicePath)
	if err != nil {
		return fmt.Errorf("fc: failed to detach disk: %s\nError: %v", mapPath, err)
	}
	glog.V(4).Infof("fc: %q is unmounted, deleting the directory", mapPath)
	err = os.RemoveAll(mapPath)
	if err != nil {
		return fmt.Errorf("fc: failed to delete the directory: %s\nError: %v", mapPath, err)
	}
	glog.V(4).Infof("fc: successfully detached disk: %s", mapPath)
	return nil
}

// GetGlobalMapPath returns global map path and error
// path: plugins/kubernetes.io/{PluginName}/volumeDevices/{WWID}/{podUid}
func (fc *fcDisk) GetGlobalMapPath(spec *volume.Spec) (string, error) {
	return fc.fcGlobalMapPath(spec)
}

// GetPodDeviceMapPath returns pod device map path and volume name
// path: pods/{podUid}/volumeDevices/kubernetes.io~fc
// volumeName: pv0001
func (fc *fcDisk) GetPodDeviceMapPath() (string, string) {
	return fc.fcPodDeviceMapPath()
}

func getVolumeSource(spec *volume.Spec) (*v1.FCVolumeSource, bool, error) {
	// fc volumes used directly in a pod have a ReadOnly flag set by the pod author.
	// fc volumes used as a PersistentVolume gets the ReadOnly flag indirectly through the persistent-claim volume used to mount the PV
	if spec.Volume != nil && spec.Volume.FC != nil {
		return spec.Volume.FC, spec.Volume.FC.ReadOnly, nil
	} else if spec.PersistentVolume != nil &&
		spec.PersistentVolume.Spec.FC != nil {
		return spec.PersistentVolume.Spec.FC, spec.ReadOnly, nil
	}

	return nil, false, fmt.Errorf("Spec does not reference a FibreChannel volume type")
}

func createPersistentVolumeFromFCVolumeSource(volumeName string, fc v1.FCVolumeSource) *v1.PersistentVolume {
	block := v1.PersistentVolumeBlock
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: volumeName,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FC: &fc,
			},
			VolumeMode: &block,
		},
	}
}

func getWwnsLunWwids(fc *v1.FCVolumeSource) ([]string, string, []string, error) {
	var lun string
	var wwids []string
	if fc.Lun != nil && len(fc.TargetWWNs) != 0 {
		lun = strconv.Itoa(int(*fc.Lun))
		return fc.TargetWWNs, lun, wwids, nil
	}
	if len(fc.WWIDs) != 0 {
		for _, wwid := range fc.WWIDs {
			wwids = append(wwids, strings.Replace(wwid, " ", "_", -1))
		}
		return fc.TargetWWNs, lun, wwids, nil
	}
	return nil, "", nil, fmt.Errorf("fc: no fc disk information found")
}
