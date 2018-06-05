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

// This file tests preemption functionality of the scheduler.

package scheduler

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/features"
	_ "k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
	testutils "k8s.io/kubernetes/test/utils"

	"github.com/golang/glog"
)

var lowPriority, mediumPriority, highPriority = int32(100), int32(200), int32(300)

func waitForNominatedNodeNameWithTimeout(cs clientset.Interface, pod *v1.Pod, timeout time.Duration) error {
	if err := wait.Poll(100*time.Millisecond, timeout, func() (bool, error) {
		pod, err := cs.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if len(pod.Status.NominatedNodeName) > 0 {
			return true, nil
		}
		return false, err
	}); err != nil {
		return fmt.Errorf("Pod %v annotation did not get set: %v", pod.Name, err)
	}
	return nil
}

func waitForNominatedNodeName(cs clientset.Interface, pod *v1.Pod) error {
	return waitForNominatedNodeNameWithTimeout(cs, pod, wait.ForeverTestTimeout)
}

// TestPreemption tests a few preemption scenarios.
func TestPreemption(t *testing.T) {
	// Enable PodPriority feature gate.
	utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%s=true", features.PodPriority))
	// Initialize scheduler.
	context := initTest(t, "preemption")
	defer cleanupTest(t, context)
	cs := context.clientSet

	defaultPodRes := &v1.ResourceRequirements{Requests: v1.ResourceList{
		v1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(100, resource.BinarySI)},
	}

	tests := []struct {
		description         string
		existingPods        []*v1.Pod
		pod                 *v1.Pod
		preemptedPodIndexes map[int]struct{}
	}{
		{
			description: "basic pod preemption",
			existingPods: []*v1.Pod{
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "victim-pod",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(400, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(200, resource.BinarySI)},
					},
				}),
			},
			pod: initPausePod(cs, &pausePodConfig{
				Name:      "preemptor-pod",
				Namespace: context.ns.Name,
				Priority:  &highPriority,
				Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(300, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(200, resource.BinarySI)},
				},
			}),
			preemptedPodIndexes: map[int]struct{}{0: {}},
		},
		{
			description: "preemption is performed to satisfy anti-affinity",
			existingPods: []*v1.Pod{
				initPausePod(cs, &pausePodConfig{
					Name: "pod-0", Namespace: context.ns.Name,
					Priority:  &mediumPriority,
					Labels:    map[string]string{"pod": "p0"},
					Resources: defaultPodRes,
				}),
				initPausePod(cs, &pausePodConfig{
					Name: "pod-1", Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Labels:    map[string]string{"pod": "p1"},
					Resources: defaultPodRes,
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "pod",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"preemptor"},
											},
										},
									},
									TopologyKey: "node",
								},
							},
						},
					},
				}),
			},
			// A higher priority pod with anti-affinity.
			pod: initPausePod(cs, &pausePodConfig{
				Name:      "preemptor-pod",
				Namespace: context.ns.Name,
				Priority:  &highPriority,
				Labels:    map[string]string{"pod": "preemptor"},
				Resources: defaultPodRes,
				Affinity: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "pod",
											Operator: metav1.LabelSelectorOpIn,
											Values:   []string{"p0"},
										},
									},
								},
								TopologyKey: "node",
							},
						},
					},
				},
			}),
			preemptedPodIndexes: map[int]struct{}{0: {}, 1: {}},
		},
		{
			// This is similar to the previous case only pod-1 is high priority.
			description: "preemption is not performed when anti-affinity is not satisfied",
			existingPods: []*v1.Pod{
				initPausePod(cs, &pausePodConfig{
					Name: "pod-0", Namespace: context.ns.Name,
					Priority:  &mediumPriority,
					Labels:    map[string]string{"pod": "p0"},
					Resources: defaultPodRes,
				}),
				initPausePod(cs, &pausePodConfig{
					Name: "pod-1", Namespace: context.ns.Name,
					Priority:  &highPriority,
					Labels:    map[string]string{"pod": "p1"},
					Resources: defaultPodRes,
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "pod",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"preemptor"},
											},
										},
									},
									TopologyKey: "node",
								},
							},
						},
					},
				}),
			},
			// A higher priority pod with anti-affinity.
			pod: initPausePod(cs, &pausePodConfig{
				Name:      "preemptor-pod",
				Namespace: context.ns.Name,
				Priority:  &highPriority,
				Labels:    map[string]string{"pod": "preemptor"},
				Resources: defaultPodRes,
				Affinity: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "pod",
											Operator: metav1.LabelSelectorOpIn,
											Values:   []string{"p0"},
										},
									},
								},
								TopologyKey: "node",
							},
						},
					},
				},
			}),
			preemptedPodIndexes: map[int]struct{}{},
		},
	}

	// Create a node with some resources and a label.
	nodeRes := &v1.ResourceList{
		v1.ResourcePods:   *resource.NewQuantity(32, resource.DecimalSI),
		v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(500, resource.BinarySI),
	}
	node, err := createNode(context.clientSet, "node1", nodeRes)
	if err != nil {
		t.Fatalf("Error creating nodes: %v", err)
	}
	nodeLabels := map[string]string{"node": node.Name}
	if err = testutils.AddLabelsToNode(context.clientSet, node.Name, nodeLabels); err != nil {
		t.Fatalf("Cannot add labels to node: %v", err)
	}
	if err = waitForNodeLabels(context.clientSet, node.Name, nodeLabels); err != nil {
		t.Fatalf("Adding labels to node didn't succeed: %v", err)
	}

	for _, test := range tests {
		pods := make([]*v1.Pod, len(test.existingPods))
		// Create and run existingPods.
		for i, p := range test.existingPods {
			pods[i], err = runPausePod(cs, p)
			if err != nil {
				t.Fatalf("Test [%v]: Error running pause pod: %v", test.description, err)
			}
		}
		// Create the "pod".
		preemptor, err := createPausePod(cs, test.pod)
		if err != nil {
			t.Errorf("Error while creating high priority pod: %v", err)
		}
		// Wait for preemption of pods and make sure the other ones are not preempted.
		for i, p := range pods {
			if _, found := test.preemptedPodIndexes[i]; found {
				if err = wait.Poll(time.Second, wait.ForeverTestTimeout, podIsGettingEvicted(cs, p.Namespace, p.Name)); err != nil {
					t.Errorf("Test [%v]: Pod %v is not getting evicted.", test.description, p.Name)
				}
			} else {
				if p.DeletionTimestamp != nil {
					t.Errorf("Test [%v]: Didn't expect pod %v to get preempted.", test.description, p.Name)
				}
			}
		}
		// Also check that the preemptor pod gets the annotation for nominated node name.
		if len(test.preemptedPodIndexes) > 0 {
			if err := waitForNominatedNodeName(cs, preemptor); err != nil {
				t.Errorf("Test [%v]: NominatedNodeName annotation was not set for pod %v: %v", test.description, preemptor.Name, err)
			}
		}

		// Cleanup
		pods = append(pods, preemptor)
		cleanupPods(cs, t, pods)
	}
}

