/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package persistentvolumeclaim

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
	storageClassName             = "storageClassName"
	annotationInUse              = "kubesphere.io/in-use"
	annotationAllowSnapshot      = "kubesphere.io/allow-snapshot"
	annotationStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
)

type persistentVolumeClaimGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &persistentVolumeClaimGetter{cache: cache}
}

func (p *persistentVolumeClaimGetter) Get(namespace, name string) (runtime.Object, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	return pvc, p.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, pvc)
}

func (p *persistentVolumeClaimGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	persistentVolumeClaims := &corev1.PersistentVolumeClaimList{}
	if err := p.cache.List(context.Background(), persistentVolumeClaims, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range persistentVolumeClaims.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, p.compare, p.filter), nil
}

func (p *persistentVolumeClaimGetter) compare(left, right runtime.Object, field query.Field) bool {
	leftPVC, ok := left.(*corev1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	rightPVC, ok := right.(*corev1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftPVC.ObjectMeta, rightPVC.ObjectMeta, field)
}

func (p *persistentVolumeClaimGetter) filter(object runtime.Object, filter query.Filter) bool {
	pvc, ok := object.(*corev1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		statuses := strings.Split(string(filter.Value), "|")
		for _, status := range statuses {
			if !strings.EqualFold(string(pvc.Status.Phase), status) {
				return false
			}
		}
		return true
	case storageClassName:
		return pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(pvc.ObjectMeta, filter)
	}
}
