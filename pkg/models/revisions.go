/*
Copyright 2018 The KubeSphere Authors.

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

package models

import (
	"fmt"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	appsV1 "k8s.io/client-go/listers/apps/v1"

	"kubesphere.io/kubesphere/pkg/models/controllers"
)

func GetDeployRevision(namespace, name, revision string) (*v1.ReplicaSet, error) {
	deployLister := controllers.ResourceControllers.Controllers[controllers.Deployments].Lister().(appsV1.DeploymentLister)
	deploy, err := deployLister.Deployments(namespace).Get(name)
	if err != nil {
		glog.Errorf("get deployment %s failed, reason: %s", name, err)
		return nil, err
	}

	labelMap := deploy.Spec.Template.Labels
	labelSelector := labels.Set(labelMap).AsSelector()

	rsLister := controllers.ResourceControllers.Controllers[controllers.Replicasets].Lister().(appsV1.ReplicaSetLister)
	rsList, err := rsLister.ReplicaSets(namespace).List(labelSelector)
	if err != nil {
		return nil, err
	}

	for _, rs := range rsList {
		if rs.Annotations["deployment.kubernetes.io/revision"] == revision {
			return rs, nil
		}
	}

	return nil, errors.NewNotFound(v1.Resource("deployment revision"), fmt.Sprintf("%s#%s", name, revision))
}

func GetDaemonSetRevision(namespace, name, revision string) (*v1.ControllerRevision, error) {
	revisionInt, err := strconv.Atoi(revision)
	if err != nil {
		return nil, err
	}

	dsLister := controllers.ResourceControllers.Controllers[controllers.Daemonsets].Lister().(appsV1.DaemonSetLister)
	ds, err := dsLister.DaemonSets(namespace).Get(name)
	if err != nil {
		glog.Errorf("get Daemonset %s failed, reason: %s", name, err)
		return nil, err
	}

	labels := ds.Spec.Template.Labels

	return getControllerRevision(namespace, name, labels, revisionInt)
}

func GetStatefulSetRevision(namespace, name, revision string) (*v1.ControllerRevision, error) {
	revisionInt, err := strconv.Atoi(revision)
	if err != nil {
		return nil, err
	}

	stLister := controllers.ResourceControllers.Controllers[controllers.Statefulsets].Lister().(appsV1.StatefulSetLister)
	st, err := stLister.StatefulSets(namespace).Get(name)
	if err != nil {
		glog.Errorf("get Daemonset %s failed, reason: %s", name, err)
		return nil, err
	}

	labels := st.Spec.Template.Labels

	return getControllerRevision(namespace, name, labels, revisionInt)
}

func getControllerRevision(namespace, name string, labelMap map[string]string, revision int) (*v1.ControllerRevision, error) {

	labelSelector := labels.Set(labelMap).AsSelector()

	revisionLister := controllers.ResourceControllers.Controllers[controllers.ControllerRevisions].Lister().(appsV1.ControllerRevisionLister)
	revisions, err := revisionLister.ControllerRevisions(namespace).List(labelSelector)
	if err != nil {
		return nil, err
	}

	for _, controllerRevision := range revisions {
		if controllerRevision.Revision == int64(revision) {
			return controllerRevision, nil
		}
	}

	return nil, errors.NewNotFound(v1.Resource("revision"), fmt.Sprintf("%s#%s", name, revision))

}
