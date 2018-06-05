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

package deployment

import (
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/kubernetes/pkg/controller/deployment"
	deploymentutil "k8s.io/kubernetes/pkg/controller/deployment/util"
	"k8s.io/kubernetes/pkg/controller/replicaset"
	"k8s.io/kubernetes/pkg/util/metrics"
	"k8s.io/kubernetes/test/integration/framework"
	testutil "k8s.io/kubernetes/test/utils"
)

const (
	pollInterval = 100 * time.Millisecond
	pollTimeout  = 60 * time.Second

	fakeContainerName = "fake-name"
	fakeImage         = "fakeimage"
)

var pauseFn = func(update *v1beta1.Deployment) {
	update.Spec.Paused = true
}

var resumeFn = func(update *v1beta1.Deployment) {
	update.Spec.Paused = false
}

type deploymentTester struct {
	t          *testing.T
	c          clientset.Interface
	deployment *v1beta1.Deployment
}

func testLabels() map[string]string {
	return map[string]string{"name": "test"}
}

// newDeployment returns a RollingUpdate Deployment with with a fake container image
func newDeployment(name, ns string, replicas int32) *v1beta1.Deployment {
	return &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: testLabels()},
			Strategy: v1beta1.DeploymentStrategy{
				Type:          v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: new(v1beta1.RollingUpdateDeployment),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: testLabels(),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  fakeContainerName,
							Image: fakeImage,
						},
					},
				},
			},
		},
	}
}

func newReplicaSet(name, ns string, replicas int32) *v1beta1.ReplicaSet {
	return &v1beta1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: v1beta1.ReplicaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: testLabels(),
			},
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: testLabels(),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  fakeContainerName,
							Image: fakeImage,
						},
					},
				},
			},
		},
	}
}

func newDeploymentRollback(name string, annotations map[string]string, revision int64) *v1beta1.DeploymentRollback {
	return &v1beta1.DeploymentRollback{
		Name:               name,
		UpdatedAnnotations: annotations,
		RollbackTo:         v1beta1.RollbackConfig{Revision: revision},
	}
}

// dcSetup sets up necessities for Deployment integration test, including master, apiserver, informers, and clientset
func dcSetup(t *testing.T) (*httptest.Server, framework.CloseFunc, *replicaset.ReplicaSetController, *deployment.DeploymentController, informers.SharedInformerFactory, clientset.Interface) {
	masterConfig := framework.NewIntegrationTestMasterConfig()
	_, s, closeFn := framework.RunAMaster(masterConfig)

	config := restclient.Config{Host: s.URL}
	clientSet, err := clientset.NewForConfig(&config)
	if err != nil {
		t.Fatalf("error in create clientset: %v", err)
	}
	resyncPeriod := 12 * time.Hour
	informers := informers.NewSharedInformerFactory(clientset.NewForConfigOrDie(restclient.AddUserAgent(&config, "deployment-informers")), resyncPeriod)

	metrics.UnregisterMetricAndUntrackRateLimiterUsage("deployment_controller")
	dc, err := deployment.NewDeploymentController(
		informers.Extensions().V1beta1().Deployments(),
		informers.Extensions().V1beta1().ReplicaSets(),
		informers.Core().V1().Pods(),
		clientset.NewForConfigOrDie(restclient.AddUserAgent(&config, "deployment-controller")),
	)
	if err != nil {
		t.Fatalf("error creating Deployment controller: %v", err)
	}
	rm := replicaset.NewReplicaSetController(
		informers.Apps().V1().ReplicaSets(),
		informers.Core().V1().Pods(),
		clientset.NewForConfigOrDie(restclient.AddUserAgent(&config, "replicaset-controller")),
		replicaset.BurstReplicas,
	)
	return s, closeFn, rm, dc, informers, clientSet
}

// dcSimpleSetup sets up necessities for Deployment integration test, including master, apiserver,
// and clientset, but not controllers and informers
func dcSimpleSetup(t *testing.T) (*httptest.Server, framework.CloseFunc, clientset.Interface) {
	masterConfig := framework.NewIntegrationTestMasterConfig()
	_, s, closeFn := framework.RunAMaster(masterConfig)

	config := restclient.Config{Host: s.URL}
	clientSet, err := clientset.NewForConfig(&config)
	if err != nil {
		t.Fatalf("error in create clientset: %v", err)
	}
	return s, closeFn, clientSet
}

