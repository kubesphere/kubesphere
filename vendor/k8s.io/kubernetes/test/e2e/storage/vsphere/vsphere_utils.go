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

package vsphere

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	vim25types "github.com/vmware/govmomi/vim25/types"
	vimtypes "github.com/vmware/govmomi/vim25/types"

	"k8s.io/api/core/v1"
	storage "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/volume/util"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/storage/utils"
)

const (
	volumesPerNode = 55
	storageclass1  = "sc-default"
	storageclass2  = "sc-vsan"
	storageclass3  = "sc-spbm"
	storageclass4  = "sc-user-specified-ds"
	DummyDiskName  = "kube-dummyDisk.vmdk"
	ProviderPrefix = "vsphere://"
)

// volumeState represents the state of a volume.
type volumeState int32

const (
	volumeStateDetached volumeState = 1
	volumeStateAttached volumeState = 2
)

// Wait until vsphere volumes are detached from the list of nodes or time out after 5 minutes
func waitForVSphereDisksToDetach(nodeVolumes map[string][]string) error {
	var (
		err            error
		disksAttached  = true
		detachTimeout  = 5 * time.Minute
		detachPollTime = 10 * time.Second
	)
	err = wait.Poll(detachPollTime, detachTimeout, func() (bool, error) {
		attachedResult, err := disksAreAttached(nodeVolumes)
		if err != nil {
			return false, err
		}
		for nodeName, nodeVolumes := range attachedResult {
			for volumePath, attached := range nodeVolumes {
				if attached {
					framework.Logf("Waiting for volumes %q to detach from %q.", volumePath, string(nodeName))
					return false, nil
				}
			}
		}
		disksAttached = false
		framework.Logf("Volume are successfully detached from all the nodes: %+v", nodeVolumes)
		return true, nil
	})
	if err != nil {
		return err
	}
	if disksAttached {
		return fmt.Errorf("Gave up waiting for volumes to detach after %v", detachTimeout)
	}
	return nil
}

// Wait until vsphere vmdk moves to expected state on the given node, or time out after 6 minutes
func waitForVSphereDiskStatus(volumePath string, nodeName string, expectedState volumeState) error {
	var (
		err          error
		diskAttached bool
		currentState volumeState
		timeout      = 6 * time.Minute
		pollTime     = 10 * time.Second
	)

	var attachedState = map[bool]volumeState{
		true:  volumeStateAttached,
		false: volumeStateDetached,
	}

	var attachedStateMsg = map[volumeState]string{
		volumeStateAttached: "attached to",
		volumeStateDetached: "detached from",
	}

	err = wait.Poll(pollTime, timeout, func() (bool, error) {
		diskAttached, err = diskIsAttached(volumePath, nodeName)
		if err != nil {
			return true, err
		}

		currentState = attachedState[diskAttached]
		if currentState == expectedState {
			framework.Logf("Volume %q has successfully %s %q", volumePath, attachedStateMsg[currentState], nodeName)
			return true, nil
		}
		framework.Logf("Waiting for Volume %q to be %s %q.", volumePath, attachedStateMsg[expectedState], nodeName)
		return false, nil
	})
	if err != nil {
		return err
	}

	if currentState != expectedState {
		err = fmt.Errorf("Gave up waiting for Volume %q to be %s %q after %v", volumePath, attachedStateMsg[expectedState], nodeName, timeout)
	}
	return err
}

// Wait until vsphere vmdk is attached from the given node or time out after 6 minutes
func waitForVSphereDiskToAttach(volumePath string, nodeName string) error {
	return waitForVSphereDiskStatus(volumePath, nodeName, volumeStateAttached)
}

// Wait until vsphere vmdk is detached from the given node or time out after 6 minutes
func waitForVSphereDiskToDetach(volumePath string, nodeName string) error {
	return waitForVSphereDiskStatus(volumePath, nodeName, volumeStateDetached)
}

