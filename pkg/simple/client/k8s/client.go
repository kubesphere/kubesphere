/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */
/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ClientInterface is the interface for kubernetes client
type ClientInterface interface {
	// Kubernetes returns the kubernetes client
	Kubernetes() kubernetes.Interface
	// Config returns the rest client config
	Config() *rest.Config
	// Master returns the master URL
	Master() string
}

// KubernetesClient is the implementation of Client
type kubernetesClient struct {
	// kubernetes client
	k8s kubernetes.Interface
	// rest client config
	config *rest.Config
	// master URL
	master string
	// For compatibility with the old implementation
	Interface kubernetes.Interface
}

// Kubernetes returns the kubernetes client
func (c *kubernetesClient) Kubernetes() kubernetes.Interface {
	return c.k8s
}

// Config returns the rest client config
func (c *kubernetesClient) Config() *rest.Config {
	return c.config
}

// Master returns the master URL
func (c *kubernetesClient) Master() string {
	return c.master
}
package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var (
	// Global client instance
	client *Client
)

// Client represents a Kubernetes client
type Client struct {
	// Kubernetes client
	Kubernetes kubernetes.Interface
	// REST config
	Config *rest.Config
}

// Initialize initializes the global client
func Initialize() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client is the interface for kubernetes client
type Client interface {
	// Kubernetes returns the kubernetes client
	Kubernetes() kubernetes.Interface
	// Config returns the rest client config
	Config() *rest.Config
	// Master returns the master URL
	Master() string
}

// kubernetesClient is the implementation of Client
type kubernetesClient struct {
	// kubernetes client
	k8s kubernetes.Interface
	// rest client config
	config *rest.Config
	// master URL
	master string
	// For compatibility with the old implementation
	Interface kubernetes.Interface
}
/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

// This file is intentionally empty.
// All Client interfaces and implementations have been moved to kubernetes.go
// to avoid duplication and redeclaration errors.
package k8s
// Kubernetes returns the kubernetes client
func (c *kubernetesClient) Kubernetes() kubernetes.Interface {
	return c.k8s
}
/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client is the interface for kubernetes client
type Client interface {
	// Kubernetes returns the kubernetes client
	Kubernetes() kubernetes.Interface
	// Config returns the rest client config
	Config() *rest.Config
	// Master returns the master URL
	Master() string
}

// kubernetesClient is the implementation of Client
type kubernetesClient struct {
	// kubernetes client
	k8s kubernetes.Interface
	// rest client config
	config *rest.Config
	// master URL
	master string
	// For compatibility with the old implementation
	Interface kubernetes.Interface
}

// Kubernetes returns the kubernetes client
func (c *kubernetesClient) Kubernetes() kubernetes.Interface {
	return c.k8s
}

// Config returns the rest client config
func (c *kubernetesClient) Config() *rest.Config {
	return c.config
}

// Master returns the master URL
func (c *kubernetesClient) Master() string {
	return c.master
}
// Config returns the rest client config
func (c *kubernetesClient) Config() *rest.Config {
	return c.config
}

// Master returns the master URL
func (c *kubernetesClient) Master() string {
	return c.master
}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	client = &Client{
		Kubernetes: clientset,
		Config:     config,
	}
	return nil
}

// Client returns the global client instance, initializing it if necessary
func Client() *Client {
	if client == nil {
		err := Initialize()
		if err != nil {
			klog.Errorf("Failed to initialize Kubernetes client: %v", err)
			// Return a default client for testing - this is not ideal in production
			defaultConfig := &rest.Config{Host: "http://localhost:8080"}
			client = &Client{Config: defaultConfig}
		}
	}
	return client
}
