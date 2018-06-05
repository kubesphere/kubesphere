/*
Copyright 2016 The Kubernetes Authors.

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

package core

import (
	"hash/fnv"
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/schedulercache"
	hashutil "k8s.io/kubernetes/pkg/util/hash"

	"github.com/golang/glog"
	"github.com/golang/groupcache/lru"
)

// We use predicate names as cache's key, its count is limited
const maxCacheEntries = 100

// EquivalenceCache holds:
// 1. a map of AlgorithmCache with node name as key
// 2. function to get equivalence pod
type EquivalenceCache struct {
	mu             sync.Mutex
	algorithmCache map[string]AlgorithmCache
}

// The AlgorithmCache stores PredicateMap with predicate name as key
type AlgorithmCache struct {
	// Only consider predicates for now
	predicatesCache *lru.Cache
}

// PredicateMap stores HostPrediacte with equivalence hash as key
type PredicateMap map[uint64]HostPredicate

// HostPredicate is the cached predicate result
type HostPredicate struct {
	Fit         bool
	FailReasons []algorithm.PredicateFailureReason
}

func newAlgorithmCache() AlgorithmCache {
	return AlgorithmCache{
		predicatesCache: lru.New(maxCacheEntries),
	}
}

// NewEquivalenceCache returns EquivalenceCache to speed up predicates by caching
// result from previous scheduling.
func NewEquivalenceCache() *EquivalenceCache {
	return &EquivalenceCache{
		algorithmCache: make(map[string]AlgorithmCache),
	}
}

// RunPredicate will return a cached predicate result. In case of a cache miss, the predicate will
// be run and its results cached for the next call.
//
// NOTE: RunPredicate will not update the equivalence cache if the given NodeInfo is stale.
func (ec *EquivalenceCache) RunPredicate(
	pred algorithm.FitPredicate,
	predicateKey string,
	pod *v1.Pod,
	meta algorithm.PredicateMetadata,
	nodeInfo *schedulercache.NodeInfo,
	equivClassInfo *equivalenceClassInfo,
	cache schedulercache.Cache,
) (bool, []algorithm.PredicateFailureReason, error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	fit, reasons, invalid := ec.lookupResult(pod.GetName(), nodeInfo.Node().GetName(), predicateKey, equivClassInfo.hash)
	if !invalid {
		return fit, reasons, nil
	}
	fit, reasons, err := pred(pod, meta, nodeInfo)
	if err != nil {
		return fit, reasons, err
	}
	// Skip update if NodeInfo is stale.
	if cache != nil && cache.IsUpToDate(nodeInfo) {
		ec.updateResult(pod.GetName(), nodeInfo.Node().GetName(), predicateKey, fit, reasons, equivClassInfo.hash)
	}
	return fit, reasons, nil
}

// updateResult updates the cached result of a predicate.
func (ec *EquivalenceCache) updateResult(
	podName, nodeName, predicateKey string,
	fit bool,
	reasons []algorithm.PredicateFailureReason,
	equivalenceHash uint64,
) {
	if _, exist := ec.algorithmCache[nodeName]; !exist {
		ec.algorithmCache[nodeName] = newAlgorithmCache()
	}
	predicateItem := HostPredicate{
		Fit:         fit,
		FailReasons: reasons,
	}
	// if cached predicate map already exists, just update the predicate by key
	if v, ok := ec.algorithmCache[nodeName].predicatesCache.Get(predicateKey); ok {
		predicateMap := v.(PredicateMap)
		// maps in golang are references, no need to add them back
		predicateMap[equivalenceHash] = predicateItem
	} else {
		ec.algorithmCache[nodeName].predicatesCache.Add(predicateKey,
			PredicateMap{
				equivalenceHash: predicateItem,
			})
	}
	glog.V(5).Infof("Updated cached predicate: %v for pod: %v on node: %s, with item %v", predicateKey, podName, nodeName, predicateItem)
}

// lookupResult returns cached predicate results:
// 1. if pod fit
// 2. reasons if pod did not fit
// 3. if cache item is not found
func (ec *EquivalenceCache) lookupResult(
	podName, nodeName, predicateKey string,
	equivalenceHash uint64,
) (bool, []algorithm.PredicateFailureReason, bool) {
	glog.V(5).Infof("Begin to calculate predicate: %v for pod: %s on node: %s based on equivalence cache",
		predicateKey, podName, nodeName)
	if algorithmCache, exist := ec.algorithmCache[nodeName]; exist {
		if cachePredicate, exist := algorithmCache.predicatesCache.Get(predicateKey); exist {
			predicateMap := cachePredicate.(PredicateMap)
			// TODO(resouer) Is it possible a race that cache failed to update immediately?
			if hostPredicate, ok := predicateMap[equivalenceHash]; ok {
				if hostPredicate.Fit {
					return true, []algorithm.PredicateFailureReason{}, false
				}
				return false, hostPredicate.FailReasons, false
			}
			// is invalid
			return false, []algorithm.PredicateFailureReason{}, true
		}
	}
	return false, []algorithm.PredicateFailureReason{}, true
}

// InvalidateCachedPredicateItem marks all items of given predicateKeys, of all pods, on the given node as invalid
func (ec *EquivalenceCache) InvalidateCachedPredicateItem(nodeName string, predicateKeys sets.String) {
	if len(predicateKeys) == 0 {
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	if algorithmCache, exist := ec.algorithmCache[nodeName]; exist {
		for predicateKey := range predicateKeys {
			algorithmCache.predicatesCache.Remove(predicateKey)
		}
	}
	glog.V(5).Infof("Done invalidating cached predicates: %v on node: %s", predicateKeys, nodeName)
}

// InvalidateCachedPredicateItemOfAllNodes marks all items of given predicateKeys, of all pods, on all node as invalid
func (ec *EquivalenceCache) InvalidateCachedPredicateItemOfAllNodes(predicateKeys sets.String) {
	if len(predicateKeys) == 0 {
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	// algorithmCache uses nodeName as key, so we just iterate it and invalid given predicates
	for _, algorithmCache := range ec.algorithmCache {
		for predicateKey := range predicateKeys {
			// just use keys is enough
			algorithmCache.predicatesCache.Remove(predicateKey)
		}
	}
	glog.V(5).Infof("Done invalidating cached predicates: %v on all node", predicateKeys)
}

// InvalidateAllCachedPredicateItemOfNode marks all cached items on given node as invalid
func (ec *EquivalenceCache) InvalidateAllCachedPredicateItemOfNode(nodeName string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	delete(ec.algorithmCache, nodeName)
	glog.V(5).Infof("Done invalidating all cached predicates on node: %s", nodeName)
}

// InvalidateCachedPredicateItemForPodAdd is a wrapper of InvalidateCachedPredicateItem for pod add case
func (ec *EquivalenceCache) InvalidateCachedPredicateItemForPodAdd(pod *v1.Pod, nodeName string) {
	// MatchInterPodAffinity: we assume scheduler can make sure newly bound pod
	// will not break the existing inter pod affinity. So we does not need to invalidate
	// MatchInterPodAffinity when pod added.
	//
	// But when a pod is deleted, existing inter pod affinity may become invalid.
	// (e.g. this pod was preferred by some else, or vice versa)
	//
	// NOTE: assumptions above will not stand when we implemented features like
	// RequiredDuringSchedulingRequiredDuringExecution.

	// NoDiskConflict: the newly scheduled pod fits to existing pods on this node,
	// it will also fits to equivalence class of existing pods

	// GeneralPredicates: will always be affected by adding a new pod
	invalidPredicates := sets.NewString(predicates.GeneralPred)

	// MaxPDVolumeCountPredicate: we check the volumes of pod to make decision.
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil {
			invalidPredicates.Insert(predicates.MaxEBSVolumeCountPred, predicates.MaxGCEPDVolumeCountPred, predicates.MaxAzureDiskVolumeCountPred)
		} else {
			if vol.AWSElasticBlockStore != nil {
				invalidPredicates.Insert(predicates.MaxEBSVolumeCountPred)
			}
			if vol.GCEPersistentDisk != nil {
				invalidPredicates.Insert(predicates.MaxGCEPDVolumeCountPred)
			}
			if vol.AzureDisk != nil {
				invalidPredicates.Insert(predicates.MaxAzureDiskVolumeCountPred)
			}
		}
	}
	ec.InvalidateCachedPredicateItem(nodeName, invalidPredicates)
}

// equivalenceClassInfo holds equivalence hash which is used for checking equivalence cache.
// We will pass this to podFitsOnNode to ensure equivalence hash is only calculated per schedule.
type equivalenceClassInfo struct {
	// Equivalence hash.
	hash uint64
}

// getEquivalenceClassInfo returns a hash of the given pod.
// The hashing function returns the same value for any two pods that are
// equivalent from the perspective of scheduling.
func (ec *EquivalenceCache) getEquivalenceClassInfo(pod *v1.Pod) *equivalenceClassInfo {
	equivalencePod := getEquivalenceHash(pod)
	if equivalencePod != nil {
		hash := fnv.New32a()
		hashutil.DeepHashObject(hash, equivalencePod)
		return &equivalenceClassInfo{
			hash: uint64(hash.Sum32()),
		}
	}
	return nil
}

// equivalencePod is the set of pod attributes which must match for two pods to
// be considered equivalent for scheduling purposes. For correctness, this must
// include any Pod field which is used by a FitPredicate.
//
// NOTE: For equivalence hash to be formally correct, lists and maps in the
// equivalencePod should be normalized. (e.g. by sorting them) However, the
// vast majority of equivalent pod classes are expected to be created from a
// single pod template, so they will all have the same ordering.
type equivalencePod struct {
	Namespace      *string
	Labels         map[string]string
	Affinity       *v1.Affinity
	Containers     []v1.Container // See note about ordering
	InitContainers []v1.Container // See note about ordering
	NodeName       *string
	NodeSelector   map[string]string
	Tolerations    []v1.Toleration
	Volumes        []v1.Volume // See note about ordering
}

// getEquivalenceHash returns the equivalencePod for a Pod.
func getEquivalenceHash(pod *v1.Pod) *equivalencePod {
	ep := &equivalencePod{
		Namespace:      &pod.Namespace,
		Labels:         pod.Labels,
		Affinity:       pod.Spec.Affinity,
		Containers:     pod.Spec.Containers,
		InitContainers: pod.Spec.InitContainers,
		NodeName:       &pod.Spec.NodeName,
		NodeSelector:   pod.Spec.NodeSelector,
		Tolerations:    pod.Spec.Tolerations,
		Volumes:        pod.Spec.Volumes,
	}
	// DeepHashObject considers nil and empty slices to be different. Normalize them.
	if len(ep.Containers) == 0 {
		ep.Containers = nil
	}
	if len(ep.InitContainers) == 0 {
		ep.InitContainers = nil
	}
	if len(ep.Tolerations) == 0 {
		ep.Tolerations = nil
	}
	if len(ep.Volumes) == 0 {
		ep.Volumes = nil
	}
	// Normalize empty maps also.
	if len(ep.Labels) == 0 {
		ep.Labels = nil
	}
	if len(ep.NodeSelector) == 0 {
		ep.NodeSelector = nil
	}
	// TODO(misterikkit): Also normalize nested maps and slices.
	return ep
}
