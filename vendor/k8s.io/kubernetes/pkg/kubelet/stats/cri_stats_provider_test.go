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

package stats

import (
	"math/rand"
	"runtime"
	"testing"
	"time"

	cadvisorfs "github.com/google/cadvisor/fs"
	cadvisorapiv2 "github.com/google/cadvisor/info/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeapi "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
	critest "k8s.io/kubernetes/pkg/kubelet/apis/cri/testing"
	statsapi "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
	cadvisortest "k8s.io/kubernetes/pkg/kubelet/cadvisor/testing"
	"k8s.io/kubernetes/pkg/kubelet/cm"
	kubecontainertest "k8s.io/kubernetes/pkg/kubelet/container/testing"
	"k8s.io/kubernetes/pkg/kubelet/kuberuntime"
	"k8s.io/kubernetes/pkg/kubelet/leaky"
	kubepodtest "k8s.io/kubernetes/pkg/kubelet/pod/testing"
	serverstats "k8s.io/kubernetes/pkg/kubelet/server/stats"
	"k8s.io/kubernetes/pkg/volume"
)

const (
	offsetInodeUsage = iota
	offsetUsage
)

func TestCRIListPodStats(t *testing.T) {
	const (
		seedRoot       = 0
		seedKubelet    = 200
		seedMisc       = 300
		seedSandbox0   = 1000
		seedContainer0 = 2000
		seedSandbox1   = 3000
		seedContainer1 = 4000
		seedContainer2 = 5000
		seedSandbox2   = 6000
		seedContainer3 = 7000
	)

	const (
		pName0 = "pod0"
		pName1 = "pod1"
		pName2 = "pod2"
	)

	const (
		cName0 = "container0-name"
		cName1 = "container1-name"
		cName2 = "container2-name"
		cName3 = "container3-name"
	)

	var (
		imageFsMountpoint = "/test/mount/point"
		unknownMountpoint = "/unknown/mount/point"
		imageFsInfo       = getTestFsInfo(2000)
		rootFsInfo        = getTestFsInfo(1000)

		sandbox0           = makeFakePodSandbox("sandbox0-name", "sandbox0-uid", "sandbox0-ns")
		sandbox0Cgroup     = "/" + cm.GetPodCgroupNameSuffix(types.UID(sandbox0.PodSandboxStatus.Metadata.Uid))
		container0         = makeFakeContainer(sandbox0, cName0, 0, false)
		containerStats0    = makeFakeContainerStats(container0, imageFsMountpoint)
		containerLogStats0 = makeFakeLogStats(1000)
		container1         = makeFakeContainer(sandbox0, cName1, 0, false)
		containerStats1    = makeFakeContainerStats(container1, unknownMountpoint)
		containerLogStats1 = makeFakeLogStats(2000)

		sandbox1           = makeFakePodSandbox("sandbox1-name", "sandbox1-uid", "sandbox1-ns")
		sandbox1Cgroup     = "/" + cm.GetPodCgroupNameSuffix(types.UID(sandbox1.PodSandboxStatus.Metadata.Uid))
		container2         = makeFakeContainer(sandbox1, cName2, 0, false)
		containerStats2    = makeFakeContainerStats(container2, imageFsMountpoint)
		containerLogStats2 = makeFakeLogStats(3000)

		sandbox2           = makeFakePodSandbox("sandbox2-name", "sandbox2-uid", "sandbox2-ns")
		sandbox2Cgroup     = "/" + cm.GetPodCgroupNameSuffix(types.UID(sandbox2.PodSandboxStatus.Metadata.Uid))
		container3         = makeFakeContainer(sandbox2, cName3, 0, true)
		containerStats3    = makeFakeContainerStats(container3, imageFsMountpoint)
		container4         = makeFakeContainer(sandbox2, cName3, 1, false)
		containerStats4    = makeFakeContainerStats(container4, imageFsMountpoint)
		containerLogStats4 = makeFakeLogStats(4000)
	)

	var (
		mockCadvisor       = new(cadvisortest.Mock)
		mockRuntimeCache   = new(kubecontainertest.MockRuntimeCache)
		mockPodManager     = new(kubepodtest.MockManager)
		resourceAnalyzer   = new(fakeResourceAnalyzer)
		fakeRuntimeService = critest.NewFakeRuntimeService()
		fakeImageService   = critest.NewFakeImageService()
	)

	infos := map[string]cadvisorapiv2.ContainerInfo{
		"/":                           getTestContainerInfo(seedRoot, "", "", ""),
		"/kubelet":                    getTestContainerInfo(seedKubelet, "", "", ""),
		"/system":                     getTestContainerInfo(seedMisc, "", "", ""),
		sandbox0.PodSandboxStatus.Id:  getTestContainerInfo(seedSandbox0, pName0, sandbox0.PodSandboxStatus.Metadata.Namespace, leaky.PodInfraContainerName),
		sandbox0Cgroup:                getTestContainerInfo(seedSandbox0, "", "", ""),
		container0.ContainerStatus.Id: getTestContainerInfo(seedContainer0, pName0, sandbox0.PodSandboxStatus.Metadata.Namespace, cName0),
		container1.ContainerStatus.Id: getTestContainerInfo(seedContainer1, pName0, sandbox0.PodSandboxStatus.Metadata.Namespace, cName1),
		sandbox1.PodSandboxStatus.Id:  getTestContainerInfo(seedSandbox1, pName1, sandbox1.PodSandboxStatus.Metadata.Namespace, leaky.PodInfraContainerName),
		sandbox1Cgroup:                getTestContainerInfo(seedSandbox1, "", "", ""),
		container2.ContainerStatus.Id: getTestContainerInfo(seedContainer2, pName1, sandbox1.PodSandboxStatus.Metadata.Namespace, cName2),
		sandbox2.PodSandboxStatus.Id:  getTestContainerInfo(seedSandbox2, pName2, sandbox2.PodSandboxStatus.Metadata.Namespace, leaky.PodInfraContainerName),
		sandbox2Cgroup:                getTestContainerInfo(seedSandbox2, "", "", ""),
		container4.ContainerStatus.Id: getTestContainerInfo(seedContainer3, pName2, sandbox2.PodSandboxStatus.Metadata.Namespace, cName3),
	}

	options := cadvisorapiv2.RequestOptions{
		IdType:    cadvisorapiv2.TypeName,
		Count:     2,
		Recursive: true,
	}

	mockCadvisor.
		On("ContainerInfoV2", "/", options).Return(infos, nil).
		On("RootFsInfo").Return(rootFsInfo, nil).
		On("GetDirFsInfo", imageFsMountpoint).Return(imageFsInfo, nil).
		On("GetDirFsInfo", unknownMountpoint).Return(cadvisorapiv2.FsInfo{}, cadvisorfs.ErrNoSuchDevice)
	fakeRuntimeService.SetFakeSandboxes([]*critest.FakePodSandbox{
		sandbox0, sandbox1, sandbox2,
	})
	fakeRuntimeService.SetFakeContainers([]*critest.FakeContainer{
		container0, container1, container2, container3, container4,
	})
	fakeRuntimeService.SetFakeContainerStats([]*runtimeapi.ContainerStats{
		containerStats0, containerStats1, containerStats2, containerStats3, containerStats4,
	})

	ephemeralVolumes := makeFakeVolumeStats([]string{"ephVolume1, ephVolumes2"})
	persistentVolumes := makeFakeVolumeStats([]string{"persisVolume1, persisVolumes2"})
	resourceAnalyzer.podVolumeStats = serverstats.PodVolumeStats{
		EphemeralVolumes:  ephemeralVolumes,
		PersistentVolumes: persistentVolumes,
	}

	fakeLogStats := map[string]*volume.Metrics{
		kuberuntime.BuildContainerLogsDirectory(types.UID("sandbox0-uid"), cName0): containerLogStats0,
		kuberuntime.BuildContainerLogsDirectory(types.UID("sandbox0-uid"), cName1): containerLogStats1,
		kuberuntime.BuildContainerLogsDirectory(types.UID("sandbox1-uid"), cName2): containerLogStats2,
		kuberuntime.BuildContainerLogsDirectory(types.UID("sandbox2-uid"), cName3): containerLogStats4,
	}
	fakeLogStatsProvider := NewFakeLogMetricsService(fakeLogStats)

	provider := NewCRIStatsProvider(
		mockCadvisor,
		resourceAnalyzer,
		mockPodManager,
		mockRuntimeCache,
		fakeRuntimeService,
		fakeImageService,
		fakeLogStatsProvider,
	)

	stats, err := provider.ListPodStats()
	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(3, len(stats))

	podStatsMap := make(map[statsapi.PodReference]statsapi.PodStats)
	for _, s := range stats {
		podStatsMap[s.PodRef] = s
	}

	p0 := podStatsMap[statsapi.PodReference{Name: "sandbox0-name", UID: "sandbox0-uid", Namespace: "sandbox0-ns"}]
	assert.Equal(sandbox0.CreatedAt, p0.StartTime.UnixNano())
	assert.Equal(2, len(p0.Containers))

	checkEphemeralStorageStats(assert, p0, ephemeralVolumes, []*runtimeapi.ContainerStats{containerStats0, containerStats1},
		[]*volume.Metrics{containerLogStats0, containerLogStats1})

	containerStatsMap := make(map[string]statsapi.ContainerStats)
	for _, s := range p0.Containers {
		containerStatsMap[s.Name] = s
	}

	c0 := containerStatsMap[cName0]
	assert.Equal(container0.CreatedAt, c0.StartTime.UnixNano())
	checkCRICPUAndMemoryStats(assert, c0, infos[container0.ContainerStatus.Id].Stats[0])
	checkCRIRootfsStats(assert, c0, containerStats0, &imageFsInfo)
	checkCRILogsStats(assert, c0, &rootFsInfo, containerLogStats0)
	c1 := containerStatsMap[cName1]
	assert.Equal(container1.CreatedAt, c1.StartTime.UnixNano())
	checkCRICPUAndMemoryStats(assert, c1, infos[container1.ContainerStatus.Id].Stats[0])
	checkCRIRootfsStats(assert, c1, containerStats1, nil)
	checkCRILogsStats(assert, c1, &rootFsInfo, containerLogStats1)
	checkCRINetworkStats(assert, p0.Network, infos[sandbox0.PodSandboxStatus.Id].Stats[0].Network)
	checkCRIPodCPUAndMemoryStats(assert, p0, infos[sandbox0Cgroup].Stats[0])

	p1 := podStatsMap[statsapi.PodReference{Name: "sandbox1-name", UID: "sandbox1-uid", Namespace: "sandbox1-ns"}]
	assert.Equal(sandbox1.CreatedAt, p1.StartTime.UnixNano())
	assert.Equal(1, len(p1.Containers))

	checkEphemeralStorageStats(assert, p1, ephemeralVolumes, []*runtimeapi.ContainerStats{containerStats2}, []*volume.Metrics{containerLogStats2})
	c2 := p1.Containers[0]
	assert.Equal(cName2, c2.Name)
	assert.Equal(container2.CreatedAt, c2.StartTime.UnixNano())
	checkCRICPUAndMemoryStats(assert, c2, infos[container2.ContainerStatus.Id].Stats[0])
	checkCRIRootfsStats(assert, c2, containerStats2, &imageFsInfo)
	checkCRILogsStats(assert, c2, &rootFsInfo, containerLogStats2)
	checkCRINetworkStats(assert, p1.Network, infos[sandbox1.PodSandboxStatus.Id].Stats[0].Network)
	checkCRIPodCPUAndMemoryStats(assert, p1, infos[sandbox1Cgroup].Stats[0])

	p2 := podStatsMap[statsapi.PodReference{Name: "sandbox2-name", UID: "sandbox2-uid", Namespace: "sandbox2-ns"}]
	assert.Equal(sandbox2.CreatedAt, p2.StartTime.UnixNano())
	assert.Equal(1, len(p2.Containers))

	checkEphemeralStorageStats(assert, p2, ephemeralVolumes, []*runtimeapi.ContainerStats{containerStats4}, []*volume.Metrics{containerLogStats4})

	c3 := p2.Containers[0]
	assert.Equal(cName3, c3.Name)
	assert.Equal(container4.CreatedAt, c3.StartTime.UnixNano())
	checkCRICPUAndMemoryStats(assert, c3, infos[container4.ContainerStatus.Id].Stats[0])
	checkCRIRootfsStats(assert, c3, containerStats4, &imageFsInfo)

	checkCRILogsStats(assert, c3, &rootFsInfo, containerLogStats4)
	checkCRINetworkStats(assert, p2.Network, infos[sandbox2.PodSandboxStatus.Id].Stats[0].Network)
	checkCRIPodCPUAndMemoryStats(assert, p2, infos[sandbox2Cgroup].Stats[0])

	mockCadvisor.AssertExpectations(t)
}

