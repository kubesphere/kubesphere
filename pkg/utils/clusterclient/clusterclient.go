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
	"net/http"
	"net/url"
	"reflect"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	clusterinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/cluster/v1alpha1"
	clusterlister "kubesphere.io/kubesphere/pkg/client/listers/cluster/v1alpha1"
)

type innerCluster struct {
	KubernetesURL *url.URL
	KubesphereURL *url.URL
	Transport     http.RoundTripper
}

type clusterClients struct {
	sync.RWMutex
	clusterLister clusterlister.ClusterLister

	// build a in memory cluster cache to speed things up
	innerClusters map[string]*innerCluster
}

type ClusterClients interface {
	IsHostCluster(cluster *clusterv1alpha1.Cluster) bool
	IsClusterReady(cluster *clusterv1alpha1.Cluster) bool
	GetClusterKubeconfig(string) (string, error)
	Get(string) (*clusterv1alpha1.Cluster, error)
	GetInnerCluster(string) *innerCluster
	GetKubernetesClientSet(string) (*kubernetes.Clientset, error)
	GetKubeSphereClientSet(string) (*kubesphere.Clientset, error)
}

func NewClusterClient(clusterInformer clusterinformer.ClusterInformer) ClusterClients {
	c := &clusterClients{
		innerClusters: make(map[string]*innerCluster),
		clusterLister: clusterInformer.Lister(),
	}

	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.addCluster(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldCluster := oldObj.(*clusterv1alpha1.Cluster)
			newCluster := newObj.(*clusterv1alpha1.Cluster)
			if !reflect.DeepEqual(oldCluster.Spec, newCluster.Spec) {
				c.addCluster(newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			c.removeCluster(obj)
		},
	})
	return c
}

func (c *clusterClients) removeCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)
	klog.V(4).Infof("remove cluster %s", cluster.Name)
	c.Lock()
	delete(c.innerClusters, cluster.Name)
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

func (c *clusterClients) addCluster(obj interface{}) *innerCluster {
	cluster := obj.(*clusterv1alpha1.Cluster)
	klog.V(4).Infof("add new cluster %s", cluster.Name)
	_, err := url.Parse(cluster.Spec.Connection.KubernetesAPIEndpoint)
	if err != nil {
		klog.Errorf("Parse kubernetes apiserver endpoint %s failed, %v", cluster.Spec.Connection.KubernetesAPIEndpoint, err)
		return nil
	}

	inner := newInnerCluster(cluster)
	c.Lock()
	c.innerClusters[cluster.Name] = inner
	c.Unlock()
	return inner
}

func (c *clusterClients) Get(clusterName string) (*clusterv1alpha1.Cluster, error) {
	return c.clusterLister.Get(clusterName)
}

func (c *clusterClients) GetClusterKubeconfig(clusterName string) (string, error) {
	cluster, err := c.clusterLister.Get(clusterName)
	if err != nil {
		return "", err
	}
	return string(cluster.Spec.Connection.KubeConfig), nil
}

func (c *clusterClients) GetInnerCluster(name string) *innerCluster {
	c.RLock()
	defer c.RUnlock()
	if inner, ok := c.innerClusters[name]; ok {
		return inner
	} else if cluster, err := c.clusterLister.Get(name); err == nil {
		// double check if the cluster exists but is not cached
		return c.addCluster(cluster)
	}
	return nil
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

func (c *clusterClients) GetKubeSphereClientSet(name string) (*kubesphere.Clientset, error) {
	kubeconfig, err := c.GetClusterKubeconfig(name)
	if err != nil {
		return nil, err
	}
	restConfig, err := newRestConfigFromString(kubeconfig)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubesphere.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func (c *clusterClients) GetKubernetesClientSet(name string) (*kubernetes.Clientset, error) {
	kubeconfig, err := c.GetClusterKubeconfig(name)
	if err != nil {
		return nil, err
	}
	restConfig, err := newRestConfigFromString(kubeconfig)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func newRestConfigFromString(kubeconfig string) (*rest.Config, error) {
	bytes, err := clientcmd.NewClientConfigFromBytes([]byte(kubeconfig))
	if err != nil {
		return nil, err
	}
	return bytes.ClientConfig()
}
