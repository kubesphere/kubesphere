// +build linux darwin windows

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

package local

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utiltesting "k8s.io/client-go/util/testing"
	"k8s.io/kubernetes/pkg/volume"
	volumetest "k8s.io/kubernetes/pkg/volume/testing"
)

const (
	testPVName     = "pvA"
	testMountPath  = "pods/poduid/volumes/kubernetes.io~local-volume/pvA"
	testGlobalPath = "plugins/kubernetes.io~local-volume/volumeDevices/pvA"
	testPodPath    = "pods/poduid/volumeDevices/kubernetes.io~local-volume"
	testNodeName   = "fakeNodeName"
)

func getPlugin(t *testing.T) (string, volume.VolumePlugin) {
	tmpDir, err := utiltesting.MkTmpdir("localVolumeTest")
	if err != nil {
		t.Fatalf("can't make a temp dir: %v", err)
	}

	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(), nil /* prober */, volumetest.NewFakeVolumeHost(tmpDir, nil, nil))

	plug, err := plugMgr.FindPluginByName(localVolumePluginName)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Can't find the plugin by name")
	}
	if plug.GetPluginName() != localVolumePluginName {
		t.Errorf("Wrong name: %s", plug.GetPluginName())
	}
	return tmpDir, plug
}

func getBlockPlugin(t *testing.T) (string, volume.BlockVolumePlugin) {
	tmpDir, err := utiltesting.MkTmpdir("localVolumeTest")
	if err != nil {
		t.Fatalf("can't make a temp dir: %v", err)
	}

	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(), nil /* prober */, volumetest.NewFakeVolumeHost(tmpDir, nil, nil))
	plug, err := plugMgr.FindMapperPluginByName(localVolumePluginName)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Can't find the plugin by name: %q", localVolumePluginName)
	}
	if plug.GetPluginName() != localVolumePluginName {
		t.Errorf("Wrong name: %s", plug.GetPluginName())
	}
	return tmpDir, plug
}

func getPersistentPlugin(t *testing.T) (string, volume.PersistentVolumePlugin) {
	tmpDir, err := utiltesting.MkTmpdir("localVolumeTest")
	if err != nil {
		t.Fatalf("can't make a temp dir: %v", err)
	}

	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(), nil /* prober */, volumetest.NewFakeVolumeHost(tmpDir, nil, nil))

	plug, err := plugMgr.FindPersistentPluginByName(localVolumePluginName)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Can't find the plugin by name")
	}
	if plug.GetPluginName() != localVolumePluginName {
		t.Errorf("Wrong name: %s", plug.GetPluginName())
	}
	return tmpDir, plug
}

func getTestVolume(readOnly bool, path string, isBlock bool) *volume.Spec {
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: testPVName,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeSource: v1.PersistentVolumeSource{
				Local: &v1.LocalVolumeSource{
					Path: path,
				},
			},
		},
	}

	if isBlock {
		blockMode := v1.PersistentVolumeBlock
		pv.Spec.VolumeMode = &blockMode
	}
	return volume.NewSpecFromPersistentVolume(pv, readOnly)
}

func TestCanSupport(t *testing.T) {
	tmpDir, plug := getPlugin(t)
	defer os.RemoveAll(tmpDir)

	if !plug.CanSupport(getTestVolume(false, tmpDir, false)) {
		t.Errorf("Expected true")
	}
}

func TestGetAccessModes(t *testing.T) {
	tmpDir, plug := getPersistentPlugin(t)
	defer os.RemoveAll(tmpDir)

	modes := plug.GetAccessModes()
	if !volumetest.ContainsAccessMode(modes, v1.ReadWriteOnce) {
		t.Errorf("Expected AccessModeType %q", v1.ReadWriteOnce)
	}

	if volumetest.ContainsAccessMode(modes, v1.ReadWriteMany) {
		t.Errorf("Found AccessModeType %q, expected not", v1.ReadWriteMany)
	}
	if volumetest.ContainsAccessMode(modes, v1.ReadOnlyMany) {
		t.Errorf("Found AccessModeType %q, expected not", v1.ReadOnlyMany)
	}
}

