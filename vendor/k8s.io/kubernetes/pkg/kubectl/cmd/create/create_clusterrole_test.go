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

package create

import (
	"testing"

	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
)

func TestCreateClusterRole(t *testing.T) {
	clusterRoleName := "my-cluster-role"

	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	tf.Namespace = "test"
	tf.Client = &fake.RESTClient{}
	tf.ClientConfigVal = defaultClientConfig()

	tests := map[string]struct {
		verbs               string
		resources           string
		nonResourceURL      string
		resourceNames       string
		expectedClusterRole *rbac.ClusterRole
	}{
		"test-duplicate-resources": {
			verbs:     "get,watch,list",
			resources: "pods,pods",
			expectedClusterRole: &rbac.ClusterRole{
				TypeMeta: v1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole"},
				ObjectMeta: v1.ObjectMeta{
					Name: clusterRoleName,
				},
				Rules: []rbac.PolicyRule{
					{
						Verbs:         []string{"get", "watch", "list"},
						Resources:     []string{"pods"},
						APIGroups:     []string{""},
						ResourceNames: []string{},
					},
				},
			},
		},
		"test-valid-case-with-multiple-apigroups": {
			verbs:     "get,watch,list",
			resources: "pods,deployments.extensions",
			expectedClusterRole: &rbac.ClusterRole{
				TypeMeta: v1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole"},
				ObjectMeta: v1.ObjectMeta{
					Name: clusterRoleName,
				},
				Rules: []rbac.PolicyRule{
					{
						Verbs:         []string{"get", "watch", "list"},
						Resources:     []string{"pods"},
						APIGroups:     []string{""},
						ResourceNames: []string{},
					},
					{
						Verbs:         []string{"get", "watch", "list"},
						Resources:     []string{"deployments"},
						APIGroups:     []string{"extensions"},
						ResourceNames: []string{},
					},
				},
			},
		},
		"test-non-resource-url": {
			verbs:          "get",
			nonResourceURL: "/logs/,/healthz",
			expectedClusterRole: &rbac.ClusterRole{
				TypeMeta: v1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole"},
				ObjectMeta: v1.ObjectMeta{
					Name: clusterRoleName,
				},
				Rules: []rbac.PolicyRule{
					{
						Verbs:           []string{"get"},
						NonResourceURLs: []string{"/logs/", "/healthz"},
					},
				},
			},
		},
		"test-resource-and-non-resource-url": {
			verbs:          "get",
			nonResourceURL: "/logs/,/healthz",
			resources:      "pods",
			expectedClusterRole: &rbac.ClusterRole{
				TypeMeta: v1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole"},
				ObjectMeta: v1.ObjectMeta{
					Name: clusterRoleName,
				},
				Rules: []rbac.PolicyRule{
					{
						Verbs:         []string{"get"},
						Resources:     []string{"pods"},
						APIGroups:     []string{""},
						ResourceNames: []string{},
					},
					{
						Verbs:           []string{"get"},
						NonResourceURLs: []string{"/logs/", "/healthz"},
					},
				},
			},
		},
	}

	for name, test := range tests {
		ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
		cmd := NewCmdCreateClusterRole(tf, ioStreams)
		cmd.Flags().Set("dry-run", "true")
		cmd.Flags().Set("output", "yaml")
		cmd.Flags().Set("verb", test.verbs)
		cmd.Flags().Set("resource", test.resources)
		cmd.Flags().Set("non-resource-url", test.nonResourceURL)
		if test.resourceNames != "" {
			cmd.Flags().Set("resource-name", test.resourceNames)
		}
		cmd.Run(cmd, []string{clusterRoleName})
		actual := &rbac.ClusterRole{}
		if err := runtime.DecodeInto(legacyscheme.Codecs.UniversalDecoder(), buf.Bytes(), actual); err != nil {
			t.Log(string(buf.Bytes()))
			t.Fatal(err)
		}
		if !equality.Semantic.DeepEqual(test.expectedClusterRole, actual) {
			t.Errorf("%s:\nexpected:\n%#v\nsaw:\n%#v", name, test.expectedClusterRole, actual)
		}
	}
}