// function to create vsphere volume spec with given VMDK volume path, Reclaim Policy and labels
func getVSpherePersistentVolumeSpec(volumePath string, persistentVolumeReclaimPolicy v1.PersistentVolumeReclaimPolicy, labels map[string]string) *v1.PersistentVolume {
	var (
		pvConfig framework.PersistentVolumeConfig
		pv       *v1.PersistentVolume
		claimRef *v1.ObjectReference
	)
	pvConfig = framework.PersistentVolumeConfig{
		NamePrefix: "vspherepv-",
		PVSource: v1.PersistentVolumeSource{
			VsphereVolume: &v1.VsphereVirtualDiskVolumeSource{
				VolumePath: volumePath,
				FSType:     "ext4",
			},
		},
		Prebind: nil,
	}

	pv = &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pvConfig.NamePrefix,
			Annotations: map[string]string{
				util.VolumeGidAnnotationKey: "777",
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: persistentVolumeReclaimPolicy,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("2Gi"),
			},
			PersistentVolumeSource: pvConfig.PVSource,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			ClaimRef: claimRef,
		},
	}
	if labels != nil {
		pv.Labels = labels
	}
	return pv
}

// function to get vsphere persistent volume spec with given selector labels.
func getVSpherePersistentVolumeClaimSpec(namespace string, labels map[string]string) *v1.PersistentVolumeClaim {
	var (
		pvc *v1.PersistentVolumeClaim
	)
	pvc = &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-",
			Namespace:    namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("2Gi"),
				},
			},
		},
	}
	if labels != nil {
		pvc.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	}

	return pvc
}

// function to write content to the volume backed by given PVC
func writeContentToVSpherePV(client clientset.Interface, pvc *v1.PersistentVolumeClaim, expectedContent string) {
	utils.RunInPodWithVolume(client, pvc.Namespace, pvc.Name, "echo "+expectedContent+" > /mnt/test/data")
	framework.Logf("Done with writing content to volume")
}

// function to verify content is matching on the volume backed for given PVC
func verifyContentOfVSpherePV(client clientset.Interface, pvc *v1.PersistentVolumeClaim, expectedContent string) {
	utils.RunInPodWithVolume(client, pvc.Namespace, pvc.Name, "grep '"+expectedContent+"' /mnt/test/data")
	framework.Logf("Successfully verified content of the volume")
}

func getVSphereStorageClassSpec(name string, scParameters map[string]string) *storage.StorageClass {
	var sc *storage.StorageClass

	sc = &storage.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind: "StorageClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Provisioner: "kubernetes.io/vsphere-volume",
	}
	if scParameters != nil {
		sc.Parameters = scParameters
	}
	return sc
}

func getVSphereClaimSpecWithStorageClass(ns string, diskSize string, storageclass *storage.StorageClass) *v1.PersistentVolumeClaim {
	claim := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-",
			Namespace:    ns,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse(diskSize),
				},
			},
			StorageClassName: &(storageclass.Name),
		},
	}
	return claim
}

// func to get pod spec with given volume claim, node selector labels and command
func getVSpherePodSpecWithClaim(claimName string, nodeSelectorKV map[string]string, command string) *v1.Pod {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pod-pvc-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "volume-tester",
					Image:   "busybox",
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", command},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "my-volume",
							MountPath: "/mnt/test",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
			Volumes: []v1.Volume{
				{
					Name: "my-volume",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: claimName,
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}
	if nodeSelectorKV != nil {
		pod.Spec.NodeSelector = nodeSelectorKV
	}
	return pod
}

// func to get pod spec with given volume paths, node selector lables and container commands
func getVSpherePodSpecWithVolumePaths(volumePaths []string, keyValuelabel map[string]string, commands []string) *v1.Pod {
	var volumeMounts []v1.VolumeMount
	var volumes []v1.Volume

	for index, volumePath := range volumePaths {
		name := fmt.Sprintf("volume%v", index+1)
		volumeMounts = append(volumeMounts, v1.VolumeMount{Name: name, MountPath: "/mnt/" + name})
		vsphereVolume := new(v1.VsphereVirtualDiskVolumeSource)
		vsphereVolume.VolumePath = volumePath
		vsphereVolume.FSType = "ext4"
		volumes = append(volumes, v1.Volume{Name: name})
		volumes[index].VolumeSource.VsphereVolume = vsphereVolume
	}

	if commands == nil || len(commands) == 0 {
		commands = []string{
			"/bin/sh",
			"-c",
			"while true; do sleep 2; done",
		}
	}
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "vsphere-e2e-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:         "vsphere-e2e-container-" + string(uuid.NewUUID()),
					Image:        "busybox",
					Command:      commands,
					VolumeMounts: volumeMounts,
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
			Volumes:       volumes,
		},
	}

	if keyValuelabel != nil {
		pod.Spec.NodeSelector = keyValuelabel
	}
	return pod
}

