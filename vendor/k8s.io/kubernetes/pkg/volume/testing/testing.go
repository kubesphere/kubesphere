/*
Copyright 2014 The Kubernetes Authors.

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

package testing

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	utiltesting "k8s.io/client-go/util/testing"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/util/io"
	"k8s.io/kubernetes/pkg/util/mount"
	utilstrings "k8s.io/kubernetes/pkg/util/strings"
	. "k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"
	"k8s.io/kubernetes/pkg/volume/util/recyclerclient"
	"k8s.io/kubernetes/pkg/volume/util/volumepathhandler"
)

// fakeVolumeHost is useful for testing volume plugins.
type fakeVolumeHost struct {
	rootDir    string
	kubeClient clientset.Interface
	pluginMgr  VolumePluginMgr
	cloud      cloudprovider.Interface
	mounter    mount.Interface
	exec       mount.Exec
	writer     io.Writer
	nodeLabels map[string]string
	nodeName   string
}

func NewFakeVolumeHost(rootDir string, kubeClient clientset.Interface, plugins []VolumePlugin) *fakeVolumeHost {
	return newFakeVolumeHost(rootDir, kubeClient, plugins, nil)
}

func NewFakeVolumeHostWithCloudProvider(rootDir string, kubeClient clientset.Interface, plugins []VolumePlugin, cloud cloudprovider.Interface) *fakeVolumeHost {
	return newFakeVolumeHost(rootDir, kubeClient, plugins, cloud)
}

func NewFakeVolumeHostWithNodeLabels(rootDir string, kubeClient clientset.Interface, plugins []VolumePlugin, labels map[string]string) *fakeVolumeHost {
	volHost := newFakeVolumeHost(rootDir, kubeClient, plugins, nil)
	volHost.nodeLabels = labels
	return volHost
}

func NewFakeVolumeHostWithNodeName(rootDir string, kubeClient clientset.Interface, plugins []VolumePlugin, nodeName string) *fakeVolumeHost {
	volHost := newFakeVolumeHost(rootDir, kubeClient, plugins, nil)
	volHost.nodeName = nodeName
	return volHost
}

func newFakeVolumeHost(rootDir string, kubeClient clientset.Interface, plugins []VolumePlugin, cloud cloudprovider.Interface) *fakeVolumeHost {
	host := &fakeVolumeHost{rootDir: rootDir, kubeClient: kubeClient, cloud: cloud}
	host.mounter = &mount.FakeMounter{}
	host.writer = &io.StdWriter{}
	host.exec = mount.NewFakeExec(nil)
	host.pluginMgr.InitPlugins(plugins, nil /* prober */, host)
	return host
}

func (f *fakeVolumeHost) GetPluginDir(podUID string) string {
	return path.Join(f.rootDir, "plugins", podUID)
}

func (f *fakeVolumeHost) GetVolumeDevicePluginDir(pluginName string) string {
	return path.Join(f.rootDir, "plugins", pluginName, "volumeDevices")
}

func (f *fakeVolumeHost) GetPodsDir() string {
	return path.Join(f.rootDir, "pods")
}

func (f *fakeVolumeHost) GetPodVolumeDir(podUID types.UID, pluginName, volumeName string) string {
	return path.Join(f.rootDir, "pods", string(podUID), "volumes", pluginName, volumeName)
}

func (f *fakeVolumeHost) GetPodVolumeDeviceDir(podUID types.UID, pluginName string) string {
	return path.Join(f.rootDir, "pods", string(podUID), "volumeDevices", pluginName)
}

func (f *fakeVolumeHost) GetPodPluginDir(podUID types.UID, pluginName string) string {
	return path.Join(f.rootDir, "pods", string(podUID), "plugins", pluginName)
}

func (f *fakeVolumeHost) GetKubeClient() clientset.Interface {
	return f.kubeClient
}

func (f *fakeVolumeHost) GetCloudProvider() cloudprovider.Interface {
	return f.cloud
}

func (f *fakeVolumeHost) GetMounter(pluginName string) mount.Interface {
	return f.mounter
}

func (f *fakeVolumeHost) GetWriter() io.Writer {
	return f.writer
}

func (f *fakeVolumeHost) NewWrapperMounter(volName string, spec Spec, pod *v1.Pod, opts VolumeOptions) (Mounter, error) {
	// The name of wrapper volume is set to "wrapped_{wrapped_volume_name}"
	wrapperVolumeName := "wrapped_" + volName
	if spec.Volume != nil {
		spec.Volume.Name = wrapperVolumeName
	}
	plug, err := f.pluginMgr.FindPluginBySpec(&spec)
	if err != nil {
		return nil, err
	}
	return plug.NewMounter(&spec, pod, opts)
}

