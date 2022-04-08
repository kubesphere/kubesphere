/*
Copyright 2021 The KubeSphere Authors.

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

package volumesnapshotclass

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	"github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/fake"
	"github.com/kubernetes-csi/external-snapshotter/client/v4/informers/externalversions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

const (
	baseVolumeSnapshotClass = `{
    "apiVersion": "snapshot.storage.k8s.io/v1",
	"deletionPolicy": "Delete",
	"driver": "disk.csi.qingcloud.com",
	"kind": "VolumeSnapshotClass",
	"metadata": {
	  	"creationTimestamp": "2021-12-23T09:53:19Z"
	}
}`
)

func newVolumeSnapshotClass(name string) *snapshotv1.VolumeSnapshotClass {
	volumeSnapshotclass := &snapshotv1.VolumeSnapshotClass{}
	err := json.Unmarshal([]byte(baseVolumeSnapshotClass), volumeSnapshotclass)
	if err != nil {
		return nil
	}
	volumeSnapshotclass.Name = name
	return volumeSnapshotclass
}

func TestListVolumeSnapshotClass(t *testing.T) {
	RegisterFailHandler(Fail)

	client := fake.NewSimpleClientset()
	informer := externalversions.NewSharedInformerFactory(client, 0)

	snapshotClass1 := newVolumeSnapshotClass("snapshotClass1")
	fmt.Println(snapshotClass1.CreationTimestamp)
	snapshotClass1.CreationTimestamp = v1.NewTime(snapshotClass1.CreationTimestamp.Add(time.Hour * 1))

	snapshotClass2 := newVolumeSnapshotClass("snapshotClass2")
	snapshotClass2.CreationTimestamp = v1.NewTime(snapshotClass1.CreationTimestamp.Add(time.Hour * 2))
	snapshotClass2.Driver = "target.csi.qingcloud.com"

	snapshotClass3 := newVolumeSnapshotClass("snapshotClass3")
	snapshotClass3.CreationTimestamp = v1.NewTime(snapshotClass1.CreationTimestamp.Add(time.Hour * 3))
	snapshotClass3.DeletionPolicy = snapshotv1.VolumeSnapshotContentRetain

	sc1Expected := snapshotClass1.DeepCopy()
	sc1Expected.Annotations = make(map[string]string)
	sc1Expected.Annotations["kubesphere.io/snapshot-count"] = "0"

	sc2Expected := snapshotClass2.DeepCopy()
	sc2Expected.Annotations = make(map[string]string)
	sc2Expected.Annotations["kubesphere.io/snapshot-count"] = "0"

	sc3Expected := snapshotClass3.DeepCopy()
	sc3Expected.Annotations = make(map[string]string)
	sc3Expected.Annotations["kubesphere.io/snapshot-count"] = "0"

	volumeSnapshotClasses := []interface{}{snapshotClass1, snapshotClass2, snapshotClass3}

	for _, s := range volumeSnapshotClasses {
		_ = informer.Snapshot().V1().VolumeSnapshotClasses().Informer().GetIndexer().Add(s)
	}
	getter := New(informer)

	Describe("condition", func() {
		It("match name", func() {
			query1 := query.New()
			query1.Filters[query.FieldName] = query.Value(snapshotClass1.Name)
			snapshotClassList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotClassList.TotalItems).To(Equal(1))
			Expect(snapshotClassList.Items[0]).To(Equal(sc1Expected))
		})

		It("match driver", func() {
			query1 := query.New()
			query1.Filters[driver] = query.Value(snapshotClass2.Driver)
			snapshotClassList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotClassList.TotalItems).To(Equal(1))
			Expect(snapshotClassList.Items[0]).To(Equal(sc2Expected))
		})

		It("match deletionPolicy", func() {
			query1 := query.New()
			query1.Filters[deletionPolicy] = query.Value("Retain")
			snapshotClassList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotClassList.TotalItems).To(Equal(1))
			Expect(snapshotClassList.Items[0]).To(Equal(sc3Expected))
		})

	})

	Describe("order", func() {
		It("by createTime ", func() {
			query1 := query.New()
			query1.SortBy = query.FieldCreateTime
			query1.Ascending = true
			snapshotClassList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotClassList.TotalItems).To(Equal(3))
			Expect(snapshotClassList.Items[0].(*snapshotv1.VolumeSnapshotClass).Name).To(Equal(snapshotClass1.Name))
			Expect(snapshotClassList.Items[1].(*snapshotv1.VolumeSnapshotClass).Name).To(Equal(snapshotClass2.Name))
			Expect(snapshotClassList.Items[2].(*snapshotv1.VolumeSnapshotClass).Name).To(Equal(snapshotClass3.Name))
		})

		It("by name", func() {
			query1 := query.New()
			query1.SortBy = query.FieldName
			query1.Ascending = true
			snapshotClassList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotClassList.TotalItems).To(Equal(3))
			Expect(snapshotClassList.Items[0].(*snapshotv1.VolumeSnapshotClass).Name).To(Equal(snapshotClass1.Name))
			Expect(snapshotClassList.Items[1].(*snapshotv1.VolumeSnapshotClass).Name).To(Equal(snapshotClass2.Name))
			Expect(snapshotClassList.Items[2].(*snapshotv1.VolumeSnapshotClass).Name).To(Equal(snapshotClass3.Name))
		})
		It("by name and reverse", func() {
			query1 := query.New()
			query1.SortBy = query.FieldName
			query1.Ascending = false
			snapshotList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotList.TotalItems).To(Equal(3))
			Expect(snapshotList.Items[0].(*snapshotv1.VolumeSnapshotClass).Name).To(Equal(snapshotClass3.Name))
			Expect(snapshotList.Items[1].(*snapshotv1.VolumeSnapshotClass).Name).To(Equal(snapshotClass2.Name))
			Expect(snapshotList.Items[2].(*snapshotv1.VolumeSnapshotClass).Name).To(Equal(snapshotClass1.Name))
		})
	})

	RunSpecs(t, "volume snapshotclass getter list")
}