func verifyFilesExistOnVSphereVolume(namespace string, podName string, filePaths ...string) {
	for _, filePath := range filePaths {
		_, err := framework.RunKubectl("exec", fmt.Sprintf("--namespace=%s", namespace), podName, "--", "/bin/ls", filePath)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to verify file: %q on the pod: %q", filePath, podName))
	}
}

func createEmptyFilesOnVSphereVolume(namespace string, podName string, filePaths []string) {
	for _, filePath := range filePaths {
		err := framework.CreateEmptyFileOnPod(namespace, podName, filePath)
		Expect(err).NotTo(HaveOccurred())
	}
}

// verify volumes are attached to the node and are accessible in pod
func verifyVSphereVolumesAccessible(c clientset.Interface, pod *v1.Pod, persistentvolumes []*v1.PersistentVolume) {
	nodeName := pod.Spec.NodeName
	namespace := pod.Namespace
	for index, pv := range persistentvolumes {
		// Verify disks are attached to the node
		isAttached, err := diskIsAttached(pv.Spec.VsphereVolume.VolumePath, nodeName)
		Expect(err).NotTo(HaveOccurred())
		Expect(isAttached).To(BeTrue(), fmt.Sprintf("disk %v is not attached with the node", pv.Spec.VsphereVolume.VolumePath))
		// Verify Volumes are accessible
		filepath := filepath.Join("/mnt/", fmt.Sprintf("volume%v", index+1), "/emptyFile.txt")
		_, err = framework.LookForStringInPodExec(namespace, pod.Name, []string{"/bin/touch", filepath}, "", time.Minute)
		Expect(err).NotTo(HaveOccurred())
	}
}

