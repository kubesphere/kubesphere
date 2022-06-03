/*
Copyright 2021 The KubeSphere Authors.

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

package v1alpha1

import (
	"errors"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"regexp"
)

var (
	ErrProviderNotFound             = errors.New("the provide of policy was not found")
	ErrTemplateOfProviderNotSupport = errors.New("the template not support the specific provider")

	ErrPolicyTemplateNotFound = errors.New("the policy template was not found")

	ErrPolicyNotFound      = errors.New("the policy was not found")
	ErrPolicyAlreadyExists = errors.New("the policy already exists")

	ErrRuleNotFound      = errors.New("the rule was not found")
	ErrRuleAlreadyExists = errors.New("the rule already exists")

	policyNameMatcher = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	ruleNameMatcher   = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

type PolicyTemplate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Targets     []PolicyTemplateTarget `json:"targets"`
	Parameters  Parameters             `json:"parameters,omitempty"  description:"policy rule parameters"`
}

type Policy struct {
	Name           string         `json:"name"`
	PolicyTemplate string         `json:"templateName,omitempty"`
	Provider       string         `json:"provider,omitempty"`
	Description    string         `json:"description,omitempty"`
	Targets        []PolicyTarget `json:"targets"`
	Parameters     Parameters     `json:"parameters,omitempty"  description:"policy rule parameters"`
}

type Rule struct {
	Name        string                 `json:"name"`
	Policy      string                 `json:"templateName,omitempty"`
	Provider    string                 `json:"provider,omitempty"`
	Description string                 `json:"description,omitempty"`
	Match       Match                  `json:"match,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// List

type PolicyTemplateList struct {
	Total int               `json:"total"`
	Items []*PolicyTemplate `json:"items"`
}

type PolicyList struct {
	Total int       `json:"total"`
	Items []*Policy `json:"items"`
}

type RuleList struct {
	Total int     `json:"total"`
	Items []*Rule `json:"items"`
}

// Get

// PolicyTemplateDetail for Get(GET) policy template
type PolicyTemplateDetail struct {
	PolicyTemplate
}

// PolicyDetail for Get(GET) policy
type PolicyDetail struct {
	Policy
}

// RuleDetail for Get(GET) rule
type RuleDetail struct {
	Rule
}

// Create and Update

// PostPolicy for Create(POST), Update(PUT) policy
type PostPolicy struct {
	Policy
}

// PostRule for Create(POST), Update(PUT) rule
type PostRule struct {
	Rule
}

type PolicyTemplateTarget struct {
	Target     string   `json:"target,omitempty" description:"target name"`
	Provider   string   `json:"provider,omitempty" description:"target admission provider"`
	Expression string   `json:"expression,omitempty" description:"expression string like rego etc."`
	Import     []string `json:"import,omitempty" description:"import from other resource"`
}

type PolicyTarget struct {
	Target     string   `json:"target,omitempty" description:"target name"`
	Expression string   `json:"expression,omitempty" description:"expression string like rego etc."`
	Import     []string `json:"import,omitempty" description:"validation for policy rule parameters"`
}

type Parameters struct {
	Validation *Validation `json:"validation,omitempty" description:"validation for policy rule parameters"`
}

type Validation struct {
	OpenAPIV3Schema *apiextensions.JSONSchemaProps `json:"openAPIV3Schema,omitempty"`
	LegacySchema    bool                           `json:"legacySchema,omitempty"`
}

// Match selects objects to apply mutations to.
type Match struct {
	Namespaces         []string `json:"namespaces,omitempty"`
	ExcludedNamespaces []string `json:"excludedNamespaces,omitempty"`
}

func (r *PostPolicy) Validate() error {
	var errs []error

	if r.Name == "" {
		errs = append(errs, errors.New("name can not be empty"))
	} else {
		if !policyNameMatcher.MatchString(r.Name) {
			errs = append(errs, errors.New("rule name must match regular expression ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"))
		}
	}

	if r.Targets == nil || len(r.Targets) == 0 {
		errs = append(errs, errors.New("targets can not be empty"))
	}

	return utilerrors.NewAggregate(errs)
}

func (r *PostRule) Validate() error {
	var errs []error

	if r.Name == "" {
		errs = append(errs, errors.New("name can not be empty"))
	} else {
		if !ruleNameMatcher.MatchString(r.Name) {
			errs = append(errs, errors.New("rule name must match regular expression ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"))
		}
	}

	return utilerrors.NewAggregate(errs)
}
