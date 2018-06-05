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

package utils

import (
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"

	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	deploymentutil "k8s.io/kubernetes/pkg/controller/deployment/util"
	labelsutil "k8s.io/kubernetes/pkg/util/labels"
)

type LogfFn func(format string, args ...interface{})

func LogReplicaSetsOfDeployment(deployment *extensions.Deployment, allOldRSs []*extensions.ReplicaSet, newRS *extensions.ReplicaSet, logf LogfFn) {
	if newRS != nil {
		logf(spew.Sprintf("New ReplicaSet %q of Deployment %q:\n%+v", newRS.Name, deployment.Name, *newRS))
	} else {
		logf("New ReplicaSet of Deployment %q is nil.", deployment.Name)
	}
	if len(allOldRSs) > 0 {
		logf("All old ReplicaSets of Deployment %q:", deployment.Name)
	}
	for i := range allOldRSs {
		logf(spew.Sprintf("%+v", *allOldRSs[i]))
	}
}

func LogPodsOfDeployment(c clientset.Interface, deployment *extensions.Deployment, rsList []*extensions.ReplicaSet, logf LogfFn) {
	minReadySeconds := deployment.Spec.MinReadySeconds
	podListFunc := func(namespace string, options metav1.ListOptions) (*v1.PodList, error) {
		return c.CoreV1().Pods(namespace).List(options)
	}

	podList, err := deploymentutil.ListPods(deployment, rsList, podListFunc)
	if err != nil {
		logf("Failed to list Pods of Deployment %q: %v", deployment.Name, err)
		return
	}
	for _, pod := range podList.Items {
		availability := "not available"
		if podutil.IsPodAvailable(&pod, minReadySeconds, metav1.Now()) {
			availability = "available"
		}
		logf(spew.Sprintf("Pod %q is %s:\n%+v", pod.Name, availability, pod))
	}
}

// Waits for the deployment to complete.
// If during a rolling update (rolling == true), returns an error if the deployment's
// rolling update strategy (max unavailable or max surge) is broken at any times.
// It's not seen as a rolling update if shortly after a scaling event or the deployment is just created.
func waitForDeploymentCompleteMaybeCheckRolling(c clientset.Interface, d *extensions.Deployment, rolling bool, logf LogfFn, pollInterval, pollTimeout time.Duration) error {
	var (
		deployment *extensions.Deployment
		reason     string
	)

	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		var err error
		deployment, err = c.ExtensionsV1beta1().Deployments(d.Namespace).Get(d.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		// If during a rolling update, make sure rolling update strategy isn't broken at any times.
		if rolling {
			reason, err = checkRollingUpdateStatus(c, deployment, logf)
			if err != nil {
				return false, err
			}
			logf(reason)
		}

		// When the deployment status and its underlying resources reach the desired state, we're done
		if deploymentutil.DeploymentComplete(d, &deployment.Status) {
			return true, nil
		}

		reason = fmt.Sprintf("deployment status: %#v", deployment.Status)
		logf(reason)

		return false, nil
	})

	if err == wait.ErrWaitTimeout {
		err = fmt.Errorf("%s", reason)
	}
	if err != nil {
		return fmt.Errorf("error waiting for deployment %q status to match expectation: %v", d.Name, err)
	}
	return nil
}

func checkRollingUpdateStatus(c clientset.Interface, deployment *extensions.Deployment, logf LogfFn) (string, error) {
	var reason string
	oldRSs, allOldRSs, newRS, err := deploymentutil.GetAllReplicaSets(deployment, c.ExtensionsV1beta1())
	if err != nil {
		return "", err
	}
	if newRS == nil {
		// New RC hasn't been created yet.
		reason = "new replica set hasn't been created yet"
		return reason, nil
	}
	allRSs := append(oldRSs, newRS)
	// The old/new ReplicaSets need to contain the pod-template-hash label
	for i := range allRSs {
		if !labelsutil.SelectorHasLabel(allRSs[i].Spec.Selector, extensions.DefaultDeploymentUniqueLabelKey) {
			reason = "all replica sets need to contain the pod-template-hash label"
			return reason, nil
		}
	}

	// Check max surge and min available
	totalCreated := deploymentutil.GetReplicaCountForReplicaSets(allRSs)
	maxCreated := *(deployment.Spec.Replicas) + deploymentutil.MaxSurge(*deployment)
	if totalCreated > maxCreated {
		LogReplicaSetsOfDeployment(deployment, allOldRSs, newRS, logf)
		LogPodsOfDeployment(c, deployment, allRSs, logf)
		return "", fmt.Errorf("total pods created: %d, more than the max allowed: %d", totalCreated, maxCreated)
	}
	minAvailable := deploymentutil.MinAvailable(deployment)
	if deployment.Status.AvailableReplicas < minAvailable {
		LogReplicaSetsOfDeployment(deployment, allOldRSs, newRS, logf)
		LogPodsOfDeployment(c, deployment, allRSs, logf)
		return "", fmt.Errorf("total pods available: %d, less than the min required: %d", deployment.Status.AvailableReplicas, minAvailable)
	}
	return "", nil
}

