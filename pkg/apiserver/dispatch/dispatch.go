package dispatch

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/client/listers/cluster/v1alpha1"
	"net/http"
	"net/url"
	"strings"
)

// Dispatcher defines how to forward request to designated cluster based on cluster name
type Dispatcher interface {
	Dispatch(w http.ResponseWriter, req *http.Request, handler http.Handler)
}

type clusterDispatch struct {
	clusterLister v1alpha1.ClusterLister
}

func NewClusterDispatch(clusterLister v1alpha1.ClusterLister) Dispatcher {
	return &clusterDispatch{
		clusterLister: clusterLister,
	}
}

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
		http.Error(w, fmt.Sprintf("cluster agent is not ready"), http.StatusInternalServerError)
		return
	}

	endpoint, err := url.Parse(cluster.Spec.Connection.KubeSphereAPIEndpoint)
	if err != nil {
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	u := *req.URL
	u.Host = endpoint.Host
	u.Path = strings.Replace(u.Path, fmt.Sprintf("/clusters/%s", info.Cluster), "", 1)

	httpProxy := proxy.NewUpgradeAwareHandler(&u, http.DefaultTransport, true, false, c)
	httpProxy.ServeHTTP(w, req)
}

func (c *clusterDispatch) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
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