func (f *fakeVolumeHost) NewWrapperUnmounter(volName string, spec Spec, podUID types.UID) (Unmounter, error) {
	// The name of wrapper volume is set to "wrapped_{wrapped_volume_name}"
	wrapperVolumeName := "wrapped_" + volName
	if spec.Volume != nil {
		spec.Volume.Name = wrapperVolumeName
	}
	plug, err := f.pluginMgr.FindPluginBySpec(&spec)
	if err != nil {
		return nil, err
	}
	return plug.NewUnmounter(spec.Name(), podUID)
}

// Returns the hostname of the host kubelet is running on
func (f *fakeVolumeHost) GetHostName() string {
	return "fakeHostName"
}

// Returns host IP or nil in the case of error.
func (f *fakeVolumeHost) GetHostIP() (net.IP, error) {
	return nil, fmt.Errorf("GetHostIP() not implemented")
}

func (f *fakeVolumeHost) GetNodeAllocatable() (v1.ResourceList, error) {
	return v1.ResourceList{}, nil
}

func (f *fakeVolumeHost) GetSecretFunc() func(namespace, name string) (*v1.Secret, error) {
	return func(namespace, name string) (*v1.Secret, error) {
		return f.kubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	}
}

func (f *fakeVolumeHost) GetExec(pluginName string) mount.Exec {
	return f.exec
}

func (f *fakeVolumeHost) GetConfigMapFunc() func(namespace, name string) (*v1.ConfigMap, error) {
	return func(namespace, name string) (*v1.ConfigMap, error) {
		return f.kubeClient.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	}
}

func (f *fakeVolumeHost) GetNodeLabels() (map[string]string, error) {
	if f.nodeLabels == nil {
		f.nodeLabels = map[string]string{"test-label": "test-value"}
	}
	return f.nodeLabels, nil
}

func (f *fakeVolumeHost) GetNodeName() types.NodeName {
	return types.NodeName(f.nodeName)
}

func (f *fakeVolumeHost) GetEventRecorder() record.EventRecorder {
	return nil
}

func ProbeVolumePlugins(config VolumeConfig) []VolumePlugin {
	if _, ok := config.OtherAttributes["fake-property"]; ok {
		return []VolumePlugin{
			&FakeVolumePlugin{
				PluginName: "fake-plugin",
				Host:       nil,
				// SomeFakeProperty: config.OtherAttributes["fake-property"] -- string, may require parsing by plugin
			},
		}
	}
	return []VolumePlugin{&FakeVolumePlugin{PluginName: "fake-plugin"}}
}

// FakeVolumePlugin is useful for testing.  It tries to be a fully compliant
// plugin, but all it does is make empty directories.
// Use as:
//   volume.RegisterPlugin(&FakePlugin{"fake-name"})
type FakeVolumePlugin struct {
	sync.RWMutex
	PluginName             string
	Host                   VolumeHost
	Config                 VolumeConfig
	LastProvisionerOptions VolumeOptions
	NewAttacherCallCount   int
	NewDetacherCallCount   int

	Mounters             []*FakeVolume
	Unmounters           []*FakeVolume
	Attachers            []*FakeVolume
	Detachers            []*FakeVolume
	BlockVolumeMappers   []*FakeVolume
	BlockVolumeUnmappers []*FakeVolume
}

var _ VolumePlugin = &FakeVolumePlugin{}
var _ BlockVolumePlugin = &FakeVolumePlugin{}
var _ RecyclableVolumePlugin = &FakeVolumePlugin{}
var _ DeletableVolumePlugin = &FakeVolumePlugin{}
var _ ProvisionableVolumePlugin = &FakeVolumePlugin{}
var _ AttachableVolumePlugin = &FakeVolumePlugin{}

func (plugin *FakeVolumePlugin) getFakeVolume(list *[]*FakeVolume) *FakeVolume {
	volume := &FakeVolume{}
	*list = append(*list, volume)
	return volume
}

func (plugin *FakeVolumePlugin) Init(host VolumeHost) error {
	plugin.Lock()
	defer plugin.Unlock()
	plugin.Host = host
	return nil
}

func (plugin *FakeVolumePlugin) GetPluginName() string {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.PluginName
}

func (plugin *FakeVolumePlugin) GetVolumeName(spec *Spec) (string, error) {
	var volumeName string
	if spec.Volume != nil && spec.Volume.GCEPersistentDisk != nil {
		volumeName = spec.Volume.GCEPersistentDisk.PDName
	} else if spec.PersistentVolume != nil &&
		spec.PersistentVolume.Spec.GCEPersistentDisk != nil {
		volumeName = spec.PersistentVolume.Spec.GCEPersistentDisk.PDName
	}
	if volumeName == "" {
		volumeName = spec.Name()
	}
	return volumeName, nil
}

