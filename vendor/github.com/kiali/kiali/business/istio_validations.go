package business

import (
	"fmt"
	"sync"

	"github.com/kiali/kiali/business/checkers"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus/internalmetrics"

	v1 "k8s.io/api/core/v1"
)

type IstioValidationsService struct {
	k8s           kubernetes.IstioClientInterface
	businessLayer *Layer
}

type ObjectChecker interface {
	Check() models.IstioValidations
}

// GetValidations returns an IstioValidations object with all the checks found when running
// all the enabled checkers. If service is "" then the whole namespace is validated.
func (in *IstioValidationsService) GetValidations(namespace, service string) (models.IstioValidations, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "IstioValidationsService", "GetValidations")
	defer promtimer.ObserveNow(&err)

	// Ensure the service or namespace exists.. do we need to block with this?
	if service != "" {
		if _, err := in.k8s.GetService(namespace, service); err != nil {
			return nil, err
		}
	} else {
		if _, err := in.k8s.GetNamespace(namespace); err != nil {
			return nil, err
		}
	}

	wg := sync.WaitGroup{}
	errChan := make(chan error, 1)

	var istioDetails kubernetes.IstioDetails
	var services []v1.Service
	var workloads models.WorkloadList
	var gatewaysPerNamespace [][]kubernetes.IstioObject
	var mtlsDetails kubernetes.MTLSDetails

	wg.Add(5) // We need to add these here to make sure we don't execute wg.Wait() before scheduler has started goroutines

	// NoServiceChecker is not necessary if we target a single service - those components with validation errors won't show up in the query
	go in.fetchServices(&services, namespace, service, errChan, &wg)

	// We fetch without target service as some validations will require full-namespace details
	go in.fetchDetails(&istioDetails, namespace, errChan, &wg)
	go in.fetchWorkloads(&workloads, namespace, errChan, &wg)
	go in.fetchGatewaysPerNamespace(&gatewaysPerNamespace, errChan, &wg)
	go in.fetchNonLocalmTLSConfigs(&mtlsDetails, errChan, &wg)

	wg.Wait()
	close(errChan)
	for e := range errChan {
		if e != nil { // Check that default value wasn't returned
			return nil, err
		}
	}

	objectCheckers := in.getAllObjectCheckers(namespace, istioDetails, services, workloads, gatewaysPerNamespace, mtlsDetails)

	// Get group validations for same kind istio objects
	return runObjectCheckers(objectCheckers), nil
}

func (in *IstioValidationsService) getAllObjectCheckers(namespace string, istioDetails kubernetes.IstioDetails, services []v1.Service, workloads models.WorkloadList, gatewaysPerNamespace [][]kubernetes.IstioObject, mtlsDetails kubernetes.MTLSDetails) []ObjectChecker {
	return []ObjectChecker{
		checkers.VirtualServiceChecker{Namespace: namespace, DestinationRules: istioDetails.DestinationRules, VirtualServices: istioDetails.VirtualServices},
		checkers.NoServiceChecker{Namespace: namespace, IstioDetails: &istioDetails, Services: services, WorkloadList: workloads, GatewaysPerNamespace: gatewaysPerNamespace},
		checkers.DestinationRulesChecker{DestinationRules: istioDetails.DestinationRules, MTLSDetails: mtlsDetails},
		checkers.GatewayChecker{GatewaysPerNamespace: gatewaysPerNamespace, Namespace: namespace},
	}
}

