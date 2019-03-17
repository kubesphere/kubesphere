package models

// GrafanaInfo provides information to access Grafana dashboards
type GrafanaInfo struct {
	URL                   string `json:"url"`
	ServiceDashboardPath  string `json:"serviceDashboardPath"`
	WorkloadDashboardPath string `json:"workloadDashboardPath"`
	VarNamespace          string `json:"varNamespace"`
	VarService            string `json:"varService"`
	VarWorkload           string `json:"varWorkload"`
}
