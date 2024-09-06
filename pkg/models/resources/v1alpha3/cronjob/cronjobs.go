/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cronjob

import (
	"context"
	"strings"

	"github.com/Masterminds/semver/v3"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/scheme"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

type cronJobsGetter struct {
	cache   runtimeclient.Reader
	gvk     schema.GroupVersionKind
	listGVK schema.GroupVersionKind
}

func New(cache runtimeclient.Reader, k8sVersion *semver.Version) v1alpha3.Interface {
	gvk := batchv1.SchemeGroupVersion.WithKind("CronJob")
	listGVK := schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind + "List",
	}
	if k8sutil.ServeBatchV1beta1(k8sVersion) {
		gvk.Version = "v1beta1"
		listGVK.Version = "v1beta1"
	}
	return &cronJobsGetter{
		cache:   cache,
		gvk:     gvk,
		listGVK: listGVK,
	}
}

func (d *cronJobsGetter) Get(namespace, name string) (runtime.Object, error) {
	obj, err := scheme.Scheme.New(d.gvk)
	if err != nil {
		return nil, err
	}
	cronJob := obj.(client.Object)
	return cronJob, d.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, cronJob)
}

func (d *cronJobsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	listObj := &unstructured.UnstructuredList{}
	listObj.SetGroupVersionKind(d.listGVK)

	if err := d.cache.List(context.Background(), listObj, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	if err := listObj.EachListItem(func(object runtime.Object) error {
		result = append(result, object)
		return nil
	}); err != nil {
		return nil, err
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func cronJobStatus(item *unstructured.Unstructured) string {
	suspend, _, _ := unstructured.NestedBool(item.UnstructuredContent(), "spec", "suspend")
	if suspend {
		return "paused"
	}
	return "running"
}

func (d *cronJobsGetter) filter(object runtime.Object, filter query.Filter) bool {
	job, ok := object.(*unstructured.Unstructured)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(cronJobStatus(job), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(k8sutil.GetObjectMeta(job), filter)
	}
}

func (d *cronJobsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftJob, ok := left.(*unstructured.Unstructured)
	if !ok {
		return false
	}

	rightJob, ok := right.(*unstructured.Unstructured)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(k8sutil.GetObjectMeta(leftJob), k8sutil.GetObjectMeta(rightJob), field)
}