// Get vSphere Volume Path from PVC
func getvSphereVolumePathFromClaim(client clientset.Interface, namespace string, claimName string) string {
	pvclaim, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(claimName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	pv, err := client.CoreV1().PersistentVolumes().Get(pvclaim.Spec.VolumeName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	return pv.Spec.VsphereVolume.VolumePath
}

// Get canonical volume path for volume Path.
// Example1: The canonical path for volume path - [vsanDatastore] kubevols/volume.vmdk will be [vsanDatastore] 25d8b159-948c-4b73-e499-02001ad1b044/volume.vmdk
// Example2: The canonical path for volume path - [vsanDatastore] 25d8b159-948c-4b73-e499-02001ad1b044/volume.vmdk will be same as volume Path.
func getCanonicalVolumePath(ctx context.Context, dc *object.Datacenter, volumePath string) (string, error) {
	var folderID string
	canonicalVolumePath := volumePath
	dsPathObj, err := getDatastorePathObjFromVMDiskPath(volumePath)
	if err != nil {
		return "", err
	}
	dsPath := strings.Split(strings.TrimSpace(dsPathObj.Path), "/")
	if len(dsPath) <= 1 {
		return canonicalVolumePath, nil
	}
	datastore := dsPathObj.Datastore
	dsFolder := dsPath[0]
	// Get the datastore folder ID if datastore or folder doesn't exist in datastoreFolderIDMap
	if !isValidUUID(dsFolder) {
		dummyDiskVolPath := "[" + datastore + "] " + dsFolder + "/" + DummyDiskName
		// Querying a non-existent dummy disk on the datastore folder.
		// It would fail and return an folder ID in the error message.
		_, err := getVirtualDiskPage83Data(ctx, dc, dummyDiskVolPath)
		if err != nil {
			re := regexp.MustCompile("File (.*?) was not found")
			match := re.FindStringSubmatch(err.Error())
			canonicalVolumePath = match[1]
		}
	}
	diskPath := getPathFromVMDiskPath(canonicalVolumePath)
	if diskPath == "" {
		return "", fmt.Errorf("Failed to parse canonicalVolumePath: %s in getcanonicalVolumePath method", canonicalVolumePath)
	}
	folderID = strings.Split(strings.TrimSpace(diskPath), "/")[0]
	canonicalVolumePath = strings.Replace(volumePath, dsFolder, folderID, 1)
	return canonicalVolumePath, nil
}

// getPathFromVMDiskPath retrieves the path from VM Disk Path.
// Example: For vmDiskPath - [vsanDatastore] kubevols/volume.vmdk, the path is kubevols/volume.vmdk
func getPathFromVMDiskPath(vmDiskPath string) string {
	datastorePathObj := new(object.DatastorePath)
	isSuccess := datastorePathObj.FromString(vmDiskPath)
	if !isSuccess {
		framework.Logf("Failed to parse vmDiskPath: %s", vmDiskPath)
		return ""
	}
	return datastorePathObj.Path
}

//getDatastorePathObjFromVMDiskPath gets the datastorePathObj from VM disk path.
func getDatastorePathObjFromVMDiskPath(vmDiskPath string) (*object.DatastorePath, error) {
	datastorePathObj := new(object.DatastorePath)
	isSuccess := datastorePathObj.FromString(vmDiskPath)
	if !isSuccess {
		framework.Logf("Failed to parse volPath: %s", vmDiskPath)
		return nil, fmt.Errorf("Failed to parse volPath: %s", vmDiskPath)
	}
	return datastorePathObj, nil
}

// getVirtualDiskPage83Data gets the virtual disk UUID by diskPath
func getVirtualDiskPage83Data(ctx context.Context, dc *object.Datacenter, diskPath string) (string, error) {
	if len(diskPath) > 0 && filepath.Ext(diskPath) != ".vmdk" {
		diskPath += ".vmdk"
	}
	vdm := object.NewVirtualDiskManager(dc.Client())
	// Returns uuid of vmdk virtual disk
	diskUUID, err := vdm.QueryVirtualDiskUuid(ctx, diskPath, dc)

	if err != nil {
		glog.Warningf("QueryVirtualDiskUuid failed for diskPath: %q. err: %+v", diskPath, err)
		return "", err
	}
	diskUUID = formatVirtualDiskUUID(diskUUID)
	return diskUUID, nil
}

// formatVirtualDiskUUID removes any spaces and hyphens in UUID
// Example UUID input is 42375390-71f9-43a3-a770-56803bcd7baa and output after format is 4237539071f943a3a77056803bcd7baa
func formatVirtualDiskUUID(uuid string) string {
	uuidwithNoSpace := strings.Replace(uuid, " ", "", -1)
	uuidWithNoHypens := strings.Replace(uuidwithNoSpace, "-", "", -1)
	return strings.ToLower(uuidWithNoHypens)
}

//isValidUUID checks if the string is a valid UUID.
func isValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

// removeStorageClusterORFolderNameFromVDiskPath removes the cluster or folder path from the vDiskPath
// for vDiskPath [DatastoreCluster/sharedVmfs-0] kubevols/e2e-vmdk-1234.vmdk, return value is [sharedVmfs-0] kubevols/e2e-vmdk-1234.vmdk
// for vDiskPath [sharedVmfs-0] kubevols/e2e-vmdk-1234.vmdk, return value remains same [sharedVmfs-0] kubevols/e2e-vmdk-1234.vmdk
func removeStorageClusterORFolderNameFromVDiskPath(vDiskPath string) string {
	datastore := regexp.MustCompile("\\[(.*?)\\]").FindStringSubmatch(vDiskPath)[1]
	if filepath.Base(datastore) != datastore {
		vDiskPath = strings.Replace(vDiskPath, datastore, filepath.Base(datastore), 1)
	}
	return vDiskPath
}

// getVirtualDeviceByPath gets the virtual device by path
func getVirtualDeviceByPath(ctx context.Context, vm *object.VirtualMachine, diskPath string) (vim25types.BaseVirtualDevice, error) {
	vmDevices, err := vm.Device(ctx)
	if err != nil {
		framework.Logf("Failed to get the devices for VM: %q. err: %+v", vm.InventoryPath, err)
		return nil, err
	}

	// filter vm devices to retrieve device for the given vmdk file identified by disk path
	for _, device := range vmDevices {
		if vmDevices.TypeName(device) == "VirtualDisk" {
			virtualDevice := device.GetVirtualDevice()
			if backing, ok := virtualDevice.Backing.(*vim25types.VirtualDiskFlatVer2BackingInfo); ok {
				if matchVirtualDiskAndVolPath(backing.FileName, diskPath) {
					framework.Logf("Found VirtualDisk backing with filename %q for diskPath %q", backing.FileName, diskPath)
					return device, nil
				} else {
					framework.Logf("VirtualDisk backing filename %q does not match with diskPath %q", backing.FileName, diskPath)
				}
			}
		}
	}
	return nil, nil
}

func matchVirtualDiskAndVolPath(diskPath, volPath string) bool {
	fileExt := ".vmdk"
	diskPath = strings.TrimSuffix(diskPath, fileExt)
	volPath = strings.TrimSuffix(volPath, fileExt)
	return diskPath == volPath
}

// convertVolPathsToDevicePaths removes cluster or folder path from volPaths and convert to canonicalPath
func convertVolPathsToDevicePaths(ctx context.Context, nodeVolumes map[string][]string) (map[string][]string, error) {
	vmVolumes := make(map[string][]string)
	for nodeName, volPaths := range nodeVolumes {
		nodeInfo := TestContext.NodeMapper.GetNodeInfo(nodeName)
		datacenter := nodeInfo.VSphere.GetDatacenterFromObjectReference(ctx, nodeInfo.DataCenterRef)
		for i, volPath := range volPaths {
			deviceVolPath, err := convertVolPathToDevicePath(ctx, datacenter, volPath)
			if err != nil {
				framework.Logf("Failed to convert vsphere volume path %s to device path for volume %s. err: %+v", volPath, deviceVolPath, err)
				return nil, err
			}
			volPaths[i] = deviceVolPath
		}
		vmVolumes[nodeName] = volPaths
	}
	return vmVolumes, nil
}

// convertVolPathToDevicePath takes volPath and returns canonical volume path
func convertVolPathToDevicePath(ctx context.Context, dc *object.Datacenter, volPath string) (string, error) {
	volPath = removeStorageClusterORFolderNameFromVDiskPath(volPath)
	// Get the canonical volume path for volPath.
	canonicalVolumePath, err := getCanonicalVolumePath(ctx, dc, volPath)
	if err != nil {
		framework.Logf("Failed to get canonical vsphere volume path for volume: %s. err: %+v", volPath, err)
		return "", err
	}
	// Check if the volume path contains .vmdk extension. If not, add the extension and update the nodeVolumes Map
	if len(canonicalVolumePath) > 0 && filepath.Ext(canonicalVolumePath) != ".vmdk" {
		canonicalVolumePath += ".vmdk"
	}
	return canonicalVolumePath, nil
}

// get .vmx file path for a virtual machine
func getVMXFilePath(vmObject *object.VirtualMachine) (vmxPath string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var nodeVM mo.VirtualMachine
	err := vmObject.Properties(ctx, vmObject.Reference(), []string{"config.files"}, &nodeVM)
	Expect(err).NotTo(HaveOccurred())
	Expect(nodeVM.Config).NotTo(BeNil())

	vmxPath = nodeVM.Config.Files.VmPathName
	framework.Logf("vmx file path is %s", vmxPath)
	return vmxPath
}

// verify ready node count. Try upto 3 minutes. Return true if count is expected count
func verifyReadyNodeCount(client clientset.Interface, expectedNodes int) bool {
	numNodes := 0
	for i := 0; i < 36; i++ {
		nodeList := framework.GetReadySchedulableNodesOrDie(client)
		Expect(nodeList.Items).NotTo(BeEmpty(), "Unable to find ready and schedulable Node")

		numNodes = len(nodeList.Items)
		if numNodes == expectedNodes {
			break
		}
		time.Sleep(5 * time.Second)
	}
	return (numNodes == expectedNodes)
}

// poweroff nodeVM and confirm the poweroff state
func poweroffNodeVM(nodeName string, vm *object.VirtualMachine) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	framework.Logf("Powering off node VM %s", nodeName)

	_, err := vm.PowerOff(ctx)
	Expect(err).NotTo(HaveOccurred())
	err = vm.WaitForPowerState(ctx, vimtypes.VirtualMachinePowerStatePoweredOff)
	Expect(err).NotTo(HaveOccurred(), "Unable to power off the node")
}

// poweron nodeVM and confirm the poweron state
func poweronNodeVM(nodeName string, vm *object.VirtualMachine) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	framework.Logf("Powering on node VM %s", nodeName)

	vm.PowerOn(ctx)
	err := vm.WaitForPowerState(ctx, vimtypes.VirtualMachinePowerStatePoweredOn)
	Expect(err).NotTo(HaveOccurred(), "Unable to power on the node")
}

