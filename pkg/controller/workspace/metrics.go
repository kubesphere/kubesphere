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
