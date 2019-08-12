// Copyright 2019 The KubeSphere Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package pbutil

import (
	"time"

	"kubesphere.io/kubesphere/pkg/constants"

	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
)

type RequestHadOffset interface {
	GetOffset() uint32
}

type RequestHadLimit interface {
	GetLimit() uint32
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
		glog.Fatalf("Cannot convert timestamp [T] to time.Time [%+v]: %+v", t, err)
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
		glog.Fatalf("Cannot convert time.Time [%+v] to ToProtoTimestamp[T]: %+v", t, err)
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

func GetOffsetFromRequest(req RequestHadOffset) uint32 {
	n := req.GetOffset()
	if n == 0 {
		return constants.DefaultOffset
	}

	return GetOffset(uint32(n))
}

func GetLimitFromRequest(req RequestHadLimit) uint32 {
	n := req.GetLimit()
	if n == 0 {
		return constants.DefaultLimit
	}
	return GetLimit(uint32(n))
}

func GetLimit(n uint32) uint32 {
	if n < 0 {
		n = 0
	}
	if n > constants.DefaultSelectLimit {
		n = constants.DefaultSelectLimit
	}
	return n
}

func GetOffset(n uint32) uint32 {
	if n < 0 {
		n = 0
	}
	return n
}
