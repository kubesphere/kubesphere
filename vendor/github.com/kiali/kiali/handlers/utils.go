package handlers

import (
	"net/http"
	"net/url"

	"k8s.io/api/core/v1"

	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus"
)

type promClientSupplier func() (*prometheus.Client, error)
type k8sClientSupplier func() (kubernetes.IstioClientInterface, error)

var defaultPromClientSupplier = prometheus.NewClient
var defaultK8SClientSupplier = func() (kubernetes.IstioClientInterface, error) {
	return kubernetes.NewClient()
}

func getService(namespace string, service string) (*v1.ServiceSpec, error) {
	client, err := kubernetes.NewClient()
	if err != nil {
		return nil, err
	}
	svc, err := client.GetService(namespace, service)
	if err != nil {
		return nil, err
	}
	return &svc.Spec, nil
}

func validateURL(serviceURL string) (*url.URL, error) {
	return url.ParseRequestURI(serviceURL)
}

func checkNamespaceAccess(w http.ResponseWriter, k8s kubernetes.IstioClientInterface, prom prometheus.ClientInterface, namespace string) *models.Namespace {
	layer := business.NewWithBackends(k8s, prom)

	if nsInfo, err := layer.Namespace.GetNamespace(namespace); err != nil {
		RespondWithError(w, http.StatusForbidden, "Cannot access namespace data: "+err.Error())
		return nil
	} else {
		return nsInfo
	}
}

func initClientsForMetrics(w http.ResponseWriter, promSupplier promClientSupplier, k8sSupplier k8sClientSupplier, namespace string) (*prometheus.Client, kubernetes.IstioClientInterface, *models.Namespace) {
	k8s, err := k8sSupplier()
	if err != nil {
		log.Error(err)
		RespondWithError(w, http.StatusServiceUnavailable, "Kubernetes client error: "+err.Error())
		return nil, nil, nil
	}
	prom, err := promSupplier()
	if err != nil {
		log.Error(err)
		RespondWithError(w, http.StatusServiceUnavailable, "Prometheus client error: "+err.Error())
		return nil, nil, nil
	}

	nsInfo := checkNamespaceAccess(w, k8s, prom, namespace)
	if nsInfo == nil {
		return nil, nil, nil
	}
	return prom, k8s, nsInfo
}
