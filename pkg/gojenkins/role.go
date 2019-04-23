package gojenkins

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type GlobalRoleResponse struct {
	RoleName      string              `json:"roleName"`
	PermissionIds GlobalPermissionIds `json:"permissionIds"`
}

type GlobalRole struct {
	Jenkins *Jenkins
	Raw     GlobalRoleResponse
}

type GlobalPermissionIds struct {
	Administer              bool `json:"hudson.model.Hudson.Administer"`
	GlobalRead              bool `json:"hudson.model.Hudson.Read"`
	CredentialCreate        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Create"`
	CredentialUpdate        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Update"`
	CredentialView          bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.View"`
	CredentialDelete        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Delete"`
	CredentialManageDomains bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.ManageDomains"`
	SlaveCreate             bool `json:"hudson.model.Computer.Create"`
	SlaveConfigure          bool `json:"hudson.model.Computer.Configure"`
	SlaveDelete             bool `json:"hudson.model.Computer.Delete"`
	SlaveBuild              bool `json:"hudson.model.Computer.Build"`
	SlaveConnect            bool `json:"hudson.model.Computer.Connect"`
	SlaveDisconnect         bool `json:"hudson.model.Computer.Disconnect"`
	ItemBuild               bool `json:"hudson.model.Item.Build"`
	ItemCreate              bool `json:"hudson.model.Item.Create"`
	ItemRead                bool `json:"hudson.model.Item.Read"`
	ItemConfigure           bool `json:"hudson.model.Item.Configure"`
	ItemCancel              bool `json:"hudson.model.Item.Cancel"`
	ItemMove                bool `json:"hudson.model.Item.Move"`
	ItemDiscover            bool `json:"hudson.model.Item.Discover"`
	ItemWorkspace           bool `json:"hudson.model.Item.Workspace"`
	ItemDelete              bool `json:"hudson.model.Item.Delete"`
	RunUpdate               bool `json:"hudson.model.Run.Update"`
	RunDelete               bool `json:"hudson.model.Run.Delete"`
	ViewCreate              bool `json:"hudson.model.View.Create"`
	ViewConfigure           bool `json:"hudson.model.View.Configure"`
	ViewRead                bool `json:"hudson.model.View.Read"`
	ViewDelete              bool `json:"hudson.model.View.Delete"`
	SCMTag                  bool `json:"hudson.scm.SCM.Tag"`
}

type ProjectRole struct {
	Jenkins *Jenkins
	Raw     ProjectRoleResponse
}

type ProjectRoleResponse struct {
	RoleName      string               `json:"roleName"`
	PermissionIds ProjectPermissionIds `json:"permissionIds"`
	Pattern       string               `json:"pattern"`
}

type ProjectPermissionIds struct {
	CredentialCreate        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Create"`
	CredentialUpdate        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Update"`
	CredentialView          bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.View"`
	CredentialDelete        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Delete"`
	CredentialManageDomains bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.ManageDomains"`
	ItemBuild               bool `json:"hudson.model.Item.Build"`
	ItemCreate              bool `json:"hudson.model.Item.Create"`
	ItemRead                bool `json:"hudson.model.Item.Read"`
	ItemConfigure           bool `json:"hudson.model.Item.Configure"`
	ItemCancel              bool `json:"hudson.model.Item.Cancel"`
	ItemMove                bool `json:"hudson.model.Item.Move"`
	ItemDiscover            bool `json:"hudson.model.Item.Discover"`
	ItemWorkspace           bool `json:"hudson.model.Item.Workspace"`
	ItemDelete              bool `json:"hudson.model.Item.Delete"`
	RunUpdate               bool `json:"hudson.model.Run.Update"`
	RunDelete               bool `json:"hudson.model.Run.Delete"`
	RunReplay               bool `json:"hudson.model.Run.Replay"`
	SCMTag                  bool `json:"hudson.scm.SCM.Tag"`
}

func (j *GlobalRole) Update(ids GlobalPermissionIds) error {
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

func (j *ProjectRole) Update(pattern string, ids ProjectPermissionIds) error {
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
