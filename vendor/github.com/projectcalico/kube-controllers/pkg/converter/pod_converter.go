// Copyright (c) 2017 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package converter

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	api "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/k8s/conversion"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// WorkloadEndpointData is an internal struct used to store the various bits
// of information that the policy controller cares about on a workload endpoint.
type WorkloadEndpointData struct {
	PodName   string
	Namespace string
	Labels    map[string]string
}

type podConverter struct {
}

// BuildWorkloadEndpointData generates the correct WorkloadEndpointData for the given
// WorkloadEndpoint, extracting fields that the policy controller is responsible for syncing.
func BuildWorkloadEndpointData(wep api.WorkloadEndpoint) WorkloadEndpointData {
	return WorkloadEndpointData{
		PodName:   wep.Spec.Pod,
		Namespace: wep.Namespace,
		Labels:    wep.Labels,
	}
}

// MergeWorkloadEndpointData applies the given WorkloadEndpointData to the provided
// WorkloadEndpoint, updating relevant fields with new values.
func MergeWorkloadEndpointData(wep *api.WorkloadEndpoint, upd WorkloadEndpointData) {
	if wep.Spec.Pod != upd.PodName || wep.Namespace != upd.Namespace {
		log.Fatalf("Bad attempt to merge data for %s/%s into wep %s/%s", upd.PodName, upd.Namespace, wep.Name, wep.Namespace)
	}
	wep.Labels = upd.Labels
}

// NewPodConverter Constructor for podConverter
func NewPodConverter() Converter {
	return &podConverter{}
}

func (p *podConverter) Convert(k8sObj interface{}) (interface{}, error) {
	// Convert Pod into a workload endpoint.
	var c conversion.Converter
	pod, ok := k8sObj.(*v1.Pod)
	if !ok {
		tombstone, ok := k8sObj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, errors.New("couldn't get object from tombstone")
		}
		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			return nil, errors.New("tombstone contained object that is not a Pod")
		}
	}

	// The conversion logic always requires a node, but we don't always have one. We don't actually
	// care about the value used for the node in this controller, so just dummy it out if it doesn't exist.
	if pod.Spec.NodeName == "" {
		pod.Spec.NodeName = "unknown.node"
	}

	kvp, err := c.PodToWorkloadEndpoint(pod)
	if err != nil {
		return nil, err
	}
	wep := kvp.Value.(*api.WorkloadEndpoint)

	// Build and return a WorkloadEndpointData struct using the data.
	return BuildWorkloadEndpointData(*wep), nil
}

// GetKey takes a WorkloadEndpointData and returns the key which
// identifies it - namespace/name
func (p *podConverter) GetKey(obj interface{}) string {
	e := obj.(WorkloadEndpointData)
	return fmt.Sprintf("%s/%s", e.Namespace, e.PodName)
}

func (p *podConverter) DeleteArgsFromKey(key string) (string, string) {
	// We don't have enough information to generate the delete args from the key that's used
	// for Pods / WorkloadEndpoints, so just panic. This should never be called but is necessary
	// to satisfy the interface.
	log.Panicf("DeleteArgsFromKey call for WorkloadEndpoints is not allowed")
	return "", ""
}