// unregister a nodeVM from VC
func unregisterNodeVM(nodeName string, vm *object.VirtualMachine) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poweroffNodeVM(nodeName, vm)

	framework.Logf("Unregistering node VM %s", nodeName)
	err := vm.Unregister(ctx)
	Expect(err).NotTo(HaveOccurred(), "Unable to unregister the node")
}

// register a nodeVM into a VC
func registerNodeVM(nodeName, workingDir, vmxFilePath string, rpool *object.ResourcePool, host *object.HostSystem) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	framework.Logf("Registering node VM %s with vmx file path %s", nodeName, vmxFilePath)

	nodeInfo := TestContext.NodeMapper.GetNodeInfo(nodeName)
	finder := find.NewFinder(nodeInfo.VSphere.Client.Client, false)

	vmFolder, err := finder.FolderOrDefault(ctx, workingDir)
	Expect(err).NotTo(HaveOccurred())

	registerTask, err := vmFolder.RegisterVM(ctx, vmxFilePath, nodeName, false, rpool, host)
	Expect(err).NotTo(HaveOccurred())
	err = registerTask.Wait(ctx)
	Expect(err).NotTo(HaveOccurred())

	vmPath := filepath.Join(workingDir, nodeName)
	vm, err := finder.VirtualMachine(ctx, vmPath)
	Expect(err).NotTo(HaveOccurred())

	poweronNodeVM(nodeName, vm)
}

