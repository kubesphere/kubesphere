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
	snapshotinformer "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"strconv"

	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sort"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	storageClassName = "storageClassName"
)

type persistentVolumeClaimSearcher struct {
	informers         informers.SharedInformerFactory
	snapshotInformers snapshotinformer.SharedInformerFactory
}

func NewPersistentVolumeClaimSearcher(informers informers.SharedInformerFactory, snapshotInformer snapshotinformer.SharedInformerFactory) v1alpha2.Interface {
	return &persistentVolumeClaimSearcher{
		informers:         informers,
		snapshotInformers: snapshotInformer,
	}
}

func (s *persistentVolumeClaimSearcher) Get(namespace, name string) (interface{}, error) {
	return s.informers.Core().V1().PersistentVolumeClaims().Lister().PersistentVolumeClaims(namespace).Get(name)
}

func pvcStatus(item *v1.PersistentVolumeClaim) string {
	status := v1alpha2.StatusPending
	if item.Status.Phase == v1.ClaimPending {
		status = v1alpha2.StatusPending
	} else if item.Status.Phase == v1.ClaimBound {
		status = v1alpha2.StatusBound
	} else if item.Status.Phase == v1.ClaimLost {
		status = v1alpha2.StatusLost
	}
	return status
}

func (*persistentVolumeClaimSearcher) match(match map[string]string, item *v1.PersistentVolumeClaim) bool {
	for k, v := range match {
		switch k {
		case v1alpha2.Status:
			statuses := strings.Split(v, "|")
			if !sliceutil.HasString(statuses, pvcStatus(item)) {
				return false
			}
		case storageClassName:
			if item.Spec.StorageClassName == nil || *item.Spec.StorageClassName != v {
				return false
			}
		default:
			if !v1alpha2.ObjectMetaExactlyMath(k, v, item.ObjectMeta) {
				return false
			}
		}
	}
	return true
}

func (*persistentVolumeClaimSearcher) fuzzy(fuzzy map[string]string, item *v1.PersistentVolumeClaim) bool {
	for k, v := range fuzzy {
		if !v1alpha2.ObjectMetaFuzzyMath(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (s *persistentVolumeClaimSearcher) Search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	persistentVolumeClaims, err := s.informers.Core().V1().PersistentVolumeClaims().Lister().PersistentVolumeClaims(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1.PersistentVolumeClaim, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = persistentVolumeClaims
	} else {
		for _, item := range persistentVolumeClaims {
			if s.match(conditions.Match, item) && s.fuzzy(conditions.Fuzzy, item) {
				result = append(result, item)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			i, j = j, i
		}
		return v1alpha2.ObjectMetaCompare(result[i].ObjectMeta, result[j].ObjectMeta, orderBy)
	})

	r := make([]interface{}, 0)
	for _, i := range result {
		inUse := s.countPods(i.Name, i.Namespace)
		isSnapshotAllow := s.isSnapshotAllowed(i.GetAnnotations()["volume.beta.kubernetes.io/storage-provisioner"])
		if i.Annotations == nil {
			i.Annotations = make(map[string]string)
		}
		i.Annotations["kubesphere.io/in-use"] = strconv.FormatBool(inUse)
		i.Annotations["kubesphere.io/allow-snapshot"] = strconv.FormatBool(isSnapshotAllow)
		r = append(r, i)
	}
	return r, nil
}

func (s *persistentVolumeClaimSearcher) countPods(name, namespace string) bool {
	pods, err := s.informers.Core().V1().Pods().Lister().Pods(namespace).List(labels.Everything())
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

func (s *persistentVolumeClaimSearcher) isSnapshotAllowed(provisioner string) bool {
	if len(provisioner) == 0 {
		return false
	}
	volumeSnapshotClasses, err := s.snapshotInformers.Snapshot().V1beta1().VolumeSnapshotClasses().Lister().List(labels.Everything())
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
