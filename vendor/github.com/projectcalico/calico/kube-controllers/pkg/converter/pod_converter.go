// Copyright (c) 2017-2020 Tigera, Inc. All rights reserved.
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

	"github.com/projectcalico/calico/libcalico-go/lib/backend/model"

	log "github.com/sirupsen/logrus"

	api "github.com/projectcalico/calico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/calico/libcalico-go/lib/backend/k8s/conversion"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// WorkloadEndpointData is an internal struct used to store the various bits
// of information that the policy controller cares about on a workload endpoint.
type WorkloadEndpointData struct {
	PodName        string
	Namespace      string
	Labels         map[string]string
	ServiceAccount string
}

type PodConverter interface {
	Convert(k8sObj interface{}) ([]WorkloadEndpointData, error)
	GetKey(obj WorkloadEndpointData) string
	DeleteArgsFromKey(key string) (string, string)
}

type podConverter struct{}

// BuildWorkloadEndpointData generates the correct WorkloadEndpointData for the given
// list of WorkloadEndpoints, extracting fields that the policy controller is responsible
// for syncing.
func BuildWorkloadEndpointData(weps ...api.WorkloadEndpoint) []WorkloadEndpointData {
	var retWEPs []WorkloadEndpointData
	for _, wep := range weps {
		retWEPs = append(retWEPs, WorkloadEndpointData{
			PodName:        wep.Spec.Pod,
			Namespace:      wep.Namespace,
			Labels:         wep.Labels,
			ServiceAccount: wep.Spec.ServiceAccountName,
		})
	}

	return retWEPs
}

// MergeWorkloadEndpointData applies the given WorkloadEndpointData to the provided
// WorkloadEndpoint, updating relevant fields with new values.
func MergeWorkloadEndpointData(wep *api.WorkloadEndpoint, upd WorkloadEndpointData) {
	if wep.Spec.Pod != upd.PodName || wep.Namespace != upd.Namespace {
		log.Fatalf("Bad attempt to merge data for %s/%s into wep %s/%s", upd.PodName, upd.Namespace, wep.Name, wep.Namespace)
	}
	wep.Labels = upd.Labels
	wep.Spec.ServiceAccountName = upd.ServiceAccount
}

// NewPodConverter Constructor for podConverter
func NewPodConverter() PodConverter {
	return &podConverter{}
}

func (p *podConverter) Convert(k8sObj interface{}) ([]WorkloadEndpointData, error) {
	// Convert Pod into a workload endpoint.
	c := conversion.NewConverter()
	pod, err := ExtractPodFromUpdate(k8sObj)
	if err != nil {
		return nil, err
	}

	// The conversion logic always requires a node, but we don't always have one. We don't actually
	// care about the value used for the node in this controller, so just dummy it out if it doesn't exist.
	if pod.Spec.NodeName == "" {
		pod.Spec.NodeName = "unknown.node"
	}

	kvps, err := c.PodToWorkloadEndpoints(pod)
	if err != nil {
		return nil, err
	}

	// Build and return a WorkloadEndpointData struct using the data.
	return BuildWorkloadEndpointData(kvpsToWEPs(kvps)...), nil
}

func kvpsToWEPs(kvps []*model.KVPair) []api.WorkloadEndpoint {
	var weps []api.WorkloadEndpoint
	for _, kvp := range kvps {
		wep := kvp.Value.(*api.WorkloadEndpoint)
		if wep != nil {
			weps = append(weps, *wep)
		}
	}

	return weps
}

// GetKey takes a WorkloadEndpointData and returns the key which
// identifies it - namespace/name
func (p *podConverter) GetKey(obj WorkloadEndpointData) string {
	return fmt.Sprintf("%s/%s", obj.Namespace, obj.PodName)
}

func (p *podConverter) DeleteArgsFromKey(key string) (string, string) {
	// We don't have enough information to generate the delete args from the key that's used
	// for Pods / WorkloadEndpoints, so just panic. This should never be called but is necessary
	// to satisfy the interface.
	log.Panicf("DeleteArgsFromKey call for WorkloadEndpoints is not allowed")
	return "", ""
}

// ExtractPodFromUpdate takes an update as received from the informer and returns the pod object, if present.
// some updates (particularly deletes) can include tombstone placeholders rather than an exact pod object. This
// function should be called in order to safely handles those cases.
func ExtractPodFromUpdate(obj interface{}) (*v1.Pod, error) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, errors.New("couldn't get object from tombstone")
		}
		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			return nil, errors.New("tombstone contained object that is not a Pod")
		}
	}
	return pod, nil
}