func (plugin *FakeVolumePlugin) CanSupport(spec *Spec) bool {
	// TODO: maybe pattern-match on spec.Name() to decide?
	return true
}

func (plugin *FakeVolumePlugin) RequiresRemount() bool {
	return false
}

func (plugin *FakeVolumePlugin) SupportsMountOption() bool {
	return true
}

func (plugin *FakeVolumePlugin) SupportsBulkVolumeVerification() bool {
	return false
}

func (plugin *FakeVolumePlugin) NewMounter(spec *Spec, pod *v1.Pod, opts VolumeOptions) (Mounter, error) {
	plugin.Lock()
	defer plugin.Unlock()
	volume := plugin.getFakeVolume(&plugin.Mounters)
	volume.PodUID = pod.UID
	volume.VolName = spec.Name()
	volume.Plugin = plugin
	volume.MetricsNil = MetricsNil{}
	return volume, nil
}

func (plugin *FakeVolumePlugin) GetMounters() (Mounters []*FakeVolume) {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.Mounters
}

func (plugin *FakeVolumePlugin) NewUnmounter(volName string, podUID types.UID) (Unmounter, error) {
	plugin.Lock()
	defer plugin.Unlock()
	volume := plugin.getFakeVolume(&plugin.Unmounters)
	volume.PodUID = podUID
	volume.VolName = volName
	volume.Plugin = plugin
	volume.MetricsNil = MetricsNil{}
	return volume, nil
}

func (plugin *FakeVolumePlugin) GetUnmounters() (Unmounters []*FakeVolume) {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.Unmounters
}

// Block volume support
func (plugin *FakeVolumePlugin) NewBlockVolumeMapper(spec *Spec, pod *v1.Pod, opts VolumeOptions) (BlockVolumeMapper, error) {
	plugin.Lock()
	defer plugin.Unlock()
	volume := plugin.getFakeVolume(&plugin.BlockVolumeMappers)
	if pod != nil {
		volume.PodUID = pod.UID
	}
	volume.VolName = spec.Name()
	volume.Plugin = plugin
	return volume, nil
}

// Block volume support
func (plugin *FakeVolumePlugin) GetBlockVolumeMapper() (BlockVolumeMappers []*FakeVolume) {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.BlockVolumeMappers
}

// Block volume support
func (plugin *FakeVolumePlugin) NewBlockVolumeUnmapper(volName string, podUID types.UID) (BlockVolumeUnmapper, error) {
	plugin.Lock()
	defer plugin.Unlock()
	volume := plugin.getFakeVolume(&plugin.BlockVolumeUnmappers)
	volume.PodUID = podUID
	volume.VolName = volName
	volume.Plugin = plugin
	return volume, nil
}

// Block volume support
func (plugin *FakeVolumePlugin) GetBlockVolumeUnmapper() (BlockVolumeUnmappers []*FakeVolume) {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.BlockVolumeUnmappers
}

func (plugin *FakeVolumePlugin) NewAttacher() (Attacher, error) {
	plugin.Lock()
	defer plugin.Unlock()
	plugin.NewAttacherCallCount = plugin.NewAttacherCallCount + 1
	return plugin.getFakeVolume(&plugin.Attachers), nil
}

func (plugin *FakeVolumePlugin) GetAttachers() (Attachers []*FakeVolume) {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.Attachers
}

func (plugin *FakeVolumePlugin) GetNewAttacherCallCount() int {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.NewAttacherCallCount
}

func (plugin *FakeVolumePlugin) NewDetacher() (Detacher, error) {
	plugin.Lock()
	defer plugin.Unlock()
	plugin.NewDetacherCallCount = plugin.NewDetacherCallCount + 1
	return plugin.getFakeVolume(&plugin.Detachers), nil
}

func (plugin *FakeVolumePlugin) GetDetachers() (Detachers []*FakeVolume) {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.Detachers
}

func (plugin *FakeVolumePlugin) GetNewDetacherCallCount() int {
	plugin.RLock()
	defer plugin.RUnlock()
	return plugin.NewDetacherCallCount
}

func (plugin *FakeVolumePlugin) Recycle(pvName string, spec *Spec, eventRecorder recyclerclient.RecycleEventRecorder) error {
	return nil
}

func (plugin *FakeVolumePlugin) NewDeleter(spec *Spec) (Deleter, error) {
	return &FakeDeleter{"/attributesTransferredFromSpec", MetricsNil{}}, nil
}

