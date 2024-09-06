/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package storageclass

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type storageClassesGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &storageClassesGetter{
		cache: cache,
	}
}

func (s *storageClassesGetter) Get(_, name string) (runtime.Object, error) {
	storageClass := &v1.StorageClass{}
	if err := s.cache.Get(context.Background(), types.NamespacedName{Name: name}, storageClass); err != nil {
		return nil, err
	}
	return s.transform(storageClass), nil
}

func (s *storageClassesGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	storageClasses := &v1.StorageClassList{}
	if err := s.cache.List(context.Background(), storageClasses, client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range storageClasses.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, s.compare, s.filter, s.transform), nil
}

func (s *storageClassesGetter) transform(obj runtime.Object) runtime.Object {
	in := obj.(*v1.StorageClass)
	out := in.DeepCopy()
	count := s.countPersistentVolumeClaims(in.Name)
	if out.Annotations == nil {
		out.Annotations = make(map[string]string)
	}
	out.Annotations["kubesphere.io/pvc-count"] = strconv.Itoa(count)
	return out
}

func (s *storageClassesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftStorageClass, ok := left.(*clusterv1alpha1.Cluster)
	if !ok {
		return false
	}

	rightStorageClass, ok := right.(*clusterv1alpha1.Cluster)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftStorageClass.ObjectMeta, rightStorageClass.ObjectMeta, field)
}

func (s *storageClassesGetter) filter(object runtime.Object, filter query.Filter) bool {
	cluster, ok := object.(*clusterv1alpha1.Cluster)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(cluster.ObjectMeta, filter)
}

func (s *storageClassesGetter) countPersistentVolumeClaims(name string) int {
	pvcs := &corev1.PersistentVolumeClaimList{}
	if err := s.cache.List(context.Background(), pvcs); err != nil {
		return 0
	}
	var count int
	for _, pvc := range pvcs.Items {
		if (pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == name) ||
			(pvc.Annotations != nil && pvc.Annotations[corev1.BetaStorageClassAnnotation] == name) {
			count++
		}
	}
	return count
}
