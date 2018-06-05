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

package eviction

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/clock"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/features"
	statsapi "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
	evictionapi "k8s.io/kubernetes/pkg/kubelet/eviction/api"
	kubetypes "k8s.io/kubernetes/pkg/kubelet/types"
	schedulerutils "k8s.io/kubernetes/pkg/scheduler/util"
)

const (
	unsupportedEvictionSignal = "unsupported eviction signal %v"
	// the reason reported back in status.
	reason = "Evicted"
	// the message associated with the reason.
	message = "The node was low on resource: %v."
	// disk, in bytes.  internal to this module, used to account for local disk usage.
	resourceDisk v1.ResourceName = "disk"
	// inodes, number. internal to this module, used to account for local disk inode consumption.
	resourceInodes v1.ResourceName = "inodes"
	// imagefs, in bytes.  internal to this module, used to account for local image filesystem usage.
	resourceImageFs v1.ResourceName = "imagefs"
	// imagefs inodes, number.  internal to this module, used to account for local image filesystem inodes.
	resourceImageFsInodes v1.ResourceName = "imagefsInodes"
	// nodefs, in bytes.  internal to this module, used to account for local node root filesystem usage.
	resourceNodeFs v1.ResourceName = "nodefs"
	// nodefs inodes, number.  internal to this module, used to account for local node root filesystem inodes.
	resourceNodeFsInodes v1.ResourceName = "nodefsInodes"
	// this prevents constantly updating the memcg notifier if synchronize
	// is run frequently.
	notifierRefreshInterval = 10 * time.Second
)

var (
	// signalToNodeCondition maps a signal to the node condition to report if threshold is met.
	signalToNodeCondition map[evictionapi.Signal]v1.NodeConditionType
	// signalToResource maps a Signal to its associated Resource.
	signalToResource map[evictionapi.Signal]v1.ResourceName
	// resourceClaimToSignal maps a Resource that can be reclaimed to its associated Signal
	resourceClaimToSignal map[v1.ResourceName][]evictionapi.Signal
)

func init() {
	// map eviction signals to node conditions
	signalToNodeCondition = map[evictionapi.Signal]v1.NodeConditionType{}
	signalToNodeCondition[evictionapi.SignalMemoryAvailable] = v1.NodeMemoryPressure
	signalToNodeCondition[evictionapi.SignalAllocatableMemoryAvailable] = v1.NodeMemoryPressure
	signalToNodeCondition[evictionapi.SignalImageFsAvailable] = v1.NodeDiskPressure
	signalToNodeCondition[evictionapi.SignalNodeFsAvailable] = v1.NodeDiskPressure
	signalToNodeCondition[evictionapi.SignalImageFsInodesFree] = v1.NodeDiskPressure
	signalToNodeCondition[evictionapi.SignalNodeFsInodesFree] = v1.NodeDiskPressure
	signalToNodeCondition[evictionapi.SignalPIDAvailable] = v1.NodePIDPressure

	// map signals to resources (and vice-versa)
	signalToResource = map[evictionapi.Signal]v1.ResourceName{}
	signalToResource[evictionapi.SignalMemoryAvailable] = v1.ResourceMemory
	signalToResource[evictionapi.SignalAllocatableMemoryAvailable] = v1.ResourceMemory
	signalToResource[evictionapi.SignalImageFsAvailable] = resourceImageFs
	signalToResource[evictionapi.SignalImageFsInodesFree] = resourceImageFsInodes
	signalToResource[evictionapi.SignalNodeFsAvailable] = resourceNodeFs
	signalToResource[evictionapi.SignalNodeFsInodesFree] = resourceNodeFsInodes

	// maps resource to signals (the following resource could be reclaimed)
	resourceClaimToSignal = map[v1.ResourceName][]evictionapi.Signal{}
	resourceClaimToSignal[resourceNodeFs] = []evictionapi.Signal{evictionapi.SignalNodeFsAvailable}
	resourceClaimToSignal[resourceImageFs] = []evictionapi.Signal{evictionapi.SignalImageFsAvailable}
	resourceClaimToSignal[resourceNodeFsInodes] = []evictionapi.Signal{evictionapi.SignalNodeFsInodesFree}
	resourceClaimToSignal[resourceImageFsInodes] = []evictionapi.Signal{evictionapi.SignalImageFsInodesFree}
}

// validSignal returns true if the signal is supported.
func validSignal(signal evictionapi.Signal) bool {
	_, found := signalToResource[signal]
	return found
}

// ParseThresholdConfig parses the flags for thresholds.
func ParseThresholdConfig(allocatableConfig []string, evictionHard, evictionSoft, evictionSoftGracePeriod, evictionMinimumReclaim map[string]string) ([]evictionapi.Threshold, error) {
	results := []evictionapi.Threshold{}
	hardThresholds, err := parseThresholdStatements(evictionHard)
	if err != nil {
		return nil, err
	}
	results = append(results, hardThresholds...)
	softThresholds, err := parseThresholdStatements(evictionSoft)
	if err != nil {
		return nil, err
	}
	gracePeriods, err := parseGracePeriods(evictionSoftGracePeriod)
	if err != nil {
		return nil, err
	}
	minReclaims, err := parseMinimumReclaims(evictionMinimumReclaim)
	if err != nil {
		return nil, err
	}
	for i := range softThresholds {
		signal := softThresholds[i].Signal
		period, found := gracePeriods[signal]
		if !found {
			return nil, fmt.Errorf("grace period must be specified for the soft eviction threshold %v", signal)
		}
		softThresholds[i].GracePeriod = period
	}
	results = append(results, softThresholds...)
	for i := range results {
		for signal, minReclaim := range minReclaims {
			if results[i].Signal == signal {
				results[i].MinReclaim = &minReclaim
				break
			}
		}
	}
	for _, key := range allocatableConfig {
		if key == kubetypes.NodeAllocatableEnforcementKey {
			results = addAllocatableThresholds(results)
			break
		}
	}
	return results, nil
}

