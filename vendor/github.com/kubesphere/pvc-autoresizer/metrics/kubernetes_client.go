package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	runtimemetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Metrics subsystem and all of the keys used by the metrics client.
const (
	KubernetesClientSubsystem    = "kubernetes_client"
	KubernetesClientFailTotalKey = "fail_total"
)

func init() {
	registerKubernetesClientMetrics()
}

type KubernetesClientFailTotalAdapter struct {
	metric prometheus.Counter
}

func (a *KubernetesClientFailTotalAdapter) Increment() {
	a.metric.Inc()
}

var (
	kubernetesClientFailTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: MetricsNamespace,
		Subsystem: KubernetesClientSubsystem,
		Name:      KubernetesClientFailTotalKey,
		Help:      "counter that indicates how many API requests to kube-api server are failed.",
	})

	KubernetesClientFailTotal *KubernetesClientFailTotalAdapter = &KubernetesClientFailTotalAdapter{metric: kubernetesClientFailTotal}
)

func registerKubernetesClientMetrics() {
	runtimemetrics.Registry.MustRegister(kubernetesClientFailTotal)
}
