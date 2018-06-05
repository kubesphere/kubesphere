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

package e2e_node

import (
	"os/exec"
	"strconv"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletmetrics "k8s.io/kubernetes/pkg/kubelet/metrics"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/metrics"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/model"
)

const (
	testPodNamePrefix = "nvidia-gpu-"
)

// Serial because the test restarts Kubelet
var _ = framework.KubeDescribe("NVIDIA GPU Device Plugin [Feature:GPUDevicePlugin] [Serial] [Disruptive]", func() {
	f := framework.NewDefaultFramework("device-plugin-gpus-errors")

	Context("DevicePlugin", func() {
		var devicePluginPod *v1.Pod
		BeforeEach(func() {
			By("Ensuring that Nvidia GPUs exists on the node")
			if !checkIfNvidiaGPUsExistOnNode() {
				Skip("Nvidia GPUs do not exist on the node. Skipping test.")
			}

			By("Creating the Google Device Plugin pod for NVIDIA GPU in GKE")
			devicePluginPod = f.PodClient().CreateSync(framework.NVIDIADevicePlugin(f.Namespace.Name))

			By("Waiting for GPUs to become available on the local node")
			Eventually(func() bool {
				return framework.NumberOfNVIDIAGPUs(getLocalNode(f)) > 0
			}, 10*time.Second, framework.Poll).Should(BeTrue())

			if framework.NumberOfNVIDIAGPUs(getLocalNode(f)) < 2 {
				Skip("Not enough GPUs to execute this test (at least two needed)")
			}
		})

		AfterEach(func() {
			l, err := f.PodClient().List(metav1.ListOptions{})
			framework.ExpectNoError(err)

			for _, p := range l.Items {
				if p.Namespace != f.Namespace.Name {
					continue
				}

				f.PodClient().Delete(p.Name, &metav1.DeleteOptions{})
			}
		})

		It("checks that when Kubelet restarts exclusive GPU assignation to pods is kept.", func() {
			By("Creating one GPU pod on a node with at least two GPUs")
			podRECMD := "devs=$(ls /dev/ | egrep '^nvidia[0-9]+$') && echo gpu devices: $devs"
			p1 := f.PodClient().CreateSync(makeBusyboxPod(framework.NVIDIAGPUResourceName, podRECMD))

			deviceIDRE := "gpu devices: (nvidia[0-9]+)"
			devId1 := parseLog(f, p1.Name, p1.Name, deviceIDRE)
			p1, err := f.PodClient().Get(p1.Name, metav1.GetOptions{})
			framework.ExpectNoError(err)

			By("Restarting Kubelet and waiting for the current running pod to restart")
			restartKubelet()

			By("Confirming that after a kubelet and pod restart, GPU assignement is kept")
			ensurePodContainerRestart(f, p1.Name, p1.Name)
			devIdRestart1 := parseLog(f, p1.Name, p1.Name, deviceIDRE)
			Expect(devIdRestart1).To(Equal(devId1))

			By("Restarting Kubelet and creating another pod")
			restartKubelet()
			framework.WaitForAllNodesSchedulable(f.ClientSet, framework.TestContext.NodeSchedulableTimeout)
			Eventually(func() bool {
				return framework.NumberOfNVIDIAGPUs(getLocalNode(f)) > 0
			}, 10*time.Second, framework.Poll).Should(BeTrue())
			p2 := f.PodClient().CreateSync(makeBusyboxPod(framework.NVIDIAGPUResourceName, podRECMD))

			By("Checking that pods got a different GPU")
			devId2 := parseLog(f, p2.Name, p2.Name, deviceIDRE)

			Expect(devId1).To(Not(Equal(devId2)))

			By("Deleting device plugin.")
			f.PodClient().Delete(devicePluginPod.Name, &metav1.DeleteOptions{})
			By("Waiting for GPUs to become unavailable on the local node")
			Eventually(func() bool {
				node, err := f.ClientSet.CoreV1().Nodes().Get(framework.TestContext.NodeName, metav1.GetOptions{})
				framework.ExpectNoError(err)
				return framework.NumberOfNVIDIAGPUs(node) <= 0
			}, 10*time.Minute, framework.Poll).Should(BeTrue())
			By("Checking that scheduled pods can continue to run even after we delete device plugin.")
			ensurePodContainerRestart(f, p1.Name, p1.Name)
			devIdRestart1 = parseLog(f, p1.Name, p1.Name, deviceIDRE)
			Expect(devIdRestart1).To(Equal(devId1))

			ensurePodContainerRestart(f, p2.Name, p2.Name)
			devIdRestart2 := parseLog(f, p2.Name, p2.Name, deviceIDRE)
			Expect(devIdRestart2).To(Equal(devId2))
			By("Restarting Kubelet.")
			restartKubelet()
			By("Checking that scheduled pods can continue to run even after we delete device plugin and restart Kubelet.")
			ensurePodContainerRestart(f, p1.Name, p1.Name)
			devIdRestart1 = parseLog(f, p1.Name, p1.Name, deviceIDRE)
			Expect(devIdRestart1).To(Equal(devId1))
			ensurePodContainerRestart(f, p2.Name, p2.Name)
			devIdRestart2 = parseLog(f, p2.Name, p2.Name, deviceIDRE)
			Expect(devIdRestart2).To(Equal(devId2))
			logDevicePluginMetrics()

			// Cleanup
			f.PodClient().DeleteSync(p1.Name, &metav1.DeleteOptions{}, framework.DefaultPodDeletionTimeout)
			f.PodClient().DeleteSync(p2.Name, &metav1.DeleteOptions{}, framework.DefaultPodDeletionTimeout)
		})
	})
})

func checkIfNvidiaGPUsExistOnNode() bool {
	// Cannot use `lspci` because it is not installed on all distros by default.
	err := exec.Command("/bin/sh", "-c", "find /sys/devices/pci* -type f | grep vendor | xargs cat | grep 0x10de").Run()
	if err != nil {
		framework.Logf("check for nvidia GPUs failed. Got Error: %v", err)
		return false
	}
	return true
}

func logDevicePluginMetrics() {
	ms, err := metrics.GrabKubeletMetricsWithoutProxy(framework.TestContext.NodeName + ":10255")
	framework.ExpectNoError(err)
	for msKey, samples := range ms {
		switch msKey {
		case kubeletmetrics.KubeletSubsystem + "_" + kubeletmetrics.DevicePluginAllocationLatencyKey:
			for _, sample := range samples {
				latency := sample.Value
				resource := string(sample.Metric["resource_name"])
				var quantile float64
				if val, ok := sample.Metric[model.QuantileLabel]; ok {
					var err error
					if quantile, err = strconv.ParseFloat(string(val), 64); err != nil {
						continue
					}
					framework.Logf("Metric: %v ResourceName: %v Quantile: %v Latency: %v", msKey, resource, quantile, latency)
				}
			}
		case kubeletmetrics.KubeletSubsystem + "_" + kubeletmetrics.DevicePluginRegistrationCountKey:
			for _, sample := range samples {
				resource := string(sample.Metric["resource_name"])
				count := sample.Value
				framework.Logf("Metric: %v ResourceName: %v Count: %v", msKey, resource, count)
			}
		}
	}
}
