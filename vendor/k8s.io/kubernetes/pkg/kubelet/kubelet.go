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

package kubelet

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/glog"

	cadvisorapi "github.com/google/cadvisor/info/v1"
	cadvisorapiv2 "github.com/google/cadvisor/info/v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/clock"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/certificate"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/integer"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/features"
	internalapi "k8s.io/kubernetes/pkg/kubelet/apis/cri"
	kubeletconfiginternal "k8s.io/kubernetes/pkg/kubelet/apis/kubeletconfig"
	"k8s.io/kubernetes/pkg/kubelet/cadvisor"
	kubeletcertificate "k8s.io/kubernetes/pkg/kubelet/certificate"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager"
	"k8s.io/kubernetes/pkg/kubelet/cm"
	"k8s.io/kubernetes/pkg/kubelet/config"
	"k8s.io/kubernetes/pkg/kubelet/configmap"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/kubelet/dockershim"
	dockerremote "k8s.io/kubernetes/pkg/kubelet/dockershim/remote"
	"k8s.io/kubernetes/pkg/kubelet/events"
	"k8s.io/kubernetes/pkg/kubelet/eviction"
	"k8s.io/kubernetes/pkg/kubelet/images"
	"k8s.io/kubernetes/pkg/kubelet/kubeletconfig"
	"k8s.io/kubernetes/pkg/kubelet/kuberuntime"
	"k8s.io/kubernetes/pkg/kubelet/lifecycle"
	"k8s.io/kubernetes/pkg/kubelet/logs"
	"k8s.io/kubernetes/pkg/kubelet/metrics"
	"k8s.io/kubernetes/pkg/kubelet/metrics/collectors"
	"k8s.io/kubernetes/pkg/kubelet/network/dns"
	"k8s.io/kubernetes/pkg/kubelet/pleg"
	kubepod "k8s.io/kubernetes/pkg/kubelet/pod"
	"k8s.io/kubernetes/pkg/kubelet/preemption"
	"k8s.io/kubernetes/pkg/kubelet/prober"
	proberesults "k8s.io/kubernetes/pkg/kubelet/prober/results"
	"k8s.io/kubernetes/pkg/kubelet/remote"
	"k8s.io/kubernetes/pkg/kubelet/secret"
	"k8s.io/kubernetes/pkg/kubelet/server"
	serverstats "k8s.io/kubernetes/pkg/kubelet/server/stats"
	"k8s.io/kubernetes/pkg/kubelet/server/streaming"
	"k8s.io/kubernetes/pkg/kubelet/stats"
	"k8s.io/kubernetes/pkg/kubelet/status"
	"k8s.io/kubernetes/pkg/kubelet/sysctl"
	kubetypes "k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/pkg/kubelet/util/format"
	"k8s.io/kubernetes/pkg/kubelet/util/queue"
	"k8s.io/kubernetes/pkg/kubelet/util/sliceutils"
	"k8s.io/kubernetes/pkg/kubelet/volumemanager"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/security/apparmor"
	utildbus "k8s.io/kubernetes/pkg/util/dbus"
	kubeio "k8s.io/kubernetes/pkg/util/io"
	utilipt "k8s.io/kubernetes/pkg/util/iptables"
	"k8s.io/kubernetes/pkg/util/mount"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
	"k8s.io/kubernetes/pkg/util/oom"
	"k8s.io/kubernetes/pkg/volume"
	utilexec "k8s.io/utils/exec"
)

const (
	// Max amount of time to wait for the container runtime to come up.
	maxWaitForContainerRuntime = 30 * time.Second

	// nodeStatusUpdateRetry specifies how many times kubelet retries when posting node status failed.
	nodeStatusUpdateRetry = 5

	// ContainerLogsDir is the location of container logs.
	ContainerLogsDir = "/var/log/containers"

	// MaxContainerBackOff is the max backoff period, exported for the e2e test
	MaxContainerBackOff = 300 * time.Second

	// Capacity of the channel for storing pods to kill. A small number should
	// suffice because a goroutine is dedicated to check the channel and does
	// not block on anything else.
	podKillingChannelCapacity = 50

	// Period for performing global cleanup tasks.
	housekeepingPeriod = time.Second * 2

	// Period for performing eviction monitoring.
	// TODO ensure this is in sync with internal cadvisor housekeeping.
	evictionMonitoringPeriod = time.Second * 10

	// The path in containers' filesystems where the hosts file is mounted.
	etcHostsPath = "/etc/hosts"

	// Capacity of the channel for receiving pod lifecycle events. This number
	// is a bit arbitrary and may be adjusted in the future.
	plegChannelCapacity = 1000

	// Generic PLEG relies on relisting for discovering container events.
	// A longer period means that kubelet will take longer to detect container
	// changes and to update pod status. On the other hand, a shorter period
	// will cause more frequent relisting (e.g., container runtime operations),
	// leading to higher cpu usage.
	// Note that even though we set the period to 1s, the relisting itself can
	// take more than 1s to finish if the container runtime responds slowly
	// and/or when there are many container changes in one cycle.
	plegRelistPeriod = time.Second * 1

	// backOffPeriod is the period to back off when pod syncing results in an
	// error. It is also used as the base period for the exponential backoff
	// container restarts and image pulls.
	backOffPeriod = time.Second * 10

	// ContainerGCPeriod is the period for performing container garbage collection.
	ContainerGCPeriod = time.Minute
	// ImageGCPeriod is the period for performing image garbage collection.
	ImageGCPeriod = 5 * time.Minute

	// Minimum number of dead containers to keep in a pod
	minDeadContainerInPod = 1
)

// SyncHandler is an interface implemented by Kubelet, for testability
type SyncHandler interface {
	HandlePodAdditions(pods []*v1.Pod)
	HandlePodUpdates(pods []*v1.Pod)
	HandlePodRemoves(pods []*v1.Pod)
	HandlePodReconcile(pods []*v1.Pod)
	HandlePodSyncs(pods []*v1.Pod)
	HandlePodCleanups() error
}

// Option is a functional option type for Kubelet
type Option func(*Kubelet)

// Bootstrap is a bootstrapping interface for kubelet, targets the initialization protocol
type Bootstrap interface {
	GetConfiguration() kubeletconfiginternal.KubeletConfiguration
	BirthCry()
	StartGarbageCollection()
	ListenAndServe(address net.IP, port uint, tlsOptions *server.TLSOptions, auth server.AuthInterface, enableDebuggingHandlers, enableContentionProfiling bool)
	ListenAndServeReadOnly(address net.IP, port uint)
	Run(<-chan kubetypes.PodUpdate)
	RunOnce(<-chan kubetypes.PodUpdate) ([]RunPodResult, error)
}

// Builder creates and initializes a Kubelet instance
type Builder func(kubeCfg *kubeletconfiginternal.KubeletConfiguration,
	kubeDeps *Dependencies,
	crOptions *config.ContainerRuntimeOptions,
	containerRuntime string,
	runtimeCgroups string,
	hostnameOverride string,
	nodeIP string,
	providerID string,
	cloudProvider string,
	certDirectory string,
	rootDirectory string,
	registerNode bool,
	registerWithTaints []api.Taint,
	allowedUnsafeSysctls []string,
	remoteRuntimeEndpoint string,
	remoteImageEndpoint string,
	experimentalMounterPath string,
	experimentalKernelMemcgNotification bool,
	experimentalCheckNodeCapabilitiesBeforeMount bool,
	experimentalNodeAllocatableIgnoreEvictionThreshold bool,
	minimumGCAge metav1.Duration,
	maxPerPodContainerCount int32,
	maxContainerCount int32,
	masterServiceNamespace string,
	registerSchedulable bool,
	nonMasqueradeCIDR string,
	keepTerminatedPodVolumes bool,
	nodeLabels map[string]string,
	seccompProfileRoot string,
	bootstrapCheckpointPath string) (Bootstrap, error)

// Dependencies is a bin for things we might consider "injected dependencies" -- objects constructed
// at runtime that are necessary for running the Kubelet. This is a temporary solution for grouping
// these objects while we figure out a more comprehensive dependency injection story for the Kubelet.
type Dependencies struct {
	Options []Option

	// Injected Dependencies
	Auth                    server.AuthInterface
	CAdvisorInterface       cadvisor.Interface
	Cloud                   cloudprovider.Interface
	ContainerManager        cm.ContainerManager
	DockerClientConfig      *dockershim.ClientConfig
	EventClient             v1core.EventsGetter
	HeartbeatClient         v1core.CoreV1Interface
	KubeClient              clientset.Interface
	ExternalKubeClient      clientset.Interface
	Mounter                 mount.Interface
	OOMAdjuster             *oom.OOMAdjuster
	OSInterface             kubecontainer.OSInterface
	PodConfig               *config.PodConfig
	Recorder                record.EventRecorder
	Writer                  kubeio.Writer
	VolumePlugins           []volume.VolumePlugin
	DynamicPluginProber     volume.DynamicPluginProber
	TLSOptions              *server.TLSOptions
	KubeletConfigController *kubeletconfig.Controller
}

// makePodSourceConfig creates a config.PodConfig from the given
// KubeletConfiguration or returns an error.
func makePodSourceConfig(kubeCfg *kubeletconfiginternal.KubeletConfiguration, kubeDeps *Dependencies, nodeName types.NodeName, bootstrapCheckpointPath string) (*config.PodConfig, error) {
	manifestURLHeader := make(http.Header)
	if len(kubeCfg.StaticPodURLHeader) > 0 {
		for k, v := range kubeCfg.StaticPodURLHeader {
			for i := range v {
				manifestURLHeader.Add(k, v[i])
			}
		}
	}

	// source of all configuration
	cfg := config.NewPodConfig(config.PodConfigNotificationIncremental, kubeDeps.Recorder)

	// define file config source
	if kubeCfg.StaticPodPath != "" {
		glog.Infof("Adding pod path: %v", kubeCfg.StaticPodPath)
		config.NewSourceFile(kubeCfg.StaticPodPath, nodeName, kubeCfg.FileCheckFrequency.Duration, cfg.Channel(kubetypes.FileSource))
	}

	// define url config source
	if kubeCfg.StaticPodURL != "" {
		glog.Infof("Adding pod url %q with HTTP header %v", kubeCfg.StaticPodURL, manifestURLHeader)
		config.NewSourceURL(kubeCfg.StaticPodURL, manifestURLHeader, nodeName, kubeCfg.HTTPCheckFrequency.Duration, cfg.Channel(kubetypes.HTTPSource))
	}

	// Restore from the checkpoint path
	// NOTE: This MUST happen before creating the apiserver source
	// below, or the checkpoint would override the source of truth.

	var updatechannel chan<- interface{}
	if bootstrapCheckpointPath != "" {
		glog.Infof("Adding checkpoint path: %v", bootstrapCheckpointPath)
		updatechannel = cfg.Channel(kubetypes.ApiserverSource)
		err := cfg.Restore(bootstrapCheckpointPath, updatechannel)
		if err != nil {
			return nil, err
		}
	}

	if kubeDeps.KubeClient != nil {
		glog.Infof("Watching apiserver")
		if updatechannel == nil {
			updatechannel = cfg.Channel(kubetypes.ApiserverSource)
		}
		config.NewSourceApiserver(kubeDeps.KubeClient, nodeName, updatechannel)
	}
	return cfg, nil
}

