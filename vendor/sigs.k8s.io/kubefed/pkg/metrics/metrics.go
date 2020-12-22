/*
Copyright 2018 The Kubernetes Authors.

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

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	kubefedClusterTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kubefedcluster_total",
			Help: "Number of total kubefed cluster in a specific state.",
		}, []string{"state", "cluster"},
	)

	joinedClusterTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "joined_cluster_total",
			Help: "Number of total joined clusters.",
		},
	)

	clusterHealthStatusDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "cluster_health_status_duration_seconds",
			Help:    "Time taken for the cluster health periodic function.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
		},
	)

	clusterClientConnectionDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "cluster_client_connection_duration_seconds",
			Help:    "Time taken for the cluster client connection function.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
		},
	)

	reconcileFederatedResourcesDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "reconcile_federated_resources_duration_seconds",
			Help:    "Time taken to reconcile federated resources in the target clusters.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
		},
	)

	joinedClusterDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "join_cluster_duration_seconds",
			Help:    "Time taken to join a cluster.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
		},
	)

	unjoinedClusterDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "unjoin_cluster_duration_seconds",
			Help:    "Time taken to unjoin a cluster.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
		},
	)

	dispatchOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dispatch_operation_duration_seconds",
			Help:    "Time taken to run dispatch operation.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
		}, []string{"action"},
	)

	controllerRuntimeReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "controller_runtime_reconcile_duration_seconds",
			Help:    "Time taken by various parts of Kubefed controllers reconciliation loops.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
		}, []string{"controller"},
	)

	controllerRuntimeReconcileDurationSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "controller_runtime_reconcile_quantile_seconds",
			Help:   "Quantiles of time taken by various parts of Kubefed controllers reconciliation loops.",
			MaxAge: time.Hour,
		}, []string{"controller"},
	)
)

const (
	// LogReconcileLongDurationThreshold defines the duration after which long function
	// duration will be logged.
	LogReconcileLongDurationThreshold = 10 * time.Second

	ClusterNotReady = "notready"
	ClusterReady    = "ready"
	ClusterOffline  = "offline"
)

// RegisterAll registers all metrics.
func RegisterAll() {
	metrics.Registry.MustRegister(
		kubefedClusterTotal,
		joinedClusterTotal,
		reconcileFederatedResourcesDuration,
		clusterHealthStatusDuration,
		clusterClientConnectionDuration,
		joinedClusterDuration,
		unjoinedClusterDuration,
		dispatchOperationDuration,
		controllerRuntimeReconcileDuration,
		controllerRuntimeReconcileDurationSummary,
	)
}

// RegisterKubefedClusterTotal records number of kubefed clusters in a specific state
func RegisterKubefedClusterTotal(state, cluster string) {
	switch state {
	case ClusterReady:
		kubefedClusterTotal.WithLabelValues(state, cluster).Set(1)
		kubefedClusterTotal.WithLabelValues(ClusterNotReady, cluster).Set(0)
		kubefedClusterTotal.WithLabelValues(ClusterOffline, cluster).Set(0)
	case ClusterNotReady:
		kubefedClusterTotal.WithLabelValues(state, cluster).Set(1)
		kubefedClusterTotal.WithLabelValues(ClusterOffline, cluster).Set(0)
		kubefedClusterTotal.WithLabelValues(ClusterReady, cluster).Set(0)
	case ClusterOffline:
		kubefedClusterTotal.WithLabelValues(state, cluster).Set(1)
		kubefedClusterTotal.WithLabelValues(ClusterNotReady, cluster).Set(0)
		kubefedClusterTotal.WithLabelValues(ClusterReady, cluster).Set(0)
	}
}

// JoinedClusterTotalInc increases by one the number of joined kubefed clusters
func JoinedClusterTotalInc() {
	joinedClusterTotal.Inc()
}

// JoinedClusterTotalDec decreases by one the number of joined kubefed clusters
func JoinedClusterTotalDec() {
	joinedClusterTotal.Dec()
}

// DispatchOperationDurationFromStart records the duration of the step identified by the action name
func DispatchOperationDurationFromStart(action string, start time.Time) {
	duration := time.Since(start)
	dispatchOperationDuration.WithLabelValues(action).Observe(duration.Seconds())
}

// ClusterHealthStatusDurationFromStart records the duration of the cluster health status operation
func ClusterHealthStatusDurationFromStart(start time.Time) {
	duration := time.Since(start)
	clusterHealthStatusDuration.Observe(duration.Seconds())
}

// ClusterClientConnectionDurationFromStart records the duration of the cluster client connection operation
func ClusterClientConnectionDurationFromStart(start time.Time) {
	duration := time.Since(start)
	clusterClientConnectionDuration.Observe(duration.Seconds())
}

// JoinedClusterDurationFromStart records the duration of the cluster joined operation
func JoinedClusterDurationFromStart(start time.Time) {
	duration := time.Since(start)
	joinedClusterDuration.Observe(duration.Seconds())
}

// UnjoinedClusterDurationFromStart records the duration of the cluster unjoined operation
func UnjoinedClusterDurationFromStart(start time.Time) {
	duration := time.Since(start)
	unjoinedClusterDuration.Observe(duration.Seconds())
}

// ReconcileFederatedResourcesDurationFromStart records the duration of the federation of resources
func ReconcileFederatedResourcesDurationFromStart(start time.Time) {
	duration := time.Since(start)
	reconcileFederatedResourcesDuration.Observe(duration.Seconds())
}

// UpdateControllerReconcileDurationFromStart records the duration of the reconcile loop
// of a controller
func UpdateControllerReconcileDurationFromStart(controller string, start time.Time) {
	duration := time.Since(start)
	UpdateControllerReconcileDuration(controller, duration)
}

// UpdateControllerReconcileDuration records the duration of the reconcile function of a controller
func UpdateControllerReconcileDuration(controller string, duration time.Duration) {
	if duration > LogReconcileLongDurationThreshold {
		klog.V(4).Infof("Reconcile loop %s took %v to complete", controller, duration)
	}

	controllerRuntimeReconcileDurationSummary.WithLabelValues(controller).Observe(duration.Seconds())
	controllerRuntimeReconcileDuration.WithLabelValues(controller).Observe(duration.Seconds())
}
