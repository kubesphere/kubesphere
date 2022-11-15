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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"unsafe"

	"github.com/emicklei/go-restful"
	"k8s.io/klog/v2"

	"github.com/google/go-cmp/cmp"
	fakesnapshot "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/fake"
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
	"kubesphere.io/kubesphere/pkg/models/components"
	resourcev1alpha2 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha2/resource"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
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
			namespace:   "",
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

	handler, err := prepare()
	if err != nil {
		t.Fatal(err)
	}

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
	apiServerService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ks-apiserver",
			Namespace: "istio-system",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "ks-apiserver-app"},
		},
	}
	ksControllerService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ks-controller",
			Namespace: "kubesphere-system",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "ks-controller-app"},
		},
	}
	deployments = []interface{}{redisDeployment, nginxDeployment}
	namespaces  = []interface{}{defaultNamespace, kubesphereNamespace}
	secrets     = []interface{}{secretFoo1, secretFoo2}
	services    = []interface{}{apiServerService, ksControllerService}
)

func prepare() (*Handler, error) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	istioClient := fakeistio.NewSimpleClientset()
	snapshotClient := fakesnapshot.NewSimpleClientset()
	apiextensionsClient := fakeapiextensions.NewSimpleClientset()

	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, istioClient, snapshotClient, apiextensionsClient, nil)

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
	for _, service := range services {
		err := k8sInformerFactory.Core().V1().Services().Informer().GetIndexer().Add(service)
		if err != nil {
			return nil, err
		}
	}

	handler := New(resourcev1alpha3.NewResourceGetter(fakeInformerFactory, nil), resourcev1alpha2.NewResourceGetter(fakeInformerFactory), components.NewComponentsGetter(fakeInformerFactory.KubernetesSharedInformerFactory()))

	return handler, nil
}

func TestHandleGetComponentStatus(t *testing.T) {
	param := map[string]string{
		"component": "ks-controller",
	}
	request, response, err := buildReqAndRes("GET", "/kapis/resources.kubesphere.io/v1alpha3/components/{component}", param, nil)
	if err != nil {
		t.Fatal("build res or req failed ")
	}
	handler, err := prepare()
	if err != nil {
		t.Fatal("init handler failed")
	}

	handler.handleGetComponentStatus(request, response)

	if status := response.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestHandleGetComponents(t *testing.T) {
	request, response, err := buildReqAndRes("GET", "/kapis/resources.kubesphere.io/v1alpha3/components", nil, nil)
	if err != nil {
		t.Fatal("build res or req failed ")
	}
	handler, err := prepare()
	if err != nil {
		t.Fatal("init handler failed")
	}

	handler.handleGetComponents(request, response)

	if status := response.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

//build req and res in *restful
func buildReqAndRes(method, target string, param map[string]string, body io.Reader) (*restful.Request, *restful.Response, error) {
	//build req
	request := httptest.NewRequest(method, target, body)
	newRequest := restful.NewRequest(request)
	if param != nil {
		err := setUnExportedFields(newRequest, "pathParameters", param)
		if err != nil {
			klog.Error("set pathParameters failed ")
			return nil, nil, err
		}
	}
	//build res
	response := httptest.NewRecorder()
	newResponse := restful.NewResponse(response)

	// assign  Key:routeProduces a value of "application/json"
	err := setUnExportedFields(newResponse, "routeProduces", []string{"application/json"})
	if err != nil {
		klog.Error("set routeProduces failed ")
		return nil, nil, err
	}
	return newRequest, newResponse, nil
}

//Setting unexported fields by using reflect
func setUnExportedFields(ptr interface{}, filedName string, newFiledValue interface{}) (err error) {
	v := reflect.ValueOf(ptr).Elem().FieldByName(filedName)
	v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	nv := reflect.ValueOf(newFiledValue)

	if v.Kind() != nv.Kind() {
		return fmt.Errorf("kind error")
	}
	v.Set(nv)
	return nil
}
