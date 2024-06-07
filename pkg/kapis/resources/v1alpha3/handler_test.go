/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
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

	"github.com/Masterminds/semver/v3"
	"github.com/emicklei/go-restful/v3"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	runtimefakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/components"
	v2 "kubesphere.io/kubesphere/pkg/models/registries/v2"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/scheme"
	"kubesphere.io/kubesphere/pkg/simple/client/overview"
)

func TestResourceV1alpha3(t *testing.T) {
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
				Items:      []runtime.Object{kubesphereNamespace, defaultNamespace},
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
				Items:      []runtime.Object{secretFoo2, secretFoo1},
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

func listResources(namespace, resourceType string, query *query.Query, h *handler) (*api.ListResult, error) {
	result, err := h.resourceGetterV1alpha3.List(resourceType, namespace, query)
	if err != nil {
		return nil, err
	}
	return result, nil
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
	deployments = []runtime.Object{redisDeployment, nginxDeployment}
	namespaces  = []runtime.Object{defaultNamespace, kubesphereNamespace}
	secrets     = []runtime.Object{secretFoo1, secretFoo2}
	services    = []runtime.Object{apiServerService, ksControllerService}
)

func prepare() (*handler, error) {
	client := runtimefakeclient.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithRuntimeObjects(namespaces...).
		WithRuntimeObjects(deployments...).
		WithRuntimeObjects(secrets...).
		WithRuntimeObjects(services...).
		Build()

	k8sVersion120, _ := semver.NewVersion("1.20.0")

	handler := &handler{
		resourceGetterV1alpha3: resourcev1alpha3.NewResourceGetter(client, k8sVersion120),
		componentsGetter:       components.NewComponentsGetter(client),
		registryHelper:         v2.NewRegistryHelper(),
		counter:                overview.New(client),
	}
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

	handler.GetComponentStatus(request, response)
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
	handler.GetComponents(request, response)
	if status := response.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// build req and res in *restful
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

// Setting unexported fields by using reflect
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
