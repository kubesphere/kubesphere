package appender

import (
	"context"
	"fmt"
	"time"

	"github.com/kiali/kiali/graph"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/prometheus/internalmetrics"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// package-private util functions (used by multiple files)

func promQuery(query string, queryTime time.Time, api v1.API, a Appender) model.Vector {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// wrap with a round() to be in line with metrics api
	query = fmt.Sprintf("round(%s,0.001)", query)
	log.Debugf("Appender query:\n%s&time=%v (now=%v, %v)\n", query, queryTime.Format(graph.TF), time.Now().Format(graph.TF), queryTime.Unix())

	promtimer := internalmetrics.GetPrometheusProcessingTimePrometheusTimer("Graph-Appender-" + a.Name())
	value, _, err := api.Query(ctx, query, queryTime)
	graph.CheckError(err)
	promtimer.ObserveDuration() // notice we only collect metrics for successful prom queries

	switch t := value.Type(); t {
	case model.ValVector: // Instant Vector
		return value.(model.Vector)
	default:
		graph.Error(fmt.Sprintf("No handling for type %v!\n", t))
	}

	return nil
}
