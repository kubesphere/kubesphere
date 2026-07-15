/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	restful "github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// ProxyHandler is the handler for proxying requests to kubernetes services
type ProxyHandler struct {
	k8sClient kubernetes.Interface
	config    *rest.Config
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(client kubernetes.Interface, config *rest.Config) *ProxyHandler {
	return &ProxyHandler{
		k8sClient: client,
		config:    config,
	}
}

// ProxyService handles proxying requests to kubernetes services
func (h *ProxyHandler) ProxyService(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	service := req.PathParameter("service")
	path := req.PathParameter("path")

	if namespace == "" || service == "" {
		http.Error(resp.ResponseWriter, "namespace and service must be specified", http.StatusBadRequest)
		return
	}

	// Build the URL to the Kubernetes API server
	kubeURL := fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s/proxy/%s", h.config.Host, namespace, service, path)
	target, err := url.Parse(kubeURL)
	if err != nil {
		klog.Errorf("Invalid URL: %v", err)
		http.Error(resp.ResponseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a transport that uses the API server authentication
	transport, err := rest.TransportFor(h.config)
	if err != nil {
		klog.Errorf("Failed to create transport: %v", err)
		http.Error(resp.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport

	// Update the request URL
	req.Request.URL = target
	req.Request.Host = target.Host

	// Add authorization header if needed
	if h.config.BearerToken != "" {
		req.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.config.BearerToken))
	}

	// Serve the request
	proxy.ServeHTTP(resp.ResponseWriter, req.Request)
}

// RegisterRoutes registers the proxy routes with the provided WebService
func (h *ProxyHandler) RegisterRoutes(ws *restful.WebService) {
	ws.Route(ws.GET("/namespaces/{namespace}/services/{service}/{path:*}").
		To(h.ProxyService).
		Doc("Proxy requests to Kubernetes services").
		Param(ws.PathParameter("namespace", "Namespace of the service")).
		Param(ws.PathParameter("service", "Name of the service")).
		Param(ws.PathParameter("path", "Path to proxy to")).
		Returns(http.StatusOK, "Success", nil).
		Returns(http.StatusBadRequest, "Bad request", nil).
		Returns(http.StatusInternalServerError, "Internal server error", nil))
}

// Register registers proxy handler with the given WebService
func Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/kapis/proxy/v1alpha1")

	// Create a client with in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to get in-cluster config: %v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	handler := NewProxyHandler(client, config)
	handler.RegisterRoutes(ws)

	container.Add(ws)
}
