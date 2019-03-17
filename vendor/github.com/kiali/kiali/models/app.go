package models

type AppList struct {
	// Namespace where the apps live in
	// required: true
	// example: bookinfo
	Namespace Namespace `json:"namespace"`

	// Applications for a given namespace
	// required: true
	Apps []AppListItem `json:"applications"`
}

// AppListItem has the necessary information to display the console app list
type AppListItem struct {
	// Name of the application
	// required: true
	// example: reviews
	Name string `json:"name"`

	// Define if all Pods related to the Workloads of this app has an IstioSidecar deployed
	// required: true
	// example: true
	IstioSidecar bool `json:"istioSidecar"`
}

type WorkloadItem struct {
	// Name of a workload member of an application
	// required: true
	// example: reviews-v1
	WorkloadName string `json:"workloadName"`

	// Define if all Pods related to the Workload has an IstioSidecar deployed
	// required: true
	// example: true
	IstioSidecar bool `json:"istioSidecar"`
}

type App struct {
	// Namespace where the app lives in
	// required: true
	// example: bookinfo
	Namespace Namespace `json:"namespace"`

	// Name of the application
	// required: true
	// example: reviews
	Name string `json:"name"`

	// Workloads for a given application
	// required: true
	Workloads []WorkloadItem `json:"workloads"`

	// List of service names linked with an application
	// required: true
	ServiceNames []string `json:"serviceNames"`

	// Runtimes and associated dashboards
	Runtimes []Runtime `json:"runtimes"`
}
