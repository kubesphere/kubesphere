// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package ctxutil

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	messageIdKey = "x-message-id"
	requestIdKey = "x-request-id"
	localeKey    = "locale"
)

type getMetadataFromContext func(ctx context.Context) (md metadata.MD, ok bool)

var getMetadataFromContextFunc = []getMetadataFromContext{
	metadata.FromOutgoingContext,
	metadata.FromIncomingContext,
}

func GetValueFromContext(ctx context.Context, key string) []string {
	if ctx == nil {
		return []string{}
	}
	for _, f := range getMetadataFromContextFunc {
		md, ok := f(ctx)
		if !ok {
			continue
		}
		m, ok := md[key]
		if ok && len(m) > 0 {
			return m
		}
	}
	m, ok := ctx.Value(key).([]string)
	if ok && len(m) > 0 {
		return m
	}
	s, ok := ctx.Value(key).(string)
	if ok && len(s) > 0 {
		return []string{s}
	}
	return []string{}
}

func Copy(src, dst context.Context) context.Context {
	ContextWithSender(dst, GetSender(src))
	SetMessageId(dst, GetMessageId(src))
	SetRequestId(dst, GetRequestId(src))
	return SetLocale(dst, GetLocale(src))
}