func getRuntimeAndImageServices(remoteRuntimeEndpoint string, remoteImageEndpoint string, runtimeRequestTimeout metav1.Duration) (internalapi.RuntimeService, internalapi.ImageManagerService, error) {
	rs, err := remote.NewRemoteRuntimeService(remoteRuntimeEndpoint, runtimeRequestTimeout.Duration)
	if err != nil {
		return nil, nil, err
	}
	is, err := remote.NewRemoteImageService(remoteImageEndpoint, runtimeRequestTimeout.Duration)
	if err != nil {
		return nil, nil, err
	}
	return rs, is, err
}

// NewMainKubelet instantiates a new Kubelet object along with all the required internal modules.
// No initialization of Kubelet and its modules should happen here.
func NewMainKubelet(kubeCfg *kubeletconfiginternal.KubeletConfiguration,
	kubeDeps *Dependencies,
	crOptions *config.ContainerRuntimeOptions,
	containerRuntime string,
	runtimeCgroups string,
	hostnameOverride string,
	nodeIP string,
	providerID string,
	cloudProvider string,
	certDirectory string,
	rootDirectory string,
	registerNode bool,
	registerWithTaints []api.Taint,
	allowedUnsafeSysctls []string,
	remoteRuntimeEndpoint string,
	remoteImageEndpoint string,
	experimentalMounterPath string,
	experimentalKernelMemcgNotification bool,
	experimentalCheckNodeCapabilitiesBeforeMount bool,
	experimentalNodeAllocatableIgnoreEvictionThreshold bool,
	minimumGCAge metav1.Duration,
	maxPerPodContainerCount int32,
	maxContainerCount int32,
	masterServiceNamespace string,
	registerSchedulable bool,
	nonMasqueradeCIDR string,
	keepTerminatedPodVolumes bool,
	nodeLabels map[string]string,
	seccompProfileRoot string,
	bootstrapCheckpointPath string) (*Kubelet, error) {
	if rootDirectory == "" {
		return nil, fmt.Errorf("invalid root directory %q", rootDirectory)
	}
	if kubeCfg.SyncFrequency.Duration <= 0 {
		return nil, fmt.Errorf("invalid sync frequency %d", kubeCfg.SyncFrequency.Duration)
	}

	if kubeCfg.MakeIPTablesUtilChains {
		if kubeCfg.IPTablesMasqueradeBit > 31 || kubeCfg.IPTablesMasqueradeBit < 0 {
			return nil, fmt.Errorf("iptables-masquerade-bit is not valid. Must be within [0, 31]")
		}
		if kubeCfg.IPTablesDropBit > 31 || kubeCfg.IPTablesDropBit < 0 {
			return nil, fmt.Errorf("iptables-drop-bit is not valid. Must be within [0, 31]")
		}
		if kubeCfg.IPTablesDropBit == kubeCfg.IPTablesMasqueradeBit {
			return nil, fmt.Errorf("iptables-masquerade-bit and iptables-drop-bit must be different")
		}
	}

	hostname := nodeutil.GetHostname(hostnameOverride)
	// Query the cloud provider for our node name, default to hostname
	nodeName := types.NodeName(hostname)
	cloudIPs := []net.IP{}
	cloudNames := []string{}
	if kubeDeps.Cloud != nil {
		var err error
		instances, ok := kubeDeps.Cloud.Instances()
		if !ok {
			return nil, fmt.Errorf("failed to get instances from cloud provider")
		}

		nodeName, err = instances.CurrentNodeName(context.TODO(), hostname)
		if err != nil {
			return nil, fmt.Errorf("error fetching current instance name from cloud provider: %v", err)
		}

		glog.V(2).Infof("cloud provider determined current node name to be %s", nodeName)

		if utilfeature.DefaultFeatureGate.Enabled(features.RotateKubeletServerCertificate) {
			nodeAddresses, err := instances.NodeAddresses(context.TODO(), nodeName)
			if err != nil {
				return nil, fmt.Errorf("failed to get the addresses of the current instance from the cloud provider: %v", err)
			}
			for _, nodeAddress := range nodeAddresses {
				switch nodeAddress.Type {
				case v1.NodeExternalIP, v1.NodeInternalIP:
					ip := net.ParseIP(nodeAddress.Address)
					if ip != nil && !ip.IsLoopback() {
						cloudIPs = append(cloudIPs, ip)
					}
				case v1.NodeExternalDNS, v1.NodeInternalDNS, v1.NodeHostName:
					cloudNames = append(cloudNames, nodeAddress.Address)
				}
			}
		}

	}

	if kubeDeps.PodConfig == nil {
		var err error
		kubeDeps.PodConfig, err = makePodSourceConfig(kubeCfg, kubeDeps, nodeName, bootstrapCheckpointPath)
		if err != nil {
			return nil, err
		}
	}

	containerGCPolicy := kubecontainer.ContainerGCPolicy{
		MinAge:             minimumGCAge.Duration,
		MaxPerPodContainer: int(maxPerPodContainerCount),
		MaxContainers:      int(maxContainerCount),
	}

	daemonEndpoints := &v1.NodeDaemonEndpoints{
		KubeletEndpoint: v1.DaemonEndpoint{Port: kubeCfg.Port},
	}

	imageGCPolicy := images.ImageGCPolicy{
		MinAge:               kubeCfg.ImageMinimumGCAge.Duration,
		HighThresholdPercent: int(kubeCfg.ImageGCHighThresholdPercent),
		LowThresholdPercent:  int(kubeCfg.ImageGCLowThresholdPercent),
	}

	enforceNodeAllocatable := kubeCfg.EnforceNodeAllocatable
	if experimentalNodeAllocatableIgnoreEvictionThreshold {
		// Do not provide kubeCfg.EnforceNodeAllocatable to eviction threshold parsing if we are not enforcing Evictions
		enforceNodeAllocatable = []string{}
	}
	thresholds, err := eviction.ParseThresholdConfig(enforceNodeAllocatable, kubeCfg.EvictionHard, kubeCfg.EvictionSoft, kubeCfg.EvictionSoftGracePeriod, kubeCfg.EvictionMinimumReclaim)
	if err != nil {
		return nil, err
	}
	evictionConfig := eviction.Config{
		PressureTransitionPeriod: kubeCfg.EvictionPressureTransitionPeriod.Duration,
		MaxPodGracePeriodSeconds: int64(kubeCfg.EvictionMaxPodGracePeriod),
		Thresholds:               thresholds,
		KernelMemcgNotification:  experimentalKernelMemcgNotification,
	}

	serviceIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	if kubeDeps.KubeClient != nil {
		serviceLW := cache.NewListWatchFromClient(kubeDeps.KubeClient.CoreV1().RESTClient(), "services", metav1.NamespaceAll, fields.Everything())
		r := cache.NewReflector(serviceLW, &v1.Service{}, serviceIndexer, 0)
		go r.Run(wait.NeverStop)
	}
	serviceLister := corelisters.NewServiceLister(serviceIndexer)

	nodeIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	if kubeDeps.KubeClient != nil {
		fieldSelector := fields.Set{api.ObjectNameField: string(nodeName)}.AsSelector()
		nodeLW := cache.NewListWatchFromClient(kubeDeps.KubeClient.CoreV1().RESTClient(), "nodes", metav1.NamespaceAll, fieldSelector)
		r := cache.NewReflector(nodeLW, &v1.Node{}, nodeIndexer, 0)
		go r.Run(wait.NeverStop)
	}
	nodeInfo := &predicates.CachedNodeInfo{NodeLister: corelisters.NewNodeLister(nodeIndexer)}

	// TODO: get the real node object of ourself,
	// and use the real node name and UID.
	// TODO: what is namespace for node?
	nodeRef := &v1.ObjectReference{
		Kind:      "Node",
		Name:      string(nodeName),
		UID:       types.UID(nodeName),
		Namespace: "",
	}

	containerRefManager := kubecontainer.NewRefManager()

	oomWatcher := NewOOMWatcher(kubeDeps.CAdvisorInterface, kubeDeps.Recorder)

	clusterDNS := make([]net.IP, 0, len(kubeCfg.ClusterDNS))
	for _, ipEntry := range kubeCfg.ClusterDNS {
		ip := net.ParseIP(ipEntry)
		if ip == nil {
			glog.Warningf("Invalid clusterDNS ip '%q'", ipEntry)
		} else {
			clusterDNS = append(clusterDNS, ip)
		}
	}
	httpClient := &http.Client{}
	parsedNodeIP := net.ParseIP(nodeIP)

	klet := &Kubelet{
		hostname:                       hostname,
		nodeName:                       nodeName,
		kubeClient:                     kubeDeps.KubeClient,
		heartbeatClient:                kubeDeps.HeartbeatClient,
		rootDirectory:                  rootDirectory,
		resyncInterval:                 kubeCfg.SyncFrequency.Duration,
		sourcesReady:                   config.NewSourcesReady(kubeDeps.PodConfig.SeenAllSources),
		registerNode:                   registerNode,
		registerWithTaints:             registerWithTaints,
		registerSchedulable:            registerSchedulable,
		dnsConfigurer:                  dns.NewConfigurer(kubeDeps.Recorder, nodeRef, parsedNodeIP, clusterDNS, kubeCfg.ClusterDomain, kubeCfg.ResolverConfig),
		serviceLister:                  serviceLister,
		nodeInfo:                       nodeInfo,
		masterServiceNamespace:         masterServiceNamespace,
		streamingConnectionIdleTimeout: kubeCfg.StreamingConnectionIdleTimeout.Duration,
		recorder:                       kubeDeps.Recorder,
		cadvisor:                       kubeDeps.CAdvisorInterface,
		cloud:                          kubeDeps.Cloud,
		externalCloudProvider:     cloudprovider.IsExternal(cloudProvider),
		providerID:                providerID,
		nodeRef:                   nodeRef,
		nodeLabels:                nodeLabels,
		nodeStatusUpdateFrequency: kubeCfg.NodeStatusUpdateFrequency.Duration,
		os:                   kubeDeps.OSInterface,
		oomWatcher:           oomWatcher,
		cgroupsPerQOS:        kubeCfg.CgroupsPerQOS,
		cgroupRoot:           kubeCfg.CgroupRoot,
		mounter:              kubeDeps.Mounter,
		writer:               kubeDeps.Writer,
		maxPods:              int(kubeCfg.MaxPods),
		podsPerCore:          int(kubeCfg.PodsPerCore),
		syncLoopMonitor:      atomic.Value{},
		daemonEndpoints:      daemonEndpoints,
		containerManager:     kubeDeps.ContainerManager,
		containerRuntimeName: containerRuntime,
		nodeIP:               parsedNodeIP,
		clock:                clock.RealClock{},
		enableControllerAttachDetach:            kubeCfg.EnableControllerAttachDetach,
		iptClient:                               utilipt.New(utilexec.New(), utildbus.New(), utilipt.ProtocolIpv4),
		makeIPTablesUtilChains:                  kubeCfg.MakeIPTablesUtilChains,
		iptablesMasqueradeBit:                   int(kubeCfg.IPTablesMasqueradeBit),
		iptablesDropBit:                         int(kubeCfg.IPTablesDropBit),
		experimentalHostUserNamespaceDefaulting: utilfeature.DefaultFeatureGate.Enabled(features.ExperimentalHostUserNamespaceDefaultingGate),
		keepTerminatedPodVolumes:                keepTerminatedPodVolumes,
	}

	if klet.cloud != nil {
		klet.cloudproviderRequestParallelism = make(chan int, 1)
		klet.cloudproviderRequestSync = make(chan int)
		// TODO(jchaloup): Make it configurable via --cloud-provider-request-timeout
		klet.cloudproviderRequestTimeout = 10 * time.Second
	}

	secretManager := secret.NewCachingSecretManager(
		kubeDeps.KubeClient, secret.GetObjectTTLFromNodeFunc(klet.GetNode))
	klet.secretManager = secretManager

	configMapManager := configmap.NewCachingConfigMapManager(
		kubeDeps.KubeClient, configmap.GetObjectTTLFromNodeFunc(klet.GetNode))
	klet.configMapManager = configMapManager

	if klet.experimentalHostUserNamespaceDefaulting {
		glog.Infof("Experimental host user namespace defaulting is enabled.")
	}

	machineInfo, err := klet.cadvisor.MachineInfo()
	if err != nil {
		return nil, err
	}
	klet.machineInfo = machineInfo

	imageBackOff := flowcontrol.NewBackOff(backOffPeriod, MaxContainerBackOff)

	klet.livenessManager = proberesults.NewManager()

	klet.podCache = kubecontainer.NewCache()
	var checkpointManager checkpointmanager.CheckpointManager
	if bootstrapCheckpointPath != "" {
		checkpointManager, err = checkpointmanager.NewCheckpointManager(bootstrapCheckpointPath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize checkpoint manager: %+v", err)
		}
	}
	// podManager is also responsible for keeping secretManager and configMapManager contents up-to-date.
	klet.podManager = kubepod.NewBasicPodManager(kubepod.NewBasicMirrorClient(klet.kubeClient), secretManager, configMapManager, checkpointManager)

	if remoteRuntimeEndpoint != "" {
		// remoteImageEndpoint is same as remoteRuntimeEndpoint if not explicitly specified
		if remoteImageEndpoint == "" {
			remoteImageEndpoint = remoteRuntimeEndpoint
		}
	}

	// TODO: These need to become arguments to a standalone docker shim.
	pluginSettings := dockershim.NetworkPluginSettings{
		HairpinMode:        kubeletconfiginternal.HairpinMode(kubeCfg.HairpinMode),
		NonMasqueradeCIDR:  nonMasqueradeCIDR,
		PluginName:         crOptions.NetworkPluginName,
		PluginConfDir:      crOptions.CNIConfDir,
		PluginBinDirString: crOptions.CNIBinDir,
		MTU:                int(crOptions.NetworkPluginMTU),
	}

	klet.resourceAnalyzer = serverstats.NewResourceAnalyzer(klet, kubeCfg.VolumeStatsAggPeriod.Duration)

	if containerRuntime == "rkt" {
		glog.Fatalln("rktnetes has been deprecated in favor of rktlet. Please see https://github.com/kubernetes-incubator/rktlet for more information.")
	}

	// if left at nil, that means it is unneeded
	var legacyLogProvider kuberuntime.LegacyLogProvider

	switch containerRuntime {
	case kubetypes.DockerContainerRuntime:
		// Create and start the CRI shim running as a grpc server.
		streamingConfig := getStreamingConfig(kubeCfg, kubeDeps)
		ds, err := dockershim.NewDockerService(kubeDeps.DockerClientConfig, crOptions.PodSandboxImage, streamingConfig,
			&pluginSettings, runtimeCgroups, kubeCfg.CgroupDriver, crOptions.DockershimRootDirectory,
			crOptions.DockerDisableSharedPID)
		if err != nil {
			return nil, err
		}
		// For now, the CRI shim redirects the streaming requests to the
		// kubelet, which handles the requests using DockerService..
		klet.criHandler = ds

		// The unix socket for kubelet <-> dockershim communication.
		glog.V(5).Infof("RemoteRuntimeEndpoint: %q, RemoteImageEndpoint: %q",
			remoteRuntimeEndpoint,
			remoteImageEndpoint)
		glog.V(2).Infof("Starting the GRPC server for the docker CRI shim.")
		server := dockerremote.NewDockerServer(remoteRuntimeEndpoint, ds)
		if err := server.Start(); err != nil {
			return nil, err
		}

		// Create dockerLegacyService when the logging driver is not supported.
		supported, err := ds.IsCRISupportedLogDriver()
		if err != nil {
			return nil, err
		}
		if !supported {
			klet.dockerLegacyService = ds
			legacyLogProvider = ds
		}
	case kubetypes.RemoteContainerRuntime:
		// No-op.
		break
	default:
		return nil, fmt.Errorf("unsupported CRI runtime: %q", containerRuntime)
	}
	runtimeService, imageService, err := getRuntimeAndImageServices(remoteRuntimeEndpoint, remoteImageEndpoint, kubeCfg.RuntimeRequestTimeout)
	if err != nil {
		return nil, err
	}
	klet.runtimeService = runtimeService
	runtime, err := kuberuntime.NewKubeGenericRuntimeManager(
		kubecontainer.FilterEventRecorder(kubeDeps.Recorder),
		klet.livenessManager,
		seccompProfileRoot,
		containerRefManager,
		machineInfo,
		klet,
		kubeDeps.OSInterface,
		klet,
		httpClient,
		imageBackOff,
		kubeCfg.SerializeImagePulls,
		float32(kubeCfg.RegistryPullQPS),
		int(kubeCfg.RegistryBurst),
		kubeCfg.CPUCFSQuota,
		runtimeService,
		imageService,
		kubeDeps.ContainerManager.InternalContainerLifecycle(),
		legacyLogProvider,
	)
	if err != nil {
		return nil, err
	}
	klet.containerRuntime = runtime
	klet.runner = runtime

	if cadvisor.UsingLegacyCadvisorStats(containerRuntime, remoteRuntimeEndpoint) {
		klet.StatsProvider = stats.NewCadvisorStatsProvider(
			klet.cadvisor,
			klet.resourceAnalyzer,
			klet.podManager,
			klet.runtimeCache,
			klet.containerRuntime)
	} else {
		klet.StatsProvider = stats.NewCRIStatsProvider(
			klet.cadvisor,
			klet.resourceAnalyzer,
			klet.podManager,
			klet.runtimeCache,
			runtimeService,
			imageService,
			stats.NewLogMetricsService())
	}

	klet.pleg = pleg.NewGenericPLEG(klet.containerRuntime, plegChannelCapacity, plegRelistPeriod, klet.podCache, clock.RealClock{})
	klet.runtimeState = newRuntimeState(maxWaitForContainerRuntime)
	klet.runtimeState.addHealthCheck("PLEG", klet.pleg.Healthy)
	klet.updatePodCIDR(kubeCfg.PodCIDR)

	// setup containerGC
	containerGC, err := kubecontainer.NewContainerGC(klet.containerRuntime, containerGCPolicy, klet.sourcesReady)
	if err != nil {
		return nil, err
	}
	klet.containerGC = containerGC
	klet.containerDeletor = newPodContainerDeletor(klet.containerRuntime, integer.IntMax(containerGCPolicy.MaxPerPodContainer, minDeadContainerInPod))

	// setup imageManager
	imageManager, err := images.NewImageGCManager(klet.containerRuntime, klet.StatsProvider, kubeDeps.Recorder, nodeRef, imageGCPolicy, crOptions.PodSandboxImage)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize image manager: %v", err)
	}
	klet.imageManager = imageManager

	if containerRuntime == kubetypes.RemoteContainerRuntime && utilfeature.DefaultFeatureGate.Enabled(features.CRIContainerLogRotation) {
		// setup containerLogManager for CRI container runtime
		containerLogManager, err := logs.NewContainerLogManager(
			klet.runtimeService,
			kubeCfg.ContainerLogMaxSize,
			int(kubeCfg.ContainerLogMaxFiles),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize container log manager: %v", err)
		}
		klet.containerLogManager = containerLogManager
	} else {
		klet.containerLogManager = logs.NewStubContainerLogManager()
	}

	klet.statusManager = status.NewManager(klet.kubeClient, klet.podManager, klet)

	if kubeCfg.ServerTLSBootstrap && kubeDeps.TLSOptions != nil && utilfeature.DefaultFeatureGate.Enabled(features.RotateKubeletServerCertificate) {
		var (
			ips   []net.IP
			names []string
		)

		// If the address was explicitly configured, use that. Otherwise, try to
		// discover addresses from the cloudprovider. Otherwise, make a best guess.
		if cfgAddress := net.ParseIP(kubeCfg.Address); cfgAddress != nil && !cfgAddress.IsUnspecified() {
			ips = []net.IP{cfgAddress}
			names = []string{klet.GetHostname(), hostnameOverride}
		} else if len(cloudIPs) != 0 || len(cloudNames) != 0 {
			ips = cloudIPs
			names = cloudNames
		} else {
			localIPs, err := allGlobalUnicastIPs()
			if err != nil {
				return nil, err
			}
			ips = localIPs
			names = []string{klet.GetHostname(), hostnameOverride}
		}

		klet.serverCertificateManager, err = kubeletcertificate.NewKubeletServerCertificateManager(klet.kubeClient, kubeCfg, klet.nodeName, ips, names, certDirectory)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize certificate manager: %v", err)
		}
		kubeDeps.TLSOptions.Config.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert := klet.serverCertificateManager.Current()
			if cert == nil {
				return nil, fmt.Errorf("no serving certificate available for the kubelet")
			}
			return cert, nil
		}
	}

	klet.probeManager = prober.NewManager(
		klet.statusManager,
		klet.livenessManager,
		klet.runner,
		containerRefManager,
		kubeDeps.Recorder)

	klet.volumePluginMgr, err =
		NewInitializedVolumePluginMgr(klet, secretManager, configMapManager, kubeDeps.VolumePlugins, kubeDeps.DynamicPluginProber)
	if err != nil {
		return nil, err
	}

	// If the experimentalMounterPathFlag is set, we do not want to
	// check node capabilities since the mount path is not the default
	if len(experimentalMounterPath) != 0 {
		experimentalCheckNodeCapabilitiesBeforeMount = false
		// Replace the nameserver in containerized-mounter's rootfs/etc/resolve.conf with kubelet.ClusterDNS
		// so that service name could be resolved
		klet.dnsConfigurer.SetupDNSinContainerizedMounter(experimentalMounterPath)
	}

	// setup volumeManager
	klet.volumeManager = volumemanager.NewVolumeManager(
		kubeCfg.EnableControllerAttachDetach,
		nodeName,
		klet.podManager,
		klet.statusManager,
		klet.kubeClient,
		klet.volumePluginMgr,
		klet.containerRuntime,
		kubeDeps.Mounter,
		klet.getPodsDir(),
		kubeDeps.Recorder,
		experimentalCheckNodeCapabilitiesBeforeMount,
		keepTerminatedPodVolumes)

	runtimeCache, err := kubecontainer.NewRuntimeCache(klet.containerRuntime)
	if err != nil {
		return nil, err
	}
	klet.runtimeCache = runtimeCache
	klet.reasonCache = NewReasonCache()
	klet.workQueue = queue.NewBasicWorkQueue(klet.clock)
	klet.podWorkers = newPodWorkers(klet.syncPod, kubeDeps.Recorder, klet.workQueue, klet.resyncInterval, backOffPeriod, klet.podCache)

	klet.backOff = flowcontrol.NewBackOff(backOffPeriod, MaxContainerBackOff)
	klet.podKillingCh = make(chan *kubecontainer.PodPair, podKillingChannelCapacity)
	klet.setNodeStatusFuncs = klet.defaultNodeStatusFuncs()

	// setup eviction manager
	evictionManager, evictionAdmitHandler := eviction.NewManager(klet.resourceAnalyzer, evictionConfig, killPodNow(klet.podWorkers, kubeDeps.Recorder), klet.imageManager, klet.containerGC, kubeDeps.Recorder, nodeRef, klet.clock)

	klet.evictionManager = evictionManager
	klet.admitHandlers.AddPodAdmitHandler(evictionAdmitHandler)

	// add sysctl admission
	runtimeSupport, err := sysctl.NewRuntimeAdmitHandler(klet.containerRuntime)
	if err != nil {
		return nil, err
	}
	safeWhitelist, err := sysctl.NewWhitelist(sysctl.SafeSysctlWhitelist(), v1.SysctlsPodAnnotationKey)
	if err != nil {
		return nil, err
	}
	// Safe, whitelisted sysctls can always be used as unsafe sysctls in the spec
	// Hence, we concatenate those two lists.
	safeAndUnsafeSysctls := append(sysctl.SafeSysctlWhitelist(), allowedUnsafeSysctls...)
	unsafeWhitelist, err := sysctl.NewWhitelist(safeAndUnsafeSysctls, v1.UnsafeSysctlsPodAnnotationKey)
	if err != nil {
		return nil, err
	}
	klet.admitHandlers.AddPodAdmitHandler(runtimeSupport)
	klet.admitHandlers.AddPodAdmitHandler(safeWhitelist)
	klet.admitHandlers.AddPodAdmitHandler(unsafeWhitelist)

	// enable active deadline handler
	activeDeadlineHandler, err := newActiveDeadlineHandler(klet.statusManager, kubeDeps.Recorder, klet.clock)
	if err != nil {
		return nil, err
	}
	klet.AddPodSyncLoopHandler(activeDeadlineHandler)
	klet.AddPodSyncHandler(activeDeadlineHandler)

	criticalPodAdmissionHandler := preemption.NewCriticalPodAdmissionHandler(klet.GetActivePods, killPodNow(klet.podWorkers, kubeDeps.Recorder), kubeDeps.Recorder)
	klet.admitHandlers.AddPodAdmitHandler(lifecycle.NewPredicateAdmitHandler(klet.getNodeAnyWay, criticalPodAdmissionHandler, klet.containerManager.UpdatePluginResources))
	// apply functional Option's
	for _, opt := range kubeDeps.Options {
		opt(klet)
	}

	klet.appArmorValidator = apparmor.NewValidator(containerRuntime)
	klet.softAdmitHandlers.AddPodAdmitHandler(lifecycle.NewAppArmorAdmitHandler(klet.appArmorValidator))
	klet.softAdmitHandlers.AddPodAdmitHandler(lifecycle.NewNoNewPrivsAdmitHandler(klet.containerRuntime))
	// Finally, put the most recent version of the config on the Kubelet, so
	// people can see how it was configured.
	klet.kubeletConfiguration = *kubeCfg
	return klet, nil
}

