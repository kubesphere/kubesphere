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

package models

type PageableResponse struct {
	Items      []interface{} `json:"items" description:"paging data"`
	TotalCount int           `json:"total_count" description:"total count"`
}

type Workspace struct {
	Group          `json:",inline"`
	Admin          string   `json:"admin,omitempty"`
	Namespaces     []string `json:"namespaces"`
	DevopsProjects []string `json:"devops_projects"`
}

type Group struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Gid         string   `json:"gid"`
	Members     []string `json:"members"`
	Logo        string   `json:"logo"`
	ChildGroups []string `json:"child_groups"`
	Description string   `json:"description"`
}

type PodInfo struct {
	Namespace string `json:"namespace" description:"namespace"`
	Pod       string `json:"pod" description:"pod name"`
	Container string `json:"container" description:"container name"`
}
