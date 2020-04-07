package dispatch

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/client/listers/cluster/v1alpha1"
	"net/http"
	"strings"
)

const defaultMultipleClusterAgentNamespace = "kubesphere-system"

// Dispatcher defines how to forward request to designated cluster based on cluster name
type Dispatcher interface {
	Dispatch(w http.ResponseWriter, req *http.Request, handler http.Handler)
}

type clusterDispatch struct {
	agentLister v1alpha1.AgentLister
}

func NewClusterDispatch(agentLister v1alpha1.AgentLister) Dispatcher {
	return &clusterDispatch{
		agentLister: agentLister,
	}
}

func (c *clusterDispatch) Dispatch(w http.ResponseWriter, req *http.Request, handler http.Handler) {

	info, _ := request.RequestInfoFrom(req.Context())
	if info.Cluster == "" { // fallback to host cluster if cluster name if empty
		handler.ServeHTTP(w, req)
		return
	}

	agent, err := c.agentLister.Get(info.Cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, fmt.Sprintf("cluster %s not found", info.Cluster), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if !isAgentReady(agent) {
		http.Error(w, fmt.Sprintf("cluster agent is not ready"), http.StatusInternalServerError)
		return
	}

	u := *req.URL
	u.Host = fmt.Sprintf("%s:%d", agent.Spec.Proxy, agent.Spec.KubeSphereAPIServerPort)
	u.Path = strings.Replace(u.Path, fmt.Sprintf("/clusters/%s", info.Cluster), "", 1)

	httpProxy := proxy.NewUpgradeAwareHandler(&u, http.DefaultTransport, true, false, c)
	httpProxy.ServeHTTP(w, req)
}

func (c *clusterDispatch) Error(w http.ResponseWriter, req *http.Request, err error) {
	responsewriters.InternalError(w, req, err)
}

func isAgentReady(agent *clusterv1alpha1.Agent) bool {
	for _, condition := range agent.Status.Conditions {
		if condition.Type == clusterv1alpha1.AgentConnected && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
