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
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"

	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	clusterinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

const proxyURLFormat = "/api/v1/namespaces/kubesphere-system/services/:ks-apiserver:/proxy%s"

// Dispatcher defines how to forward request to designated cluster based on cluster name
// This should only be used in host cluster when multicluster mode enabled, use in any other cases may cause
// unexpected behavior
type Dispatcher interface {
	Dispatch(w http.ResponseWriter, req *http.Request, handler http.Handler)
}

type clusterDispatch struct {
	clusterclient.ClusterClients
}

func NewClusterDispatch(clusterInformer clusterinformer.ClusterInformer) Dispatcher {
	return &clusterDispatch{clusterclient.NewClusterClient(clusterInformer)}
}

// Dispatch dispatch requests to designated cluster
func (c *clusterDispatch) Dispatch(w http.ResponseWriter, req *http.Request, handler http.Handler) {
	info, _ := request.RequestInfoFrom(req.Context())

	if len(info.Cluster) == 0 {
		klog.Warningf("Request with empty cluster, %v", req.URL)
		http.Error(w, fmt.Sprintf("Bad request, empty cluster"), http.StatusBadRequest)
		return
	}

	cluster, err := c.Get(info.Cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, fmt.Sprintf("cluster %s not found", info.Cluster), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// request cluster is host cluster, no need go through agent
	if c.IsHostCluster(cluster) {
		req.URL.Path = strings.Replace(req.URL.Path, fmt.Sprintf("/clusters/%s", info.Cluster), "", 1)
		handler.ServeHTTP(w, req)
		return
	}

	if !c.IsClusterReady(cluster) {
		http.Error(w, fmt.Sprintf("cluster %s is not ready", cluster.Name), http.StatusBadRequest)
		return
	}

	innCluster := c.GetInnerCluster(cluster.Name)
	if innCluster == nil {
		http.Error(w, fmt.Sprintf("cluster %s is not ready", cluster.Name), http.StatusBadRequest)
		return
	}

	transport := http.DefaultTransport

	// change request host to actually cluster hosts
	u := *req.URL
	u.Path = strings.Replace(u.Path, fmt.Sprintf("/clusters/%s", info.Cluster), "", 1)

	// if cluster connection is direct and kubesphere apiserver endpoint is empty
	// we use kube-apiserver proxy way
	if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeDirect &&
		len(cluster.Spec.Connection.KubeSphereAPIEndpoint) == 0 {

		u.Scheme = innCluster.KubernetesURL.Scheme
		u.Host = innCluster.KubernetesURL.Host
		u.Path = fmt.Sprintf(proxyURLFormat, u.Path)
		transport = innCluster.Transport

		// The reason we need this is kube-apiserver doesn't behave like a standard proxy, it will strip
		// authorization header of proxy requests. Use custom header to avoid stripping by kube-apiserver.
		// https://github.com/kubernetes/kubernetes/issues/38775#issuecomment-277915961
		// We first copy req.Header['Authorization'] to req.Header['X-KubeSphere-Authorization'] before sending
		// designated cluster kube-apiserver, then copy req.Header['X-KubeSphere-Authorization'] to
		// req.Header['Authorization'] before authentication.
		req.Header.Set("X-KubeSphere-Authorization", req.Header.Get("Authorization"))

		// If cluster kubeconfig using token authentication, transport will not override authorization header,
		// this will cause requests reject by kube-apiserver since kubesphere authorization header is not
		// acceptable. Delete this header is safe since we are using X-KubeSphere-Authorization.
		// https://github.com/kubernetes/client-go/blob/master/transport/round_trippers.go#L285
		req.Header.Del("Authorization")

		// Dirty trick again. The kube-apiserver apiserver proxy rejects all proxy requests with dryRun parameter
		// https://github.com/kubernetes/kubernetes/pull/66083
		// Really don't understand why they do this. And here we are, bypass with replacing 'dryRun'
		// with dryrun and switch bach before send to kube-apiserver on the other side.
		if len(u.Query()["dryRun"]) != 0 {
			req.URL.RawQuery = strings.Replace(req.URL.RawQuery, "dryRun", "dryrun", 1)
		}

		// kube-apiserver lost query string when proxy websocket requests, there are several issues opened
		// tracking this, like https://github.com/kubernetes/kubernetes/issues/89360. Also there is a promising
		// PR aim to fix this, but it's unlikely it will get merged soon. So here we are again. Put raw query
		// string in Header and extract it on member cluster.
		if httpstream.IsUpgradeRequest(req) && len(req.URL.RawQuery) != 0 {
			req.Header.Set("X-KubeSphere-Rawquery", req.URL.RawQuery)
		}
	} else {
		// everything else goes to ks-apiserver, since our ks-apiserver has the ability to proxy kube-apiserver requests

		u.Host = innCluster.KubesphereURL.Host
		u.Scheme = innCluster.KubesphereURL.Scheme
	}

	httpProxy := proxy.NewUpgradeAwareHandler(&u, transport, false, false, c)
	httpProxy.UpgradeTransport = proxy.NewUpgradeRequestRoundTripper(transport, transport)
	httpProxy.ServeHTTP(w, req)
}

func (c *clusterDispatch) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}