type serviceLister interface {
	List(labels.Selector) ([]*v1.Service, error)
}

// Kubelet is the main kubelet implementation.
type Kubelet struct {
	kubeletConfiguration kubeletconfiginternal.KubeletConfiguration

	hostname        string
	nodeName        types.NodeName
	runtimeCache    kubecontainer.RuntimeCache
	kubeClient      clientset.Interface
	heartbeatClient v1core.CoreV1Interface
	iptClient       utilipt.Interface
	rootDirectory   string

	// podWorkers handle syncing Pods in response to events.
	podWorkers PodWorkers

	// resyncInterval is the interval between periodic full reconciliations of
	// pods on this node.
	resyncInterval time.Duration

	// sourcesReady records the sources seen by the kubelet, it is thread-safe.
	sourcesReady config.SourcesReady

	// podManager is a facade that abstracts away the various sources of pods
	// this Kubelet services.
	podManager kubepod.Manager

	// Needed to observe and respond to situations that could impact node stability
	evictionManager eviction.Manager

	// Optional, defaults to /logs/ from /var/log
	logServer http.Handler
	// Optional, defaults to simple Docker implementation
	runner kubecontainer.ContainerCommandRunner

	// cAdvisor used for container information.
	cadvisor cadvisor.Interface

	// Set to true to have the node register itself with the apiserver.
	registerNode bool
	// List of taints to add to a node object when the kubelet registers itself.
	registerWithTaints []api.Taint
	// Set to true to have the node register itself as schedulable.
	registerSchedulable bool
	// for internal book keeping; access only from within registerWithApiserver
	registrationCompleted bool

	// dnsConfigurer is used for setting up DNS resolver configuration when launching pods.
	dnsConfigurer *dns.Configurer

	// masterServiceNamespace is the namespace that the master service is exposed in.
	masterServiceNamespace string
	// serviceLister knows how to list services
	serviceLister serviceLister
	// nodeInfo knows how to get information about the node for this kubelet.
	nodeInfo predicates.NodeInfo

	// a list of node labels to register
	nodeLabels map[string]string

	// Last timestamp when runtime responded on ping.
	// Mutex is used to protect this value.
	runtimeState *runtimeState

	// Volume plugins.
	volumePluginMgr *volume.VolumePluginMgr

	// Handles container probing.
	probeManager prober.Manager
	// Manages container health check results.
	livenessManager proberesults.Manager

	// How long to keep idle streaming command execution/port forwarding
	// connections open before terminating them
	streamingConnectionIdleTimeout time.Duration

	// The EventRecorder to use
	recorder record.EventRecorder

	// Policy for handling garbage collection of dead containers.
	containerGC kubecontainer.ContainerGC

	// Manager for image garbage collection.
	imageManager images.ImageGCManager

	// Manager for container logs.
	containerLogManager logs.ContainerLogManager

	// Secret manager.
	secretManager secret.Manager

	// ConfigMap manager.
	configMapManager configmap.Manager

	// Cached MachineInfo returned by cadvisor.
	machineInfo *cadvisorapi.MachineInfo

	//Cached RootFsInfo returned by cadvisor
	rootfsInfo *cadvisorapiv2.FsInfo

	// Handles certificate rotations.
	serverCertificateManager certificate.Manager

	// Syncs pods statuses with apiserver; also used as a cache of statuses.
	statusManager status.Manager

	// VolumeManager runs a set of asynchronous loops that figure out which
	// volumes need to be attached/mounted/unmounted/detached based on the pods
	// scheduled on this node and makes it so.
	volumeManager volumemanager.VolumeManager

	// Cloud provider interface.
	cloud cloudprovider.Interface
	// To keep exclusive access to the cloudproviderRequestParallelism
	cloudproviderRequestMux sync.Mutex
	// Keep the count of requests processed in parallel (expected to be 1 at most at a given time)
	cloudproviderRequestParallelism chan int
	// Sync with finished requests
	cloudproviderRequestSync chan int
	// Request timeout
	cloudproviderRequestTimeout time.Duration

	// Indicates that the node initialization happens in an external cloud controller
	externalCloudProvider bool
	// Reference to this node.
	nodeRef *v1.ObjectReference

	// The name of the container runtime
	containerRuntimeName string

	// Container runtime.
	containerRuntime kubecontainer.Runtime

	// Container runtime service (needed by container runtime Start()).
	// TODO(CD): try to make this available without holding a reference in this
	//           struct. For example, by adding a getter to generic runtime.
	runtimeService internalapi.RuntimeService

	// reasonCache caches the failure reason of the last creation of all containers, which is
	// used for generating ContainerStatus.
	reasonCache *ReasonCache

	// nodeStatusUpdateFrequency specifies how often kubelet posts node status to master.
	// Note: be cautious when changing the constant, it must work with nodeMonitorGracePeriod
	// in nodecontroller. There are several constraints:
	// 1. nodeMonitorGracePeriod must be N times more than nodeStatusUpdateFrequency, where
	//    N means number of retries allowed for kubelet to post node status. It is pointless
	//    to make nodeMonitorGracePeriod be less than nodeStatusUpdateFrequency, since there
	//    will only be fresh values from Kubelet at an interval of nodeStatusUpdateFrequency.
	//    The constant must be less than podEvictionTimeout.
	// 2. nodeStatusUpdateFrequency needs to be large enough for kubelet to generate node
	//    status. Kubelet may fail to update node status reliably if the value is too small,
	//    as it takes time to gather all necessary node information.
	nodeStatusUpdateFrequency time.Duration

	// Generates pod events.
	pleg pleg.PodLifecycleEventGenerator

	// Store kubecontainer.PodStatus for all pods.
	podCache kubecontainer.Cache

	// os is a facade for various syscalls that need to be mocked during testing.
	os kubecontainer.OSInterface

	// Watcher of out of memory events.
	oomWatcher OOMWatcher

	// Monitor resource usage
	resourceAnalyzer serverstats.ResourceAnalyzer

	// Whether or not we should have the QOS cgroup hierarchy for resource management
	cgroupsPerQOS bool

	// If non-empty, pass this to the container runtime as the root cgroup.
	cgroupRoot string

	// Mounter to use for volumes.
	mounter mount.Interface

	// Writer interface to use for volumes.
	writer kubeio.Writer

	// Manager of non-Runtime containers.
	containerManager cm.ContainerManager

	// Maximum Number of Pods which can be run by this Kubelet
	maxPods int

	// Monitor Kubelet's sync loop
	syncLoopMonitor atomic.Value

	// Container restart Backoff
	backOff *flowcontrol.Backoff

	// Channel for sending pods to kill.
	podKillingCh chan *kubecontainer.PodPair

	// Information about the ports which are opened by daemons on Node running this Kubelet server.
	daemonEndpoints *v1.NodeDaemonEndpoints

	// A queue used to trigger pod workers.
	workQueue queue.WorkQueue

	// oneTimeInitializer is used to initialize modules that are dependent on the runtime to be up.
	oneTimeInitializer sync.Once

	// If non-nil, use this IP address for the node
	nodeIP net.IP

	// If non-nil, this is a unique identifier for the node in an external database, eg. cloudprovider
	providerID string

	// clock is an interface that provides time related functionality in a way that makes it
	// easy to test the code.
	clock clock.Clock

	// handlers called during the tryUpdateNodeStatus cycle
	setNodeStatusFuncs []func(*v1.Node) error

	// TODO: think about moving this to be centralized in PodWorkers in follow-on.
	// the list of handlers to call during pod admission.
	admitHandlers lifecycle.PodAdmitHandlers

	// softAdmithandlers are applied to the pod after it is admitted by the Kubelet, but before it is
	// run. A pod rejected by a softAdmitHandler will be left in a Pending state indefinitely. If a
	// rejected pod should not be recreated, or the scheduler is not aware of the rejection rule, the
	// admission rule should be applied by a softAdmitHandler.
	softAdmitHandlers lifecycle.PodAdmitHandlers

	// the list of handlers to call during pod sync loop.
	lifecycle.PodSyncLoopHandlers

	// the list of handlers to call during pod sync.
	lifecycle.PodSyncHandlers

	// the number of allowed pods per core
	podsPerCore int

	// enableControllerAttachDetach indicates the Attach/Detach controller
	// should manage attachment/detachment of volumes scheduled to this node,
	// and disable kubelet from executing any attach/detach operations
	enableControllerAttachDetach bool

	// trigger deleting containers in a pod
	containerDeletor *podContainerDeletor

	// config iptables util rules
	makeIPTablesUtilChains bool

	// The bit of the fwmark space to mark packets for SNAT.
	iptablesMasqueradeBit int

	// The bit of the fwmark space to mark packets for dropping.
	iptablesDropBit int

	// The AppArmor validator for checking whether AppArmor is supported.
	appArmorValidator apparmor.Validator

	// The handler serving CRI streaming calls (exec/attach/port-forward).
	criHandler http.Handler

	// experimentalHostUserNamespaceDefaulting sets userns=true when users request host namespaces (pid, ipc, net),
	// are using non-namespaced capabilities (mknod, sys_time, sys_module), the pod contains a privileged container,
	// or using host path volumes.
	// This should only be enabled when the container runtime is performing user remapping AND if the
	// experimental behavior is desired.
	experimentalHostUserNamespaceDefaulting bool

	// dockerLegacyService contains some legacy methods for backward compatibility.
	// It should be set only when docker is using non json-file logging driver.
	dockerLegacyService dockershim.DockerLegacyService

	// StatsProvider provides the node and the container stats.
	*stats.StatsProvider

	// This flag, if set, instructs the kubelet to keep volumes from terminated pods mounted to the node.
	// This can be useful for debugging volume related issues.
	keepTerminatedPodVolumes bool // DEPRECATED
}