func addAllocatableThresholds(thresholds []evictionapi.Threshold) []evictionapi.Threshold {
	additionalThresholds := []evictionapi.Threshold{}
	for _, threshold := range thresholds {
		if threshold.Signal == evictionapi.SignalMemoryAvailable && isHardEvictionThreshold(threshold) {
			// Copy the SignalMemoryAvailable to SignalAllocatableMemoryAvailable
			additionalThresholds = append(additionalThresholds, evictionapi.Threshold{
				Signal:     evictionapi.SignalAllocatableMemoryAvailable,
				Operator:   threshold.Operator,
				Value:      threshold.Value,
				MinReclaim: threshold.MinReclaim,
			})
		}
	}
	return append(thresholds, additionalThresholds...)
}

// parseThresholdStatements parses the input statements into a list of Threshold objects.
func parseThresholdStatements(statements map[string]string) ([]evictionapi.Threshold, error) {
	if len(statements) == 0 {
		return nil, nil
	}
	results := []evictionapi.Threshold{}
	for signal, val := range statements {
		result, err := parseThresholdStatement(evictionapi.Signal(signal), val)
		if err != nil {
			return nil, err
		}
		if result != nil {
			results = append(results, *result)
		}
	}
	return results, nil
}

// parseThresholdStatement parses a threshold statement and returns a threshold,
// or nil if the threshold should be ignored.
func parseThresholdStatement(signal evictionapi.Signal, val string) (*evictionapi.Threshold, error) {
	if !validSignal(signal) {
		return nil, fmt.Errorf(unsupportedEvictionSignal, signal)
	}
	operator := evictionapi.OpForSignal[signal]
	if strings.HasSuffix(val, "%") {
		// ignore 0% and 100%
		if val == "0%" || val == "100%" {
			return nil, nil
		}
		percentage, err := parsePercentage(val)
		if err != nil {
			return nil, err
		}
		if percentage < 0 {
			return nil, fmt.Errorf("eviction percentage threshold %v must be >= 0%%: %s", signal, val)
		}
		if percentage > 100 {
			return nil, fmt.Errorf("eviction percentage threshold %v must be <= 100%%: %s", signal, val)
		}
		return &evictionapi.Threshold{
			Signal:   signal,
			Operator: operator,
			Value: evictionapi.ThresholdValue{
				Percentage: percentage,
			},
		}, nil
	}
	quantity, err := resource.ParseQuantity(val)
	if err != nil {
		return nil, err
	}
	if quantity.Sign() < 0 || quantity.IsZero() {
		return nil, fmt.Errorf("eviction threshold %v must be positive: %s", signal, &quantity)
	}
	return &evictionapi.Threshold{
		Signal:   signal,
		Operator: operator,
		Value: evictionapi.ThresholdValue{
			Quantity: &quantity,
		},
	}, nil
}

// parsePercentage parses a string representing a percentage value
func parsePercentage(input string) (float32, error) {
	value, err := strconv.ParseFloat(strings.TrimRight(input, "%"), 32)
	if err != nil {
		return 0, err
	}
	return float32(value) / 100, nil
}

// parseGracePeriods parses the grace period statements
func parseGracePeriods(statements map[string]string) (map[evictionapi.Signal]time.Duration, error) {
	if len(statements) == 0 {
		return nil, nil
	}
	results := map[evictionapi.Signal]time.Duration{}
	for signal, val := range statements {
		signal := evictionapi.Signal(signal)
		if !validSignal(signal) {
			return nil, fmt.Errorf(unsupportedEvictionSignal, signal)
		}
		gracePeriod, err := time.ParseDuration(val)
		if err != nil {
			return nil, err
		}
		if gracePeriod < 0 {
			return nil, fmt.Errorf("invalid eviction grace period specified: %v, must be a positive value", val)
		}
		results[signal] = gracePeriod
	}
	return results, nil
}

// parseMinimumReclaims parses the minimum reclaim statements
func parseMinimumReclaims(statements map[string]string) (map[evictionapi.Signal]evictionapi.ThresholdValue, error) {
	if len(statements) == 0 {
		return nil, nil
	}
	results := map[evictionapi.Signal]evictionapi.ThresholdValue{}
	for signal, val := range statements {
		signal := evictionapi.Signal(signal)
		if !validSignal(signal) {
			return nil, fmt.Errorf(unsupportedEvictionSignal, signal)
		}
		if strings.HasSuffix(val, "%") {
			percentage, err := parsePercentage(val)
			if err != nil {
				return nil, err
			}
			if percentage <= 0 {
				return nil, fmt.Errorf("eviction percentage minimum reclaim %v must be positive: %s", signal, val)
			}
			results[signal] = evictionapi.ThresholdValue{
				Percentage: percentage,
			}
			continue
		}
		quantity, err := resource.ParseQuantity(val)
		if err != nil {
			return nil, err
		}
		if quantity.Sign() < 0 {
			return nil, fmt.Errorf("negative eviction minimum reclaim specified for %v", signal)
		}
		results[signal] = evictionapi.ThresholdValue{
			Quantity: &quantity,
		}
	}
	return results, nil
}

// diskUsage converts used bytes into a resource quantity.
func diskUsage(fsStats *statsapi.FsStats) *resource.Quantity {
	if fsStats == nil || fsStats.UsedBytes == nil {
		return &resource.Quantity{Format: resource.BinarySI}
	}
	usage := int64(*fsStats.UsedBytes)
	return resource.NewQuantity(usage, resource.BinarySI)
}

