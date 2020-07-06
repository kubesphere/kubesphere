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

package volumesnapshot

import (
	"encoding/json"
	"github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	"github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/clientset/versioned/fake"
	"github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"testing"
	"time"
)

const (
	baseVolumeSnapshot = `{
    "apiVersion": "snapshot.storage.k8s.io/v1beta1",
    "kind": "VolumeSnapshot",
    "metadata": {
        "creationTimestamp": "2020-04-29T06:52:06Z",
        "finalizers": [
            "snapshot.storage.kubernetes.io/volumesnapshot-as-source-protection",
            "snapshot.storage.kubernetes.io/volumesnapshot-bound-protection"
        ],
        "generation": 1,
        "name": "snap-1",
        "namespace": "default",
        "resourceVersion": "5027277",
        "selfLink": "/apis/snapshot.storage.k8s.io/v1beta1/namespaces/default/volumesnapshots/snap-1",
        "uid": "dc66842d-17bf-4087-a8e8-7592d129a956"
    },
    "spec": {
        "source": {
            "persistentVolumeClaimName": "pvc-source"
        },
        "volumeSnapshotClassName": "csi-neonsan"
    },
    "status": {
        "boundVolumeSnapshotContentName": "snapcontent-dc66842d-17bf-4087-a8e8-7592d129a956",
        "creationTime": "2020-04-29T06:52:06Z",
        "readyToUse": true,
        "restoreSize": "20Gi"
    }
}`
	defaultNamespace = "default"
)

func newVolumeSnapshot(name string) *v1beta1.VolumeSnapshot {
	volumeSnapshot := &v1beta1.VolumeSnapshot{}
	err := json.Unmarshal([]byte(baseVolumeSnapshot), volumeSnapshot)
	if err != nil {
		return nil
	}
	volumeSnapshot.Name = name
	return volumeSnapshot
}

