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

const (
	permissionUndefined = "permission undefined"
	defaultRegoQuery    = "data.authz.allow"
)

// Make decision by request attributes
func (o *opaAuthorizer) Authorize(attr authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {

	// Make decisions based on the authorization policy of different levels of roles
	// Error returned when an internal error occurs
	// Reason must be returned when access is denied
	globalRole, err := o.am.GetRoleOfUserInTargetScope(iamv1alpha2.GlobalScope, "", attr.GetUser().GetName())
	if err != nil {
		if errors.IsNotFound(err) {
			return authorizer.DecisionDeny, err.Error(), nil
		}
		return authorizer.DecisionDeny, "", err
	}

	// check global role policy rules
	if authorized, reason, err = o.makeDecision(globalRole, attr); authorized == authorizer.DecisionAllow {
		return authorized, reason, nil
	}

	// it's not in cluster resource, permission denied
	if attr.GetCluster() == "" {
		return authorizer.DecisionDeny, permissionUndefined, nil
	}

	clusterRole, err := o.am.GetRoleOfUserInTargetScope(iamv1alpha2.ClusterScope, attr.GetCluster(), attr.GetUser().GetName())
	if err != nil {
		if errors.IsNotFound(err) {
			return authorizer.DecisionDeny, err.Error(), nil
		}
		return authorizer.DecisionDeny, "", err
	}

	// check cluster role policy rules
	if authorized, reason, err := o.makeDecision(clusterRole, attr); authorized == authorizer.DecisionAllow {
		return authorized, reason, nil
	} else if err != nil {
		return authorizer.DecisionDeny, "", err
	}

	// it's not in cluster resource, permission denied
	if attr.GetWorkspace() == "" && attr.GetNamespace() == "" {
		return authorizer.DecisionDeny, permissionUndefined, nil
	}

	workspaceRole, err := o.am.GetRoleOfUserInTargetScope(iamv1alpha2.WorkspaceScope, attr.GetWorkspace(), attr.GetUser().GetName())
	if err != nil {
		if errors.IsNotFound(err) {
			return authorizer.DecisionDeny, err.Error(), nil
		}
		return authorizer.DecisionDeny, "", err
	}

	// check workspace role policy rules
	if authorized, reason, err := o.makeDecision(workspaceRole, attr); authorized == authorizer.DecisionAllow {
		return authorized, reason, err
	} else if err != nil {
		return authorizer.DecisionDeny, "", err
	}

	// it's not in workspace resource, permission denied
	if attr.GetNamespace() == "" {
		return authorizer.DecisionDeny, permissionUndefined, nil
	}

	namespaceRole, err := o.am.GetRoleOfUserInTargetScope(iamv1alpha2.NamespaceScope, attr.GetNamespace(), attr.GetUser().GetName())
	if err != nil {
		if errors.IsNotFound(err) {
			return authorizer.DecisionDeny, err.Error(), nil
		}
		return authorizer.DecisionDeny, "", err
	}
	// check namespace role policy rules
	if authorized, reason, err := o.makeDecision(namespaceRole, attr); authorized == authorizer.DecisionAllow {
		return authorized, reason, err
	} else if err != nil {
		return authorizer.DecisionDeny, "", err
	}

	return authorizer.DecisionDeny, permissionUndefined, nil
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
		query, err := rego.New(rego.Query(defaultRegoQuery), rego.Module("authz.rego", rule.Rego)).PrepareForEval(context.Background())

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

	return authorizer.DecisionDeny, permissionUndefined, nil
}

func NewOPAAuthorizer(am am.AccessManagementInterface) *opaAuthorizer {
	return &opaAuthorizer{am: am}
}
