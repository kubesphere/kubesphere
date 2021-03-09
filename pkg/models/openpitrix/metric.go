/*
Copyright 2020 The KubeSphere Authors.

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

package openpitrix

import (
	compbasemetrics "k8s.io/component-base/metrics"
	"kubesphere.io/kubesphere/pkg/utils/metrics"
)

var (
	appTemplateCreationCounter = compbasemetrics.NewCounterVec(
		&compbasemetrics.CounterOpts{
			Name:           "application_template_creation",
			Help:           "Counter of application template creation broken out for each workspace, name and create state",
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{"workspace", "name", "state"},
	)
)

func init() {
	metrics.MustRegister(appTemplateCreationCounter)
}
