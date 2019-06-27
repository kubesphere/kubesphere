package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type History struct {
	HistoryId      string    `gorm:"column:history_id" json:"history_id"`
	HistoryName    string    `gorm:"column:history_name" json:"history_name"`
	Event          string    `gorm:"column:event" json:"event"`
	Content        string    `gorm:"column:content" json:"content"`
	NotificationId string    `gorm:"column:notification_id" json:"notification_id"`
	CreateTime     time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime     time.Time `gorm:"column:update_time" json:"update_time"`
	AlertId        string    `gorm:"column:alert_id" json:"alert_id"`
	RuleId         string    `gorm:"column:rule_id" json:"rule_id"`
	ResourceName   string    `gorm:"column:resource_name" json:"resource_name"`
}

//table name
const (
	TableHistory = "history"
)

const (
	HistoryIdPrefix = "hs-"
)

//field name
//Hs is short for history.
const (
	HsColId             = "history_id"
	HsColName           = "history_name"
	HsColEvent          = "event"
	HsColContent        = "content"
	HsColNotificationId = "notification_id"
	HsColCreateTime     = "create_time"
	HsColUpdateTime     = "update_time"
	HsColAlertId        = "alert_id"
	HsColRuleId         = "rule_id"
	HsColResourceName   = "resource_name"
)

func NewHistoryId() string {
	return idutils.GetUuid(HistoryIdPrefix)
}

func NewHistory(historyName string, event string, content string, notificationId string, alertId string, ruleId string, resourceName string) *History {
	history := &History{
		HistoryId:      NewHistoryId(),
		HistoryName:    historyName,
		Event:          event,
		Content:        content,
		NotificationId: notificationId,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
		AlertId:        alertId,
		RuleId:         ruleId,
		ResourceName:   resourceName,
	}
	return history
}

func HistoryToPb(history *History) *pb.History {
	pbHistory := pb.History{}
	pbHistory.HistoryId = history.HistoryId
	pbHistory.HistoryName = history.HistoryName
	pbHistory.Event = history.Event
	pbHistory.Content = history.Content
	pbHistory.NotificationId = history.NotificationId
	pbHistory.CreateTime = pbutil.ToProtoTimestamp(history.CreateTime)
	pbHistory.UpdateTime = pbutil.ToProtoTimestamp(history.UpdateTime)
	pbHistory.AlertId = history.AlertId
	pbHistory.RuleId = history.RuleId
	pbHistory.ResourceName = history.ResourceName
	return &pbHistory
}

func ParseHsSet2PbSet(inHss []*History) []*pb.History {
	var pbHss []*pb.History
	for _, inHs := range inHss {
		pbHs := HistoryToPb(inHs)
		pbHss = append(pbHss, pbHs)
	}
	return pbHss
}
