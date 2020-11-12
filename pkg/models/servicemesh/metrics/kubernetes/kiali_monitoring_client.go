package kubernetes

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

// KialiMonitoringInterface for mocks (only mocked function are necessary here)
type KialiMonitoringInterface interface {
	GetDashboard(namespace string, name string) (*MonitoringDashboard, error)
}

// KialiMonitoringClient is the client struct for Kiali Monitoring API over Kubernetes
// API to get MonitoringDashboards
type KialiMonitoringClient struct {
	KialiMonitoringInterface
	client *rest.RESTClient
}

// NewKialiMonitoringClient creates a new client able to fetch Kiali Monitoring API.
func NewKialiMonitoringClient() (*KialiMonitoringClient, error) {
	config, err := ConfigClient()
	if err != nil {
		return nil, err
	}

	types := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			return nil
		})

	err = schemeBuilder.AddToScheme(types)
	if err != nil {
		return nil, err
	}

	client, err := newClientForAPI(config, kialiMonitoringGroupVersion, types)
	return &KialiMonitoringClient{
		client: client,
	}, err
}

// GetDashboard returns a MonitoringDashboard for the given name
func (in *KialiMonitoringClient) GetDashboard(namespace string, name string) (*MonitoringDashboard, error) {
	result, err := in.client.Get().Namespace(namespace).Resource("monitoringdashboards").SubResource(name).Do().Raw()
	if err != nil {
		return nil, err
	}
	var dashboard MonitoringDashboard
	err = json.Unmarshal(result, &dashboard)
	return &dashboard, err
}
