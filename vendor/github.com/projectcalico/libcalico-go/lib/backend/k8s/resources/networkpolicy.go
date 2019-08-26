// Copyright (c) 2017-2019 Tigera, Inc. All rights reserved.

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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/k8s/conversion"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	NetworkPolicyResourceName = "NetworkPolicies"
	NetworkPolicyCRDName      = "networkpolicies.crd.projectcalico.org"
)

func NewNetworkPolicyClient(c *kubernetes.Clientset, r *rest.RESTClient) K8sResourceClient {
	crdClient := &customK8sResourceClient{
		restClient:      r,
		name:            NetworkPolicyCRDName,
		resource:        NetworkPolicyResourceName,
		description:     "Calico Network Policies",
		k8sResourceType: reflect.TypeOf(apiv3.NetworkPolicy{}),
		k8sResourceTypeMeta: metav1.TypeMeta{
			Kind:       apiv3.KindNetworkPolicy,
			APIVersion: apiv3.GroupVersionCurrent,
		},
		k8sListType:  reflect.TypeOf(apiv3.NetworkPolicyList{}),
		resourceKind: apiv3.KindNetworkPolicy,
		namespaced:   true,
	}
	return &networkPolicyClient{
		clientSet: c,
		crdClient: crdClient,
	}
}

// Implements the api.Client interface for NetworkPolicys.
type networkPolicyClient struct {
	conversion.Converter
	resourceName string
	clientSet    *kubernetes.Clientset
	crdClient    *customK8sResourceClient
}

func (c *networkPolicyClient) Create(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	log.Debug("Received Create request on NetworkPolicy type")
	key := kvp.Key.(model.ResourceKey)
	if strings.HasPrefix(key.Name, conversion.K8sNetworkPolicyNamePrefix) {
		// We don't support Create of a Kubernetes NetworkPolicy.
		return nil, cerrors.ErrorOperationNotSupported{
			Identifier: kvp.Key,
			Operation:  "Create",
		}
	}

	kvp, err := c.crdClient.Create(ctx, kvp)
	if kvp != nil {
		// Convert the revision to the combined CRD/k8s revision - the k8s rev will be empty, but this
		// format will allow the revision to be passed into List and Watch calls.
		kvp.Revision = c.JoinNetworkPolicyRevisions(kvp.Revision, "")
	}
	return kvp, err
}

func (c *networkPolicyClient) Update(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	log.Debug("Received Update request on NetworkPolicy type")

	key := kvp.Key.(model.ResourceKey)
	if strings.HasPrefix(key.Name, conversion.K8sNetworkPolicyNamePrefix) {
		// We don't support Update of a Kubernetes NetworkPolicy.
		return nil, cerrors.ErrorOperationNotSupported{
			Identifier: kvp.Key,
			Operation:  "Update",
		}
	}

	// The revision, if supplied, will be a combination of CRD and k8s-backed revisions.  Extract
	// the CRD rev and use that for the update.
	crdRev, _, err := c.SplitNetworkPolicyRevision(kvp.Revision)
	if err != nil {
		return nil, err
	}
	kvp.Revision = crdRev
	kvp, err = c.crdClient.Update(ctx, kvp)

	if kvp != nil {
		// Convert the revision back to the combined CRD/k8s revision - the k8s rev will be empty, but this
		// format will allow the revision to be passed into List and Watch calls.
		kvp.Revision = c.JoinNetworkPolicyRevisions(kvp.Revision, "")
	}
	return kvp, err
}

func (c *networkPolicyClient) Apply(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: kvp.Key,
		Operation:  "Apply",
	}
}
func (c *networkPolicyClient) DeleteKVP(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	return c.Delete(ctx, kvp.Key, kvp.Revision, kvp.UID)
}

func (c *networkPolicyClient) Delete(ctx context.Context, key model.Key, revision string, uid *types.UID) (*model.KVPair, error) {
	log.Debug("Received Delete request on NetworkPolicy type")
	k := key.(model.ResourceKey)
	if strings.HasPrefix(k.Name, conversion.K8sNetworkPolicyNamePrefix) {
		// We don't support Delete of a Kubernetes NetworkPolicy.
		return nil, cerrors.ErrorOperationNotSupported{
			Identifier: key,
			Operation:  "Delete",
		}
	}

	// The revision, if supplied, will be a combination of CRD and k8s-backed revisions.  Extract
	// the CRD rev and use that for the delete.
	crdRev, _, err := c.SplitNetworkPolicyRevision(revision)
	if err != nil {
		return nil, err
	}
	kvp, err := c.crdClient.Delete(ctx, key, crdRev, uid)

	if kvp != nil {
		// Convert the revision back to the combined CRD/k8s revision - the k8s rev will be empty.
		kvp.Revision = c.JoinNetworkPolicyRevisions(kvp.Revision, "")
	}
	return kvp, err
}