// inodeUsage converts inodes consumed into a resource quantity.
func inodeUsage(fsStats *statsapi.FsStats) *resource.Quantity {
	if fsStats == nil || fsStats.InodesUsed == nil {
		return &resource.Quantity{Format: resource.BinarySI}
	}
	usage := int64(*fsStats.InodesUsed)
	return resource.NewQuantity(usage, resource.BinarySI)
}

// memoryUsage converts working set into a resource quantity.
func memoryUsage(memStats *statsapi.MemoryStats) *resource.Quantity {
	if memStats == nil || memStats.WorkingSetBytes == nil {
		return &resource.Quantity{Format: resource.BinarySI}
	}
	usage := int64(*memStats.WorkingSetBytes)
	return resource.NewQuantity(usage, resource.BinarySI)
}

// localVolumeNames returns the set of volumes for the pod that are local
// TODO: sumamry API should report what volumes consume local storage rather than hard-code here.
func localVolumeNames(pod *v1.Pod) []string {
	result := []string{}
	for _, volume := range pod.Spec.Volumes {
		if volume.HostPath != nil ||
			(volume.EmptyDir != nil && volume.EmptyDir.Medium != v1.StorageMediumMemory) ||
			volume.ConfigMap != nil ||
			volume.GitRepo != nil {
			result = append(result, volume.Name)
		}
	}
	return result
}

// containerUsage aggregates container disk usage and inode consumption for the specified stats to measure.
func containerUsage(podStats statsapi.PodStats, statsToMeasure []fsStatsType) v1.ResourceList {
	disk := resource.Quantity{Format: resource.BinarySI}
	inodes := resource.Quantity{Format: resource.BinarySI}
	for _, container := range podStats.Containers {
		if hasFsStatsType(statsToMeasure, fsStatsRoot) {
			disk.Add(*diskUsage(container.Rootfs))
			inodes.Add(*inodeUsage(container.Rootfs))
		}
		if hasFsStatsType(statsToMeasure, fsStatsLogs) {
			disk.Add(*diskUsage(container.Logs))
			inodes.Add(*inodeUsage(container.Logs))
		}
	}
	return v1.ResourceList{
		resourceDisk:   disk,
		resourceInodes: inodes,
	}
}

// podLocalVolumeUsage aggregates pod local volumes disk usage and inode consumption for the specified stats to measure.
func podLocalVolumeUsage(volumeNames []string, podStats statsapi.PodStats) v1.ResourceList {
	disk := resource.Quantity{Format: resource.BinarySI}
	inodes := resource.Quantity{Format: resource.BinarySI}
	for _, volumeName := range volumeNames {
		for _, volumeStats := range podStats.VolumeStats {
			if volumeStats.Name == volumeName {
				disk.Add(*diskUsage(&volumeStats.FsStats))
				inodes.Add(*inodeUsage(&volumeStats.FsStats))
				break
			}
		}
	}
	return v1.ResourceList{
		resourceDisk:   disk,
		resourceInodes: inodes,
	}
}

// podDiskUsage aggregates pod disk usage and inode consumption for the specified stats to measure.
func podDiskUsage(podStats statsapi.PodStats, pod *v1.Pod, statsToMeasure []fsStatsType) (v1.ResourceList, error) {
	disk := resource.Quantity{Format: resource.BinarySI}
	inodes := resource.Quantity{Format: resource.BinarySI}

	containerUsageList := containerUsage(podStats, statsToMeasure)
	disk.Add(containerUsageList[resourceDisk])
	inodes.Add(containerUsageList[resourceInodes])

	if hasFsStatsType(statsToMeasure, fsStatsLocalVolumeSource) {
		volumeNames := localVolumeNames(pod)
		podLocalVolumeUsageList := podLocalVolumeUsage(volumeNames, podStats)
		disk.Add(podLocalVolumeUsageList[resourceDisk])
		inodes.Add(podLocalVolumeUsageList[resourceInodes])
	}
	return v1.ResourceList{
		resourceDisk:   disk,
		resourceInodes: inodes,
	}, nil
}

// podMemoryUsage aggregates pod memory usage.
func podMemoryUsage(podStats statsapi.PodStats) (v1.ResourceList, error) {
	memory := resource.Quantity{Format: resource.BinarySI}
	for _, container := range podStats.Containers {
		// memory usage (if known)
		memory.Add(*memoryUsage(container.Memory))
	}
	return v1.ResourceList{v1.ResourceMemory: memory}, nil
}

// localEphemeralVolumeNames returns the set of ephemeral volumes for the pod that are local
func localEphemeralVolumeNames(pod *v1.Pod) []string {
	result := []string{}
	for _, volume := range pod.Spec.Volumes {
		if volume.GitRepo != nil ||
			(volume.EmptyDir != nil && volume.EmptyDir.Medium != v1.StorageMediumMemory) ||
			volume.ConfigMap != nil || volume.DownwardAPI != nil {
			result = append(result, volume.Name)
		}
	}
	return result
}

// podLocalEphemeralStorageUsage aggregates pod local ephemeral storage usage and inode consumption for the specified stats to measure.
func podLocalEphemeralStorageUsage(podStats statsapi.PodStats, pod *v1.Pod, statsToMeasure []fsStatsType) (v1.ResourceList, error) {
	disk := resource.Quantity{Format: resource.BinarySI}
	inodes := resource.Quantity{Format: resource.BinarySI}

	containerUsageList := containerUsage(podStats, statsToMeasure)
	disk.Add(containerUsageList[resourceDisk])
	inodes.Add(containerUsageList[resourceInodes])

	if hasFsStatsType(statsToMeasure, fsStatsLocalVolumeSource) {
		volumeNames := localEphemeralVolumeNames(pod)
		podLocalVolumeUsageList := podLocalVolumeUsage(volumeNames, podStats)
		disk.Add(podLocalVolumeUsageList[resourceDisk])
		inodes.Add(podLocalVolumeUsageList[resourceInodes])
	}
	return v1.ResourceList{
		resourceDisk:   disk,
		resourceInodes: inodes,
	}, nil
}

