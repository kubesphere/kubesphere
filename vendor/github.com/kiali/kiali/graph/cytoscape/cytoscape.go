// Cytoscape package provides conversion from our graph to the CystoscapeJS
// configuration json model.
//
// The following links are useful for understanding CytoscapeJS and it's configuration:
//
// Main page:   http://js.cytoscape.org/
// JSON config: http://js.cytoscape.org/#notation/elements-json
// Demos:       http://js.cytoscape.org/#demos
//
// Algorithm: Process the graph structure adding nodes and edges, decorating each
//            with information provided.  An optional second pass generates compound
//            nodes for version grouping.
//
package cytoscape

import (
	"crypto/md5"
	"fmt"
	"sort"

	"github.com/kiali/kiali/graph"
	"github.com/kiali/kiali/graph/options"
)

type ProtocolTraffic struct {
	Protocol string            `json:"protocol"` // protocol
	Rates    map[string]string `json:"rates"`    // map[rate]value
}

type NodeData struct {
	// Cytoscape Fields
	Id     string `json:"id"`               // unique internal node ID (n0, n1...)
	Parent string `json:"parent,omitempty"` // Compound Node parent ID

	// App Fields (not required by Cytoscape)
	NodeType        string            `json:"nodeType"`
	Namespace       string            `json:"namespace"`
	Workload        string            `json:"workload,omitempty"`
	App             string            `json:"app,omitempty"`
	Version         string            `json:"version,omitempty"`
	Service         string            `json:"service,omitempty"`         // requested service for NodeTypeService
	DestServices    map[string]bool   `json:"destServices,omitempty"`    // requested services for [dest] node
	Traffic         []ProtocolTraffic `json:"traffic,omitempty"`         // traffic rates for all detected protocols
	HasCB           bool              `json:"hasCB,omitempty"`           // true (has circuit breaker) | false
	HasMissingSC    bool              `json:"hasMissingSC,omitempty"`    // true (has missing sidecar) | false
	HasVS           bool              `json:"hasVS,omitempty"`           // true (has route rule) | false
	IsDead          bool              `json:"isDead,omitempty"`          // true (has no pods) | false
	IsGroup         string            `json:"isGroup,omitempty"`         // set to the grouping type, current values: [ 'app', 'version' ]
	IsInaccessible  bool              `json:"isInaccessible,omitempty"`  // true if the node exists in an inaccessible namespace
	IsMisconfigured string            `json:"isMisconfigured,omitempty"` // set to misconfiguration list, current values: [ 'labels' ]
	IsOutside       bool              `json:"isOutside,omitempty"`       // true | false
	IsRoot          bool              `json:"isRoot,omitempty"`          // true | false
	IsServiceEntry  string            `json:"isServiceEntry,omitempty"`  // set to the location, current values: [ 'MESH_EXTERNAL', 'MESH_INTERNAL' ]
	IsUnused        bool              `json:"isUnused,omitempty"`        // true | false
}

type EdgeData struct {
	// Cytoscape Fields
	Id     string `json:"id"`     // unique internal edge ID (e0, e1...)
	Source string `json:"source"` // parent node ID
	Target string `json:"target"` // child node ID

	// App Fields (not required by Cytoscape)
	Traffic      ProtocolTraffic `json:"traffic,omitempty"`      // traffic rates for the edge protocol
	ResponseTime string          `json:"responseTime,omitempty"` // in millis
	IsMTLS       string          `json:"isMTLS,omitempty"`       // set to the percentage of traffic using a mutual TLS connection
	IsUnused     bool            `json:"isUnused,omitempty"`     // true | false
}

type NodeWrapper struct {
	Data *NodeData `json:"data"`
}

type EdgeWrapper struct {
	Data *EdgeData `json:"data"`
}

type Elements struct {
	Nodes []*NodeWrapper `json:"nodes"`
	Edges []*EdgeWrapper `json:"edges"`
}

type Config struct {
	Timestamp int64    `json:"timestamp"`
	Duration  int64    `json:"duration"`
	GraphType string   `json:"graphType"`
	Elements  Elements `json:"elements"`
}

func nodeHash(id string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(id)))
}

func edgeHash(from, to, protocol string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s.%s.%s", from, to, protocol))))
}