func TestListVolumeSnapshot(t *testing.T) {
	RegisterFailHandler(Fail)

	client := fake.NewSimpleClientset()
	informer := externalversions.NewSharedInformerFactory(client, 0)

	pvcName1, pvcName2, pvcName3 := "pvc-1", "pvc-2", "pvc-3"
	snapshot1 := newVolumeSnapshot("snap-1")
	snapshot1.CreationTimestamp = v1.NewTime(snapshot1.CreationTimestamp.Add(time.Hour * 3))
	snapshot1.Spec.Source.PersistentVolumeClaimName = &pvcName1

	snapshot2 := newVolumeSnapshot("snap-2")
	snapshot2.CreationTimestamp = v1.NewTime(snapshot2.CreationTimestamp.Add(time.Hour * 2))
	snapshot2.Spec.Source.PersistentVolumeClaimName = &pvcName2

	snapshot3 := newVolumeSnapshot("snap-3")
	snapshot3.CreationTimestamp = v1.NewTime(snapshot3.CreationTimestamp.Add(time.Hour))
	snapshot3.Spec.Source.PersistentVolumeClaimName = &pvcName3
	readyToUse := false
	snapshot3.Status.ReadyToUse = &readyToUse
	volumeSnapshotClassNameTest := "csi.aws.com"
	snapshot3.Spec.VolumeSnapshotClassName = &volumeSnapshotClassNameTest

	volumeSnapshots := []interface{}{snapshot1, snapshot2, snapshot3}

	for _, s := range volumeSnapshots {
		_ = informer.Snapshot().V1beta1().VolumeSnapshots().Informer().GetIndexer().Add(s)
	}
	getter := New(informer)

	Describe("condition", func() {
		It("match name", func() {
			query1 := query.New()
			query1.Filters[query.FieldName] = query.Value(snapshot1.Name)
			snapshotList, err := getter.List(defaultNamespace, query1)
			Expect(err).To(BeNil())
			Expect(snapshotList.TotalItems).To(Equal(1))
			Expect(snapshotList.Items[0]).To(Equal(snapshot1))
		})

		It("match persistentVolumeClaimName", func() {
			query1 := query.New()
			query1.Filters[persistentVolumeClaimName] = query.Value(*snapshot2.Spec.Source.PersistentVolumeClaimName)
			snapshotList, err := getter.List(defaultNamespace, query1)
			Expect(err).To(BeNil())
			Expect(snapshotList.TotalItems).To(Equal(1))
			Expect(snapshotList.Items[0]).To(Equal(snapshot2))
		})

		It("match status", func() {
			query1 := query.New()
			query1.Filters[query.FieldStatus] = query.Value(statusCreating)
			snapshotList, err := getter.List(defaultNamespace, query1)
			Expect(err).To(BeNil())
			Expect(snapshotList.TotalItems).To(Equal(1))
			Expect(snapshotList.Items[0]).To(Equal(snapshot3))
		})

		It("match volumeSnapshotClassName", func() {
			query1 := query.New()
			query1.Filters[volumeSnapshotClassName] = query.Value(*snapshot3.Spec.VolumeSnapshotClassName)
			snapshotList, err := getter.List(defaultNamespace, query1)
			Expect(err).To(BeNil())
			Expect(snapshotList.TotalItems).To(Equal(1))
			Expect(snapshotList.Items[0]).To(Equal(snapshot3))
		})

	})

	Describe("order", func() {
		It("by createTime ", func() {
			query1 := query.New()
			query1.SortBy = query.FieldCreateTime
			query1.Ascending = true
			snapshotList, err := getter.List(defaultNamespace, query1)
			Expect(err).To(BeNil())
			Expect(snapshotList.TotalItems).To(Equal(3))
			Expect(snapshotList.Items[0].(*v1beta1.VolumeSnapshot).Name).To(Equal(snapshot3.Name))
			Expect(snapshotList.Items[1].(*v1beta1.VolumeSnapshot).Name).To(Equal(snapshot2.Name))
			Expect(snapshotList.Items[2].(*v1beta1.VolumeSnapshot).Name).To(Equal(snapshot1.Name))
		})

		It("by name", func() {
			query1 := query.New()
			query1.SortBy = query.FieldName
			query1.Ascending = true
			snapshotList, err := getter.List(defaultNamespace, query1)
			Expect(err).To(BeNil())
			Expect(snapshotList.TotalItems).To(Equal(3))
			Expect(snapshotList.Items[0].(*v1beta1.VolumeSnapshot).Name).To(Equal(snapshot1.Name))
			Expect(snapshotList.Items[1].(*v1beta1.VolumeSnapshot).Name).To(Equal(snapshot2.Name))
			Expect(snapshotList.Items[2].(*v1beta1.VolumeSnapshot).Name).To(Equal(snapshot3.Name))
		})
		It("by name and reverse", func() {
			query1 := query.New()
			query1.SortBy = query.FieldName
			query1.Ascending = false
			snapshotList, err := getter.List(defaultNamespace, query1)
			Expect(err).To(BeNil())
			Expect(snapshotList.TotalItems).To(Equal(3))
			Expect(snapshotList.Items[0].(*v1beta1.VolumeSnapshot).Name).To(Equal(snapshot3.Name))
			Expect(snapshotList.Items[1].(*v1beta1.VolumeSnapshot).Name).To(Equal(snapshot2.Name))
			Expect(snapshotList.Items[2].(*v1beta1.VolumeSnapshot).Name).To(Equal(snapshot1.Name))
		})
	})

	Describe("snapshot status", func() {
		snapshot := newVolumeSnapshot("snap-0")
		It("snapshot == nil", func() {
			Expect(snapshotStatus(nil)).To(Equal(statusCreating))
		})
		It("snapshot.Status == nil", func() {
			snapshot.Status = nil
			Expect(snapshotStatus(snapshot)).To(Equal(statusCreating))
		})
		It("snapshot.Status.ReadyToUse == nil", func() {
			snapshot.Status = &v1beta1.VolumeSnapshotStatus{
				ReadyToUse: nil,
			}
			Expect(snapshotStatus(snapshot)).To(Equal(statusCreating))
		})
		It("snapshot.Status.ReadyToUse == false", func() {
			readyToUse := false
			snapshot.Status = &v1beta1.VolumeSnapshotStatus{
				ReadyToUse: &readyToUse,
			}
			Expect(snapshotStatus(snapshot)).To(Equal(statusCreating))
		})

		It("snapshot.Status.ReadyToUse == true", func() {
			readyToUse := true
			snapshot.Status = &v1beta1.VolumeSnapshotStatus{
				ReadyToUse: &readyToUse,
			}
			Expect(snapshotStatus(snapshot)).To(Equal(statusReady))
		})
	})

	RunSpecs(t, "volume snapshot getter list")
}

//func TestVolumeSnapshotStatus( t *testing.T)  {
//	RegisterFailHandler(Fail)
//
//	RunSpecs(t, "volume snapshot status")
//}