// TestDisablePreemption tests disable pod preemption of scheduler works as expected.
func TestDisablePreemption(t *testing.T) {
	// Enable PodPriority feature gate.
	utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%s=true", features.PodPriority))
	// Initialize scheduler, and disable preemption.
	context := initTestDisablePreemption(t, "disable-preemption")
	defer cleanupTest(t, context)
	cs := context.clientSet

	tests := []struct {
		description  string
		existingPods []*v1.Pod
		pod          *v1.Pod
	}{
		{
			description: "pod preemption will not happen",
			existingPods: []*v1.Pod{
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "victim-pod",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(400, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(200, resource.BinarySI)},
					},
				}),
			},
			pod: initPausePod(cs, &pausePodConfig{
				Name:      "preemptor-pod",
				Namespace: context.ns.Name,
				Priority:  &highPriority,
				Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(300, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(200, resource.BinarySI)},
				},
			}),
		},
	}

	// Create a node with some resources and a label.
	nodeRes := &v1.ResourceList{
		v1.ResourcePods:   *resource.NewQuantity(32, resource.DecimalSI),
		v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(500, resource.BinarySI),
	}
	_, err := createNode(context.clientSet, "node1", nodeRes)
	if err != nil {
		t.Fatalf("Error creating nodes: %v", err)
	}

	for _, test := range tests {
		pods := make([]*v1.Pod, len(test.existingPods))
		// Create and run existingPods.
		for i, p := range test.existingPods {
			pods[i], err = runPausePod(cs, p)
			if err != nil {
				t.Fatalf("Test [%v]: Error running pause pod: %v", test.description, err)
			}
		}
		// Create the "pod".
		preemptor, err := createPausePod(cs, test.pod)
		if err != nil {
			t.Errorf("Error while creating high priority pod: %v", err)
		}
		// Ensure preemptor should keep unschedulable.
		if err := waitForPodUnschedulable(cs, preemptor); err != nil {
			t.Errorf("Test [%v]: Preemptor %v should not become scheduled",
				test.description, preemptor.Name)
		}

		// Ensure preemptor should not be nominated.
		if err := waitForNominatedNodeNameWithTimeout(cs, preemptor, 5*time.Second); err == nil {
			t.Errorf("Test [%v]: Preemptor %v should not be nominated",
				test.description, preemptor.Name)
		}

		// Cleanup
		pods = append(pods, preemptor)
		cleanupPods(cs, t, pods)
	}
}