func TestCRIImagesFsStats(t *testing.T) {
	var (
		imageFsMountpoint = "/test/mount/point"
		imageFsInfo       = getTestFsInfo(2000)
		imageFsUsage      = makeFakeImageFsUsage(imageFsMountpoint)
	)
	var (
		mockCadvisor         = new(cadvisortest.Mock)
		mockRuntimeCache     = new(kubecontainertest.MockRuntimeCache)
		mockPodManager       = new(kubepodtest.MockManager)
		resourceAnalyzer     = new(fakeResourceAnalyzer)
		fakeRuntimeService   = critest.NewFakeRuntimeService()
		fakeImageService     = critest.NewFakeImageService()
		fakeLogStatsProvider = NewFakeLogMetricsService(nil)
	)

	mockCadvisor.On("GetDirFsInfo", imageFsMountpoint).Return(imageFsInfo, nil)
	fakeImageService.SetFakeFilesystemUsage([]*runtimeapi.FilesystemUsage{
		imageFsUsage,
	})

	provider := NewCRIStatsProvider(
		mockCadvisor,
		resourceAnalyzer,
		mockPodManager,
		mockRuntimeCache,
		fakeRuntimeService,
		fakeImageService,
		fakeLogStatsProvider,
	)

	stats, err := provider.ImageFsStats()
	assert := assert.New(t)
	assert.NoError(err)

	assert.Equal(imageFsUsage.Timestamp, stats.Time.UnixNano())
	assert.Equal(imageFsInfo.Available, *stats.AvailableBytes)
	assert.Equal(imageFsInfo.Capacity, *stats.CapacityBytes)
	assert.Equal(imageFsInfo.InodesFree, stats.InodesFree)
	assert.Equal(imageFsInfo.Inodes, stats.Inodes)
	assert.Equal(imageFsUsage.UsedBytes.Value, *stats.UsedBytes)
	assert.Equal(imageFsUsage.InodesUsed.Value, *stats.InodesUsed)

	mockCadvisor.AssertExpectations(t)
}

