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
	"kubesphere.io/kubesphere/pkg/server/params"
	dsClient "kubesphere.io/kubesphere/pkg/simple/client/devops"
)

type DevOpsProjectLister interface {
	ListDevOpsProjects(workspace string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error)
}

type devopsProjectLister struct {
	dsProject dsClient.ProjectOperator
}

func newProjectLister(client dsClient.ProjectOperator) DevOpsProjectLister {
	return &devopsProjectLister{
		dsProject: client,
	}
}

func (o *devopsProjectLister) ListDevOpsProjects(workspace string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error) {
	//TODO: @runzexia use informer to impl it
	return nil, nil
}
