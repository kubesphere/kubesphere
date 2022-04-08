// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package rules

import (
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"kubesphere.io/kubesphere/pkg/api/alerting/v2alpha1"
)

type ResourceRuleCollection struct {
	GroupSet  map[string]struct{}
	IdRules   map[string]*ResourceRuleItem
	NameRules map[string][]*ResourceRuleItem
}

type ResourceRuleItem struct {
	ResourceName string
	RuleWithGroup
}

type ResourceRule struct {
	Level  v2alpha1.RuleLevel
	Custom bool
	ResourceRuleItem
}

type ResourceRuleChunk struct {
	Level            v2alpha1.RuleLevel
	Custom           bool
	ResourceRulesMap map[string]*ResourceRuleCollection
}

type RuleWithGroup struct {
	Group string
	Id    string
	promresourcesv1.Rule
}