func (c *networkPolicyClient) Get(ctx context.Context, key model.Key, revision string) (*model.KVPair, error) {
	log.Debug("Received Get request on NetworkPolicy type")
	k := key.(model.ResourceKey)
	if k.Name == "" {
		return nil, errors.New("Missing policy name")
	}
	if k.Namespace == "" {
		return nil, errors.New("Missing policy namespace")
	}

	// The revision, if supplied, will be a combination of CRD and k8s-backed revisions.  Extract
	// the k8s rev and use the correct version depending on whether we are querying the CRD or the
	// k8s NetworkPolicy.
	crdRev, k8sRev, err := c.SplitNetworkPolicyRevision(revision)
	if err != nil {
		return nil, err
	}

	// Check to see if this is backed by a NetworkPolicy.
	if strings.HasPrefix(k.Name, conversion.K8sNetworkPolicyNamePrefix) {
		// Backed by a NetworkPolicy - extract the name.
		policyName := strings.TrimPrefix(k.Name, conversion.K8sNetworkPolicyNamePrefix)

		// Get the NetworkPolicy from the API and convert it.
		networkPolicy := networkingv1.NetworkPolicy{}
		err = c.clientSet.NetworkingV1().RESTClient().
			Get().
			Resource("networkpolicies").
			Namespace(k.Namespace).
			Name(policyName).
			VersionedParams(&metav1.GetOptions{ResourceVersion: k8sRev}, scheme.ParameterCodec).
			Do().Into(&networkPolicy)
		if err != nil {
			return nil, K8sErrorToCalico(err, k)
		}
		kvp, err := c.K8sNetworkPolicyToCalico(&networkPolicy)

		if kvp != nil {
			// Convert the revision back to the combined CRD/k8s revision - the CRD rev will be empty.
			kvp.Revision = c.JoinNetworkPolicyRevisions("", kvp.Revision)
		}
		return kvp, err
	} else {
		kvp, err := c.crdClient.Get(ctx, k, crdRev)

		if kvp != nil {
			// Convert the revision back to the combined CRD/k8s revision - the k8s rev will be empty.
			kvp.Revision = c.JoinNetworkPolicyRevisions(kvp.Revision, "")
		}
		return kvp, err
	}
}

func (c *networkPolicyClient) List(ctx context.Context, list model.ListInterface, revision string) (*model.KVPairList, error) {
	log.Debug("Received List request on NetworkPolicy type")
	l := list.(model.ResourceListOptions)
	if l.Name != "" {
		// Exact lookup on a NetworkPolicy.
		kvp, err := c.Get(ctx, model.ResourceKey{Name: l.Name, Namespace: l.Namespace, Kind: l.Kind}, revision)
		if err != nil {
			// Return empty slice of KVPair if the object doesn't exist, return the error otherwise.
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); ok {
				return &model.KVPairList{
					KVPairs:  []*model.KVPair{},
					Revision: revision,
				}, nil
			} else {
				return nil, err
			}
		}

		return &model.KVPairList{
			KVPairs:  []*model.KVPair{kvp},
			Revision: revision,
		}, nil
	}

	// List all Namespaced Calico Network Policies.
	npKvps, err := c.crdClient.List(ctx, l, revision)
	if err != nil {
		log.WithError(err).Info("Unable to list Calico CRD-backed Network Policy resources")
		return nil, err
	}

	// Convert the revision to the combined CRD/k8s revision - the k8s rev will be empty.
	for _, kvp := range npKvps.KVPairs {
		kvp.Revision = c.JoinNetworkPolicyRevisions(kvp.Revision, "")
	}

	// List all of the k8s NetworkPolicy objects in all Namespaces.
	networkPolicies := networkingv1.NetworkPolicyList{}
	req := c.clientSet.NetworkingV1().RESTClient().
		Get().
		Resource("networkpolicies")
	if l.Namespace != "" {
		// Add the namespace if requested.
		req = req.Namespace(l.Namespace)
	}
	err = req.Do().Into(&networkPolicies)
	if err != nil {
		log.WithError(err).Info("Unable to list K8s Network Policy resources")
		return nil, K8sErrorToCalico(err, l)
	}

	// For each policy, turn it into a Policy and generate the list.
	for _, p := range networkPolicies.Items {
		kvp, err := c.K8sNetworkPolicyToCalico(&p)
		if err != nil {
			log.WithError(err).Info("Failed to convert K8s Network Policy")
			return nil, err
		}

		// Convert the revision to the combined CRD/k8s revision - the CRD rev will be empty.
		kvp.Revision = c.JoinNetworkPolicyRevisions("", kvp.Revision)
		npKvps.KVPairs = append(npKvps.KVPairs, kvp)
	}

	// Combine the two resource versions to a single resource version for the List
	// that can be decoded by the Watch.
	npKvps.Revision = c.JoinNetworkPolicyRevisions(npKvps.Revision, networkPolicies.ResourceVersion)

	log.WithField("KVPs", npKvps).Info("Returning NP KVPs")
	return npKvps, nil
}

