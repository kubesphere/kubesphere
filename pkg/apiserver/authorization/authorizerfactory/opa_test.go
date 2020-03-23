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
	"testing"
)

func TestPlatformRole(t *testing.T) {
	platformRoles := map[string]am.FakeRole{"admin": {
		Name: "admin",
		Rego: "package authz\ndefault allow = true",
	}, "anonymous": {
		Name: "anonymous",
		Rego: "package authz\ndefault allow = false",
	}, "tom": {
		Name: "tom",
		Rego: `package authz
default allow = false
allow {
  resources_in_cluster1
}
resources_in_cluster1 {
	input.Cluster == "cluster1"
}`,
	},
	}

	operator := am.NewFakeAMOperator()
	operator.Prepare(platformRoles, nil, nil, nil)

	opa := NewOPAAuthorizer(operator)

	tests := []struct {
		name             string
		request          authorizer.AttributesRecord
		expectedDecision authorizer.Decision
	}{
		{
			name: "admin can list nodes",
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
			name: "anonymous can not list nodes",
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
		}, {
			name: "tom can list nodes in cluster1",
			request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "tom",
				},
				Verb:              "list",
				Cluster:           "cluster1",
				Workspace:         "",
				Namespace:         "",
				APIGroup:          "",
				APIVersion:        "v1",
				Resource:          "nodes",
				Subresource:       "",
				Name:              "",
				KubernetesRequest: true,
				ResourceRequest:   true,
				Path:              "/api/v1/clusters/cluster1/nodes",
			},
			expectedDecision: authorizer.DecisionAllow,
		},
		{
			name: "tom can not list nodes in cluster2",
			request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "tom",
				},
				Verb:              "list",
				Cluster:           "cluster2",
				Workspace:         "",
				Namespace:         "",
				APIGroup:          "",
				APIVersion:        "v1",
				Resource:          "nodes",
				Subresource:       "",
				Name:              "",
				KubernetesRequest: true,
				ResourceRequest:   true,
				Path:              "/api/v1/clusters/cluster2/nodes",
			},
			expectedDecision: authorizer.DecisionDeny,
		},
	}

	for _, test := range tests {
		decision, _, err := opa.Authorize(test.request)
		if err != nil {
			t.Errorf("test failed: %s, %v", test.name, err)
		}
		if decision != test.expectedDecision {
			t.Errorf("%s: expected decision %v, actual %+v", test.name, test.expectedDecision, decision)
		}
	}
}