func allGlobalUnicastIPs() ([]net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("could not list network interfaces: %v", err)
	}
	var ips []net.IP
	for _, i := range interfaces {
		addresses, err := i.Addrs()
		if err != nil {
			return nil, fmt.Errorf("could not list the addresses for network interface %v: %v", i, err)
		}
		for _, address := range addresses {
			switch v := address.(type) {
			case *net.IPNet:
				if v.IP.IsGlobalUnicast() {
					ips = append(ips, v.IP)
				}
			}
		}
	}
	return ips, nil
}

// setupDataDirs creates:
// 1.  the root directory
// 2.  the pods directory
// 3.  the plugins directory
func (kl *Kubelet) setupDataDirs() error {
	kl.rootDirectory = path.Clean(kl.rootDirectory)
	if err := os.MkdirAll(kl.getRootDir(), 0750); err != nil {
		return fmt.Errorf("error creating root directory: %v", err)
	}
	if err := kl.mounter.MakeRShared(kl.getRootDir()); err != nil {
		return fmt.Errorf("error configuring root directory: %v", err)
	}
	if err := os.MkdirAll(kl.getPodsDir(), 0750); err != nil {
		return fmt.Errorf("error creating pods directory: %v", err)
	}
	if err := os.MkdirAll(kl.getPluginsDir(), 0750); err != nil {
		return fmt.Errorf("error creating plugins directory: %v", err)
	}
	return nil
}

