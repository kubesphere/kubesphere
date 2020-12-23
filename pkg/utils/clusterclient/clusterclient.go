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

package clusterclient

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	clusterinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/cluster/v1alpha1"
	"net/http"
	"net/url"
	"sync"
)

var (
	ClusterNotExistsFormat = "cluster %s not exists"
)

type innerCluster struct {
	KubernetesURL *url.URL
	KubesphereURL *url.URL
	Transport     http.RoundTripper
}

type clusterClients struct {
	sync.RWMutex
	clusterMap        map[string]*clusterv1alpha1.Cluster
	clusterKubeconfig map[string]string

	// build a in memory cluster cache to speed things up
	innerClusters map[string]*innerCluster
}

type ClusterClients interface {
	IsHostCluster(cluster *clusterv1alpha1.Cluster) bool
	IsClusterReady(cluster *clusterv1alpha1.Cluster) bool
	GetClusterKubeconfig(string) (string, error)
	Get(string) (*clusterv1alpha1.Cluster, error)
	GetInnerCluster(string) *innerCluster
}

func (c *clusterClients) IsClusterReady(cluster *clusterv1alpha1.Cluster) bool {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == clusterv1alpha1.ClusterReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (c *clusterClients) IsHostCluster(cluster *clusterv1alpha1.Cluster) bool {
	if _, ok := cluster.Labels[clusterv1alpha1.HostCluster]; ok {
		return true
	}
	return false
}

func (c *clusterClients) GetInnerCluster(name string) *innerCluster {
	c.RLock()
	defer c.RUnlock()
	if cluster, ok := c.innerClusters[name]; ok {
		return cluster
	}
	return nil
}

var c *clusterClients
var lock sync.Mutex

func NewClusterClient(clusterInformer clusterinformer.ClusterInformer) ClusterClients {

	if c == nil {
		lock.Lock()
		defer lock.Unlock()

		if c != nil {
			return c
		}

		c = &clusterClients{
			clusterMap:        map[string]*clusterv1alpha1.Cluster{},
			clusterKubeconfig: map[string]string{},
			innerClusters:     make(map[string]*innerCluster),
		}

		clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.addCluster(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				c.removeCluster(oldObj)
				c.addCluster(newObj)
			},
			DeleteFunc: func(obj interface{}) {
				c.removeCluster(obj)
			},
		})
	}

	return c
}

func (c *clusterClients) removeCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)
	klog.V(4).Infof("remove cluster %s", cluster.Name)
	c.Lock()
	if _, ok := c.clusterMap[cluster.Name]; ok {
		delete(c.clusterMap, cluster.Name)
		delete(c.innerClusters, cluster.Name)
		delete(c.clusterKubeconfig, cluster.Name)
	}
	c.Unlock()
}

func newInnerCluster(cluster *clusterv1alpha1.Cluster) *innerCluster {
	kubernetesEndpoint, err := url.Parse(cluster.Spec.Connection.KubernetesAPIEndpoint)
	if err != nil {
		klog.Errorf("Parse kubernetes apiserver endpoint %s failed, %v", cluster.Spec.Connection.KubernetesAPIEndpoint, err)
		return nil
	}

	kubesphereEndpoint, err := url.Parse(cluster.Spec.Connection.KubeSphereAPIEndpoint)
	if err != nil {
		klog.Errorf("Parse kubesphere apiserver endpoint %s failed, %v", cluster.Spec.Connection.KubeSphereAPIEndpoint, err)
		return nil
	}

	// prepare for
	clientConfig, err := clientcmd.NewClientConfigFromBytes(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		klog.Errorf("Unable to create client config from kubeconfig bytes, %#v", err)
		return nil
	}

	clusterConfig, err := clientConfig.ClientConfig()
	if err != nil {
		klog.Errorf("Failed to get client config, %#v", err)
		return nil
	}

	transport, err := rest.TransportFor(clusterConfig)
	if err != nil {
		klog.Errorf("Create transport failed, %v", err)
		return nil
	}

	return &innerCluster{
		KubernetesURL: kubernetesEndpoint,
		KubesphereURL: kubesphereEndpoint,
		Transport:     transport,
	}
}

func (c *clusterClients) addCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)
	klog.V(4).Infof("add new cluster %s", cluster.Name)
	_, err := url.Parse(cluster.Spec.Connection.KubernetesAPIEndpoint)
	if err != nil {
		klog.Errorf("Parse kubernetes apiserver endpoint %s failed, %v", cluster.Spec.Connection.KubernetesAPIEndpoint, err)
		return
	}

	innerCluster := newInnerCluster(cluster)
	c.Lock()
	c.clusterMap[cluster.Name] = cluster
	c.clusterKubeconfig[cluster.Name] = string(cluster.Spec.Connection.KubeConfig)
	c.innerClusters[cluster.Name] = innerCluster
	c.Unlock()
}

func (c *clusterClients) GetClusterKubeconfig(clusterName string) (string, error) {
	c.RLock()
	defer c.RUnlock()
	if c, exists := c.clusterKubeconfig[clusterName]; exists {
		return c, nil
	} else {
		return "", fmt.Errorf(ClusterNotExistsFormat, clusterName)
	}
}

func (c *clusterClients) Get(clusterName string) (*clusterv1alpha1.Cluster, error) {
	c.RLock()
	defer c.RUnlock()
	if cluster, exists := c.clusterMap[clusterName]; exists {
		return cluster, nil
	} else {
		return nil, fmt.Errorf(ClusterNotExistsFormat, clusterName)
	}
}
