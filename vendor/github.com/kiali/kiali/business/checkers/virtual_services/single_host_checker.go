package virtual_services

import (
	"reflect"
	"strings"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
)

type SingleHostChecker struct {
	Namespace       string
	VirtualServices []kubernetes.IstioObject
}

type Host struct {
	Service   string
	Namespace string
	Cluster   string
}

func (in SingleHostChecker) Check() models.IstioValidations {
	hostCounter := make(map[string]map[string]map[string][]*kubernetes.IstioObject)
	validations := models.IstioValidations{}

	for _, vs := range in.VirtualServices {
		for _, host := range getHost(vs) {
			storeHost(hostCounter, vs, host)
		}
	}

	for _, clusterCounter := range hostCounter {
		for _, namespaceCounter := range clusterCounter {
			isNamespaceWildcard := len(namespaceCounter["*"]) > 0
			for _, serviceCounter := range namespaceCounter {
				targetSameHost := len(serviceCounter) > 1
				otherServiceHosts := len(namespaceCounter) > 1
				for _, virtualService := range serviceCounter {
					// Marking virtualService as invalid if:
					// - there is more than one virtual service per a host
					// - there is one virtual service with wildcard and there are other virtual services pointing
					//   a host for that namespace
					if targetSameHost || isNamespaceWildcard && otherServiceHosts {
						if !hasGateways(virtualService) {
							multipleVirtualServiceCheck(*virtualService, validations)
						}
					}
				}
			}
		}
	}

	return validations
}

func multipleVirtualServiceCheck(virtualService kubernetes.IstioObject, validations models.IstioValidations) {
	virtualServiceName := virtualService.GetObjectMeta().Name
	key := models.IstioValidationKey{Name: virtualServiceName, ObjectType: "virtualservice"}
	checks := models.Build("virtualservices.singlehost", "spec/hosts")
	rrValidation := &models.IstioValidation{
		Name:       virtualServiceName,
		ObjectType: "virtualservice",
		Valid:      true,
		Checks: []*models.IstioCheck{
			&checks,
		},
	}

	validations.MergeValidations(models.IstioValidations{key: rrValidation})
}

func storeHost(hostCounter map[string]map[string]map[string][]*kubernetes.IstioObject, vs kubernetes.IstioObject, host Host) {
	vsList := []*kubernetes.IstioObject{&vs}

	if hostCounter[host.Cluster] == nil {
		hostCounter[host.Cluster] = map[string]map[string][]*kubernetes.IstioObject{
			host.Namespace: {
				host.Service: vsList,
			},
		}
	} else if hostCounter[host.Cluster][host.Namespace] == nil {
		hostCounter[host.Cluster][host.Namespace] = map[string][]*kubernetes.IstioObject{
			host.Service: vsList,
		}
	} else if _, ok := hostCounter[host.Cluster][host.Namespace][host.Service]; !ok {
		hostCounter[host.Cluster][host.Namespace][host.Service] = vsList
	} else {
		hostCounter[host.Cluster][host.Namespace][host.Service] = append(hostCounter[host.Cluster][host.Namespace][host.Service], &vs)

	}
}

func getHost(virtualService kubernetes.IstioObject) []Host {
	hosts := virtualService.GetSpec()["hosts"]
	if hosts == nil {
		return []Host{}
	}

	slice := reflect.ValueOf(hosts)
	if slice.Kind() != reflect.Slice {
		return []Host{}
	}

	targetHosts := make([]Host, 0, slice.Len())

	for hostIdx := 0; hostIdx < slice.Len(); hostIdx++ {
		hostName, ok := slice.Index(hostIdx).Interface().(string)
		if !ok {
			continue
		}

		targetHosts = append(targetHosts, formatHostForSearch(hostName, virtualService.GetObjectMeta().Namespace))
	}

	return targetHosts
}

// Convert host to Host struct for searching
// e.g. reviews -> reviews, virtualService.Namespace, svc.cluster.local
// e.g. reviews.bookinfo.svc.cluster.local -> reviews, bookinfo, svc.cluster.local
// e.g. *.bookinfo.svc.cluster.local -> *, bookinfo, svc.cluster.local
// e.g. * -> *, *, *
func formatHostForSearch(hostName, virtualServiceNamespace string) Host {
	domainParts := strings.Split(hostName, ".")
	host := Host{}

	host.Service = domainParts[0]
	if len(domainParts) > 1 {
		host.Namespace = domainParts[1]

		if len(domainParts) > 2 {
			host.Cluster = strings.Join(domainParts[2:], ".")
		}
	} else if host.Service != "*" {
		host.Namespace = virtualServiceNamespace
		host.Cluster = "svc.cluster.local"
	} else if host.Service == "*" {
		host.Namespace = "*"
		host.Cluster = "*"
	}

	return host
}

func hasGateways(virtualService *kubernetes.IstioObject) bool {
	if gateways, ok := (*virtualService).GetSpec()["gateways"]; ok {
		vsGateways, ok := (gateways).([]interface{})
		return ok && vsGateways != nil && len(vsGateways) > 0
	}
	return false
}
