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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/kube/pkg/quota/v1"
	"kubesphere.io/kubesphere/kube/pkg/quota/v1/generic"
)

// legacyObjectCountAliases are what we used to do simple object counting quota with mapped to alias
var legacyObjectCountAliases = map[schema.GroupVersionResource]corev1.ResourceName{
	corev1.SchemeGroupVersion.WithResource(string(corev1.ResourceConfigMaps)):             corev1.ResourceConfigMaps,
	corev1.SchemeGroupVersion.WithResource(string(corev1.ResourceQuotas)):                 corev1.ResourceQuotas,
	corev1.SchemeGroupVersion.WithResource(string(corev1.ResourceReplicationControllers)): corev1.ResourceReplicationControllers,
	corev1.SchemeGroupVersion.WithResource(string(corev1.ResourceSecrets)):                corev1.ResourceSecrets,
}

// NewEvaluators returns the list of static evaluators that manage more than counts
func NewEvaluators(client client.Client) []quota.Evaluator {
	// these evaluators have special logic
	result := []quota.Evaluator{
		NewPodEvaluator(client, clock.RealClock{}),
		NewServiceEvaluator(client),
		NewPersistentVolumeClaimEvaluator(client),
	}
	// these evaluators require an alias for backwards compatibility
	for gvk, alias := range legacyObjectCountAliases {
		result = append(result,
			generic.NewObjectCountEvaluator(gvk.GroupVersion().WithResource(string(alias)).GroupResource(), generic.ListResourceUsingCacheFunc(client, gvk), alias))
	}
	return result
}
