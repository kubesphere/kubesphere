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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/constants"
	kserr "kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"net/http"
	"sort"
	"strings"
)

func ListDevopsProjects(workspace, username string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error) {

	db := mysql.Client()

	var workspaceDOPBindings []models.WorkspaceDPBinding

	if err := db.Where("workspace = ?", workspace).Find(&workspaceDOPBindings).Error; err != nil {
		return nil, err
	}

	devOpsProjects := make([]models.DevopsProject, 0)

	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/api/v1alpha/projects", constants.DevopsAPIServer), nil)
	request.Header.Add(constants.UserNameHeader, username)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 200 {
		return nil, kserr.Parse(data)
	}

	err = json.Unmarshal(data, &devOpsProjects)

	if err != nil {
		return nil, err
	}

	if keyword := conditions.Match["keyword"]; keyword != "" {
		for i := 0; i < len(devOpsProjects); i++ {
			if !strings.Contains(devOpsProjects[i].Name, keyword) {
				devOpsProjects = append(devOpsProjects[:i], devOpsProjects[i+1:]...)
				i--
			}
		}
	}

	sort.Slice(devOpsProjects, func(i, j int) bool {
		switch orderBy {
		case "name":
			if reverse {
				return devOpsProjects[i].Name < devOpsProjects[j].Name
			} else {
				return devOpsProjects[i].Name > devOpsProjects[j].Name
			}
		default:
			if reverse {
				return devOpsProjects[i].CreateTime.After(*devOpsProjects[j].CreateTime)
			} else {
				return devOpsProjects[i].CreateTime.Before(*devOpsProjects[j].CreateTime)
			}
		}
	})

	for i := 0; i < len(devOpsProjects); i++ {
		inWorkspace := false

		for _, binding := range workspaceDOPBindings {
			if binding.DevOpsProject == *devOpsProjects[i].ProjectId {
				inWorkspace = true
			}
		}
		if !inWorkspace {
			devOpsProjects = append(devOpsProjects[:i], devOpsProjects[i+1:]...)
			i--
		}
	}

	// limit offset
	result := make([]interface{}, 0)
	for i, v := range devOpsProjects {
		if len(result) < limit && i >= offset {
			result = append(result, v)
		}
	}

	return &models.PageableResponse{Items: result, TotalCount: len(devOpsProjects)}, nil
}
