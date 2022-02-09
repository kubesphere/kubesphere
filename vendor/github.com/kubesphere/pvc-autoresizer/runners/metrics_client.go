package runners

import (
	"context"
	"fmt"
	"time"

	"github.com/kubesphere/pvc-autoresizer/metrics"
	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/types"
)

const (
	volumeAvailableQuery = "kubelet_volume_stats_available_bytes"
	volumeCapacityQuery  = "kubelet_volume_stats_capacity_bytes"
	inodesAvailableQuery = "kubelet_volume_stats_inodes_free"
	inodesCapacityQuery  = "kubelet_volume_stats_inodes"
)

// NewPrometheusClient returns a new prometheusClient
func NewPrometheusClient(url string) (MetricsClient, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, err
	}
	v1api := prometheusv1.NewAPI(client)

	return &prometheusClient{
		prometheusAPI: v1api,
	}, nil
}

// MetricsClient is an interface for getting metrics
type MetricsClient interface {
	GetMetrics(ctx context.Context) (map[types.NamespacedName]*VolumeStats, error)
}

// VolumeStats is a struct containing metrics used by pvc-autoresizer
type VolumeStats struct {
	AvailableBytes     int64
	CapacityBytes      int64
	AvailableInodeSize int64
	CapacityInodeSize  int64
}

type prometheusClient struct {
	prometheusAPI prometheusv1.API
}

func (c *prometheusClient) GetMetrics(ctx context.Context) (map[types.NamespacedName]*VolumeStats, error) {
	volumeStatsMap := make(map[types.NamespacedName]*VolumeStats)

	availableBytes, err := c.getMetricValues(ctx, volumeAvailableQuery)
	if err != nil {
		return nil, err
	}

	capacityBytes, err := c.getMetricValues(ctx, volumeCapacityQuery)
	if err != nil {
		return nil, err
	}

	availableInodeSize, err := c.getMetricValues(ctx, inodesAvailableQuery)
	if err != nil {
		return nil, err
	}

	capacityInodeSize, err := c.getMetricValues(ctx, inodesCapacityQuery)
	if err != nil {
		return nil, err
	}

	for key, val := range availableBytes {
		vs := &VolumeStats{AvailableBytes: val}
		if cb, ok := capacityBytes[key]; !ok {
			continue
		} else {
			vs.CapacityBytes = cb
		}
		if ais, ok := availableInodeSize[key]; ok {
			vs.AvailableInodeSize = ais
		}
		if cis, ok := capacityInodeSize[key]; ok {
			vs.CapacityInodeSize = cis
		}
		volumeStatsMap[key] = vs
	}

	return volumeStatsMap, nil
}

func (c *prometheusClient) getMetricValues(ctx context.Context, query string) (map[types.NamespacedName]int64, error) {
	res, _, err := c.prometheusAPI.Query(ctx, query, time.Now())
	if err != nil {
		metrics.MetricsClientFailTotal.Increment()
		return nil, err
	}

	if res.Type() != model.ValVector {
		return nil, fmt.Errorf("unknown response type: %s", res.Type().String())
	}
	resultMap := make(map[types.NamespacedName]int64)
	vec := res.(model.Vector)
	for _, val := range vec {
		nn := types.NamespacedName{
			Namespace: string(val.Metric["namespace"]),
			Name:      string(val.Metric["persistentvolumeclaim"]),
		}
		resultMap[nn] = int64(val.Value)
	}
	return resultMap, nil
}