func mkPriorityPodWithGrace(tc *TestContext, name string, priority int32, grace int64) *v1.Pod {
	defaultPodRes := &v1.ResourceRequirements{Requests: v1.ResourceList{
		v1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(100, resource.BinarySI)},
	}
	pod := initPausePod(tc.clientSet, &pausePodConfig{
		Name:      name,
		Namespace: tc.ns.Name,
		Priority:  &priority,
		Labels:    map[string]string{"pod": name},
		Resources: defaultPodRes,
	})
	// Setting grace period to zero. Otherwise, we may never see the actual deletion
	// of the pods in integration tests.
	pod.Spec.TerminationGracePeriodSeconds = &grace
	return pod
}

// This test ensures that while the preempting pod is waiting for the victims to
// terminate, other pending lower priority pods are not scheduled in the room created
// after preemption and while the higher priority pods is not scheduled yet.
func TestPreemptionStarvation(t *testing.T) {
	// Enable PodPriority feature gate.
	utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%s=true", features.PodPriority))
	// Initialize scheduler.
	context := initTest(t, "preemption")
	defer cleanupTest(t, context)
	cs := context.clientSet

	tests := []struct {
		description        string
		numExistingPod     int
		numExpectedPending int
		preemptor          *v1.Pod
	}{
		{
			// This test ensures that while the preempting pod is waiting for the victims
			// terminate, other lower priority pods are not scheduled in the room created
			// after preemption and while the higher priority pods is not scheduled yet.
			description:        "starvation test: higher priority pod is scheduled before the lower priority ones",
			numExistingPod:     10,
			numExpectedPending: 5,
			preemptor: initPausePod(cs, &pausePodConfig{
				Name:      "preemptor-pod",
				Namespace: context.ns.Name,
				Priority:  &highPriority,
				Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(300, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(200, resource.BinarySI)},
				},
			}),
		},
	}

	// Create a node with some resources and a label.
	nodeRes := &v1.ResourceList{
		v1.ResourcePods:   *resource.NewQuantity(32, resource.DecimalSI),
		v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(500, resource.BinarySI),
	}
	_, err := createNode(context.clientSet, "node1", nodeRes)
	if err != nil {
		t.Fatalf("Error creating nodes: %v", err)
	}

	for _, test := range tests {
		pendingPods := make([]*v1.Pod, test.numExpectedPending)
		numRunningPods := test.numExistingPod - test.numExpectedPending
		runningPods := make([]*v1.Pod, numRunningPods)
		// Create and run existingPods.
		for i := 0; i < numRunningPods; i++ {
			runningPods[i], err = createPausePod(cs, mkPriorityPodWithGrace(context, fmt.Sprintf("rpod-%v", i), mediumPriority, 0))
			if err != nil {
				t.Fatalf("Test [%v]: Error creating pause pod: %v", test.description, err)
			}
		}
		// make sure that runningPods are all scheduled.
		for _, p := range runningPods {
			if err := waitForPodToSchedule(cs, p); err != nil {
				t.Fatalf("Pod %v didn't get scheduled: %v", p.Name, err)
			}
		}
		// Create pending pods.
		for i := 0; i < test.numExpectedPending; i++ {
			pendingPods[i], err = createPausePod(cs, mkPriorityPodWithGrace(context, fmt.Sprintf("ppod-%v", i), mediumPriority, 0))
			if err != nil {
				t.Fatalf("Test [%v]: Error creating pending pod: %v", test.description, err)
			}
		}
		// Make sure that all pending pods are being marked unschedulable.
		for _, p := range pendingPods {
			if err := wait.Poll(100*time.Millisecond, wait.ForeverTestTimeout,
				podUnschedulable(cs, p.Namespace, p.Name)); err != nil {
				t.Errorf("Pod %v didn't get marked unschedulable: %v", p.Name, err)
			}
		}
		// Create the preemptor.
		preemptor, err := createPausePod(cs, test.preemptor)
		if err != nil {
			t.Errorf("Error while creating the preempting pod: %v", err)
		}
		// Check that the preemptor pod gets the annotation for nominated node name.
		if err := waitForNominatedNodeName(cs, preemptor); err != nil {
			t.Errorf("Test [%v]: NominatedNodeName annotation was not set for pod %v: %v", test.description, preemptor.Name, err)
		}
		// Make sure that preemptor is scheduled after preemptions.
		if err := waitForPodToScheduleWithTimeout(cs, preemptor, 60*time.Second); err != nil {
			t.Errorf("Preemptor pod %v didn't get scheduled: %v", preemptor.Name, err)
		}
		// Cleanup
		glog.Info("Cleaning up all pods...")
		allPods := pendingPods
		allPods = append(allPods, runningPods...)
		allPods = append(allPods, preemptor)
		cleanupPods(cs, t, allPods)
	}
}

