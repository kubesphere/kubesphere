package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type AlertDetail struct {
	AlertId             string    `gorm:"column:alert_id" json:"alert_id"`
	AlertName           string    `gorm:"column:alert_name" json:"alert_name"`
	Disabled            bool      `gorm:"column:disabled" json:"disabled"`
	CreateTime          time.Time `gorm:"column:create_time" json:"create_time"`
	RunningStatus       string    `gorm:"column:running_status" json:"running_status"`
	AlertStatus         string    `gorm:"column:alert_status" json:"alert_status"`
	PolicyId            string    `gorm:"column:policy_id" json:"policy_id"`
	RsFilterName        string    `gorm:"column:rs_filter_name" json:"rs_filter_name"`
	RsFilterParam       string    `gorm:"column:rs_filter_param" json:"rs_filter_param"`
	RsTypeName          string    `gorm:"column:rs_type_name" json:"rs_type_name"`
	ExecutorId          string    `gorm:"column:executor_id" json:"executor_id"`
	PolicyName          string    `gorm:"column:policy_name" json:"policy_name"`
	PolicyDescription   string    `gorm:"column:policy_description" json:"policy_description"`
	PolicyConfig        string    `gorm:"column:policy_config" json:"policy_config"`
	Creator             string    `gorm:"column:creator" json:"creator"`
	AvailableStartTime  string    `gorm:"column:available_start_time" json:"available_start_time"`
	AvailableEndTime    string    `gorm:"column:available_end_time" json:"available_end_time"`
	Metrics             []string  `gorm:"column:metrics" json:"metrics"`
	RulesCount          uint32    `json:"rules_count"`
	PositivesCount      uint32    `json:"positives_count"`
	MostRecentAlertTime string    `json:"most_recent_alert_time"`
	NfAddressListId     string    `gorm:"column:nf_address_list_id" json:"nf_address_list_id"`
}

func AlertDetailToPb(alertDetail *AlertDetail) *pb.AlertDetail {
	pbAlertDetail := pb.AlertDetail{}
	pbAlertDetail.AlertId = alertDetail.AlertId
	pbAlertDetail.AlertName = alertDetail.AlertName
	pbAlertDetail.Disabled = alertDetail.Disabled
	pbAlertDetail.CreateTime = pbutil.ToProtoTimestamp(alertDetail.CreateTime)
	pbAlertDetail.RunningStatus = alertDetail.RunningStatus
	pbAlertDetail.AlertStatus = alertDetail.AlertStatus
	pbAlertDetail.PolicyId = alertDetail.PolicyId
	pbAlertDetail.RsFilterName = alertDetail.RsFilterName
	pbAlertDetail.RsFilterParam = alertDetail.RsFilterParam
	pbAlertDetail.RsTypeName = alertDetail.RsTypeName
	pbAlertDetail.ExecutorId = alertDetail.ExecutorId
	pbAlertDetail.PolicyName = alertDetail.PolicyName
	pbAlertDetail.PolicyDescription = alertDetail.PolicyDescription
	pbAlertDetail.PolicyConfig = alertDetail.PolicyConfig
	pbAlertDetail.Creator = alertDetail.Creator
	pbAlertDetail.AvailableStartTime = alertDetail.AvailableStartTime
	pbAlertDetail.AvailableEndTime = alertDetail.AvailableEndTime
	pbAlertDetail.Metrics = alertDetail.Metrics
	pbAlertDetail.RulesCount = alertDetail.RulesCount
	pbAlertDetail.PositivesCount = alertDetail.PositivesCount
	pbAlertDetail.MostRecentAlertTime = alertDetail.MostRecentAlertTime
	pbAlertDetail.NfAddressListId = alertDetail.NfAddressListId
	return &pbAlertDetail
}

func ParseAldSet2PbSet(inAlds []*AlertDetail) []*pb.AlertDetail {
	var pbAlds []*pb.AlertDetail
	for _, inAld := range inAlds {
		pbAld := AlertDetailToPb(inAld)
		pbAlds = append(pbAlds, pbAld)
	}
	return pbAlds
}

