// Copyright (c) 2016-2019 Tigera, Inc. All rights reserved.

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

package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	kapiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/k8s/conversion"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"
	"k8s.io/apimachinery/pkg/types"
)

func NewWorkloadEndpointClient(c *kubernetes.Clientset) K8sResourceClient {
	return &WorkloadEndpointClient{
		clientSet: c,
		converter: conversion.Converter{},
	}
}

// Implements the api.Client interface for WorkloadEndpoints.
type WorkloadEndpointClient struct {
	clientSet *kubernetes.Clientset
	converter conversion.Converter
}

func (c *WorkloadEndpointClient) Create(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	log.Debug("Received Create request on WorkloadEndpoint type")
	// As a special case for the CNI plugin, try to patch the Pod with the IP that we've calculated.
	// This works around a bug in kubelet that causes it to delay writing the Pod IP for a long time:
	// https://github.com/kubernetes/kubernetes/issues/39113.
	//
	// Note: it's a bit odd to do this in the Create, but the CNI plugin uses CreateOrUpdate().  Doing it
	// here makes sure that, if the update fails: we retry here, and, we don't report success without
	// making the patch.
	return c.patchPodIP(ctx, kvp)
}

func (c *WorkloadEndpointClient) Update(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	log.Debug("Received Update request on WorkloadEndpoint type")
	// As a special case for the CNI plugin, try to patch the Pod with the IP that we've calculated.
	// This works around a bug in kubelet that causes it to delay writing the Pod IP for a long time:
	// https://github.com/kubernetes/kubernetes/issues/39113.
	return c.patchPodIP(ctx, kvp)
}

func (c *WorkloadEndpointClient) DeleteKVP(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	return c.Delete(ctx, kvp.Key, kvp.Revision, kvp.UID)
}

func (c *WorkloadEndpointClient) Delete(ctx context.Context, key model.Key, revision string, uid *types.UID) (*model.KVPair, error) {
	log.Warn("Operation Delete is not supported on WorkloadEndpoint type")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: key,
		Operation:  "Delete",
	}
}

// patchPodIP PATCHes the Kubernetes Pod associated with the given KVPair with the IP address it contains.
// This is a no-op if there is no IP address.
//
// We store the IP address in an annotation because patching the PodIP directly races with changes that
// kubelet makes so kubelet can undo our changes.
func (c *WorkloadEndpointClient) patchPodIP(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	ips := kvp.Value.(*apiv3.WorkloadEndpoint).Spec.IPNetworks
	if len(ips) == 0 {
		return kvp, nil
	}

	log.Debugf("PATCHing pod with IP: %v", ips[0])
	wepID, err := c.converter.ParseWorkloadEndpointName(kvp.Key.(model.ResourceKey).Name)
	if err != nil {
		return nil, err
	}
	if wepID.Pod == "" {
		return nil, cerrors.ErrorInsufficientIdentifiers{Name: kvp.Key.(model.ResourceKey).Name}
	}
	// Write the IP address into an annotation.  This generates an event more quickly than
	// waiting for kubelet to update the Status.PodIP field.
	ns := kvp.Key.(model.ResourceKey).Namespace
	patch, err := calculateAnnotationPatch(conversion.AnnotationPodIP, ips[0])
	if err != nil {
		log.WithError(err).Error("Failed to calculate Pod patch.")
		return nil, err
	}
	pod, err := c.clientSet.CoreV1().Pods(ns).Patch(wepID.Pod, types.StrategicMergePatchType, patch, "status")
	if err != nil {
		return nil, K8sErrorToCalico(err, kvp.Key)
	}
	log.Debugf("Successfully PATCHed pod to add podIP annotation: %+v", pod)
	return c.converter.PodToWorkloadEndpoint(pod)
}

const annotationPatchTemplate = `{"metadata": {"annotations": {%s: %s}}}`

func calculateAnnotationPatch(name, value string) ([]byte, error) {
	// Marshal the key and value in order to make sure all the escaping is done correctly.
	nameJson, err := json.Marshal(name)
	if err != nil {
		return nil, err
	}
	valueJson, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	patch := []byte(fmt.Sprintf(annotationPatchTemplate, nameJson, valueJson))
	return patch, nil
}

