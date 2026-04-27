package v1alpha1

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// RegisterProxyHandlers registers the proxy handler with the Gin engine
func RegisterProxyHandlers(g *gin.Engine) {
	g.GET("/kapis/proxy/v1alpha1/namespaces/:namespace/services/:service/*path", handleServiceProxy)
}

func handleServiceProxy(c *gin.Context) {
	namespace := c.Param("namespace")
	service := c.Param("service")
	path := c.Param("path")
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// Use in-cluster config for Kubernetes API
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("Failed to get Kubernetes config: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	kubeURL := fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s/proxy/%s", config.Host, namespace, service, path)
	target, err := url.Parse(kubeURL)
	if err != nil {
		klog.Errorf("Invalid URL: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transport, err := rest.TransportFor(config)
	if err != nil {
		klog.Errorf("Failed to create transport: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport

	c.Request.URL = target
	c.Request.Host = target.Host
	if config.BearerToken != "" {
		c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.BearerToken))
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}