// addPodConditionReady sets given pod status to ready at given time
func addPodConditionReady(pod *v1.Pod, time metav1.Time) {
	pod.Status = v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{
			{
				Type:               v1.PodReady,
				Status:             v1.ConditionTrue,
				LastTransitionTime: time,
			},
		},
	}
}

func (d *deploymentTester) waitForDeploymentRevisionAndImage(revision, image string) error {
	if err := testutil.WaitForDeploymentRevisionAndImage(d.c, d.deployment.Namespace, d.deployment.Name, revision, image, d.t.Logf, pollInterval, pollTimeout); err != nil {
		return fmt.Errorf("failed to wait for Deployment revision %s: %v", d.deployment.Name, err)
	}
	return nil
}

func markPodReady(c clientset.Interface, ns string, pod *v1.Pod) error {
	addPodConditionReady(pod, metav1.Now())
	_, err := c.CoreV1().Pods(ns).UpdateStatus(pod)
	return err
}

func intOrStrP(num int) *intstr.IntOrString {
	intstr := intstr.FromInt(num)
	return &intstr
}

// markUpdatedPodsReady manually marks updated Deployment pods status to ready,
// until the deployment is complete
func (d *deploymentTester) markUpdatedPodsReady(wg *sync.WaitGroup) {
	defer wg.Done()

	ns := d.deployment.Namespace
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		// We're done when the deployment is complete
		if completed, err := d.deploymentComplete(); err != nil {
			return false, err
		} else if completed {
			return true, nil
		}
		// Otherwise, mark remaining pods as ready
		pods, err := d.listUpdatedPods()
		if err != nil {
			d.t.Log(err)
			return false, nil
		}
		d.t.Logf("%d/%d of deployment pods are created", len(pods), *d.deployment.Spec.Replicas)
		for i := range pods {
			pod := pods[i]
			if podutil.IsPodReady(&pod) {
				continue
			}
			if err = markPodReady(d.c, ns, &pod); err != nil {
				d.t.Logf("failed to update Deployment pod %s, will retry later: %v", pod.Name, err)
			}
		}
		return false, nil
	})
	if err != nil {
		d.t.Fatalf("failed to mark updated Deployment pods to ready: %v", err)
	}
}