func TestGetVolumeName(t *testing.T) {
	tmpDir, plug := getPersistentPlugin(t)
	defer os.RemoveAll(tmpDir)

	volName, err := plug.GetVolumeName(getTestVolume(false, tmpDir, false))
	if err != nil {
		t.Errorf("Failed to get volume name: %v", err)
	}
	if volName != testPVName {
		t.Errorf("Expected volume name %q, got %q", testPVName, volName)
	}
}

func TestInvalidLocalPath(t *testing.T) {
	tmpDir, plug := getPlugin(t)
	defer os.RemoveAll(tmpDir)

	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: types.UID("poduid")}}
	mounter, err := plug.NewMounter(getTestVolume(false, "/no/backsteps/allowed/..", false), pod, volume.VolumeOptions{})
	if err != nil {
		t.Fatal(err)
	}

	err = mounter.SetUp(nil)
	expectedMsg := "invalid path: /no/backsteps/allowed/.. must not contain '..'"
	if err.Error() != expectedMsg {
		t.Fatalf("expected error `%s` but got `%s`", expectedMsg, err)
	}
}

func TestMountUnmount(t *testing.T) {
	tmpDir, plug := getPlugin(t)
	defer os.RemoveAll(tmpDir)

	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: types.UID("poduid")}}
	mounter, err := plug.NewMounter(getTestVolume(false, tmpDir, false), pod, volume.VolumeOptions{})
	if err != nil {
		t.Errorf("Failed to make a new Mounter: %v", err)
	}
	if mounter == nil {
		t.Fatalf("Got a nil Mounter")
	}

	volPath := path.Join(tmpDir, testMountPath)
	path := mounter.GetPath()
	if path != volPath {
		t.Errorf("Got unexpected path: %s", path)
	}

	if err := mounter.SetUp(nil); err != nil {
		t.Errorf("Expected success, got: %v", err)
	}

	if runtime.GOOS != "windows" {
		// skip this check in windows since the "bind mount" logic is implemented differently in mount_wiondows.go
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				t.Errorf("SetUp() failed, volume path not created: %s", path)
			} else {
				t.Errorf("SetUp() failed: %v", err)
			}
		}
	}

	unmounter, err := plug.NewUnmounter(testPVName, pod.UID)
	if err != nil {
		t.Errorf("Failed to make a new Unmounter: %v", err)
	}
	if unmounter == nil {
		t.Fatalf("Got a nil Unmounter")
	}

	if err := unmounter.TearDown(); err != nil {
		t.Errorf("Expected success, got: %v", err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Errorf("TearDown() failed, volume path still exists: %s", path)
	} else if !os.IsNotExist(err) {
		t.Errorf("TearDown() failed: %v", err)
	}
}

// TestMapUnmap tests block map and unmap interfaces.
func TestMapUnmap(t *testing.T) {
	tmpDir, plug := getBlockPlugin(t)
	defer os.RemoveAll(tmpDir)

	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: types.UID("poduid")}}
	volSpec := getTestVolume(false, tmpDir, true /*isBlock*/)
	mapper, err := plug.NewBlockVolumeMapper(volSpec, pod, volume.VolumeOptions{})
	if err != nil {
		t.Errorf("Failed to make a new Mounter: %v", err)
	}
	if mapper == nil {
		t.Fatalf("Got a nil Mounter")
	}

	expectedGlobalPath := path.Join(tmpDir, testGlobalPath)
	globalPath, err := mapper.GetGlobalMapPath(volSpec)
	if err != nil {
		t.Errorf("Failed to get global path: %v", err)
	}
	if globalPath != expectedGlobalPath {
		t.Errorf("Got unexpected path: %s, expected %s", globalPath, expectedGlobalPath)
	}
	expectedPodPath := path.Join(tmpDir, testPodPath)
	podPath, volName := mapper.GetPodDeviceMapPath()
	if podPath != expectedPodPath {
		t.Errorf("Got unexpected pod path: %s, expected %s", podPath, expectedPodPath)
	}
	if volName != testPVName {
		t.Errorf("Got unexpected volNamne: %s, expected %s", volName, testPVName)
	}
	devPath, err := mapper.SetUpDevice()
	if err != nil {
		t.Errorf("Failed to SetUpDevice, err: %v", err)
	}

	if _, err := os.Stat(devPath); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("SetUpDevice() failed, volume path not created: %s", devPath)
		} else {
			t.Errorf("SetUpDevice() failed: %v", err)
		}
	}

	unmapper, err := plug.NewBlockVolumeUnmapper(testPVName, pod.UID)
	if err != nil {
		t.Fatalf("Failed to make a new Unmapper: %v", err)
	}
	if unmapper == nil {
		t.Fatalf("Got a nil Unmapper")
	}

	if err := unmapper.TearDownDevice(globalPath, devPath); err != nil {
		t.Errorf("TearDownDevice failed, err: %v", err)
	}
}

