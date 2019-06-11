package business

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus"
)

// DashboardsService deals with fetching dashboards from k8s client
type DashboardsService struct {
	prom prometheus.ClientInterface
	mon  kubernetes.KialiMonitoringInterface
}

// NewDashboardsService initializes this business service
func NewDashboardsService(mon kubernetes.KialiMonitoringInterface, prom prometheus.ClientInterface) DashboardsService {
	return DashboardsService{prom: prom, mon: mon}
}

func (in *DashboardsService) loadDashboardResource(namespace, template string) (*kubernetes.MonitoringDashboard, error) {
	// There is an override mechanism with dashboards: default dashboards can be provided in Kiali namespace,
	// and can be overriden in app namespace.
	// So we look for the one in app namespace first, and only if not found fallback to the one in istio-system.
	dashboard, err := in.mon.GetDashboard(namespace, template)
	if err != nil {
		cfg := config.Get()
		dashboard, err = in.mon.GetDashboard(cfg.IstioNamespace, template)
		if err != nil {
			return nil, err
		}
	}

	return dashboard, nil
}

// GetDashboard returns a dashboard filled-in with target data
func (in *DashboardsService) GetDashboard(params prometheus.CustomMetricsQuery, template string) (*models.MonitoringDashboard, error) {
	dashboard, err := in.loadDashboardResource(params.Namespace, template)
	if err != nil {
		return nil, err
	}

	aggLabels := models.ConvertAggregations(dashboard.Spec)
	labels := fmt.Sprintf(`{namespace="%s",app="%s"`, params.Namespace, params.App)
	if params.Version != "" {
		labels += fmt.Sprintf(`,version="%s"`, params.Version)
	} else {
		// For app-based dashboards, we automatically add a possible aggregation/grouping over versions
		versionsAgg := models.Aggregation{
			Label:       "version",
			DisplayName: "Version",
		}
		aggLabels = append([]models.Aggregation{versionsAgg}, aggLabels...)
	}
	labels += "}"
	grouping := strings.Join(params.ByLabels, ",")

	wg := sync.WaitGroup{}
	wg.Add(len(dashboard.Spec.Charts))
	filledCharts := make([]models.Chart, len(dashboard.Spec.Charts))

	for i, c := range dashboard.Spec.Charts {
		go func(idx int, chart kubernetes.MonitoringDashboardChart) {
			defer wg.Done()
			filledCharts[idx] = models.ConvertChart(chart)
			if chart.DataType == kubernetes.Raw {
				aggregator := params.RawDataAggregator
				if chart.Aggregator != "" {
					aggregator = chart.Aggregator
				}
				filledCharts[idx].Metric = in.prom.FetchRange(chart.MetricName, labels, grouping, aggregator, &params.BaseMetricsQuery)
			} else if chart.DataType == kubernetes.Rate {
				filledCharts[idx].Metric = in.prom.FetchRateRange(chart.MetricName, labels, grouping, &params.BaseMetricsQuery)
			} else {
				filledCharts[idx].Histogram = in.prom.FetchHistogramRange(chart.MetricName, labels, grouping, &params.BaseMetricsQuery)
			}
		}(i, c)
	}

	wg.Wait()
	return &models.MonitoringDashboard{
		Title:        dashboard.Spec.Title,
		Charts:       filledCharts,
		Aggregations: aggLabels,
	}, nil
}

type istioChart struct {
	models.Chart
	refName string
}

var istioCharts = []istioChart{
	{
		Chart: models.Chart{
			Name:  "Request volume",
			Unit:  "ops",
			Spans: 6,
		},
		refName: "request_count",
	},
	{
		Chart: models.Chart{
			Name:  "Request duration",
			Unit:  "s",
			Spans: 6,
		},
		refName: "request_duration",
	},
	{
		Chart: models.Chart{
			Name:  "Request size",
			Unit:  "B",
			Spans: 6,
		},
		refName: "request_size",
	},
	{
		Chart: models.Chart{
			Name:  "Response size",
			Unit:  "B",
			Spans: 6,
		},
		refName: "response_size",
	},
	{
		Chart: models.Chart{
			Name:  "TCP received",
			Unit:  "bps",
			Spans: 6,
		},
		refName: "tcp_received",
	},
	{
		Chart: models.Chart{
			Name:  "TCP sent",
			Unit:  "bps",
			Spans: 6,
		},
		refName: "tcp_sent",
	},
}

// GetIstioDashboard returns Istio dashboard (currently hard-coded) filled-in with metrics
func (in *DashboardsService) GetIstioDashboard(params prometheus.IstioMetricsQuery) (*models.MonitoringDashboard, error) {
	var dashboard models.MonitoringDashboard
	// Copy dashboard
	if params.Direction == "inbound" {
		dashboard = models.PrepareIstioDashboard("Inbound", "destination", "source")
	} else {
		dashboard = models.PrepareIstioDashboard("Outbound", "source", "destination")
	}

	metrics := in.prom.GetMetrics(&params)

	for _, chartTpl := range istioCharts {
		newChart := chartTpl.Chart
		if metric, ok := metrics.Metrics[chartTpl.refName]; ok {
			newChart.Metric = metric
		}
		if histo, ok := metrics.Histograms[chartTpl.refName]; ok {
			newChart.Histogram = histo
		}
		dashboard.Charts = append(dashboard.Charts, newChart)
	}

	return &dashboard, nil
}

func (in *DashboardsService) buildRuntimesList(namespace string, templatesNames []string) []models.Runtime {
	dashboards := make([]*kubernetes.MonitoringDashboard, len(templatesNames))
	wg := sync.WaitGroup{}
	wg.Add(len(templatesNames))
	for idx, template := range templatesNames {
		go func(i int, tpl string) {
			defer wg.Done()
			dashboard, err := in.loadDashboardResource(namespace, tpl)
			if err != nil {
				log.Errorf("Cannot get dashboard %s in namespace %s. Error was: %v", tpl, namespace, err)
			} else {
				dashboards[i] = dashboard
			}
		}(idx, template)
	}

	wg.Wait()

	runtimes := []models.Runtime{}
	for _, dashboard := range dashboards {
		if dashboard == nil {
			continue
		}
		runtime := getDashboardRuntime(dashboard)
		ref := models.DashboardRef{
			Template: dashboard.Metadata["name"].(string),
			Title:    dashboard.Spec.Title,
		}
		found := false
		for i := range runtimes {
			rtObj := &runtimes[i]
			if rtObj.Name == runtime {
				rtObj.DashboardRefs = append(rtObj.DashboardRefs, ref)
				found = true
				break
			}
		}
		if !found {
			runtimes = append(runtimes, models.Runtime{
				Name:          runtime,
				DashboardRefs: []models.DashboardRef{ref},
			})
		}
	}
	return runtimes
}

func getDashboardRuntime(dashboard *kubernetes.MonitoringDashboard) string {
	if labels, ok := dashboard.Metadata["labels"]; ok {
		if labelsMap, ok := labels.(map[string]interface{}); ok {
			if runtime, ok := labelsMap["runtime"]; ok {
				return runtime.(string)
			}
		}
	}
	return dashboard.Spec.Title
}
