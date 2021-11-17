// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"encoding/json"
	"fmt"
)

// Resource has the same definition as https://github.com/casbin/casdoor/blob/master/object/resource.go#L24
// used to obtain resource-related information from Casdoor
type Resource struct {
	Owner string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name  string `xorm:"varchar(100) notnull pk" json:"name"`
}

func UploadResource(user string, tag string, parent string, fullFilePath string, fileBytes []byte) (string, string, error) {
	queryMap := map[string]string{
		"owner":        authConfig.OrganizationName,
		"user":         user,
		"application":  authConfig.ApplicationName,
		"tag":          tag,
		"parent":       parent,
		"fullFilePath": fullFilePath,
	}

	resp, err := doPost("upload-resource", queryMap, fileBytes, true)
	if err != nil {
		return "", "", err
	}

	if resp.Status != "ok" {
		return "", "", fmt.Errorf(resp.Msg)
	}

	fileUrl := resp.Data.(string)
	name := resp.Data2.(string)
	return fileUrl, name, nil
}

func DeleteResource(name string) (bool, error) {
	resource := Resource{
		Owner: authConfig.OrganizationName,
		Name:  name,
	}
	postBytes, err := json.Marshal(resource)
	if err != nil {
		return false, err
	}

	resp, err := doPost("delete-resource", nil, postBytes, false)
	if err != nil {
		return false, err
	}

	return resp.Data == "Affected", nil
}
