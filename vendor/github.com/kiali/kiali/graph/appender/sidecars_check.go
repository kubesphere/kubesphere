package appender

import (
	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/graph"
)

const SidecarsCheckAppenderName = "sidecarsCheck"

// SidecarsCheckAppender flags nodes whose backing workloads are missing at least one Envoy sidecar. Note that
// a node with no backing workloads is not flagged.
// Name: sidecarsCheck
type SidecarsCheckAppender struct{}

// Name implements Appender
func (a SidecarsCheckAppender) Name() string {
	return SidecarsCheckAppenderName
}

// AppendGraph implements Appender
func (a SidecarsCheckAppender) AppendGraph(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo) {
	if len(trafficMap) == 0 {
		return
	}

	if globalInfo.Business == nil {
		var err error
		globalInfo.Business, err = business.Get()
		graph.CheckError(err)
	}
	if namespaceInfo.WorkloadList == nil {
		workloadList, err := globalInfo.Business.Workload.GetWorkloadList(namespaceInfo.Namespace)
		graph.CheckError(err)
		namespaceInfo.WorkloadList = &workloadList
	}

	a.applySidecarsChecks(trafficMap, namespaceInfo)
}

func (a *SidecarsCheckAppender) applySidecarsChecks(trafficMap graph.TrafficMap, namespaceInfo *NamespaceInfo) {
	cfg := config.Get()
	istioNamespace := cfg.IstioNamespace
	workloadList := namespaceInfo.WorkloadList

	for _, n := range trafficMap {
		// Skip the check if this node is outside the requested namespace, we limit badging to the requested namespaces
		if n.Namespace != namespaceInfo.Namespace {
			continue
		}

		// We whitelist istio components because they may not report telemetry using injected sidecars.
		if n.Namespace == istioNamespace {
			continue
		}

		// dead nodes tell no tales (er, have no pods)
		if isDead, ok := n.Metadata["isDead"]; ok && isDead.(bool) {
			continue
		}

		// get the workloads for the node and check to see if they have sidecars. Note that
		// if there are no workloads/pods we don't flag it as missing sidecars.  No pods means
		// no missing sidecars.  (In most cases this means it was flagged as dead, and handled above)
		hasIstioSidecar := true
		switch n.NodeType {
		case graph.NodeTypeWorkload:
			if workload, found := getWorkload(n.Workload, workloadList); found {
				hasIstioSidecar = workload.IstioSidecar
			}
		case graph.NodeTypeApp:
			workloads := getAppWorkloads(n.App, n.Version, workloadList)
			if len(workloads) > 0 {
				for _, workload := range workloads {
					if !workload.IstioSidecar {
						hasIstioSidecar = false
						break
					}
				}
			}
		default:
			continue
		}

		if !hasIstioSidecar {
			n.Metadata["hasMissingSC"] = true
		}
	}
}
