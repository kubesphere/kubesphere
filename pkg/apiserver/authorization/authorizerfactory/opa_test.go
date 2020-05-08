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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	iamvealpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	factory "kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"testing"
)

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
				APIVersion:        "v1",
				Resource:          "nodes",
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
				APIVersion:        "v1",
				Resource:          "nodes",
				KubernetesRequest: true,
				ResourceRequest:   true,
				Path:              "/api/v1/nodes",
			},
			expectedDecision: authorizer.DecisionNoOpinion,
		}, {
			name: "tom can list nodes in cluster1",
			request: authorizer.AttributesRecord{
				User: &user.DefaultInfo{
					Name: "tom",
				},
				Verb:              "list",
				Cluster:           "cluster1",
				APIVersion:        "v1",
				Resource:          "nodes",
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
				APIVersion:        "v1",
				Resource:          "nodes",
				KubernetesRequest: true,
				ResourceRequest:   true,
				Path:              "/api/v1/clusters/cluster2/nodes",
			},
			expectedDecision: authorizer.DecisionNoOpinion,
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

func prepare() (am.AccessManagementInterface, error) {
	globalRoles := []*iamvealpha2.GlobalRole{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.ResourceKindGlobalRole,
				APIVersion: iamvealpha2.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "global-admin",
				Annotations: map[string]string{iamvealpha2.RegoOverrideAnnotation: "package authz\ndefault allow = true"},
			},
		}, {
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.ResourceKindGlobalRole,
				APIVersion: iamvealpha2.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "anonymous",
				Annotations: map[string]string{iamvealpha2.RegoOverrideAnnotation: "package authz\ndefault allow = false"},
			},
		}, {
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.ResourceKindGlobalRole,
				APIVersion: iamvealpha2.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster1-admin",
				Annotations: map[string]string{iamvealpha2.RegoOverrideAnnotation: `package authz
default allow = false
allow {
  resources_in_cluster1
}
resources_in_cluster1 {
	input.Cluster == "cluster1"
}`},
			},
		},
	}

	roleBindings := []*iamvealpha2.GlobalRoleBinding{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.ResourceKindGlobalRoleBinding,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "global-admin",
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: iamvealpha2.SchemeGroupVersion.String(),
				Kind:     iamvealpha2.ResourceKindGlobalRole,
				Name:     "global-admin",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:     iamvealpha2.ResourceKindUser,
					APIGroup: iamvealpha2.SchemeGroupVersion.String(),
					Name:     "admin",
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.ResourceKindGlobalRoleBinding,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "anonymous",
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: iamvealpha2.SchemeGroupVersion.String(),
				Kind:     iamvealpha2.ResourceKindGlobalRole,
				Name:     "anonymous",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:     iamvealpha2.ResourceKindUser,
					APIGroup: iamvealpha2.SchemeGroupVersion.String(),
					Name:     user.Anonymous,
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       iamvealpha2.ResourceKindGlobalRoleBinding,
				APIVersion: iamvealpha2.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster1-admin",
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: iamvealpha2.SchemeGroupVersion.String(),
				Kind:     iamvealpha2.ResourceKindGlobalRole,
				Name:     "cluster1-admin",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:     iamvealpha2.ResourceKindUser,
					APIGroup: iamvealpha2.SchemeGroupVersion.String(),
					Name:     "tom",
				},
			},
		},
	}

	ksClient := fake.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	factory := factory.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)
	for _, role := range globalRoles {
		err := factory.KubeSphereSharedInformerFactory().Iam().V1alpha2().GlobalRoles().Informer().GetIndexer().Add(role)
		if err != nil {
			return nil, fmt.Errorf("add role:%s", err)
		}
	}

	for _, roleBinding := range roleBindings {
		err := factory.KubeSphereSharedInformerFactory().Iam().V1alpha2().GlobalRoleBindings().Informer().GetIndexer().Add(roleBinding)
		if err != nil {
			return nil, fmt.Errorf("add role binding:%s", err)
		}
	}

	operator := am.NewAMOperator(factory)

	return operator, nil
}
