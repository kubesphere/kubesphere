package status

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	kversion "k8s.io/apimachinery/pkg/version"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/log"
)

type externalService func() (*ExternalServiceInfo, error)

var (
	// Example Maistra product version is:
	//   redhat@redhat-docker.io/maistra-0.1.0-1-3a136c90ec5e308f236e0d7ebb5c4c5e405217f4-unknown
	// Example Maistra upstream project version is:
	//   redhat@redhat-pulp.abc.xyz.redhat.com:8888/openshift-istio-tech-preview-0.1.0-1-3a136c90ec5e308f236e0d7ebb5c4c5e405217f4-Custom
	// Example Istio snapshot version is:
	//   root@f72e3d3ef3c2-docker.io/istio-release-1.0-20180927-21-10-cbe9c05c470ec1924f7bcf02334b183e7e6175cb-Clean
	maistraProductVersionExpr = regexp.MustCompile("maistra-([0-9]+\\.[0-9]+\\.[0-9]+)")
	maistraProjectVersionExpr = regexp.MustCompile("openshift-istio.*-([0-9]+\\.[0-9]+\\.[0-9]+)")
	istioVersionExpr          = regexp.MustCompile("([0-9]+\\.[0-9]+\\.[0-9]+)")
	istioSnapshotVersionExpr  = regexp.MustCompile("istio-release-([0-9]+\\.[0-9]+)(-[0-9]{8})")
)

func getVersions() {
	components := []externalService{
		istioVersion,
		prometheusVersion,
		kubernetesVersion,
	}

	if config.Get().ExternalServices.Jaeger.URL != "" {
		components = append(components, jaegerVersion)
	}

	if config.Get().ExternalServices.Grafana.URL != "" {
		components = append(components, grafanaVersion)
	}

	for _, comp := range components {
		getVersionComponent(comp)
	}
}

func getVersionComponent(serviceComponent externalService) {
	componentInfo, err := serviceComponent()
	if err == nil {
		info.ExternalServices = append(info.ExternalServices, *componentInfo)
	}
}

func validateVersion(istioReq string, installedVersion string) bool {
	reqWords := strings.Split(istioReq, " ")
	requirementV, errReqV := version.NewVersion(reqWords[1])
	installedV, errInsV := version.NewVersion(installedVersion)
	if errReqV != nil || errInsV != nil {
		return false
	}
	switch operator := reqWords[0]; operator {
	case "==":
		return installedV.Equal(requirementV)
	case ">=":
		return installedV.GreaterThan(requirementV) || installedV.Equal(requirementV)
	case ">":
		return installedV.GreaterThan(requirementV)
	case "<=":
		return installedV.LessThan(requirementV) || installedV.Equal(requirementV)
	case "<":
		return installedV.LessThan(requirementV)
	}
	return false
}

func istioVersion() (*ExternalServiceInfo, error) {
	var (
		body    []byte
		err     error
		product *ExternalServiceInfo
		resp    *http.Response
	)

	istioConfig := config.Get().ExternalServices.Istio
	resp, err = http.Get(istioConfig.UrlServiceVersion)
	if err == nil {
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err == nil {
			rawVersion := string(body)
			product, err = parseIstioRawVersion(rawVersion)
			return product, err
		}
	}
	return nil, err
}

