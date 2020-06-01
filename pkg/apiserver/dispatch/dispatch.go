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

package dispatch

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	clusterinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/cluster/v1alpha1"
	clusterlister "kubesphere.io/kubesphere/pkg/client/listers/cluster/v1alpha1"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const proxyURLFormat = "/api/v1/namespaces/kubesphere-system/services/:ks-apiserver:/proxy%s"

// Dispatcher defines how to forward request to designated cluster based on cluster name
type Dispatcher interface {
	Dispatch(w http.ResponseWriter, req *http.Request, handler http.Handler)
}

type innerCluster struct {
	kubernetesURL *url.URL
	kubesphereURL *url.URL
	transport     http.RoundTripper
}

type clusterDispatch struct {
	clusterLister clusterlister.ClusterLister

	// dispatcher will build a in memory cluster cache to speed things up
	innerClusters map[string]*innerCluster

	clusterInformerSynced cache.InformerSynced

	mutex sync.RWMutex
}

func NewClusterDispatch(clusterInformer clusterinformer.ClusterInformer, clusterLister clusterlister.ClusterLister) Dispatcher {
	clusterDispatcher := &clusterDispatch{
		clusterLister: clusterLister,
		innerClusters: make(map[string]*innerCluster),
		mutex:         sync.RWMutex{},
	}

	clusterDispatcher.clusterInformerSynced = clusterInformer.Informer().HasSynced
	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: clusterDispatcher.updateInnerClusters,
		UpdateFunc: func(oldObj, newObj interface{}) {
			clusterDispatcher.updateInnerClusters(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			cluster := obj.(*clusterv1alpha1.Cluster)
			clusterDispatcher.mutex.Lock()
			if _, ok := clusterDispatcher.innerClusters[cluster.Name]; ok {
				delete(clusterDispatcher.innerClusters, cluster.Name)
			}
			clusterDispatcher.mutex.Unlock()

		},
	})

	return clusterDispatcher
}

// Dispatch dispatch requests to designated cluster
func (c *clusterDispatch) Dispatch(w http.ResponseWriter, req *http.Request, handler http.Handler) {
	info, _ := request.RequestInfoFrom(req.Context())

	if len(info.Cluster) == 0 {
		klog.Warningf("Request with empty cluster, %v", req.URL)
		http.Error(w, fmt.Sprintf("Bad request, empty cluster"), http.StatusBadRequest)
		return
	}

	cluster, err := c.clusterLister.Get(info.Cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, fmt.Sprintf("cluster %s not found", info.Cluster), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// request cluster is host cluster, no need go through agent
	if isClusterHostCluster(cluster) {
		req.URL.Path = strings.Replace(req.URL.Path, fmt.Sprintf("/clusters/%s", info.Cluster), "", 1)
		handler.ServeHTTP(w, req)
		return
	}

	if !isClusterReady(cluster) {
		http.Error(w, fmt.Sprintf("cluster is not ready"), http.StatusInternalServerError)
		return
	}

	innCluster := c.getInnerCluster(cluster.Name)
	if innCluster == nil {
		http.Error(w, fmt.Sprintf("cluster not ready"), http.StatusInternalServerError)
		return
	}

	transport := http.DefaultTransport

	u := *req.URL
	u.Path = strings.Replace(u.Path, fmt.Sprintf("/clusters/%s", info.Cluster), "", 1)

	if info.IsKubernetesRequest {
		u.Host = innCluster.kubernetesURL.Host
		u.Scheme = innCluster.kubernetesURL.Scheme
	} else {
		u.Host = innCluster.kubesphereURL.Host

		// if cluster connection is direct and kubesphere apiserver endpoint is empty
		// we use kube-apiserver proxy
		if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeDirect &&
			len(cluster.Spec.Connection.KubeSphereAPIEndpoint) == 0 {

			u.Scheme = innCluster.kubernetesURL.Scheme
			u.Host = innCluster.kubernetesURL.Host
			u.Path = fmt.Sprintf(proxyURLFormat, u.Path)
			transport = innCluster.transport
		}
	}

	httpProxy := proxy.NewUpgradeAwareHandler(&u, transport, false, false, c)
	httpProxy.ServeHTTP(w, req)
}

func (c *clusterDispatch) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}

func (c *clusterDispatch) getInnerCluster(name string) *innerCluster {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if cluster, ok := c.innerClusters[name]; ok {
		return cluster
	}
	return nil
}

func (c *clusterDispatch) updateInnerClusters(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)

	kubernetesEndpoint, err := url.Parse(cluster.Spec.Connection.KubernetesAPIEndpoint)
	if err != nil {
		klog.Errorf("Parse kubernetes apiserver endpoint %s failed, %v", cluster.Spec.Connection.KubernetesAPIEndpoint, err)
		return
	}

	kubesphereEndpoint, err := url.Parse(cluster.Spec.Connection.KubeSphereAPIEndpoint)
	if err != nil {
		klog.Errorf("Parse kubesphere apiserver endpoint %s failed, %v", cluster.Spec.Connection.KubeSphereAPIEndpoint, err)
		return
	}

	// prepare for
	clientConfig, err := clientcmd.NewClientConfigFromBytes(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		klog.Errorf("Unable to create client config from kubeconfig bytes, %#v", err)
		return
	}

	clusterConfig, err := clientConfig.ClientConfig()
	if err != nil {
		klog.Errorf("Failed to get client config, %#v", err)
		return
	}

	transport, err := rest.TransportFor(clusterConfig)
	if err != nil {
		klog.Errorf("Create transport failed, %v", err)
	}

	c.mutex.Lock()
	c.innerClusters[cluster.Name] = &innerCluster{
		kubernetesURL: kubernetesEndpoint,
		kubesphereURL: kubesphereEndpoint,
		transport:     transport,
	}
	c.mutex.Unlock()
}

func isClusterReady(cluster *clusterv1alpha1.Cluster) bool {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == clusterv1alpha1.ClusterReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}

func isClusterHostCluster(cluster *clusterv1alpha1.Cluster) bool {
	for key, value := range cluster.Annotations {
		if key == clusterv1alpha1.IsHostCluster && value == "true" {
			return true
		}
	}

	return false
}
