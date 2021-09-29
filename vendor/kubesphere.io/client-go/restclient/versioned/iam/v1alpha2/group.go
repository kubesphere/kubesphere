/*
Copyright 2021 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"context"

	resty "github.com/go-resty/resty/v2"
)

type GroupsGetter interface {
	Groups() GroupInterface
}

type GroupInterface interface {
	CreateBinding(ctx context.Context, workspace, group, user string) (string, error)
}

type groups struct {
	client *resty.Client
}

func newGroups(c *IamV1alpha2Client) *groups {
	return &groups{
		client: c.client,
	}
}

//TODO: to be remoted once we move kubesphere.io/apis out of kubesphere package
type groupMember struct {
	UserName  string `json:"userName"`
	GroupName string `json:"groupName"`
}

// Create takes the representation of a group and creates it.  Returns the server's representation of the group, and an error, if there is any.
func (c *groups) CreateBinding(ctx context.Context, workspace, group, user string) (result string, err error) {

	members := []groupMember{{
		UserName:  user,
		GroupName: group,
	}}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(members).
		SetPathParams(map[string]string{
			"workspace": workspace,
		}).
		Post("/kapis/iam.kubesphere.io/v1alpha2/workspaces/{workspace}/groupbindings")
	return resp.String(), err
}