func (in *IstioValidationsService) GetIstioObjectValidations(namespace string, objectType string, object string) (models.IstioValidations, error) {
	var err error
	promtimer := internalmetrics.GetGoFunctionMetric("business", "IstioValidationsService", "GetIstioObjectValidations")
	defer promtimer.ObserveNow(&err)

	var istioDetails kubernetes.IstioDetails
	var services []v1.Service
	var workloads models.WorkloadList
	var gatewaysPerNamespace [][]kubernetes.IstioObject
	var mtlsDetails kubernetes.MTLSDetails

	var objectCheckers []ObjectChecker

	wg := sync.WaitGroup{}
	errChan := make(chan error, 1)

	// Get all the Istio objects from a Namespace and all gateways from every namespace
	wg.Add(5)
	go in.fetchDetails(&istioDetails, namespace, errChan, &wg)
	go in.fetchServices(&services, namespace, "", errChan, &wg)
	go in.fetchWorkloads(&workloads, namespace, errChan, &wg)
	go in.fetchGatewaysPerNamespace(&gatewaysPerNamespace, errChan, &wg)
	go in.fetchNonLocalmTLSConfigs(&mtlsDetails, errChan, &wg)
	wg.Wait()

	switch objectType {
	case Gateways:
		objectCheckers = []ObjectChecker{
			checkers.GatewayChecker{GatewaysPerNamespace: gatewaysPerNamespace, Namespace: namespace},
		}
	case VirtualServices:
		virtualServiceChecker := checkers.VirtualServiceChecker{Namespace: namespace, VirtualServices: istioDetails.VirtualServices, DestinationRules: istioDetails.DestinationRules}
		noServiceChecker := checkers.NoServiceChecker{Namespace: namespace, Services: services, IstioDetails: &istioDetails, WorkloadList: workloads, GatewaysPerNamespace: gatewaysPerNamespace}
		objectCheckers = []ObjectChecker{noServiceChecker, virtualServiceChecker}
	case DestinationRules:
		destinationRulesChecker := checkers.DestinationRulesChecker{DestinationRules: istioDetails.DestinationRules, MTLSDetails: mtlsDetails}
		noServiceChecker := checkers.NoServiceChecker{Namespace: namespace, Services: services, IstioDetails: &istioDetails, WorkloadList: workloads, GatewaysPerNamespace: gatewaysPerNamespace}
		objectCheckers = []ObjectChecker{noServiceChecker, destinationRulesChecker}
	case ServiceEntries:
		// Validations on ServiceEntries are not yet in place
	case Rules:
		// Validations on Istio Rules are not yet in place
	case Templates:
		// Validations on Templates are not yet in place
		// TODO Support subtypes
	case Adapters:
		// Validations on Adapters are not yet in place
		// TODO Support subtypes
	case QuotaSpecs:
		// Validations on QuotaSpecs are not yet in place
	case QuotaSpecBindings:
		// Validations on QuotaSpecBindings are not yet in place
	case Policies:
		// Validations on Policies are not yet in place
	case MeshPolicies:
		// Validations on MeshPolicies are not yet in place
	case ClusterRbacConfigs:
		// Validations on ClusterRbacConfigs are not yet in place
	case ServiceRoles:
		// Validations on ServiceRoles are not yet in place
	case ServiceRoleBindings:
		// Validations on ServiceRoleBindings are not yet in place
	default:
		err = fmt.Errorf("Object type not found: %v", objectType)
	}

	close(errChan)
	for e := range errChan {
		if e != nil { // Check that default value wasn't returned
			return nil, err
		}
	}

	if objectCheckers == nil {
		return models.IstioValidations{}, err
	}

	return runObjectCheckers(objectCheckers).FilterByKey(models.ObjectTypeSingular[objectType], object), nil
}

func runObjectCheckers(objectCheckers []ObjectChecker) models.IstioValidations {
	objectTypeValidations := models.IstioValidations{}

	// Run checks for each IstioObject type
	for _, objectChecker := range objectCheckers {
		objectTypeValidations.MergeValidations(objectChecker.Check())
	}

	return objectTypeValidations
}

// The following idea is used underneath: if errChan has at least one record, we'll effectively cancel the request (if scheduled in such order). On the other hand, if we can't
// write to the buffered errChan, we just ignore the error as select does not block even if channel is full. This is because a single error is enough to cancel the whole request.

func (in *IstioValidationsService) fetchGatewaysPerNamespace(gatewaysPerNamespace *[][]kubernetes.IstioObject, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	if nss, err := in.businessLayer.Namespace.GetNamespaces(); err == nil {
		gwss := make([][]kubernetes.IstioObject, len(nss))
		for i := range nss {
			gwss[i] = make([]kubernetes.IstioObject, 0)
		}
		*gatewaysPerNamespace = gwss

		wg.Add(len(nss))
		for i, ns := range nss {
			go fetchNoEntry(&gwss[i], ns.Name, in.k8s.GetGateways, wg, errChan)
		}
	} else {
		select {
		case errChan <- err:
		default:
		}
	}
}

func fetchNoEntry(rValue *[]kubernetes.IstioObject, namespace string, fetcher func(string) ([]kubernetes.IstioObject, error), wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()
	if len(errChan) == 0 {
		fetched, err := fetcher(namespace)
		*rValue = append(*rValue, fetched...)
		if err != nil {
			select {
			case errChan <- err:
			default:
			}
		}
	}
}

func (in *IstioValidationsService) fetchServices(rValue *[]v1.Service, namespace, serviceName string, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(errChan) == 0 {
		services, err := in.k8s.GetServices(namespace, nil)
		if err != nil {
			select {
			case errChan <- err:
			default:
			}
		} else {
			*rValue = services
		}
	}
}

func (in *IstioValidationsService) fetchWorkloads(rValue *models.WorkloadList, namespace string, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(errChan) == 0 {
		workloadList, err := in.businessLayer.Workload.GetWorkloadList(namespace)
		if err != nil {
			select {
			case errChan <- err:
			default:
			}
		} else {
			*rValue = workloadList
		}
	}
}

func (in *IstioValidationsService) fetchDetails(rValue *kubernetes.IstioDetails, namespace string, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(errChan) == 0 {
		istioDetails, err := in.k8s.GetIstioDetails(namespace, "")
		if err != nil {
			select {
			case errChan <- err:
			default:
			}
		} else {
			*rValue = *istioDetails
		}
	}
}

func (in *IstioValidationsService) fetchNonLocalmTLSConfigs(mtlsDetails *kubernetes.MTLSDetails, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(errChan) > 0 {
		return
	}

	namespaces, err := in.businessLayer.Namespace.GetNamespaces()
	if err != nil {
		errChan <- err
		return
	}

	nsNames := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		nsNames = append(nsNames, ns.Name)
	}

	destinationRules, err := in.businessLayer.IstioConfig.getAllDestinationRules(nsNames)
	if err != nil {
		errChan <- err
	} else {
		mtlsDetails.DestinationRules = destinationRules
	}
}
