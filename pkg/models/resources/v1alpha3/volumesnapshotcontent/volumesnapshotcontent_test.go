package volumesnapshotcontent

import (
	"encoding/json"
	"testing"
	"time"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	"github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/fake"
	"github.com/kubernetes-csi/external-snapshotter/client/v4/informers/externalversions"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/apiserver/query"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	baseVolumeSnapshotContent = `{
	"apiVersion": "snapshot.storage.k8s.io/v1",
	"kind": "VolumeSnapshotContent",
	"metadata": {
		"creationTimestamp": "2020-04-29T06:52:06Z",
  		"finalizers": [
			"snapshot.storage.kubernetes.io/volumesnapshotcontent-bound-protection"
		],
		"generation": 1,
		"name": "snapcontent-1",
		"resourceVersion": "827984",
		"uid": "80dce0bc-67dd-4d87-b91c-e1d4ad7350f3"
	},
	"spec": {
		"deletionPolicy": "Delete",
		"driver": "disk.csi.qingcloud.com",
		"source": {
			"volumeHandle": "vol-hrguk3bo"
		},
		"volumeSnapshotClassName": "csi-qingcloud",
		"volumeSnapshotRef": {
			"apiVersion": "snapshot.storage.k8s.io/v1",
			"kind": "VolumeSnapshot",
			"name": "tt",
			"namespace": "kubesphere-monitoring-system",
			"resourceVersion": "827830",
			"uid": "45534028-c659-4dfe-9498-ccc25e03afc2"
		}
	},
	"status": {
		"creationTime": 1638866716000000000,
		"readyToUse": true,
		"restoreSize": 21474836480,
		"snapshotHandle": "ss-83jv1p2a"
	}
}`
)

func newVolumeSnapshotContent(name string) *snapshotv1.VolumeSnapshotContent {
	volumeSnapshotContent := &snapshotv1.VolumeSnapshotContent{}
	err := json.Unmarshal([]byte(baseVolumeSnapshotContent), volumeSnapshotContent)
	if err != nil {
		return nil
	}
	volumeSnapshotContent.Name = name
	return volumeSnapshotContent
}

func TestListVolumeSnapshot(t *testing.T) {
	RegisterFailHandler(Fail)

	client := fake.NewSimpleClientset()
	informer := externalversions.NewSharedInformerFactory(client, 0)

	snapshotContent1 := newVolumeSnapshotContent("snapshotContent-1")
	snapshotContent1.CreationTimestamp = v1.NewTime(snapshotContent1.CreationTimestamp.Add(time.Hour * 3))

	snapshotContent2 := newVolumeSnapshotContent("snapshotContent-2")
	snapshotContent2.CreationTimestamp = v1.NewTime(snapshotContent2.CreationTimestamp.Add(time.Hour * 2))
	targetName := "targetVolumeSnapshotClassName"
	snapshotContent2.Spec.VolumeSnapshotClassName = &targetName

	snapshotContent3 := newVolumeSnapshotContent("snapshotContent-3")
	snapshotContent3.CreationTimestamp = v1.NewTime(snapshotContent3.CreationTimestamp.Add(time.Hour * 1))
	snapshotContent3.Spec.VolumeSnapshotRef.Name = "target-snapshot"

	snapshotContents := []interface{}{snapshotContent1, snapshotContent2, snapshotContent3}

	for _, s := range snapshotContents {
		_ = informer.Snapshot().V1().VolumeSnapshotContents().Informer().GetIndexer().Add(s)
	}
	getter := New(informer)

	Describe("condition", func() {
		It("match name", func() {
			query1 := query.New()
			query1.Filters[query.FieldName] = query.Value(snapshotContent1.Name)
			snapshotContentList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotContentList.TotalItems).To(Equal(1))
			Expect(snapshotContentList.Items[0]).To(Equal(snapshotContent1))
		})

		It("match volumeSnapshotClassName", func() {
			query1 := query.New()
			query1.Filters[volumeSnapshotClassName] = query.Value(*snapshotContent2.Spec.VolumeSnapshotClassName)
			snapshotContentList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotContentList.TotalItems).To(Equal(1))
			Expect(snapshotContentList.Items[0]).To(Equal(snapshotContent2))
		})

		It("match volumeSnapshotName", func() {
			query1 := query.New()
			query1.Filters[volumeSnapshotName] = query.Value(snapshotContent3.Spec.VolumeSnapshotRef.Name)
			snapshotContentList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotContentList.TotalItems).To(Equal(1))
			Expect(snapshotContentList.Items[0]).To(Equal(snapshotContent3))
		})
	})

	Describe("list", func() {
		It("by createTime", func() {
			query1 := query.New()
			query1.SortBy = query.FieldCreateTime
			query1.Ascending = true
			snapshotContentList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotContentList.Items[0].(*snapshotv1.VolumeSnapshotContent).Name).To(Equal(snapshotContent3.Name))
			Expect(snapshotContentList.Items[1].(*snapshotv1.VolumeSnapshotContent).Name).To(Equal(snapshotContent2.Name))
			Expect(snapshotContentList.Items[2].(*snapshotv1.VolumeSnapshotContent).Name).To(Equal(snapshotContent1.Name))
		})

		It("by name", func() {
			query1 := query.New()
			query1.SortBy = query.FieldName
			query1.Ascending = true
			snapshotContentList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotContentList.Items[0].(*snapshotv1.VolumeSnapshotContent).Name).To(Equal(snapshotContent1.Name))
			Expect(snapshotContentList.Items[1].(*snapshotv1.VolumeSnapshotContent).Name).To(Equal(snapshotContent2.Name))
			Expect(snapshotContentList.Items[2].(*snapshotv1.VolumeSnapshotContent).Name).To(Equal(snapshotContent3.Name))

		})

		It("by name and reverse", func() {
			query1 := query.New()
			query1.SortBy = query.FieldName
			query1.Ascending = false
			snapshotContentList, err := getter.List("", query1)
			Expect(err).To(BeNil())
			Expect(snapshotContentList.Items[0].(*snapshotv1.VolumeSnapshotContent).Name).To(Equal(snapshotContent3.Name))
			Expect(snapshotContentList.Items[1].(*snapshotv1.VolumeSnapshotContent).Name).To(Equal(snapshotContent2.Name))
			Expect(snapshotContentList.Items[2].(*snapshotv1.VolumeSnapshotContent).Name).To(Equal(snapshotContent1.Name))

		})
	})
	RunSpecs(t, "volume snapshot content list")
}
