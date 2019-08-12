package metrics

import (
	"encoding/json"
	"fmt"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/pkg/textparse"
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"net/http"
	"net/url"
	"strconv"
)

const (
	ServiceMonitorLabel      = "k8s-app"
	ServiceMonitorAnnotation = "kubesphere.io/enable"
)

var (
	clientset = k8s.MonitClient()
)

func CheckConnectionAndDataFormat(ns string, smon ServiceMonitorConfig) error {

	var urls []url.URL

	svcLister := informers.SharedInformerFactory().Core().V1().Services().Lister()
	svc, err := svcLister.Services(ns).Get(smon.Service)
	if err != nil {
		return err
	}

	for _, ep := range smon.Endpoints {
		for _, p := range svc.Spec.Ports {
			if ep.Port == p.Name {
				urls = append(urls, url.URL{
					Scheme: "http",
					Host:   fmt.Sprintf("%s.%s.svc:%d", svc.Name, svc.Namespace, p.Port),
					Path:   ep.Path})
			}
		}
	}
	if len(urls) != len(smon.Endpoints) || len(urls) == 0 {
		return fmt.Errorf("Invalid ports")
	}

	// Try to connect each url, and check Prometheus data format
	for _, u := range urls {
		resp, err := http.DefaultClient.Get(u.String())
		if err != nil {
			return err
		}

		contentType := resp.Header.Get("Content-Type")
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()

		err = isPrometheusFormat(b, contentType)
		if err != nil {
			return err
		}
	}

	return err
}

func isPrometheusFormat(b []byte, contentType string) (err error) {
	p := textparse.New(b, contentType)
	for {
		if _, err = p.Next(); err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
	}
}

func ApplyServiceMonitor(ns string, cfg ServiceMonitorConfig) error {

	smonName := makeServiceMonitorName(ns, cfg.Service)
	obj := monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      smonName,
			Namespace: ns,
			Labels: map[string]string{
				ServiceMonitorLabel: smonName,
			},
			Annotations: map[string]string{
				ServiceMonitorAnnotation: strconv.FormatBool(cfg.Enable),
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: func() (endpoints []monitoringv1.Endpoint) {
				for _, ep := range cfg.Endpoints {
					endpoints = append(endpoints, monitoringv1.Endpoint{
						Port:          ep.Port,
						Path:          ep.Path,
						Interval:      cfg.Interval,
						ScrapeTimeout: cfg.ScrapeTimeout,
					})
				}
				return
			}(),
			Selector: metav1.LabelSelector{
				MatchLabels: func() map[string]string {
					svcLister := informers.SharedInformerFactory().Core().V1().Services().Lister()
					svc, err := svcLister.Services(ns).Get(cfg.Service)
					if err != nil {
						return nil
					}
					return svc.Labels
				}(),
			},
		},
	}

	_, err := clientset.MonitoringV1().ServiceMonitors(ns).Get(smonName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = clientset.MonitoringV1().ServiceMonitors(ns).Create(&obj)
			if err != nil {
				return err
			}
		}
		return err
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = clientset.MonitoringV1().ServiceMonitors(ns).Patch(smonName, types.MergePatchType, b)
	if err != nil {
		return err
	}

	return err
}

func GetServiceMonitorConfig(ns string, svc string) (cfg ServiceMonitorConfig, err error) {

	obj, err := clientset.MonitoringV1().ServiceMonitors(ns).Get(makeServiceMonitorName(ns, svc), metav1.GetOptions{})
	if err != nil {
		return
	}

	// endpoints are required
	if len(obj.Spec.Endpoints) == 0 {
		err = fmt.Errorf("Invalid endpoints.")
		return
	}

	cfg = ServiceMonitorConfig{
		Service:       svc,
		Interval:      obj.Spec.Endpoints[0].Interval,
		ScrapeTimeout: obj.Spec.Endpoints[0].ScrapeTimeout,
		Endpoints: func() (eps []ServiceMonitorEndpoint) {
			for _, ep := range obj.Spec.Endpoints {
				eps = append(eps, ServiceMonitorEndpoint{
					Port: ep.Port,
					Path: ep.Path,
				})
			}
			return
		}(),
		Enable: func() bool {
			for k, v := range obj.Annotations {
				if k == ServiceMonitorAnnotation && v == "true" {
					return true
				}
			}
			return false
		}(),
	}

	return
}

func DeleteServiceMonitor(ns string, svc string) error {
	return clientset.MonitoringV1().ServiceMonitors(ns).Delete(makeServiceMonitorName(ns, svc), &metav1.DeleteOptions{})
}

func makeServiceMonitorName(ns string, svc string) string {
	return fmt.Sprintf("%s.%s", ns, svc)
}
