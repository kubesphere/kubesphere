package appender

import (
	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/graph"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/models"
)

const UnusedNodeAppenderName = "unusedNode"

// UnusedNodeAppender looks for services that have never seen request traffic.  It adds nodes to represent the
// unused definitions.  The added node types depend on the graph type and/or labeling on the definition.
// Name: unusedNode
type UnusedNodeAppender struct {
	GraphType   string // This appender does not operate on service graphs because it adds workload nodes.
	IsNodeGraph bool   // This appender does not operate on node detail graphs because we want to focus on the specific node.
}

// Name implements Appender
func (a UnusedNodeAppender) Name() string {
	return UnusedNodeAppenderName
}

// AppendGraph implements Appender
func (a UnusedNodeAppender) AppendGraph(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo) {
	if graph.GraphTypeService == a.GraphType || a.IsNodeGraph {
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

	a.addUnusedNodes(trafficMap, namespaceInfo.Namespace, namespaceInfo.WorkloadList.Workloads)
}

func (a UnusedNodeAppender) addUnusedNodes(trafficMap graph.TrafficMap, namespace string, workloads []models.WorkloadListItem) {
	unusedTrafficMap := a.buildUnusedTrafficMap(trafficMap, namespace, workloads)

	// If trafficMap is empty just populate it with the unused nodes and return
	if len(trafficMap) == 0 {
		for k, v := range unusedTrafficMap {
			trafficMap[k] = v
		}
		return
	}

	// Integrate the unused nodes into the existing traffic map
	for _, v := range unusedTrafficMap {
		addUnusedNodeToTrafficMap(trafficMap, v)
	}
}

func (a UnusedNodeAppender) buildUnusedTrafficMap(trafficMap graph.TrafficMap, namespace string, workloads []models.WorkloadListItem) graph.TrafficMap {
	unusedTrafficMap := graph.NewTrafficMap()
	cfg := config.Get()
	appLabel := cfg.IstioLabels.AppLabelName
	versionLabel := cfg.IstioLabels.VersionLabelName
	for _, w := range workloads {
		labels := w.Labels
		app := graph.Unknown
		version := graph.Unknown
		if v, ok := labels[appLabel]; ok {
			app = v
		}
		if v, ok := labels[versionLabel]; ok {
			version = v
		}
		id, nodeType := graph.Id(namespace, w.Name, app, version, "", a.GraphType)
		if _, found := trafficMap[id]; !found {
			if _, found = unusedTrafficMap[id]; !found {
				log.Debugf("Adding unused node for workload [%s] with labels [%v]", w.Name, labels)
				node := graph.NewNodeExplicit(id, namespace, w.Name, app, version, "", nodeType, a.GraphType)
				// note: we don't know what the protocol really should be, http is most common, it's a dead edge anyway
				node.Metadata = map[string]interface{}{"httpIn": 0.0, "httpOut": 0.0, "isUnused": true}
				unusedTrafficMap[id] = &node
			}
		}
	}
	return unusedTrafficMap
}

func addUnusedNodeToTrafficMap(trafficMap graph.TrafficMap, unusedNode *graph.Node) {
	// add unused node to traffic map
	trafficMap[unusedNode.ID] = unusedNode

	// Add a "sibling" edge to any node with an edge to the same app
	for _, n := range trafficMap {
		findAndAddSibling(n, unusedNode)
	}
}

func findAndAddSibling(parent, unusedNode *graph.Node) {
	if unusedNode.App == graph.Unknown {
		return
	}

	found := false
	for _, edge := range parent.Edges {
		if found = edge.Dest.App == unusedNode.App; found {
			break
		}
	}
	if found {
		parent.AddEdge(unusedNode)
	}
}
