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
package workspace

import (
	compbasemetrics "k8s.io/component-base/metrics"

	"kubesphere.io/kubesphere/pkg/utils/metrics"
)

var (
	workspaceOperation = compbasemetrics.NewCounterVec(
		&compbasemetrics.CounterOpts{
			Name: "ks_controller_manager_workspace_operation",
			Help: "Counter of ks controller manager workspace operation broken out for each operation, name",
			// This metric is used for verifying api call latencies SLO,
			// as well as tracking regressions in this aspects.
			// Thus we customize buckets significantly, to empower both usecases.
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{"operation", "name"},
	)
)

func init() {
	metrics.MustRegister(workspaceOperation)
}