func TestClusterRoleValidate(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	tf.Namespace = "test"

	tests := map[string]struct {
		clusterRoleOptions *CreateClusterRoleOptions
		expectErr          bool
	}{
		"test-missing-name": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{},
			},
			expectErr: true,
		},
		"test-missing-verb": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name: "my-clusterrole",
				},
			},
			expectErr: true,
		},
		"test-missing-resource": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"get"},
				},
			},
			expectErr: true,
		},
		"test-missing-resource-existing-apigroup": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"get"},
					Resources: []ResourceOptions{
						{
							Group: "extensions",
						},
					},
				},
			},
			expectErr: true,
		},
		"test-missing-resource-existing-subresource": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"get"},
					Resources: []ResourceOptions{
						{
							SubResource: "scale",
						},
					},
				},
			},
			expectErr: true,
		},
		"test-invalid-verb": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"invalid-verb"},
					Resources: []ResourceOptions{
						{
							Resource: "pods",
						},
					},
				},
			},
			expectErr: true,
		},
		"test-nonresource-verb": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"post"},
					Resources: []ResourceOptions{
						{
							Resource: "pods",
						},
					},
				},
			},
			expectErr: true,
		},
		"test-special-verb": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"use"},
					Resources: []ResourceOptions{
						{
							Resource: "pods",
						},
					},
				},
			},
			expectErr: true,
		},
		"test-mix-verbs": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"impersonate", "use"},
					Resources: []ResourceOptions{
						{
							Resource:    "userextras",
							SubResource: "scopes",
						},
					},
				},
			},
			expectErr: true,
		},
		"test-special-verb-with-wrong-apigroup": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"impersonate"},
					Resources: []ResourceOptions{
						{
							Resource:    "userextras",
							SubResource: "scopes",
							Group:       "extensions",
						},
					},
				},
			},
			expectErr: true,
		},
		"test-invalid-resource": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"get"},
					Resources: []ResourceOptions{
						{
							Resource: "invalid-resource",
						},
					},
				},
			},
			expectErr: true,
		},
		"test-resource-name-with-multiple-resources": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"get"},
					Resources: []ResourceOptions{
						{
							Resource: "pods",
						},
						{
							Resource: "deployments",
							Group:    "extensions",
						},
					},
				},
			},
			expectErr: false,
		},
		"test-valid-case": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "role-binder",
					Verbs: []string{"get", "list", "bind"},
					Resources: []ResourceOptions{
						{
							Resource: "roles",
							Group:    "rbac.authorization.k8s.io",
						},
					},
				},
			},
			expectErr: false,
		},
		"test-valid-case-with-subresource": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"get", "list"},
					Resources: []ResourceOptions{
						{
							Resource:    "replicasets",
							SubResource: "scale",
						},
					},
				},
			},
			expectErr: false,
		},
		"test-valid-case-with-additional-resource": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"impersonate"},
					Resources: []ResourceOptions{
						{
							Resource:    "userextras",
							SubResource: "scopes",
							Group:       "authentication.k8s.io",
						},
					},
				},
			},
			expectErr: false,
		},
		"test-invalid-empty-non-resource-url": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"create"},
				},
				NonResourceURLs: []string{""},
			},
			expectErr: true,
		},
		"test-invalid-non-resource-url": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"create"},
				},
				NonResourceURLs: []string{"logs"},
			},
			expectErr: true,
		},
		"test-invalid-non-resource-url-with-*": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"create"},
				},
				NonResourceURLs: []string{"/logs/*/"},
			},
			expectErr: true,
		},
		"test-invalid-non-resource-url-with-multiple-*": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"create"},
				},
				NonResourceURLs: []string{"/logs*/*"},
			},
			expectErr: true,
		},
		"test-invalid-verb-for-non-resource-url": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"create"},
				},
				NonResourceURLs: []string{"/logs/"},
			},
			expectErr: true,
		},
		"test-resource-and-non-resource-url-specified-together": {
			clusterRoleOptions: &CreateClusterRoleOptions{
				CreateRoleOptions: &CreateRoleOptions{
					Name:  "my-clusterrole",
					Verbs: []string{"get"},
					Resources: []ResourceOptions{
						{
							Resource:    "replicasets",
							SubResource: "scale",
						},
					},
				},
				NonResourceURLs: []string{"/logs/", "/logs/*"},
			},
			expectErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var err error
			test.clusterRoleOptions.Mapper, err = tf.RESTMapper()
			if err != nil {
				t.Fatal(err)
			}
			err = test.clusterRoleOptions.Validate()
			if test.expectErr && err == nil {
				t.Errorf("%s: expect error happens, but validate passes.", name)
			}
			if !test.expectErr && err != nil {
				t.Errorf("%s: unexpected error: %v", name, err)
			}
		})
	}
}

func defaultClientConfig() *restclient.Config {
	return &restclient.Config{
		APIPath: "/api",
		ContentConfig: restclient.ContentConfig{
			NegotiatedSerializer: scheme.Codecs,
			ContentType:          runtime.ContentTypeJSON,
			GroupVersion:         &schema.GroupVersion{Version: "v1"},
		},
	}
}