// formatThreshold formats a threshold for logging.
func formatThreshold(threshold evictionapi.Threshold) string {
	return fmt.Sprintf("threshold(signal=%v, operator=%v, value=%v, gracePeriod=%v)", threshold.Signal, threshold.Operator, evictionapi.ThresholdValue(threshold.Value), threshold.GracePeriod)
}

// formatevictionapi.ThresholdValue formats a thresholdValue for logging.
func formatThresholdValue(value evictionapi.ThresholdValue) string {
	if value.Quantity != nil {
		return value.Quantity.String()
	}
	return fmt.Sprintf("%f%%", value.Percentage*float32(100))
}

// cachedStatsFunc returns a statsFunc based on the provided pod stats.
func cachedStatsFunc(podStats []statsapi.PodStats) statsFunc {
	uid2PodStats := map[string]statsapi.PodStats{}
	for i := range podStats {
		uid2PodStats[podStats[i].PodRef.UID] = podStats[i]
	}
	return func(pod *v1.Pod) (statsapi.PodStats, bool) {
		stats, found := uid2PodStats[string(pod.UID)]
		return stats, found
	}
}

// Cmp compares p1 and p2 and returns:
//
//   -1 if p1 <  p2
//    0 if p1 == p2
//   +1 if p1 >  p2
//
type cmpFunc func(p1, p2 *v1.Pod) int

// multiSorter implements the Sort interface, sorting changes within.
type multiSorter struct {
	pods []*v1.Pod
	cmp  []cmpFunc
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *multiSorter) Sort(pods []*v1.Pod) {
	ms.pods = pods
	sort.Sort(ms)
}

// OrderedBy returns a Sorter that sorts using the cmp functions, in order.
// Call its Sort method to sort the data.
func orderedBy(cmp ...cmpFunc) *multiSorter {
	return &multiSorter{
		cmp: cmp,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.pods)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.pods[i], ms.pods[j] = ms.pods[j], ms.pods[i]
}

// Less is part of sort.Interface.
func (ms *multiSorter) Less(i, j int) bool {
	p1, p2 := ms.pods[i], ms.pods[j]
	var k int
	for k = 0; k < len(ms.cmp)-1; k++ {
		cmpResult := ms.cmp[k](p1, p2)
		// p1 is less than p2
		if cmpResult < 0 {
			return true
		}
		// p1 is greater than p2
		if cmpResult > 0 {
			return false
		}
		// we don't know yet
	}
	// the last cmp func is the final decider
	return ms.cmp[k](p1, p2) < 0
}

// priority compares pods by Priority, if priority is enabled.
func priority(p1, p2 *v1.Pod) int {
	if !utilfeature.DefaultFeatureGate.Enabled(features.PodPriority) {
		// If priority is not enabled, all pods are equal.
		return 0
	}
	priority1 := schedulerutils.GetPodPriority(p1)
	priority2 := schedulerutils.GetPodPriority(p2)
	if priority1 == priority2 {
		return 0
	}
	if priority1 > priority2 {
		return 1
	}
	return -1
}

// exceedMemoryRequests compares whether or not pods' memory usage exceeds their requests
func exceedMemoryRequests(stats statsFunc) cmpFunc {
	return func(p1, p2 *v1.Pod) int {
		p1Stats, p1Found := stats(p1)
		p2Stats, p2Found := stats(p2)
		if !p1Found || !p2Found {
			// prioritize evicting the pod for which no stats were found
			return cmpBool(!p1Found, !p2Found)
		}

		p1Usage, p1Err := podMemoryUsage(p1Stats)
		p2Usage, p2Err := podMemoryUsage(p2Stats)
		if p1Err != nil || p2Err != nil {
			// prioritize evicting the pod which had an error getting stats
			return cmpBool(p1Err != nil, p2Err != nil)
		}

		p1Memory := p1Usage[v1.ResourceMemory]
		p2Memory := p2Usage[v1.ResourceMemory]
		p1ExceedsRequests := p1Memory.Cmp(podRequest(p1, v1.ResourceMemory)) == 1
		p2ExceedsRequests := p2Memory.Cmp(podRequest(p2, v1.ResourceMemory)) == 1
		// prioritize evicting the pod which exceeds its requests
		return cmpBool(p1ExceedsRequests, p2ExceedsRequests)
	}
}

// memory compares pods by largest consumer of memory relative to request.
func memory(stats statsFunc) cmpFunc {
	return func(p1, p2 *v1.Pod) int {
		p1Stats, p1Found := stats(p1)
		p2Stats, p2Found := stats(p2)
		if !p1Found || !p2Found {
			// prioritize evicting the pod for which no stats were found
			return cmpBool(!p1Found, !p2Found)
		}

		p1Usage, p1Err := podMemoryUsage(p1Stats)
		p2Usage, p2Err := podMemoryUsage(p2Stats)
		if p1Err != nil || p2Err != nil {
			// prioritize evicting the pod which had an error getting stats
			return cmpBool(p1Err != nil, p2Err != nil)
		}

		// adjust p1, p2 usage relative to the request (if any)
		p1Memory := p1Usage[v1.ResourceMemory]
		p1Request := podRequest(p1, v1.ResourceMemory)
		p1Memory.Sub(p1Request)

		p2Memory := p2Usage[v1.ResourceMemory]
		p2Request := podRequest(p2, v1.ResourceMemory)
		p2Memory.Sub(p2Request)

		// prioritize evicting the pod which has the larger consumption of memory
		return p2Memory.Cmp(p1Memory)
	}
}

