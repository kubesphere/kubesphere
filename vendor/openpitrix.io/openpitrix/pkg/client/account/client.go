// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package account

import (
	"context"
	"fmt"
	"math"
	"strings"

	"openpitrix.io/openpitrix/pkg/constants"
	"openpitrix.io/openpitrix/pkg/logger"
	"openpitrix.io/openpitrix/pkg/manager"
	"openpitrix.io/openpitrix/pkg/pb"
	"openpitrix.io/openpitrix/pkg/util/pbutil"
	"openpitrix.io/openpitrix/pkg/util/stringutil"
)

type Client struct {
	pb.AccountManagerClient
}

func NewClient() (*Client, error) {
	conn, err := manager.NewClient(constants.AccountServiceHost, constants.AccountServicePort)
	if err != nil {
		return nil, err
	}
	return &Client{
		AccountManagerClient: pb.NewAccountManagerClient(conn),
	}, nil
}

func (c *Client) GetUsers(ctx context.Context, userIds []string) ([]*pb.User, error) {
	var internalUsers []*pb.User
	var noInternalUserIds []string
	for _, userId := range userIds {
		if stringutil.StringIn(userId, constants.InternalUsers) {
			internalUsers = append(internalUsers, &pb.User{
				UserId: pbutil.ToProtoString(userId),
			})
		} else {
			noInternalUserIds = append(noInternalUserIds, userId)
		}
	}

	if len(noInternalUserIds) == 0 {
		return internalUsers, nil
	}

	response, err := c.DescribeUsers(ctx, &pb.DescribeUsersRequest{
		UserId: noInternalUserIds,
	})
	if err != nil {
		logger.Error(ctx, "Describe users %s failed: %+v", noInternalUserIds, err)
		return nil, err
	}
	if len(response.UserSet) != len(noInternalUserIds) {
		logger.Error(ctx, "Describe users %s with return count [%d]", userIds, len(response.UserSet)+len(internalUsers))
		return nil, fmt.Errorf("describe users %s with return count [%d]", userIds, len(response.UserSet)+len(internalUsers))
	}
	response.UserSet = append(response.UserSet, internalUsers...)
	return response.UserSet, nil
}

func (c *Client) GetUser(ctx context.Context, userId string) (*pb.User, error) {
	users, err := c.GetUsers(ctx, []string{userId})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("not found user [%s]", userId)
	}
	return users[0], nil
}

func (c *Client) GetUserGroupPath(ctx context.Context, userId string) (string, error) {
	var userGroupPath string

	response, err := c.DescribeUsersDetail(ctx, &pb.DescribeUsersRequest{
		UserId: []string{userId},
	})
	if err != nil || len(response.UserDetailSet) == 0 {
		logger.Error(ctx, "Describe user [%s] failed: %+v", userId, err)
		return "", err
	}

	groups := response.UserDetailSet[0].GroupSet

	//If one user under different groups, get the highest group path.
	minLevel := math.MaxInt32
	for _, group := range groups {
		level := len(strings.Split(group.GroupPath.GetValue(), "."))
		if level < minLevel {
			minLevel = level
			userGroupPath = group.GetGroupPath().GetValue()
		}
	}

	return userGroupPath, nil

}

func (c *Client) GetRoleUsers(ctx context.Context, roleIds []string) ([]*pb.User, error) {
	response, err := c.DescribeUsers(ctx, &pb.DescribeUsersRequest{
		RoleId: roleIds,
		Status: []string{constants.StatusActive},
	})
	if err != nil {
		logger.Error(ctx, "Describe users failed: %+v", err)
		return nil, err
	}

	return response.UserSet, nil
}

func (c *Client) GetIsvFromUser(ctx context.Context, userId string) (*pb.User, error) {
	groupPath, err := c.GetUserGroupPath(ctx, userId)
	if err != nil {
		return nil, err
	}

	rootGroupId := strings.Split(groupPath, ".")[0]

	describeUsersResponse, err := c.DescribeUsers(ctx, &pb.DescribeUsersRequest{
		RootGroupId: []string{rootGroupId},
		Status:      []string{constants.StatusActive},
		RoleId:      []string{constants.RoleIsv},
	})
	if err != nil {
		logger.Error(ctx, "Failed to describe users: %+v", err)
		return nil, err
	}

	if len(describeUsersResponse.UserSet) == 0 {
		logger.Error(ctx, "Isv not exist with root group id [%s]", rootGroupId)
		return nil, fmt.Errorf("isv not exist")
	}

	return describeUsersResponse.UserSet[0], nil
}
