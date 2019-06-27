package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type ResourceType struct {
	RsTypeId    string    `gorm:"column:rs_type_id" json:"rs_type_id"`
	RsTypeName  string    `gorm:"column:rs_type_name" json:"rs_type_name"`
	RsTypeParam string    `gorm:"column:rs_type_param" json:"rs_type_param"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"update_time"`
}

//table name
const (
	TableResourceType = "resource_type"
)

const (
	ResourceTypeIdPrefix = "rst-"
)

//field name
//Rt is short for resource_type.
const (
	RtColId         = "rs_type_id"
	RtColName       = "rs_type_name"
	RtColParam      = "rs_type_param"
	RtColCreateTime = "create_time"
	RtColUpdateTime = "update_time"
)

func NewResourceTypeId() string {
	return idutils.GetUuid(ResourceTypeIdPrefix)
}

func NewResourceType(rsTypeName string, rsUriTmpl string) *ResourceType {
	resourceType := &ResourceType{
		RsTypeId:    NewResourceTypeId(),
		RsTypeName:  rsTypeName,
		RsTypeParam: rsUriTmpl,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	return resourceType
}

func ResourceTypeToPb(resourceType *ResourceType) *pb.ResourceType {
	pbResourceType := pb.ResourceType{}
	pbResourceType.RsTypeId = resourceType.RsTypeId
	pbResourceType.RsTypeName = resourceType.RsTypeName
	pbResourceType.RsTypeParam = resourceType.RsTypeParam
	pbResourceType.CreateTime = pbutil.ToProtoTimestamp(resourceType.CreateTime)
	pbResourceType.UpdateTime = pbutil.ToProtoTimestamp(resourceType.UpdateTime)
	return &pbResourceType
}

func ParseRtSet2PbSet(inRts []*ResourceType) []*pb.ResourceType {
	var pbRts []*pb.ResourceType
	for _, inRt := range inRts {
		pbRt := ResourceTypeToPb(inRt)
		pbRts = append(pbRts, pbRt)
	}
	return pbRts
}
