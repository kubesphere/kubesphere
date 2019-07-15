package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type Action struct {
	ActionId        string    `gorm:"column:action_id" json:"action_id"`
	ActionName      string    `gorm:"column:action_name" json:"action_name"`
	TriggerStatus   string    `gorm:"column:trigger_status" json:"trigger_status"`
	TriggerAction   string    `gorm:"column:trigger_action" json:"trigger_action"`
	CreateTime      time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime      time.Time `gorm:"column:update_time" json:"update_time"`
	PolicyId        string    `gorm:"column:policy_id" json:"policy_id"`
	NfAddressListId string    `gorm:"column:nf_address_list_id" json:"nf_address_list_id"`
}

//table name
const (
	TableAction = "action"
)

const (
	ActionIdPrefix = "ac-"
)

//field name
//Ac is short for action.
const (
	AcColId              = "action_id"
	AcColName            = "action_name"
	AcColTriggerStatus   = "trigger_status"
	AcColTriggerAction   = "trigger_action"
	AcColCreateTime      = "create_time"
	AcColUpdateTime      = "update_time"
	AcColPolicyId        = "policy_id"
	AcColNfAddressListId = "nf_address_list_id"
)

func NewActionId() string {
	return idutils.GetUuid(ActionIdPrefix)
}

func NewAction(actionName string, triggerStatus string, triggerAction string, policyId string, nfAddressListId string) *Action {
	action := &Action{
		ActionId:        NewActionId(),
		ActionName:      actionName,
		TriggerStatus:   triggerStatus,
		TriggerAction:   triggerAction,
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
		PolicyId:        policyId,
		NfAddressListId: nfAddressListId,
	}
	return action
}

func ActionToPb(action *Action) *pb.Action {
	pbAction := pb.Action{}
	pbAction.ActionId = action.ActionId
	pbAction.ActionName = action.ActionName
	pbAction.TriggerStatus = action.TriggerStatus
	pbAction.TriggerAction = action.TriggerAction
	pbAction.CreateTime = pbutil.ToProtoTimestamp(action.CreateTime)
	pbAction.UpdateTime = pbutil.ToProtoTimestamp(action.UpdateTime)
	pbAction.PolicyId = action.PolicyId
	pbAction.NfAddressListId = action.NfAddressListId
	return &pbAction
}

func ParseAcSet2PbSet(inAcs []*Action) []*pb.Action {
	var pbAcs []*pb.Action
	for _, inAc := range inAcs {
		pbAc := ActionToPb(inAc)
		pbAcs = append(pbAcs, pbAc)
	}
	return pbAcs
}
