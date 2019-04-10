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
package tenant

import (
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/kubesphere"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"sort"
	"strings"
)

func ListDevopsProjects(workspace, username string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error) {

	db := mysql.Client()

	var workspaceDOPBindings []models.WorkspaceDPBinding

	if err := db.Where("workspace = ?", workspace).Find(&workspaceDOPBindings).Error; err != nil {
		return nil, err
	}

	projects, err := kubesphere.Client().ListDevopsProjects(username)
	if err != nil {
		return nil, err
	}

	if keyword := conditions.Match["keyword"]; keyword != "" {
		for i := 0; i < len(projects); i++ {
			if !strings.Contains(projects[i].Name, keyword) {
				projects = append(projects[:i], projects[i+1:]...)
				i--
			}
		}
	}

	sort.Slice(projects, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		switch orderBy {
		case "name":
			return projects[i].Name > projects[j].Name
		default:
			return projects[i].CreateTime.Before(*projects[j].CreateTime)
		}
	})

	for i := 0; i < len(projects); i++ {
		inWorkspace := false

		for _, binding := range workspaceDOPBindings {
			if binding.DevOpsProject == projects[i].ProjectId {
				inWorkspace = true
			}
		}
		if !inWorkspace {
			projects = append(projects[:i], projects[i+1:]...)
			i--
		}
	}

	// limit offset
	result := make([]interface{}, 0)
	for i, v := range projects {
		if len(result) < limit && i >= offset {
			result = append(result, v)
		}
	}

	return &models.PageableResponse{Items: result, TotalCount: len(projects)}, nil
}
