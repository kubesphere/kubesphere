package apis

import monitoringdashboardv1alpha1 "kubesphere.io/monitoring-dashboard/api/v1alpha1"

func init() {
	AddToSchemes = append(AddToSchemes, monitoringdashboardv1alpha1.SchemeBuilder.AddToScheme)
}