func NewConfig(trafficMap graph.TrafficMap, o options.VendorOptions) (result Config) {
	nodes := []*NodeWrapper{}
	edges := []*EdgeWrapper{}

	buildConfig(trafficMap, &nodes, &edges, o)

	// Add compound nodes as needed
	switch o.GroupBy {
	case options.GroupByApp:
		if o.GraphType != graph.GraphTypeService {
			groupByApp(&nodes)
		}
	case options.GroupByVersion:
		if o.GraphType == graph.GraphTypeVersionedApp {
			groupByVersion(&nodes)
		}
	default:
		// no grouping
	}

	// sort nodes and edges for better json presentation (and predictable testing)
	// kiali-1258 compound/isGroup/parent nodes must come before the child references
	sort.Slice(nodes, func(i, j int) bool {
		switch {
		case nodes[i].Data.Namespace != nodes[j].Data.Namespace:
			return nodes[i].Data.Namespace < nodes[j].Data.Namespace
		case nodes[i].Data.IsGroup != nodes[j].Data.IsGroup:
			return nodes[i].Data.IsGroup > nodes[j].Data.IsGroup
		case nodes[i].Data.App != nodes[j].Data.App:
			return nodes[i].Data.App < nodes[j].Data.App
		case nodes[i].Data.Version != nodes[j].Data.Version:
			return nodes[i].Data.Version < nodes[j].Data.Version
		case nodes[i].Data.Service != nodes[j].Data.Service:
			return nodes[i].Data.Service < nodes[j].Data.Service
		default:
			return nodes[i].Data.Workload < nodes[j].Data.Workload
		}
	})
	sort.Slice(edges, func(i, j int) bool {
		switch {
		case edges[i].Data.Source < edges[j].Data.Source:
			return true
		case edges[i].Data.Source > edges[j].Data.Source:
			return false
		default:
			return edges[i].Data.Target < edges[j].Data.Target
		}
	})

	elements := Elements{nodes, edges}
	result = Config{
		Duration:  int64(o.Duration.Seconds()),
		Timestamp: o.QueryTime,
		GraphType: o.GraphType,
		Elements:  elements,
	}
	return result
}

func buildConfig(trafficMap graph.TrafficMap, nodes *[]*NodeWrapper, edges *[]*EdgeWrapper, o options.VendorOptions) {
	for id, n := range trafficMap {
		nodeId := nodeHash(id)

		nd := &NodeData{
			Id:        nodeId,
			NodeType:  n.NodeType,
			Namespace: n.Namespace,
			Workload:  n.Workload,
			App:       n.App,
			Version:   n.Version,
			Service:   n.Service,
		}

		addNodeTelemetry(n, nd)

		// node may have deployment but no pods running)
		if val, ok := n.Metadata["isDead"]; ok {
			nd.IsDead = val.(bool)
		}

		// node may be a root
		if val, ok := n.Metadata["isRoot"]; ok {
			nd.IsRoot = val.(bool)
		}

		// node may be unused
		if val, ok := n.Metadata["isUnused"]; ok {
			nd.IsUnused = val.(bool)
		}

		// node is not accessible to the current user
		if val, ok := n.Metadata["isInaccessible"]; ok {
			nd.IsInaccessible = val.(bool)
		}

		// node may have a circuit breaker
		if val, ok := n.Metadata["hasCB"]; ok {
			nd.HasCB = val.(bool)
		}

		// node may have a virtual service
		if val, ok := n.Metadata["hasVS"]; ok {
			nd.HasVS = val.(bool)
		}

		// set sidecars checks, if available
		if val, ok := n.Metadata["hasMissingSC"]; ok {
			nd.HasMissingSC = val.(bool)
		}

		// check if node is misconfigured
		if val, ok := n.Metadata["isMisconfigured"]; ok {
			nd.IsMisconfigured = val.(string)
		}

		// check if node is on another namespace
		if val, ok := n.Metadata["isOutside"]; ok {
			nd.IsOutside = val.(bool)
		}

		// node may have destination service info
		if val, ok := n.Metadata["destServices"]; ok {
			nd.DestServices = val.(map[string]bool)
		}

		// node may be a service entry
		if val, ok := n.Metadata["isServiceEntry"]; ok {
			nd.IsServiceEntry = val.(string)
		}

		nw := NodeWrapper{
			Data: nd,
		}

		*nodes = append(*nodes, &nw)

		for _, e := range n.Edges {
			sourceIdHash := nodeHash(n.ID)
			destIdHash := nodeHash(e.Dest.ID)
			protocol := ""
			if e.Metadata["protocol"] != nil {
				protocol = e.Metadata["protocol"].(string)
			}
			edgeId := edgeHash(sourceIdHash, destIdHash, protocol)
			ed := EdgeData{
				Id:     edgeId,
				Source: sourceIdHash,
				Target: destIdHash,
			}
			addEdgeTelemetry(e, &ed)

			ew := EdgeWrapper{
				Data: &ed,
			}
			*edges = append(*edges, &ew)
		}
	}
}

func addNodeTelemetry(n *graph.Node, nd *NodeData) {
	nd.Traffic = []ProtocolTraffic{}
	for _, p := range graph.Protocols {
		protocolTraffic := ProtocolTraffic{Protocol: p.Name}
		for _, r := range p.NodeRates {
			if rateVal := getRate(n.Metadata, r.Name); rateVal > 0.0 {
				if protocolTraffic.Rates == nil {
					protocolTraffic.Rates = make(map[string]string)
				}
				protocolTraffic.Rates[r.Name] = fmt.Sprintf("%.*f", r.Precision, rateVal)
			}
		}
		if protocolTraffic.Rates != nil {
			nd.Traffic = append(nd.Traffic, protocolTraffic)
		}
	}
}

