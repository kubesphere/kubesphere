package apis

import monitoringdashboardv1alpha2 "kubesphere.io/monitoring-dashboard/api/v1alpha2"

func init() {
	AddToSchemes = append(AddToSchemes, monitoringdashboardv1alpha2.SchemeBuilder.AddToScheme)
}