func (d *deploymentTester) deploymentComplete() (bool, error) {
	latest, err := d.c.ExtensionsV1beta1().Deployments(d.deployment.Namespace).Get(d.deployment.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return deploymentutil.DeploymentComplete(d.deployment, &latest.Status), nil
}

// Waits for the deployment to complete, and check rolling update strategy isn't broken at any times.
// Rolling update strategy should not be broken during a rolling update.
func (d *deploymentTester) waitForDeploymentCompleteAndCheckRolling() error {
	return testutil.WaitForDeploymentCompleteAndCheckRolling(d.c, d.deployment, d.t.Logf, pollInterval, pollTimeout)
}

// Waits for the deployment to complete, and don't check if rolling update strategy is broken.
// Rolling update strategy is used only during a rolling update, and can be violated in other situations,
// such as shortly after a scaling event or the deployment is just created.
func (d *deploymentTester) waitForDeploymentComplete() error {
	return testutil.WaitForDeploymentComplete(d.c, d.deployment, d.t.Logf, pollInterval, pollTimeout)
}

// waitForDeploymentCompleteAndCheckRollingAndMarkPodsReady waits for the Deployment to complete
// while marking updated Deployment pods as ready at the same time.
// Uses hard check to make sure rolling update strategy is not violated at any times.
func (d *deploymentTester) waitForDeploymentCompleteAndCheckRollingAndMarkPodsReady() error {
	var wg sync.WaitGroup

	// Manually mark updated Deployment pods as ready in a separate goroutine
	wg.Add(1)
	go d.markUpdatedPodsReady(&wg)

	// Wait for the Deployment status to complete while Deployment pods are becoming ready
	err := d.waitForDeploymentCompleteAndCheckRolling()
	if err != nil {
		return fmt.Errorf("failed to wait for Deployment %s to complete: %v", d.deployment.Name, err)
	}

	// Wait for goroutine to finish
	wg.Wait()

	return nil
}

// waitForDeploymentCompleteAndMarkPodsReady waits for the Deployment to complete
// while marking updated Deployment pods as ready at the same time.
func (d *deploymentTester) waitForDeploymentCompleteAndMarkPodsReady() error {
	var wg sync.WaitGroup

	// Manually mark updated Deployment pods as ready in a separate goroutine
	wg.Add(1)
	go d.markUpdatedPodsReady(&wg)

	// Wait for the Deployment status to complete using soft check, while Deployment pods are becoming ready
	err := d.waitForDeploymentComplete()
	if err != nil {
		return fmt.Errorf("failed to wait for Deployment status %s: %v", d.deployment.Name, err)
	}

	// Wait for goroutine to finish
	wg.Wait()

	return nil
}

func (d *deploymentTester) updateDeployment(applyUpdate testutil.UpdateDeploymentFunc) (*v1beta1.Deployment, error) {
	return testutil.UpdateDeploymentWithRetries(d.c, d.deployment.Namespace, d.deployment.Name, applyUpdate, d.t.Logf, pollInterval, pollTimeout)
}

func (d *deploymentTester) waitForObservedDeployment(desiredGeneration int64) error {
	if err := testutil.WaitForObservedDeployment(d.c, d.deployment.Namespace, d.deployment.Name, desiredGeneration); err != nil {
		return fmt.Errorf("failed waiting for ObservedGeneration of deployment %s to become %d: %v", d.deployment.Name, desiredGeneration, err)
	}
	return nil
}

func (d *deploymentTester) getNewReplicaSet() (*v1beta1.ReplicaSet, error) {
	deployment, err := d.c.ExtensionsV1beta1().Deployments(d.deployment.Namespace).Get(d.deployment.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed retrieving deployment %s: %v", d.deployment.Name, err)
	}
	rs, err := deploymentutil.GetNewReplicaSet(deployment, d.c.ExtensionsV1beta1())
	if err != nil {
		return nil, fmt.Errorf("failed retrieving new replicaset of deployment %s: %v", d.deployment.Name, err)
	}
	return rs, nil
}

func (d *deploymentTester) expectNoNewReplicaSet() error {
	rs, err := d.getNewReplicaSet()
	if err != nil {
		return err
	}
	if rs != nil {
		return fmt.Errorf("expected deployment %s not to create a new replicaset, got %v", d.deployment.Name, rs)
	}
	return nil
}

func (d *deploymentTester) expectNewReplicaSet() (*v1beta1.ReplicaSet, error) {
	rs, err := d.getNewReplicaSet()
	if err != nil {
		return nil, err
	}
	if rs == nil {
		return nil, fmt.Errorf("expected deployment %s to create a new replicaset, got nil", d.deployment.Name)
	}
	return rs, nil
}

func (d *deploymentTester) updateReplicaSet(name string, applyUpdate testutil.UpdateExtensionsReplicaSetFunc) (*v1beta1.ReplicaSet, error) {
	return testutil.UpdateExtensionsReplicaSetWithRetries(d.c, d.deployment.Namespace, name, applyUpdate, d.t.Logf, pollInterval, pollTimeout)
}

func (d *deploymentTester) updateReplicaSetStatus(name string, applyStatusUpdate testutil.UpdateExtensionsReplicaSetFunc) (*v1beta1.ReplicaSet, error) {
	return testutil.UpdateExtensionsReplicaSetStatusWithRetries(d.c, d.deployment.Namespace, name, applyStatusUpdate, d.t.Logf, pollInterval, pollTimeout)
}

// waitForDeploymentRollbackCleared waits for deployment either started rolling back or doesn't need to rollback.
func (d *deploymentTester) waitForDeploymentRollbackCleared() error {
	return testutil.WaitForDeploymentRollbackCleared(d.c, d.deployment.Namespace, d.deployment.Name, pollInterval, pollTimeout)
}

// checkDeploymentRevisionAndImage checks if the input deployment's and its new replica set's revision and image are as expected.
func (d *deploymentTester) checkDeploymentRevisionAndImage(revision, image string) error {
	return testutil.CheckDeploymentRevisionAndImage(d.c, d.deployment.Namespace, d.deployment.Name, revision, image)
}

func (d *deploymentTester) waitForDeploymentUpdatedReplicasGTE(minUpdatedReplicas int32) error {
	return testutil.WaitForDeploymentUpdatedReplicasGTE(d.c, d.deployment.Namespace, d.deployment.Name, minUpdatedReplicas, d.deployment.Generation, pollInterval, pollTimeout)
}

func (d *deploymentTester) waitForDeploymentWithCondition(reason string, condType v1beta1.DeploymentConditionType) error {
	return testutil.WaitForDeploymentWithCondition(d.c, d.deployment.Namespace, d.deployment.Name, reason, condType, d.t.Logf, pollInterval, pollTimeout)
}

func (d *deploymentTester) listUpdatedPods() ([]v1.Pod, error) {
	selector, err := metav1.LabelSelectorAsSelector(d.deployment.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment selector: %v", err)
	}
	pods, err := d.c.CoreV1().Pods(d.deployment.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployment pods, will retry later: %v", err)
	}
	newRS, err := d.getNewReplicaSet()
	if err != nil {
		return nil, fmt.Errorf("failed to get new replicaset of deployment %q: %v", d.deployment.Name, err)
	}
	if newRS == nil {
		return nil, fmt.Errorf("unable to find new replicaset of deployment %q", d.deployment.Name)
	}

	var ownedPods []v1.Pod
	for _, pod := range pods.Items {
		rs := metav1.GetControllerOf(&pod)
		if rs.UID == newRS.UID {
			ownedPods = append(ownedPods, pod)
		}
	}
	return ownedPods, nil
}

func (d *deploymentTester) waitRSStable(replicaset *v1beta1.ReplicaSet) error {
	return testutil.WaitExtensionsRSStable(d.t, d.c, replicaset, pollInterval, pollTimeout)
}

func (d *deploymentTester) scaleDeployment(newReplicas int32) error {
	var err error
	d.deployment, err = d.updateDeployment(func(update *v1beta1.Deployment) {
		update.Spec.Replicas = &newReplicas
	})
	if err != nil {
		return fmt.Errorf("failed updating deployment %q: %v", d.deployment.Name, err)
	}

	if err := d.waitForDeploymentCompleteAndMarkPodsReady(); err != nil {
		return err
	}

	rs, err := d.expectNewReplicaSet()
	if err != nil {
		return err
	}
	if *rs.Spec.Replicas != newReplicas {
		return fmt.Errorf("expected new replicaset replicas = %d, got %d", newReplicas, *rs.Spec.Replicas)
	}
	return nil
}

// waitForReadyReplicas waits for number of ready replicas to equal number of replicas.
func (d *deploymentTester) waitForReadyReplicas() error {
	if err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		deployment, err := d.c.ExtensionsV1beta1().Deployments(d.deployment.Namespace).Get(d.deployment.Name, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to get deployment %q: %v", d.deployment.Name, err)
		}
		return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas, nil
	}); err != nil {
		return fmt.Errorf("failed to wait for .readyReplicas to equal .replicas: %v", err)
	}
	return nil
}