func parseIstioRawVersion(rawVersion string) (*ExternalServiceInfo, error) {
	product := ExternalServiceInfo{Name: "Unknown", Version: "Unknown"}

	// First see if we detect Maistra (either product or upstream project).
	// If it is not Maistra, see if it is upstream Istio (either a release or snapshot).
	// If it is neither then it is some unknown Istio implementation that we do not support.

	maistraVersionStringArr := maistraProductVersionExpr.FindStringSubmatch(rawVersion)
	if maistraVersionStringArr != nil {
		log.Debugf("Detected Maistra product version [%v]", rawVersion)
		if len(maistraVersionStringArr) > 1 {
			product.Name = "Maistra"
			product.Version = maistraVersionStringArr[1] // get regex group #1 ,which is the "#.#.#" version string
			if !validateVersion(config.MaistraVersionSupported, product.Version) {
				info.WarningMessages = append(info.WarningMessages, "Maistra version "+product.Version+" is not supported, the version should be "+config.MaistraVersionSupported)
			}

			// we know this is Maistra - either a supported or unsupported version - return now
			return &product, nil
		}
	}

	maistraVersionStringArr = maistraProjectVersionExpr.FindStringSubmatch(rawVersion)
	if maistraVersionStringArr != nil {
		log.Debugf("Detected Maistra project version [%v]", rawVersion)
		if len(maistraVersionStringArr) > 1 {
			product.Name = "Maistra Project"
			product.Version = maistraVersionStringArr[1] // get regex group #1 ,which is the "#.#.#" version string
			if !validateVersion(config.MaistraVersionSupported, product.Version) {
				info.WarningMessages = append(info.WarningMessages, "Maistra project version "+product.Version+" is not supported, the version should be "+config.MaistraVersionSupported)
			}

			// we know this is Maistra - either a supported or unsupported version - return now
			return &product, nil
		}
	}

	// see if it is a released version of Istio
	istioVersionStringArr := istioVersionExpr.FindStringSubmatch(rawVersion)
	if istioVersionStringArr != nil {
		log.Debugf("Detected Istio version [%v]", rawVersion)
		if len(istioVersionStringArr) > 1 {
			product.Name = "Istio"
			product.Version = istioVersionStringArr[1] // get regex group #1 ,which is the "#.#.#" version string
			if !validateVersion(config.IstioVersionSupported, product.Version) {
				info.WarningMessages = append(info.WarningMessages, "Istio version "+product.Version+" is not supported, the version should be "+config.IstioVersionSupported)
			}
			// we know this is Istio upstream - either a supported or unsupported version - return now
			return &product, nil
		}
	}

	// see if it is a snapshot version of Istio
	istioVersionStringArr = istioSnapshotVersionExpr.FindStringSubmatch(rawVersion)
	if istioVersionStringArr != nil {
		log.Debugf("Detected Istio snapshot version [%v]", rawVersion)
		if len(istioVersionStringArr) > 2 {
			product.Name = "Istio Snapshot"
			majorMinor := istioVersionStringArr[1]  // regex group #1 is the "#.#" version numbers
			snapshotStr := istioVersionStringArr[2] // regex group #2 is the date/time stamp
			product.Version = majorMinor + snapshotStr
			if !validateVersion(config.IstioVersionSupported, majorMinor) {
				info.WarningMessages = append(info.WarningMessages, "Istio snapshot version "+product.Version+" is not supported, the version should be "+config.IstioVersionSupported)
			}
			// we know this is Istio upstream - either a supported or unsupported version - return now
			return &product, nil
		}
	}

	log.Debugf("Detected unknown Istio implementation version [%v]", rawVersion)
	product.Name = "Unknown Istio Implementation"
	product.Version = rawVersion
	info.WarningMessages = append(info.WarningMessages, "Unknown Istio implementation version "+product.Version+" is not recognized, thus not supported.")
	return &product, nil
}

type p8sResponseVersion struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
}

func jaegerVersion() (*ExternalServiceInfo, error) {
	product := ExternalServiceInfo{}
	product.Name = "Jaeger"
	product.Url = config.Get().ExternalServices.Jaeger.URL

	return &product, nil
}

func grafanaVersion() (*ExternalServiceInfo, error) {
	product := ExternalServiceInfo{}
	product.Name = "Grafana"
	product.Url = config.Get().ExternalServices.Grafana.URL

	return &product, nil
}

func prometheusVersion() (*ExternalServiceInfo, error) {
	product := ExternalServiceInfo{}
	prometheusV := new(p8sResponseVersion)
	prometheusUrl := config.Get().ExternalServices.PrometheusServiceURL
	resp, err := http.Get(prometheusUrl + "/version")
	if err == nil {
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&prometheusV)
		if err == nil {
			product.Name = "Prometheus"
			product.Version = prometheusV.Version
			return &product, nil
		}
	}
	return nil, err
}

func kubernetesVersion() (*ExternalServiceInfo, error) {
	var (
		err           error
		k8sConfig     *rest.Config
		k8s           *kube.Clientset
		serverVersion *kversion.Info
	)

	product := ExternalServiceInfo{}
	k8sConfig, err = kubernetes.ConfigClient()
	if err == nil {
		k8sConfig.QPS = config.Get().KubernetesConfig.QPS
		k8sConfig.Burst = config.Get().KubernetesConfig.Burst
		k8s, err = kube.NewForConfig(k8sConfig)
		if err == nil {
			serverVersion, err = k8s.Discovery().ServerVersion()
			if err == nil {
				product.Name = "Kubernetes"
				product.Version = serverVersion.GitVersion
				return &product, nil
			}
		}
	}
	return nil, err
}
