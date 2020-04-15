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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	iamvealpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"testing"
)

func prepare() (am.AccessManagementInterface, error) {
	rules := []*iamvealpha2.PolicyRule{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.PolicyRuleKind,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "always-allow",
			},
			Rego: "package authz\ndefault allow = true",
		}, {
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.PolicyRuleKind,
				APIVersion: iamvealpha2.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "always-deny",
			},
			Rego: "package authz\ndefault allow = false",
		}, {
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.PolicyRuleKind,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "manage-cluster1-resources",
			},
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

	roles := []*iamvealpha2.Role{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.RoleKind,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "global-admin",
			},
			Target: iamvealpha2.Target{
				Scope: iamvealpha2.GlobalScope,
				Name:  "",
			},
			Rules: []iamvealpha2.RuleRef{
				{
					APIGroup: iamvealpha2.SchemeGroupVersion.String(),
					Kind:     iamvealpha2.PolicyRuleKind,
					Name:     "always-allow",
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.RoleKind,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "anonymous",
			},
			Target: iamvealpha2.Target{
				Scope: iamvealpha2.GlobalScope,
				Name:  "",
			},
			Rules: []iamvealpha2.RuleRef{
				{
					APIGroup: iamvealpha2.SchemeGroupVersion.String(),
					Kind:     iamvealpha2.PolicyRuleKind,
					Name:     "always-deny",
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.RoleKind,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster1-admin",
			},
			Target: iamvealpha2.Target{
				Scope: iamvealpha2.GlobalScope,
				Name:  "",
			},
			Rules: []iamvealpha2.RuleRef{
				{
					APIGroup: iamvealpha2.SchemeGroupVersion.String(),
					Kind:     iamvealpha2.PolicyRuleKind,
					Name:     "manage-cluster1-resources",
				},
			},
		},
	}

	roleBindings := []*iamvealpha2.RoleBinding{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.RoleBindingKind,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "global-admin",
			},
			Scope: iamvealpha2.GlobalScope,
			RoleRef: iamvealpha2.RoleRef{
				APIGroup: iamvealpha2.SchemeGroupVersion.String(),
				Kind:     iamvealpha2.RoleKind,
				Name:     "global-admin",
			},
			Subjects: []iamvealpha2.Subject{
				{
					Kind:     iamvealpha2.UserKind,
					APIGroup: iamvealpha2.SchemeGroupVersion.String(),
					Name:     "admin",
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.RoleBindingKind,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "anonymous",
			},
			Scope: iamvealpha2.GlobalScope,
			RoleRef: iamvealpha2.RoleRef{
				APIGroup: iamvealpha2.SchemeGroupVersion.String(),
				Kind:     iamvealpha2.RoleKind,
				Name:     "anonymous",
			},
			Subjects: []iamvealpha2.Subject{
				{
					Kind:     iamvealpha2.UserKind,
					APIGroup: iamvealpha2.SchemeGroupVersion.String(),
					Name:     user.Anonymous,
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.RoleBindingKind,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster1-admin",
			},
			Scope: iamvealpha2.GlobalScope,
			RoleRef: iamvealpha2.RoleRef{
				APIGroup: iamvealpha2.SchemeGroupVersion.String(),
				Kind:     iamvealpha2.RoleKind,
				Name:     "cluster1-admin",
			},
			Subjects: []iamvealpha2.Subject{
				{
					Kind:     iamvealpha2.UserKind,
					APIGroup: iamvealpha2.SchemeGroupVersion.String(),
					Name:     "tom",
				},
			},
		},
	}

	ksClient := fake.NewSimpleClientset()
	informerFactory := externalversions.NewSharedInformerFactory(ksClient, 0)

	for _, rule := range rules {
		err := informerFactory.Iam().V1alpha2().PolicyRules().Informer().GetIndexer().Add(rule)
		if err != nil {
			return nil, fmt.Errorf("add rule:%s", err)
		}
	}
	for _, role := range roles {
		err := informerFactory.Iam().V1alpha2().Roles().Informer().GetIndexer().Add(role)
		if err != nil {
			return nil, fmt.Errorf("add role:%s", err)
		}
	}
	for _, roleBinding := range roleBindings {
		err := informerFactory.Iam().V1alpha2().RoleBindings().Informer().GetIndexer().Add(roleBinding)
		if err != nil {
			return nil, fmt.Errorf("add role binding:%s", err)
		}
	}

	operator := am.NewAMOperator(ksClient, informerFactory)

	return operator, nil
}

func TestGlobalRole(t *testing.T) {

	operator, err := prepare()
	if err != nil {
		t.Fatal(err)
	}

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