// disksAreAttached takes map of node and it's volumes and returns map of node, its volumes and attachment state
func disksAreAttached(nodeVolumes map[string][]string) (nodeVolumesAttachMap map[string]map[string]bool, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	disksAttached := make(map[string]map[string]bool)
	if len(nodeVolumes) == 0 {
		return disksAttached, nil
	}
	// Convert VolPaths into canonical form so that it can be compared with the VM device path.
	vmVolumes, err := convertVolPathsToDevicePaths(ctx, nodeVolumes)
	if err != nil {
		framework.Logf("Failed to convert volPaths to devicePaths: %+v. err: %+v", nodeVolumes, err)
		return nil, err
	}
	for vm, volumes := range vmVolumes {
		volumeAttachedMap := make(map[string]bool)
		for _, volume := range volumes {
			attached, err := diskIsAttached(volume, vm)
			if err != nil {
				return nil, err
			}
			volumeAttachedMap[volume] = attached
		}
		nodeVolumesAttachMap[vm] = volumeAttachedMap
	}
	return disksAttached, nil
}

// diskIsAttached returns if disk is attached to the VM using controllers supported by the plugin.
func diskIsAttached(volPath string, nodeName string) (bool, error) {
	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nodeInfo := TestContext.NodeMapper.GetNodeInfo(nodeName)
	Connect(ctx, nodeInfo.VSphere)
	vm := object.NewVirtualMachine(nodeInfo.VSphere.Client.Client, nodeInfo.VirtualMachineRef)
	volPath = removeStorageClusterORFolderNameFromVDiskPath(volPath)
	device, err := getVirtualDeviceByPath(ctx, vm, volPath)
	if err != nil {
		framework.Logf("diskIsAttached failed to determine whether disk %q is still attached on node %q",
			volPath,
			nodeName)
		return false, err
	}
	if device == nil {
		return false, nil
	}
	framework.Logf("diskIsAttached found the disk %q attached on node %q", volPath, nodeName)
	return true, nil
}

// getUUIDFromProviderID strips ProviderPrefix - "vsphere://" from the providerID
// this gives the VM UUID which can be used to find Node VM from vCenter
func getUUIDFromProviderID(providerID string) string {
	return strings.TrimPrefix(providerID, ProviderPrefix)
}