type ResourceStatus struct {
	ResourceName       string `gorm:"column:resource_name" json:"resource_name"`
	CurrentLevel       string `json:"current_level"`
	PositiveCount      uint32 `json:"positive_count"`
	CumulatedSendCount uint32 `json:"cumulated_send_count"`
	NextResendInterval uint32 `json:"next_resend_interval"`
	NextSendableTime   string `json:"next_sendable_time"`
	AggregatedAlerts   string `json:"aggregated_alerts"`
}

type AlertStatus struct {
	RuleId           string           `gorm:"column:rule_id" json:"rule_id"`
	RuleName         string           `gorm:"column:rule_name" json:"rule_name"`
	Disabled         bool             `gorm:"column:disabled" json:"disabled"`
	MonitorPeriods   uint32           `gorm:"column:monitor_periods" json:"monitor_periods"`
	Severity         string           `gorm:"column:severity" json:"severity"`
	MetricsType      string           `gorm:"column:metrics_type" json:"metrics_type"`
	ConditionType    string           `gorm:"column:condition_type" json:"condition_type"`
	Thresholds       string           `gorm:"column:thresholds" json:"thresholds"`
	Unit             string           `gorm:"column:unit" json:"unit"`
	ConsecutiveCount uint32           `gorm:"column:consecutive_count" json:"consecutive_count"`
	Inhibit          bool             `gorm:"column:inhibit" json:"inhibit"`
	MetricName       string           `gorm:"column:metric_name" json:"metric_name"`
	Resources        []ResourceStatus `gorm:"column:resources" json:"resources"`
	CreateTime       time.Time        `gorm:"column:create_time" json:"create_time"`
	UpdateTime       time.Time        `gorm:"column:update_time" json:"update_time"`
	AlertStatus      string           `gorm:"column:alert_status"`
}

func AlertStatusToPb(alertStatus AlertStatus) *pb.AlertStatus {
	pbAlertStatus := pb.AlertStatus{}
	pbAlertStatus.RuleId = alertStatus.RuleId
	pbAlertStatus.RuleName = alertStatus.RuleName
	pbAlertStatus.MetricName = alertStatus.MetricName
	pbAlertStatus.Disabled = alertStatus.Disabled
	pbAlertStatus.MonitorPeriods = alertStatus.MonitorPeriods
	pbAlertStatus.Severity = alertStatus.Severity
	pbAlertStatus.MetricsType = alertStatus.MetricsType
	pbAlertStatus.ConditionType = alertStatus.ConditionType
	pbAlertStatus.Thresholds = alertStatus.Thresholds
	pbAlertStatus.Unit = alertStatus.Unit
	pbAlertStatus.ConsecutiveCount = alertStatus.ConsecutiveCount
	pbAlertStatus.Inhibit = alertStatus.Inhibit
	for _, resource := range alertStatus.Resources {
		pbResource := pb.ResourceStatus{}
		pbResource.ResourceName = resource.ResourceName
		pbResource.CurrentLevel = resource.CurrentLevel
		pbResource.PositiveCount = resource.PositiveCount
		pbResource.CumulatedSendCount = resource.CumulatedSendCount
		pbResource.NextResendInterval = resource.NextResendInterval
		pbResource.NextSendableTime = resource.NextSendableTime
		pbResource.AggregatedAlerts = resource.AggregatedAlerts

		pbAlertStatus.Resources = append(pbAlertStatus.Resources, &pbResource)
	}
	pbAlertStatus.CreateTime = pbutil.ToProtoTimestamp(alertStatus.CreateTime)
	pbAlertStatus.UpdateTime = pbutil.ToProtoTimestamp(alertStatus.UpdateTime)
	return &pbAlertStatus
}

func ParseAlsSet2PbSet(inAlss []AlertStatus) []*pb.AlertStatus {
	var pbAlss []*pb.AlertStatus
	for _, inAls := range inAlss {
		pbAls := AlertStatusToPb(inAls)
		pbAlss = append(pbAlss, pbAls)
	}
	return pbAlss
}
