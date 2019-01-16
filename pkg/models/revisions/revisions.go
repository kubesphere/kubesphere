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

	"github.com/golang/glog"
	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	lister "k8s.io/client-go/listers/apps/v1"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/informers"
)

var (
	daemonSetLister          lister.DaemonSetLister
	deploymentLister         lister.DeploymentLister
	replicaSetLister         lister.ReplicaSetLister
	statefulSetLister        lister.StatefulSetLister
	controllerRevisionLister lister.ControllerRevisionLister
)

func init() {
	daemonSetLister = informers.SharedInformerFactory().Apps().V1().DaemonSets().Lister()
	deploymentLister = informers.SharedInformerFactory().Apps().V1().Deployments().Lister()
	replicaSetLister = informers.SharedInformerFactory().Apps().V1().ReplicaSets().Lister()
	statefulSetLister = informers.SharedInformerFactory().Apps().V1().StatefulSets().Lister()
	controllerRevisionLister = informers.SharedInformerFactory().Apps().V1().ControllerRevisions().Lister()
}

func GetDeployRevision(namespace, name, revision string) (*v1.ReplicaSet, error) {
	deploy, err := deploymentLister.Deployments(namespace).Get(name)
	if err != nil {
		glog.Errorf("get deployment %s failed, reason: %s", name, err)
		return nil, err
	}

	labelMap := deploy.Spec.Template.Labels
	labelSelector := labels.Set(labelMap).AsSelector()

	rsList, err := replicaSetLister.ReplicaSets(namespace).List(labelSelector)
	if err != nil {
		return nil, err
	}

	for _, rs := range rsList {
		if rs.Annotations["deployment.kubernetes.io/revision"] == revision {
			return rs, nil
		}
	}

	return nil, errors.New(errors.NotFound, fmt.Sprintf("revision not found %v#%v", name, revision))
}

func GetDaemonSetRevision(namespace, name string, revisionInt int) (*v1.ControllerRevision, error) {

	ds, err := daemonSetLister.DaemonSets(namespace).Get(name)

	if err != nil {
		return nil, err
	}

	labels := ds.Spec.Template.Labels

	return getControllerRevision(namespace, name, labels, revisionInt)
}

func GetStatefulSetRevision(namespace, name string, revisionInt int) (*v1.ControllerRevision, error) {

	st, err := statefulSetLister.StatefulSets(namespace).Get(name)

	if err != nil {
		return nil, err
	}

	return getControllerRevision(namespace, name, st.Spec.Template.Labels, revisionInt)
}

func getControllerRevision(namespace, name string, labelMap map[string]string, revision int) (*v1.ControllerRevision, error) {

	labelSelector := labels.Set(labelMap).AsSelector()

	revisions, err := controllerRevisionLister.ControllerRevisions(namespace).List(labelSelector)

	if err != nil {
		return nil, err
	}

	for _, controllerRevision := range revisions {
		if controllerRevision.Revision == int64(revision) {
			return controllerRevision, nil
		}
	}

	return nil, errors.New(errors.NotFound, fmt.Sprintf("revision not found %v#%v", name, revision))
}
