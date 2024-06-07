/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package job

import (
	"context"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	jobFailed    = "failed"
	jobCompleted = "completed"
	jobRunning   = "running"
)

type jobsGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &jobsGetter{cache: cache}
}

func (d *jobsGetter) Get(namespace, name string) (runtime.Object, error) {
	job := &batchv1.Job{}
	return job, d.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, job)
}

func (d *jobsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	jobs := &batchv1.JobList{}
	if err := d.cache.List(context.Background(), jobs, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range jobs.Items {
		result = append(result, item.DeepCopy())
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