// podRequest returns the total resource request of a pod which is the
// max(max of init container requests, sum of container requests)
func podRequest(pod *v1.Pod, resourceName v1.ResourceName) resource.Quantity {
	containerValue := resource.Quantity{Format: resource.BinarySI}
	if resourceName == resourceDisk && !utilfeature.DefaultFeatureGate.Enabled(features.LocalStorageCapacityIsolation) {
		// if the local storage capacity isolation feature gate is disabled, pods request 0 disk
		return containerValue
	}
	for i := range pod.Spec.Containers {
		switch resourceName {
		case v1.ResourceMemory:
			containerValue.Add(*pod.Spec.Containers[i].Resources.Requests.Memory())
		case resourceDisk:
			containerValue.Add(*pod.Spec.Containers[i].Resources.Requests.StorageEphemeral())
		}
	}
	initValue := resource.Quantity{Format: resource.BinarySI}
	for i := range pod.Spec.InitContainers {
		switch resourceName {
		case v1.ResourceMemory:
			if initValue.Cmp(*pod.Spec.InitContainers[i].Resources.Requests.Memory()) < 0 {
				initValue = *pod.Spec.InitContainers[i].Resources.Requests.Memory()
			}
		case resourceDisk:
			if initValue.Cmp(*pod.Spec.InitContainers[i].Resources.Requests.StorageEphemeral()) < 0 {
				initValue = *pod.Spec.InitContainers[i].Resources.Requests.StorageEphemeral()
			}
		}
	}
	if containerValue.Cmp(initValue) > 0 {
		return containerValue
	}
	return initValue
}

// exceedDiskRequests compares whether or not pods' disk usage exceeds their requests
func exceedDiskRequests(stats statsFunc, fsStatsToMeasure []fsStatsType, diskResource v1.ResourceName) cmpFunc {
	return func(p1, p2 *v1.Pod) int {
		p1Stats, p1Found := stats(p1)
		p2Stats, p2Found := stats(p2)
		if !p1Found || !p2Found {
			// prioritize evicting the pod for which no stats were found
			return cmpBool(!p1Found, !p2Found)
		}

		p1Usage, p1Err := podDiskUsage(p1Stats, p1, fsStatsToMeasure)
		p2Usage, p2Err := podDiskUsage(p2Stats, p2, fsStatsToMeasure)
		if p1Err != nil || p2Err != nil {
			// prioritize evicting the pod which had an error getting stats
			return cmpBool(p1Err != nil, p2Err != nil)
		}

		p1Disk := p1Usage[diskResource]
		p2Disk := p2Usage[diskResource]
		p1ExceedsRequests := p1Disk.Cmp(podRequest(p1, diskResource)) == 1
		p2ExceedsRequests := p2Disk.Cmp(podRequest(p2, diskResource)) == 1
		// prioritize evicting the pod which exceeds its requests
		return cmpBool(p1ExceedsRequests, p2ExceedsRequests)
	}
}

// disk compares pods by largest consumer of disk relative to request for the specified disk resource.
func disk(stats statsFunc, fsStatsToMeasure []fsStatsType, diskResource v1.ResourceName) cmpFunc {
	return func(p1, p2 *v1.Pod) int {
		p1Stats, p1Found := stats(p1)
		p2Stats, p2Found := stats(p2)
		if !p1Found || !p2Found {
			// prioritize evicting the pod for which no stats were found
			return cmpBool(!p1Found, !p2Found)
		}
		p1Usage, p1Err := podDiskUsage(p1Stats, p1, fsStatsToMeasure)
		p2Usage, p2Err := podDiskUsage(p2Stats, p2, fsStatsToMeasure)
		if p1Err != nil || p2Err != nil {
			// prioritize evicting the pod which had an error getting stats
			return cmpBool(p1Err != nil, p2Err != nil)
		}

		// adjust p1, p2 usage relative to the request (if any)
		p1Disk := p1Usage[diskResource]
		p2Disk := p2Usage[diskResource]
		p1Request := podRequest(p1, resourceDisk)
		p1Disk.Sub(p1Request)
		p2Request := podRequest(p2, resourceDisk)
		p2Disk.Sub(p2Request)
		// prioritize evicting the pod which has the larger consumption of disk
		return p2Disk.Cmp(p1Disk)
	}
}

// cmpBool compares booleans, placing true before false
func cmpBool(a, b bool) int {
	if a == b {
		return 0
	}
	if !b {
		return -1
	}
	return 1
}

// rankMemoryPressure orders the input pods for eviction in response to memory pressure.
// It ranks by whether or not the pod's usage exceeds its requests, then by priority, and
// finally by memory usage above requests.
func rankMemoryPressure(pods []*v1.Pod, stats statsFunc) {
	orderedBy(exceedMemoryRequests(stats), priority, memory(stats)).Sort(pods)
}

// rankDiskPressureFunc returns a rankFunc that measures the specified fs stats.
func rankDiskPressureFunc(fsStatsToMeasure []fsStatsType, diskResource v1.ResourceName) rankFunc {
	return func(pods []*v1.Pod, stats statsFunc) {
		orderedBy(exceedDiskRequests(stats, fsStatsToMeasure, diskResource), priority, disk(stats, fsStatsToMeasure, diskResource)).Sort(pods)
	}
}

// byEvictionPriority implements sort.Interface for []v1.ResourceName.
type byEvictionPriority []v1.ResourceName

