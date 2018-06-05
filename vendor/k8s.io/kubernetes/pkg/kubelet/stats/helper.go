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
	"fmt"
	"time"

	"github.com/golang/glog"

	cadvisorapiv1 "github.com/google/cadvisor/info/v1"
	cadvisorapiv2 "github.com/google/cadvisor/info/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	statsapi "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
	"k8s.io/kubernetes/pkg/kubelet/cadvisor"
)

// defaultNetworkInterfaceName is used for collectng network stats.
// This logic relies on knowledge of the container runtime implementation and
// is not reliable.
const defaultNetworkInterfaceName = "eth0"

func cadvisorInfoToCPUandMemoryStats(info *cadvisorapiv2.ContainerInfo) (*statsapi.CPUStats, *statsapi.MemoryStats) {
	cstat, found := latestContainerStats(info)
	if !found {
		return nil, nil
	}
	var cpuStats *statsapi.CPUStats
	var memoryStats *statsapi.MemoryStats
	if info.Spec.HasCpu {
		cpuStats = &statsapi.CPUStats{
			Time: metav1.NewTime(cstat.Timestamp),
		}
		if cstat.CpuInst != nil {
			cpuStats.UsageNanoCores = &cstat.CpuInst.Usage.Total
		}
		if cstat.Cpu != nil {
			cpuStats.UsageCoreNanoSeconds = &cstat.Cpu.Usage.Total
		}
	}
	if info.Spec.HasMemory {
		pageFaults := cstat.Memory.ContainerData.Pgfault
		majorPageFaults := cstat.Memory.ContainerData.Pgmajfault
		memoryStats = &statsapi.MemoryStats{
			Time:            metav1.NewTime(cstat.Timestamp),
			UsageBytes:      &cstat.Memory.Usage,
			WorkingSetBytes: &cstat.Memory.WorkingSet,
			RSSBytes:        &cstat.Memory.RSS,
			PageFaults:      &pageFaults,
			MajorPageFaults: &majorPageFaults,
		}
		// availableBytes = memory limit (if known) - workingset
		if !isMemoryUnlimited(info.Spec.Memory.Limit) {
			availableBytes := info.Spec.Memory.Limit - cstat.Memory.WorkingSet
			memoryStats.AvailableBytes = &availableBytes
		}
	}
	return cpuStats, memoryStats
}

// cadvisorInfoToContainerStats returns the statsapi.ContainerStats converted
// from the container and filesystem info.
func cadvisorInfoToContainerStats(name string, info *cadvisorapiv2.ContainerInfo, rootFs, imageFs *cadvisorapiv2.FsInfo) *statsapi.ContainerStats {
	result := &statsapi.ContainerStats{
		StartTime: metav1.NewTime(info.Spec.CreationTime),
		Name:      name,
	}
	cstat, found := latestContainerStats(info)
	if !found {
		return result
	}

	cpu, memory := cadvisorInfoToCPUandMemoryStats(info)
	result.CPU = cpu
	result.Memory = memory

	if rootFs != nil {
		// The container logs live on the node rootfs device
		result.Logs = buildLogsStats(cstat, rootFs)
	}

	if imageFs != nil {
		// The container rootFs lives on the imageFs devices (which may not be the node root fs)
		result.Rootfs = buildRootfsStats(cstat, imageFs)
	}

	cfs := cstat.Filesystem
	if cfs != nil {
		if cfs.BaseUsageBytes != nil {
			if result.Rootfs != nil {
				rootfsUsage := *cfs.BaseUsageBytes
				result.Rootfs.UsedBytes = &rootfsUsage
			}
			if cfs.TotalUsageBytes != nil && result.Logs != nil {
				logsUsage := *cfs.TotalUsageBytes - *cfs.BaseUsageBytes
				result.Logs.UsedBytes = &logsUsage
			}
		}
		if cfs.InodeUsage != nil && result.Rootfs != nil {
			rootInodes := *cfs.InodeUsage
			result.Rootfs.InodesUsed = &rootInodes
		}
	}

	for _, acc := range cstat.Accelerators {
		result.Accelerators = append(result.Accelerators, statsapi.AcceleratorStats{
			Make:        acc.Make,
			Model:       acc.Model,
			ID:          acc.ID,
			MemoryTotal: acc.MemoryTotal,
			MemoryUsed:  acc.MemoryUsed,
			DutyCycle:   acc.DutyCycle,
		})
	}

	result.UserDefinedMetrics = cadvisorInfoToUserDefinedMetrics(info)

	return result
}

// cadvisorInfoToNetworkStats returns the statsapi.NetworkStats converted from
// the container info from cadvisor.
func cadvisorInfoToNetworkStats(name string, info *cadvisorapiv2.ContainerInfo) *statsapi.NetworkStats {
	if !info.Spec.HasNetwork {
		return nil
	}
	cstat, found := latestContainerStats(info)
	if !found {
		return nil
	}

	iStats := statsapi.NetworkStats{
		Time: metav1.NewTime(cstat.Timestamp),
	}

	for i := range cstat.Network.Interfaces {
		inter := cstat.Network.Interfaces[i]
		iStat := statsapi.InterfaceStats{
			Name:     inter.Name,
			RxBytes:  &inter.RxBytes,
			RxErrors: &inter.RxErrors,
			TxBytes:  &inter.TxBytes,
			TxErrors: &inter.TxErrors,
		}

		if inter.Name == defaultNetworkInterfaceName {
			iStats.InterfaceStats = iStat
		}

		iStats.Interfaces = append(iStats.Interfaces, iStat)
	}

	return &iStats
}

