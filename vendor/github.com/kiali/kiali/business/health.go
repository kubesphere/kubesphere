package business

import (
	"time"

	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus"
	"github.com/kiali/kiali/prometheus/internalmetrics"
)

// HealthService deals with fetching health from various sources and convert to kiali model
type HealthService struct {
	prom prometheus.ClientInterface
	k8s  kubernetes.IstioClientInterface
}

// GetServiceHealth returns a service health (service request error rate)
func (in *HealthService) GetServiceHealth(namespace, service, rateInterval string, queryTime time.Time) (models.ServiceHealth, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "HealthService", "GetServiceHealth")
	defer promtimer.ObserveNow(&err)

	rqHealth, err := in.getServiceRequestsHealth(namespace, service, rateInterval, queryTime)
	return models.ServiceHealth{Requests: rqHealth}, err
}

// GetAppHealth returns an app health from just Namespace and app name (thus, it fetches data from K8S and Prometheus)
func (in *HealthService) GetAppHealth(namespace, app, rateInterval string, queryTime time.Time) (models.AppHealth, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "HealthService", "GetAppHealth")
	defer promtimer.ObserveNow(&err)

	appLabel := config.Get().IstioLabels.AppLabelName

	selectorLabels := make(map[string]string)
	selectorLabels[appLabel] = app
	labelSelector := labels.FormatLabels(selectorLabels)

	ws, err := fetchWorkloads(in.k8s, namespace, labelSelector)
	if err != nil {
		log.Errorf("Error fetching Workloads per namespace %s and app %s: %s", namespace, app, err)
		return models.AppHealth{}, err
	}

	return in.getAppHealth(namespace, app, rateInterval, queryTime, ws)
}

func (in *HealthService) getAppHealth(namespace, app, rateInterval string, queryTime time.Time, ws models.Workloads) (models.AppHealth, error) {
	health := models.EmptyAppHealth()

	// Perf: do not bother fetching request rate if not a single workload has sidecar
	hasSidecar := false
	for _, w := range ws {
		if w.IstioSidecar {
			hasSidecar = true
			break
		}
	}

	// Fetch services requests rates
	var errRate error
	if hasSidecar {
		rate, err := in.getAppRequestsHealth(namespace, app, rateInterval, queryTime)
		health.Requests = rate
		errRate = err
	}

	// Deployment status
	health.WorkloadStatuses = castWorkloadStatuses(ws)

	return health, errRate
}

// GetWorkloadHealth returns a workload health from just Namespace and workload (thus, it fetches data from K8S and Prometheus)
func (in *HealthService) GetWorkloadHealth(namespace, workload, rateInterval string, queryTime time.Time) (models.WorkloadHealth, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "HealthService", "GetWorkloadHealth")
	defer promtimer.ObserveNow(&err)

	w, err := fetchWorkload(in.k8s, namespace, workload)
	if err != nil {
		return models.WorkloadHealth{}, err
	}
	status := models.WorkloadStatus{
		Name:              w.Name,
		Replicas:          w.Replicas,
		AvailableReplicas: w.AvailableReplicas,
	}

	// Perf: do not bother fetching request rate if workload has no sidecar
	if !w.IstioSidecar {
		return models.WorkloadHealth{
			WorkloadStatus: status,
			Requests:       models.NewEmptyRequestHealth(),
		}, nil
	}

	rate, err := in.getWorkloadRequestsHealth(namespace, workload, rateInterval, queryTime)
	return models.WorkloadHealth{
		WorkloadStatus: status,
		Requests:       rate,
	}, err
}

// GetNamespaceAppHealth returns a health for all apps in given Namespace (thus, it fetches data from K8S and Prometheus)
func (in *HealthService) GetNamespaceAppHealth(namespace, rateInterval string, queryTime time.Time) (models.NamespaceAppHealth, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "HealthService", "GetNamespaceAppHealth")
	defer promtimer.ObserveNow(&err)

	appEntities, err := fetchNamespaceApps(in.k8s, namespace, "")
	if err != nil {
		return nil, err
	}
	return in.getNamespaceAppHealth(namespace, appEntities, rateInterval, queryTime)
}

func (in *HealthService) getNamespaceAppHealth(namespace string, appEntities namespaceApps, rateInterval string, queryTime time.Time) (models.NamespaceAppHealth, error) {
	allHealth := make(models.NamespaceAppHealth)

	// Perf: do not bother fetching request rate if not a single workload has sidecar
	hasSidecar := false

	// Prepare all data
	for app, entities := range appEntities {
		if app != "" {
			h := models.EmptyAppHealth()
			allHealth[app] = &h
			if entities != nil {
				h.WorkloadStatuses = castWorkloadStatuses(entities.Workloads)
				for _, w := range entities.Workloads {
					if w.IstioSidecar {
						hasSidecar = true
						break
					}
				}
			}
		}
	}

	var errRate error
	if hasSidecar {
		// Fetch services requests rates
		rates, err := in.prom.GetAllRequestRates(namespace, rateInterval, queryTime)
		errRate = err
		// Fill with collected request rates
		fillAppRequestRates(allHealth, rates)
	}

	return allHealth, errRate
}

// GetNamespaceServiceHealth returns a health for all services in given Namespace (thus, it fetches data from K8S and Prometheus)
func (in *HealthService) GetNamespaceServiceHealth(namespace, rateInterval string, queryTime time.Time) (models.NamespaceServiceHealth, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "HealthService", "GetNamespaceServiceHealth")
	defer promtimer.ObserveNow(&err)

	return in.getNamespaceServiceHealth(namespace, rateInterval, queryTime), nil
}