func testFSGroupMount(plug volume.VolumePlugin, pod *v1.Pod, tmpDir string, fsGroup int64) error {
	mounter, err := plug.NewMounter(getTestVolume(false, tmpDir, false), pod, volume.VolumeOptions{})
	if err != nil {
		return err
	}
	if mounter == nil {
		return fmt.Errorf("Got a nil Mounter")
	}

	volPath := path.Join(tmpDir, testMountPath)
	path := mounter.GetPath()
	if path != volPath {
		return fmt.Errorf("Got unexpected path: %s", path)
	}

	if err := mounter.SetUp(&fsGroup); err != nil {
		return err
	}
	return nil
}

func TestConstructVolumeSpec(t *testing.T) {
	tmpDir, plug := getPlugin(t)
	defer os.RemoveAll(tmpDir)

	volPath := path.Join(tmpDir, testMountPath)
	spec, err := plug.ConstructVolumeSpec(testPVName, volPath)
	if err != nil {
		t.Errorf("ConstructVolumeSpec() failed: %v", err)
	}
	if spec == nil {
		t.Fatalf("ConstructVolumeSpec() returned nil")
	}

	volName := spec.Name()
	if volName != testPVName {
		t.Errorf("Expected volume name %q, got %q", testPVName, volName)
	}

	if spec.Volume != nil {
		t.Errorf("Volume object returned, expected nil")
	}

	pv := spec.PersistentVolume
	if pv == nil {
		t.Fatalf("PersistentVolume object nil")
	}

	ls := pv.Spec.PersistentVolumeSource.Local
	if ls == nil {
		t.Fatalf("LocalVolumeSource object nil")
	}
}

func TestConstructBlockVolumeSpec(t *testing.T) {
	tmpDir, plug := getBlockPlugin(t)
	defer os.RemoveAll(tmpDir)

	podPath := path.Join(tmpDir, testPodPath)
	spec, err := plug.ConstructBlockVolumeSpec(types.UID("poduid"), testPVName, podPath)
	if err != nil {
		t.Errorf("ConstructBlockVolumeSpec() failed: %v", err)
	}
	if spec == nil {
		t.Fatalf("ConstructBlockVolumeSpec() returned nil")
	}

	volName := spec.Name()
	if volName != testPVName {
		t.Errorf("Expected volume name %q, got %q", testPVName, volName)
	}

	if spec.Volume != nil {
		t.Errorf("Volume object returned, expected nil")
	}

	pv := spec.PersistentVolume
	if pv == nil {
		t.Fatalf("PersistentVolume object nil")
	}

	if spec.PersistentVolume.Spec.VolumeMode == nil {
		t.Fatalf("Volume mode has not been set.")
	}

	if *spec.PersistentVolume.Spec.VolumeMode != v1.PersistentVolumeBlock {
		t.Errorf("Unexpected volume mode %q", *spec.PersistentVolume.Spec.VolumeMode)
	}

	ls := pv.Spec.PersistentVolumeSource.Local
	if ls == nil {
		t.Fatalf("LocalVolumeSource object nil")
	}
}

