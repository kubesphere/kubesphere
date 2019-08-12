package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type ResourceFilter struct {
	RsFilterId    string    `gorm:"column:rs_filter_id" json:"rs_filter_id"`
	RsFilterName  string    `gorm:"column:rs_filter_name" json:"rs_filter_name"`
	RsFilterParam string    `gorm:"column:rs_filter_param" json:"rs_filter_param"`
	Status        string    `gorm:"column:status" json:"status"`
	CreateTime    time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime    time.Time `gorm:"column:update_time" json:"update_time"`
	RsTypeId      string    `gorm:"column:rs_type_id" json:"rs_type_id"`
}

//table name
const (
	TableResourceFilter = "resource_filter"
)

const (
	RsFilterIdPrefix = "rf-"
)

//field name
//Rf is short for resource filter.
const (
	RfColId         = "rs_filter_id"
	RfColName       = "rs_filter_name"
	RfColParam      = "rs_filter_param"
	RfColStatus     = "status"
	RfColCreateTime = "create_time"
	RfColUpdateTime = "update_time"
	RfColTypeId     = "rs_type_id"
)

func NewRsFilterId() string {
	return idutils.GetUuid(RsFilterIdPrefix)
}

func NewResourceFilter(rsFilterName string, rsFilterUri string, rsTypeId string) *ResourceFilter {
	rsFilter := &ResourceFilter{
		RsFilterId:    NewRsFilterId(),
		RsFilterName:  rsFilterName,
		RsFilterParam: rsFilterUri,
		Status:        "active",
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		RsTypeId:      rsTypeId,
	}
	return rsFilter
}

func ResourceFilterToPb(rsFilter *ResourceFilter) *pb.ResourceFilter {
	pbResourceFilter := pb.ResourceFilter{}
	pbResourceFilter.RsFilterId = rsFilter.RsFilterId
	pbResourceFilter.RsFilterName = rsFilter.RsFilterName
	pbResourceFilter.RsFilterParam = rsFilter.RsFilterParam
	pbResourceFilter.Status = rsFilter.Status
	pbResourceFilter.CreateTime = pbutil.ToProtoTimestamp(rsFilter.CreateTime)
	pbResourceFilter.UpdateTime = pbutil.ToProtoTimestamp(rsFilter.UpdateTime)
	pbResourceFilter.RsTypeId = rsFilter.RsTypeId
	return &pbResourceFilter
}

func ParseRfSet2PbSet(inRfs []*ResourceFilter) []*pb.ResourceFilter {
	var pbRfs []*pb.ResourceFilter
	for _, inRf := range inRfs {
		pbRf := ResourceFilterToPb(inRf)
		pbRfs = append(pbRfs, pbRf)
	}
	return pbRfs
}
