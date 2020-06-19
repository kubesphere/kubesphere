/*
Copyright 2018 The Kubernetes Authors.

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

package proxy

import (
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"testing"
)

func TestNewAuthorizer(t *testing.T) {
	tests := []struct {
		multiClusterEnabled bool
		request             authorizer.AttributesRecord
		expectResult        authorizer.Decision
	}{
		{
			multiClusterEnabled: false,
			request: authorizer.AttributesRecord{
				Workspace:         "ws",
				Namespace:         "ns",
				KubernetesRequest: false,
				ResourceRequest:   false,
			},
			expectResult: authorizer.DecisionNoOpinion,
		},
		{
			multiClusterEnabled: false,
			request: authorizer.AttributesRecord{
				Cluster:           "cluster1",
				Workspace:         "ws",
				Namespace:         "ns",
				KubernetesRequest: false,
				ResourceRequest:   false,
			},
			expectResult: authorizer.DecisionNoOpinion,
		},
		{
			multiClusterEnabled: true,
			request: authorizer.AttributesRecord{
				Cluster:           "cluster1",
				Workspace:         "ws",
				Namespace:         "ns",
				KubernetesRequest: false,
				ResourceRequest:   false,
			},
			expectResult: authorizer.DecisionAllow,
		},
		{
			multiClusterEnabled: true,
			request: authorizer.AttributesRecord{
				Workspace:         "ws",
				Namespace:         "ns",
				KubernetesRequest: false,
				ResourceRequest:   false,
			},
			expectResult: authorizer.DecisionNoOpinion,
		},
	}
	for i, test := range tests {
		a := NewAuthorizer(test.multiClusterEnabled)
		result, _, _ := a.Authorize(test.request)
		if result != test.expectResult {
			t.Errorf("case %d, got %#v, expected %#v", i, result, test.expectResult)
		}
	}
}