func makeFakePodSandbox(name, uid, namespace string) *critest.FakePodSandbox {
	p := &critest.FakePodSandbox{
		PodSandboxStatus: runtimeapi.PodSandboxStatus{
			Metadata: &runtimeapi.PodSandboxMetadata{
				Name:      name,
				Uid:       uid,
				Namespace: namespace,
			},
			State:     runtimeapi.PodSandboxState_SANDBOX_READY,
			CreatedAt: time.Now().UnixNano(),
		},
	}
	p.PodSandboxStatus.Id = critest.BuildSandboxName(p.PodSandboxStatus.Metadata)
	return p
}

func makeFakeContainer(sandbox *critest.FakePodSandbox, name string, attempt uint32, terminated bool) *critest.FakeContainer {
	sandboxID := sandbox.PodSandboxStatus.Id
	c := &critest.FakeContainer{
		SandboxID: sandboxID,
		ContainerStatus: runtimeapi.ContainerStatus{
			Metadata:  &runtimeapi.ContainerMetadata{Name: name, Attempt: attempt},
			Image:     &runtimeapi.ImageSpec{},
			ImageRef:  "fake-image-ref",
			CreatedAt: time.Now().UnixNano(),
		},
	}
	c.ContainerStatus.Labels = map[string]string{
		"io.kubernetes.pod.name":       sandbox.Metadata.Name,
		"io.kubernetes.pod.uid":        sandbox.Metadata.Uid,
		"io.kubernetes.pod.namespace":  sandbox.Metadata.Namespace,
		"io.kubernetes.container.name": name,
	}
	if terminated {
		c.ContainerStatus.State = runtimeapi.ContainerState_CONTAINER_EXITED
	} else {
		c.ContainerStatus.State = runtimeapi.ContainerState_CONTAINER_RUNNING
	}
	c.ContainerStatus.Id = critest.BuildContainerName(c.ContainerStatus.Metadata, sandboxID)
	return c
}