// Waits for the deployment to complete, and check rolling update strategy isn't broken at any times.
// Rolling update strategy should not be broken during a rolling update.
func WaitForDeploymentCompleteAndCheckRolling(c clientset.Interface, d *extensions.Deployment, logf LogfFn, pollInterval, pollTimeout time.Duration) error {
	rolling := true
	return waitForDeploymentCompleteMaybeCheckRolling(c, d, rolling, logf, pollInterval, pollTimeout)
}

// Waits for the deployment to complete, and don't check if rolling update strategy is broken.
// Rolling update strategy is used only during a rolling update, and can be violated in other situations,
// such as shortly after a scaling event or the deployment is just created.
func WaitForDeploymentComplete(c clientset.Interface, d *extensions.Deployment, logf LogfFn, pollInterval, pollTimeout time.Duration) error {
	rolling := false
	return waitForDeploymentCompleteMaybeCheckRolling(c, d, rolling, logf, pollInterval, pollTimeout)
}

// WaitForDeploymentRevisionAndImage waits for the deployment's and its new RS's revision and container image to match the given revision and image.
// Note that deployment revision and its new RS revision should be updated shortly, so we only wait for 1 minute here to fail early.
func WaitForDeploymentRevisionAndImage(c clientset.Interface, ns, deploymentName string, revision, image string, logf LogfFn, pollInterval, pollTimeout time.Duration) error {
	var deployment *extensions.Deployment
	var newRS *extensions.ReplicaSet
	var reason string
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		var err error
		deployment, err = c.ExtensionsV1beta1().Deployments(ns).Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		// The new ReplicaSet needs to be non-nil and contain the pod-template-hash label
		newRS, err = deploymentutil.GetNewReplicaSet(deployment, c.ExtensionsV1beta1())
		if err != nil {
			return false, err
		}
		if err := checkRevisionAndImage(deployment, newRS, revision, image); err != nil {
			reason = err.Error()
			logf(reason)
			return false, nil
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		LogReplicaSetsOfDeployment(deployment, nil, newRS, logf)
		err = fmt.Errorf(reason)
	}
	if newRS == nil {
		return fmt.Errorf("deployment %q failed to create new replica set", deploymentName)
	}
	if err != nil {
		return fmt.Errorf("error waiting for deployment %q (got %s / %s) and new replica set %q (got %s / %s) revision and image to match expectation (expected %s / %s): %v", deploymentName, deployment.Annotations[deploymentutil.RevisionAnnotation], deployment.Spec.Template.Spec.Containers[0].Image, newRS.Name, newRS.Annotations[deploymentutil.RevisionAnnotation], newRS.Spec.Template.Spec.Containers[0].Image, revision, image, err)
	}
	return nil
}

// CheckDeploymentRevisionAndImage checks if the input deployment's and its new replica set's revision and image are as expected.
func CheckDeploymentRevisionAndImage(c clientset.Interface, ns, deploymentName, revision, image string) error {
	deployment, err := c.ExtensionsV1beta1().Deployments(ns).Get(deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get deployment %s during revision check: %v", deploymentName, err)
	}

	// Check revision of the new replica set of this deployment
	newRS, err := deploymentutil.GetNewReplicaSet(deployment, c.ExtensionsV1beta1())
	if err != nil {
		return fmt.Errorf("unable to get new replicaset of deployment %s during revision check: %v", deploymentName, err)
	}
	return checkRevisionAndImage(deployment, newRS, revision, image)
}

func checkRevisionAndImage(deployment *extensions.Deployment, newRS *extensions.ReplicaSet, revision, image string) error {
	// The new ReplicaSet needs to be non-nil and contain the pod-template-hash label
	if newRS == nil {
		return fmt.Errorf("new replicaset for deployment %q is yet to be created", deployment.Name)
	}
	if !labelsutil.SelectorHasLabel(newRS.Spec.Selector, extensions.DefaultDeploymentUniqueLabelKey) {
		return fmt.Errorf("new replica set %q doesn't have %q label selector", newRS.Name, extensions.DefaultDeploymentUniqueLabelKey)
	}
	// Check revision of this deployment, and of the new replica set of this deployment
	if deployment.Annotations == nil || deployment.Annotations[deploymentutil.RevisionAnnotation] != revision {
		return fmt.Errorf("deployment %q doesn't have the required revision set", deployment.Name)
	}
	if newRS.Annotations == nil || newRS.Annotations[deploymentutil.RevisionAnnotation] != revision {
		return fmt.Errorf("new replicaset %q doesn't have the required revision set", newRS.Name)
	}
	// Check the image of this deployment, and of the new replica set of this deployment
	if !containsImage(deployment.Spec.Template.Spec.Containers, image) {
		return fmt.Errorf("deployment %q doesn't have the required image %s set", deployment.Name, image)
	}
	if !containsImage(newRS.Spec.Template.Spec.Containers, image) {
		return fmt.Errorf("new replica set %q doesn't have the required image %s.", newRS.Name, image)
	}
	return nil
}

