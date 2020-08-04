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

package jenkins

import (
	"errors"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type GlobalRoleResponse struct {
	RoleName      string                     `json:"roleName"`
	PermissionIds devops.GlobalPermissionIds `json:"permissionIds"`
}

type GlobalRole struct {
	Jenkins *Jenkins
	Raw     GlobalRoleResponse
}

type ProjectRole struct {
	Jenkins *Jenkins
	Raw     ProjectRoleResponse
}

type ProjectRoleResponse struct {
	RoleName      string                      `json:"roleName"`
	PermissionIds devops.ProjectPermissionIds `json:"permissionIds"`
	Pattern       string                      `json:"pattern"`
}

func (j *GlobalRole) Update(ids devops.GlobalPermissionIds) error {
	var idArray []string
	values := reflect.ValueOf(ids)
	for i := 0; i < values.NumField(); i++ {
		field := values.Field(i)
		if field.Bool() {
			idArray = append(idArray, values.Type().Field(i).Tag.Get("json"))
		}
	}
	param := map[string]string{
		"roleName":      j.Raw.RoleName,
		"type":          GLOBAL_ROLE,
		"permissionIds": strings.Join(idArray, ","),
		"overwrite":     strconv.FormatBool(true),
	}
	responseString := ""
	response, err := j.Jenkins.Requester.Post("/role-strategy/strategy/addRole", nil, &responseString, param)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}

// call jenkins api to update global role
func (j *GlobalRole) AssignRole(sid string) error {
	param := map[string]string{
		"type":     GLOBAL_ROLE,
		"roleName": j.Raw.RoleName,
		"sid":      sid,
	}
	responseString := ""
	response, err := j.Jenkins.Requester.Post("/role-strategy/strategy/assignRole", nil, &responseString, param)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}

func (j *GlobalRole) UnAssignRole(sid string) error {
	param := map[string]string{
		"type":     GLOBAL_ROLE,
		"roleName": j.Raw.RoleName,
		"sid":      sid,
	}
	responseString := ""
	response, err := j.Jenkins.Requester.Post("/role-strategy/strategy/unassignRole", nil, &responseString, param)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}

// update ProjectPermissionIds to Project
// pattern string means some project, like project-name/*
func (j *ProjectRole) Update(pattern string, ids devops.ProjectPermissionIds) error {
	var idArray []string
	values := reflect.ValueOf(ids)
	for i := 0; i < values.NumField(); i++ {
		field := values.Field(i)
		if field.Bool() {
			idArray = append(idArray, values.Type().Field(i).Tag.Get("json"))
		}
	}
	param := map[string]string{
		"roleName":      j.Raw.RoleName,
		"type":          PROJECT_ROLE,
		"permissionIds": strings.Join(idArray, ","),
		"overwrite":     strconv.FormatBool(true),
		"pattern":       pattern,
	}
	responseString := ""
	response, err := j.Jenkins.Requester.Post("/role-strategy/strategy/addRole", nil, &responseString, param)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}

func (j *ProjectRole) AssignRole(sid string) error {
	param := map[string]string{
		"type":     PROJECT_ROLE,
		"roleName": j.Raw.RoleName,
		"sid":      sid,
	}
	responseString := ""
	response, err := j.Jenkins.Requester.Post("/role-strategy/strategy/assignRole", nil, &responseString, param)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}

func (j *ProjectRole) UnAssignRole(sid string) error {
	param := map[string]string{
		"type":     PROJECT_ROLE,
		"roleName": j.Raw.RoleName,
		"sid":      sid,
	}
	responseString := ""
	response, err := j.Jenkins.Requester.Post("/role-strategy/strategy/unassignRole", nil, &responseString, param)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}
