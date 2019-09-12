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

package expansion

import (
	"encoding/json"
	"flag"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "/tmp")
	flag.Set("v", "3")
	flag.Parse()
	ret := m.Run()
	os.Exit(ret)
}

func TestSyncHandler(t *testing.T) {
	retryTime = wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2,
		Steps:    2,
	}
	tests := []struct {
		name     string
		pvc      *v1.PersistentVolumeClaim
		pod      *v1.Pod
		deploy   *appsv1.Deployment
		rs       *appsv1.ReplicaSet
		sts      *appsv1.StatefulSet
		sc       *storagev1.StorageClass
		pvcKey   string
		hasError bool
	}{
		{
			name: "mount pvc on deploy",
			pvc:  getFakePersistentVolumeClaim("fake-pvc", "vol-12345", "fake-sc", types.UID(123)),
			sc:   getFakeStorageClass("fake-sc", "fake.sc.com"),
			deploy: getFakeDeployment("fake-deploy", "234", 1,
				getFakePersistentVolumeClaim("fake-pvc", "vol-12345", "fake-sc", types.UID(123))),
			pvcKey:   "default/fake-pvc",
			hasError: false,
		},
		{
			name:     "unmounted pvc",
			pvc:      getFakePersistentVolumeClaim("fake-pvc", "vol-12345", "fake-sc", types.UID(123)),
			sc:       getFakeStorageClass("fake-sc", "fake.sc.com"),
			pvcKey:   "default/fake-pvc",
			hasError: true,
		},
	}
	for _, tc := range tests {
		test := tc
		fakeKubeClient := &fake.Clientset{}
		fakeWatch := watch.NewFake()
		fakeKubeClient.AddWatchReactor("*", clientgotesting.DefaultWatchReactor(fakeWatch, nil))
		informerFactory := informers.NewSharedInformerFactory(fakeKubeClient, 0)
		pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
		storageClassInformer := informerFactory.Storage().V1().StorageClasses()
		podInformer := informerFactory.Core().V1().Pods()
		deployInformer := informerFactory.Apps().V1().Deployments()
		rsInformer := informerFactory.Apps().V1().ReplicaSets()
		stsInformer := informerFactory.Apps().V1().StatefulSets()
		pvc := tc.pvc

		if tc.pvc != nil {
			informerFactory.Core().V1().PersistentVolumeClaims().Informer().GetIndexer().Add(pvc)
		}
		if tc.sc != nil {
			informerFactory.Storage().V1().StorageClasses().Informer().GetIndexer().Add(tc.sc)
		}
		if tc.deploy != nil {
			informerFactory.Apps().V1().Deployments().Informer().GetIndexer().Add(tc.deploy)
			tc.rs = generateReplicaSetFromDeployment(tc.deploy)
		}
		if tc.rs != nil {
			informerFactory.Apps().V1().ReplicaSets().Informer().GetIndexer().Add(tc.rs)
			tc.pod = generatePodFromReplicaSet(tc.rs)
		}
		if tc.sts != nil {
			informerFactory.Apps().V1().StatefulSets().Informer().GetIndexer().Add(tc.sts)
		}
		if tc.pod != nil {
			informerFactory.Core().V1().Pods().Informer().GetIndexer().Add(tc.pod)
		}
		expc := NewVolumeExpansionController(fakeKubeClient, pvcInformer, storageClassInformer, podInformer, deployInformer,
			rsInformer, stsInformer)
		fakeKubeClient.AddReactor("patch", "persistentvolumeclaims", func(action clientgotesting.Action) (bool, runtime.Object,
			error) {
			if action.GetSubresource() == "status" {
				patchActionaction, _ := action.(clientgotesting.PatchAction)
				pvc, err := applyPVCPatch(pvc, patchActionaction.GetPatch())
				if err != nil {
					return false, nil, err
				}
				return true, pvc, nil
			}
			return true, pvc, nil
		})
		err := expc.syncHandler(test.pvcKey)
		if err != nil && !test.hasError {
			t.Fatalf("for: %s; unexpected error while running handler : %v", test.name, err)
		}
	}
}

func applyPVCPatch(originalPVC *v1.PersistentVolumeClaim, patch []byte) (*v1.PersistentVolumeClaim, error) {
	pvcData, err := json.Marshal(originalPVC)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pvc with %v", err)
	}
	updated, err := strategicpatch.StrategicMergePatch(pvcData, patch, v1.PersistentVolumeClaim{})
	if err != nil {
		return nil, fmt.Errorf("failed to apply patch on pvc %v", err)
	}
	updatedPVC := &v1.PersistentVolumeClaim{}
	if err := json.Unmarshal(updated, updatedPVC); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated pvc : %v", err)
	}
	return updatedPVC, nil
}

func getFakePersistentVolumeClaim(pvcName, volumeName, scName string, uid types.UID) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: pvcName, Namespace: "default", UID: uid},
		Spec:       v1.PersistentVolumeClaimSpec{},
	}
	if volumeName != "" {
		pvc.Spec.VolumeName = volumeName
	}

	if scName != "" {
		pvc.Spec.StorageClassName = &scName
	}
	return pvc
}

func getFakeStorageClass(scName, pluginName string) *storagev1.StorageClass {
	return &storagev1.StorageClass{
		ObjectMeta:  metav1.ObjectMeta{Name: scName},
		Provisioner: pluginName,
	}
}

func getFakePod(podName string, ownerRef *metav1.OwnerReference, uid types.UID) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            podName,
			Namespace:       "default",
			UID:             uid,
			OwnerReferences: []metav1.OwnerReference{*ownerRef},
		},
		Spec: v1.PodSpec{},
	}
	return pod
}

func getFakeReplicaSet(rsName string, ownerRef *metav1.OwnerReference, uid types.UID, replicas int) *appsv1.ReplicaSet {
	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            rsName,
			Namespace:       "default",
			UID:             uid,
			OwnerReferences: []metav1.OwnerReference{*ownerRef},
		},
		Spec: appsv1.ReplicaSetSpec{},
	}
	return rs
}

func getFakeDeployment(deployName string, uid types.UID, replicas int32, mountPVC *v1.PersistentVolumeClaim) *appsv1.
	Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployName,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "test",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: mountPVC.GetName(),
								},
							},
						},
					},
				},
			},
		},
	}
}

func getFakeStatefulSet(stsName string, uid types.UID, replicas int32, mountPVC *v1.PersistentVolumeClaim) *appsv1.
	StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stsName,
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "test",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: mountPVC.GetName(),
								},
							},
						},
					},
				},
			},
		},
	}
}

// generatePodFromRS creates a pod, with the input ReplicaSet's selector and its template
func generatePodFromReplicaSet(rs *appsv1.ReplicaSet) *v1.Pod {
	trueVar := true
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rs.Name + "-pod",
			Namespace: rs.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{UID: rs.UID, APIVersion: "v1beta1", Kind: "ReplicaSet", Name: rs.Name, Controller: &trueVar},
			},
		},
		Spec: rs.Spec.Template.Spec,
	}
}

func generateReplicaSetFromDeployment(deploy *appsv1.Deployment) *appsv1.ReplicaSet {
	trueVar := true
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploy.Name + "-rs",
			Namespace: deploy.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{UID: deploy.UID, APIVersion: "v1beta1", Kind: "Deployment", Name: deploy.Name, Controller: &trueVar},
			},
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: deploy.Spec.Replicas,
			Template: deploy.Spec.Template,
		},
	}
}
