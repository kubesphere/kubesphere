package appender

import (
	"fmt"
	"math"
	"time"

	"github.com/prometheus/common/model"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/graph"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/prometheus"
)

const (
	DefaultQuantile          = 0.95 // 95th percentile
	ResponseTimeAppenderName = "responseTime"
)

// ResponseTimeAppender is responsible for adding responseTime information to the graph. ResponseTime
// is represented as a percentile value. The default is 95th percentile, which means that
// 95% of requests executed in no more than the resulting milliseconds.
// Name: responseTime
type ResponseTimeAppender struct {
	GraphType          string
	InjectServiceNodes bool
	IncludeIstio       bool
	Namespaces         map[string]graph.NamespaceInfo
	Quantile           float64
	QueryTime          int64 // unix time in seconds
}

// Name implements Appender
func (a ResponseTimeAppender) Name() string {
	return ResponseTimeAppenderName
}

// AppendGraph implements Appender
func (a ResponseTimeAppender) AppendGraph(trafficMap graph.TrafficMap, globalInfo *GlobalInfo, namespaceInfo *NamespaceInfo) {
	if len(trafficMap) == 0 {
		return
	}

	if globalInfo.PromClient == nil {
		var err error
		globalInfo.PromClient, err = prometheus.NewClient()
		graph.CheckError(err)
	}

	a.appendGraph(trafficMap, namespaceInfo.Namespace, globalInfo.PromClient)
}

func (a ResponseTimeAppender) appendGraph(trafficMap graph.TrafficMap, namespace string, client *prometheus.Client) {
	quantile := a.Quantile
	if a.Quantile <= 0.0 || a.Quantile >= 100.0 {
		log.Warningf("Replacing invalid quantile [%.2f] with default [%.2f]", a.Quantile, DefaultQuantile)
		quantile = DefaultQuantile
	}
	log.Debugf("Generating responseTime using quantile [%.2f]; namespace = %v", quantile, namespace)
	duration := a.Namespaces[namespace].Duration

	// create map to quickly look up responseTime
	responseTimeMap := make(map[string]float64)

	// query prometheus for the responseTime info in three queries:
	// 1) query for responseTime originating from "unknown" (i.e. the internet)
	groupBy := "le,source_workload_namespace,source_workload,source_app,source_version,destination_service_namespace,destination_service_name,destination_workload,destination_app,destination_version"
	query := fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(%s{reporter="destination",source_workload="unknown",destination_service_namespace="%v",response_code=~"%s"}[%vs])) by (%s))`,
		quantile,
		"istio_request_duration_seconds_bucket",
		namespace,
		"2[0-9]{2}|^0$",         // must match success for all expected protocols
		int(duration.Seconds()), // range duration for the query
		groupBy)
	unkVector := promQuery(query, time.Unix(a.QueryTime, 0), client.API(), a)
	a.populateResponseTimeMap(responseTimeMap, &unkVector)

	// 2) query for responseTime originating from a workload outside of the namespace. Exclude any "unknown" source telemetry (an unusual corner case)
	query = fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(%s{reporter="source",source_workload_namespace!="%v",source_workload!="unknown",destination_service_namespace="%v",response_code=~"%s"}[%vs])) by (%s))`,
		quantile,
		"istio_request_duration_seconds_bucket",
		namespace,
		namespace,
		"2[0-9]{2}|^0$",         // must match success for all expected protocols
		int(duration.Seconds()), // range duration for the query
		groupBy)
	outVector := promQuery(query, time.Unix(a.QueryTime, 0), client.API(), a)
	a.populateResponseTimeMap(responseTimeMap, &outVector)

	// 3) query for responseTime originating from a workload inside of the namespace
	query = fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(%s{reporter="source",source_workload_namespace="%v",response_code=~"%s"}[%vs])) by (%s))`,
		quantile,
		"istio_request_duration_seconds_bucket",
		namespace,
		"2[0-9]{2}|^0$",         // must match success for all expected protocols
		int(duration.Seconds()), // range duration for the query
		groupBy)
	inVector := promQuery(query, time.Unix(a.QueryTime, 0), client.API(), a)
	a.populateResponseTimeMap(responseTimeMap, &inVector)

	// istio component telemetry is only reported destination-side, so we must perform additional queries
	if a.IncludeIstio {
		istioNamespace := config.Get().IstioNamespace

		// 4) if the target namespace is istioNamespace re-query for traffic originating from outside (other than unknown, covered in query #1)
		if namespace == istioNamespace {
			query = fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(%s{reporter="destination",source_workload!="unknown",source_workload_namespace!="%v",destination_service_namespace="%v",response_code=~"%s"}[%vs])) by (%s))`,
				quantile,
				"istio_request_duration_seconds_bucket",
				namespace,
				namespace,
				"2[0-9]{2}|^0$",         // must match success for all expected protocols
				int(duration.Seconds()), // range duration for the query
				groupBy)

			// fetch the externally originating request traffic time-series
			outIstioVector := promQuery(query, time.Unix(a.QueryTime, 0), client.API(), a)
			a.populateResponseTimeMap(responseTimeMap, &outIstioVector)
		}

		// 5) supplemental query for traffic originating from a workload inside of the namespace with istioSystem destination
		query = fmt.Sprintf(`histogram_quantile(%.2f, sum(rate(%s{reporter="destination",source_workload_namespace="%v",destination_service_namespace="%v",response_code=~"%s"}[%vs])) by (%s))`,
			quantile,
			"istio_request_duration_seconds_bucket",
			namespace,
			istioNamespace,
			"2[0-9]{2}|^0$",         // must match success for all expected protocols
			int(duration.Seconds()), // range duration for the query
			groupBy)

		// fetch the internally originating request traffic time-series
		inIstioVector := promQuery(query, time.Unix(a.QueryTime, 0), client.API(), a)
		a.populateResponseTimeMap(responseTimeMap, &inIstioVector)
	}

	applyResponseTime(trafficMap, responseTimeMap)
}