// StartGarbageCollection starts garbage collection threads.
func (kl *Kubelet) StartGarbageCollection() {
	loggedContainerGCFailure := false
	go wait.Until(func() {
		if err := kl.containerGC.GarbageCollect(); err != nil {
			glog.Errorf("Container garbage collection failed: %v", err)
			kl.recorder.Eventf(kl.nodeRef, v1.EventTypeWarning, events.ContainerGCFailed, err.Error())
			loggedContainerGCFailure = true
		} else {
			var vLevel glog.Level = 4
			if loggedContainerGCFailure {
				vLevel = 1
				loggedContainerGCFailure = false
			}

			glog.V(vLevel).Infof("Container garbage collection succeeded")
		}
	}, ContainerGCPeriod, wait.NeverStop)

	stopChan := make(chan struct{})
	defer close(stopChan)
	// when the high threshold is set to 100, stub the image GC manager
	if kl.kubeletConfiguration.ImageGCHighThresholdPercent == 100 {
		glog.V(2).Infof("ImageGCHighThresholdPercent is set 100, Disable image GC")
		go func() { stopChan <- struct{}{} }()
	}

	prevImageGCFailed := false
	go wait.Until(func() {
		if err := kl.imageManager.GarbageCollect(); err != nil {
			if prevImageGCFailed {
				glog.Errorf("Image garbage collection failed multiple times in a row: %v", err)
				// Only create an event for repeated failures
				kl.recorder.Eventf(kl.nodeRef, v1.EventTypeWarning, events.ImageGCFailed, err.Error())
			} else {
				glog.Errorf("Image garbage collection failed once. Stats initialization may not have completed yet: %v", err)
			}
			prevImageGCFailed = true
		} else {
			var vLevel glog.Level = 4
			if prevImageGCFailed {
				vLevel = 1
				prevImageGCFailed = false
			}

			glog.V(vLevel).Infof("Image garbage collection succeeded")
		}
	}, ImageGCPeriod, stopChan)
}

// initializeModules will initialize internal modules that do not require the container runtime to be up.
// Note that the modules here must not depend on modules that are not initialized here.
func (kl *Kubelet) initializeModules() error {
	// Prometheus metrics.
	metrics.Register(kl.runtimeCache, collectors.NewVolumeStatsCollector(kl))

	// Setup filesystem directories.
	if err := kl.setupDataDirs(); err != nil {
		return err
	}

	// If the container logs directory does not exist, create it.
	if _, err := os.Stat(ContainerLogsDir); err != nil {
		if err := kl.os.MkdirAll(ContainerLogsDir, 0755); err != nil {
			glog.Errorf("Failed to create directory %q: %v", ContainerLogsDir, err)
		}
	}

	// Start the image manager.
	kl.imageManager.Start()

	// Start the certificate manager if it was enabled.
	if kl.serverCertificateManager != nil {
		kl.serverCertificateManager.Start()
	}

	// Start out of memory watcher.
	if err := kl.oomWatcher.Start(kl.nodeRef); err != nil {
		return fmt.Errorf("Failed to start OOM watcher %v", err)
	}

	// Start resource analyzer
	kl.resourceAnalyzer.Start()

	return nil
}

// initializeRuntimeDependentModules will initialize internal modules that require the container runtime to be up.
func (kl *Kubelet) initializeRuntimeDependentModules() {
	if err := kl.cadvisor.Start(); err != nil {
		// Fail kubelet and rely on the babysitter to retry starting kubelet.
		// TODO(random-liu): Add backoff logic in the babysitter
		glog.Fatalf("Failed to start cAdvisor %v", err)
	}

	// trigger on-demand stats collection once so that we have capacity information for ephemeral storage.
	// ignore any errors, since if stats collection is not successful, the container manager will fail to start below.
	kl.StatsProvider.GetCgroupStats("/", true)
	// Start container manager.
	node, err := kl.getNodeAnyWay()
	if err != nil {
		// Fail kubelet and rely on the babysitter to retry starting kubelet.
		glog.Fatalf("Kubelet failed to get node info: %v", err)
	}
	// containerManager must start after cAdvisor because it needs filesystem capacity information
	if err := kl.containerManager.Start(node, kl.GetActivePods, kl.sourcesReady, kl.statusManager, kl.runtimeService); err != nil {
		// Fail kubelet and rely on the babysitter to retry starting kubelet.
		glog.Fatalf("Failed to start ContainerManager %v", err)
	}
	// eviction manager must start after cadvisor because it needs to know if the container runtime has a dedicated imagefs
	kl.evictionManager.Start(kl.StatsProvider, kl.GetActivePods, kl.podResourcesAreReclaimed, evictionMonitoringPeriod)

	// container log manager must start after container runtime is up to retrieve information from container runtime
	// and inform container to reopen log file after log rotation.
	kl.containerLogManager.Start()
}

// Run starts the kubelet reacting to config updates
func (kl *Kubelet) Run(updates <-chan kubetypes.PodUpdate) {
	if kl.logServer == nil {
		kl.logServer = http.StripPrefix("/logs/", http.FileServer(http.Dir("/var/log/")))
	}
	if kl.kubeClient == nil {
		glog.Warning("No api server defined - no node status update will be sent.")
	}

	if err := kl.initializeModules(); err != nil {
		kl.recorder.Eventf(kl.nodeRef, v1.EventTypeWarning, events.KubeletSetupFailed, err.Error())
		glog.Fatal(err)
	}

	// Start volume manager
	go kl.volumeManager.Run(kl.sourcesReady, wait.NeverStop)

	if kl.kubeClient != nil {
		// Start syncing node status immediately, this may set up things the runtime needs to run.
		go wait.Until(kl.syncNodeStatus, kl.nodeStatusUpdateFrequency, wait.NeverStop)
	}
	go wait.Until(kl.updateRuntimeUp, 5*time.Second, wait.NeverStop)

	// Start loop to sync iptables util rules
	if kl.makeIPTablesUtilChains {
		go wait.Until(kl.syncNetworkUtil, 1*time.Minute, wait.NeverStop)
	}

	// Start a goroutine responsible for killing pods (that are not properly
	// handled by pod workers).
	go wait.Until(kl.podKiller, 1*time.Second, wait.NeverStop)

	// Start gorouting responsible for checking limits in resolv.conf
	if kl.dnsConfigurer.ResolverConfig != "" {
		go wait.Until(func() { kl.dnsConfigurer.CheckLimitsForResolvConf() }, 30*time.Second, wait.NeverStop)
	}

	// Start component sync loops.
	kl.statusManager.Start()
	kl.probeManager.Start()

	// Start the pod lifecycle event generator.
	kl.pleg.Start()
	kl.syncLoop(updates, kl)
}

