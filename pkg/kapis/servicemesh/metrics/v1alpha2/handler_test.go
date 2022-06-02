package v1alpha2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/simple/client/kiali"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

func prepare() (*Handler, error) {
	var namespaceName = "kubesphere-system"
	var serviceAccountName = "kubesphere"
	var secretName = "kiali"
	clientset := fakek8s.NewSimpleClientset()

	ctx := context.Background()
	namespacesClient := clientset.CoreV1().Namespaces()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	_, err := namespacesClient.Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("create namespace failed ")
		return nil, err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespaceName,
		},
	}

	object := &corev1.ObjectReference{
		Name: secretName,
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespaceName,
		},
		Secrets: []corev1.ObjectReference{*object},
	}

	serviceAccountClient := clientset.CoreV1().ServiceAccounts(namespaceName)

	_, err = serviceAccountClient.Create(ctx, sa, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("create serviceAccount failed ")
		return nil, err
	}

	secretClient := clientset.CoreV1().Secrets(namespaceName)

	_, err = secretClient.Create(ctx, secret, metav1.CreateOptions{})

	if err != nil {
		klog.Errorf("create secret failed ")
		return nil, err
	}

	// mock jaeger server
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	options := &servicemesh.Options{
		IstioPilotHost:            "",
		KialiQueryHost:            "",
		JaegerQueryHost:           ts.URL,
		ServicemeshPrometheusHost: "",
	}
	handler := NewHandler(options, clientset, nil)

	token, _ := json.Marshal(
		&kiali.TokenResponse{
			Username: "test",
			Token:    "test",
		},
	)

	mc := &kiali.MockClient{
		TokenResult:   token,
		RequestResult: "fake",
	}

	client := kiali.NewClient("token", nil, mc, "token", options.KialiQueryHost)

	err = reflectutils.SetUnExportedField(handler, "client", client)
	if err != nil {
		klog.Errorf("apply mock client failed")
		return nil, err
	}
	return handler, nil
}

func TestGetServiceTracing(t *testing.T) {
	handler, err := prepare()
	if err != nil {
		t.Fatalf("init handler failed")
	}

	namespaceName := "namespace-test"
	serviceName := "service-test"
	url := fmt.Sprintf("/namespaces/%s/services/%s/traces", namespaceName, serviceName)
	request, _ := http.NewRequest("GET", url, nil)
	query := request.URL.Query()
	query.Add("start", "1650167872000000")
	query.Add("end", "1650211072000000")
	query.Add("limit", "10")
	request.URL.RawQuery = query.Encode()

	restfulRequest := restful.NewRequest(request)
	pathMap := make(map[string]string)
	pathMap["namespace"] = namespaceName
	pathMap["service"] = serviceName
	if err := reflectutils.SetUnExportedField(restfulRequest, "pathParameters", pathMap); err != nil {
		t.Fatalf("set pathParameters failed")
	}

	recorder := httptest.NewRecorder()
	restfulResponse := restful.NewResponse(recorder)
	restfulResponse.SetRequestAccepts("application/json")
	handler.GetServiceTracing(restfulRequest, restfulResponse)
	if status := restfulResponse.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
