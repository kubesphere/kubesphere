package volumesnapshotcontent

import (
	"strings"

	v1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	"github.com/kubernetes-csi/external-snapshotter/client/v4/informers/externalversions"

	"k8s.io/apimachinery/pkg/runtime"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	volumeSnapshotClassName = "volumeSnapshotClassName"
	volumeSnapshotName      = "volumeSnapshotName"
	volumeSnapshotNameSpace = "volumeSnapshotNamespace"
)

type volumesnapshotcontentGetter struct {
	informers externalversions.SharedInformerFactory
}

func New(informer externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &volumesnapshotcontentGetter{informers: informer}
}

func (v *volumesnapshotcontentGetter) Get(namespace, name string) (runtime.Object, error) {
	return v.informers.Snapshot().V1().VolumeSnapshotContents().Lister().Get(name)
}

func (v *volumesnapshotcontentGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	all, err := v.informers.Snapshot().V1().VolumeSnapshotContents().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, snapshotContent := range all {
		result = append(result, snapshotContent)
	}

	return v1alpha3.DefaultList(result, query, v.compare, v.filter), nil
}

func (v *volumesnapshotcontentGetter) compare(left, right runtime.Object, field query.Field) bool {
	leftSnapshotContent, ok := left.(*v1.VolumeSnapshotContent)
	if !ok {
		return false
	}
	rightSnapshotContent, ok := right.(*v1.VolumeSnapshotContent)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftSnapshotContent.ObjectMeta, rightSnapshotContent.ObjectMeta, field)
}

func (v *volumesnapshotcontentGetter) filter(object runtime.Object, filter query.Filter) bool {
	snapshotcontent, ok := object.(*v1.VolumeSnapshotContent)
	if !ok {
		return false
	}

	switch filter.Field {
	case volumeSnapshotClassName:
		return strings.EqualFold(*snapshotcontent.Spec.VolumeSnapshotClassName, string(filter.Value))
	case volumeSnapshotName:
		return strings.EqualFold(snapshotcontent.Spec.VolumeSnapshotRef.Name, string(filter.Value))
	case volumeSnapshotNameSpace:
		return strings.EqualFold(snapshotcontent.Spec.VolumeSnapshotRef.Namespace, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(snapshotcontent.ObjectMeta, filter)
	}
}