// syncPod is the transaction script for the sync of a single pod.
//
// Arguments:
//
// o - the SyncPodOptions for this invocation
//
// The workflow is:
// * If the pod is being created, record pod worker start latency
// * Call generateAPIPodStatus to prepare an v1.PodStatus for the pod
// * If the pod is being seen as running for the first time, record pod
//   start latency
// * Update the status of the pod in the status manager
// * Kill the pod if it should not be running
// * Create a mirror pod if the pod is a static pod, and does not
//   already have a mirror pod
// * Create the data directories for the pod if they do not exist
// * Wait for volumes to attach/mount
// * Fetch the pull secrets for the pod
// * Call the container runtime's SyncPod callback
// * Update the traffic shaping for the pod's ingress and egress limits
//
// If any step of this workflow errors, the error is returned, and is repeated
// on the next syncPod call.
//
// This operation writes all events that are dispatched in order to provide
// the most accurate information possible about an error situation to aid debugging.
// Callers should not throw an event if this operation returns an error.
func (kl *Kubelet) syncPod(o syncPodOptions) error {
	// pull out the required options
	pod := o.pod
	mirrorPod := o.mirrorPod
	podStatus := o.podStatus
	updateType := o.updateType

	// if we want to kill a pod, do it now!
	if updateType == kubetypes.SyncPodKill {
		killPodOptions := o.killPodOptions
		if killPodOptions == nil || killPodOptions.PodStatusFunc == nil {
			return fmt.Errorf("kill pod options are required if update type is kill")
		}
		apiPodStatus := killPodOptions.PodStatusFunc(pod, podStatus)
		kl.statusManager.SetPodStatus(pod, apiPodStatus)
		// we kill the pod with the specified grace period since this is a termination
		if err := kl.killPod(pod, nil, podStatus, killPodOptions.PodTerminationGracePeriodSecondsOverride); err != nil {
			kl.recorder.Eventf(pod, v1.EventTypeWarning, events.FailedToKillPod, "error killing pod: %v", err)
			// there was an error killing the pod, so we return that error directly
			utilruntime.HandleError(err)
			return err
		}
		return nil
	}

	// Latency measurements for the main workflow are relative to the
	// first time the pod was seen by the API server.
	var firstSeenTime time.Time
	if firstSeenTimeStr, ok := pod.Annotations[kubetypes.ConfigFirstSeenAnnotationKey]; ok {
		firstSeenTime = kubetypes.ConvertToTimestamp(firstSeenTimeStr).Get()
	}

	// Record pod worker start latency if being created
	// TODO: make pod workers record their own latencies
	if updateType == kubetypes.SyncPodCreate {
		if !firstSeenTime.IsZero() {
			// This is the first time we are syncing the pod. Record the latency
			// since kubelet first saw the pod if firstSeenTime is set.
			metrics.PodWorkerStartLatency.Observe(metrics.SinceInMicroseconds(firstSeenTime))
		} else {
			glog.V(3).Infof("First seen time not recorded for pod %q", pod.UID)
		}
	}

	// Generate final API pod status with pod and status manager status
	apiPodStatus := kl.generateAPIPodStatus(pod, podStatus)
	// The pod IP may be changed in generateAPIPodStatus if the pod is using host network. (See #24576)
	// TODO(random-liu): After writing pod spec into container labels, check whether pod is using host network, and
	// set pod IP to hostIP directly in runtime.GetPodStatus
	podStatus.IP = apiPodStatus.PodIP

	// Record the time it takes for the pod to become running.
	existingStatus, ok := kl.statusManager.GetPodStatus(pod.UID)
	if !ok || existingStatus.Phase == v1.PodPending && apiPodStatus.Phase == v1.PodRunning &&
		!firstSeenTime.IsZero() {
		metrics.PodStartLatency.Observe(metrics.SinceInMicroseconds(firstSeenTime))
	}

	runnable := kl.canRunPod(pod)
	if !runnable.Admit {
		// Pod is not runnable; update the Pod and Container statuses to why.
		apiPodStatus.Reason = runnable.Reason
		apiPodStatus.Message = runnable.Message
		// Waiting containers are not creating.
		const waitingReason = "Blocked"
		for _, cs := range apiPodStatus.InitContainerStatuses {
			if cs.State.Waiting != nil {
				cs.State.Waiting.Reason = waitingReason
			}
		}
		for _, cs := range apiPodStatus.ContainerStatuses {
			if cs.State.Waiting != nil {
				cs.State.Waiting.Reason = waitingReason
			}
		}
	}

	// Update status in the status manager
	kl.statusManager.SetPodStatus(pod, apiPodStatus)

	// Kill pod if it should not be running
	if !runnable.Admit || pod.DeletionTimestamp != nil || apiPodStatus.Phase == v1.PodFailed {
		var syncErr error
		if err := kl.killPod(pod, nil, podStatus, nil); err != nil {
			kl.recorder.Eventf(pod, v1.EventTypeWarning, events.FailedToKillPod, "error killing pod: %v", err)
			syncErr = fmt.Errorf("error killing pod: %v", err)
			utilruntime.HandleError(syncErr)
		} else {
			if !runnable.Admit {
				// There was no error killing the pod, but the pod cannot be run.
				// Return an error to signal that the sync loop should back off.
				syncErr = fmt.Errorf("pod cannot be run: %s", runnable.Message)
			}
		}
		return syncErr
	}

	// If the network plugin is not ready, only start the pod if it uses the host network
	if rs := kl.runtimeState.networkErrors(); len(rs) != 0 && !kubecontainer.IsHostNetworkPod(pod) {
		kl.recorder.Eventf(pod, v1.EventTypeWarning, events.NetworkNotReady, "network is not ready: %v", rs)
		return fmt.Errorf("network is not ready: %v", rs)
	}

	// Create Cgroups for the pod and apply resource parameters
	// to them if cgroups-per-qos flag is enabled.
	pcm := kl.containerManager.NewPodContainerManager()
	// If pod has already been terminated then we need not create
	// or update the pod's cgroup
	if !kl.podIsTerminated(pod) {
		// When the kubelet is restarted with the cgroups-per-qos
		// flag enabled, all the pod's running containers
		// should be killed intermittently and brought back up
		// under the qos cgroup hierarchy.
		// Check if this is the pod's first sync
		firstSync := true
		for _, containerStatus := range apiPodStatus.ContainerStatuses {
			if containerStatus.State.Running != nil {
				firstSync = false
				break
			}
		}
		// Don't kill containers in pod if pod's cgroups already
		// exists or the pod is running for the first time
		podKilled := false
		if !pcm.Exists(pod) && !firstSync {
			if err := kl.killPod(pod, nil, podStatus, nil); err == nil {
				podKilled = true
			}
		}
		// Create and Update pod's Cgroups
		// Don't create cgroups for run once pod if it was killed above
		// The current policy is not to restart the run once pods when
		// the kubelet is restarted with the new flag as run once pods are
		// expected to run only once and if the kubelet is restarted then
		// they are not expected to run again.
		// We don't create and apply updates to cgroup if its a run once pod and was killed above
		if !(podKilled && pod.Spec.RestartPolicy == v1.RestartPolicyNever) {
			if !pcm.Exists(pod) {
				if err := kl.containerManager.UpdateQOSCgroups(); err != nil {
					glog.V(2).Infof("Failed to update QoS cgroups while syncing pod: %v", err)
				}
				if err := pcm.EnsureExists(pod); err != nil {
					kl.recorder.Eventf(pod, v1.EventTypeWarning, events.FailedToCreatePodContainer, "unable to ensure pod container exists: %v", err)
					return fmt.Errorf("failed to ensure that the pod: %v cgroups exist and are correctly applied: %v", pod.UID, err)
				}
			}
		}
	}

	// Create Mirror Pod for Static Pod if it doesn't already exist
	if kubepod.IsStaticPod(pod) {
		podFullName := kubecontainer.GetPodFullName(pod)
		deleted := false
		if mirrorPod != nil {
			if mirrorPod.DeletionTimestamp != nil || !kl.podManager.IsMirrorPodOf(mirrorPod, pod) {
				// The mirror pod is semantically different from the static pod. Remove
				// it. The mirror pod will get recreated later.
				glog.Warningf("Deleting mirror pod %q because it is outdated", format.Pod(mirrorPod))
				if err := kl.podManager.DeleteMirrorPod(podFullName); err != nil {
					glog.Errorf("Failed deleting mirror pod %q: %v", format.Pod(mirrorPod), err)
				} else {
					deleted = true
				}
			}
		}
		if mirrorPod == nil || deleted {
			node, err := kl.GetNode()
			if err != nil || node.DeletionTimestamp != nil {
				glog.V(4).Infof("No need to create a mirror pod, since node %q has been removed from the cluster", kl.nodeName)
			} else {
				glog.V(4).Infof("Creating a mirror pod for static pod %q", format.Pod(pod))
				if err := kl.podManager.CreateMirrorPod(pod); err != nil {
					glog.Errorf("Failed creating a mirror pod for %q: %v", format.Pod(pod), err)
				}
			}
		}
	}

	// Make data directories for the pod
	if err := kl.makePodDataDirs(pod); err != nil {
		kl.recorder.Eventf(pod, v1.EventTypeWarning, events.FailedToMakePodDataDirectories, "error making pod data directories: %v", err)
		glog.Errorf("Unable to make pod data directories for pod %q: %v", format.Pod(pod), err)
		return err
	}

	// Volume manager will not mount volumes for terminated pods
	if !kl.podIsTerminated(pod) {
		// Wait for volumes to attach/mount
		if err := kl.volumeManager.WaitForAttachAndMount(pod); err != nil {
			kl.recorder.Eventf(pod, v1.EventTypeWarning, events.FailedMountVolume, "Unable to mount volumes for pod %q: %v", format.Pod(pod), err)
			glog.Errorf("Unable to mount volumes for pod %q: %v; skipping pod", format.Pod(pod), err)
			return err
		}
	}

	// Fetch the pull secrets for the pod
	pullSecrets := kl.getPullSecretsForPod(pod)

	// Call the container runtime's SyncPod callback
	result := kl.containerRuntime.SyncPod(pod, apiPodStatus, podStatus, pullSecrets, kl.backOff)
	kl.reasonCache.Update(pod.UID, result)
	if err := result.Error(); err != nil {
		// Do not record an event here, as we keep all event logging for sync pod failures
		// local to container runtime so we get better errors
		return err
	}

	return nil
}

// Get pods which should be resynchronized. Currently, the following pod should be resynchronized:
//   * pod whose work is ready.
//   * internal modules that request sync of a pod.
func (kl *Kubelet) getPodsToSync() []*v1.Pod {
	allPods := kl.podManager.GetPods()
	podUIDs := kl.workQueue.GetWork()
	podUIDSet := sets.NewString()
	for _, podUID := range podUIDs {
		podUIDSet.Insert(string(podUID))
	}
	var podsToSync []*v1.Pod
	for _, pod := range allPods {
		if podUIDSet.Has(string(pod.UID)) {
			// The work of the pod is ready
			podsToSync = append(podsToSync, pod)
			continue
		}
		for _, podSyncLoopHandler := range kl.PodSyncLoopHandlers {
			if podSyncLoopHandler.ShouldSync(pod) {
				podsToSync = append(podsToSync, pod)
				break
			}
		}
	}
	return podsToSync
}