func (plugin *FakeVolumePlugin) NewProvisioner(options VolumeOptions) (Provisioner, error) {
	plugin.Lock()
	defer plugin.Unlock()
	plugin.LastProvisionerOptions = options
	return &FakeProvisioner{options, plugin.Host}, nil
}

func (plugin *FakeVolumePlugin) GetAccessModes() []v1.PersistentVolumeAccessMode {
	return []v1.PersistentVolumeAccessMode{}
}

func (plugin *FakeVolumePlugin) ConstructVolumeSpec(volumeName, mountPath string) (*Spec, error) {
	return &Spec{
		Volume: &v1.Volume{
			Name: volumeName,
		},
	}, nil
}

// Block volume support
func (plugin *FakeVolumePlugin) ConstructBlockVolumeSpec(podUID types.UID, volumeName, mountPath string) (*Spec, error) {
	return &Spec{
		Volume: &v1.Volume{
			Name: volumeName,
		},
	}, nil
}

func (plugin *FakeVolumePlugin) GetDeviceMountRefs(deviceMountPath string) ([]string, error) {
	return []string{}, nil
}

type FakeFileVolumePlugin struct {
}

func (plugin *FakeFileVolumePlugin) Init(host VolumeHost) error {
	return nil
}

func (plugin *FakeFileVolumePlugin) GetPluginName() string {
	return "fake-file-plugin"
}

func (plugin *FakeFileVolumePlugin) GetVolumeName(spec *Spec) (string, error) {
	return "", nil
}

func (plugin *FakeFileVolumePlugin) CanSupport(spec *Spec) bool {
	return true
}

func (plugin *FakeFileVolumePlugin) RequiresRemount() bool {
	return false
}

func (plugin *FakeFileVolumePlugin) SupportsMountOption() bool {
	return false
}

func (plugin *FakeFileVolumePlugin) SupportsBulkVolumeVerification() bool {
	return false
}

func (plugin *FakeFileVolumePlugin) NewMounter(spec *Spec, podRef *v1.Pod, opts VolumeOptions) (Mounter, error) {
	return nil, nil
}

func (plugin *FakeFileVolumePlugin) NewUnmounter(name string, podUID types.UID) (Unmounter, error) {
	return nil, nil
}

func (plugin *FakeFileVolumePlugin) ConstructVolumeSpec(volumeName, mountPath string) (*Spec, error) {
	return nil, nil
}

func NewFakeFileVolumePlugin() []VolumePlugin {
	return []VolumePlugin{&FakeFileVolumePlugin{}}
}

type FakeVolume struct {
	sync.RWMutex
	PodUID  types.UID
	VolName string
	Plugin  *FakeVolumePlugin
	MetricsNil

	SetUpCallCount              int
	TearDownCallCount           int
	AttachCallCount             int
	DetachCallCount             int
	WaitForAttachCallCount      int
	MountDeviceCallCount        int
	UnmountDeviceCallCount      int
	GetDeviceMountPathCallCount int
	SetUpDeviceCallCount        int
	TearDownDeviceCallCount     int
	GlobalMapPathCallCount      int
	PodDeviceMapPathCallCount   int
}

func (_ *FakeVolume) GetAttributes() Attributes {
	return Attributes{
		ReadOnly:        false,
		Managed:         true,
		SupportsSELinux: true,
	}
}

func (fv *FakeVolume) CanMount() error {
	return nil
}

func (fv *FakeVolume) SetUp(fsGroup *int64) error {
	fv.Lock()
	defer fv.Unlock()
	fv.SetUpCallCount++
	return fv.SetUpAt(fv.getPath(), fsGroup)
}

func (fv *FakeVolume) GetSetUpCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.SetUpCallCount
}

func (fv *FakeVolume) SetUpAt(dir string, fsGroup *int64) error {
	return os.MkdirAll(dir, 0750)
}

func (fv *FakeVolume) GetPath() string {
	fv.RLock()
	defer fv.RUnlock()
	return fv.getPath()
}

func (fv *FakeVolume) getPath() string {
	return path.Join(fv.Plugin.Host.GetPodVolumeDir(fv.PodUID, utilstrings.EscapeQualifiedNameForDisk(fv.Plugin.PluginName), fv.VolName))
}

func (fv *FakeVolume) TearDown() error {
	fv.Lock()
	defer fv.Unlock()
	fv.TearDownCallCount++
	return fv.TearDownAt(fv.getPath())
}

func (fv *FakeVolume) GetTearDownCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.TearDownCallCount
}

func (fv *FakeVolume) TearDownAt(dir string) error {
	return os.RemoveAll(dir)
}

// Block volume support
func (fv *FakeVolume) SetUpDevice() (string, error) {
	fv.Lock()
	defer fv.Unlock()
	fv.SetUpDeviceCallCount++
	return "", nil
}

