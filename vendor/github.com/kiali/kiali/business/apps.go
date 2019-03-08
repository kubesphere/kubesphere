package business

import (
	"fmt"
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus"
	"github.com/kiali/kiali/prometheus/internalmetrics"
)

// AppService deals with fetching Workloads group by "app" label, which will be identified as an "application"
type AppService struct {
	prom prometheus.ClientInterface
	k8s  kubernetes.IstioClientInterface
}

// Temporal map of Workloads group by app label
type appsWorkload map[string][]*models.Workload

// Helper method to build a map of workloads for a given labelSelector
func (in *AppService) fetchWorkloadsPerApp(namespace, labelSelector string) (appsWorkload, error) {
	cfg := config.Get()

	ws, err := fetchWorkloads(in.k8s, namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	apps := make(appsWorkload)
	for _, w := range ws {
		if appLabel, ok := w.Labels[cfg.IstioLabels.AppLabelName]; ok {
			apps[appLabel] = append(apps[appLabel], w)
		}
	}
	return apps, nil
}

// GetAppList is the API handler to fetch the list of applications in a given namespace
func (in *AppService) GetAppList(namespace string) (models.AppList, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "AppService", "GetAppList")
	defer promtimer.ObserveNow(&err)

	appList := &models.AppList{
		Namespace: models.Namespace{Name: namespace},
		Apps:      []models.AppListItem{},
	}
	apps, err := fetchNamespaceApps(in.k8s, namespace, "")
	if err != nil {
		return *appList, err
	}

	for keyApp, valueApp := range apps {
		appItem := &models.AppListItem{Name: keyApp}
		appItem.IstioSidecar = false
		if len(valueApp.Workloads) > 0 {
			appItem.IstioSidecar = true
		}
		for _, w := range valueApp.Workloads {
			appItem.IstioSidecar = appItem.IstioSidecar && w.Pods.HasIstioSideCar()
		}
		(*appList).Apps = append((*appList).Apps, *appItem)
	}

	return *appList, nil
}

// GetApp is the API handler to fetch the details for a given namespace and app name
func (in *AppService) GetApp(namespace string, appName string) (models.App, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "AppService", "GetApp")
	defer promtimer.ObserveNow(&err)

	appInstance := &models.App{Namespace: models.Namespace{Name: namespace}, Name: appName}
	namespaceApps, err := fetchNamespaceApps(in.k8s, namespace, appName)
	if err != nil {
		return *appInstance, err
	}

	var appDetails *appDetails
	var ok bool
	// Send a NewNotFound if the app is not found in the deployment list, instead to send an empty result
	if appDetails, ok = namespaceApps[appName]; !ok {
		return *appInstance, kubernetes.NewNotFound(appName, "Kiali", "App")
	}

	(*appInstance).Workloads = make([]models.WorkloadItem, len(appDetails.Workloads))
	for i, wkd := range appDetails.Workloads {
		wkdSvc := &models.WorkloadItem{WorkloadName: wkd.Name}
		wkdSvc.IstioSidecar = wkd.Pods.HasIstioSideCar()
		(*appInstance).Workloads[i] = *wkdSvc
	}

	(*appInstance).ServiceNames = make([]string, len(appDetails.Services))
	for i, svc := range appDetails.Services {
		(*appInstance).ServiceNames[i] = svc.Name
	}

	in.fillCustomDashboardRefs(namespace, appInstance, appDetails)

	return *appInstance, nil
}

// AppDetails holds Services and Workloads having the same "app" label
type appDetails struct {
	app       string
	Services  []v1.Service
	Workloads models.Workloads
}

// NamespaceApps is a map of app_name x AppDetails
type namespaceApps = map[string]*appDetails

func castAppDetails(services []v1.Service, ws models.Workloads) namespaceApps {
	allEntities := make(namespaceApps)
	appLabel := config.Get().IstioLabels.AppLabelName
	for _, service := range services {
		if app, ok := service.Spec.Selector[appLabel]; ok {
			if appEntities, ok := allEntities[app]; ok {
				appEntities.Services = append(appEntities.Services, service)
			} else {
				allEntities[app] = &appDetails{
					app:      app,
					Services: []v1.Service{service},
				}
			}
		}
	}
	for _, w := range ws {
		if app, ok := w.Labels[appLabel]; ok {
			if appEntities, ok := allEntities[app]; ok {
				appEntities.Workloads = append(appEntities.Workloads, w)
			} else {
				allEntities[app] = &appDetails{
					app:       app,
					Workloads: models.Workloads{w},
				}
			}
		}
	}
	return allEntities
}

// Helper method to fetch all applications for a given namespace.
// Optionally if appName parameter is provided, it filters apps for that name.
// Return an error on any problem.
func fetchNamespaceApps(k8s kubernetes.IstioClientInterface, namespace string, appName string) (namespaceApps, error) {
	var services []v1.Service
	var ws models.Workloads
	cfg := config.Get()

	labelSelector := cfg.IstioLabels.AppLabelName
	if appName != "" {
		labelSelector = fmt.Sprintf("%s=%s", cfg.IstioLabels.AppLabelName, appName)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	errChan := make(chan error, 2)

	go func() {
		defer wg.Done()
		var err error
		services, err = k8s.GetServices(namespace, nil)
		if appName != "" {
			selector := labels.Set(map[string]string{cfg.IstioLabels.AppLabelName: appName}).AsSelector()
			services = kubernetes.FilterServicesForSelector(selector, services)
		}
		if err != nil {
			log.Errorf("Error fetching Services per namespace %s: %s", namespace, err)
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		ws, err = fetchWorkloads(k8s, namespace, labelSelector)
		if err != nil {
			log.Errorf("Error fetching Workload per namespace %s: %s", namespace, err)
			errChan <- err
		}
	}()

	wg.Wait()
	if len(errChan) != 0 {
		err := <-errChan
		return nil, err
	}

	return castAppDetails(services, ws), nil
}

// fillCustomDashboardRefs finds all dashboard IDs and Titles associated to this app and add them to the model
func (in *AppService) fillCustomDashboardRefs(namespace string, app *models.App, details *appDetails) {
	allPods := models.Pods{}
	for _, workload := range details.Workloads {
		allPods = append(allPods, workload.Pods...)
	}
	uniqueRefsList := getUniqueRuntimes(allPods)
	mon, err := kubernetes.NewKialiMonitoringClient()
	if err != nil {
		// Do not fail the whole query, just log & return
		log.Error("Cannot initialize Kiali Monitoring Client")
		return
	}
	dash := NewDashboardsService(mon, in.prom)
	app.Runtimes = dash.buildRuntimesList(namespace, uniqueRefsList)
}
