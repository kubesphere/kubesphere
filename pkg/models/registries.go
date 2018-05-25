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
	"github.com/emicklei/go-restful"
	"context"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"

	"kubesphere.io/kubesphere/pkg/constants"
	"github.com/golang/glog"

)


type AuthInfo struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ServerHost string `json:"serverhost"`
}


func RegistryLoginAuth(request *restful.Request, response *restful.Response) {

	var result constants.ResultMessage

	authinfo := AuthInfo{}

	err := request.ReadEntity(&authinfo)

	if err == nil {

		datastr := []byte(authinfo.Username + ":" + authinfo.Password)
		auth := base64.StdEncoding.EncodeToString(datastr)
		ctx := context.Background()
		cli, err := client.NewEnvClient()

		if err != nil {
			panic(err)
		}

		authcfg := types.AuthConfig{

			Username:      authinfo.Username,
			Password:      authinfo.Password,
			Auth:          auth,
			ServerAddress: authinfo.ServerHost,
		}

		auth_msg, err := cli.RegistryLogin(ctx, authcfg)
		data := make(map[string]string)

		if err == nil {


			data["status"] = auth_msg.Status
			result.Data = data
			result.ApiVersion = constants.APIVERSION
			result.Kind = constants.KIND
			glog.Infoln(result)
			response.WriteAsJson(result)
		} else {


			data["status"] = "Login Failed"
			result.Data = data
			result.ApiVersion = constants.APIVERSION
			result.Kind = constants.KIND
			glog.Infoln(result)
			response.WriteAsJson(result)
		}

	} else {

		result.Data = err
		result.ApiVersion = constants.APIVERSION
		result.Kind = constants.KIND
		glog.Infoln(result)
		response.WriteAsJson(result)


	}

}

func RegistryKey(request *restful.Request, response *restful.Response)  {


	var result constants.ResultMessage

	authinfo := AuthInfo{}



	err := request.ReadEntity(&authinfo)

	if err == nil {

		datastr := []byte(authinfo.Username + ":" + authinfo.Password)
		auth := base64.StdEncoding.EncodeToString(datastr)

		dockercfg := "{\"auths\":{\""+authinfo.ServerHost+"\":{\"username\":\""+authinfo.Username+"\",\"password\":\""+authinfo.Password+"\",\"auth\":\""+auth+"\"}}}"

		dockerconfigjson := base64.StdEncoding.EncodeToString([]byte(dockercfg))


		data := make(map[string]string)

		data["dockerconfigjson"] = dockerconfigjson

		result.Data = data

		result.ApiVersion = constants.APIVERSION
		result.Kind = constants.KIND

		glog.Infoln(result)

		response.WriteAsJson(result)


	} else {


		result.Data = err
		result.ApiVersion = constants.APIVERSION
		result.Kind = constants.KIND
		glog.Infoln(result)
		response.WriteAsJson(result)
	}

}