// Block volume support
func (fv *FakeVolume) GetSetUpDeviceCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.SetUpDeviceCallCount
}

// Block volume support
func (fv *FakeVolume) GetGlobalMapPath(spec *Spec) (string, error) {
	fv.RLock()
	defer fv.RUnlock()
	fv.GlobalMapPathCallCount++
	return fv.getGlobalMapPath()
}

// Block volume support
func (fv *FakeVolume) getGlobalMapPath() (string, error) {
	return path.Join(fv.Plugin.Host.GetVolumeDevicePluginDir(utilstrings.EscapeQualifiedNameForDisk(fv.Plugin.PluginName)), "pluginDependentPath"), nil
}

// Block volume support
func (fv *FakeVolume) GetGlobalMapPathCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.GlobalMapPathCallCount
}

// Block volume support
func (fv *FakeVolume) GetPodDeviceMapPath() (string, string) {
	fv.RLock()
	defer fv.RUnlock()
	fv.PodDeviceMapPathCallCount++
	return fv.getPodDeviceMapPath()
}

// Block volume support
func (fv *FakeVolume) getPodDeviceMapPath() (string, string) {
	return path.Join(fv.Plugin.Host.GetPodVolumeDeviceDir(fv.PodUID, utilstrings.EscapeQualifiedNameForDisk(fv.Plugin.PluginName))), fv.VolName
}

// Block volume support
func (fv *FakeVolume) GetPodDeviceMapPathCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.PodDeviceMapPathCallCount
}

// Block volume support
func (fv *FakeVolume) TearDownDevice(mapPath string, devicePath string) error {
	fv.Lock()
	defer fv.Unlock()
	fv.TearDownDeviceCallCount++
	return nil
}

// Block volume support
func (fv *FakeVolume) GetTearDownDeviceCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.TearDownDeviceCallCount
}

func (fv *FakeVolume) Attach(spec *Spec, nodeName types.NodeName) (string, error) {
	fv.Lock()
	defer fv.Unlock()
	fv.AttachCallCount++
	return "/dev/vdb-test", nil
}

func (fv *FakeVolume) GetAttachCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.AttachCallCount
}

func (fv *FakeVolume) WaitForAttach(spec *Spec, devicePath string, pod *v1.Pod, spectimeout time.Duration) (string, error) {
	fv.Lock()
	defer fv.Unlock()
	fv.WaitForAttachCallCount++
	return "/dev/sdb", nil
}

func (fv *FakeVolume) GetWaitForAttachCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.WaitForAttachCallCount
}

func (fv *FakeVolume) GetDeviceMountPath(spec *Spec) (string, error) {
	fv.Lock()
	defer fv.Unlock()
	fv.GetDeviceMountPathCallCount++
	return "", nil
}

func (fv *FakeVolume) MountDevice(spec *Spec, devicePath string, deviceMountPath string) error {
	fv.Lock()
	defer fv.Unlock()
	fv.MountDeviceCallCount++
	return nil
}

func (fv *FakeVolume) GetMountDeviceCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.MountDeviceCallCount
}

func (fv *FakeVolume) Detach(volumeName string, nodeName types.NodeName) error {
	fv.Lock()
	defer fv.Unlock()
	fv.DetachCallCount++
	return nil
}

func (fv *FakeVolume) VolumesAreAttached(spec []*Spec, nodeName types.NodeName) (map[*Spec]bool, error) {
	fv.Lock()
	defer fv.Unlock()
	return nil, nil
}

func (fv *FakeVolume) GetDetachCallCount() int {
	fv.RLock()
	defer fv.RUnlock()
	return fv.DetachCallCount
}

func (fv *FakeVolume) UnmountDevice(globalMountPath string) error {
	fv.Lock()
	defer fv.Unlock()
	fv.UnmountDeviceCallCount++
	return nil
}

type FakeDeleter struct {
	path string
	MetricsNil
}

func (fd *FakeDeleter) Delete() error {
	// nil is success, else error
	return nil
}

func (fd *FakeDeleter) GetPath() string {
	return fd.path
}

type FakeProvisioner struct {
	Options VolumeOptions
	Host    VolumeHost
}

func (fc *FakeProvisioner) Provision() (*v1.PersistentVolume, error) {
	fullpath := fmt.Sprintf("/tmp/hostpath_pv/%s", uuid.NewUUID())

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: fc.Options.PVName,
			Annotations: map[string]string{
				util.VolumeDynamicallyCreatedByKey: "fakeplugin-provisioner",
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: fc.Options.PersistentVolumeReclaimPolicy,
			AccessModes:                   fc.Options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): fc.Options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: fullpath,
				},
			},
		},
	}

	return pv, nil
}

