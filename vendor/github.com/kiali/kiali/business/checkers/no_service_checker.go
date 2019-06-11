package checkers

import (
	"github.com/kiali/kiali/business/checkers/destinationrules"
	"github.com/kiali/kiali/business/checkers/virtual_services"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
	v1 "k8s.io/api/core/v1"
)

type NoServiceChecker struct {
	Namespace            string
	IstioDetails         *kubernetes.IstioDetails
	Services             []v1.Service
	WorkloadList         models.WorkloadList
	GatewaysPerNamespace [][]kubernetes.IstioObject
}

func (in NoServiceChecker) Check() models.IstioValidations {
	validations := models.IstioValidations{}

	if in.IstioDetails == nil || in.Services == nil {
		return validations
	}

	log.Infof("ServiceEntries: %v\n", in.IstioDetails.ServiceEntries)

	serviceNames := getServiceNames(in.Services)
	serviceHosts := kubernetes.ServiceEntryHostnames(in.IstioDetails.ServiceEntries)
	gatewayNames := kubernetes.GatewayNames(in.GatewaysPerNamespace)

	for _, virtualService := range in.IstioDetails.VirtualServices {
		validations.MergeValidations(runVirtualServiceCheck(virtualService, in.Namespace, serviceNames, serviceHosts))
		validations.MergeValidations(runGatewayCheck(virtualService, gatewayNames))
	}
	for _, destinationRule := range in.IstioDetails.DestinationRules {
		validations.MergeValidations(runDestinationRuleCheck(destinationRule, in.Namespace, in.WorkloadList, serviceNames, serviceHosts))
	}

	return validations
}

func runVirtualServiceCheck(virtualService kubernetes.IstioObject, namespace string, serviceNames []string, serviceHosts map[string][]string) models.IstioValidations {
	result, valid := virtual_services.NoHostChecker{
		Namespace:         namespace,
		ServiceNames:      serviceNames,
		VirtualService:    virtualService,
		ServiceEntryHosts: serviceHosts,
	}.Check()

	istioObjectName := virtualService.GetObjectMeta().Name
	key := models.IstioValidationKey{ObjectType: "virtualservice", Name: istioObjectName}
	vsvalidations := models.IstioValidations{}
	vsvalidations[key] = &models.IstioValidation{
		Name:       istioObjectName,
		ObjectType: "virtualservice",
		Checks:     result,
		Valid:      valid,
	}
	return vsvalidations
}

func runGatewayCheck(virtualService kubernetes.IstioObject, gatewayNames map[string]struct{}) models.IstioValidations {
	result, valid := virtual_services.NoGatewayChecker{
		VirtualService: virtualService,
		GatewayNames:   gatewayNames,
	}.Check()

	istioObjectName := virtualService.GetObjectMeta().Name
	key := models.IstioValidationKey{ObjectType: "virtualservice", Name: istioObjectName}
	vsvalidations := models.IstioValidations{}
	vsvalidations[key] = &models.IstioValidation{
		Name:       istioObjectName,
		ObjectType: "virtualservice",
		Checks:     result,
		Valid:      valid,
	}
	return vsvalidations
}

func runDestinationRuleCheck(destinationRule kubernetes.IstioObject, namespace string, workloads models.WorkloadList, serviceNames []string, serviceHosts map[string][]string) models.IstioValidations {
	result, valid := destinationrules.NoDestinationChecker{
		Namespace:       namespace,
		WorkloadList:    workloads,
		DestinationRule: destinationRule,
		ServiceNames:    serviceNames,
		ServiceEntries:  serviceHosts,
	}.Check()

	istioObjectName := destinationRule.GetObjectMeta().Name
	key := models.IstioValidationKey{ObjectType: "destinationrule", Name: istioObjectName}
	drvalidations := models.IstioValidations{}
	drvalidations[key] = &models.IstioValidation{
		Name:       istioObjectName,
		ObjectType: "destinationrule",
		Checks:     result,
		Valid:      valid,
	}
	return drvalidations
}

func getServiceNames(services []v1.Service) []string {
	serviceNames := make([]string, 0)
	for _, item := range services {
		serviceNames = append(serviceNames, item.Name)
	}
	return serviceNames
}