func makeFakeContainerStats(container *critest.FakeContainer, imageFsMountpoint string) *runtimeapi.ContainerStats {
	containerStats := &runtimeapi.ContainerStats{
		Attributes: &runtimeapi.ContainerAttributes{
			Id:       container.ContainerStatus.Id,
			Metadata: container.ContainerStatus.Metadata,
		},
		WritableLayer: &runtimeapi.FilesystemUsage{
			Timestamp:  time.Now().UnixNano(),
			FsId:       &runtimeapi.FilesystemIdentifier{Mountpoint: imageFsMountpoint},
			UsedBytes:  &runtimeapi.UInt64Value{Value: rand.Uint64() / 100},
			InodesUsed: &runtimeapi.UInt64Value{Value: rand.Uint64() / 100},
		},
	}
	if container.State == runtimeapi.ContainerState_CONTAINER_EXITED {
		containerStats.Cpu = nil
		containerStats.Memory = nil
	} else {
		containerStats.Cpu = &runtimeapi.CpuUsage{
			Timestamp:            time.Now().UnixNano(),
			UsageCoreNanoSeconds: &runtimeapi.UInt64Value{Value: rand.Uint64()},
		}
		containerStats.Memory = &runtimeapi.MemoryUsage{
			Timestamp:       time.Now().UnixNano(),
			WorkingSetBytes: &runtimeapi.UInt64Value{Value: rand.Uint64()},
		}
	}
	return containerStats
}

