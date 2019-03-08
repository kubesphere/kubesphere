package business

import (
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/prometheus"
)

// Layer is a container for fast access to inner services
type Layer struct {
	Svc         SvcService
	Health      HealthService
	Validations IstioValidationsService
	IstioConfig IstioConfigService
	Workload    WorkloadService
	App         AppService
	Namespace   NamespaceService
	k8s         kubernetes.IstioClientInterface
}

// Global business.Layer; currently only used for tests to inject mocks,
//	whereas production code recreates services in a stateless way
var layer *Layer

// Get the business.Layer, create it if necessary
func Get() (*Layer, error) {
	if layer != nil {
		return layer, nil
	}
	k8s, err := kubernetes.NewClient()
	if err != nil {
		return nil, err
	}
	prom, err := prometheus.NewClient()
	if err != nil {
		return nil, err
	}
	// Business needs to maintain a minimal state as kubernetes package will maintain a cache
	SetWithBackends(k8s, prom)
	return layer, nil
}

// SetWithBackends creates all services with injected clients to external APIs
func SetWithBackends(k8s kubernetes.IstioClientInterface, prom prometheus.ClientInterface) *Layer {
	layer = NewWithBackends(k8s, prom)
	return layer
}

// NewWithBackends creates the business layer using the passed k8s and prom clients
func NewWithBackends(k8s kubernetes.IstioClientInterface, prom prometheus.ClientInterface) *Layer {
	temporaryLayer := &Layer{}
	temporaryLayer.Health = HealthService{prom: prom, k8s: k8s}
	temporaryLayer.Svc = SvcService{prom: prom, k8s: k8s, businessLayer: temporaryLayer}
	temporaryLayer.IstioConfig = IstioConfigService{k8s: k8s}
	temporaryLayer.Workload = WorkloadService{k8s: k8s, prom: prom, businessLayer: temporaryLayer}
	temporaryLayer.Validations = IstioValidationsService{k8s: k8s, businessLayer: temporaryLayer}
	temporaryLayer.App = AppService{prom: prom, k8s: k8s}
	temporaryLayer.Namespace = NewNamespaceService(k8s)
	temporaryLayer.k8s = k8s
	return temporaryLayer
}

func (in *Layer) Stop() {
	if in.k8s != nil {
		in.k8s.Stop()
	}
}
