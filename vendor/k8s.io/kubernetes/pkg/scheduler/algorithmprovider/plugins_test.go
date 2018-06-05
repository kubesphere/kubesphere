/*
Copyright 2014 The Kubernetes Authors.

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

package algorithmprovider

import (
	"testing"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/scheduler/factory"
)

var (
	algorithmProviderNames = []string{
		factory.DefaultProvider,
	}
)

func TestDefaultConfigExists(t *testing.T) {
	p, err := factory.GetAlgorithmProvider(factory.DefaultProvider)
	if err != nil {
		t.Errorf("error retrieving default provider: %v", err)
	}
	if p == nil {
		t.Error("algorithm provider config should not be nil")
	}
	if len(p.FitPredicateKeys) == 0 {
		t.Error("default algorithm provider shouldn't have 0 fit predicates")
	}
}

func TestAlgorithmProviders(t *testing.T) {
	for _, pn := range algorithmProviderNames {
		p, err := factory.GetAlgorithmProvider(pn)
		if err != nil {
			t.Errorf("error retrieving '%s' provider: %v", pn, err)
			break
		}
		if len(p.PriorityFunctionKeys) == 0 {
			t.Errorf("%s algorithm provider shouldn't have 0 priority functions", pn)
		}
		for _, pf := range p.PriorityFunctionKeys.List() {
			if !factory.IsPriorityFunctionRegistered(pf) {
				t.Errorf("priority function %s is not registered but is used in the %s algorithm provider", pf, pn)
			}
		}
		for _, fp := range p.FitPredicateKeys.List() {
			if !factory.IsFitPredicateRegistered(fp) {
				t.Errorf("fit predicate %s is not registered but is used in the %s algorithm provider", fp, pn)
			}
		}
	}
}

func TestApplyFeatureGates(t *testing.T) {
	for _, pn := range algorithmProviderNames {
		p, err := factory.GetAlgorithmProvider(pn)
		if err != nil {
			t.Errorf("Error retrieving '%s' provider: %v", pn, err)
			break
		}

		if !p.FitPredicateKeys.Has("CheckNodeCondition") {
			t.Errorf("Failed to find predicate: 'CheckNodeCondition'")
			break
		}

		if !p.FitPredicateKeys.Has("PodToleratesNodeTaints") {
			t.Errorf("Failed to find predicate: 'PodToleratesNodeTaints'")
			break
		}
	}

	// Apply features for algorithm providers.
	utilfeature.DefaultFeatureGate.Set("TaintNodesByCondition=True")

	ApplyFeatureGates()

	for _, pn := range algorithmProviderNames {
		p, err := factory.GetAlgorithmProvider(pn)
		if err != nil {
			t.Errorf("Error retrieving '%s' provider: %v", pn, err)
			break
		}

		if !p.FitPredicateKeys.Has("PodToleratesNodeTaints") {
			t.Errorf("Failed to find predicate: 'PodToleratesNodeTaints'")
			break
		}

		if p.FitPredicateKeys.Has("CheckNodeCondition") {
			t.Errorf("Unexpected predicate: 'CheckNodeCondition'")
			break
		}
	}
}
