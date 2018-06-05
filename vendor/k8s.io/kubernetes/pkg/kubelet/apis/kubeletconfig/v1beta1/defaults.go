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

package v1beta1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/kubelet/qos"
	kubetypes "k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/pkg/master/ports"
	utilpointer "k8s.io/kubernetes/pkg/util/pointer"
)

const (
	DefaultIPTablesMasqueradeBit = 14
	DefaultIPTablesDropBit       = 15
)

var (
	zeroDuration = metav1.Duration{}
	// Refer to [Node Allocatable](https://git.k8s.io/community/contributors/design-proposals/node/node-allocatable.md) doc for more information.
	DefaultNodeAllocatableEnforcement = []string{"pods"}
)

func addDefaultingFuncs(scheme *kruntime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_KubeletConfiguration(obj *KubeletConfiguration) {
	if obj.SyncFrequency == zeroDuration {
		obj.SyncFrequency = metav1.Duration{Duration: 1 * time.Minute}
	}
	if obj.FileCheckFrequency == zeroDuration {
		obj.FileCheckFrequency = metav1.Duration{Duration: 20 * time.Second}
	}
	if obj.HTTPCheckFrequency == zeroDuration {
		obj.HTTPCheckFrequency = metav1.Duration{Duration: 20 * time.Second}
	}
	if obj.Address == "" {
		obj.Address = "0.0.0.0"
	}
	if obj.Port == 0 {
		obj.Port = ports.KubeletPort
	}
	if obj.Authentication.Anonymous.Enabled == nil {
		obj.Authentication.Anonymous.Enabled = utilpointer.BoolPtr(false)
	}
	if obj.Authentication.Webhook.Enabled == nil {
		obj.Authentication.Webhook.Enabled = utilpointer.BoolPtr(true)
	}
	if obj.Authentication.Webhook.CacheTTL == zeroDuration {
		obj.Authentication.Webhook.CacheTTL = metav1.Duration{Duration: 2 * time.Minute}
	}
	if obj.Authorization.Mode == "" {
		obj.Authorization.Mode = KubeletAuthorizationModeWebhook
	}
	if obj.Authorization.Webhook.CacheAuthorizedTTL == zeroDuration {
		obj.Authorization.Webhook.CacheAuthorizedTTL = metav1.Duration{Duration: 5 * time.Minute}
	}
	if obj.Authorization.Webhook.CacheUnauthorizedTTL == zeroDuration {
		obj.Authorization.Webhook.CacheUnauthorizedTTL = metav1.Duration{Duration: 30 * time.Second}
	}
	if obj.RegistryPullQPS == nil {
		obj.RegistryPullQPS = utilpointer.Int32Ptr(5)
	}
	if obj.RegistryBurst == 0 {
		obj.RegistryBurst = 10
	}
	if obj.EventRecordQPS == nil {
		obj.EventRecordQPS = utilpointer.Int32Ptr(5)
	}
	if obj.EventBurst == 0 {
		obj.EventBurst = 10
	}
	if obj.EnableDebuggingHandlers == nil {
		obj.EnableDebuggingHandlers = utilpointer.BoolPtr(true)
	}
	if obj.HealthzPort == nil {
		obj.HealthzPort = utilpointer.Int32Ptr(10248)
	}
	if obj.HealthzBindAddress == "" {
		obj.HealthzBindAddress = "127.0.0.1"
	}
	if obj.OOMScoreAdj == nil {
		obj.OOMScoreAdj = utilpointer.Int32Ptr(int32(qos.KubeletOOMScoreAdj))
	}
	if obj.StreamingConnectionIdleTimeout == zeroDuration {
		obj.StreamingConnectionIdleTimeout = metav1.Duration{Duration: 4 * time.Hour}
	}
	if obj.NodeStatusUpdateFrequency == zeroDuration {
		obj.NodeStatusUpdateFrequency = metav1.Duration{Duration: 10 * time.Second}
	}
	if obj.ImageMinimumGCAge == zeroDuration {
		obj.ImageMinimumGCAge = metav1.Duration{Duration: 2 * time.Minute}
	}
	if obj.ImageGCHighThresholdPercent == nil {
		// default is below docker's default dm.min_free_space of 90%
		obj.ImageGCHighThresholdPercent = utilpointer.Int32Ptr(85)
	}
	if obj.ImageGCLowThresholdPercent == nil {
		obj.ImageGCLowThresholdPercent = utilpointer.Int32Ptr(80)
	}
	if obj.VolumeStatsAggPeriod == zeroDuration {
		obj.VolumeStatsAggPeriod = metav1.Duration{Duration: time.Minute}
	}
	if obj.CgroupsPerQOS == nil {
		obj.CgroupsPerQOS = utilpointer.BoolPtr(true)
	}
	if obj.CgroupDriver == "" {
		obj.CgroupDriver = "cgroupfs"
	}
	if obj.CPUManagerPolicy == "" {
		obj.CPUManagerPolicy = "none"
	}
	if obj.CPUManagerReconcilePeriod == zeroDuration {
		// Keep the same as default NodeStatusUpdateFrequency
		obj.CPUManagerReconcilePeriod = metav1.Duration{Duration: 10 * time.Second}
	}
	if obj.RuntimeRequestTimeout == zeroDuration {
		obj.RuntimeRequestTimeout = metav1.Duration{Duration: 2 * time.Minute}
	}
	if obj.HairpinMode == "" {
		obj.HairpinMode = PromiscuousBridge
	}
	if obj.MaxPods == 0 {
		obj.MaxPods = 110
	}
	if obj.PodPidsLimit == nil {
		temp := int64(-1)
		obj.PodPidsLimit = &temp
	}
	if obj.ResolverConfig == "" {
		obj.ResolverConfig = kubetypes.ResolvConfDefault
	}
	if obj.CPUCFSQuota == nil {
		obj.CPUCFSQuota = utilpointer.BoolPtr(true)
	}
	if obj.MaxOpenFiles == 0 {
		obj.MaxOpenFiles = 1000000
	}
	if obj.ContentType == "" {
		obj.ContentType = "application/vnd.kubernetes.protobuf"
	}
	if obj.KubeAPIQPS == nil {
		obj.KubeAPIQPS = utilpointer.Int32Ptr(5)
	}
	if obj.KubeAPIBurst == 0 {
		obj.KubeAPIBurst = 10
	}
	if obj.SerializeImagePulls == nil {
		obj.SerializeImagePulls = utilpointer.BoolPtr(true)
	}
	if obj.EvictionHard == nil {
		obj.EvictionHard = map[string]string{
			"memory.available":  "100Mi",
			"nodefs.available":  "10%",
			"nodefs.inodesFree": "5%",
			"imagefs.available": "15%",
		}
	}
	if obj.EvictionPressureTransitionPeriod == zeroDuration {
		obj.EvictionPressureTransitionPeriod = metav1.Duration{Duration: 5 * time.Minute}
	}
	if obj.EnableControllerAttachDetach == nil {
		obj.EnableControllerAttachDetach = utilpointer.BoolPtr(true)
	}
	if obj.MakeIPTablesUtilChains == nil {
		obj.MakeIPTablesUtilChains = utilpointer.BoolPtr(true)
	}
	if obj.IPTablesMasqueradeBit == nil {
		obj.IPTablesMasqueradeBit = utilpointer.Int32Ptr(DefaultIPTablesMasqueradeBit)
	}
	if obj.IPTablesDropBit == nil {
		obj.IPTablesDropBit = utilpointer.Int32Ptr(DefaultIPTablesDropBit)
	}
	if obj.FailSwapOn == nil {
		obj.FailSwapOn = utilpointer.BoolPtr(true)
	}
	if obj.ContainerLogMaxSize == "" {
		obj.ContainerLogMaxSize = "10Mi"
	}
	if obj.ContainerLogMaxFiles == nil {
		obj.ContainerLogMaxFiles = utilpointer.Int32Ptr(5)
	}
	if obj.EnforceNodeAllocatable == nil {
		obj.EnforceNodeAllocatable = DefaultNodeAllocatableEnforcement
	}
}