// markUpdatedPodsReadyWithoutComplete marks updated Deployment pods as ready without waiting for deployment to complete.
func (d *deploymentTester) markUpdatedPodsReadyWithoutComplete() error {
	if err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		pods, err := d.listUpdatedPods()
		if err != nil {
			return false, err
		}
		for i := range pods {
			pod := pods[i]
			if podutil.IsPodReady(&pod) {
				continue
			}
			if err = markPodReady(d.c, d.deployment.Namespace, &pod); err != nil {
				d.t.Logf("failed to update Deployment pod %q, will retry later: %v", pod.Name, err)
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("failed to mark all updated pods as ready: %v", err)
	}
	return nil
}

// Verify all replicas fields of DeploymentStatus have desired count.
// Immediately return an error when found a non-matching replicas field.
func (d *deploymentTester) checkDeploymentStatusReplicasFields(replicas, updatedReplicas, readyReplicas, availableReplicas, unavailableReplicas int32) error {
	deployment, err := d.c.ExtensionsV1beta1().Deployments(d.deployment.Namespace).Get(d.deployment.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %q: %v", d.deployment.Name, err)
	}
	if deployment.Status.Replicas != replicas {
		return fmt.Errorf("unexpected .replicas: expect %d, got %d", replicas, deployment.Status.Replicas)
	}
	if deployment.Status.UpdatedReplicas != updatedReplicas {
		return fmt.Errorf("unexpected .updatedReplicas: expect %d, got %d", updatedReplicas, deployment.Status.UpdatedReplicas)
	}
	if deployment.Status.ReadyReplicas != readyReplicas {
		return fmt.Errorf("unexpected .readyReplicas: expect %d, got %d", readyReplicas, deployment.Status.ReadyReplicas)
	}
	if deployment.Status.AvailableReplicas != availableReplicas {
		return fmt.Errorf("unexpected .replicas: expect %d, got %d", availableReplicas, deployment.Status.AvailableReplicas)
	}
	if deployment.Status.UnavailableReplicas != unavailableReplicas {
		return fmt.Errorf("unexpected .replicas: expect %d, got %d", unavailableReplicas, deployment.Status.UnavailableReplicas)
	}
	return nil
}