func containsImage(containers []v1.Container, imageName string) bool {
	for _, container := range containers {
		if container.Image == imageName {
			return true
		}
	}
	return false
}

type UpdateDeploymentFunc func(d *extensions.Deployment)

func UpdateDeploymentWithRetries(c clientset.Interface, namespace, name string, applyUpdate UpdateDeploymentFunc, logf LogfFn, pollInterval, pollTimeout time.Duration) (*extensions.Deployment, error) {
	var deployment *extensions.Deployment
	var updateErr error
	pollErr := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		var err error
		if deployment, err = c.ExtensionsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{}); err != nil {
			return false, err
		}
		// Apply the update, then attempt to push it to the apiserver.
		applyUpdate(deployment)
		if deployment, err = c.ExtensionsV1beta1().Deployments(namespace).Update(deployment); err == nil {
			logf("Updating deployment %s", name)
			return true, nil
		}
		updateErr = err
		return false, nil
	})
	if pollErr == wait.ErrWaitTimeout {
		pollErr = fmt.Errorf("couldn't apply the provided updated to deployment %q: %v", name, updateErr)
	}
	return deployment, pollErr
}

func WaitForObservedDeployment(c clientset.Interface, ns, deploymentName string, desiredGeneration int64) error {
	return deploymentutil.WaitForObservedDeployment(func() (*extensions.Deployment, error) {
		return c.ExtensionsV1beta1().Deployments(ns).Get(deploymentName, metav1.GetOptions{})
	}, desiredGeneration, 2*time.Second, 1*time.Minute)
}

// WaitForDeploymentRollbackCleared waits for given deployment either started rolling back or doesn't need to rollback.
func WaitForDeploymentRollbackCleared(c clientset.Interface, ns, deploymentName string, pollInterval, pollTimeout time.Duration) error {
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		deployment, err := c.ExtensionsV1beta1().Deployments(ns).Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		// Rollback not set or is kicked off
		if deployment.Spec.RollbackTo == nil {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("error waiting for deployment %s rollbackTo to be cleared: %v", deploymentName, err)
	}
	return nil
}

// WaitForDeploymentUpdatedReplicasGTE waits for given deployment to be observed by the controller and has at least a number of updatedReplicas
func WaitForDeploymentUpdatedReplicasGTE(c clientset.Interface, ns, deploymentName string, minUpdatedReplicas int32, desiredGeneration int64, pollInterval, pollTimeout time.Duration) error {
	var deployment *extensions.Deployment
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		d, err := c.ExtensionsV1beta1().Deployments(ns).Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		deployment = d
		return deployment.Status.ObservedGeneration >= desiredGeneration && deployment.Status.UpdatedReplicas >= minUpdatedReplicas, nil
	})
	if err != nil {
		return fmt.Errorf("error waiting for deployment %q to have at least %d updatedReplicas: %v; latest .status.updatedReplicas: %d", deploymentName, minUpdatedReplicas, err, deployment.Status.UpdatedReplicas)
	}
	return nil
}

func WaitForDeploymentWithCondition(c clientset.Interface, ns, deploymentName, reason string, condType extensions.DeploymentConditionType, logf LogfFn, pollInterval, pollTimeout time.Duration) error {
	var deployment *extensions.Deployment
	pollErr := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		d, err := c.ExtensionsV1beta1().Deployments(ns).Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		deployment = d
		cond := deploymentutil.GetDeploymentCondition(deployment.Status, condType)
		return cond != nil && cond.Reason == reason, nil
	})
	if pollErr == wait.ErrWaitTimeout {
		pollErr = fmt.Errorf("deployment %q never updated with the desired condition and reason, latest deployment conditions: %+v", deployment.Name, deployment.Status.Conditions)
		_, allOldRSs, newRS, err := deploymentutil.GetAllReplicaSets(deployment, c.ExtensionsV1beta1())
		if err == nil {
			LogReplicaSetsOfDeployment(deployment, allOldRSs, newRS, logf)
			LogPodsOfDeployment(c, deployment, append(allOldRSs, newRS), logf)
		}
	}
	return pollErr
}
