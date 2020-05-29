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

package persistentvolumeclaim

import (
	snapshotinformers "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"strconv"
	"strings"
)

const (
	storageClassName = "storageClassName"

	annotationInUse              = "kubesphere.io/in-use"
	annotationAllowSnapshot      = "kubesphere.io/allow-snapshot"
	annotationStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
)

type persistentVolumeClaimGetter struct {
	informers         informers.SharedInformerFactory
	snapshotInformers snapshotinformers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory, snapshotInformer snapshotinformers.SharedInformerFactory) v1alpha3.Interface {
	return &persistentVolumeClaimGetter{informers: informer, snapshotInformers: snapshotInformer}
}

func (p *persistentVolumeClaimGetter) Get(namespace, name string) (runtime.Object, error) {
	pvc, err := p.informers.Core().V1().PersistentVolumeClaims().Lister().PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		return pvc, err
	}
	p.annotatePVC(pvc)
	return pvc, nil
}

func (p *persistentVolumeClaimGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	all, err := p.informers.Core().V1().PersistentVolumeClaims().Lister().PersistentVolumeClaims(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, pvc := range all {
		p.annotatePVC(pvc)
		result = append(result, pvc)
	}
	return v1alpha3.DefaultList(result, query, p.compare, p.filter), nil
}

func (p *persistentVolumeClaimGetter) compare(left, right runtime.Object, field query.Field) bool {
	leftSnapshot, ok := left.(*v1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	rightSnapshot, ok := right.(*v1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftSnapshot.ObjectMeta, rightSnapshot.ObjectMeta, field)
}

func (p *persistentVolumeClaimGetter) filter(object runtime.Object, filter query.Filter) bool {
	pvc, ok := object.(*v1.PersistentVolumeClaim)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.EqualFold(string(pvc.Status.Phase), string(filter.Value))
	case storageClassName:
		return pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(pvc.ObjectMeta, filter)
	}
}

func (p *persistentVolumeClaimGetter) annotatePVC(pvc *v1.PersistentVolumeClaim) {
	inUse := p.countPods(pvc.Name, pvc.Namespace)
	isSnapshotAllow := p.isSnapshotAllowed(pvc.GetAnnotations()[annotationStorageProvisioner])
	if pvc.Annotations == nil {
		pvc.Annotations = make(map[string]string)
	}
	pvc.Annotations[annotationInUse] = strconv.FormatBool(inUse)
	pvc.Annotations[annotationAllowSnapshot] = strconv.FormatBool(isSnapshotAllow)
}

func (p *persistentVolumeClaimGetter) countPods(name, namespace string) bool {
	pods, err := p.informers.Core().V1().Pods().Lister().Pods(namespace).List(labels.Everything())
	if err != nil {
		return false
	}
	for _, pod := range pods {
		for _, pvc := range pod.Spec.Volumes {
			if pvc.PersistentVolumeClaim != nil && pvc.PersistentVolumeClaim.ClaimName == name {
				return true
			}
		}
	}
	return false
}

func (p *persistentVolumeClaimGetter) isSnapshotAllowed(provisioner string) bool {
	if len(provisioner) == 0 {
		return false
	}
	volumeSnapshotClasses, err := p.snapshotInformers.Snapshot().V1beta1().VolumeSnapshotClasses().Lister().List(labels.Everything())
	if err != nil {
		return false
	}
	for _, volumeSnapshotClass := range volumeSnapshotClasses {
		if volumeSnapshotClass.Driver == provisioner {
			return true
		}
	}
	return false
}