func addEdgeTelemetry(e *graph.Edge, ed *EdgeData) {
	if val, ok := e.Metadata["isMTLS"]; ok {
		ed.IsMTLS = fmt.Sprintf("%.0f", val.(float64))
	}
	if val, ok := e.Metadata["responseTime"]; ok {
		responseTime := val.(float64)
		ed.ResponseTime = fmt.Sprintf("%.0f", responseTime)
	}
	if val, ok := e.Source.Metadata["isUnused"]; ok {
		ed.IsUnused = val.(bool)
	}

	// an edge represents traffic for at most one protocol
	ed.Traffic = ProtocolTraffic{}
	for _, p := range graph.Protocols {
		protocolTraffic := ProtocolTraffic{Protocol: p.Name}
		total := 0.0
		err := 0.0
		var percentErr, percentReq graph.Rate
		for _, r := range p.EdgeRates {
			rateVal := getRate(e.Metadata, r.Name)
			switch {
			case r.IsTotal:
				// there is one field holding the total traffic
				total = rateVal
			case r.IsErr:
				// error rates can be reported for several error status codes, so sum up all
				// of the error traffic to be used in the percentErr calculation below.
				err += rateVal
			case r.IsPercentErr:
				// hold onto the percentErr field so we know how to report it below
				percentErr = r
			case r.IsPercentReq:
				// hold onto the percentReq field so we know how to report it below
				percentReq = r
			}
			if rateVal := getRate(e.Metadata, r.Name); rateVal > 0.0 {
				if protocolTraffic.Rates == nil {
					protocolTraffic.Rates = make(map[string]string)
				}
				protocolTraffic.Rates[r.Name] = fmt.Sprintf("%.*f", r.Precision, rateVal)
			}
		}
		if protocolTraffic.Rates != nil {
			if total > 0 {
				if percentErr.Name != "" {
					rateVal := err / total * 100
					if rateVal > 0.0 {
						protocolTraffic.Rates[percentErr.Name] = fmt.Sprintf("%.*f", percentErr.Precision, rateVal)
					}
				}
				if percentReq.Name != "" {
					rateVal := 0.0
					for _, r := range p.NodeRates {
						if !r.IsOut {
							continue
						}
						rateVal = total / getRate(e.Source.Metadata, r.Name) * 100.0
						break
					}
					if rateVal > 0.0 {
						protocolTraffic.Rates[percentReq.Name] = fmt.Sprintf("%.*f", percentReq.Precision, rateVal)
					}
				}
			}
			ed.Traffic = protocolTraffic
			break
		}
	}
}

func getRate(md map[string]interface{}, k string) float64 {
	if rate, ok := md[k]; ok {
		return rate.(float64)
	}
	return 0.0
}

// groupByVersion adds compound nodes to group multiple versions of the same app
func groupByVersion(nodes *[]*NodeWrapper) {
	appBox := make(map[string][]*NodeData)

	for _, nw := range *nodes {
		if nw.Data.NodeType == graph.NodeTypeApp {
			k := fmt.Sprintf("box_%s_%s", nw.Data.Namespace, nw.Data.App)
			appBox[k] = append(appBox[k], nw.Data)
		}
	}

	generateGroupCompoundNodes(appBox, nodes, options.GroupByVersion)
}

// groupByApp adds compound nodes to group all nodes for the same app
func groupByApp(nodes *[]*NodeWrapper) {
	appBox := make(map[string][]*NodeData)

	for _, nw := range *nodes {
		if nw.Data.App != "unknown" && nw.Data.App != "" {
			k := fmt.Sprintf("box_%s_%s", nw.Data.Namespace, nw.Data.App)
			appBox[k] = append(appBox[k], nw.Data)
		}
	}

	generateGroupCompoundNodes(appBox, nodes, options.GroupByApp)
}

func generateGroupCompoundNodes(appBox map[string][]*NodeData, nodes *[]*NodeWrapper, groupBy string) {
	for k, members := range appBox {
		if len(members) > 1 {
			// create the compound (parent) node for the member nodes
			nodeId := nodeHash(k)
			nd := NodeData{
				Id:        nodeId,
				NodeType:  graph.NodeTypeApp,
				Namespace: members[0].Namespace,
				App:       members[0].App,
				Version:   "",
				IsGroup:   groupBy,
			}

			nw := NodeWrapper{
				Data: &nd,
			}

			// assign each member node to the compound parent
			nd.HasMissingSC = false // TODO: this is probably unecessarily noisy
			nd.IsInaccessible = false
			nd.IsOutside = false

			for _, n := range members {
				n.Parent = nodeId

				// copy some member attributes to to the compound node (aka app box)
				nd.HasMissingSC = nd.HasMissingSC || n.HasMissingSC
				nd.IsInaccessible = nd.IsInaccessible || n.IsInaccessible
				nd.IsOutside = nd.IsOutside || n.IsOutside
			}

			// add the compound node to the list of nodes
			*nodes = append(*nodes, &nw)
		}
	}
}
