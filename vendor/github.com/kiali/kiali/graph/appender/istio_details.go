package appender

import (
	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/graph"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
)

const IstioAppenderName = "istio"

// IstioAppender is responsible for badging nodes with special Istio significance:
// - CircuitBreaker: n.Metadata["hasCB"] = true
// - VirtualService: n.Metadata["hasVS"] = true
// Name: istio
type IstioAppender struct{}

// Name implements Appender
func (a IstioAppender) Name() string {
	return IstioAppenderName
}

// AppendGraph implements Appender
func (a IstioAppender) AppendGraph(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo) {
	if len(trafficMap) == 0 {
		return
	}

	if globalInfo.Business == nil {
		var err error
		globalInfo.Business, err = business.Get()
		graph.CheckError(err)
	}

	addBadging(trafficMap, globalInfo, namespaceInfo)
	addLabels(trafficMap, globalInfo)
}

func addBadging(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo) {
	// Currently no other appenders use DestinationRules or VirtualServices, so they are not cached in NamespaceInfo
	istioCfg, err := globalInfo.Business.IstioConfig.GetIstioConfigList(business.IstioConfigCriteria{
		IncludeDestinationRules: true,
		IncludeVirtualServices:  true,
		Namespace:               namespaceInfo.Namespace,
	})
	graph.CheckError(err)

	applyCircuitBreakers(trafficMap, namespaceInfo.Namespace, istioCfg)
	applyVirtualServices(trafficMap, namespaceInfo.Namespace, istioCfg)
}

func applyCircuitBreakers(trafficMap graph.TrafficMap, namespace string, istioCfg models.IstioConfigList) {
NODES:
	for _, n := range trafficMap {
		// Skip the check if this node is outside the requested namespace, we limit badging to the requested namespaces
		if n.Namespace != namespace {
			continue
		}

		// Note, Because DestinationRules are applied to services we limit CB badges to service nodes and app nodes.
		// Whether we should add to workload nodes is debatable, we could add it later if needed.
		versionOk := graph.IsOK(n.Version)
		switch {
		case n.NodeType == graph.NodeTypeService:
			for _, destinationRule := range istioCfg.DestinationRules.Items {
				if destinationRule.HasCircuitBreaker(namespace, n.Service, "") {
					n.Metadata["hasCB"] = true
					continue NODES
				}
			}
		case !versionOk && (n.NodeType == graph.NodeTypeApp):
			if destServices, ok := n.Metadata["destServices"]; ok {
				for serviceName, _ := range destServices.(map[string]bool) {
					for _, destinationRule := range istioCfg.DestinationRules.Items {
						if destinationRule.HasCircuitBreaker(namespace, serviceName, "") {
							n.Metadata["hasCB"] = true
							continue NODES
						}
					}
				}
			}
		case versionOk:
			if destServices, ok := n.Metadata["destServices"]; ok {
				for serviceName, _ := range destServices.(map[string]bool) {
					for _, destinationRule := range istioCfg.DestinationRules.Items {
						if destinationRule.HasCircuitBreaker(namespace, serviceName, n.Version) {
							n.Metadata["hasCB"] = true
							continue NODES
						}
					}
				}
			}
		default:
			continue
		}
	}
}

func applyVirtualServices(trafficMap graph.TrafficMap, namespace string, istioCfg models.IstioConfigList) {
NODES:
	for _, n := range trafficMap {
		if n.NodeType != graph.NodeTypeService {
			continue
		}
		if n.Namespace != namespace {
			continue
		}
		for _, virtualService := range istioCfg.VirtualServices.Items {
			if virtualService.IsValidHost(namespace, n.Service) {
				n.Metadata["hasVS"] = true
				continue NODES
			}
		}
	}
}

// addLabels is a chance to add any missing label info to nodes when the telemetry does not provide enough information.
// TODO: For efficiency we may want to consider pulling all namespace service definitions in one call (the call does not
//       exist at this writing).  As written we pull each service individually, which can be a fair number of round
//       trips when services are injected (as they are by default). Note also that currently we do query for
//       outsider service nodes.  That may be a security problem f the outside namespace is inaccessible to the user. If
//       that becomes an issue we can limit to accessible namespaces or only to the NamespaceInfo namespace.
func addLabels(trafficMap graph.TrafficMap, globalInfo *GlobalInfo) {
	appLabelName := config.Get().IstioLabels.AppLabelName
	for _, n := range trafficMap {
		// make sure service nodes have the defined app label so it can be used for app grouping in the UI.
		if n.NodeType == graph.NodeTypeService && n.App == "" {
			service, err := globalInfo.Business.Svc.GetServiceDefinition(n.Namespace, n.Service)
			if err != nil {
				log.Debugf("Error fetching service definition, may not apply app label correctly for namespace=%s svc=%s: %s", n.Namespace, n.Service, err.Error())
				if service == nil {
					continue
				}
			}

			if app, ok := service.Service.Labels[appLabelName]; ok {
				n.App = app
			}
		}
	}
}
