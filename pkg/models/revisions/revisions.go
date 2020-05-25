/*
Copyright 2019 The KubeSphere Authors.

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

package revisions

import (
	"fmt"
	"k8s.io/client-go/informers"
	"k8s.io/klog"

	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type RevisionGetter interface {
	GetDeploymentRevision(namespace, name, revision string) (*v1.ReplicaSet, error)
	GetStatefulSetRevision(namespace, name string, revision int) (*v1.ControllerRevision, error)
	GetDaemonSetRevision(namespace, name string, revision int) (*v1.ControllerRevision, error)
}

type revisionGetter struct {
	informers informers.SharedInformerFactory
}

func NewRevisionGetter(informers informers.SharedInformerFactory) RevisionGetter {
	return &revisionGetter{informers: informers}
}

func (c *revisionGetter) GetDeploymentRevision(namespace, name, revision string) (*v1.ReplicaSet, error) {
	deploymentLister := c.informers.Apps().V1().Deployments().Lister()
	deploy, err := deploymentLister.Deployments(namespace).Get(name)
	if err != nil {
		klog.Errorf("get deployment %s failed, reason: %s", name, err)
		return nil, err
	}

	labelMap := deploy.Spec.Template.Labels
	labelSelector := labels.Set(labelMap).AsSelector()

	replicaSetLister := c.informers.Apps().V1().ReplicaSets().Lister()
	rsList, err := replicaSetLister.ReplicaSets(namespace).List(labelSelector)
	if err != nil {
		return nil, err
	}

	for _, rs := range rsList {
		if rs.Annotations["deployment.kubernetes.io/revision"] == revision {
			return rs, nil
		}
	}

	return nil, fmt.Errorf("revision not found %v#%v", name, revision)
}

func (c *revisionGetter) GetDaemonSetRevision(namespace, name string, revisionInt int) (*v1.ControllerRevision, error) {
	daemonSetLister := c.informers.Apps().V1().DaemonSets().Lister()
	ds, err := daemonSetLister.DaemonSets(namespace).Get(name)

	if err != nil {
		return nil, err
	}

	lbs := ds.Spec.Template.Labels

	return c.getControllerRevision(namespace, name, lbs, revisionInt)
}

func (c *revisionGetter) GetStatefulSetRevision(namespace, name string, revisionInt int) (*v1.ControllerRevision, error) {
	statefulSetLister := c.informers.Apps().V1().StatefulSets().Lister()
	st, err := statefulSetLister.StatefulSets(namespace).Get(name)

	if err != nil {
		return nil, err
	}

	return c.getControllerRevision(namespace, name, st.Spec.Template.Labels, revisionInt)
}

func (c *revisionGetter) getControllerRevision(namespace, name string, labelMap map[string]string, revision int) (*v1.ControllerRevision, error) {

	labelSelector := labels.Set(labelMap).AsSelector()
	controllerRevisionLister := c.informers.Apps().V1().ControllerRevisions().Lister()
	revisions, err := controllerRevisionLister.ControllerRevisions(namespace).List(labelSelector)

	if err != nil {
		return nil, err
	}

	for _, controllerRevision := range revisions {
		if controllerRevision.Revision == int64(revision) {
			return controllerRevision, nil
		}
	}

	return nil, fmt.Errorf("revision not found %v#%v", name, revision)
}
