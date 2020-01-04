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
package deployment

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/api/apps/v1"
)

const (
	applicationLabel = "app.kubernetes.io/name"
	ReleaseLabel     = "relase"

	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type deploymentsGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &deploymentsGetter{sharedInformers: sharedInformers}
}

func (d *deploymentsGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.sharedInformers.Apps().V1().Deployments().Lister().Deployments(namespace).Get(name)
}

func (d *deploymentsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	// first retrieves all deployments within given namespace
	all, err := d.sharedInformers.Apps().V1().Deployments().Lister().Deployments(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, deploy := range all {
		result = append(result, deploy)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *deploymentsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftDeployment, ok := left.(*v1.Deployment)
	if !ok {
		return false
	}

	rightDeployment, ok := right.(*v1.Deployment)
	if !ok {
		return false
	}

	switch field {
	case query.FieldCreationTimeStamp:
		return leftDeployment.CreationTimestamp.After(rightDeployment.CreationTimestamp.Time)
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftDeployment).After(lastUpdateTime(rightDeployment))
	default:
		fallthrough
	case query.FieldName:
		return strings.Compare(leftDeployment.Name, rightDeployment.Name) > 0
	}
}

func (d *deploymentsGetter) filter(object runtime.Object, filter query.Filter) bool {
	deployment, ok := object.(*v1.Deployment)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldName:
		return query.ComparableString(deployment.Name).Contains(filter.Value)
	case query.FieldApplication:
		if app, ok := deployment.Labels[applicationLabel]; ok {
			return query.ComparableString(app).Contains(filter.Value)
		}
	case query.FieldStatus:
		return filter.Value.Compare(query.ComparableString(deploymentStatus(deployment.Status))) == 0
	default:
		return false
	}
	return false
}

func deploymentStatus(status v1.DeploymentStatus) string {
	if status.ReadyReplicas == 0 && status.Replicas == 0 {
		return statusStopped
	} else if status.ReadyReplicas == status.Replicas {
		return statusRunning
	} else {
		return statusUpdating
	}
}

func lastUpdateTime(deployment *v1.Deployment) time.Time {
	lut := deployment.CreationTimestamp.Time
	for _, condition := range deployment.Status.Conditions {
		if condition.LastUpdateTime.After(lut) {
			lut = condition.LastUpdateTime.Time
		}
	}
	return lut
}
