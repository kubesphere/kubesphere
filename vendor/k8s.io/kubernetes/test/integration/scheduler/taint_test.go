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

package scheduler

// This file tests the Taint feature.

import (
	"reflect"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	internalinformers "k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion"
	"k8s.io/kubernetes/pkg/controller/nodelifecycle"
	kubeadmission "k8s.io/kubernetes/pkg/kubeapiserver/admission"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
	"k8s.io/kubernetes/plugin/pkg/admission/podtolerationrestriction"
	pluginapi "k8s.io/kubernetes/plugin/pkg/admission/podtolerationrestriction/apis/podtolerationrestriction"
)

// TestTaintNodeByCondition verifies:
//   1. MemoryPressure Toleration is added to non-BestEffort Pod by PodTolerationRestriction
//   2. NodeController taints nodes by node condition
//   3. Scheduler allows pod to tolerate node condition taints, e.g. network unavailable
func TestTaintNodeByCondition(t *testing.T) {
	enabled := utilfeature.DefaultFeatureGate.Enabled("TaintNodesByCondition")
	defer func() {
		if !enabled {
			utilfeature.DefaultFeatureGate.Set("TaintNodesByCondition=False")
		}
	}()
	// Enable TaintNodeByCondition
	utilfeature.DefaultFeatureGate.Set("TaintNodesByCondition=True")

	// Build PodToleration Admission.
	admission := podtolerationrestriction.NewPodTolerationsPlugin(&pluginapi.Configuration{})

	context := initTestMaster(t, "default", admission)

	// Build clientset and informers for controllers.
	internalClientset := internalclientset.NewForConfigOrDie(&restclient.Config{
		QPS:           -1,
		Host:          context.httpServer.URL,
		ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})
	internalInformers := internalinformers.NewSharedInformerFactory(internalClientset, time.Second)

	kubeadmission.WantsInternalKubeClientSet(admission).SetInternalKubeClientSet(internalClientset)
	kubeadmission.WantsInternalKubeInformerFactory(admission).SetInternalKubeInformerFactory(internalInformers)

	controllerCh := make(chan struct{})
	defer close(controllerCh)

	// Apply feature gates to enable TaintNodesByCondition
	algorithmprovider.ApplyFeatureGates()

	context = initTestScheduler(t, context, controllerCh, false, nil)
	clientset := context.clientSet
	informers := context.informerFactory
	nsName := context.ns.Name

	// Start NodeLifecycleController for taint.
	nc, err := nodelifecycle.NewNodeLifecycleController(
		informers.Core().V1().Pods(),
		informers.Core().V1().Nodes(),
		informers.Extensions().V1beta1().DaemonSets(),
		nil, // CloudProvider
		clientset,
		time.Second, // Node monitor grace period
		time.Second, // Node startup grace period
		time.Second, // Node monitor period
		time.Second, // Pod eviction timeout
		100,         // Eviction limiter QPS
		100,         // Secondary eviction limiter QPS
		100,         // Large cluster threshold
		100,         // Unhealthy zone threshold
		true,        // Run taint manager
		true,        // Use taint based evictions
		true,        // Enabled TaintNodeByCondition feature
	)
	if err != nil {
		t.Errorf("Failed to create node controller: %v", err)
		return
	}
	go nc.Run(controllerCh)

	// Waiting for all controller sync.
	internalInformers.Start(controllerCh)
	internalInformers.WaitForCacheSync(controllerCh)

	// -------------------------------------------
	// Test TaintNodeByCondition feature.
	// -------------------------------------------
	memoryPressureToleration := v1.Toleration{
		Key:      algorithm.TaintNodeMemoryPressure,
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}

	// Case 1: Add MememoryPressure Toleration for non-BestEffort pod.
	burstablePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "burstable-pod",
			Namespace: nsName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "busybox",
					Image: "busybox",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("100m"),
						},
					},
				},
			},
		},
	}

	burstablePodInServ, err := clientset.CoreV1().Pods(nsName).Create(burstablePod)
	if err != nil {
		t.Errorf("Case 1: Failed to create pod: %v", err)
	} else if !reflect.DeepEqual(burstablePodInServ.Spec.Tolerations, []v1.Toleration{memoryPressureToleration}) {
		t.Errorf("Case 1: Unexpected toleration of non-BestEffort pod, expected: %+v, got: %v",
			[]v1.Toleration{memoryPressureToleration},
			burstablePodInServ.Spec.Tolerations)
	}

	// Case 2: No MemoryPressure Toleration for BestEffort pod.
	besteffortPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "best-effort-pod",
			Namespace: nsName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "busybox",
					Image: "busybox",
				},
			},
		},
	}

	besteffortPodInServ, err := clientset.CoreV1().Pods(nsName).Create(besteffortPod)
	if err != nil {
		t.Errorf("Case 2: Failed to create pod: %v", err)
	} else if len(besteffortPodInServ.Spec.Tolerations) != 0 {
		t.Errorf("Case 2: Unexpected toleration # of BestEffort pod, expected: 0, got: %v",
			len(besteffortPodInServ.Spec.Tolerations))
	}

	// Case 3: Taint Node by NetworkUnavailable condition.
	networkUnavailableNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-1",
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4000m"),
				v1.ResourceMemory: resource.MustParse("16Gi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4000m"),
				v1.ResourceMemory: resource.MustParse("16Gi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeNetworkUnavailable,
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	}

	nodeInformerCh := make(chan bool)
	nodeInformer := informers.Core().V1().Nodes().Informer()
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			curNode := cur.(*v1.Node)
			if curNode.Name != "node-1" {
				return
			}
			for _, taint := range curNode.Spec.Taints {
				if taint.Key == algorithm.TaintNodeNetworkUnavailable &&
					taint.Effect == v1.TaintEffectNoSchedule {
					nodeInformerCh <- true
					break
				}
			}
		},
	})

	if _, err := clientset.CoreV1().Nodes().Create(networkUnavailableNode); err != nil {
		t.Errorf("Case 3: Failed to create node: %v", err)
	} else {
		select {
		case <-time.After(60 * time.Second):
			t.Errorf("Case 3: Failed to taint node after 60s.")
		case <-nodeInformerCh:
		}
	}

	// Case 4: Schedule Pod with NetworkUnavailable toleration.
	networkDaemonPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "network-daemon-pod",
			Namespace: nsName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "busybox",
					Image: "busybox",
				},
			},
			Tolerations: []v1.Toleration{
				{
					Key:      algorithm.TaintNodeNetworkUnavailable,
					Operator: v1.TolerationOpExists,
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
		},
	}

	if _, err := clientset.CoreV1().Pods(nsName).Create(networkDaemonPod); err != nil {
		t.Errorf("Case 4: Failed to create pod for network daemon: %v", err)
	} else {
		if err := waitForPodToScheduleWithTimeout(clientset, networkDaemonPod, time.Second*60); err != nil {
			t.Errorf("Case 4: Failed to schedule network daemon pod in 60s.")
		}
	}

	// Case 5: Taint node by unschedulable condition
	unschedulableNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-2",
		},
		Spec: v1.NodeSpec{
			Unschedulable: true,
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4000m"),
				v1.ResourceMemory: resource.MustParse("16Gi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4000m"),
				v1.ResourceMemory: resource.MustParse("16Gi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	nodeInformerCh2 := make(chan bool)
	nodeInformer2 := informers.Core().V1().Nodes().Informer()
	nodeInformer2.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			curNode := cur.(*v1.Node)
			if curNode.Name != "node-2" {
				return
			}

			for _, taint := range curNode.Spec.Taints {
				if taint.Key == algorithm.TaintNodeUnschedulable &&
					taint.Effect == v1.TaintEffectNoSchedule {
					nodeInformerCh2 <- true
					break
				}
			}
		},
	})

	if _, err := clientset.CoreV1().Nodes().Create(unschedulableNode); err != nil {
		t.Errorf("Case 5: Failed to create node: %v", err)
	} else {
		select {
		case <-time.After(60 * time.Second):
			t.Errorf("Case 5: Failed to taint node after 60s.")
		case <-nodeInformerCh2:
		}
	}
}