// deletePod deletes the pod from the internal state of the kubelet by:
// 1.  stopping the associated pod worker asynchronously
// 2.  signaling to kill the pod by sending on the podKillingCh channel
//
// deletePod returns an error if not all sources are ready or the pod is not
// found in the runtime cache.
func (kl *Kubelet) deletePod(pod *v1.Pod) error {
	if pod == nil {
		return fmt.Errorf("deletePod does not allow nil pod")
	}
	if !kl.sourcesReady.AllReady() {
		// If the sources aren't ready, skip deletion, as we may accidentally delete pods
		// for sources that haven't reported yet.
		return fmt.Errorf("skipping delete because sources aren't ready yet")
	}
	kl.podWorkers.ForgetWorker(pod.UID)

	// Runtime cache may not have been updated to with the pod, but it's okay
	// because the periodic cleanup routine will attempt to delete again later.
	runningPods, err := kl.runtimeCache.GetPods()
	if err != nil {
		return fmt.Errorf("error listing containers: %v", err)
	}
	runningPod := kubecontainer.Pods(runningPods).FindPod("", pod.UID)
	if runningPod.IsEmpty() {
		return fmt.Errorf("pod not found")
	}
	podPair := kubecontainer.PodPair{APIPod: pod, RunningPod: &runningPod}

	kl.podKillingCh <- &podPair
	// TODO: delete the mirror pod here?

	// We leave the volume/directory cleanup to the periodic cleanup routine.
	return nil
}

// rejectPod records an event about the pod with the given reason and message,
// and updates the pod to the failed phase in the status manage.
func (kl *Kubelet) rejectPod(pod *v1.Pod, reason, message string) {
	kl.recorder.Eventf(pod, v1.EventTypeWarning, reason, message)
	kl.statusManager.SetPodStatus(pod, v1.PodStatus{
		Phase:   v1.PodFailed,
		Reason:  reason,
		Message: "Pod " + message})
}

// canAdmitPod determines if a pod can be admitted, and gives a reason if it
// cannot. "pod" is new pod, while "pods" are all admitted pods
// The function returns a boolean value indicating whether the pod
// can be admitted, a brief single-word reason and a message explaining why
// the pod cannot be admitted.
func (kl *Kubelet) canAdmitPod(pods []*v1.Pod, pod *v1.Pod) (bool, string, string) {
	// the kubelet will invoke each pod admit handler in sequence
	// if any handler rejects, the pod is rejected.
	// TODO: move out of disk check into a pod admitter
	// TODO: out of resource eviction should have a pod admitter call-out
	attrs := &lifecycle.PodAdmitAttributes{Pod: pod, OtherPods: pods}
	for _, podAdmitHandler := range kl.admitHandlers {
		if result := podAdmitHandler.Admit(attrs); !result.Admit {
			return false, result.Reason, result.Message
		}
	}

	return true, "", ""
}

func (kl *Kubelet) canRunPod(pod *v1.Pod) lifecycle.PodAdmitResult {
	attrs := &lifecycle.PodAdmitAttributes{Pod: pod}
	// Get "OtherPods". Rejected pods are failed, so only include admitted pods that are alive.
	attrs.OtherPods = kl.filterOutTerminatedPods(kl.podManager.GetPods())

	for _, handler := range kl.softAdmitHandlers {
		if result := handler.Admit(attrs); !result.Admit {
			return result
		}
	}

	// TODO: Refactor as a soft admit handler.
	if err := canRunPod(pod); err != nil {
		return lifecycle.PodAdmitResult{
			Admit:   false,
			Reason:  "Forbidden",
			Message: err.Error(),
		}
	}

	return lifecycle.PodAdmitResult{Admit: true}
}

// syncLoop is the main loop for processing changes. It watches for changes from
// three channels (file, apiserver, and http) and creates a union of them. For
// any new change seen, will run a sync against desired state and running state. If
// no changes are seen to the configuration, will synchronize the last known desired
// state every sync-frequency seconds. Never returns.
func (kl *Kubelet) syncLoop(updates <-chan kubetypes.PodUpdate, handler SyncHandler) {
	glog.Info("Starting kubelet main sync loop.")
	// The resyncTicker wakes up kubelet to checks if there are any pod workers
	// that need to be sync'd. A one-second period is sufficient because the
	// sync interval is defaulted to 10s.
	syncTicker := time.NewTicker(time.Second)
	defer syncTicker.Stop()
	housekeepingTicker := time.NewTicker(housekeepingPeriod)
	defer housekeepingTicker.Stop()
	plegCh := kl.pleg.Watch()
	const (
		base   = 100 * time.Millisecond
		max    = 5 * time.Second
		factor = 2
	)
	duration := base
	for {
		if rs := kl.runtimeState.runtimeErrors(); len(rs) != 0 {
			glog.Infof("skipping pod synchronization - %v", rs)
			// exponential backoff
			time.Sleep(duration)
			duration = time.Duration(math.Min(float64(max), factor*float64(duration)))
			continue
		}
		// reset backoff if we have a success
		duration = base

		kl.syncLoopMonitor.Store(kl.clock.Now())
		if !kl.syncLoopIteration(updates, handler, syncTicker.C, housekeepingTicker.C, plegCh) {
			break
		}
		kl.syncLoopMonitor.Store(kl.clock.Now())
	}
}

// syncLoopIteration reads from various channels and dispatches pods to the
// given handler.
//
// Arguments:
// 1.  configCh:       a channel to read config events from
// 2.  handler:        the SyncHandler to dispatch pods to
// 3.  syncCh:         a channel to read periodic sync events from
// 4.  houseKeepingCh: a channel to read housekeeping events from
// 5.  plegCh:         a channel to read PLEG updates from
//
// Events are also read from the kubelet liveness manager's update channel.
//
// The workflow is to read from one of the channels, handle that event, and
// update the timestamp in the sync loop monitor.
//
// Here is an appropriate place to note that despite the syntactical
// similarity to the switch statement, the case statements in a select are
// evaluated in a pseudorandom order if there are multiple channels ready to
// read from when the select is evaluated.  In other words, case statements
// are evaluated in random order, and you can not assume that the case
// statements evaluate in order if multiple channels have events.
//
// With that in mind, in truly no particular order, the different channels
// are handled as follows:
//
// * configCh: dispatch the pods for the config change to the appropriate
//             handler callback for the event type
// * plegCh: update the runtime cache; sync pod
// * syncCh: sync all pods waiting for sync
// * houseKeepingCh: trigger cleanup of pods
// * liveness manager: sync pods that have failed or in which one or more
//                     containers have failed liveness checks
func (kl *Kubelet) syncLoopIteration(configCh <-chan kubetypes.PodUpdate, handler SyncHandler,
	syncCh <-chan time.Time, housekeepingCh <-chan time.Time, plegCh <-chan *pleg.PodLifecycleEvent) bool {
	select {
	case u, open := <-configCh:
		// Update from a config source; dispatch it to the right handler
		// callback.
		if !open {
			glog.Errorf("Update channel is closed. Exiting the sync loop.")
			return false
		}

		switch u.Op {
		case kubetypes.ADD:
			glog.V(2).Infof("SyncLoop (ADD, %q): %q", u.Source, format.Pods(u.Pods))
			// After restarting, kubelet will get all existing pods through
			// ADD as if they are new pods. These pods will then go through the
			// admission process and *may* be rejected. This can be resolved
			// once we have checkpointing.
			handler.HandlePodAdditions(u.Pods)
		case kubetypes.UPDATE:
			glog.V(2).Infof("SyncLoop (UPDATE, %q): %q", u.Source, format.PodsWithDeletiontimestamps(u.Pods))
			handler.HandlePodUpdates(u.Pods)
		case kubetypes.REMOVE:
			glog.V(2).Infof("SyncLoop (REMOVE, %q): %q", u.Source, format.Pods(u.Pods))
			handler.HandlePodRemoves(u.Pods)
		case kubetypes.RECONCILE:
			glog.V(4).Infof("SyncLoop (RECONCILE, %q): %q", u.Source, format.Pods(u.Pods))
			handler.HandlePodReconcile(u.Pods)
		case kubetypes.DELETE:
			glog.V(2).Infof("SyncLoop (DELETE, %q): %q", u.Source, format.Pods(u.Pods))
			// DELETE is treated as a UPDATE because of graceful deletion.
			handler.HandlePodUpdates(u.Pods)
		case kubetypes.RESTORE:
			glog.V(2).Infof("SyncLoop (RESTORE, %q): %q", u.Source, format.Pods(u.Pods))
			// These are pods restored from the checkpoint. Treat them as new
			// pods.
			handler.HandlePodAdditions(u.Pods)
		case kubetypes.SET:
			// TODO: Do we want to support this?
			glog.Errorf("Kubelet does not support snapshot update")
		}

		if u.Op != kubetypes.RESTORE {
			// If the update type is RESTORE, it means that the update is from
			// the pod checkpoints and may be incomplete. Do not mark the
			// source as ready.

			// Mark the source ready after receiving at least one update from the
			// source. Once all the sources are marked ready, various cleanup
			// routines will start reclaiming resources. It is important that this
			// takes place only after kubelet calls the update handler to process
			// the update to ensure the internal pod cache is up-to-date.
			kl.sourcesReady.AddSource(u.Source)
		}
	case e := <-plegCh:
		if isSyncPodWorthy(e) {
			// PLEG event for a pod; sync it.
			if pod, ok := kl.podManager.GetPodByUID(e.ID); ok {
				glog.V(2).Infof("SyncLoop (PLEG): %q, event: %#v", format.Pod(pod), e)
				handler.HandlePodSyncs([]*v1.Pod{pod})
			} else {
				// If the pod no longer exists, ignore the event.
				glog.V(4).Infof("SyncLoop (PLEG): ignore irrelevant event: %#v", e)
			}
		}

		if e.Type == pleg.ContainerDied {
			if containerID, ok := e.Data.(string); ok {
				kl.cleanUpContainersInPod(e.ID, containerID)
			}
		}
	case <-syncCh:
		// Sync pods waiting for sync
		podsToSync := kl.getPodsToSync()
		if len(podsToSync) == 0 {
			break
		}
		glog.V(4).Infof("SyncLoop (SYNC): %d pods; %s", len(podsToSync), format.Pods(podsToSync))
		handler.HandlePodSyncs(podsToSync)
	case update := <-kl.livenessManager.Updates():
		if update.Result == proberesults.Failure {
			// The liveness manager detected a failure; sync the pod.

			// We should not use the pod from livenessManager, because it is never updated after
			// initialization.
			pod, ok := kl.podManager.GetPodByUID(update.PodUID)
			if !ok {
				// If the pod no longer exists, ignore the update.
				glog.V(4).Infof("SyncLoop (container unhealthy): ignore irrelevant update: %#v", update)
				break
			}
			glog.V(1).Infof("SyncLoop (container unhealthy): %q", format.Pod(pod))
			handler.HandlePodSyncs([]*v1.Pod{pod})
		}
	case <-housekeepingCh:
		if !kl.sourcesReady.AllReady() {
			// If the sources aren't ready or volume manager has not yet synced the states,
			// skip housekeeping, as we may accidentally delete pods from unready sources.
			glog.V(4).Infof("SyncLoop (housekeeping, skipped): sources aren't ready yet.")
		} else {
			glog.V(4).Infof("SyncLoop (housekeeping)")
			if err := handler.HandlePodCleanups(); err != nil {
				glog.Errorf("Failed cleaning pods: %v", err)
			}
		}
	}
	return true
}

