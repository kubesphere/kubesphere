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
	"context"
	"fmt"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/admission"
	utilfeature "k8s.io/apiserver/pkg/util/feature"

	"kubesphere.io/kubesphere/kube/pkg/apis/core/v1/helper"
	k8sfeatures "kubesphere.io/kubesphere/kube/pkg/features"
	"kubesphere.io/kubesphere/kube/pkg/quota/v1"
	"kubesphere.io/kubesphere/kube/pkg/quota/v1/generic"
)

// the name used for object count quota
var pvcObjectCountName = generic.ObjectCountQuotaResourceNameFor(corev1.SchemeGroupVersion.WithResource("persistentvolumeclaims").GroupResource())

// pvcResources are the set of static resources managed by quota associated with pvcs.
// for each resource in this list, it may be refined dynamically based on storage class.
var pvcResources = []corev1.ResourceName{
	corev1.ResourcePersistentVolumeClaims,
	corev1.ResourceRequestsStorage,
}

// storageClassSuffix is the suffix to the qualified portion of storage class resource name.
// For example, if you want to quota storage by storage class, you would have a declaration
// that follows <storage-class>.storageclass.storage.k8s.io/<resource>.
// For example:
// * gold.storageclass.storage.k8s.io/: 500Gi
// * bronze.storageclass.storage.k8s.io/requests.storage: 500Gi
const storageClassSuffix string = ".storageclass.storage.k8s.io/"

// NewPersistentVolumeClaimEvaluator returns an evaluator that can evaluate persistent volume claims
func NewPersistentVolumeClaimEvaluator(cache client.Reader) quota.Evaluator {
	pvcEvaluator := &pvcEvaluator{cache: cache}
	return pvcEvaluator
}

// pvcEvaluator knows how to evaluate quota usage for persistent volume claims
type pvcEvaluator struct {
	// listFuncByNamespace knows how to list pvc claims
	cache client.Reader
}

// Constraints verifies that all required resources are present on the item.
func (p *pvcEvaluator) Constraints(required []corev1.ResourceName, item runtime.Object) error {
	// no-op for persistent volume claims
	return nil
}

// GroupResource that this evaluator tracks
func (p *pvcEvaluator) GroupResource() schema.GroupResource {
	return corev1.SchemeGroupVersion.WithResource("persistentvolumeclaims").GroupResource()
}

// Handles returns true if the evaluator should handle the specified operation.
func (p *pvcEvaluator) Handles(a admission.Attributes) bool {
	op := a.GetOperation()
	if op == admission.Create {
		return true
	}
	if op == admission.Update && utilfeature.DefaultFeatureGate.Enabled(k8sfeatures.ExpandPersistentVolumes) {
		return true
	}
	return false
}

// Matches returns true if the evaluator matches the specified quota with the provided input item
func (p *pvcEvaluator) Matches(resourceQuota *corev1.ResourceQuota, item runtime.Object) (bool, error) {
	return generic.Matches(resourceQuota, item, p.MatchingResources, generic.MatchesNoScopeFunc)
}

// MatchingScopes takes the input specified list of scopes and input object. Returns the set of scopes resource matches.
func (p *pvcEvaluator) MatchingScopes(item runtime.Object, scopes []corev1.ScopedResourceSelectorRequirement) ([]corev1.ScopedResourceSelectorRequirement, error) {
	return []corev1.ScopedResourceSelectorRequirement{}, nil
}

// UncoveredQuotaScopes takes the input matched scopes which are limited by configuration and the matched quota scopes.
// It returns the scopes which are in limited scopes but dont have a corresponding covering quota scope
func (p *pvcEvaluator) UncoveredQuotaScopes(limitedScopes []corev1.ScopedResourceSelectorRequirement, matchedQuotaScopes []corev1.ScopedResourceSelectorRequirement) ([]corev1.ScopedResourceSelectorRequirement, error) {
	return []corev1.ScopedResourceSelectorRequirement{}, nil
}

// MatchingResources takes the input specified list of resources and returns the set of resources it matches.
func (p *pvcEvaluator) MatchingResources(items []corev1.ResourceName) []corev1.ResourceName {
	var result []corev1.ResourceName
	for _, item := range items {
		// match object count quota fields
		if quota.Contains([]corev1.ResourceName{pvcObjectCountName}, item) {
			result = append(result, item)
			continue
		}
		// match pvc resources
		if quota.Contains(pvcResources, item) {
			result = append(result, item)
			continue
		}
		// match pvc resources scoped by storage class (<storage-class-name>.storage-class.kubernetes.io/<resource>)
		for _, resource := range pvcResources {
			byStorageClass := storageClassSuffix + string(resource)
			if strings.HasSuffix(string(item), byStorageClass) {
				result = append(result, item)
				break
			}
		}
	}
	return result
}

// Usage knows how to measure usage associated with item.
func (p *pvcEvaluator) Usage(item runtime.Object) (corev1.ResourceList, error) {
	result := corev1.ResourceList{}
	pvc, err := toExternalPersistentVolumeClaimOrError(item)
	if err != nil {
		return result, err
	}

	// charge for claim
	result[corev1.ResourcePersistentVolumeClaims] = *(resource.NewQuantity(1, resource.DecimalSI))
	result[pvcObjectCountName] = *(resource.NewQuantity(1, resource.DecimalSI))
	storageClassRef := helper.GetPersistentVolumeClaimClass(pvc)
	if len(storageClassRef) > 0 {
		storageClassClaim := corev1.ResourceName(storageClassRef + storageClassSuffix + string(corev1.ResourcePersistentVolumeClaims))
		result[storageClassClaim] = *(resource.NewQuantity(1, resource.DecimalSI))
	}

	// charge for storage
	if request, found := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; found {
		result[corev1.ResourceRequestsStorage] = request
		// charge usage to the storage class (if present)
		if len(storageClassRef) > 0 {
			storageClassStorage := corev1.ResourceName(storageClassRef + storageClassSuffix + string(corev1.ResourceRequestsStorage))
			result[storageClassStorage] = request
		}
	}
	return result, nil
}

func (p *pvcEvaluator) listPVC(namespace string) ([]runtime.Object, error) {
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := p.cache.List(context.Background(), pvcList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	pvcs := make([]runtime.Object, 0)
	for _, pvc := range pvcList.Items {
		pvcs = append(pvcs, &pvc)
	}
	return pvcs, nil
}

// UsageStats calculates aggregate usage for the object.
func (p *pvcEvaluator) UsageStats(options quota.UsageStatsOptions) (quota.UsageStats, error) {
	return generic.CalculateUsageStats(options, p.listPVC, generic.MatchesNoScopeFunc, p.Usage)
}

// ensure we implement required interface
var _ quota.Evaluator = &pvcEvaluator{}

func toExternalPersistentVolumeClaimOrError(obj runtime.Object) (*corev1.PersistentVolumeClaim, error) {
	var pvc *corev1.PersistentVolumeClaim
	switch t := obj.(type) {
	case *corev1.PersistentVolumeClaim:
		pvc = t
	default:
		return nil, fmt.Errorf("expect *v1.PersistentVolumeClaim, got %v", t)
	}
	return pvc, nil
}
