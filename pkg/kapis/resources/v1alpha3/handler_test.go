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

package v1alpha3

import (
	"github.com/google/go-cmp/cmp"
	fakesnapshot "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/clientset/versioned/fake"
	fakeistio "istio.io/client-go/pkg/clientset/versioned/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	fakeapiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	fakeapp "sigs.k8s.io/application/pkg/client/clientset/versioned/fake"
	"testing"
)

func TestResourceV1alpha2Fallback(t *testing.T) {
	tests := []struct {
		description   string
		namespace     string
		resource      string
		query         *query.Query
		expectedError error
		expected      *api.ListResult
	}{
		{
			description: "list namespaces",
			namespace:   "default",
			resource:    "namespaces",
			query: &query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    "name",
				Ascending: false,
				Filters:   nil,
			},
			expectedError: nil,
			expected: &api.ListResult{
				Items:      []interface{}{kubesphereNamespace, defaultNamespace},
				TotalItems: 2,
			},
		},
		{
			description: "list secrets fallback",
			namespace:   "default",
			resource:    "secrets",
			query: &query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    "name",
				Ascending: false,
				Filters:   nil,
			},
			expectedError: nil,
			expected: &api.ListResult{
				Items:      []interface{}{secretFoo2, secretFoo1},
				TotalItems: 2,
			},
		},
	}

	factory, err := prepare()
	if err != nil {
		t.Fatal(err)
	}

	handler := New(factory)

	for _, test := range tests {
		got, err := listResources(test.namespace, test.resource, test.query, handler)

		if err != test.expectedError {
			t.Fatalf("expected error: %s, got: %s", test.expectedError, err)
		}
		if diff := cmp.Diff(got, test.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
		}
	}
}

func listResources(namespace, resourceType string, query *query.Query, h *Handler) (*api.ListResult, error) {

	result, err := h.resourceGetterV1alpha3.List(resourceType, namespace, query)

	if err == nil {
		return result, nil
	}

	if err != resource.ErrResourceNotSupported {
		return nil, err
	}

	// fallback to v1alpha2
	return h.fallback(resourceType, namespace, query)
}

var (
	defaultNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "default",
			Labels: map[string]string{"kubesphere.io/workspace": "system-workspace"},
		},
	}
	kubesphereNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kubesphere-system",
			Labels: map[string]string{"kubesphere.io/workspace": "system-workspace"},
		},
	}
	secretFoo1 = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo1",
			Namespace: "default",
		},
	}
	secretFoo2 = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo2",
			Namespace: "default",
		},
	}

	replicas = int32(1)

	nginxDeployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 1,
		},
	}
	redisDeployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis",
			Namespace: "default",
			Labels:    map[string]string{"kubesphere.io/creator": "admin"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 0,
		},
	}
	deployments = []interface{}{redisDeployment, nginxDeployment}
	namespaces  = []interface{}{defaultNamespace, kubesphereNamespace}
	secrets     = []interface{}{secretFoo1, secretFoo2}
)

func prepare() (informers.InformerFactory, error) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	istioClient := fakeistio.NewSimpleClientset()
	appClient := fakeapp.NewSimpleClientset()
	snapshotClient := fakesnapshot.NewSimpleClientset()
	apiextensionsClient := fakeapiextensions.NewSimpleClientset()

	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, istioClient, appClient, snapshotClient, apiextensionsClient)

	k8sInformerFactory := fakeInformerFactory.KubernetesSharedInformerFactory()

	for _, namespace := range namespaces {
		err := k8sInformerFactory.Core().V1().Namespaces().Informer().GetIndexer().Add(namespace)
		if err != nil {
			return nil, err
		}
	}
	for _, deployment := range deployments {
		err := k8sInformerFactory.Apps().V1().Deployments().Informer().GetIndexer().Add(deployment)
		if err != nil {
			return nil, err
		}
	}
	for _, secret := range secrets {
		err := k8sInformerFactory.Core().V1().Secrets().Informer().GetIndexer().Add(secret)
		if err != nil {
			return nil, err
		}
	}

	return fakeInformerFactory, nil
}
