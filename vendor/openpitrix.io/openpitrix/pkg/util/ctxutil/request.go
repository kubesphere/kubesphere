// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package ctxutil

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func GetRequestId(ctx context.Context) string {
	rid := GetValueFromContext(ctx, requestIdKey)
	if len(rid) == 0 {
		return ""
	}
	return rid[0]
}

func SetRequestId(ctx context.Context, requestId string) context.Context {
	ctx = context.WithValue(ctx, requestIdKey, []string{requestId})
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md[requestIdKey] = []string{requestId}
	return metadata.NewOutgoingContext(ctx, md)
}
