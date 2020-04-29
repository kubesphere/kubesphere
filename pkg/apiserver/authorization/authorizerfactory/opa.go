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
	rbacv1 "k8s.io/api/rbac/v1"
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
	defaultRegoQuery = "data.authz.allow"
)

// Make decision by request attributes
func (o *opaAuthorizer) Authorize(attr authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
	// Make decisions based on the authorization policy of different levels of roles
	// Error returned when an internal error occurs
	// Reason must be returned when access is denied
	globalRole, err := o.am.GetGlobalRoleOfUser(attr.GetUser().GetName())

	if err != nil {
		if errors.IsNotFound(err) {
			return authorizer.DecisionNoOpinion, "", nil
		}
		return authorizer.DecisionNoOpinion, "", err
	}

	// check global policy rules
	if authorized, reason, err = o.makeDecision(globalRole, attr); authorized == authorizer.DecisionAllow {
		return authorized, reason, nil
	}

	// it's global resource, permission denied
	if attr.GetResourceScope() == iamv1alpha2.GlobalScope {
		return authorizer.DecisionNoOpinion, "", nil
	}

	if attr.GetResourceScope() == iamv1alpha2.WorkspaceScope {
		workspaceRole, err := o.am.GetWorkspaceRoleOfUser(attr.GetUser().GetName(), attr.GetWorkspace())
		if err != nil {
			if errors.IsNotFound(err) {
				return authorizer.DecisionNoOpinion, "", nil
			}
			return authorizer.DecisionNoOpinion, "", err
		}

		// check workspace role policy rules
		if authorized, reason, err := o.makeDecision(workspaceRole, attr); authorized == authorizer.DecisionAllow {
			return authorized, reason, err
		} else if err != nil {
			return authorizer.DecisionNoOpinion, "", err
		}

		return authorizer.DecisionNoOpinion, "", nil
	}

	if attr.GetResourceScope() == iamv1alpha2.NamespaceScope {
		role, err := o.am.GetNamespaceRoleOfUser(attr.GetUser().GetName(), attr.GetNamespace())
		if err != nil {
			if errors.IsNotFound(err) {
				return authorizer.DecisionNoOpinion, "", nil
			}
			return authorizer.DecisionNoOpinion, "", err
		}
		// check namespace role policy rules
		if authorized, reason, err := o.makeDecision(role, attr); authorized == authorizer.DecisionAllow {
			return authorized, reason, err
		} else if err != nil {
			return authorizer.DecisionNoOpinion, "", err
		}

		return authorizer.DecisionNoOpinion, "", nil
	}

	clusterRole, err := o.am.GetClusterRoleOfUser(attr.GetUser().GetName(), attr.GetCluster())

	if errors.IsNotFound(err) {
		return authorizer.DecisionNoOpinion, "", nil
	}

	if err != nil {
		return authorizer.DecisionNoOpinion, "", err
	}

	// check cluster role policy rules
	if authorized, reason, err := o.makeDecision(clusterRole, attr); authorized == authorizer.DecisionAllow {
		return authorized, reason, nil
	} else if err != nil {
		return authorizer.DecisionNoOpinion, "", err
	}

	return authorizer.DecisionNoOpinion, "", nil
}

// Make decision base on role
func (o *opaAuthorizer) makeDecision(role interface{}, a authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {

	regoPolicy := ""

	// override
	if globalRole, ok := role.(*iamv1alpha2.GlobalRole); ok {
		if overrideRego, ok := globalRole.Annotations[iamv1alpha2.RegoOverrideAnnotation]; ok {
			regoPolicy = overrideRego
		}
	} else if workspaceRole, ok := role.(*iamv1alpha2.WorkspaceRole); ok {
		if overrideRego, ok := workspaceRole.Annotations[iamv1alpha2.RegoOverrideAnnotation]; ok {
			regoPolicy = overrideRego
		}
	} else if clusterRole, ok := role.(*rbacv1.ClusterRole); ok {
		if overrideRego, ok := clusterRole.Annotations[iamv1alpha2.RegoOverrideAnnotation]; ok {
			regoPolicy = overrideRego
		}
	} else if role, ok := role.(*rbacv1.Role); ok {
		if overrideRego, ok := role.Annotations[iamv1alpha2.RegoOverrideAnnotation]; ok {
			regoPolicy = overrideRego
		}
	}

	if regoPolicy == "" {
		return authorizer.DecisionNoOpinion, "", nil
	}

	// Call the rego.New function to create an object that can be prepared or evaluated
	//  After constructing a new rego.Rego object you can call PrepareForEval() to obtain an executable query
	query, err := rego.New(rego.Query(defaultRegoQuery), rego.Module("authz.rego", regoPolicy)).PrepareForEval(context.Background())

	if err != nil {
		klog.Errorf("syntax error:%s,refer: %s+v", err, role)
		return authorizer.DecisionNoOpinion, "", err
	}

	// The policy decision is contained in the results returned by the Eval() call. You can inspect the decision and handle it accordingly.
	results, err := query.Eval(context.Background(), rego.EvalInput(a))

	if err != nil {
		klog.Errorf("syntax error:%s,refer: %s+v", err, role)
		return authorizer.DecisionNoOpinion, "", err
	}

	if len(results) > 0 && results[0].Expressions[0].Value == true {
		return authorizer.DecisionAllow, "", nil
	}

	return authorizer.DecisionNoOpinion, "", nil
}

func NewOPAAuthorizer(am am.AccessManagementInterface) *opaAuthorizer {
	return &opaAuthorizer{am: am}
}
