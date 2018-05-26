/*
Copyright 2018 The KubeSphere Authors.

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

import (
	"encoding/base64"
	"context"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"

	"kubesphere.io/kubesphere/pkg/constants"

)


type AuthInfo struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ServerHost string `json:"serverhost"`
}


const DOCKERCLIENTERROR = "Docker client error"

func RegistryLoginAuth(authinfo AuthInfo) constants.ResultMessage {

	var result constants.ResultMessage

	data := make(map[string]interface{})

	datastr := []byte(authinfo.Username + ":" + authinfo.Password)
	auth := base64.StdEncoding.EncodeToString(datastr)
	ctx := context.Background()
	cli, err := client.NewClientWithOpts()

	if err != nil {

		data["message"] = DOCKERCLIENTERROR
		data["reason"] = err.Error()
	}

	authcfg := types.AuthConfig{

		Username:      authinfo.Username,
		Password:      authinfo.Password,
		Auth:          auth,
		ServerAddress: authinfo.ServerHost,
	}

	authmsg, err := cli.RegistryLogin(ctx, authcfg)

	cli.Close()

	if err != nil {

		data["message"] = DOCKERCLIENTERROR
		data["reason"] = err.Error()

	}

	if authmsg.Status == "Login Succeeded" {

		data["message"] = "Verified"

	} else {

		data["message"] = "Unverified"
		data["reason"] = "Username or password is incorrect "

	}

	result.Data = data


	return result

}