func (c *networkPolicyClient) EnsureInitialized() error {
	return nil
}

func (c *networkPolicyClient) Watch(ctx context.Context, list model.ListInterface, revision string) (api.WatchInterface, error) {
	// Build watch options to pass to k8s.
	opts := metav1.ListOptions{Watch: true}
	rlo, ok := list.(model.ResourceListOptions)
	if !ok {
		return nil, fmt.Errorf("ListInterface is not a ResourceListOptions: %s", list)
	}

	// Setting to Watch all networkPolicies in all namespaces; overriden below
	watchK8s, watchCrd := true, true

	// Watch a specific networkPolicy
	if len(rlo.Name) != 0 {
		if len(rlo.Namespace) == 0 {
			return nil, errors.New("cannot watch a specific NetworkPolicy without a namespace")
		}
		// We've been asked to watch a specific networkpolicy.
		log.WithField("name", rlo.Name).Debug("Watching a single networkpolicy")
		// Backed by a NetworkPolicy - extract the name.
		policyName := rlo.Name
		if strings.HasPrefix(rlo.Name, conversion.K8sNetworkPolicyNamePrefix) {
			watchCrd = false
			policyName = strings.TrimPrefix(rlo.Name, conversion.K8sNetworkPolicyNamePrefix)
		} else {
			watchK8s = false
		}
		// write back in rlo for custom resource watch below
		rlo.Name = policyName
		opts.FieldSelector = fields.OneTermEqualSelector("metadata.name", policyName).String()
	}

	// If a revision is specified, see if it contains a "/" and if so split into separate
	// revisions for the CRD and for the K8s resource.
	crdNPRev, k8sNPRev, err := c.SplitNetworkPolicyRevision(revision)
	if err != nil {
		return nil, err
	}

	opts.ResourceVersion = k8sNPRev
	var k8sRawWatch kwatch.Interface = kwatch.NewFake()
	if watchK8s {
		log.Debugf("Watching networkPolicy (k8s) at revision %q", k8sNPRev)
		k8sRawWatch, err = c.clientSet.NetworkingV1().NetworkPolicies(rlo.Namespace).Watch(opts)
		if err != nil {
			return nil, K8sErrorToCalico(err, list)
		}
	}
	converter := func(r Resource) (*model.KVPair, error) {
		np, ok := r.(*networkingv1.NetworkPolicy)
		if !ok {
			return nil, errors.New("NetworkPolicy conversion with incorrect k8s resource type")
		}
		return c.K8sNetworkPolicyToCalico(np)
	}
	k8sWatch := newK8sWatcherConverter(ctx, "NetworkPolicy (namespaced)", converter, k8sRawWatch)

	var calicoWatch api.WatchInterface = api.NewFake()
	if watchCrd {
		log.Debugf("Watching networkPolicy (crd) at revision %q", crdNPRev)
		calicoWatch, err = c.crdClient.Watch(ctx, rlo, crdNPRev)
		if err != nil {
			k8sWatch.Stop()
			return nil, err
		}
	}

	return newNetworkPolicyWatcher(ctx, k8sNPRev, crdNPRev, k8sWatch, calicoWatch), nil
}

func newNetworkPolicyWatcher(ctx context.Context, k8sRev, crdRev string, k8sWatch, calicoWatch api.WatchInterface) api.WatchInterface {
	ctx, cancel := context.WithCancel(ctx)
	wc := &networkPolicyWatcher{
		k8sNPRev:   k8sRev,
		crdNPRev:   crdRev,
		k8sNPWatch: k8sWatch,
		crdNPWatch: calicoWatch,
		context:    ctx,
		cancel:     cancel,
		resultChan: make(chan api.WatchEvent, resultsBufSize),
	}
	go wc.processNPEvents()
	return wc
}

type networkPolicyWatcher struct {
	conversion.Converter
	converter  ConvertK8sResourceToKVPair
	k8sNPRev   string
	crdNPRev   string
	k8sNPWatch api.WatchInterface
	crdNPWatch api.WatchInterface
	context    context.Context
	cancel     context.CancelFunc
	resultChan chan api.WatchEvent
	terminated uint32
}

