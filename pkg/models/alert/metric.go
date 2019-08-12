package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type Metric struct {
	MetricId    string    `gorm:"column:metric_id" json:"metric_id"`
	MetricName  string    `gorm:"column:metric_name" json:"metric_name"`
	MetricParam string    `gorm:"column:metric_param" json:"metric_param"`
	Status      string    `gorm:"column:status" json:"status"`
	RsTypeId    string    `gorm:"column:rs_type_id" json:"rs_type_id"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"update_time"`
}

//table name
const (
	TableMetric = "metric"
)

const (
	MetricIdPrefix = "mt-"
)

//field name
//Mt is short for metric.
const (
	MtColId         = "metric_id"
	MtColName       = "metric_name"
	MtColParam      = "metric_param"
	MtColStatus     = "status"
	MtColCreateTime = "create_time"
	MtColUpdateTime = "update_time"
	MtColTypeId     = "rs_type_id"
)

func NewMetricId() string {
	return idutils.GetUuid(MetricIdPrefix)
}

func NewMetric(metricName string, metricParam string, rsTypeId string) *Metric {
	metric := &Metric{
		MetricId:    NewMetricId(),
		MetricName:  metricName,
		MetricParam: metricParam,
		Status:      "active",
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		RsTypeId:    rsTypeId,
	}
	return metric
}

func MetricToPb(metric *Metric) *pb.Metric {
	pbMetric := pb.Metric{}
	pbMetric.MetricId = metric.MetricId
	pbMetric.MetricName = metric.MetricName
	pbMetric.MetricParam = metric.MetricParam
	pbMetric.Status = metric.Status
	pbMetric.CreateTime = pbutil.ToProtoTimestamp(metric.CreateTime)
	pbMetric.UpdateTime = pbutil.ToProtoTimestamp(metric.UpdateTime)
	pbMetric.RsTypeId = metric.RsTypeId
	return &pbMetric
}

func ParseMtSet2PbSet(inMts []*Metric) []*pb.Metric {
	var pbMts []*pb.Metric
	for _, inMt := range inMts {
		pbMt := MetricToPb(inMt)
		pbMts = append(pbMts, pbMt)
	}
	return pbMts
}