var _ volumepathhandler.BlockVolumePathHandler = &FakeVolumePathHandler{}

//NewDeviceHandler Create a new IoHandler implementation
func NewBlockVolumePathHandler() volumepathhandler.BlockVolumePathHandler {
	return &FakeVolumePathHandler{}
}

type FakeVolumePathHandler struct {
	sync.RWMutex
}

func (fv *FakeVolumePathHandler) MapDevice(devicePath string, mapDir string, linkName string) error {
	// nil is success, else error
	return nil
}

func (fv *FakeVolumePathHandler) UnmapDevice(mapDir string, linkName string) error {
	// nil is success, else error
	return nil
}

func (fv *FakeVolumePathHandler) RemoveMapPath(mapPath string) error {
	// nil is success, else error
	return nil
}

func (fv *FakeVolumePathHandler) IsSymlinkExist(mapPath string) (bool, error) {
	// nil is success, else error
	return true, nil
}

func (fv *FakeVolumePathHandler) GetDeviceSymlinkRefs(devPath string, mapPath string) ([]string, error) {
	// nil is success, else error
	return []string{}, nil
}

func (fv *FakeVolumePathHandler) FindGlobalMapPathUUIDFromPod(pluginDir, mapPath string, podUID types.UID) (string, error) {
	// nil is success, else error
	return "", nil
}

func (fv *FakeVolumePathHandler) AttachFileDevice(path string) (string, error) {
	// nil is success, else error
	return "", nil
}

func (fv *FakeVolumePathHandler) GetLoopDevice(path string) (string, error) {
	// nil is success, else error
	return "/dev/loop1", nil
}

func (fv *FakeVolumePathHandler) RemoveLoopDevice(device string) error {
	// nil is success, else error
	return nil
}

// FindEmptyDirectoryUsageOnTmpfs finds the expected usage of an empty directory existing on
// a tmpfs filesystem on this system.
func FindEmptyDirectoryUsageOnTmpfs() (*resource.Quantity, error) {
	tmpDir, err := utiltesting.MkTmpdir("metrics_du_test")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	out, err := exec.Command("nice", "-n", "19", "du", "-s", "-B", "1", tmpDir).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed command 'du' on %s with error %v", tmpDir, err)
	}
	used, err := resource.ParseQuantity(strings.Fields(string(out))[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'du' output %s due to error %v", out, err)
	}
	used.Format = resource.BinarySI
	return &used, nil
}

// VerifyAttachCallCount ensures that at least one of the Attachers for this
// plugin has the expectedAttachCallCount number of calls. Otherwise it returns
// an error.
func VerifyAttachCallCount(
	expectedAttachCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, attacher := range fakeVolumePlugin.GetAttachers() {
		actualCallCount := attacher.GetAttachCallCount()
		if actualCallCount == expectedAttachCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No attachers have expected AttachCallCount. Expected: <%v>.",
		expectedAttachCallCount)
}

// VerifyZeroAttachCalls ensures that all of the Attachers for this plugin have
// a zero AttachCallCount. Otherwise it returns an error.
func VerifyZeroAttachCalls(fakeVolumePlugin *FakeVolumePlugin) error {
	for _, attacher := range fakeVolumePlugin.GetAttachers() {
		actualCallCount := attacher.GetAttachCallCount()
		if actualCallCount != 0 {
			return fmt.Errorf(
				"At least one attacher has non-zero AttachCallCount: <%v>.",
				actualCallCount)
		}
	}

	return nil
}

// VerifyWaitForAttachCallCount ensures that at least one of the Mounters for
// this plugin has the expectedWaitForAttachCallCount number of calls. Otherwise
// it returns an error.
func VerifyWaitForAttachCallCount(
	expectedWaitForAttachCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, attacher := range fakeVolumePlugin.GetAttachers() {
		actualCallCount := attacher.GetWaitForAttachCallCount()
		if actualCallCount == expectedWaitForAttachCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No Attachers have expected WaitForAttachCallCount. Expected: <%v>.",
		expectedWaitForAttachCallCount)
}

// VerifyZeroWaitForAttachCallCount ensures that all Attachers for this plugin
// have a zero WaitForAttachCallCount. Otherwise it returns an error.
func VerifyZeroWaitForAttachCallCount(fakeVolumePlugin *FakeVolumePlugin) error {
	for _, attacher := range fakeVolumePlugin.GetAttachers() {
		actualCallCount := attacher.GetWaitForAttachCallCount()
		if actualCallCount != 0 {
			return fmt.Errorf(
				"At least one attacher has non-zero WaitForAttachCallCount: <%v>.",
				actualCallCount)
		}
	}

	return nil
}