// dispatchWork starts the asynchronous sync of the pod in a pod worker.
// If the pod is terminated, dispatchWork
func (kl *Kubelet) dispatchWork(pod *v1.Pod, syncType kubetypes.SyncPodType, mirrorPod *v1.Pod, start time.Time) {
	if kl.podIsTerminated(pod) {
		if pod.DeletionTimestamp != nil {
			// If the pod is in a terminated state, there is no pod worker to
			// handle the work item. Check if the DeletionTimestamp has been
			// set, and force a status update to trigger a pod deletion request
			// to the apiserver.
			kl.statusManager.TerminatePod(pod)
		}
		return
	}
	// Run the sync in an async worker.
	kl.podWorkers.UpdatePod(&UpdatePodOptions{
		Pod:        pod,
		MirrorPod:  mirrorPod,
		UpdateType: syncType,
		OnCompleteFunc: func(err error) {
			if err != nil {
				metrics.PodWorkerLatency.WithLabelValues(syncType.String()).Observe(metrics.SinceInMicroseconds(start))
			}
		},
	})
	// Note the number of containers for new pods.
	if syncType == kubetypes.SyncPodCreate {
		metrics.ContainersPerPodCount.Observe(float64(len(pod.Spec.Containers)))
	}
}

// TODO: handle mirror pods in a separate component (issue #17251)
func (kl *Kubelet) handleMirrorPod(mirrorPod *v1.Pod, start time.Time) {
	// Mirror pod ADD/UPDATE/DELETE operations are considered an UPDATE to the
	// corresponding static pod. Send update to the pod worker if the static
	// pod exists.
	if pod, ok := kl.podManager.GetPodByMirrorPod(mirrorPod); ok {
		kl.dispatchWork(pod, kubetypes.SyncPodUpdate, mirrorPod, start)
	}
}

// HandlePodAdditions is the callback in SyncHandler for pods being added from
// a config source.
func (kl *Kubelet) HandlePodAdditions(pods []*v1.Pod) {
	start := kl.clock.Now()
	sort.Sort(sliceutils.PodsByCreationTime(pods))
	for _, pod := range pods {
		existingPods := kl.podManager.GetPods()
		// Always add the pod to the pod manager. Kubelet relies on the pod
		// manager as the source of truth for the desired state. If a pod does
		// not exist in the pod manager, it means that it has been deleted in
		// the apiserver and no action (other than cleanup) is required.
		kl.podManager.AddPod(pod)

		if kubepod.IsMirrorPod(pod) {
			kl.handleMirrorPod(pod, start)
			continue
		}

		if !kl.podIsTerminated(pod) {
			// Only go through the admission process if the pod is not
			// terminated.

			// We failed pods that we rejected, so activePods include all admitted
			// pods that are alive.
			activePods := kl.filterOutTerminatedPods(existingPods)

			// Check if we can admit the pod; if not, reject it.
			if ok, reason, message := kl.canAdmitPod(activePods, pod); !ok {
				kl.rejectPod(pod, reason, message)
				continue
			}
		}
		mirrorPod, _ := kl.podManager.GetMirrorPodByPod(pod)
		kl.dispatchWork(pod, kubetypes.SyncPodCreate, mirrorPod, start)
		kl.probeManager.AddPod(pod)
	}
}

// HandlePodUpdates is the callback in the SyncHandler interface for pods
// being updated from a config source.
func (kl *Kubelet) HandlePodUpdates(pods []*v1.Pod) {
	start := kl.clock.Now()
	for _, pod := range pods {
		kl.podManager.UpdatePod(pod)
		if kubepod.IsMirrorPod(pod) {
			kl.handleMirrorPod(pod, start)
			continue
		}
		// TODO: Evaluate if we need to validate and reject updates.

		mirrorPod, _ := kl.podManager.GetMirrorPodByPod(pod)
		kl.dispatchWork(pod, kubetypes.SyncPodUpdate, mirrorPod, start)
	}
}

// HandlePodRemoves is the callback in the SyncHandler interface for pods
// being removed from a config source.
func (kl *Kubelet) HandlePodRemoves(pods []*v1.Pod) {
	start := kl.clock.Now()
	for _, pod := range pods {
		kl.podManager.DeletePod(pod)
		if kubepod.IsMirrorPod(pod) {
			kl.handleMirrorPod(pod, start)
			continue
		}
		// Deletion is allowed to fail because the periodic cleanup routine
		// will trigger deletion again.
		if err := kl.deletePod(pod); err != nil {
			glog.V(2).Infof("Failed to delete pod %q, err: %v", format.Pod(pod), err)
		}
		kl.probeManager.RemovePod(pod)
	}
}

// HandlePodReconcile is the callback in the SyncHandler interface for pods
// that should be reconciled.
func (kl *Kubelet) HandlePodReconcile(pods []*v1.Pod) {
	for _, pod := range pods {
		// Update the pod in pod manager, status manager will do periodically reconcile according
		// to the pod manager.
		kl.podManager.UpdatePod(pod)

		// After an evicted pod is synced, all dead containers in the pod can be removed.
		if eviction.PodIsEvicted(pod.Status) {
			if podStatus, err := kl.podCache.Get(pod.UID); err == nil {
				kl.containerDeletor.deleteContainersInPod("", podStatus, true)
			}
		}
	}
}

// HandlePodSyncs is the callback in the syncHandler interface for pods
// that should be dispatched to pod workers for sync.
func (kl *Kubelet) HandlePodSyncs(pods []*v1.Pod) {
	start := kl.clock.Now()
	for _, pod := range pods {
		mirrorPod, _ := kl.podManager.GetMirrorPodByPod(pod)
		kl.dispatchWork(pod, kubetypes.SyncPodSync, mirrorPod, start)
	}
}

// LatestLoopEntryTime returns the last time in the sync loop monitor.
func (kl *Kubelet) LatestLoopEntryTime() time.Time {
	val := kl.syncLoopMonitor.Load()
	if val == nil {
		return time.Time{}
	}
	return val.(time.Time)
}

// updateRuntimeUp calls the container runtime status callback, initializing
// the runtime dependent modules when the container runtime first comes up,
// and returns an error if the status check fails.  If the status check is OK,
// update the container runtime uptime in the kubelet runtimeState.
func (kl *Kubelet) updateRuntimeUp() {
	s, err := kl.containerRuntime.Status()
	if err != nil {
		glog.Errorf("Container runtime sanity check failed: %v", err)
		return
	}
	if s == nil {
		glog.Errorf("Container runtime status is nil")
		return
	}
	// Periodically log the whole runtime status for debugging.
	// TODO(random-liu): Consider to send node event when optional
	// condition is unmet.
	glog.V(4).Infof("Container runtime status: %v", s)
	networkReady := s.GetRuntimeCondition(kubecontainer.NetworkReady)
	if networkReady == nil || !networkReady.Status {
		glog.Errorf("Container runtime network not ready: %v", networkReady)
		kl.runtimeState.setNetworkState(fmt.Errorf("runtime network not ready: %v", networkReady))
	} else {
		// Set nil if the container runtime network is ready.
		kl.runtimeState.setNetworkState(nil)
	}
	// TODO(random-liu): Add runtime error in runtimeState, and update it
	// when runtime is not ready, so that the information in RuntimeReady
	// condition will be propagated to NodeReady condition.
	runtimeReady := s.GetRuntimeCondition(kubecontainer.RuntimeReady)
	// If RuntimeReady is not set or is false, report an error.
	if runtimeReady == nil || !runtimeReady.Status {
		glog.Errorf("Container runtime not ready: %v", runtimeReady)
		return
	}
	kl.oneTimeInitializer.Do(kl.initializeRuntimeDependentModules)
	kl.runtimeState.setRuntimeSync(kl.clock.Now())
}

// GetConfiguration returns the KubeletConfiguration used to configure the kubelet.
func (kl *Kubelet) GetConfiguration() kubeletconfiginternal.KubeletConfiguration {
	return kl.kubeletConfiguration
}

// BirthCry sends an event that the kubelet has started up.
func (kl *Kubelet) BirthCry() {
	// Make an event that kubelet restarted.
	kl.recorder.Eventf(kl.nodeRef, v1.EventTypeNormal, events.StartingKubelet, "Starting kubelet.")
}

// StreamingConnectionIdleTimeout returns the timeout for streaming connections to the HTTP server.
func (kl *Kubelet) StreamingConnectionIdleTimeout() time.Duration {
	return kl.streamingConnectionIdleTimeout
}

// ResyncInterval returns the interval used for periodic syncs.
func (kl *Kubelet) ResyncInterval() time.Duration {
	return kl.resyncInterval
}

// ListenAndServe runs the kubelet HTTP server.
func (kl *Kubelet) ListenAndServe(address net.IP, port uint, tlsOptions *server.TLSOptions, auth server.AuthInterface, enableDebuggingHandlers, enableContentionProfiling bool) {
	server.ListenAndServeKubeletServer(kl, kl.resourceAnalyzer, address, port, tlsOptions, auth, enableDebuggingHandlers, enableContentionProfiling, kl.containerRuntime, kl.criHandler)
}

// ListenAndServeReadOnly runs the kubelet HTTP server in read-only mode.
func (kl *Kubelet) ListenAndServeReadOnly(address net.IP, port uint) {
	server.ListenAndServeKubeletReadOnlyServer(kl, kl.resourceAnalyzer, address, port, kl.containerRuntime)
}

// Delete the eligible dead container instances in a pod. Depending on the configuration, the latest dead containers may be kept around.
func (kl *Kubelet) cleanUpContainersInPod(podID types.UID, exitedContainerID string) {
	if podStatus, err := kl.podCache.Get(podID); err == nil {
		removeAll := false
		if syncedPod, ok := kl.podManager.GetPodByUID(podID); ok {
			// generate the api status using the cached runtime status to get up-to-date ContainerStatuses
			apiPodStatus := kl.generateAPIPodStatus(syncedPod, podStatus)
			// When an evicted or deleted pod has already synced, all containers can be removed.
			removeAll = eviction.PodIsEvicted(syncedPod.Status) || (syncedPod.DeletionTimestamp != nil && notRunning(apiPodStatus.ContainerStatuses))
		}
		kl.containerDeletor.deleteContainersInPod(exitedContainerID, podStatus, removeAll)
	}
}

// isSyncPodWorthy filters out events that are not worthy of pod syncing
func isSyncPodWorthy(event *pleg.PodLifecycleEvent) bool {
	// ContatnerRemoved doesn't affect pod state
	return event.Type != pleg.ContainerRemoved
}

// Gets the streaming server configuration to use with in-process CRI shims.
func getStreamingConfig(kubeCfg *kubeletconfiginternal.KubeletConfiguration, kubeDeps *Dependencies) *streaming.Config {
	config := &streaming.Config{
		// Use a relative redirect (no scheme or host).
		BaseURL: &url.URL{
			Path: "/cri/",
		},
		StreamIdleTimeout:               kubeCfg.StreamingConnectionIdleTimeout.Duration,
		StreamCreationTimeout:           streaming.DefaultConfig.StreamCreationTimeout,
		SupportedRemoteCommandProtocols: streaming.DefaultConfig.SupportedRemoteCommandProtocols,
		SupportedPortForwardProtocols:   streaming.DefaultConfig.SupportedPortForwardProtocols,
	}
	if kubeDeps.TLSOptions != nil {
		config.TLSConfig = kubeDeps.TLSOptions.Config
	}
	return config
}
