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

//
type HealthStatus struct {
	KubeSphereComponents []ComponentStatus `json:"kubesphereStatus" description:"kubesphere components status"`
	NodeStatus           NodeStatus        `json:"nodeStatus" description:"nodes status"`
}
