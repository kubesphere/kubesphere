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

package kubelet

import (
	"fmt"
	"net"
	"runtime"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/kubelet/configmap"
	"k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/kubelet/mountpod"
	"k8s.io/kubernetes/pkg/kubelet/secret"
	"k8s.io/kubernetes/pkg/util/io"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"
)

// NewInitializedVolumePluginMgr returns a new instance of
// volume.VolumePluginMgr initialized with kubelets implementation of the
// volume.VolumeHost interface.
//
// kubelet - used by VolumeHost methods to expose kubelet specific parameters
// plugins - used to initialize volumePluginMgr
func NewInitializedVolumePluginMgr(
	kubelet *Kubelet,
	secretManager secret.Manager,
	configMapManager configmap.Manager,
	plugins []volume.VolumePlugin,
	prober volume.DynamicPluginProber) (*volume.VolumePluginMgr, error) {

	mountPodManager, err := mountpod.NewManager(kubelet.getRootDir(), kubelet.podManager)
	if err != nil {
		return nil, err
	}
	kvh := &kubeletVolumeHost{
		kubelet:          kubelet,
		volumePluginMgr:  volume.VolumePluginMgr{},
		secretManager:    secretManager,
		configMapManager: configMapManager,
		mountPodManager:  mountPodManager,
	}

	if err := kvh.volumePluginMgr.InitPlugins(plugins, prober, kvh); err != nil {
		return nil, fmt.Errorf(
			"Could not initialize volume plugins for KubeletVolumePluginMgr: %v",
			err)
	}

	return &kvh.volumePluginMgr, nil
}

// Compile-time check to ensure kubeletVolumeHost implements the VolumeHost interface
var _ volume.VolumeHost = &kubeletVolumeHost{}

func (kvh *kubeletVolumeHost) GetPluginDir(pluginName string) string {
	return kvh.kubelet.getPluginDir(pluginName)
}

type kubeletVolumeHost struct {
	kubelet          *Kubelet
	volumePluginMgr  volume.VolumePluginMgr
	secretManager    secret.Manager
	configMapManager configmap.Manager
	mountPodManager  mountpod.Manager
}

func (kvh *kubeletVolumeHost) GetVolumeDevicePluginDir(pluginName string) string {
	return kvh.kubelet.getVolumeDevicePluginDir(pluginName)
}

func (kvh *kubeletVolumeHost) GetPodsDir() string {
	return kvh.kubelet.getPodsDir()
}

func (kvh *kubeletVolumeHost) GetPodVolumeDir(podUID types.UID, pluginName string, volumeName string) string {
	dir := kvh.kubelet.getPodVolumeDir(podUID, pluginName, volumeName)
	if runtime.GOOS == "windows" {
		dir = util.GetWindowsPath(dir)
	}
	return dir
}

func (kvh *kubeletVolumeHost) GetPodVolumeDeviceDir(podUID types.UID, pluginName string) string {
	return kvh.kubelet.getPodVolumeDeviceDir(podUID, pluginName)
}

func (kvh *kubeletVolumeHost) GetPodPluginDir(podUID types.UID, pluginName string) string {
	return kvh.kubelet.getPodPluginDir(podUID, pluginName)
}

func (kvh *kubeletVolumeHost) GetKubeClient() clientset.Interface {
	return kvh.kubelet.kubeClient
}

func (kvh *kubeletVolumeHost) NewWrapperMounter(
	volName string,
	spec volume.Spec,
	pod *v1.Pod,
	opts volume.VolumeOptions) (volume.Mounter, error) {
	// The name of wrapper volume is set to "wrapped_{wrapped_volume_name}"
	wrapperVolumeName := "wrapped_" + volName
	if spec.Volume != nil {
		spec.Volume.Name = wrapperVolumeName
	}

	return kvh.kubelet.newVolumeMounterFromPlugins(&spec, pod, opts)
}

func (kvh *kubeletVolumeHost) NewWrapperUnmounter(volName string, spec volume.Spec, podUID types.UID) (volume.Unmounter, error) {
	// The name of wrapper volume is set to "wrapped_{wrapped_volume_name}"
	wrapperVolumeName := "wrapped_" + volName
	if spec.Volume != nil {
		spec.Volume.Name = wrapperVolumeName
	}

	plugin, err := kvh.kubelet.volumePluginMgr.FindPluginBySpec(&spec)
	if err != nil {
		return nil, err
	}

	return plugin.NewUnmounter(spec.Name(), podUID)
}

