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
	"context"
	"github.com/open-policy-agent/opa/rego"
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"testing"
)

func TestPlatformRole(t *testing.T) {

	module := `package platform.authz
default allow = false

allow {
    input.User.name == "admin"
}

allow {
    is_admin
}

is_admin {
    input.User.Groups[_] == "admin"
}
`
	query, err := rego.New(rego.Query("data.authz.allow"), rego.Module("authz.rego", module)).PrepareForEval(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	input := authorizer.AttributesRecord{
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
		DevopsProject:     "",
		APIGroup:          "",
		APIVersion:        "v1",
		Resource:          "nodes",
		Subresource:       "",
		Name:              "",
		KubernetesRequest: true,
		ResourceRequest:   true,
		Path:              "/api/v1/nodes",
	}

	results, err := query.Eval(context.Background(), rego.EvalInput(input))

	if err != nil {
		t.Log(err)
	}

	if len(results) > 0 && results[0].Expressions[0].Value == true {
		t.Log("allowed")
	} else {
		t.Log("deny")
	}
}
