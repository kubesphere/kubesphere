/*
Copyright 2020 KubeSphere Authors

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

package components

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/api/resource/v1alpha2"
	"testing"
	"time"
)

func service(name, namespace string, selector map[string]string) runtime.Object {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: selector,
		},
	}
}

func pods(name, namespace string, labels map[string]string, healthPods, totalPods int) []runtime.Object {
	var ps []runtime.Object

	for index := 0; index < totalPods; index++ {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", name, index),
				Namespace: namespace,
				Labels:    labels,
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:  fmt.Sprintf("%s-%d", name, index),
						Ready: true,
					},
				},
			},
		}

		if index >= healthPods {
			pod.Status.Phase = v1.PodPending
			pod.Status.ContainerStatuses[0].Ready = false
		}

		ps = append(ps, pod)
	}

	return ps
}

func nodes(name string, healthNodes, totalNodes int) []runtime.Object {
	var ns []runtime.Object

	for index := 0; index < totalNodes; index++ {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%d", name, index),
			},
			Status: v1.NodeStatus{
				Phase: v1.NodeRunning,
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		if index >= healthNodes {
			node.Status.Phase = v1.NodePending
			node.Status.Conditions[0].Status = v1.ConditionFalse
		}

		ns = append(ns, node)
	}

	return ns
}

func TestGetSystemHealthStatus(t *testing.T) {
	var tests = []struct {
		description string
		labels      map[string]string
		namespace   string
		name        string
		healthPods  int
		totalPods   int
		healthNodes int
		totalNodes  int
		expected    v1alpha2.HealthStatus
	}{
		{
			"no backends",
			map[string]string{"app": "foo"},
			"kube-system",
			"",
			0,
			0,
			0,
			0,
			v1alpha2.HealthStatus{
				KubeSphereComponents: []v1alpha2.ComponentStatus{
					{
						Namespace:       "kube-system",
						Label:           map[string]string{"app": "foo"},
						TotalBackends:   0,
						HealthyBackends: 0,
					},
				},
				NodeStatus: v1alpha2.NodeStatus{},
			},
		},
		{
			"all healthy",
			map[string]string{"app": "foo"},
			"kubesphere-system",
			"ks-apiserver",
			2,
			2,
			2,
			2,
			v1alpha2.HealthStatus{
				KubeSphereComponents: []v1alpha2.ComponentStatus{
					{
						Name:            "ks-apiserver",
						Namespace:       "kubesphere-system",
						Label:           map[string]string{"app": "foo"},
						TotalBackends:   2,
						HealthyBackends: 2,
					},
				},
				NodeStatus: v1alpha2.NodeStatus{
					TotalNodes:   2,
					HealthyNodes: 2,
				},
			},
		},
		{
			"all unhealthy",
			map[string]string{"app": "foo"},
			"kubesphere-system",
			"ks-apiserver",
			0,
			2,
			0,
			2,
			v1alpha2.HealthStatus{
				KubeSphereComponents: []v1alpha2.ComponentStatus{
					{
						Name:            "ks-apiserver",
						Namespace:       "kubesphere-system",
						Label:           map[string]string{"app": "foo"},
						TotalBackends:   2,
						HealthyBackends: 0,
					},
				},
				NodeStatus: v1alpha2.NodeStatus{
					TotalNodes:   2,
					HealthyNodes: 0,
				},
			},
		},
		{
			"half healthy",
			map[string]string{"app": "foo"},
			"kubesphere-system",
			"ks-apiserver",
			2,
			4,
			2,
			4,
			v1alpha2.HealthStatus{
				KubeSphereComponents: []v1alpha2.ComponentStatus{
					{
						Name:            "ks-apiserver",
						Namespace:       "kubesphere-system",
						Label:           map[string]string{"app": "foo"},
						TotalBackends:   4,
						HealthyBackends: 2,
					},
				},
				NodeStatus: v1alpha2.NodeStatus{
					TotalNodes:   4,
					HealthyNodes: 2,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ps := pods(test.name, test.namespace, test.labels, test.healthPods, test.totalPods)
			svc := service(test.name, test.namespace, test.labels)
			ns := nodes(test.name, test.healthNodes, test.totalNodes)

			var objs []runtime.Object
			objs = append(objs, ps...)
			objs = append(objs, svc)
			objs = append(objs, ns...)

			client := fake.NewSimpleClientset(objs...)

			informer := informers.NewSharedInformerFactory(client, time.Minute*10)

			informer.Core().V1().Services().Informer().GetIndexer().Add(svc)

			for _, obj := range ps {
				informer.Core().V1().Pods().Informer().GetIndexer().Add(obj)
			}

			for _, obj := range ns {
				informer.Core().V1().Nodes().Informer().GetIndexer().Add(obj)
			}

			c := NewComponentsGetter(informer)
			healthStatus, err := c.GetSystemHealthStatus()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(healthStatus, test.expected); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
				return
			}
		})
	}

}

func TestGetComponentStatus(t *testing.T) {
	var tests = []struct {
		description   string
		name          string
		namespace     string
		labels        map[string]string
		healthPods    int
		totalPods     int
		expected      v1alpha2.ComponentStatus
		expectedError bool
	}{
		{
			"no component",
			"random",
			"foo",
			map[string]string{"app": "foo"},
			2,
			4,
			v1alpha2.ComponentStatus{
				Name:            "",
				Namespace:       "",
				SelfLink:        "",
				Label:           nil,
				StartedAt:       time.Time{},
				TotalBackends:   0,
				HealthyBackends: 0,
			},
			true,
		},
		{
			"all healthy",
			"ks-apiserver",
			"kubesphere-system",
			map[string]string{"app": "foo"},
			2,
			4,
			v1alpha2.ComponentStatus{
				Name:            "ks-apiserver",
				Namespace:       "kubesphere-system",
				Label:           map[string]string{"app": "foo"},
				TotalBackends:   4,
				HealthyBackends: 2,
			},
			false,
		},
		{
			"all unhealthy",
			"ks-apiserver",
			"kubesphere-system",
			map[string]string{"app": "foo"},
			0,
			4,
			v1alpha2.ComponentStatus{
				Name:            "ks-apiserver",
				Namespace:       "kubesphere-system",
				Label:           map[string]string{"app": "foo"},
				TotalBackends:   4,
				HealthyBackends: 0,
			},
			false,
		},
		{
			"half healthy",
			"ks-apiserver",
			"kubesphere-system",
			map[string]string{"app": "foo"},
			2,
			4,
			v1alpha2.ComponentStatus{
				Name:            "ks-apiserver",
				Namespace:       "kubesphere-system",
				Label:           map[string]string{"app": "foo"},
				TotalBackends:   4,
				HealthyBackends: 2,
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ps := pods(test.name, test.namespace, test.labels, test.healthPods, test.totalPods)
			svc := service(test.name, test.namespace, test.labels)

			var objs []runtime.Object
			objs = append(objs, ps...)
			objs = append(objs, svc)

			client := fake.NewSimpleClientset(objs...)

			informer := informers.NewSharedInformerFactory(client, time.Minute*10)

			informer.Core().V1().Services().Informer().GetIndexer().Add(svc)

			for _, obj := range ps {
				informer.Core().V1().Pods().Informer().GetIndexer().Add(obj)
			}

			c := NewComponentsGetter(informer)
			healthStatus, err := c.GetComponentStatus(test.name)
			if err == nil && test.expectedError {
				t.Fatalf("expected error while got nothing")
			} else if err != nil && !test.expectedError {
				t.Fatal(err)
			}

			if diff := cmp.Diff(healthStatus, test.expected); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
				return
			}
		})
	}
}
