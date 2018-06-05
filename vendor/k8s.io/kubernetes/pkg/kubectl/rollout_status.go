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

package kubectl

import (
	"fmt"

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	clientappsv1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	clientextensionsv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/controller/deployment/util"
)

// StatusViewer provides an interface for resources that have rollout status.
type StatusViewer interface {
	Status(namespace, name string, revision int64) (string, bool, error)
}

// StatusViewerFor returns a StatusViewer for the resource specified by kind.
func StatusViewerFor(kind schema.GroupKind, c kubernetes.Interface) (StatusViewer, error) {
	switch kind {
	case extensionsv1beta1.SchemeGroupVersion.WithKind("Deployment").GroupKind(), apps.Kind("Deployment"):
		return &DeploymentStatusViewer{c.ExtensionsV1beta1()}, nil
	case extensionsv1beta1.SchemeGroupVersion.WithKind("DaemonSet").GroupKind(), apps.Kind("DaemonSet"):
		return &DaemonSetStatusViewer{c.ExtensionsV1beta1()}, nil
	case apps.Kind("StatefulSet"):
		return &StatefulSetStatusViewer{c.AppsV1beta1()}, nil
	}
	return nil, fmt.Errorf("no status viewer has been implemented for %v", kind)
}

// DeploymentStatusViewer implements the StatusViewer interface.
type DeploymentStatusViewer struct {
	c clientextensionsv1beta1.DeploymentsGetter
}

// DaemonSetStatusViewer implements the StatusViewer interface.
type DaemonSetStatusViewer struct {
	c clientextensionsv1beta1.DaemonSetsGetter
}

// StatefulSetStatusViewer implements the StatusViewer interface.
type StatefulSetStatusViewer struct {
	c clientappsv1beta1.StatefulSetsGetter
}

// Status returns a message describing deployment status, and a bool value indicating if the status is considered done.
func (s *DeploymentStatusViewer) Status(namespace, name string, revision int64) (string, bool, error) {
	deployment, err := s.c.Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", false, err
	}
	if revision > 0 {
		deploymentRev, err := util.Revision(deployment)
		if err != nil {
			return "", false, fmt.Errorf("cannot get the revision of deployment %q: %v", deployment.Name, err)
		}
		if revision != deploymentRev {
			return "", false, fmt.Errorf("desired revision (%d) is different from the running revision (%d)", revision, deploymentRev)
		}
	}
	if deployment.Generation <= deployment.Status.ObservedGeneration {
		cond := util.GetDeploymentCondition(deployment.Status, extensionsv1beta1.DeploymentProgressing)
		if cond != nil && cond.Reason == util.TimedOutReason {
			return "", false, fmt.Errorf("deployment %q exceeded its progress deadline", name)
		}
		if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
			return fmt.Sprintf("Waiting for rollout to finish: %d out of %d new replicas have been updated...\n", deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas), false, nil
		}
		if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
			return fmt.Sprintf("Waiting for rollout to finish: %d old replicas are pending termination...\n", deployment.Status.Replicas-deployment.Status.UpdatedReplicas), false, nil
		}
		if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
			return fmt.Sprintf("Waiting for rollout to finish: %d of %d updated replicas are available...\n", deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas), false, nil
		}
		return fmt.Sprintf("deployment %q successfully rolled out\n", name), true, nil
	}
	return fmt.Sprintf("Waiting for deployment spec update to be observed...\n"), false, nil
}

// Status returns a message describing daemon set status, and a bool value indicating if the status is considered done.
func (s *DaemonSetStatusViewer) Status(namespace, name string, revision int64) (string, bool, error) {
	//ignoring revision as DaemonSets does not have history yet

	daemon, err := s.c.DaemonSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", false, err
	}
	if daemon.Spec.UpdateStrategy.Type != extensionsv1beta1.RollingUpdateDaemonSetStrategyType {
		return "", true, fmt.Errorf("Status is available only for RollingUpdate strategy type")
	}
	if daemon.Generation <= daemon.Status.ObservedGeneration {
		if daemon.Status.UpdatedNumberScheduled < daemon.Status.DesiredNumberScheduled {
			return fmt.Sprintf("Waiting for rollout to finish: %d out of %d new pods have been updated...\n", daemon.Status.UpdatedNumberScheduled, daemon.Status.DesiredNumberScheduled), false, nil
		}
		if daemon.Status.NumberAvailable < daemon.Status.DesiredNumberScheduled {
			return fmt.Sprintf("Waiting for rollout to finish: %d of %d updated pods are available...\n", daemon.Status.NumberAvailable, daemon.Status.DesiredNumberScheduled), false, nil
		}
		return fmt.Sprintf("daemon set %q successfully rolled out\n", name), true, nil
	}
	return fmt.Sprintf("Waiting for daemon set spec update to be observed...\n"), false, nil
}

// Status returns a message describing statefulset status, and a bool value indicating if the status is considered done.
func (s *StatefulSetStatusViewer) Status(namespace, name string, revision int64) (string, bool, error) {
	sts, err := s.c.StatefulSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", false, err
	}
	if sts.Spec.UpdateStrategy.Type == apps.OnDeleteStatefulSetStrategyType {
		return "", true, fmt.Errorf("%s updateStrategy does not have a Status`", apps.OnDeleteStatefulSetStrategyType)
	}
	if sts.Status.ObservedGeneration == nil || sts.Generation > *sts.Status.ObservedGeneration {
		return "Waiting for statefulset spec update to be observed...\n", false, nil
	}
	if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas {
		return fmt.Sprintf("Waiting for %d pods to be ready...\n", *sts.Spec.Replicas-sts.Status.ReadyReplicas), false, nil
	}
	if sts.Spec.UpdateStrategy.Type == apps.RollingUpdateStatefulSetStrategyType && sts.Spec.UpdateStrategy.RollingUpdate != nil {
		if sts.Spec.Replicas != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			if sts.Status.UpdatedReplicas < (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition) {
				return fmt.Sprintf("Waiting for partitioned roll out to finish: %d out of %d new pods have been updated...\n",
					sts.Status.UpdatedReplicas, (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition)), false, nil
			}
		}
		return fmt.Sprintf("partitioned roll out complete: %d new pods have been updated...\n",
			sts.Status.UpdatedReplicas), true, nil
	}
	if sts.Status.UpdateRevision != sts.Status.CurrentRevision {
		return fmt.Sprintf("waiting for statefulset rolling update to complete %d pods at revision %s...\n",
			sts.Status.UpdatedReplicas, sts.Status.UpdateRevision), false, nil
	}
	return fmt.Sprintf("statefulset rolling update complete %d pods at revision %s...\n", sts.Status.CurrentReplicas, sts.Status.CurrentRevision), true, nil

}