// cadvisorInfoToUserDefinedMetrics returns the statsapi.UserDefinedMetric
// converted from the container info from cadvisor.
func cadvisorInfoToUserDefinedMetrics(info *cadvisorapiv2.ContainerInfo) []statsapi.UserDefinedMetric {
	type specVal struct {
		ref     statsapi.UserDefinedMetricDescriptor
		valType cadvisorapiv1.DataType
		time    time.Time
		value   float64
	}
	udmMap := map[string]*specVal{}
	for _, spec := range info.Spec.CustomMetrics {
		udmMap[spec.Name] = &specVal{
			ref: statsapi.UserDefinedMetricDescriptor{
				Name:  spec.Name,
				Type:  statsapi.UserDefinedMetricType(spec.Type),
				Units: spec.Units,
			},
			valType: spec.Format,
		}
	}
	for _, stat := range info.Stats {
		for name, values := range stat.CustomMetrics {
			specVal, ok := udmMap[name]
			if !ok {
				glog.Warningf("spec for custom metric %q is missing from cAdvisor output. Spec: %+v, Metrics: %+v", name, info.Spec, stat.CustomMetrics)
				continue
			}
			for _, value := range values {
				// Pick the most recent value
				if value.Timestamp.Before(specVal.time) {
					continue
				}
				specVal.time = value.Timestamp
				specVal.value = value.FloatValue
				if specVal.valType == cadvisorapiv1.IntType {
					specVal.value = float64(value.IntValue)
				}
			}
		}
	}
	var udm []statsapi.UserDefinedMetric
	for _, specVal := range udmMap {
		udm = append(udm, statsapi.UserDefinedMetric{
			UserDefinedMetricDescriptor: specVal.ref,
			Time:  metav1.NewTime(specVal.time),
			Value: specVal.value,
		})
	}
	return udm
}

// latestContainerStats returns the latest container stats from cadvisor, or nil if none exist
func latestContainerStats(info *cadvisorapiv2.ContainerInfo) (*cadvisorapiv2.ContainerStats, bool) {
	stats := info.Stats
	if len(stats) < 1 {
		return nil, false
	}
	latest := stats[len(stats)-1]
	if latest == nil {
		return nil, false
	}
	return latest, true
}

func isMemoryUnlimited(v uint64) bool {
	// Size after which we consider memory to be "unlimited". This is not
	// MaxInt64 due to rounding by the kernel.
	// TODO: cadvisor should export this https://github.com/google/cadvisor/blob/master/metrics/prometheus.go#L596
	const maxMemorySize = uint64(1 << 62)

	return v > maxMemorySize
}

// getCgroupInfo returns the information of the container with the specified
// containerName from cadvisor.
func getCgroupInfo(cadvisor cadvisor.Interface, containerName string, updateStats bool) (*cadvisorapiv2.ContainerInfo, error) {
	var maxAge *time.Duration
	if updateStats {
		age := 0 * time.Second
		maxAge = &age
	}
	infoMap, err := cadvisor.ContainerInfoV2(containerName, cadvisorapiv2.RequestOptions{
		IdType:    cadvisorapiv2.TypeName,
		Count:     2, // 2 samples are needed to compute "instantaneous" CPU
		Recursive: false,
		MaxAge:    maxAge,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container info for %q: %v", containerName, err)
	}
	if len(infoMap) != 1 {
		return nil, fmt.Errorf("unexpected number of containers: %v", len(infoMap))
	}
	info := infoMap[containerName]
	return &info, nil
}

// getCgroupStats returns the latest stats of the container having the
// specified containerName from cadvisor.
func getCgroupStats(cadvisor cadvisor.Interface, containerName string, updateStats bool) (*cadvisorapiv2.ContainerStats, error) {
	info, err := getCgroupInfo(cadvisor, containerName, updateStats)
	if err != nil {
		return nil, err
	}
	stats, found := latestContainerStats(info)
	if !found {
		return nil, fmt.Errorf("failed to get latest stats from container info for %q", containerName)
	}
	return stats, nil
}

func buildLogsStats(cstat *cadvisorapiv2.ContainerStats, rootFs *cadvisorapiv2.FsInfo) *statsapi.FsStats {
	fsStats := &statsapi.FsStats{
		Time:           metav1.NewTime(cstat.Timestamp),
		AvailableBytes: &rootFs.Available,
		CapacityBytes:  &rootFs.Capacity,
		InodesFree:     rootFs.InodesFree,
		Inodes:         rootFs.Inodes,
	}

	if rootFs.Inodes != nil && rootFs.InodesFree != nil {
		logsInodesUsed := *rootFs.Inodes - *rootFs.InodesFree
		fsStats.InodesUsed = &logsInodesUsed
	}
	return fsStats
}

func buildRootfsStats(cstat *cadvisorapiv2.ContainerStats, imageFs *cadvisorapiv2.FsInfo) *statsapi.FsStats {
	return &statsapi.FsStats{
		Time:           metav1.NewTime(cstat.Timestamp),
		AvailableBytes: &imageFs.Available,
		CapacityBytes:  &imageFs.Capacity,
		InodesFree:     imageFs.InodesFree,
		Inodes:         imageFs.Inodes,
	}
}
