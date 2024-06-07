/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspace

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	workspaceOperation = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ks_controller_manager_workspace_operation",
			Help: "Counter of ks controller manager workspace operation broken out for each operation, name",
			// This metric is used for verifying api call latencies SLO,
			// as well as tracking regressions in this aspects.
			// Thus we customize buckets significantly, to empower both usecases.
		},
		[]string{"operation", "name"},
	)
)

func init() {
	metrics.Registry.MustRegister(workspaceOperation)
}
