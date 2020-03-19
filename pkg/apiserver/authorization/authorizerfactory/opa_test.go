/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package authorizerfactory

import (
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"testing"
)

func TestPlatformRole(t *testing.T) {

	opa := NewOPAAuthorizer(am.NewFakeAMOperator(cache.NewSimpleCache()))

	tests := []struct {
		name             string
		request          authorizer.AttributesRecord
		expectedDecision authorizer.Decision
	}{
		{
			name: "list nodes",
			request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name:   "admin",
					UID:    "0",
					Groups: []string{"admin"},
					Extra:  nil,
				},
				Verb:              "list",
				Cluster:           "",
				Workspace:         "",
				Namespace:         "",
				APIGroup:          "",
				APIVersion:        "v1",
				Resource:          "nodes",
				Subresource:       "",
				Name:              "",
				KubernetesRequest: true,
				ResourceRequest:   true,
				Path:              "/api/v1/nodes",
			},
			expectedDecision: authorizer.DecisionAllow,
		},
		{
			name: "list nodes",
			request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name:   user.Anonymous,
					UID:    "0",
					Groups: []string{"admin"},
					Extra:  nil,
				},
				Verb:              "list",
				Cluster:           "",
				Workspace:         "",
				Namespace:         "",
				APIGroup:          "",
				APIVersion:        "v1",
				Resource:          "nodes",
				Subresource:       "",
				Name:              "",
				KubernetesRequest: true,
				ResourceRequest:   true,
				Path:              "/api/v1/nodes",
			},
			expectedDecision: authorizer.DecisionDeny,
		},
	}

	for _, test := range tests {
		decision, _, err := opa.Authorize(test.request)
		if err != nil {
			t.Error(err)
		}
		if decision != test.expectedDecision {
			t.Errorf("%s: expected decision %v, actual %+v", test.name, test.expectedDecision, decision)
		}
	}
}
