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

package options

import (
	"net"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	apiserveroptions "k8s.io/apiserver/pkg/server/options"
	cmoptions "k8s.io/kubernetes/cmd/controller-manager/app/options"
	"k8s.io/kubernetes/pkg/apis/componentconfig"
)

func TestAddFlags(t *testing.T) {
	f := pflag.NewFlagSet("addflagstest", pflag.ContinueOnError)
	s := NewKubeControllerManagerOptions()
	s.AddFlags(f, []string{""}, []string{""})

	args := []string{
		"--address=192.168.4.10",
		"--allocate-node-cidrs=true",
		"--attach-detach-reconcile-sync-period=30s",
		"--cidr-allocator-type=CloudAllocator",
		"--cloud-config=/cloud-config",
		"--cloud-provider=gce",
		"--cluster-cidr=1.2.3.4/24",
		"--cluster-name=k8s",
		"--cluster-signing-cert-file=/cluster-signing-cert",
		"--cluster-signing-key-file=/cluster-signing-key",
		"--concurrent-deployment-syncs=10",
		"--concurrent-endpoint-syncs=10",
		"--concurrent-gc-syncs=30",
		"--concurrent-namespace-syncs=20",
		"--concurrent-replicaset-syncs=10",
		"--concurrent-resource-quota-syncs=10",
		"--concurrent-service-syncs=2",
		"--concurrent-serviceaccount-token-syncs=10",
		"--concurrent_rc_syncs=10",
		"--configure-cloud-routes=false",
		"--contention-profiling=true",
		"--controller-start-interval=2m",
		"--controllers=foo,bar",
		"--deployment-controller-sync-period=45s",
		"--disable-attach-detach-reconcile-sync=true",
		"--enable-dynamic-provisioning=false",
		"--enable-garbage-collector=false",
		"--enable-hostpath-provisioner=true",
		"--enable-taint-manager=false",
		"--experimental-cluster-signing-duration=10h",
		"--flex-volume-plugin-dir=/flex-volume-plugin",
		"--horizontal-pod-autoscaler-downscale-delay=2m",
		"--horizontal-pod-autoscaler-sync-period=45s",
		"--horizontal-pod-autoscaler-upscale-delay=1m",
		"--http2-max-streams-per-connection=47",
		"--kube-api-burst=100",
		"--kube-api-content-type=application/json",
		"--kube-api-qps=50.0",
		"--kubeconfig=/kubeconfig",
		"--large-cluster-size-threshold=100",
		"--leader-elect=false",
		"--leader-elect-lease-duration=30s",
		"--leader-elect-renew-deadline=15s",
		"--leader-elect-resource-lock=configmap",
		"--leader-elect-retry-period=5s",
		"--master=192.168.4.20",
		"--min-resync-period=8h",
		"--namespace-sync-period=10m",
		"--node-cidr-mask-size=48",
		"--node-eviction-rate=0.2",
		"--node-monitor-grace-period=30s",
		"--node-monitor-period=10s",
		"--node-startup-grace-period=30s",
		"--pod-eviction-timeout=2m",
		"--port=10000",
		"--profiling=false",
		"--pv-recycler-increment-timeout-nfs=45",
		"--pv-recycler-minimum-timeout-hostpath=45",
		"--pv-recycler-minimum-timeout-nfs=200",
		"--pv-recycler-timeout-increment-hostpath=45",
		"--pvclaimbinder-sync-period=30s",
		"--resource-quota-sync-period=10m",
		"--route-reconciliation-period=30s",
		"--secondary-node-eviction-rate=0.05",
		"--service-account-private-key-file=/service-account-private-key",
		"--terminated-pod-gc-threshold=12000",
		"--unhealthy-zone-threshold=0.6",
		"--use-service-account-credentials=true",
		"--cert-dir=/a/b/c",
		"--bind-address=192.168.4.21",
		"--secure-port=10001",
	}
	f.Parse(args)
	// Sort GCIgnoredResources because it's built from a map, which means the
	// insertion order is random.
	sort.Sort(sortedGCIgnoredResources(s.Generic.GarbageCollectorController.GCIgnoredResources))

	expected := &KubeControllerManagerOptions{
		Generic: &cmoptions.GenericControllerManagerOptions{
			CloudProvider: &cmoptions.CloudProviderOptions{
				Name:            "gce",
				CloudConfigFile: "/cloud-config",
			},
			Debugging: &cmoptions.DebuggingOptions{
				EnableProfiling:           false,
				EnableContentionProfiling: true,
			},
			GenericComponent: &cmoptions.GenericComponentConfigOptions{
				MinResyncPeriod:         metav1.Duration{Duration: 8 * time.Hour},
				ContentType:             "application/json",
				KubeAPIQPS:              50.0,
				KubeAPIBurst:            100,
				ControllerStartInterval: metav1.Duration{Duration: 2 * time.Minute},
				LeaderElection: componentconfig.LeaderElectionConfiguration{
					ResourceLock:  "configmap",
					LeaderElect:   false,
					LeaseDuration: metav1.Duration{Duration: 30 * time.Second},
					RenewDeadline: metav1.Duration{Duration: 15 * time.Second},
					RetryPeriod:   metav1.Duration{Duration: 5 * time.Second},
				},
			},
			KubeCloudShared: &cmoptions.KubeCloudSharedOptions{
				Port:    10252,     // Note: InsecureServingOptions.ApplyTo will write the flag value back into the component config		 +				AllocateNodeCIDRs:         true,
				Address: "0.0.0.0", // Note: InsecureServingOptions.ApplyTo will write the flag value back into the component config
				UseServiceAccountCredentials: true,
				RouteReconciliationPeriod:    metav1.Duration{Duration: 30 * time.Second},
				NodeMonitorPeriod:            metav1.Duration{Duration: 10 * time.Second},
				ClusterName:                  "k8s",
				ClusterCIDR:                  "1.2.3.4/24",
				AllocateNodeCIDRs:            true,
				CIDRAllocatorType:            "CloudAllocator",
				ConfigureCloudRoutes:         false,
				ServiceAccountKeyFile:        "/service-account-private-key",
			},
			AttachDetachController: &cmoptions.AttachDetachControllerOptions{
				ReconcilerSyncLoopPeriod:          metav1.Duration{Duration: 30 * time.Second},
				DisableAttachDetachReconcilerSync: true,
			},
			CSRSigningController: &cmoptions.CSRSigningControllerOptions{
				ClusterSigningCertFile: "/cluster-signing-cert",
				ClusterSigningKeyFile:  "/cluster-signing-key",
				ClusterSigningDuration: metav1.Duration{Duration: 10 * time.Hour},
			},
			DaemonSetController: &cmoptions.DaemonSetControllerOptions{
				ConcurrentDaemonSetSyncs: 2,
			},
			DeploymentController: &cmoptions.DeploymentControllerOptions{
				ConcurrentDeploymentSyncs:      10,
				DeploymentControllerSyncPeriod: metav1.Duration{Duration: 45 * time.Second},
			},
			DeprecatedFlags: &cmoptions.DeprecatedControllerOptions{
				DeletingPodsQPS:    0.1,
				RegisterRetryCount: 10,
			},
			EndPointController: &cmoptions.EndPointControllerOptions{
				ConcurrentEndpointSyncs: 10,
			},
			GarbageCollectorController: &cmoptions.GarbageCollectorControllerOptions{
				ConcurrentGCSyncs: 30,
				GCIgnoredResources: []componentconfig.GroupResource{
					{Group: "extensions", Resource: "replicationcontrollers"},
					{Group: "", Resource: "bindings"},
					{Group: "", Resource: "componentstatuses"},
					{Group: "", Resource: "events"},
					{Group: "authentication.k8s.io", Resource: "tokenreviews"},
					{Group: "authorization.k8s.io", Resource: "subjectaccessreviews"},
					{Group: "authorization.k8s.io", Resource: "selfsubjectaccessreviews"},
					{Group: "authorization.k8s.io", Resource: "localsubjectaccessreviews"},
					{Group: "authorization.k8s.io", Resource: "selfsubjectrulesreviews"},
					{Group: "apiregistration.k8s.io", Resource: "apiservices"},
					{Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions"},
				},
				EnableGarbageCollector: false,
			},
			HPAController: &cmoptions.HPAControllerOptions{
				HorizontalPodAutoscalerSyncPeriod:               metav1.Duration{Duration: 45 * time.Second},
				HorizontalPodAutoscalerUpscaleForbiddenWindow:   metav1.Duration{Duration: 1 * time.Minute},
				HorizontalPodAutoscalerDownscaleForbiddenWindow: metav1.Duration{Duration: 2 * time.Minute},
				HorizontalPodAutoscalerTolerance:                0.1,
				HorizontalPodAutoscalerUseRESTClients:           true,
			},
			JobController: &cmoptions.JobControllerOptions{
				ConcurrentJobSyncs: 5,
			},
			NamespaceController: &cmoptions.NamespaceControllerOptions{
				NamespaceSyncPeriod:      metav1.Duration{Duration: 10 * time.Minute},
				ConcurrentNamespaceSyncs: 20,
			},
			NodeIpamController: &cmoptions.NodeIpamControllerOptions{
				NodeCIDRMaskSize: 48,
			},
			NodeLifecycleController: &cmoptions.NodeLifecycleControllerOptions{
				EnableTaintManager:        false,
				NodeEvictionRate:          0.2,
				SecondaryNodeEvictionRate: 0.05,
				NodeMonitorGracePeriod:    metav1.Duration{Duration: 30 * time.Second},
				NodeStartupGracePeriod:    metav1.Duration{Duration: 30 * time.Second},
				PodEvictionTimeout:        metav1.Duration{Duration: 2 * time.Minute},
				LargeClusterSizeThreshold: 100,
				UnhealthyZoneThreshold:    0.6,
			},
			PersistentVolumeBinderController: &cmoptions.PersistentVolumeBinderControllerOptions{
				PVClaimBinderSyncPeriod: metav1.Duration{Duration: 30 * time.Second},
				VolumeConfiguration: componentconfig.VolumeConfiguration{
					EnableDynamicProvisioning:  false,
					EnableHostPathProvisioning: true,
					FlexVolumePluginDir:        "/flex-volume-plugin",
					PersistentVolumeRecyclerConfiguration: componentconfig.PersistentVolumeRecyclerConfiguration{
						MaximumRetry:             3,
						MinimumTimeoutNFS:        200,
						IncrementTimeoutNFS:      45,
						MinimumTimeoutHostPath:   45,
						IncrementTimeoutHostPath: 45,
					},
				},
			},
			PodGCController: &cmoptions.PodGCControllerOptions{
				TerminatedPodGCThreshold: 12000,
			},
			ReplicaSetController: &cmoptions.ReplicaSetControllerOptions{
				ConcurrentRSSyncs: 10,
			},
			ReplicationController: &cmoptions.ReplicationControllerOptions{
				ConcurrentRCSyncs: 10,
			},
			ResourceQuotaController: &cmoptions.ResourceQuotaControllerOptions{
				ResourceQuotaSyncPeriod:      metav1.Duration{Duration: 10 * time.Minute},
				ConcurrentResourceQuotaSyncs: 10,
			},
			SAController: &cmoptions.SAControllerOptions{
				ConcurrentSATokenSyncs: 10,
			},
			ServiceController: &cmoptions.ServiceControllerOptions{
				ConcurrentServiceSyncs: 2,
			},
			Controllers: []string{"foo", "bar"},
			SecureServing: &apiserveroptions.SecureServingOptions{
				BindPort:    10001,
				BindAddress: net.ParseIP("192.168.4.21"),
				ServerCert: apiserveroptions.GeneratableKeyCert{
					CertDirectory: "/a/b/c",
					PairName:      "kube-controller-manager",
				},
				HTTP2MaxStreamsPerConnection: 47,
			},
			InsecureServing: &cmoptions.InsecureServingOptions{
				BindAddress: net.ParseIP("192.168.4.10"),
				BindPort:    int(10000),
				BindNetwork: "tcp",
			},
			Kubeconfig: "/kubeconfig",
			Master:     "192.168.4.20",
		},
	}

	// Sort GCIgnoredResources because it's built from a map, which means the
	// insertion order is random.
	sort.Sort(sortedGCIgnoredResources(expected.Generic.GarbageCollectorController.GCIgnoredResources))

	if !reflect.DeepEqual(expected, s) {
		t.Errorf("Got different run options than expected.\nDifference detected on:\n%s", diff.ObjectReflectDiff(expected, s))
	}
}

type sortedGCIgnoredResources []componentconfig.GroupResource

func (r sortedGCIgnoredResources) Len() int {
	return len(r)
}

func (r sortedGCIgnoredResources) Less(i, j int) bool {
	if r[i].Group < r[j].Group {
		return true
	} else if r[i].Group > r[j].Group {
		return false
	}
	return r[i].Resource < r[j].Resource
}

func (r sortedGCIgnoredResources) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
