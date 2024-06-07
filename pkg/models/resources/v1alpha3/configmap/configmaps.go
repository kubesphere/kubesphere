/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package configmap

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type configmapsGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &configmapsGetter{cache: cache}
}

func (d *configmapsGetter) Get(namespace, name string) (runtime.Object, error) {
	configMap := &corev1.ConfigMap{}
	return configMap, d.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, configMap)
}

func (d *configmapsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	configMaps := &corev1.ConfigMapList{}
	if err := d.cache.List(context.Background(), configMaps, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range configMaps.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *configmapsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftCM, ok := left.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	rightCM, ok := right.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftCM.ObjectMeta, rightCM.ObjectMeta, field)
}

func (d *configmapsGetter) filter(object runtime.Object, filter query.Filter) bool {
	configMap, ok := object.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(configMap.ObjectMeta, filter)
}