func (c *WorkloadEndpointClient) Get(ctx context.Context, key model.Key, revision string) (*model.KVPair, error) {
	log.Debug("Received Get request on WorkloadEndpoint type")
	k := key.(model.ResourceKey)

	// Parse resource name so we can get get the podName
	wepID, err := c.converter.ParseWorkloadEndpointName(key.(model.ResourceKey).Name)
	if err != nil {
		return nil, err
	}
	if wepID.Pod == "" {
		return nil, cerrors.ErrorResourceDoesNotExist{
			Identifier: key,
			Err:        errors.New("malformed WorkloadEndpoint name - unable to determine Pod name"),
		}
	}

	pod, err := c.clientSet.CoreV1().Pods(k.Namespace).Get(wepID.Pod, metav1.GetOptions{ResourceVersion: revision})
	if err != nil {
		return nil, K8sErrorToCalico(err, k)
	}

	// Decide if this pod should be displayed.
	if !c.converter.IsValidCalicoWorkloadEndpoint(pod) {
		return nil, cerrors.ErrorResourceDoesNotExist{Identifier: k}
	}
	return c.converter.PodToWorkloadEndpoint(pod)
}

func (c *WorkloadEndpointClient) List(ctx context.Context, list model.ListInterface, revision string) (*model.KVPairList, error) {
	log.Debug("Received List request on WorkloadEndpoint type")
	l := list.(model.ResourceListOptions)

	// If a workload is provided, we can do an exact lookup of this
	// workload endpoint.
	if l.Name != "" {
		kvp, err := c.Get(ctx, model.ResourceKey{
			Name:      l.Name,
			Namespace: l.Namespace,
			Kind:      l.Kind,
		}, revision)
		if err != nil {
			switch err.(type) {
			// Return empty slice of KVPair if the object doesn't exist, return the error otherwise.
			case cerrors.ErrorResourceDoesNotExist:
				return &model.KVPairList{
					KVPairs:  []*model.KVPair{},
					Revision: revision,
				}, nil
			default:
				return nil, err
			}
		}

		return &model.KVPairList{
			KVPairs:  []*model.KVPair{kvp},
			Revision: revision,
		}, nil
	}

	// Otherwise, enumerate all pods in a namespace.
	pods, err := c.clientSet.CoreV1().Pods(l.Namespace).List(metav1.ListOptions{ResourceVersion: revision})
	if err != nil {
		return nil, K8sErrorToCalico(err, l)
	}

	// For each Pod, return a workload endpoint.
	ret := []*model.KVPair{}
	for _, pod := range pods.Items {
		// Decide if this pod should be included.
		if !c.converter.IsValidCalicoWorkloadEndpoint(&pod) {
			continue
		}

		kvp, err := c.converter.PodToWorkloadEndpoint(&pod)
		if err != nil {
			return nil, err
		}
		ret = append(ret, kvp)
	}
	return &model.KVPairList{
		KVPairs:  ret,
		Revision: revision,
	}, nil
}

func (c *WorkloadEndpointClient) EnsureInitialized() error {
	return nil
}

func (c *WorkloadEndpointClient) Watch(ctx context.Context, list model.ListInterface, revision string) (api.WatchInterface, error) {
	// Build watch options to pass to k8s.
	opts := metav1.ListOptions{ResourceVersion: revision, Watch: true}
	rlo, ok := list.(model.ResourceListOptions)
	if !ok {
		return nil, fmt.Errorf("ListInterface is not a ResourceListOptions: %s", list)
	}
	if len(rlo.Name) != 0 {
		if len(rlo.Namespace) == 0 {
			return nil, errors.New("cannot watch a specific WorkloadEndpoint without a namespace")
		}
		// We've been asked to watch a specific workloadendpoint
		wepids, err := c.converter.ParseWorkloadEndpointName(rlo.Name)
		if err != nil {
			return nil, err
		}
		log.WithField("name", wepids.Pod).Debug("Watching a single workloadendpoint")
		opts.FieldSelector = fields.OneTermEqualSelector("metadata.name", wepids.Pod).String()
	}

	ns := list.(model.ResourceListOptions).Namespace
	k8sWatch, err := c.clientSet.CoreV1().Pods(ns).Watch(opts)
	if err != nil {
		return nil, K8sErrorToCalico(err, list)
	}
	converter := func(r Resource) (*model.KVPair, error) {
		k8sPod, ok := r.(*kapiv1.Pod)
		if !ok {
			return nil, errors.New("Pod conversion with incorrect k8s resource type")
		}
		if !c.converter.IsValidCalicoWorkloadEndpoint(k8sPod) {
			// If this is not a valid Calico workload endpoint then don't return in the watch.
			// Returning a nil KVP and a nil error swallows the event.
			return nil, nil
		}
		return c.converter.PodToWorkloadEndpoint(k8sPod)
	}
	return newK8sWatcherConverter(ctx, "Pod", converter, k8sWatch), nil
}
