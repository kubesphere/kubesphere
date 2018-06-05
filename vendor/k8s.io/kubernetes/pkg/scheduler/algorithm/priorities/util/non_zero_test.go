/*
Copyright 2017 The Kubernetes Authors.

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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGetNonzeroRequests(t *testing.T) {
	tds := []struct {
		name           string
		requests       v1.ResourceList
		expectedCPU    int64
		expectedMemory int64
	}{
		{
			"cpu_and_memory_not_found",
			v1.ResourceList{},
			DefaultMilliCPURequest,
			DefaultMemoryRequest,
		},
		{
			"only_cpu_exist",
			v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("200m"),
			},
			200,
			DefaultMemoryRequest,
		},
		{
			"only_memory_exist",
			v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("400Mi"),
			},
			DefaultMilliCPURequest,
			400 * 1024 * 1024,
		},
		{
			"cpu_memory_exist",
			v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("200m"),
				v1.ResourceMemory: resource.MustParse("400Mi"),
			},
			200,
			400 * 1024 * 1024,
		},
	}

	for _, td := range tds {
		realCPU, realMemory := GetNonzeroRequests(&td.requests)
		assert.EqualValuesf(t, td.expectedCPU, realCPU, "Failed to test: %s", td.name)
		assert.EqualValuesf(t, td.expectedMemory, realMemory, "Failed to test: %s", td.name)
	}
}
