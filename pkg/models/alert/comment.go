package models

import (
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"kubesphere.io/kubesphere/pkg/utils/pbutil"
)

type Comment struct {
	CommentId  string    `gorm:"column:comment_id" json:"comment_id"`
	Addresser  string    `gorm:"column:addresser" json:"addresser"`
	Content    string    `gorm:"column:content" json:"content"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time" json:"update_time"`
	HistoryId  string    `gorm:"column:history_id" json:"history_id"`
}

//table name
const (
	TableComment = "comment"
)

const (
	CommentIdPrefix = "cm-"
)

//field name
//Cm is short for comment.
const (
	CmColId         = "comment_id"
	CmColAddresser  = "addresser"
	CmColContent    = "content"
	CmColCreateTime = "create_time"
	CmColUpdateTime = "update_time"
	CmColHistoryId  = "history_id"
)

func NewCommentId() string {
	return idutils.GetUuid(CommentIdPrefix)
}

func NewComment(addresser string, content string, historyId string) *Comment {
	comment := &Comment{
		CommentId:  NewCommentId(),
		Addresser:  addresser,
		Content:    content,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		HistoryId:  historyId,
	}
	return comment
}

func CommentToPb(comment *Comment) *pb.Comment {
	pbComment := pb.Comment{}
	pbComment.CommentId = comment.CommentId
	pbComment.Addresser = comment.Addresser
	pbComment.Content = comment.Content
	pbComment.CreateTime = pbutil.ToProtoTimestamp(comment.CreateTime)
	pbComment.UpdateTime = pbutil.ToProtoTimestamp(comment.UpdateTime)
	pbComment.HistoryId = comment.HistoryId
	return &pbComment
}

func ParseCmtSet2PbSet(inCmts []*Comment) []*pb.Comment {
	var pbCmts []*pb.Comment
	for _, inCmt := range inCmts {
		pbCmt := CommentToPb(inCmt)
		pbCmts = append(pbCmts, pbCmt)
	}
	return pbCmts
}