func (a byEvictionPriority) Len() int      { return len(a) }
func (a byEvictionPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less ranks memory before all other resources.
func (a byEvictionPriority) Less(i, j int) bool {
	return a[i] == v1.ResourceMemory
}

// makeSignalObservations derives observations using the specified summary provider.
func makeSignalObservations(summary *statsapi.Summary) (signalObservations, statsFunc) {
	// build the function to work against for pod stats
	statsFunc := cachedStatsFunc(summary.Pods)
	// build an evaluation context for current eviction signals
	result := signalObservations{}

	if memory := summary.Node.Memory; memory != nil && memory.AvailableBytes != nil && memory.WorkingSetBytes != nil {
		result[evictionapi.SignalMemoryAvailable] = signalObservation{
			available: resource.NewQuantity(int64(*memory.AvailableBytes), resource.BinarySI),
			capacity:  resource.NewQuantity(int64(*memory.AvailableBytes+*memory.WorkingSetBytes), resource.BinarySI),
			time:      memory.Time,
		}
	}
	if allocatableContainer, err := getSysContainer(summary.Node.SystemContainers, statsapi.SystemContainerPods); err != nil {
		glog.Errorf("eviction manager: failed to construct signal: %q error: %v", evictionapi.SignalAllocatableMemoryAvailable, err)
	} else {
		if memory := allocatableContainer.Memory; memory != nil && memory.AvailableBytes != nil && memory.WorkingSetBytes != nil {
			result[evictionapi.SignalAllocatableMemoryAvailable] = signalObservation{
				available: resource.NewQuantity(int64(*memory.AvailableBytes), resource.BinarySI),
				capacity:  resource.NewQuantity(int64(*memory.AvailableBytes+*memory.WorkingSetBytes), resource.BinarySI),
				time:      memory.Time,
			}
		}
	}
	if nodeFs := summary.Node.Fs; nodeFs != nil {
		if nodeFs.AvailableBytes != nil && nodeFs.CapacityBytes != nil {
			result[evictionapi.SignalNodeFsAvailable] = signalObservation{
				available: resource.NewQuantity(int64(*nodeFs.AvailableBytes), resource.BinarySI),
				capacity:  resource.NewQuantity(int64(*nodeFs.CapacityBytes), resource.BinarySI),
				time:      nodeFs.Time,
			}
		}
		if nodeFs.InodesFree != nil && nodeFs.Inodes != nil {
			result[evictionapi.SignalNodeFsInodesFree] = signalObservation{
				available: resource.NewQuantity(int64(*nodeFs.InodesFree), resource.BinarySI),
				capacity:  resource.NewQuantity(int64(*nodeFs.Inodes), resource.BinarySI),
				time:      nodeFs.Time,
			}
		}
	}
	if summary.Node.Runtime != nil {
		if imageFs := summary.Node.Runtime.ImageFs; imageFs != nil {
			if imageFs.AvailableBytes != nil && imageFs.CapacityBytes != nil {
				result[evictionapi.SignalImageFsAvailable] = signalObservation{
					available: resource.NewQuantity(int64(*imageFs.AvailableBytes), resource.BinarySI),
					capacity:  resource.NewQuantity(int64(*imageFs.CapacityBytes), resource.BinarySI),
					time:      imageFs.Time,
				}
				if imageFs.InodesFree != nil && imageFs.Inodes != nil {
					result[evictionapi.SignalImageFsInodesFree] = signalObservation{
						available: resource.NewQuantity(int64(*imageFs.InodesFree), resource.BinarySI),
						capacity:  resource.NewQuantity(int64(*imageFs.Inodes), resource.BinarySI),
						time:      imageFs.Time,
					}
				}
			}
		}
	}
	if rlimit := summary.Node.Rlimit; rlimit != nil {
		if rlimit.NumOfRunningProcesses != nil && rlimit.MaxPID != nil {
			available := int64(*rlimit.MaxPID) - int64(*rlimit.NumOfRunningProcesses)
			result[evictionapi.SignalPIDAvailable] = signalObservation{
				available: resource.NewQuantity(available, resource.BinarySI),
				capacity:  resource.NewQuantity(int64(*rlimit.MaxPID), resource.BinarySI),
				time:      rlimit.Time,
			}
		}
	}
	return result, statsFunc
}

func getSysContainer(sysContainers []statsapi.ContainerStats, name string) (*statsapi.ContainerStats, error) {
	for _, cont := range sysContainers {
		if cont.Name == name {
			return &cont, nil
		}
	}
	return nil, fmt.Errorf("system container %q not found in metrics", name)
}

// thresholdsMet returns the set of thresholds that were met independent of grace period
func thresholdsMet(thresholds []evictionapi.Threshold, observations signalObservations, enforceMinReclaim bool) []evictionapi.Threshold {
	results := []evictionapi.Threshold{}
	for i := range thresholds {
		threshold := thresholds[i]
		observed, found := observations[threshold.Signal]
		if !found {
			glog.Warningf("eviction manager: no observation found for eviction signal %v", threshold.Signal)
			continue
		}
		// determine if we have met the specified threshold
		thresholdMet := false
		quantity := evictionapi.GetThresholdQuantity(threshold.Value, observed.capacity)
		// if enforceMinReclaim is specified, we compare relative to value - minreclaim
		if enforceMinReclaim && threshold.MinReclaim != nil {
			quantity.Add(*evictionapi.GetThresholdQuantity(*threshold.MinReclaim, observed.capacity))
		}
		thresholdResult := quantity.Cmp(*observed.available)
		switch threshold.Operator {
		case evictionapi.OpLessThan:
			thresholdMet = thresholdResult > 0
		}
		if thresholdMet {
			results = append(results, threshold)
		}
	}
	return results
}

func debugLogObservations(logPrefix string, observations signalObservations) {
	if !glog.V(3) {
		return
	}
	for k, v := range observations {
		if !v.time.IsZero() {
			glog.Infof("eviction manager: %v: signal=%v, available: %v, capacity: %v, time: %v", logPrefix, k, v.available, v.capacity, v.time)
		} else {
			glog.Infof("eviction manager: %v: signal=%v, available: %v, capacity: %v", logPrefix, k, v.available, v.capacity)
		}
	}
}

func debugLogThresholdsWithObservation(logPrefix string, thresholds []evictionapi.Threshold, observations signalObservations) {
	if !glog.V(3) {
		return
	}
	for i := range thresholds {
		threshold := thresholds[i]
		observed, found := observations[threshold.Signal]
		if found {
			quantity := evictionapi.GetThresholdQuantity(threshold.Value, observed.capacity)
			glog.Infof("eviction manager: %v: threshold [signal=%v, quantity=%v] observed %v", logPrefix, threshold.Signal, quantity, observed.available)
		} else {
			glog.Infof("eviction manager: %v: threshold [signal=%v] had no observation", logPrefix, threshold.Signal)
		}
	}
}

func thresholdsUpdatedStats(thresholds []evictionapi.Threshold, observations, lastObservations signalObservations) []evictionapi.Threshold {
	results := []evictionapi.Threshold{}
	for i := range thresholds {
		threshold := thresholds[i]
		observed, found := observations[threshold.Signal]
		if !found {
			glog.Warningf("eviction manager: no observation found for eviction signal %v", threshold.Signal)
			continue
		}
		last, found := lastObservations[threshold.Signal]
		if !found || observed.time.IsZero() || observed.time.After(last.time.Time) {
			results = append(results, threshold)
		}
	}
	return results
}

// thresholdsFirstObservedAt merges the input set of thresholds with the previous observation to determine when active set of thresholds were initially met.
func thresholdsFirstObservedAt(thresholds []evictionapi.Threshold, lastObservedAt thresholdsObservedAt, now time.Time) thresholdsObservedAt {
	results := thresholdsObservedAt{}
	for i := range thresholds {
		observedAt, found := lastObservedAt[thresholds[i]]
		if !found {
			observedAt = now
		}
		results[thresholds[i]] = observedAt
	}
	return results
}

// thresholdsMetGracePeriod returns the set of thresholds that have satisfied associated grace period
func thresholdsMetGracePeriod(observedAt thresholdsObservedAt, now time.Time) []evictionapi.Threshold {
	results := []evictionapi.Threshold{}
	for threshold, at := range observedAt {
		duration := now.Sub(at)
		if duration < threshold.GracePeriod {
			glog.V(2).Infof("eviction manager: eviction criteria not yet met for %v, duration: %v", formatThreshold(threshold), duration)
			continue
		}
		results = append(results, threshold)
	}
	return results
}

// nodeConditions returns the set of node conditions associated with a threshold
func nodeConditions(thresholds []evictionapi.Threshold) []v1.NodeConditionType {
	results := []v1.NodeConditionType{}
	for _, threshold := range thresholds {
		if nodeCondition, found := signalToNodeCondition[threshold.Signal]; found {
			if !hasNodeCondition(results, nodeCondition) {
				results = append(results, nodeCondition)
			}
		}
	}
	return results
}

// nodeConditionsLastObservedAt merges the input with the previous observation to determine when a condition was most recently met.
func nodeConditionsLastObservedAt(nodeConditions []v1.NodeConditionType, lastObservedAt nodeConditionsObservedAt, now time.Time) nodeConditionsObservedAt {
	results := nodeConditionsObservedAt{}
	// the input conditions were observed "now"
	for i := range nodeConditions {
		results[nodeConditions[i]] = now
	}
	// the conditions that were not observed now are merged in with their old time
	for key, value := range lastObservedAt {
		_, found := results[key]
		if !found {
			results[key] = value
		}
	}
	return results
}

// nodeConditionsObservedSince returns the set of conditions that have been observed within the specified period
func nodeConditionsObservedSince(observedAt nodeConditionsObservedAt, period time.Duration, now time.Time) []v1.NodeConditionType {
	results := []v1.NodeConditionType{}
	for nodeCondition, at := range observedAt {
		duration := now.Sub(at)
		if duration < period {
			results = append(results, nodeCondition)
		}
	}
	return results
}

// hasFsStatsType returns true if the fsStat is in the input list
func hasFsStatsType(inputs []fsStatsType, item fsStatsType) bool {
	for _, input := range inputs {
		if input == item {
			return true
		}
	}
	return false
}

// hasNodeCondition returns true if the node condition is in the input list
func hasNodeCondition(inputs []v1.NodeConditionType, item v1.NodeConditionType) bool {
	for _, input := range inputs {
		if input == item {
			return true
		}
	}
	return false
}

// mergeThresholds will merge both threshold lists eliminating duplicates.
func mergeThresholds(inputsA []evictionapi.Threshold, inputsB []evictionapi.Threshold) []evictionapi.Threshold {
	results := inputsA
	for _, threshold := range inputsB {
		if !hasThreshold(results, threshold) {
			results = append(results, threshold)
		}
	}
	return results
}

// hasThreshold returns true if the threshold is in the input list
func hasThreshold(inputs []evictionapi.Threshold, item evictionapi.Threshold) bool {
	for _, input := range inputs {
		if input.GracePeriod == item.GracePeriod && input.Operator == item.Operator && input.Signal == item.Signal && compareThresholdValue(input.Value, item.Value) {
			return true
		}
	}
	return false
}

// compareThresholdValue returns true if the two thresholdValue objects are logically the same
func compareThresholdValue(a evictionapi.ThresholdValue, b evictionapi.ThresholdValue) bool {
	if a.Quantity != nil {
		if b.Quantity == nil {
			return false
		}
		return a.Quantity.Cmp(*b.Quantity) == 0
	}
	if b.Quantity != nil {
		return false
	}
	return a.Percentage == b.Percentage
}

// getStarvedResources returns the set of resources that are starved based on thresholds met.
func getStarvedResources(thresholds []evictionapi.Threshold) []v1.ResourceName {
	results := []v1.ResourceName{}
	for _, threshold := range thresholds {
		if starvedResource, found := signalToResource[threshold.Signal]; found {
			results = append(results, starvedResource)
		}
	}
	return results
}

// isSoftEviction returns true if the thresholds met for the starved resource are only soft thresholds
func isSoftEvictionThresholds(thresholds []evictionapi.Threshold, starvedResource v1.ResourceName) bool {
	for _, threshold := range thresholds {
		if resourceToCheck := signalToResource[threshold.Signal]; resourceToCheck != starvedResource {
			continue
		}
		if isHardEvictionThreshold(threshold) {
			return false
		}
	}
	return true
}

// isHardEvictionThreshold returns true if eviction should immediately occur
func isHardEvictionThreshold(threshold evictionapi.Threshold) bool {
	return threshold.GracePeriod == time.Duration(0)
}

// buildResourceToRankFunc returns ranking functions associated with resources
func buildResourceToRankFunc(withImageFs bool) map[v1.ResourceName]rankFunc {
	resourceToRankFunc := map[v1.ResourceName]rankFunc{
		v1.ResourceMemory: rankMemoryPressure,
	}
	// usage of an imagefs is optional
	if withImageFs {
		// with an imagefs, nodefs pod rank func for eviction only includes logs and local volumes
		resourceToRankFunc[resourceNodeFs] = rankDiskPressureFunc([]fsStatsType{fsStatsLogs, fsStatsLocalVolumeSource}, resourceDisk)
		resourceToRankFunc[resourceNodeFsInodes] = rankDiskPressureFunc([]fsStatsType{fsStatsLogs, fsStatsLocalVolumeSource}, resourceInodes)
		// with an imagefs, imagefs pod rank func for eviction only includes rootfs
		resourceToRankFunc[resourceImageFs] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot}, resourceDisk)
		resourceToRankFunc[resourceImageFsInodes] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot}, resourceInodes)
	} else {
		// without an imagefs, nodefs pod rank func for eviction looks at all fs stats.
		// since imagefs and nodefs share a common device, they share common ranking functions.
		resourceToRankFunc[resourceNodeFs] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot, fsStatsLogs, fsStatsLocalVolumeSource}, resourceDisk)
		resourceToRankFunc[resourceNodeFsInodes] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot, fsStatsLogs, fsStatsLocalVolumeSource}, resourceInodes)
		resourceToRankFunc[resourceImageFs] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot, fsStatsLogs, fsStatsLocalVolumeSource}, resourceDisk)
		resourceToRankFunc[resourceImageFsInodes] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot, fsStatsLogs, fsStatsLocalVolumeSource}, resourceInodes)
	}
	return resourceToRankFunc
}

