// Copyright 2017 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package pbutil

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"

	"openpitrix.io/openpitrix/pkg/db"
	"openpitrix.io/openpitrix/pkg/logger"
)

type RequestHadOffset interface {
	GetOffset() uint32
}

type RequestHadLimit interface {
	GetLimit() uint32
}

const (
	DefaultOffset = uint64(0)
	DefaultLimit  = uint64(20)
)

func GetOffsetFromRequest(req RequestHadOffset) uint64 {
	n := req.GetOffset()
	if n == 0 {
		return DefaultOffset
	}
	return db.GetOffset(uint64(n))
}

func GetLimitFromRequest(req RequestHadLimit) uint64 {
	n := req.GetLimit()
	if n == 0 {
		return DefaultLimit
	}
	return db.GetLimit(uint64(n))
}

func GetTime(t *timestamp.Timestamp) (tt time.Time) {
	if t == nil {
		return time.Now()
	} else {
		return FromProtoTimestamp(t)
	}
}

func FromProtoTimestamp(t *timestamp.Timestamp) (tt time.Time) {
	tt, err := ptypes.Timestamp(t)
	if err != nil {
		logger.Critical(nil, "Cannot convert timestamp [T] to time.Time [%+v]: %+v", t, err)
		panic(err)
	}
	return
}

func ToProtoTimestamp(t time.Time) (tt *timestamp.Timestamp) {
	if t.IsZero() {
		return nil
	}
	tt, err := ptypes.TimestampProto(t)
	if err != nil {
		logger.Critical(nil, "Cannot convert time.Time [%+v] to ToProtoTimestamp[T]: %+v", t, err)
		panic(err)
	}
	return
}

func ToProtoString(str string) *wrappers.StringValue {
	return &wrappers.StringValue{Value: str}
}

func ToProtoUInt32(uint32 uint32) *wrappers.UInt32Value {
	return &wrappers.UInt32Value{Value: uint32}
}

func ToProtoInt32(i int32) *wrappers.Int32Value {
	return &wrappers.Int32Value{Value: i}
}

func ToProtoBool(bool bool) *wrappers.BoolValue {
	return &wrappers.BoolValue{Value: bool}
}

func ToProtoBytes(bytes []byte) *wrappers.BytesValue {
	return &wrappers.BytesValue{Value: bytes}
}

type DescribeResponse interface {
	GetTotalCount() uint32
}

type DescribeApi interface {
	SetRequest(ctx context.Context, req interface{}, limit, offset uint32) error
	Describe(ctx context.Context, req interface{}, advancedParams ...string) (DescribeResponse, error)
}

func DescribeAllResponses(ctx context.Context, describeApi DescribeApi, req interface{}, advancedParams ...string) ([]DescribeResponse, error) {
	limit := uint32(db.DefaultSelectLimit)
	offset := uint32(0)
	var responses []DescribeResponse

	if err := describeApi.SetRequest(ctx, req, limit, offset); err != nil {
		return nil, err
	}
	response, err := describeApi.Describe(ctx, req, advancedParams...)
	if err != nil {
		return nil, err
	}

	totalCount := response.GetTotalCount()

	responses = append(responses, response)
	offset = offset + db.DefaultSelectLimit
	for {
		if totalCount > uint32(offset) {
			if err := describeApi.SetRequest(ctx, req, limit, offset); err != nil {
				return nil, err
			}
			response, err = describeApi.Describe(ctx, req)
			if err != nil {
				return nil, err
			}
			responses = append(responses, response)
			offset = offset + db.DefaultSelectLimit
		} else {
			break
		}
	}
	return responses, nil
}
