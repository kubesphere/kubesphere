/*
Copyright 2020 KubeSphere Authors

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

/**
project operator, providing API for creating/getting/deleting projects
The actual data of the project is stored in the CRD,
so we only need to create the project with the corresponding ID in the CI/CD system.
*/
type ProjectOperator interface {
	CreateDevOpsProject(projectId string) (string, error)
	DeleteDevOpsProject(projectId string) error
	GetDevOpsProject(projectId string) (string, error)
}