// Stop stops the watcher and releases associated resources.
// This calls through to the context cancel function.
func (npw *networkPolicyWatcher) Stop() {
	npw.cancel()
	npw.k8sNPWatch.Stop()
	npw.crdNPWatch.Stop()
}

// ResultChan returns a channel used to receive WatchEvents.
func (npw *networkPolicyWatcher) ResultChan() <-chan api.WatchEvent {
	return npw.resultChan
}

// HasTerminated returns true when the watcher has completed termination processing.
func (npw *networkPolicyWatcher) HasTerminated() bool {
	terminated := atomic.LoadUint32(&npw.terminated) != 0

	if npw.k8sNPWatch != nil {
		terminated = terminated && npw.k8sNPWatch.HasTerminated()
	}
	if npw.crdNPWatch != nil {
		terminated = terminated && npw.crdNPWatch.HasTerminated()
	}

	return terminated
}

// Loop to process the events stream from the underlying k8s Watcher and convert them to
// backend KVPs.
func (npw *networkPolicyWatcher) processNPEvents() {
	log.Debug("Watcher process started")
	defer func() {
		log.Debug("Watcher process terminated")
		npw.Stop()
		close(npw.resultChan)
		atomic.AddUint32(&npw.terminated, 1)
	}()

	for {
		var ok bool
		var e api.WatchEvent
		var isCRDEvent bool
		select {
		case e, ok = <-npw.crdNPWatch.ResultChan():
			if !ok {
				// We shouldn't get a closed channel without first getting a terminating error,
				// so write a warning log and convert to a termination error.
				log.Warn("Calico NP channel closed")
				e = api.WatchEvent{
					Type: api.WatchError,
					Error: cerrors.ErrorWatchTerminated{
						ClosedByRemote: true,
						Err:            errors.New("Calico NP watch channel closed"),
					},
				}
			}
			log.Debug("Processing Calico NP event")
			isCRDEvent = true

		case e, ok = <-npw.k8sNPWatch.ResultChan():
			if !ok {
				// We shouldn't get a closed channel without first getting a terminating error,
				// so write a warning log and convert to a termination error.
				log.Warn("Kubernetes NP channel closed")
				e = api.WatchEvent{
					Type: api.WatchError,
					Error: cerrors.ErrorWatchTerminated{
						ClosedByRemote: true,
						Err:            errors.New("Kubernetes NP watch channel closed"),
					},
				}
			}
			log.Debug("Processing Kubernetes NP event")
			isCRDEvent = false

		case <-npw.context.Done(): // user cancel
			log.Debug("Process watcher done event in KDD client")
			return
		}

		// Update the resource version of the Object in the watcher.  The version returned on a watch
		// event needs to able to be passed back into a Watch client so that we can resume watching
		// when a watch fails.  The watch client is expecting a slash separated list of resource
		// versions in the format <CRD NP Revision>/<k8s NP Revision>.
		var value interface{}
		switch e.Type {
		case api.WatchModified, api.WatchAdded:
			value = e.New.Value
		case api.WatchDeleted:
			value = e.Old.Value
		}

		if value != nil {
			oma, ok := value.(metav1.ObjectMetaAccessor)
			if !ok {
				log.WithField("event", e).Error(
					"Resource returned from watch does not implement the ObjectMetaAccessor interface")
				e = api.WatchEvent{
					Type: api.WatchError,
					Error: cerrors.ErrorWatchTerminated{
						Err: errors.New("Resource returned from watch does not implement the ObjectMetaAccessor interface"),
					},
				}
			}
			if isCRDEvent {
				npw.crdNPRev = oma.GetObjectMeta().GetResourceVersion()
			} else {
				npw.k8sNPRev = oma.GetObjectMeta().GetResourceVersion()
			}
			oma.GetObjectMeta().SetResourceVersion(npw.JoinNetworkPolicyRevisions(npw.crdNPRev, npw.k8sNPRev))
		} else if e.Error == nil {
			log.WithField("event", e).Warning("Event had nil error and value")
		}

		// Send the processed event.
		select {
		case npw.resultChan <- e:
			// If this is an error event, check to see if it's a terminating one.
			// If so, terminate this watcher.
			if e.Type == api.WatchError {
				log.WithError(e.Error).Debug("Kubernetes event converted to backend watcher error event")
				if _, ok := e.Error.(cerrors.ErrorWatchTerminated); ok {
					log.Debug("Watch terminated event")
					return
				}
			}

		case <-npw.context.Done():
			log.Debug("Process watcher done event during watch event in kdd client")
			return
		}
	}
}
