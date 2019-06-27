package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type AlertInfo struct {
	RsFilter ResourceFilter `json:"resource_filter"`
	Policy   Policy         `json:"policy"`
	Rules    []Rule         `json:"rules"`
	Action   Action         `json:"action"`
	Alert    Alert          `json:"alert"`
}

type Alert struct {
	AlertId       string    `gorm:"column:alert_id" json:"alert_id"`
	AlertName     string    `gorm:"column:alert_name" json:"alert_name"`
	Disabled      bool      `gorm:"column:disabled" json:"disabled"`
	RunningStatus string    `gorm:"column:running_status" json:"running_status"`
	AlertStatus   string    `gorm:"column:alert_status" json:"alert_status"`
	CreateTime    time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime    time.Time `gorm:"column:update_time" json:"update_time"`
	PolicyId      string    `gorm:"column:policy_id" json:"policy_id"`
	RsFilterId    string    `gorm:"column:rs_filter_id" json:"rs_filter_id"`
	ExecutorId    string    `gorm:"column:executor_id" json:"executor_id"`
}

//table name
const (
	TableAlert = "alert"
)

const (
	AlertIdPrefix = "al-"
)

//field name
//Al is short for alert.
const (
	AlColId            = "alert_id"
	AlColName          = "alert_name"
	AlColDisabled      = "disabled"
	AlColRunningStatus = "running_status"
	AlColAlertStatus   = "alert_status"
	AlColCreateTime    = "create_time"
	AlColUpdateTime    = "update_time"
	AlColPolicyId      = "policy_id"
	AlColRsFilterId    = "rs_filter_id"
	AlColExecutorId    = "executor_id"
)

func NewAlertId() string {
	return idutils.GetUuid(AlertIdPrefix)
}

func NewAlert(alertName string, disabled bool, runningStatus string, alertStatus string, policyId string, rsFilterId string) *Alert {
	alert := &Alert{
		AlertId:       NewAlertId(),
		AlertName:     alertName,
		Disabled:      false,
		RunningStatus: runningStatus,
		AlertStatus:   alertStatus,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		PolicyId:      policyId,
		RsFilterId:    rsFilterId,
		ExecutorId:    "",
	}
	return alert
}

func AlertToPb(alert *Alert) *pb.Alert {
	pbAlert := pb.Alert{}
	pbAlert.AlertId = alert.AlertId
	pbAlert.AlertName = alert.AlertName
	pbAlert.Disabled = alert.Disabled
	pbAlert.RunningStatus = alert.RunningStatus
	pbAlert.AlertStatus = alert.AlertStatus
	pbAlert.CreateTime = pbutil.ToProtoTimestamp(alert.CreateTime)
	pbAlert.UpdateTime = pbutil.ToProtoTimestamp(alert.UpdateTime)
	pbAlert.PolicyId = alert.PolicyId
	pbAlert.RsFilterId = alert.RsFilterId
	pbAlert.ExecutorId = alert.ExecutorId
	return &pbAlert
}

func ParseAlSet2PbSet(inAls []*Alert) []*pb.Alert {
	var pbAls []*pb.Alert
	for _, inAl := range inAls {
		pbAl := AlertToPb(inAl)
		pbAls = append(pbAls, pbAl)
	}
	return pbAls
}
