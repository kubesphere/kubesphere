package appender

import (
	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/graph"
)

const DeadNodeAppenderName = "deadNode"

// DeadNodeAppender is responsible for removing from the graph unwanted nodes:
// - nodes for which there is no traffic reported and a backing workload that can't be found
//   (presumably removed from K8S). (kiali-621)
//   - this includes "unknown"
// - service nodes that are not service entries (kiali-1526) and for which there is no incoming
//   error traffic and no outgoing edges (kiali-1326).
// Name: deadNode
type DeadNodeAppender struct{}

// Name implements Appender
func (a DeadNodeAppender) Name() string {
	return DeadNodeAppenderName
}

// AppendGraph implements Appender
func (a DeadNodeAppender) AppendGraph(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo) {
	if len(trafficMap) == 0 {
		return
	}

	var err error
	if globalInfo.Business == nil {
		globalInfo.Business, err = business.Get()
		graph.CheckError(err)
	}
	if namespaceInfo.WorkloadList == nil {
		workloadList, err := globalInfo.Business.Workload.GetWorkloadList(namespaceInfo.Namespace)
		graph.CheckError(err)
		namespaceInfo.WorkloadList = &workloadList
	}

	a.applyDeadNodes(trafficMap, globalInfo, namespaceInfo)
}

func (a DeadNodeAppender) applyDeadNodes(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo) {
	numRemoved := 0
	for id, n := range trafficMap {
		switch n.NodeType {
		case graph.NodeTypeService:
			// a service node with outgoing edges is never considered dead (or egress)
			if len(n.Edges) > 0 {
				continue
			}

			// A service node that is a service entry is never considered dead
			if _, ok := n.Metadata["isServiceEntry"]; ok {
				continue
			}

			// a service node with no incoming error traffic and no outgoing edges, is dead.
			// Incoming non-error traffic can not raise the dead because it is caused by an
			// edge case (pod life-cycle change) that we don't want to see.
			isDead := true
		ServiceCase:
			for _, p := range graph.Protocols {
				for _, r := range p.NodeRates {
					if r.IsErr {
						if errRate, hasErrRate := n.Metadata[r.Name]; hasErrRate && errRate.(float64) > 0 {
							isDead = false
							break ServiceCase
						}
					}
				}
			}
			if isDead {
				delete(trafficMap, id)
				numRemoved++
			}
		default:
			// a node with traffic is not dead, skip
			isDead := true
		DefaultCase:
			for _, p := range graph.Protocols {
				for _, r := range p.NodeRates {
					if r.IsIn || r.IsOut {
						if rate, hasRate := n.Metadata[r.Name]; hasRate && rate.(float64) > 0 {
							isDead = false
							break DefaultCase
						}
					}
				}
			}
			if !isDead {
				continue
			}

			// There are some node types that are never associated with backing workloads (such as versionless app nodes).
			// Nodes of those types are never dead because their workload clearly can't be missing (they don't have workloads).
			// - note: unknown is not saved by this rule (kiali-2078) - i.e. unknown nodes can be declared dead
			if n.NodeType != graph.NodeTypeUnknown && !graph.IsOK(n.Workload) {
				continue
			}

			// Remove if backing workload is not defined (always true for "unknown"), flag if there are no pods
			if workload, found := getWorkload(n.Workload, namespaceInfo.WorkloadList); !found {
				delete(trafficMap, id)
				numRemoved++
			} else {
				if workload.PodCount == 0 {
					n.Metadata["isDead"] = true
				}
			}
		}
	}

	// If we removed any nodes we need to remove any edges to them as well...
	if numRemoved == 0 {
		return
	}

	for _, s := range trafficMap {
		goodEdges := []*graph.Edge{}
		for _, e := range s.Edges {
			if _, found := trafficMap[e.Dest.ID]; found {
				goodEdges = append(goodEdges, e)
			}
		}
		s.Edges = goodEdges
	}
}