func TestPersistentClaimReadOnlyFlag(t *testing.T) {
	tmpDir, plug := getPlugin(t)
	defer os.RemoveAll(tmpDir)

	// Read only == true
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: types.UID("poduid")}}
	mounter, err := plug.NewMounter(getTestVolume(true, tmpDir, false), pod, volume.VolumeOptions{})
	if err != nil {
		t.Errorf("Failed to make a new Mounter: %v", err)
	}
	if mounter == nil {
		t.Fatalf("Got a nil Mounter")
	}
	if !mounter.GetAttributes().ReadOnly {
		t.Errorf("Expected true for mounter.IsReadOnly")
	}

	// Read only == false
	mounter, err = plug.NewMounter(getTestVolume(false, tmpDir, false), pod, volume.VolumeOptions{})
	if err != nil {
		t.Errorf("Failed to make a new Mounter: %v", err)
	}
	if mounter == nil {
		t.Fatalf("Got a nil Mounter")
	}
	if mounter.GetAttributes().ReadOnly {
		t.Errorf("Expected false for mounter.IsReadOnly")
	}
}

func TestUnsupportedPlugins(t *testing.T) {
	tmpDir, err := utiltesting.MkTmpdir("localVolumeTest")
	if err != nil {
		t.Fatalf("can't make a temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(), nil /* prober */, volumetest.NewFakeVolumeHost(tmpDir, nil, nil))
	spec := getTestVolume(false, tmpDir, false)

	recyclePlug, err := plugMgr.FindRecyclablePluginBySpec(spec)
	if err == nil && recyclePlug != nil {
		t.Errorf("Recyclable plugin found, expected none")
	}

	deletePlug, err := plugMgr.FindDeletablePluginByName(localVolumePluginName)
	if err == nil && deletePlug != nil {
		t.Errorf("Deletable plugin found, expected none")
	}

	attachPlug, err := plugMgr.FindAttachablePluginByName(localVolumePluginName)
	if err == nil && attachPlug != nil {
		t.Errorf("Attachable plugin found, expected none")
	}

	createPlug, err := plugMgr.FindCreatablePluginBySpec(spec)
	if err == nil && createPlug != nil {
		t.Errorf("Creatable plugin found, expected none")
	}

	provisionPlug, err := plugMgr.FindProvisionablePluginByName(localVolumePluginName)
	if err == nil && provisionPlug != nil {
		t.Errorf("Provisionable plugin found, expected none")
	}
}

func TestFilterPodMounts(t *testing.T) {
	tmpDir, plug := getPlugin(t)
	defer os.RemoveAll(tmpDir)

	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: types.UID("poduid")}}
	mounter, err := plug.NewMounter(getTestVolume(false, tmpDir, false), pod, volume.VolumeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	lvMounter, ok := mounter.(*localVolumeMounter)
	if !ok {
		t.Fatal("mounter is not localVolumeMounter")
	}

	host := volumetest.NewFakeVolumeHost(tmpDir, nil, nil)
	podsDir := host.GetPodsDir()

	cases := map[string]struct {
		input    []string
		expected []string
	}{
		"empty": {
			[]string{},
			[]string{},
		},
		"not-pod-mount": {
			[]string{"/mnt/outside"},
			[]string{},
		},
		"pod-mount": {
			[]string{filepath.Join(podsDir, "pod-mount")},
			[]string{filepath.Join(podsDir, "pod-mount")},
		},
		"not-directory-prefix": {
			[]string{podsDir + "pod-mount"},
			[]string{},
		},
		"mix": {
			[]string{"/mnt/outside",
				filepath.Join(podsDir, "pod-mount"),
				"/another/outside",
				filepath.Join(podsDir, "pod-mount2")},
			[]string{filepath.Join(podsDir, "pod-mount"),
				filepath.Join(podsDir, "pod-mount2")},
		},
	}
	for name, test := range cases {
		output := lvMounter.filterPodMounts(test.input)
		if !reflect.DeepEqual(output, test.expected) {
			t.Errorf("%v failed: output %+v doesn't equal expected %+v", name, output, test.expected)
		}
	}
}
