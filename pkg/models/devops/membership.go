/*
Copyright 2019 The KubeSphere Authors.
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

package devops

import "kubesphere.io/kubesphere/pkg/simple/client/devops"

const (
	ProjectMembershipTableName       = "project_membership"
	ProjectMembershipUsernameColumn  = "project_membership.username"
	ProjectMembershipProjectIdColumn = "project_membership.project_id"
	ProjectMembershipRoleColumn      = "project_membership.role"
)

var ProjectMembershipColumns = GetColumnsFromStruct(&devops.ProjectMembership{})

func NewDevOpsProjectMemberShip(username, projectId, role, grantBy string) *devops.ProjectMembership {
	return &devops.ProjectMembership{
		Username:  username,
		ProjectId: projectId,
		Role:      role,
		Status:    StatusActive,
		GrantBy:   grantBy,
	}
}
