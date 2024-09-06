/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import "time"

// ComponentStatus represents system component status.
type ComponentStatus struct {
	Name            string      `json:"name" description:"component name"`
	Namespace       string      `json:"namespace" description:"the name of the namespace"`
	SelfLink        string      `json:"selfLink" description:"self link"`
	Label           interface{} `json:"label" description:"labels"`
	StartedAt       time.Time   `json:"startedAt" description:"started time"`
	TotalBackends   int         `json:"totalBackends" description:"the total replicas of each backend system component"`
	HealthyBackends int         `json:"healthyBackends" description:"the number of healthy backend components"`
}

// NodeStatus assembles cluster nodes status, simply wrap unhealthy and total nodes.
type NodeStatus struct {
	// total nodes of cluster, including master nodes
	TotalNodes int `json:"totalNodes" description:"total number of nodes"`

	// healthy nodes means nodes whose state is NodeReady
	HealthyNodes int `json:"healthyNodes" description:"the number of healthy nodes"`
}

type HealthStatus struct {
	KubeSphereComponents []ComponentStatus `json:"kubesphereStatus" description:"kubesphere components status"`
	NodeStatus           NodeStatus        `json:"nodeStatus" description:"nodes status"`
}