func applyResponseTime(trafficMap graph.TrafficMap, responseTimeMap map[string]float64) {
	for _, n := range trafficMap {
		for _, e := range n.Edges {
			key := fmt.Sprintf("%s %s", e.Source.ID, e.Dest.ID)
			e.Metadata["responseTime"] = responseTimeMap[key]
		}
	}
}

func (a ResponseTimeAppender) populateResponseTimeMap(responseTimeMap map[string]float64, vector *model.Vector) {
	for _, s := range *vector {
		m := s.Metric
		lSourceWlNs, sourceWlNsOk := m["source_workload_namespace"]
		lSourceWl, sourceWlOk := m["source_workload"]
		lSourceApp, sourceAppOk := m["source_app"]
		lSourceVer, sourceVerOk := m["source_version"]
		lDestSvcNs, destSvcNsOk := m["destination_service_namespace"]
		lDestSvcName, destSvcNameOk := m["destination_service_name"]
		lDestWl, destWlOk := m["destination_workload"]
		lDestApp, destAppOk := m["destination_app"]
		lDestVer, destVerOk := m["destination_version"]
		if !sourceWlNsOk || !sourceWlOk || !sourceAppOk || !sourceVerOk || !destSvcNsOk || !destSvcNameOk || !destWlOk || !destAppOk || !destVerOk {
			log.Warningf("Skipping %v, missing expected labels", m.String())
			continue
		}

		sourceWlNs := string(lSourceWlNs)
		sourceWl := string(lSourceWl)
		sourceApp := string(lSourceApp)
		sourceVer := string(lSourceVer)
		destSvcNs := string(lDestSvcNs)
		destSvcName := string(lDestSvcName)
		destWl := string(lDestWl)
		destApp := string(lDestApp)
		destVer := string(lDestVer)

		// to best preserve precision convert from secs to millis now, otherwise the
		// thousandths place is dropped downstream.
		val := float64(s.Value) * 1000.0

		// It is possible to get a NaN if there is no traffic (or possibly other reasons). Just skip it
		if math.IsNaN(val) {
			continue
		}

		if a.InjectServiceNodes {
			// don't inject a service node if the dest node is already a service node.  Also, we can't inject if destSvcName is not set.
			_, destNodeType := graph.Id(destSvcNs, destWl, destApp, destVer, destSvcName, a.GraphType)
			if destSvcNameOk && destNodeType != graph.NodeTypeService {
				// Do not set response time on the incoming edge, we can't validly aggregate response times of the outgoing edges (kiali-2297)
				a.addResponseTime(responseTimeMap, val, destSvcNs, "", "", "", destSvcName, destSvcNs, destWl, destApp, destVer, destSvcName)
			} else {
				a.addResponseTime(responseTimeMap, val, sourceWlNs, sourceWl, sourceApp, sourceVer, "", destSvcNs, destWl, destApp, destVer, destSvcName)
			}
		} else {
			a.addResponseTime(responseTimeMap, val, sourceWlNs, sourceWl, sourceApp, sourceVer, "", destSvcNs, destWl, destApp, destVer, destSvcName)
		}
	}
}

func (a ResponseTimeAppender) addResponseTime(responseTimeMap map[string]float64, val float64, sourceWlNs, sourceWl, sourceApp, sourceVer, sourceSvcName, destSvcNs, destWl, destApp, destVer, destSvcName string) {
	sourceId, _ := graph.Id(sourceWlNs, sourceWl, sourceApp, sourceVer, sourceSvcName, a.GraphType)
	destId, _ := graph.Id(destSvcNs, destWl, destApp, destVer, destSvcName, a.GraphType)
	key := fmt.Sprintf("%s %s", sourceId, destId)

	responseTimeMap[key] = val
}
