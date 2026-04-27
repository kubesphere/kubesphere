/*
Copyright 2024 the KubeSphere Authors.

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
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mockReader struct {
	client.Reader
}

func TestPodEvaluatorMatchingResources(t *testing.T) {
	evaluator := NewPodEvaluator(&mockReader{}, clock.RealClock{})

	input := []corev1.ResourceName{
		corev1.ResourceCPU,
		corev1.ResourceMemory,
		corev1.ResourceName("nvidia.com/gpu"),
		corev1.ResourceName("requests.nvidia.com/gpu"),
	}

	// We expect nvidia.com/gpu to be matched after our fix
	matches := evaluator.MatchingResources(input)
	
	foundGPU := false
	for _, m := range matches {
		if m == corev1.ResourceName("nvidia.com/gpu") {
			foundGPU = true
		}
	}
	
	if !foundGPU {
		t.Errorf("Expected nvidia.com/gpu to be matched, but got %v", matches)
	}
}

func TestExtendedResourceQuotaSupport(t *testing.T) {
	evaluator := NewPodEvaluator(&mockReader{}, clock.RealClock{})

	// Case 1: MatchingResources should match non-prefixed extended resources
	input := []corev1.ResourceName{
		corev1.ResourceName("nvidia.com/gpu"),
	}
	
	matches := evaluator.MatchingResources(input)
	
	foundGPU := false
	for _, m := range matches {
		if m == corev1.ResourceName("nvidia.com/gpu") {
			foundGPU = true
		}
	}
	
	if !foundGPU {
		t.Errorf("MatchingResources failed to match nvidia.com/gpu. Matches: %v", matches)
	}

	// Case 2: Usage should include non-prefixed extended resources
	requests := corev1.ResourceList{
		corev1.ResourceName("nvidia.com/gpu"): resource.MustParse("1"),
	}
	limits := corev1.ResourceList{}
	
	usage := podComputeUsageHelper(requests, limits)
	
	if _, ok := usage[corev1.ResourceName("nvidia.com/gpu")]; !ok {
		t.Errorf("podComputeUsageHelper failed to include nvidia.com/gpu in usage. Usage: %v", usage)
	}
}

func TestPodComputeUsageHelper(t *testing.T) {
	requests := corev1.ResourceList{
		corev1.ResourceName("nvidia.com/gpu"): resource.MustParse("1"),
	}
	limits := corev1.ResourceList{}
	
	usage := podComputeUsageHelper(requests, limits)
	
	// We expect usage to contain BOTH the prefixed and non-prefixed versions
	if _, ok := usage[corev1.ResourceName("nvidia.com/gpu")]; !ok {
		t.Errorf("Expected usage to contain nvidia.com/gpu, got %v", usage)
	}
}
