package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type Rule struct {
	RuleId           string    `gorm:"column:rule_id" json:"rule_id"`
	RuleName         string    `gorm:"column:rule_name" json:"rule_name"`
	Disabled         bool      `gorm:"column:disabled" json:"disabled"`
	MonitorPeriods   uint32    `gorm:"column:monitor_periods" json:"monitor_periods"`
	Severity         string    `gorm:"column:severity" json:"severity"`
	MetricsType      string    `gorm:"column:metrics_type" json:"metrics_type"`
	ConditionType    string    `gorm:"column:condition_type" json:"condition_type"`
	Thresholds       string    `gorm:"column:thresholds" json:"thresholds"`
	Unit             string    `gorm:"column:unit" json:"unit"`
	ConsecutiveCount uint32    `gorm:"column:consecutive_count" json:"consecutive_count"`
	Inhibit          bool      `gorm:"column:inhibit" json:"inhibit"`
	CreateTime       time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime       time.Time `gorm:"column:update_time" json:"update_time"`
	PolicyId         string    `gorm:"column:policy_id" json:"policy_id"`
	MetricId         string    `gorm:"column:metric_id" json:"metric_id"`
}

//table name
const (
	TableRule = "rule"
)

const (
	RuleIdPrefix = "rl-"
)

//field name
//Rl is short for rule.
const (
	RlColId               = "rule_id"
	RlColName             = "rule_name"
	RlColDisabled         = "disabled"
	RlColMonitorPeriods   = "monitor_periods"
	RlColSeverity         = "severity"
	RlColMetricsType      = "metrics_type"
	RlColConditionType    = "condition_type"
	RlColThresholds       = "thresholds"
	RlColUnit             = "unit"
	RlColConsecutiveCount = "consecutive_count"
	RlColInhibit          = "inhibit"
	RlColCreateTime       = "create_time"
	RlColUpdateTime       = "update_time"
	RlColPolicyId         = "policy_id"
	RlColMetricId         = "metric_id"
)

func NewRuleId() string {
	return idutils.GetUuid(RuleIdPrefix)
}

func NewRule(ruleName string, disabled bool, monitorPeriods uint32, severity string, metricsType string, conditionType string, thresholds string, unit string, consecutiveCount uint32, inhibit bool, policyId string, metricId string) *Rule {
	rule := &Rule{
		RuleId:           NewRuleId(),
		RuleName:         ruleName,
		Disabled:         disabled,
		MonitorPeriods:   monitorPeriods,
		Severity:         severity,
		MetricsType:      metricsType,
		ConditionType:    conditionType,
		Thresholds:       thresholds,
		Unit:             unit,
		ConsecutiveCount: consecutiveCount,
		Inhibit:          inhibit,
		CreateTime:       time.Now(),
		UpdateTime:       time.Now(),
		PolicyId:         policyId,
		MetricId:         metricId,
	}
	return rule
}

func RuleToPb(rule *Rule) *pb.Rule {
	pbRule := pb.Rule{}
	pbRule.RuleId = rule.RuleId
	pbRule.RuleName = rule.RuleName
	pbRule.Disabled = rule.Disabled
	pbRule.MonitorPeriods = rule.MonitorPeriods
	pbRule.Severity = rule.Severity
	pbRule.MetricsType = rule.MetricsType
	pbRule.ConditionType = rule.ConditionType
	pbRule.Thresholds = rule.Thresholds
	pbRule.Unit = rule.Unit
	pbRule.ConsecutiveCount = rule.ConsecutiveCount
	pbRule.Inhibit = rule.Inhibit
	pbRule.CreateTime = pbutil.ToProtoTimestamp(rule.CreateTime)
	pbRule.UpdateTime = pbutil.ToProtoTimestamp(rule.UpdateTime)
	pbRule.PolicyId = rule.PolicyId
	pbRule.MetricId = rule.MetricId
	return &pbRule
}

func ParseRlSet2PbSet(inRls []*Rule) []*pb.Rule {
	var pbRls []*pb.Rule
	for _, inRl := range inRls {
		pbRl := RuleToPb(inRl)
		pbRls = append(pbRls, pbRl)
	}
	return pbRls
}

type RuleDetail struct {
	RuleId           string    `gorm:"column:rule_id" json:"rule_id"`
	RuleName         string    `gorm:"column:rule_name" json:"rule_name"`
	Disabled         bool      `gorm:"column:disabled" json:"disabled"`
	MonitorPeriods   uint32    `gorm:"column:monitor_periods" json:"monitor_periods"`
	Severity         string    `gorm:"column:severity" json:"severity"`
	MetricsType      string    `gorm:"column:metrics_type" json:"metrics_type"`
	ConditionType    string    `gorm:"column:condition_type" json:"condition_type"`
	Thresholds       string    `gorm:"column:thresholds" json:"thresholds"`
	MetricParam      string    `gorm:"column:metric_param" json:"metric_param"`
	Unit             string    `gorm:"column:unit" json:"unit"`
	ConsecutiveCount uint32    `gorm:"column:consecutive_count" json:"consecutive_count"`
	Inhibit          bool      `gorm:"column:inhibit" json:"inhibit"`
	CreateTime       time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime       time.Time `gorm:"column:update_time" json:"update_time"`
	PolicyId         string    `gorm:"column:policy_id" json:"policy_id"`
	MetricId         string    `gorm:"column:metric_id" json:"metric_id"`
}
