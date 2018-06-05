/*
Copyright 2018 The Kubernetes Authors.

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

package statefulset

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	typedv1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	//svc "k8s.io/kubernetes/pkg/api/v1/service"
	"k8s.io/kubernetes/pkg/controller/statefulset"
	"k8s.io/kubernetes/test/integration/framework"
)

const (
	pollInterval = 100 * time.Millisecond
	pollTimeout  = 60 * time.Second

	fakeImageName = "fake-name"
	fakeImage     = "fakeimage"
)

type statefulsetTester struct {
	t           *testing.T
	c           clientset.Interface
	service     *v1.Service
	statefulset *v1beta1.StatefulSet
}

func labelMap() map[string]string {
	return map[string]string{"foo": "bar"}
}

// newService returns a service with a fake name for StatefulSet to be created soon
func newHeadlessService(namespace string) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "fake-service-name",
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{Port: 80, Name: "http", Protocol: "TCP"},
			},
			Selector: labelMap(),
		},
	}
}

// newSTS returns a StatefulSet with with a fake container image
func newSTS(name, namespace string, replicas int) *v1beta1.StatefulSet {
	replicasCopy := int32(replicas)
	return &v1beta1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1beta1.StatefulSetSpec{
			PodManagementPolicy: v1beta1.ParallelPodManagement,
			Replicas:            &replicasCopy,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelMap(),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelMap(),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "fake-name",
							Image: "fakeimage",
							VolumeMounts: []v1.VolumeMount{
								{Name: "datadir", MountPath: "/data/"},
								{Name: "home", MountPath: "/home"},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "datadir",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: fmt.Sprintf("/tmp/%v", "datadir"),
								},
							},
						},
						{
							Name: "home",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: fmt.Sprintf("/tmp/%v", "home"),
								},
							},
						},
					},
				},
			},
			ServiceName: "fake-service-name",
			UpdateStrategy: v1beta1.StatefulSetUpdateStrategy{
				Type: v1beta1.RollingUpdateStatefulSetStrategyType,
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				// for volume mount "datadir"
				newStatefulSetPVC("fake-pvc-name"),
			},
		},
	}
}

func newStatefulSetPVC(name string) v1.PersistentVolumeClaim {
	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"volume.alpha.kubernetes.io/storage-class": "anything",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: *resource.NewQuantity(1, resource.BinarySI),
				},
			},
		},
	}
}

// scSetup sets up necessities for Statefulset integration test, including master, apiserver, informers, and clientset
func scSetup(t *testing.T) (*httptest.Server, framework.CloseFunc, *statefulset.StatefulSetController, informers.SharedInformerFactory, clientset.Interface) {
	masterConfig := framework.NewIntegrationTestMasterConfig()
	_, s, closeFn := framework.RunAMaster(masterConfig)

	config := restclient.Config{Host: s.URL}
	clientSet, err := clientset.NewForConfig(&config)
	if err != nil {
		t.Fatalf("error in create clientset: %v", err)
	}
	resyncPeriod := 12 * time.Hour
	informers := informers.NewSharedInformerFactory(clientset.NewForConfigOrDie(restclient.AddUserAgent(&config, "statefulset-informers")), resyncPeriod)

	sc := statefulset.NewStatefulSetController(
		informers.Core().V1().Pods(),
		informers.Apps().V1().StatefulSets(),
		informers.Core().V1().PersistentVolumeClaims(),
		informers.Apps().V1().ControllerRevisions(),
		clientset.NewForConfigOrDie(restclient.AddUserAgent(&config, "statefulset-controller")),
	)

	return s, closeFn, sc, informers, clientSet
}

// Run STS controller and informers
func runControllerAndInformers(sc *statefulset.StatefulSetController, informers informers.SharedInformerFactory) chan struct{} {
	stopCh := make(chan struct{})
	informers.Start(stopCh)
	go sc.Run(5, stopCh)
	return stopCh
}

func createHeadlessService(t *testing.T, clientSet clientset.Interface, headlessService *v1.Service) {
	_, err := clientSet.Core().Services(headlessService.Namespace).Create(headlessService)
	if err != nil {
		t.Fatalf("failed creating headless service: %v", err)
	}
}

func createSTSsPods(t *testing.T, clientSet clientset.Interface, stss []*v1beta1.StatefulSet, pods []*v1.Pod) ([]*v1beta1.StatefulSet, []*v1.Pod) {
	var createdSTSs []*v1beta1.StatefulSet
	var createdPods []*v1.Pod
	for _, sts := range stss {
		createdSTS, err := clientSet.AppsV1beta1().StatefulSets(sts.Namespace).Create(sts)
		if err != nil {
			t.Fatalf("failed to create sts %s: %v", sts.Name, err)
		}
		createdSTSs = append(createdSTSs, createdSTS)
	}
	for _, pod := range pods {
		createdPod, err := clientSet.Core().Pods(pod.Namespace).Create(pod)
		if err != nil {
			t.Fatalf("failed to create pod %s: %v", pod.Name, err)
		}
		createdPods = append(createdPods, createdPod)
	}

	return createdSTSs, createdPods
}

// Verify .Status.Replicas is equal to .Spec.Replicas
func waitSTSStable(t *testing.T, clientSet clientset.Interface, sts *v1beta1.StatefulSet) {
	stsClient := clientSet.AppsV1beta1().StatefulSets(sts.Namespace)
	desiredGeneration := sts.Generation
	if err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		newSTS, err := stsClient.Get(sts.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return newSTS.Status.Replicas == *newSTS.Spec.Replicas && *newSTS.Status.ObservedGeneration >= desiredGeneration, nil
	}); err != nil {
		t.Fatalf("failed to verify .Status.Replicas is equal to .Spec.Replicas for sts %s: %v", sts.Name, err)
	}
}

func updatePod(t *testing.T, podClient typedv1.PodInterface, podName string, updateFunc func(*v1.Pod)) *v1.Pod {
	var pod *v1.Pod
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newPod, err := podClient.Get(podName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		updateFunc(newPod)
		pod, err = podClient.Update(newPod)
		return err
	}); err != nil {
		t.Fatalf("failed to update pod %s: %v", podName, err)
	}
	return pod
}

func updatePodStatus(t *testing.T, podClient typedv1.PodInterface, podName string, updateStatusFunc func(*v1.Pod)) *v1.Pod {
	var pod *v1.Pod
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newPod, err := podClient.Get(podName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		updateStatusFunc(newPod)
		pod, err = podClient.UpdateStatus(newPod)
		return err
	}); err != nil {
		t.Fatalf("failed to update status of pod %s: %v", podName, err)
	}
	return pod
}

func getPods(t *testing.T, podClient typedv1.PodInterface, labelMap map[string]string) *v1.PodList {
	podSelector := labels.Set(labelMap).AsSelector()
	options := metav1.ListOptions{LabelSelector: podSelector.String()}
	pods, err := podClient.List(options)
	if err != nil {
		t.Fatalf("failed obtaining a list of pods that match the pod labels %v: %v", labelMap, err)
	}
	if pods == nil {
		t.Fatalf("obtained a nil list of pods")
	}
	return pods
}

func updateSTS(t *testing.T, stsClient typedv1beta1.StatefulSetInterface, stsName string, updateFunc func(*v1beta1.StatefulSet)) *v1beta1.StatefulSet {
	var sts *v1beta1.StatefulSet
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newSTS, err := stsClient.Get(stsName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		updateFunc(newSTS)
		sts, err = stsClient.Update(newSTS)
		return err
	}); err != nil {
		t.Fatalf("failed to update sts %s: %v", stsName, err)
	}
	return sts
}

// Update .Spec.Replicas to replicas and verify .Status.Replicas is changed accordingly
func scaleSTS(t *testing.T, c clientset.Interface, sts *v1beta1.StatefulSet, replicas int32) {
	stsClient := c.AppsV1beta1().StatefulSets(sts.Namespace)
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newSTS, err := stsClient.Get(sts.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		*newSTS.Spec.Replicas = replicas
		sts, err = stsClient.Update(newSTS)
		return err
	}); err != nil {
		t.Fatalf("failed to update .Spec.Replicas to %d for sts %s: %v", replicas, sts.Name, err)
	}
	waitSTSStable(t, c, sts)
}