func makeFakeImageFsUsage(fsMountpoint string) *runtimeapi.FilesystemUsage {
	return &runtimeapi.FilesystemUsage{
		Timestamp:  time.Now().UnixNano(),
		FsId:       &runtimeapi.FilesystemIdentifier{Mountpoint: fsMountpoint},
		UsedBytes:  &runtimeapi.UInt64Value{Value: rand.Uint64()},
		InodesUsed: &runtimeapi.UInt64Value{Value: rand.Uint64()},
	}
}

func makeFakeVolumeStats(volumeNames []string) []statsapi.VolumeStats {
	volumes := make([]statsapi.VolumeStats, len(volumeNames))
	availableBytes := rand.Uint64()
	capacityBytes := rand.Uint64()
	usedBytes := rand.Uint64() / 100
	inodes := rand.Uint64()
	inodesFree := rand.Uint64()
	inodesUsed := rand.Uint64() / 100
	for i, name := range volumeNames {
		fsStats := statsapi.FsStats{
			Time:           metav1.NewTime(time.Now()),
			AvailableBytes: &availableBytes,
			CapacityBytes:  &capacityBytes,
			UsedBytes:      &usedBytes,
			Inodes:         &inodes,
			InodesFree:     &inodesFree,
			InodesUsed:     &inodesUsed,
		}
		volumes[i] = statsapi.VolumeStats{
			FsStats: fsStats,
			Name:    name,
		}
	}
	return volumes
}

func checkCRICPUAndMemoryStats(assert *assert.Assertions, actual statsapi.ContainerStats, cs *cadvisorapiv2.ContainerStats) {
	assert.Equal(cs.Timestamp.UnixNano(), actual.CPU.Time.UnixNano())
	assert.Equal(cs.Cpu.Usage.Total, *actual.CPU.UsageCoreNanoSeconds)
	assert.Equal(cs.CpuInst.Usage.Total, *actual.CPU.UsageNanoCores)

	assert.Equal(cs.Memory.Usage, *actual.Memory.UsageBytes)
	assert.Equal(cs.Memory.WorkingSet, *actual.Memory.WorkingSetBytes)
	assert.Equal(cs.Memory.RSS, *actual.Memory.RSSBytes)
	assert.Equal(cs.Memory.ContainerData.Pgfault, *actual.Memory.PageFaults)
	assert.Equal(cs.Memory.ContainerData.Pgmajfault, *actual.Memory.MajorPageFaults)
}

func checkCRIRootfsStats(assert *assert.Assertions, actual statsapi.ContainerStats, cs *runtimeapi.ContainerStats, imageFsInfo *cadvisorapiv2.FsInfo) {
	assert.Equal(cs.WritableLayer.Timestamp, actual.Rootfs.Time.UnixNano())
	if imageFsInfo != nil {
		assert.Equal(imageFsInfo.Available, *actual.Rootfs.AvailableBytes)
		assert.Equal(imageFsInfo.Capacity, *actual.Rootfs.CapacityBytes)
		assert.Equal(*imageFsInfo.InodesFree, *actual.Rootfs.InodesFree)
		assert.Equal(*imageFsInfo.Inodes, *actual.Rootfs.Inodes)
	} else {
		assert.Nil(actual.Rootfs.AvailableBytes)
		assert.Nil(actual.Rootfs.CapacityBytes)
		assert.Nil(actual.Rootfs.InodesFree)
		assert.Nil(actual.Rootfs.Inodes)
	}
	assert.Equal(cs.WritableLayer.UsedBytes.Value, *actual.Rootfs.UsedBytes)
	assert.Equal(cs.WritableLayer.InodesUsed.Value, *actual.Rootfs.InodesUsed)
}

