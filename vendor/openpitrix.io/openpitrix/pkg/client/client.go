// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package client

import (
	"context"

	accessclient "openpitrix.io/openpitrix/pkg/client/access"
	"openpitrix.io/openpitrix/pkg/pb"
	"openpitrix.io/openpitrix/pkg/sender"
	"openpitrix.io/openpitrix/pkg/util/ctxutil"
)

func SetSystemUserToContext(ctx context.Context) context.Context {
	return ctxutil.ContextWithSender(ctx, sender.GetSystemSender())
}

func SetUserToContext(ctx context.Context, userId, apiMethod string) (context.Context, error) {
	accessClient, err := accessclient.NewClient()
	if err != nil {
		return nil, err
	}
	response, err := accessClient.CanDo(ctx, &pb.CanDoRequest{
		UserId:    userId,
		ApiMethod: apiMethod,
	})
	if err != nil {
		return nil, err
	}

	userSender := sender.New(response.UserId, sender.OwnerPath(response.OwnerPath), sender.OwnerPath(response.AccessPath))
	return ctxutil.ContextWithSender(ctx, userSender), nil
}