// TestNominatedNodeCleanUp checks that when there are nominated pods on a
// node and a higher priority pod is nominated to run on the node, the nominated
// node name of the lower priority pods is cleared.
// Test scenario:
// 1. Create a few low priority pods with long grade period that fill up a node.
// 2. Create a medium priority pod that preempt some of those pods.
// 3. Check that nominated node name of the medium priority pod is set.
// 4. Create a high priority pod that preempts some pods on that node.
// 5. Check that nominated node name of the high priority pod is set and nominated
//    node name of the medium priority pod is cleared.
func TestNominatedNodeCleanUp(t *testing.T) {
	// Enable PodPriority feature gate.
	utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%s=true", features.PodPriority))
	// Initialize scheduler.
	context := initTest(t, "preemption")
	defer cleanupTest(t, context)
	cs := context.clientSet
	// Create a node with some resources and a label.
	nodeRes := &v1.ResourceList{
		v1.ResourcePods:   *resource.NewQuantity(32, resource.DecimalSI),
		v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(500, resource.BinarySI),
	}
	_, err := createNode(context.clientSet, "node1", nodeRes)
	if err != nil {
		t.Fatalf("Error creating nodes: %v", err)
	}

	// Step 1. Create a few low priority pods.
	lowPriPods := make([]*v1.Pod, 4)
	for i := 0; i < len(lowPriPods); i++ {
		lowPriPods[i], err = createPausePod(cs, mkPriorityPodWithGrace(context, fmt.Sprintf("lpod-%v", i), lowPriority, 60))
		if err != nil {
			t.Fatalf("Error creating pause pod: %v", err)
		}
	}
	// make sure that the pods are all scheduled.
	for _, p := range lowPriPods {
		if err := waitForPodToSchedule(cs, p); err != nil {
			t.Fatalf("Pod %v didn't get scheduled: %v", p.Name, err)
		}
	}
	// Step 2. Create a medium priority pod.
	podConf := initPausePod(cs, &pausePodConfig{
		Name:      "medium-priority",
		Namespace: context.ns.Name,
		Priority:  &mediumPriority,
		Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
			v1.ResourceCPU:    *resource.NewMilliQuantity(400, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(400, resource.BinarySI)},
		},
	})
	medPriPod, err := createPausePod(cs, podConf)
	if err != nil {
		t.Errorf("Error while creating the medium priority pod: %v", err)
	}
	// Step 3. Check that nominated node name of the medium priority pod is set.
	if err := waitForNominatedNodeName(cs, medPriPod); err != nil {
		t.Errorf("NominatedNodeName annotation was not set for pod %v: %v", medPriPod.Name, err)
	}
	// Step 4. Create a high priority pod.
	podConf = initPausePod(cs, &pausePodConfig{
		Name:      "high-priority",
		Namespace: context.ns.Name,
		Priority:  &highPriority,
		Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
			v1.ResourceCPU:    *resource.NewMilliQuantity(300, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(200, resource.BinarySI)},
		},
	})
	highPriPod, err := createPausePod(cs, podConf)
	if err != nil {
		t.Errorf("Error while creating the high priority pod: %v", err)
	}
	// Step 5. Check that nominated node name of the high priority pod is set.
	if err := waitForNominatedNodeName(cs, highPriPod); err != nil {
		t.Errorf("NominatedNodeName annotation was not set for pod %v: %v", medPriPod.Name, err)
	}
	// And the nominated node name of the medium priority pod is cleared.
	if err := wait.Poll(100*time.Millisecond, wait.ForeverTestTimeout, func() (bool, error) {
		pod, err := cs.CoreV1().Pods(medPriPod.Namespace).Get(medPriPod.Name, metav1.GetOptions{})
		if err != nil {
			t.Errorf("Error getting the medium priority pod info: %v", err)
		}
		if len(pod.Status.NominatedNodeName) == 0 {
			return true, nil
		}
		return false, err
	}); err != nil {
		t.Errorf("The nominated node name of the medium priority pod was not cleared: %v", err)
	}
}

