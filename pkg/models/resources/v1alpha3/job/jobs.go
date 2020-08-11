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

package job

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"strings"
	"time"
)

const (
	jobFailed    = "failed"
	jobCompleted = "completed"
	jobRunning   = "running"
)

type jobsGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &jobsGetter{sharedInformers: sharedInformers}
}

func (d *jobsGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.sharedInformers.Batch().V1().Jobs().Lister().Jobs(namespace).Get(name)
}

func (d *jobsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	jobs, err := d.sharedInformers.Batch().V1().Jobs().Lister().Jobs(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, job := range jobs {
		result = append(result, job)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *jobsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftJob, ok := left.(*batchv1.Job)
	if !ok {
		return false
	}

	rightJob, ok := right.(*batchv1.Job)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftJob).After(lastUpdateTime(rightJob))
	case query.FieldStatus:
		return strings.Compare(jobStatus(leftJob.Status), jobStatus(rightJob.Status)) > 0
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftJob.ObjectMeta, rightJob.ObjectMeta, field)
	}
}

func (d *jobsGetter) filter(object runtime.Object, filter query.Filter) bool {
	job, ok := object.(*batchv1.Job)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(jobStatus(job.Status), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(job.ObjectMeta, filter)
	}
}

func jobStatus(status batchv1.JobStatus) string {
	for _, condition := range status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			return jobCompleted
		} else if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			return jobFailed
		}
	}

	return jobRunning
}

func lastUpdateTime(job *batchv1.Job) time.Time {
	lut := job.CreationTimestamp.Time
	for _, condition := range job.Status.Conditions {
		if condition.LastTransitionTime.After(lut) {
			lut = condition.LastTransitionTime.Time
		}
	}
	return lut
}