func (kvh *kubeletVolumeHost) GetCloudProvider() cloudprovider.Interface {
	return kvh.kubelet.cloud
}

func (kvh *kubeletVolumeHost) GetMounter(pluginName string) mount.Interface {
	exec, err := kvh.getMountExec(pluginName)
	if err != nil {
		glog.V(2).Infof("Error finding mount pod for plugin %s: %s", pluginName, err.Error())
		// Use the default mounter
		exec = nil
	}
	if exec == nil {
		return kvh.kubelet.mounter
	}
	return mount.NewExecMounter(exec, kvh.kubelet.mounter)
}

func (kvh *kubeletVolumeHost) GetWriter() io.Writer {
	return kvh.kubelet.writer
}

func (kvh *kubeletVolumeHost) GetHostName() string {
	return kvh.kubelet.hostname
}

func (kvh *kubeletVolumeHost) GetHostIP() (net.IP, error) {
	return kvh.kubelet.GetHostIP()
}

func (kvh *kubeletVolumeHost) GetNodeAllocatable() (v1.ResourceList, error) {
	node, err := kvh.kubelet.getNodeAnyWay()
	if err != nil {
		return nil, fmt.Errorf("error retrieving node: %v", err)
	}
	return node.Status.Allocatable, nil
}

func (kvh *kubeletVolumeHost) GetSecretFunc() func(namespace, name string) (*v1.Secret, error) {
	return kvh.secretManager.GetSecret
}

func (kvh *kubeletVolumeHost) GetConfigMapFunc() func(namespace, name string) (*v1.ConfigMap, error) {
	return kvh.configMapManager.GetConfigMap
}

func (kvh *kubeletVolumeHost) GetNodeLabels() (map[string]string, error) {
	node, err := kvh.kubelet.GetNode()
	if err != nil {
		return nil, fmt.Errorf("error retrieving node: %v", err)
	}
	return node.Labels, nil
}

func (kvh *kubeletVolumeHost) GetNodeName() types.NodeName {
	return kvh.kubelet.nodeName
}

func (kvh *kubeletVolumeHost) GetEventRecorder() record.EventRecorder {
	return kvh.kubelet.recorder
}

func (kvh *kubeletVolumeHost) GetExec(pluginName string) mount.Exec {
	exec, err := kvh.getMountExec(pluginName)
	if err != nil {
		glog.V(2).Infof("Error finding mount pod for plugin %s: %s", pluginName, err.Error())
		// Use the default exec
		exec = nil
	}
	if exec == nil {
		return mount.NewOsExec()
	}
	return exec
}

// getMountExec returns mount.Exec implementation that leads to pod with mount
// utilities. It returns nil,nil when there is no such pod and default mounter /
// os.Exec should be used.
func (kvh *kubeletVolumeHost) getMountExec(pluginName string) (mount.Exec, error) {
	if !utilfeature.DefaultFeatureGate.Enabled(features.MountContainers) {
		glog.V(5).Infof("using default mounter/exec for %s", pluginName)
		return nil, nil
	}

	pod, container, err := kvh.mountPodManager.GetMountPod(pluginName)
	if err != nil {
		return nil, err
	}
	if pod == nil {
		// Use default mounter/exec for this plugin
		glog.V(5).Infof("using default mounter/exec for %s", pluginName)
		return nil, nil
	}
	glog.V(5).Infof("using container %s/%s/%s to execute mount utilites for %s", pod.Namespace, pod.Name, container, pluginName)
	return &containerExec{
		pod:           pod,
		containerName: container,
		kl:            kvh.kubelet,
	}, nil
}

// containerExec is implementation of mount.Exec that executes commands in given
// container in given pod.
type containerExec struct {
	pod           *v1.Pod
	containerName string
	kl            *Kubelet
}

var _ mount.Exec = &containerExec{}

func (e *containerExec) Run(cmd string, args ...string) ([]byte, error) {
	cmdline := append([]string{cmd}, args...)
	glog.V(5).Infof("Exec mounter running in pod %s/%s/%s: %v", e.pod.Namespace, e.pod.Name, e.containerName, cmdline)
	return e.kl.RunInContainer(container.GetPodFullName(e.pod), e.pod.UID, e.containerName, cmdline)
}
