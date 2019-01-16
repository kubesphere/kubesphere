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
package registries

import (
	"context"
	"encoding/base64"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/errors"
)

type AuthInfo struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ServerHost string `json:"serverhost"`
}

const loginSuccess = "Login Succeeded"

func RegistryVerify(authInfo AuthInfo) error {
	auth := base64.StdEncoding.EncodeToString([]byte(authInfo.Username + ":" + authInfo.Password))
	ctx := context.Background()
	cli, err := client.NewEnvClient()

	if err != nil {
		glog.Error(err)
	}

	config := types.AuthConfig{
		Username:      authInfo.Username,
		Password:      authInfo.Password,
		Auth:          auth,
		ServerAddress: authInfo.ServerHost,
	}

	resp, err := cli.RegistryLogin(ctx, config)
	cli.Close()

	if err != nil {
		return err
	}

	if resp.Status == loginSuccess {
		return nil
	} else {
		return errors.New(errors.InvalidArgument, resp.Status)
	}
}