func (in *HealthService) getNamespaceServiceHealth(namespace string, rateInterval string, queryTime time.Time) models.NamespaceServiceHealth {
	allHealth := make(models.NamespaceServiceHealth)

	// Fetch services requests rates
	rates, _ := in.prom.GetNamespaceServicesRequestRates(namespace, rateInterval, queryTime)
	// Fill with collected request rates
	lblDestSvc := model.LabelName("destination_service_name")
	for _, sample := range rates {
		service := string(sample.Metric[lblDestSvc])
		health, ok := allHealth[service]
		if !ok {
			health = &models.ServiceHealth{Requests: models.NewEmptyRequestHealth()}
			allHealth[service] = health
		}
		health.Requests.AggregateInbound(sample)
	}

	return allHealth
}

// GetNamespaceWorkloadHealth returns a health for all workloads in given Namespace (thus, it fetches data from K8S and Prometheus)
func (in *HealthService) GetNamespaceWorkloadHealth(namespace, rateInterval string, queryTime time.Time) (models.NamespaceWorkloadHealth, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "HealthService", "GetNamespaceWorkloadHealth")
	defer promtimer.ObserveNow(&err)

	wl, err := fetchWorkloads(in.k8s, namespace, "")
	if err != nil {
		return nil, err
	}

	return in.getNamespaceWorkloadHealth(namespace, wl, rateInterval, queryTime), nil
}

func (in *HealthService) getNamespaceWorkloadHealth(namespace string, ws models.Workloads, rateInterval string, queryTime time.Time) models.NamespaceWorkloadHealth {
	// Perf: do not bother fetching request rate if not a single workload has sidecar
	hasSidecar := false

	allHealth := make(models.NamespaceWorkloadHealth)
	for _, w := range ws {
		allHealth[w.Name] = &models.WorkloadHealth{}
		allHealth[w.Name].WorkloadStatus = models.WorkloadStatus{
			Name:              w.Name,
			Replicas:          w.Replicas,
			AvailableReplicas: w.AvailableReplicas,
		}
		if w.IstioSidecar {
			hasSidecar = true
		}
	}

	if hasSidecar {
		// Fetch services requests rates
		rates, _ := in.prom.GetAllRequestRates(namespace, rateInterval, queryTime)
		// Fill with collected request rates
		fillWorkloadRequestRates(allHealth, rates)
	}

	return allHealth
}

// fillAppRequestRates aggregates requests rates from metrics fetched from Prometheus, and stores the result in the health map.
func fillAppRequestRates(allHealth models.NamespaceAppHealth, rates model.Vector) {
	lblDest := model.LabelName("destination_app")
	lblSrc := model.LabelName("source_app")
	for _, sample := range rates {
		name := string(sample.Metric[lblDest])
		if health, ok := allHealth[name]; ok {
			health.Requests.AggregateInbound(sample)
		}
		name = string(sample.Metric[lblSrc])
		if health, ok := allHealth[name]; ok {
			health.Requests.AggregateOutbound(sample)
		}
	}
}

// fillWorkloadRequestRates aggregates requests rates from metrics fetched from Prometheus, and stores the result in the health map.
func fillWorkloadRequestRates(allHealth models.NamespaceWorkloadHealth, rates model.Vector) {
	lblDest := model.LabelName("destination_workload")
	lblSrc := model.LabelName("source_workload")
	for _, sample := range rates {
		name := string(sample.Metric[lblDest])
		if health, ok := allHealth[name]; ok {
			health.Requests.AggregateInbound(sample)
		}
		name = string(sample.Metric[lblSrc])
		if health, ok := allHealth[name]; ok {
			health.Requests.AggregateOutbound(sample)
		}
	}
}

func (in *HealthService) getServiceRequestsHealth(namespace, service, rateInterval string, queryTime time.Time) (models.RequestHealth, error) {
	rqHealth := models.NewEmptyRequestHealth()
	inbound, err := in.prom.GetServiceRequestRates(namespace, service, rateInterval, queryTime)
	for _, sample := range inbound {
		rqHealth.AggregateInbound(sample)
	}
	return rqHealth, err
}

func (in *HealthService) getAppRequestsHealth(namespace, app, rateInterval string, queryTime time.Time) (models.RequestHealth, error) {
	rqHealth := models.NewEmptyRequestHealth()
	inbound, outbound, err := in.prom.GetAppRequestRates(namespace, app, rateInterval, queryTime)
	for _, sample := range inbound {
		rqHealth.AggregateInbound(sample)
	}
	for _, sample := range outbound {
		rqHealth.AggregateOutbound(sample)
	}
	return rqHealth, err
}

func (in *HealthService) getWorkloadRequestsHealth(namespace, workload, rateInterval string, queryTime time.Time) (models.RequestHealth, error) {
	rqHealth := models.NewEmptyRequestHealth()
	inbound, outbound, err := in.prom.GetWorkloadRequestRates(namespace, workload, rateInterval, queryTime)
	for _, sample := range inbound {
		rqHealth.AggregateInbound(sample)
	}
	for _, sample := range outbound {
		rqHealth.AggregateOutbound(sample)
	}
	return rqHealth, err
}

func castWorkloadStatuses(ws models.Workloads) []models.WorkloadStatus {
	statuses := make([]models.WorkloadStatus, 0)
	for _, w := range ws {
		status := models.WorkloadStatus{
			Name:              w.Name,
			Replicas:          w.Replicas,
			AvailableReplicas: w.AvailableReplicas}
		statuses = append(statuses, status)

	}
	return statuses
}
