/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package revisions

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type RevisionGetter interface {
	GetDeploymentRevision(namespace, name, revision string) (*appsv1.ReplicaSet, error)
	GetStatefulSetRevision(namespace, name string, revision int) (*appsv1.ControllerRevision, error)
	GetDaemonSetRevision(namespace, name string, revision int) (*appsv1.ControllerRevision, error)
}

type revisionGetter struct {
	cache runtimeclient.Reader
}

func NewRevisionGetter(cacheReader runtimeclient.Reader) RevisionGetter {
	return &revisionGetter{cache: cacheReader}
}

func (c *revisionGetter) GetDeploymentRevision(namespace, name, revision string) (*appsv1.ReplicaSet, error) {
	deployment := &appsv1.Deployment{}
	if err := c.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, deployment); err != nil {
		klog.Errorf("get deployment %s failed, reason: %s", name, err)
		return nil, err
	}
	replicaSetList := &appsv1.ReplicaSetList{}
	if err := c.cache.List(context.Background(), replicaSetList, client.InNamespace(namespace), client.MatchingLabels(deployment.Spec.Template.Labels)); err != nil {
		klog.Errorf("get deployment %s failed, reason: %s", name, err)
		return nil, err
	}

	for _, rs := range replicaSetList.Items {
		result := rs.DeepCopy()
		if result.Annotations["deployment.kubernetes.io/revision"] == revision {
			return result, nil
		}
	}

	return nil, fmt.Errorf("revision not found %v#%v", name, revision)
}

func (c *revisionGetter) GetDaemonSetRevision(namespace, name string, revision int) (*appsv1.ControllerRevision, error) {
	daemonSet := &appsv1.DaemonSet{}
	if err := c.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, daemonSet); err != nil {
		klog.Errorf("get deployment %s failed, reason: %s", name, err)
		return nil, err
	}
	return c.getControllerRevision(namespace, name, daemonSet.Spec.Template.Labels, revision)
}

func (c *revisionGetter) GetStatefulSetRevision(namespace, name string, revisionInt int) (*appsv1.ControllerRevision, error) {
	statefulSet := &appsv1.StatefulSet{}
	if err := c.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, statefulSet); err != nil {
		klog.Errorf("get deployment %s failed, reason: %s", name, err)
		return nil, err
	}
	return c.getControllerRevision(namespace, name, statefulSet.Spec.Template.Labels, revisionInt)
}

func (c *revisionGetter) getControllerRevision(namespace, name string, labelMap map[string]string, revision int) (*appsv1.ControllerRevision, error) {
	controllerRevisionList := &appsv1.ControllerRevisionList{}
	if err := c.cache.List(context.Background(), controllerRevisionList, client.InNamespace(namespace), client.MatchingLabels(labelMap)); err != nil {
		return nil, err
	}
	for _, controllerRevision := range controllerRevisionList.Items {
		if controllerRevision.Revision == int64(revision) {
			return controllerRevision.DeepCopy(), nil
		}
	}
	return nil, fmt.Errorf("revision not found %v#%v", name, revision)
}
