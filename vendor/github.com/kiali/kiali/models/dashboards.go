package models

import (
	"fmt"
	"sort"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/prometheus"
)

// MonitoringDashboard is the model representing custom monitoring dashboard, transformed from MonitoringDashboard k8s resource
type MonitoringDashboard struct {
	Title        string        `json:"title"`
	Charts       []Chart       `json:"charts"`
	Aggregations []Aggregation `json:"aggregations"`
}

// Chart is the model representing a custom chart, transformed from charts in MonitoringDashboard k8s resource
type Chart struct {
	Name      string               `json:"name"`
	Unit      string               `json:"unit"`
	Spans     int                  `json:"spans"`
	Metric    *prometheus.Metric   `json:"metric"`
	Histogram prometheus.Histogram `json:"histogram"`
}

// ConvertChart converts a k8s chart (from MonitoringDashboard k8s resource) into this models chart
func ConvertChart(from kubernetes.MonitoringDashboardChart) Chart {
	return Chart{
		Name:  from.Name,
		Unit:  from.Unit,
		Spans: from.Spans,
	}
}

// Aggregation is the model representing label's allowed aggregation, transformed from aggregation in MonitoringDashboard k8s resource
type Aggregation struct {
	Label       string `json:"label"`
	DisplayName string `json:"displayName"`
}

// ConvertAggregations converts a k8s aggregations (from MonitoringDashboard k8s resource) into this models aggregations
// Results are sorted by DisplayName
func ConvertAggregations(from kubernetes.MonitoringDashboardSpec) []Aggregation {
	uniqueAggs := make(map[string]Aggregation)
	for _, chart := range from.Charts {
		for _, agg := range chart.Aggregations {
			uniqueAggs[agg.DisplayName] = Aggregation{Label: agg.Label, DisplayName: agg.DisplayName}
		}
	}
	aggs := []Aggregation{}
	for _, agg := range uniqueAggs {
		aggs = append(aggs, agg)
	}
	sort.Slice(aggs, func(i, j int) bool {
		return aggs[i].DisplayName < aggs[j].DisplayName
	})
	return aggs
}

func buildIstioAggregations(local, remote string) []Aggregation {
	return []Aggregation{
		{
			Label:       fmt.Sprintf("%s_version", local),
			DisplayName: "Local version",
		},
		{
			Label:       fmt.Sprintf("%s_app", remote),
			DisplayName: "Remote app",
		},
		{
			Label:       fmt.Sprintf("%s_version", remote),
			DisplayName: "Remote version",
		},
		{
			Label:       "response_code",
			DisplayName: "Response code",
		},
	}
}

// PrepareIstioDashboard prepares the Istio dashboard title and aggregations dynamically for input values
func PrepareIstioDashboard(direction, local, remote string) MonitoringDashboard {
	return MonitoringDashboard{
		Title:        fmt.Sprintf("%s Metrics", direction),
		Aggregations: buildIstioAggregations(local, remote),
	}
}

// Runtime holds the runtime title and associated dashboard template(s)
type Runtime struct {
	Name          string         `json:"name"`
	DashboardRefs []DashboardRef `json:"dashboardRefs"`
}

// DashboardRef holds template name and title for a custom dashboard
type DashboardRef struct {
	Template string `json:"template"`
	Title    string `json:"title"`
}
