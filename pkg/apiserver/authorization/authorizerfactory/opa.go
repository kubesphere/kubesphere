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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
)

type opaAuthorizer struct {
	am am.AccessManagementInterface
}

// Make decision by request attributes
func (o *opaAuthorizer) Authorize(attr authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {

	// Make decisions based on the authorization policy of different levels of roles
	globalRole, err := o.am.GetRoleOfUserInTargetScope(iamv1alpha2.GlobalScope, "", attr.GetUser().GetName())
	if err != nil {
		if errors.IsNotFound(err) {
			return authorizer.DecisionDeny, err.Error(), nil
		}
		return authorizer.DecisionDeny, "", err
	}

	// check platform role policy rules
	if authorized, reason, err = o.makeDecision(globalRole, attr); authorized == authorizer.DecisionAllow {
		return authorized, reason, err
	}

	// it's not in cluster resource, permission denied
	if attr.GetCluster() == "" {
		return authorizer.DecisionDeny, "permission undefined", nil
	}

	clusterRole, err := o.am.GetRoleOfUserInTargetScope(iamv1alpha2.ClusterScope, attr.GetCluster(), attr.GetUser().GetName())
	if err != nil {
		if errors.IsNotFound(err) {
			return authorizer.DecisionDeny, err.Error(), nil
		}
		return authorizer.DecisionDeny, "", err
	}

	// check cluster role policy rules
	if a, r, e := o.makeDecision(clusterRole, attr); a == authorizer.DecisionAllow {
		return a, r, e
	}

	// it's not in cluster resource, permission denied
	if attr.GetWorkspace() == "" && attr.GetNamespace() == "" {
		return authorizer.DecisionDeny, "permission undefined", nil
	}

	workspaceRole, err := o.am.GetRoleOfUserInTargetScope(iamv1alpha2.WorkspaceScope, attr.GetWorkspace(), attr.GetUser().GetName())
	if err != nil {
		if errors.IsNotFound(err) {
			return authorizer.DecisionDeny, err.Error(), nil
		}
		return authorizer.DecisionDeny, "", err
	}

	// check workspace role policy rules
	if a, r, e := o.makeDecision(workspaceRole, attr); a == authorizer.DecisionAllow {
		return a, r, e
	}

	// it's not in workspace resource, permission denied
	if attr.GetNamespace() == "" {
		return authorizer.DecisionDeny, "permission undefined", nil
	}

	if attr.GetNamespace() != "" {
		namespaceRole, err := o.am.GetRoleOfUserInTargetScope(iamv1alpha2.NamespaceScope, attr.GetNamespace(), attr.GetUser().GetName())
		if err != nil {
			if errors.IsNotFound(err) {
				return authorizer.DecisionDeny, err.Error(), nil
			}
			return authorizer.DecisionDeny, "", err
		}
		// check namespace role policy rules
		if a, r, e := o.makeDecision(namespaceRole, attr); a == authorizer.DecisionAllow {
			return a, r, e
		}
	}

	return authorizer.DecisionDeny, "", nil
}

// Make decision base on role
func (o *opaAuthorizer) makeDecision(role *iamv1alpha2.Role, a authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {

	for _, ruleRef := range role.Rules {
		rule, err := o.am.GetPolicyRule(ruleRef.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return authorizer.DecisionDeny, "", err
		}
		// Call the rego.New function to create an object that can be prepared or evaluated
		//  After constructing a new rego.Rego object you can call PrepareForEval() to obtain an executable query
		query, err := rego.New(rego.Query("data.authz.allow"), rego.Module("authz.rego", rule.Rego)).PrepareForEval(context.Background())

		if err != nil {
			klog.Errorf("rule syntax error:%s", err)
			continue
		}

		// The policy decision is contained in the results returned by the Eval() call. You can inspect the decision and handle it accordingly.
		results, err := query.Eval(context.Background(), rego.EvalInput(a))

		if err != nil {
			klog.Errorf("rule syntax error:%s", err)
			continue
		}

		if len(results) > 0 && results[0].Expressions[0].Value == true {
			return authorizer.DecisionAllow, "", nil
		}
	}

	return authorizer.DecisionDeny, "permission undefined", nil
}

func NewOPAAuthorizer(am am.AccessManagementInterface) *opaAuthorizer {
	return &opaAuthorizer{am: am}
}