func checkCRILogsStats(assert *assert.Assertions, actual statsapi.ContainerStats, rootFsInfo *cadvisorapiv2.FsInfo, logStats *volume.Metrics) {
	assert.Equal(rootFsInfo.Timestamp, actual.Logs.Time.Time)
	assert.Equal(rootFsInfo.Available, *actual.Logs.AvailableBytes)
	assert.Equal(rootFsInfo.Capacity, *actual.Logs.CapacityBytes)
	assert.Equal(*rootFsInfo.InodesFree, *actual.Logs.InodesFree)
	assert.Equal(*rootFsInfo.Inodes, *actual.Logs.Inodes)
	assert.Equal(uint64(logStats.Used.Value()), *actual.Logs.UsedBytes)
	assert.Equal(uint64(logStats.InodesUsed.Value()), *actual.Logs.InodesUsed)
}

func checkEphemeralStorageStats(assert *assert.Assertions,
	actual statsapi.PodStats,
	volumes []statsapi.VolumeStats,
	containers []*runtimeapi.ContainerStats,
	containerLogStats []*volume.Metrics) {
	var totalUsed, inodesUsed uint64
	for _, container := range containers {
		totalUsed = totalUsed + container.WritableLayer.UsedBytes.Value
		inodesUsed = inodesUsed + container.WritableLayer.InodesUsed.Value
	}

	for _, volume := range volumes {
		totalUsed = totalUsed + *volume.FsStats.UsedBytes
		inodesUsed = inodesUsed + *volume.FsStats.InodesUsed
	}

	for _, logStats := range containerLogStats {
		totalUsed = totalUsed + uint64(logStats.Used.Value())
	}

	assert.Equal(int(totalUsed), int(*actual.EphemeralStorage.UsedBytes))
	assert.Equal(int(inodesUsed), int(*actual.EphemeralStorage.InodesUsed))
}

func checkCRINetworkStats(assert *assert.Assertions, actual *statsapi.NetworkStats, expected *cadvisorapiv2.NetworkStats) {
	assert.Equal(expected.Interfaces[0].RxBytes, *actual.RxBytes)
	assert.Equal(expected.Interfaces[0].RxErrors, *actual.RxErrors)
	assert.Equal(expected.Interfaces[0].TxBytes, *actual.TxBytes)
	assert.Equal(expected.Interfaces[0].TxErrors, *actual.TxErrors)
}

func checkCRIPodCPUAndMemoryStats(assert *assert.Assertions, actual statsapi.PodStats, cs *cadvisorapiv2.ContainerStats) {
	if runtime.GOOS != "linux" {
		return
	}
	assert.Equal(cs.Timestamp.UnixNano(), actual.CPU.Time.UnixNano())
	assert.Equal(cs.Cpu.Usage.Total, *actual.CPU.UsageCoreNanoSeconds)
	assert.Equal(cs.CpuInst.Usage.Total, *actual.CPU.UsageNanoCores)

	assert.Equal(cs.Memory.Usage, *actual.Memory.UsageBytes)
	assert.Equal(cs.Memory.WorkingSet, *actual.Memory.WorkingSetBytes)
	assert.Equal(cs.Memory.RSS, *actual.Memory.RSSBytes)
	assert.Equal(cs.Memory.ContainerData.Pgfault, *actual.Memory.PageFaults)
	assert.Equal(cs.Memory.ContainerData.Pgmajfault, *actual.Memory.MajorPageFaults)
}

func makeFakeLogStats(seed int) *volume.Metrics {
	m := &volume.Metrics{}
	m.Used = resource.NewQuantity(int64(seed+offsetUsage), resource.BinarySI)
	m.InodesUsed = resource.NewQuantity(int64(seed+offsetInodeUsage), resource.BinarySI)
	return m
}
