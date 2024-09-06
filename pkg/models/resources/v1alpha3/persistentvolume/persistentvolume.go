/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package persistentvolume

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	storageClassName = "storageClassName"
)

type persistentVolumeGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &persistentVolumeGetter{cache: cache}
}

func (p *persistentVolumeGetter) Get(_, name string) (runtime.Object, error) {
	pv := &corev1.PersistentVolume{}
	return pv, p.cache.Get(context.Background(), types.NamespacedName{Name: name}, pv)
}

func (p *persistentVolumeGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	persistentVolumes := &corev1.PersistentVolumeList{}
	if err := p.cache.List(context.Background(), persistentVolumes, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range persistentVolumes.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, p.compare, p.filter), nil
}

func (p *persistentVolumeGetter) compare(obj1, obj2 runtime.Object, field query.Field) bool {
	pv1, ok := obj1.(*corev1.PersistentVolume)
	if !ok {
		return false
	}
	pv2, ok := obj2.(*corev1.PersistentVolume)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(pv1.ObjectMeta, pv2.ObjectMeta, field)
}

func (p *persistentVolumeGetter) filter(object runtime.Object, filter query.Filter) bool {
	pv, ok := object.(*corev1.PersistentVolume)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.EqualFold(string(pv.Status.Phase), string(filter.Value))
	case storageClassName:
		return pv.Spec.StorageClassName != "" && pv.Spec.StorageClassName == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(pv.ObjectMeta, filter)
	}
}
