// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package access

import (
	"context"

	accountclient "openpitrix.io/openpitrix/pkg/client/account"
	"openpitrix.io/openpitrix/pkg/constants"
	"openpitrix.io/openpitrix/pkg/logger"
	"openpitrix.io/openpitrix/pkg/manager"
	"openpitrix.io/openpitrix/pkg/pb"
)

type Client struct {
	pb.AccessManagerClient
}

func NewClient() (*Client, error) {
	conn, err := manager.NewClient(constants.AccountServiceHost, constants.AccountServicePort)
	if err != nil {
		return nil, err
	}
	return &Client{
		AccessManagerClient: pb.NewAccessManagerClient(conn),
	}, nil
}

func (c *Client) CheckActionBundleUser(ctx context.Context, actionBundleIds []string, userId string) bool {
	users, err := c.GetActionBundleUsers(ctx, actionBundleIds)
	if err != nil {
		return false
	}
	for _, user := range users {
		if user.GetUserId().GetValue() == userId {
			return true
		}
	}
	return false
}

func (c *Client) GetActionBundleRoles(ctx context.Context, actionBundleIds []string) ([]*pb.Role, error) {
	response, err := c.DescribeRoles(ctx, &pb.DescribeRolesRequest{
		ActionBundleId: actionBundleIds,
		Status:         []string{constants.StatusActive},
	})
	if err != nil {
		logger.Error(ctx, "Describe roles failed: %+v", err)
		return nil, err
	}

	return response.RoleSet, nil
}

func (c *Client) GetActionBundleUsers(ctx context.Context, actionBundleIds []string) ([]*pb.User, error) {
	roles, err := c.GetActionBundleRoles(ctx, actionBundleIds)
	if err != nil {
		return nil, err
	}
	var roleIds []string
	for _, role := range roles {
		roleIds = append(roleIds, role.RoleId)
	}

	accountClient, err := accountclient.NewClient()
	if err != nil {
		logger.Error(ctx, "Get account manager client failed: %+v", err)
		return nil, err
	}
	return accountClient.GetRoleUsers(ctx, roleIds)
}