func mkMinAvailablePDB(name, namespace string, uid types.UID, minAvailable int, matchLabels map[string]string) *policy.PodDisruptionBudget {
	intMinAvailable := intstr.FromInt(minAvailable)
	return &policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: policy.PodDisruptionBudgetSpec{
			MinAvailable: &intMinAvailable,
			Selector:     &metav1.LabelSelector{MatchLabels: matchLabels},
		},
	}
}

// TestPDBInPreemption tests PodDisruptionBudget support in preemption.
func TestPDBInPreemption(t *testing.T) {
	// Enable PodPriority feature gate.
	utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%s=true", features.PodPriority))
	// Initialize scheduler.
	context := initTest(t, "preemption-pdb")
	defer cleanupTest(t, context)
	cs := context.clientSet

	defaultPodRes := &v1.ResourceRequirements{Requests: v1.ResourceList{
		v1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(100, resource.BinarySI)},
	}
	defaultNodeRes := &v1.ResourceList{
		v1.ResourcePods:   *resource.NewQuantity(32, resource.DecimalSI),
		v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(500, resource.BinarySI),
	}

	type nodeConfig struct {
		name string
		res  *v1.ResourceList
	}

	tests := []struct {
		description         string
		nodes               []*nodeConfig
		pdbs                []*policy.PodDisruptionBudget
		existingPods        []*v1.Pod
		pod                 *v1.Pod
		preemptedPodIndexes map[int]struct{}
	}{
		{
			description: "A non-PDB violating pod is preempted despite its higher priority",
			nodes:       []*nodeConfig{{name: "node-1", res: defaultNodeRes}},
			pdbs: []*policy.PodDisruptionBudget{
				mkMinAvailablePDB("pdb-1", context.ns.Name, types.UID("pdb-1-uid"), 2, map[string]string{"foo": "bar"}),
			},
			existingPods: []*v1.Pod{
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "low-pod1",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: defaultPodRes,
					Labels:    map[string]string{"foo": "bar"},
				}),
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "low-pod2",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: defaultPodRes,
					Labels:    map[string]string{"foo": "bar"},
				}),
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "mid-pod3",
					Namespace: context.ns.Name,
					Priority:  &mediumPriority,
					Resources: defaultPodRes,
				}),
			},
			pod: initPausePod(cs, &pausePodConfig{
				Name:      "preemptor-pod",
				Namespace: context.ns.Name,
				Priority:  &highPriority,
				Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(300, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(200, resource.BinarySI)},
				},
			}),
			preemptedPodIndexes: map[int]struct{}{2: {}},
		},
		{
			description: "A node without any PDB violating pods is preferred for preemption",
			nodes: []*nodeConfig{
				{name: "node-1", res: defaultNodeRes},
				{name: "node-2", res: defaultNodeRes},
			},
			pdbs: []*policy.PodDisruptionBudget{
				mkMinAvailablePDB("pdb-1", context.ns.Name, types.UID("pdb-1-uid"), 2, map[string]string{"foo": "bar"}),
			},
			existingPods: []*v1.Pod{
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "low-pod1",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: defaultPodRes,
					NodeName:  "node-1",
					Labels:    map[string]string{"foo": "bar"},
				}),
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "mid-pod2",
					Namespace: context.ns.Name,
					Priority:  &mediumPriority,
					NodeName:  "node-2",
					Resources: defaultPodRes,
				}),
			},
			pod: initPausePod(cs, &pausePodConfig{
				Name:      "preemptor-pod",
				Namespace: context.ns.Name,
				Priority:  &highPriority,
				Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(200, resource.BinarySI)},
				},
			}),
			preemptedPodIndexes: map[int]struct{}{1: {}},
		},
		{
			description: "A node with fewer PDB violating pods is preferred for preemption",
			nodes: []*nodeConfig{
				{name: "node-1", res: defaultNodeRes},
				{name: "node-2", res: defaultNodeRes},
				{name: "node-3", res: defaultNodeRes},
			},
			pdbs: []*policy.PodDisruptionBudget{
				mkMinAvailablePDB("pdb-1", context.ns.Name, types.UID("pdb-1-uid"), 2, map[string]string{"foo1": "bar"}),
				mkMinAvailablePDB("pdb-2", context.ns.Name, types.UID("pdb-2-uid"), 2, map[string]string{"foo2": "bar"}),
			},
			existingPods: []*v1.Pod{
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "low-pod1",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: defaultPodRes,
					NodeName:  "node-1",
					Labels:    map[string]string{"foo1": "bar"},
				}),
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "mid-pod1",
					Namespace: context.ns.Name,
					Priority:  &mediumPriority,
					Resources: defaultPodRes,
					NodeName:  "node-1",
				}),
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "low-pod2",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: defaultPodRes,
					NodeName:  "node-2",
					Labels:    map[string]string{"foo2": "bar"},
				}),
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "mid-pod2",
					Namespace: context.ns.Name,
					Priority:  &mediumPriority,
					Resources: defaultPodRes,
					NodeName:  "node-2",
					Labels:    map[string]string{"foo2": "bar"},
				}),
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "low-pod4",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: defaultPodRes,
					NodeName:  "node-3",
					Labels:    map[string]string{"foo2": "bar"},
				}),
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "low-pod5",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: defaultPodRes,
					NodeName:  "node-3",
					Labels:    map[string]string{"foo2": "bar"},
				}),
				initPausePod(context.clientSet, &pausePodConfig{
					Name:      "low-pod6",
					Namespace: context.ns.Name,
					Priority:  &lowPriority,
					Resources: defaultPodRes,
					NodeName:  "node-3",
					Labels:    map[string]string{"foo2": "bar"},
				}),
			},
			pod: initPausePod(cs, &pausePodConfig{
				Name:      "preemptor-pod",
				Namespace: context.ns.Name,
				Priority:  &highPriority,
				Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(200, resource.BinarySI)},
				},
			}),
			preemptedPodIndexes: map[int]struct{}{0: {}, 1: {}},
		},
	}

	for _, test := range tests {
		for _, nodeConf := range test.nodes {
			_, err := createNode(cs, nodeConf.name, nodeConf.res)
			if err != nil {
				t.Fatalf("Error creating node %v: %v", nodeConf.name, err)
			}
		}
		// Create PDBs.
		for _, pdb := range test.pdbs {
			_, err := context.clientSet.PolicyV1beta1().PodDisruptionBudgets(context.ns.Name).Create(pdb)
			if err != nil {
				t.Fatalf("Failed to create PDB: %v", err)
			}
		}
		// Wait for PDBs to show up in the scheduler's cache.
		if err := wait.Poll(time.Second, 15*time.Second, func() (bool, error) {
			cachedPDBs, err := context.scheduler.Config().SchedulerCache.ListPDBs(labels.Everything())
			if err != nil {
				t.Errorf("Error while polling for PDB: %v", err)
				return false, err
			}
			return len(cachedPDBs) == len(test.pdbs), err
		}); err != nil {
			t.Fatalf("Not all PDBs were added to the cache: %v", err)
		}

		pods := make([]*v1.Pod, len(test.existingPods))
		var err error
		// Create and run existingPods.
		for i, p := range test.existingPods {
			if pods[i], err = runPausePod(cs, p); err != nil {
				t.Fatalf("Test [%v]: Error running pause pod: %v", test.description, err)
			}
		}
		// Create the "pod".
		preemptor, err := createPausePod(cs, test.pod)
		if err != nil {
			t.Errorf("Error while creating high priority pod: %v", err)
		}
		// Wait for preemption of pods and make sure the other ones are not preempted.
		for i, p := range pods {
			if _, found := test.preemptedPodIndexes[i]; found {
				if err = wait.Poll(time.Second, wait.ForeverTestTimeout, podIsGettingEvicted(cs, p.Namespace, p.Name)); err != nil {
					t.Errorf("Test [%v]: Pod %v is not getting evicted.", test.description, p.Name)
				}
			} else {
				if p.DeletionTimestamp != nil {
					t.Errorf("Test [%v]: Didn't expect pod %v to get preempted.", test.description, p.Name)
				}
			}
		}
		// Also check that the preemptor pod gets the annotation for nominated node name.
		if len(test.preemptedPodIndexes) > 0 {
			if err := waitForNominatedNodeName(cs, preemptor); err != nil {
				t.Errorf("Test [%v]: NominatedNodeName annotation was not set for pod %v: %v", test.description, preemptor.Name, err)
			}
		}

		// Cleanup
		pods = append(pods, preemptor)
		cleanupPods(cs, t, pods)
		cs.PolicyV1beta1().PodDisruptionBudgets(context.ns.Name).DeleteCollection(nil, metav1.ListOptions{})
		cs.CoreV1().Nodes().DeleteCollection(nil, metav1.ListOptions{})
	}
}
