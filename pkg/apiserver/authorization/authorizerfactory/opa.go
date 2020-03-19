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
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	am2 "kubesphere.io/kubesphere/pkg/models/iam/am"
)

type opaAuthorizer struct {
	am am2.AccessManagementInterface
}

func (o *opaAuthorizer) Authorize(a authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {

	platformRole, err := o.am.GetPlatformRole(a.GetUser().GetName())
	if err != nil {
		return authorizer.DecisionDeny, "", err
	}

	// check platform role policy rules
	if a, r, e := o.roleAuthorize(platformRole, a); a == authorizer.DecisionAllow {
		return a, r, e
	}

	// it's not in cluster resource, permission denied
	// TODO declare implicit cluster info in request Info
	if a.GetCluster() == "" {
		return authorizer.DecisionDeny, "permission undefined", nil
	}

	clusterRole, err := o.am.GetClusterRole(a.GetCluster(), a.GetUser().GetName())
	if err != nil {
		return authorizer.DecisionDeny, "", err
	}

	// check cluster role policy rules
	if a, r, e := o.roleAuthorize(clusterRole, a); a == authorizer.DecisionAllow {
		return a, r, e
	}

	// it's not in cluster resource, permission denied
	if a.GetWorkspace() == "" && a.GetNamespace() == "" && a.GetDevopsProject() == "" {
		return authorizer.DecisionDeny, "permission undefined", nil
	}

	workspaceRole, err := o.am.GetWorkspaceRole(a.GetWorkspace(), a.GetUser().GetName())
	if err != nil {
		return authorizer.DecisionDeny, "", err
	}

	// check workspace role policy rules
	if a, r, e := o.roleAuthorize(workspaceRole, a); a == authorizer.DecisionAllow {
		return a, r, e
	}

	// it's not in workspace resource, permission denied
	if a.GetNamespace() == "" && a.GetDevopsProject() == "" {
		return authorizer.DecisionDeny, "permission undefined", nil
	}

	if a.GetNamespace() != "" {
		namespaceRole, err := o.am.GetNamespaceRole(a.GetNamespace(), a.GetUser().GetName())
		if err != nil {
			return authorizer.DecisionDeny, "", err
		}
		// check namespace role policy rules
		if a, r, e := o.roleAuthorize(namespaceRole, a); a == authorizer.DecisionAllow {
			return a, r, e
		}
	}

	if a.GetDevopsProject() != "" {
		devOpsRole, err := o.am.GetDevOpsRole(a.GetNamespace(), a.GetUser().GetName())
		if err != nil {
			return authorizer.DecisionDeny, "", err
		}
		// check devops role policy rules
		if a, r, e := o.roleAuthorize(devOpsRole, a); a == authorizer.DecisionAllow {
			return a, r, e
		}
	}

	return authorizer.DecisionDeny, "", nil
}

func (o *opaAuthorizer) roleAuthorize(role am2.Role, a authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {

	query, err := rego.New(rego.Query("data.authz.allow"), rego.Module("authz.rego", role.GetRego())).PrepareForEval(context.Background())

	if err != nil {
		return authorizer.DecisionDeny, "", err
	}

	results, err := query.Eval(context.Background(), rego.EvalInput(a))

	if err != nil {
		return authorizer.DecisionDeny, "", err
	}

	if len(results) > 0 && results[0].Expressions[0].Value == true {
		return authorizer.DecisionAllow, "", nil
	}

	return authorizer.DecisionDeny, "permission undefined", nil
}

func NewOPAAuthorizer(am am2.AccessManagementInterface) *opaAuthorizer {
	return &opaAuthorizer{am: am}
}
