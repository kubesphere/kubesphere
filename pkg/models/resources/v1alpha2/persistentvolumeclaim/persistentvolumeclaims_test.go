package persistentvolumeclaim

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/google/go-cmp/cmp"
	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	snapshotefakeclient "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/fake"
	snapshotinformers "github.com/kubernetes-csi/external-snapshotter/client/v4/informers/externalversions"

	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
)

var (
	testStorageClassName = "sc1"
)

var (
	pvc1 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-1",
			Namespace: "default",
			Annotations: map[string]string{
				"kubesphere.io/in-use":         "false",
				"kubesphere.io/allow-snapshot": "false",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &testStorageClassName,
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimPending,
		},
	}

	pvc2 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-2",
			Namespace: "default",
			Annotations: map[string]string{
				"kubesphere.io/in-use":         "false",
				"kubesphere.io/allow-snapshot": "false",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &testStorageClassName,
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimLost,
		},
	}

	pvc3 = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-3",
			Namespace: "default",
			Annotations: map[string]string{
				"kubesphere.io/in-use":         "true",
				"kubesphere.io/allow-snapshot": "false",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &testStorageClassName,
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimBound,
		},
	}

	pod1 = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc3.Name,
						},
					},
				},
			},
		},
	}

	vsc1 = &snapshot.VolumeSnapshotClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "VolumeSnapshotClass-1",
			Namespace: "default",
		},
		Driver: testStorageClassName,
	}

	persistentVolumeClaims = []interface{}{pvc1, pvc2, pvc3}
	pods                   = []interface{}{pod1}
	volumeSnapshotClasses  = []interface{}{vsc1}
)

func prepare() (v1alpha2.Interface, error) {
	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)
	snapshotClient := snapshotefakeclient.NewSimpleClientset()
	snapshotInformers := snapshotinformers.NewSharedInformerFactory(snapshotClient, 0)

	for _, persistentVolumeClaim := range persistentVolumeClaims {
		err := informer.Core().V1().PersistentVolumeClaims().Informer().GetIndexer().Add(persistentVolumeClaim)
		if err != nil {
			return nil, err
		}
	}

	for _, pod := range pods {
		err := informer.Core().V1().Pods().Informer().GetIndexer().Add(pod)
		if err != nil {
			return nil, err
		}
	}

	for _, volumeSnapshotClass := range volumeSnapshotClasses {
		err := snapshotInformers.Snapshot().V1().VolumeSnapshotClasses().Informer().GetIndexer().Add(volumeSnapshotClass)
		if err != nil {
			return nil, err
		}
	}

	return NewPersistentVolumeClaimSearcher(informer, snapshotInformers), nil
}

func TestGet(t *testing.T) {
	tests := []struct {
		Namespace   string
		Name        string
		Expected    interface{}
		ExpectedErr error
	}{
		{
			"default",
			"pvc-1",
			pvc1,
			nil,
		},
	}

	getter, err := prepare()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		got, err := getter.Get(test.Namespace, test.Name)
		if test.ExpectedErr != nil && err != test.ExpectedErr {
			t.Errorf("expected error, got nothing")
		} else if err != nil {
			t.Fatal(err)
		}

		diff := cmp.Diff(got, test.Expected)
		if diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.Expected, diff)
		}
	}
}

func TestSearch(t *testing.T) {
	tests := []struct {
		Namespace   string
		Conditions  *params.Conditions
		OrderBy     string
		Reverse     bool
		Expected    []interface{}
		ExpectedErr error
	}{
		{
			Namespace: "default",
			Conditions: &params.Conditions{
				Match: map[string]string{
					v1alpha2.Status: v1alpha2.StatusPending,
				},
				Fuzzy: nil,
			},
			OrderBy:     "name",
			Reverse:     false,
			Expected:    []interface{}{pvc1},
			ExpectedErr: nil,
		},
		{
			Namespace: "default",
			Conditions: &params.Conditions{
				Match: map[string]string{
					v1alpha2.Status: v1alpha2.StatusLost,
				},
				Fuzzy: nil,
			},
			OrderBy:     "name",
			Reverse:     false,
			Expected:    []interface{}{pvc2},
			ExpectedErr: nil,
		},
		{
			Namespace: "default",
			Conditions: &params.Conditions{
				Match: map[string]string{
					v1alpha2.Status: v1alpha2.StatusBound,
				},
				Fuzzy: nil,
			},
			OrderBy:     "name",
			Reverse:     false,
			Expected:    []interface{}{pvc3},
			ExpectedErr: nil,
		},
	}

	searcher, err := prepare()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		got, err := searcher.Search(test.Namespace, test.Conditions, test.OrderBy, test.Reverse)
		if test.ExpectedErr != nil && err != test.ExpectedErr {
			t.Errorf("expected error, got nothing")
		} else if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(got, test.Expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.Expected, diff)
		}
	}
}
