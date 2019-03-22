package appender

import (
	"time"

	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/graph"
	"github.com/kiali/kiali/log"
)

const ServiceEntryAppenderName = "serviceEntry"

// ServiceEntryAppender is responsible for identifying service nodes that are
// Istio Service Entries.
// Name: serviceEntry
type ServiceEntryAppender struct {
	AccessibleNamespaces map[string]time.Time
}

// Name implements Appender
func (a ServiceEntryAppender) Name() string {
	return ServiceEntryAppenderName
}

// AppendGraph implements Appender
func (a ServiceEntryAppender) AppendGraph(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo) {
	if len(trafficMap) == 0 {
		return
	}

	var err error
	if globalInfo.Business == nil {
		globalInfo.Business, err = business.Get()
		graph.CheckError(err)
	}

	a.applyServiceEntries(trafficMap, globalInfo, namespaceInfo)
}

func (a ServiceEntryAppender) applyServiceEntries(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo) {
	for _, n := range trafficMap {
		// only a service node can be a service entry
		if n.NodeType != graph.NodeTypeService {
			continue
		}
		// only a terminal node can be a service entry (no outgoing edges because the service is performed outside the mesh)
		if len(n.Edges) > 0 {
			continue
		}

		// A service node with no outgoing edges may be an egress.
		// If so flag it, don't discard it (kiali-1526, see also kiali-2014).
		// The flag will be passed to the UI to inhibit links to non-existent detail pages.
		if location, ok := a.getServiceEntry(n.Service, globalInfo); ok {
			n.Metadata["isServiceEntry"] = location
		}
	}
}

// getServiceEntry queries the cluster API to resolve service entries
// across all accessible namespaces in the cluster. All ServiceEntries are needed because
// Istio does not distinguish where a ServiceEntry is created when routing traffic (i.e.
// a ServiceEntry can be in any namespace and it will still work).
func (a ServiceEntryAppender) getServiceEntry(service string, globalInfo *GlobalInfo) (string, bool) {
	if globalInfo.ServiceEntries == nil {
		globalInfo.ServiceEntries = make(map[string]string)

		for ns := range a.AccessibleNamespaces {
			istioCfg, err := globalInfo.Business.IstioConfig.GetIstioConfigList(business.IstioConfigCriteria{
				IncludeServiceEntries: true,
				Namespace:             ns,
			})
			graph.CheckError(err)

			for _, entry := range istioCfg.ServiceEntries {
				if entry.Spec.Hosts != nil {
					location := "MESH_EXTERNAL"
					if entry.Spec.Location == "MESH_INTERNAL" {
						location = "MESH_INTERNAL"
					}
					for _, host := range entry.Spec.Hosts.([]interface{}) {
						globalInfo.ServiceEntries[host.(string)] = location
					}
				}
			}
		}
		log.Tracef("Found [%v] service entries", len(globalInfo.ServiceEntries))
	}

	location, ok := globalInfo.ServiceEntries[service]
	return location, ok
}