// GetAllReadySchedulableNodeInfos returns NodeInfo objects for all nodes with Ready and schedulable state
func GetReadySchedulableNodeInfos() []*NodeInfo {
	nodeList := framework.GetReadySchedulableNodesOrDie(f.ClientSet)
	Expect(nodeList.Items).NotTo(BeEmpty(), "Unable to find ready and schedulable Node")
	var nodesInfo []*NodeInfo
	for _, node := range nodeList.Items {
		nodeInfo := TestContext.NodeMapper.GetNodeInfo(node.Name)
		if nodeInfo != nil {
			nodesInfo = append(nodesInfo, nodeInfo)
		}
	}
	return nodesInfo
}

// GetReadySchedulableRandomNodeInfo returns NodeInfo object for one of the Ready and Schedulable Node.
// if multiple nodes are present with Ready and Scheduable state then one of the Node is selected randomly
// and it's associated NodeInfo object is returned.
func GetReadySchedulableRandomNodeInfo() *NodeInfo {
	nodesInfo := GetReadySchedulableNodeInfos()
	rand.Seed(time.Now().Unix())
	Expect(nodesInfo).NotTo(BeEmpty())
	return nodesInfo[rand.Int()%len(nodesInfo)]
}

// invokeVCenterServiceControl invokes the given command for the given service
// via service-control on the given vCenter host over SSH.
func invokeVCenterServiceControl(command, service, host string) error {
	sshCmd := fmt.Sprintf("service-control --%s %s", command, service)
	framework.Logf("Invoking command %v on vCenter host %v", sshCmd, host)
	result, err := framework.SSH(sshCmd, host, framework.TestContext.Provider)
	if err != nil || result.Code != 0 {
		framework.LogSSHResult(result)
		return fmt.Errorf("couldn't execute command: %s on vCenter host: %v", sshCmd, err)
	}
	return nil
}

// expectVolumeToBeAttached checks if the given Volume is attached to the given
// Node, else fails.
func expectVolumeToBeAttached(nodeName, volumePath string) {
	isAttached, err := diskIsAttached(volumePath, nodeName)
	Expect(err).NotTo(HaveOccurred())
	Expect(isAttached).To(BeTrue(), fmt.Sprintf("disk: %s is not attached with the node", volumePath))
}

// expectVolumesToBeAttached checks if the given Volumes are attached to the
// corresponding set of Nodes, else fails.
func expectVolumesToBeAttached(pods []*v1.Pod, volumePaths []string) {
	for i, pod := range pods {
		nodeName := pod.Spec.NodeName
		volumePath := volumePaths[i]
		By(fmt.Sprintf("Verifying that volume %v is attached to node %v", volumePath, nodeName))
		expectVolumeToBeAttached(nodeName, volumePath)
	}
}

// expectFilesToBeAccessible checks if the given files are accessible on the
// corresponding set of Nodes, else fails.
func expectFilesToBeAccessible(namespace string, pods []*v1.Pod, filePaths []string) {
	for i, pod := range pods {
		podName := pod.Name
		filePath := filePaths[i]
		By(fmt.Sprintf("Verifying that file %v is accessible on pod %v", filePath, podName))
		verifyFilesExistOnVSphereVolume(namespace, podName, filePath)
	}
}

// writeContentToPodFile writes the given content to the specified file.
func writeContentToPodFile(namespace, podName, filePath, content string) error {
	_, err := framework.RunKubectl("exec", fmt.Sprintf("--namespace=%s", namespace), podName,
		"--", "/bin/sh", "-c", fmt.Sprintf("echo '%s' > %s", content, filePath))
	return err
}

// expectFileContentToMatch checks if a given file contains the specified
// content, else fails.
func expectFileContentToMatch(namespace, podName, filePath, content string) {
	_, err := framework.RunKubectl("exec", fmt.Sprintf("--namespace=%s", namespace), podName,
		"--", "/bin/sh", "-c", fmt.Sprintf("grep '%s' %s", content, filePath))
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to match content of file: %q on the pod: %q", filePath, podName))
}

// expectFileContentsToMatch checks if the given contents match the ones present
// in corresponding files on respective Pods, else fails.
func expectFileContentsToMatch(namespace string, pods []*v1.Pod, filePaths []string, contents []string) {
	for i, pod := range pods {
		podName := pod.Name
		filePath := filePaths[i]
		By(fmt.Sprintf("Matching file content for %v on pod %v", filePath, podName))
		expectFileContentToMatch(namespace, podName, filePath, contents[i])
	}
}
