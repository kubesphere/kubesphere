package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type Policy struct {
	PolicyId           string    `gorm:"column:policy_id" json:"policy_id"`
	PolicyName         string    `gorm:"column:policy_name" json:"policy_name"`
	PolicyDescription  string    `gorm:"column:policy_description" json:"policy_description"`
	PolicyConfig       string    `gorm:"column:policy_config" json:"policy_config"`
	Creator            string    `gorm:"column:creator" json:"creator"`
	AvailableStartTime string    `gorm:"column:available_start_time" json:"available_start_time"`
	AvailableEndTime   string    `gorm:"column:available_end_time" json:"available_end_time"`
	CreateTime         time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime         time.Time `gorm:"column:update_time" json:"update_time"`
	RsTypeId           string    `gorm:"column:rs_type_id" json:"rs_type_id"`
}

//table name
const (
	TablePolicy = "policy"
)

const (
	PolicyIdPrefix = "pl-"
)

//field name
//Pl is short for policy.
const (
	PlColId                 = "policy_id"
	PlColName               = "policy_name"
	PlColDescription        = "policy_description"
	PlColConfig             = "policy_config"
	PlColCreator            = "creator"
	PlColAvailableStartTime = "available_start_time"
	PlColAvailableEndTime   = "available_end_time"
	PlColCreateTime         = "create_time"
	PlColUpdateTime         = "update_time"
	PlColTypeId             = "rs_type_id"
)

func NewPolicyId() string {
	return idutils.GetUuid(PolicyIdPrefix)
}

func NewPolicy(policyName string, policyDescription string, policyConfig string, creator string, availableStartTime string, availableEndTime string, rsTypeId string) *Policy {
	policy := &Policy{
		PolicyId:           NewPolicyId(),
		PolicyName:         policyName,
		PolicyDescription:  policyDescription,
		PolicyConfig:       policyConfig,
		Creator:            creator,
		AvailableStartTime: availableStartTime,
		AvailableEndTime:   availableEndTime,
		CreateTime:         time.Now(),
		UpdateTime:         time.Now(),
		RsTypeId:           rsTypeId,
	}
	return policy
}

func PolicyToPb(policy *Policy) *pb.Policy {
	pbPolicy := pb.Policy{}
	pbPolicy.PolicyId = policy.PolicyId
	pbPolicy.PolicyName = policy.PolicyName
	pbPolicy.PolicyDescription = policy.PolicyDescription
	pbPolicy.PolicyConfig = policy.PolicyConfig
	pbPolicy.Creator = policy.Creator
	pbPolicy.AvailableStartTime = policy.AvailableStartTime
	pbPolicy.AvailableEndTime = policy.AvailableEndTime
	pbPolicy.CreateTime = pbutil.ToProtoTimestamp(policy.CreateTime)
	pbPolicy.UpdateTime = pbutil.ToProtoTimestamp(policy.UpdateTime)
	pbPolicy.RsTypeId = policy.RsTypeId
	return &pbPolicy
}

func ParsePlSet2PbSet(inPls []*Policy) []*pb.Policy {
	var pbPls []*pb.Policy
	for _, inPl := range inPls {
		pbPl := PolicyToPb(inPl)
		pbPls = append(pbPls, pbPl)
	}
	return pbPls
}

type PolicyByAlert struct {
	AlertName          string    `gorm:"column:alert_name" json:"alert_name"`
	PolicyName         string    `gorm:"column:policy_name" json:"policy_name"`
	PolicyDescription  string    `gorm:"column:policy_description" json:"policy_description"`
	PolicyConfig       string    `gorm:"column:policy_config" json:"policy_config"`
	Creator            string    `gorm:"column:creator" json:"creator"`
	AvailableStartTime string    `gorm:"column:available_start_time" json:"available_start_time"`
	AvailableEndTime   string    `gorm:"column:available_end_time" json:"available_end_time"`
	CreateTime         time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime         time.Time `gorm:"column:update_time" json:"update_time"`
	RsTypeId           string    `gorm:"column:rs_type_id" json:"rs_type_id"`
}