// PodIsEvicted returns true if the reported pod status is due to an eviction.
func PodIsEvicted(podStatus v1.PodStatus) bool {
	return podStatus.Phase == v1.PodFailed && podStatus.Reason == reason
}

// buildResourceToNodeReclaimFuncs returns reclaim functions associated with resources.
func buildResourceToNodeReclaimFuncs(imageGC ImageGC, containerGC ContainerGC, withImageFs bool) map[v1.ResourceName]nodeReclaimFuncs {
	resourceToReclaimFunc := map[v1.ResourceName]nodeReclaimFuncs{}
	// usage of an imagefs is optional
	if withImageFs {
		// with an imagefs, nodefs pressure should just delete logs
		resourceToReclaimFunc[resourceNodeFs] = nodeReclaimFuncs{}
		resourceToReclaimFunc[resourceNodeFsInodes] = nodeReclaimFuncs{}
		// with an imagefs, imagefs pressure should delete unused images
		resourceToReclaimFunc[resourceImageFs] = nodeReclaimFuncs{containerGC.DeleteAllUnusedContainers, imageGC.DeleteUnusedImages}
		resourceToReclaimFunc[resourceImageFsInodes] = nodeReclaimFuncs{containerGC.DeleteAllUnusedContainers, imageGC.DeleteUnusedImages}
	} else {
		// without an imagefs, nodefs pressure should delete logs, and unused images
		// since imagefs and nodefs share a common device, they share common reclaim functions
		resourceToReclaimFunc[resourceNodeFs] = nodeReclaimFuncs{containerGC.DeleteAllUnusedContainers, imageGC.DeleteUnusedImages}
		resourceToReclaimFunc[resourceNodeFsInodes] = nodeReclaimFuncs{containerGC.DeleteAllUnusedContainers, imageGC.DeleteUnusedImages}
		resourceToReclaimFunc[resourceImageFs] = nodeReclaimFuncs{containerGC.DeleteAllUnusedContainers, imageGC.DeleteUnusedImages}
		resourceToReclaimFunc[resourceImageFsInodes] = nodeReclaimFuncs{containerGC.DeleteAllUnusedContainers, imageGC.DeleteUnusedImages}
	}
	return resourceToReclaimFunc
}

// thresholdStopCh is a ThresholdStopCh which can only be closed after notifierRefreshInterval time has passed
type thresholdStopCh struct {
	lock      sync.Mutex
	ch        chan struct{}
	startTime time.Time
	//  used to track time
	clock clock.Clock
}

// NewInitialStopCh returns a ThresholdStopCh which can be closed immediately
func NewInitialStopCh(clock clock.Clock) ThresholdStopCh {
	return &thresholdStopCh{ch: make(chan struct{}), clock: clock}
}

// implements ThresholdStopCh.Reset
func (t *thresholdStopCh) Reset() (closed bool) {
	t.lock.Lock()
	defer t.lock.Unlock()
	closed = t.clock.Since(t.startTime) > notifierRefreshInterval
	if closed {
		// close the old channel and reopen a new one
		close(t.ch)
		t.startTime = t.clock.Now()
		t.ch = make(chan struct{})
	}
	return
}

// implements ThresholdStopCh.Ch
func (t *thresholdStopCh) Ch() <-chan struct{} {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.ch
}
