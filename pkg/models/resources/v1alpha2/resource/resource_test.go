/*
Copyright 2020 The KubeSphere Authors.

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

package resource

import (
	"github.com/google/go-cmp/cmp"
	fakesnapshot "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/clientset/versioned/fake"
	fakeistio "istio.io/client-go/pkg/clientset/versioned/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	fakeapp "sigs.k8s.io/application/pkg/client/clientset/versioned/fake"
	"testing"
)

func TestConditions(t *testing.T) {
	factory, err := prepare()
	if err != nil {
		t.Fatal(err)
	}
	resource := NewResourceGetter(factory)

	tests := []struct {
		Name           string
		Namespace      string
		Resource       string
		Conditions     *params.Conditions
		OrderBy        string
		Reverse        bool
		Limit          int
		Offset         int
		ExpectResponse *models.PageableResponse
		ExpectError    error
	}{{
		Name:       "list namespace order by name asc",
		Namespace:  "",
		Resource:   "namespaces",
		Conditions: &params.Conditions{},
		OrderBy:    "name",
		Reverse:    false,
		Limit:      10,
		Offset:     0,
		ExpectResponse: &models.PageableResponse{
			Items:      []interface{}{defaultNamespace, kubesphereNamespace},
			TotalCount: 2,
		},
		ExpectError: nil,
	}, {
		Name:       "list namespace order by name desc",
		Namespace:  "",
		Resource:   "namespaces",
		Conditions: &params.Conditions{},
		OrderBy:    "name",
		Reverse:    true,
		Limit:      10,
		Offset:     0,
		ExpectResponse: &models.PageableResponse{
			Items:      []interface{}{kubesphereNamespace, defaultNamespace},
			TotalCount: 2,
		},
		ExpectError: nil,
	},
		{
			Name:       "list deployment",
			Namespace:  "default",
			Resource:   "deployments",
			Conditions: &params.Conditions{},
			OrderBy:    "name",
			Reverse:    false,
			Limit:      10,
			Offset:     0,
			ExpectResponse: &models.PageableResponse{
				Items:      []interface{}{nginxDeployment, redisDeployment},
				TotalCount: 2,
			},
			ExpectError: nil,
		},
		{
			Name:      "filter deployment by keyword",
			Namespace: "default",
			Resource:  "deployments",
			Conditions: &params.Conditions{
				Match: map[string]string{v1alpha2.Keyword: "ngin"},
				Fuzzy: nil,
			},
			OrderBy: "name",
			Reverse: true,
			Limit:   10,
			Offset:  0,
			ExpectResponse: &models.PageableResponse{
				Items:      []interface{}{nginxDeployment},
				TotalCount: 1,
			},
			ExpectError: nil,
		},
		{
			Name:      "filter deployment by label",
			Namespace: "default",
			Resource:  "deployments",
			Conditions: &params.Conditions{
				Match: map[string]string{"kubesphere.io/creator": "admin"},
				Fuzzy: nil,
			},
			OrderBy: "",
			Reverse: true,
			Limit:   10,
			Offset:  0,
			ExpectResponse: &models.PageableResponse{
				Items:      []interface{}{redisDeployment},
				TotalCount: 1,
			},
			ExpectError: nil,
		}, {
			Name:      "filter deployment by status",
			Namespace: "default",
			Resource:  "deployments",
			Conditions: &params.Conditions{
				Match: map[string]string{v1alpha2.Status: v1alpha2.StatusRunning},
				Fuzzy: nil,
			},
			OrderBy: "",
			Reverse: true,
			Limit:   10,
			Offset:  0,
			ExpectResponse: &models.PageableResponse{
				Items:      []interface{}{nginxDeployment},
				TotalCount: 1,
			},
			ExpectError: nil,
		},
	}

	for _, test := range tests {
		response, err := resource.ListResources(test.Namespace, test.Resource, test.Conditions, test.OrderBy, test.Reverse, test.Limit, test.Offset)
		if err != test.ExpectError {
			t.Fatalf("expected error: %s, got: %s", test.ExpectError, err)
		}
		if diff := cmp.Diff(test.ExpectResponse, response); diff != "" {
			t.Errorf(diff)
		}
	}

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
)

func prepare() (informers.InformerFactory, error) {

	namespaces := []interface{}{defaultNamespace, kubesphereNamespace}
	deployments := []interface{}{nginxDeployment, redisDeployment}

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	istioClient := fakeistio.NewSimpleClientset()
	appClient := fakeapp.NewSimpleClientset()
	snapshotClient := fakesnapshot.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, istioClient, appClient, snapshotClient, nil)

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

	return fakeInformerFactory, nil
}