// VerifyMountDeviceCallCount ensures that at least one of the Mounters for
// this plugin has the expectedMountDeviceCallCount number of calls. Otherwise
// it returns an error.
func VerifyMountDeviceCallCount(
	expectedMountDeviceCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, attacher := range fakeVolumePlugin.GetAttachers() {
		actualCallCount := attacher.GetMountDeviceCallCount()
		if actualCallCount == expectedMountDeviceCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No Attachers have expected MountDeviceCallCount. Expected: <%v>.",
		expectedMountDeviceCallCount)
}

// VerifyZeroMountDeviceCallCount ensures that all Attachers for this plugin
// have a zero MountDeviceCallCount. Otherwise it returns an error.
func VerifyZeroMountDeviceCallCount(fakeVolumePlugin *FakeVolumePlugin) error {
	for _, attacher := range fakeVolumePlugin.GetAttachers() {
		actualCallCount := attacher.GetMountDeviceCallCount()
		if actualCallCount != 0 {
			return fmt.Errorf(
				"At least one attacher has non-zero MountDeviceCallCount: <%v>.",
				actualCallCount)
		}
	}

	return nil
}

// VerifySetUpCallCount ensures that at least one of the Mounters for this
// plugin has the expectedSetUpCallCount number of calls. Otherwise it returns
// an error.
func VerifySetUpCallCount(
	expectedSetUpCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, mounter := range fakeVolumePlugin.GetMounters() {
		actualCallCount := mounter.GetSetUpCallCount()
		if actualCallCount >= expectedSetUpCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No Mounters have expected SetUpCallCount. Expected: <%v>.",
		expectedSetUpCallCount)
}

// VerifyZeroSetUpCallCount ensures that all Mounters for this plugin have a
// zero SetUpCallCount. Otherwise it returns an error.
func VerifyZeroSetUpCallCount(fakeVolumePlugin *FakeVolumePlugin) error {
	for _, mounter := range fakeVolumePlugin.GetMounters() {
		actualCallCount := mounter.GetSetUpCallCount()
		if actualCallCount != 0 {
			return fmt.Errorf(
				"At least one mounter has non-zero SetUpCallCount: <%v>.",
				actualCallCount)
		}
	}

	return nil
}

// VerifyTearDownCallCount ensures that at least one of the Unounters for this
// plugin has the expectedTearDownCallCount number of calls. Otherwise it
// returns an error.
func VerifyTearDownCallCount(
	expectedTearDownCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, unmounter := range fakeVolumePlugin.GetUnmounters() {
		actualCallCount := unmounter.GetTearDownCallCount()
		if actualCallCount >= expectedTearDownCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No Unmounters have expected SetUpCallCount. Expected: <%v>.",
		expectedTearDownCallCount)
}

// VerifyZeroTearDownCallCount ensures that all Mounters for this plugin have a
// zero TearDownCallCount. Otherwise it returns an error.
func VerifyZeroTearDownCallCount(fakeVolumePlugin *FakeVolumePlugin) error {
	for _, mounter := range fakeVolumePlugin.GetMounters() {
		actualCallCount := mounter.GetTearDownCallCount()
		if actualCallCount != 0 {
			return fmt.Errorf(
				"At least one mounter has non-zero TearDownCallCount: <%v>.",
				actualCallCount)
		}
	}

	return nil
}

// VerifyDetachCallCount ensures that at least one of the Attachers for this
// plugin has the expectedDetachCallCount number of calls. Otherwise it returns
// an error.
func VerifyDetachCallCount(
	expectedDetachCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, detacher := range fakeVolumePlugin.GetDetachers() {
		actualCallCount := detacher.GetDetachCallCount()
		if actualCallCount == expectedDetachCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No Detachers have expected DetachCallCount. Expected: <%v>.",
		expectedDetachCallCount)
}

// VerifyZeroDetachCallCount ensures that all Detachers for this plugin have a
// zero DetachCallCount. Otherwise it returns an error.
func VerifyZeroDetachCallCount(fakeVolumePlugin *FakeVolumePlugin) error {
	for _, detacher := range fakeVolumePlugin.GetDetachers() {
		actualCallCount := detacher.GetDetachCallCount()
		if actualCallCount != 0 {
			return fmt.Errorf(
				"At least one detacher has non-zero DetachCallCount: <%v>.",
				actualCallCount)
		}
	}

	return nil
}

