// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package ctxutil

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/metadata"

	"openpitrix.io/openpitrix/pkg/sender"
)

const (
	SenderKey = "sender"
	TokenType = "Bearer"
)

func GetSender(ctx context.Context) *sender.Sender {
	values := GetValueFromContext(ctx, SenderKey)
	if len(values) == 0 || len(values[0]) == 0 {
		return nil
	}
	s := sender.Sender{}
	err := json.Unmarshal([]byte(values[0]), &s)
	if err != nil {
		panic(err)
	}
	return &s
}

func ContextWithSender(ctx context.Context, user *sender.Sender) context.Context {
	if user == nil {
		return ctx
	}
	ctx = context.WithValue(ctx, SenderKey, []string{user.ToJson()})
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md[SenderKey] = []string{user.ToJson()}
	return metadata.NewOutgoingContext(ctx, md)
}
