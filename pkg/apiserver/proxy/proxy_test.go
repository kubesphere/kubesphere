package proxy

import (
	restful "github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyService(t *testing.T) {
	client := fake.NewSimpleClientset()
	config := &rest.Config{Host: "http://fake-apiserver"}
	handler := NewProxyHandler(client, config)

	req := restful.NewRequest(httptest.NewRequest("GET", "/proxy/namespaces/default/services/test-service/test-path", nil))

	resp := restful.NewResponse(httptest.NewRecorder())

	handler.ProxyService(req, resp)
	// Since we use a fake client, just check for status code 502 (Bad Gateway)
	if resp.StatusCode() != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", resp.StatusCode())
	}
}