// VerifySetUpDeviceCallCount ensures that at least one of the Mappers for this
// plugin has the expectedSetUpDeviceCallCount number of calls. Otherwise it
// returns an error.
func VerifySetUpDeviceCallCount(
	expectedSetUpDeviceCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, mapper := range fakeVolumePlugin.GetBlockVolumeMapper() {
		actualCallCount := mapper.GetSetUpDeviceCallCount()
		if actualCallCount >= expectedSetUpDeviceCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No Mapper have expected SetUpDeviceCallCount. Expected: <%v>.",
		expectedSetUpDeviceCallCount)
}

// VerifyTearDownDeviceCallCount ensures that at least one of the Unmappers for this
// plugin has the expectedTearDownDeviceCallCount number of calls. Otherwise it
// returns an error.
func VerifyTearDownDeviceCallCount(
	expectedTearDownDeviceCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, unmapper := range fakeVolumePlugin.GetBlockVolumeUnmapper() {
		actualCallCount := unmapper.GetTearDownDeviceCallCount()
		if actualCallCount >= expectedTearDownDeviceCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No Unmapper have expected TearDownDeviceCallCount. Expected: <%v>.",
		expectedTearDownDeviceCallCount)
}

// VerifyZeroTearDownDeviceCallCount ensures that all Mappers for this plugin have a
// zero TearDownDeviceCallCount. Otherwise it returns an error.
func VerifyZeroTearDownDeviceCallCount(fakeVolumePlugin *FakeVolumePlugin) error {
	for _, unmapper := range fakeVolumePlugin.GetBlockVolumeUnmapper() {
		actualCallCount := unmapper.GetTearDownDeviceCallCount()
		if actualCallCount != 0 {
			return fmt.Errorf(
				"At least one unmapper has non-zero TearDownDeviceCallCount: <%v>.",
				actualCallCount)
		}
	}

	return nil
}

// VerifyGetGlobalMapPathCallCount ensures that at least one of the Mappers for this
// plugin has the expectedGlobalMapPathCallCount number of calls. Otherwise it returns
// an error.
func VerifyGetGlobalMapPathCallCount(
	expectedGlobalMapPathCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, mapper := range fakeVolumePlugin.GetBlockVolumeMapper() {
		actualCallCount := mapper.GetGlobalMapPathCallCount()
		if actualCallCount == expectedGlobalMapPathCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No Mappers have expected GetGlobalMapPathCallCount. Expected: <%v>.",
		expectedGlobalMapPathCallCount)
}

// VerifyGetPodDeviceMapPathCallCount ensures that at least one of the Mappers for this
// plugin has the expectedPodDeviceMapPathCallCount number of calls. Otherwise it returns
// an error.
func VerifyGetPodDeviceMapPathCallCount(
	expectedPodDeviceMapPathCallCount int,
	fakeVolumePlugin *FakeVolumePlugin) error {
	for _, mapper := range fakeVolumePlugin.GetBlockVolumeMapper() {
		actualCallCount := mapper.GetPodDeviceMapPathCallCount()
		if actualCallCount == expectedPodDeviceMapPathCallCount {
			return nil
		}
	}

	return fmt.Errorf(
		"No Mappers have expected GetPodDeviceMapPathCallCount. Expected: <%v>.",
		expectedPodDeviceMapPathCallCount)
}

// GetTestVolumePluginMgr creates, initializes, and returns a test volume plugin
// manager and fake volume plugin using a fake volume host.
func GetTestVolumePluginMgr(
	t *testing.T) (*VolumePluginMgr, *FakeVolumePlugin) {
	v := NewFakeVolumeHost(
		"",  /* rootDir */
		nil, /* kubeClient */
		nil, /* plugins */
	)
	plugins := ProbeVolumePlugins(VolumeConfig{})
	if err := v.pluginMgr.InitPlugins(plugins, nil /* prober */, v); err != nil {
		t.Fatal(err)
	}

	return &v.pluginMgr, plugins[0].(*FakeVolumePlugin)
}

// CreateTestPVC returns a provisionable PVC for tests
func CreateTestPVC(capacity string, accessModes []v1.PersistentVolumeAccessMode) *v1.PersistentVolumeClaim {
	claim := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy",
			Namespace: "default",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse(capacity),
				},
			},
		},
	}
	return &claim
}

func MetricsEqualIgnoreTimestamp(a *Metrics, b *Metrics) bool {
	available := a.Available == b.Available
	capacity := a.Capacity == b.Capacity
	used := a.Used == b.Used
	inodes := a.Inodes == b.Inodes
	inodesFree := a.InodesFree == b.InodesFree
	inodesUsed := a.InodesUsed == b.InodesUsed
	return available && capacity && used && inodes && inodesFree && inodesUsed
}

func ContainsAccessMode(modes []v1.PersistentVolumeAccessMode, mode v1.PersistentVolumeAccessMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}
