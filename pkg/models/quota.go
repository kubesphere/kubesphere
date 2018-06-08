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
	"k8s.io/api/core/v1"

	"encoding/json"
	"errors"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"time"
)

type resourceQuota struct {
	NameSpace       string `json:"namespace"`
	Data            v1.ResourceQuotaStatus `json:"data"`
	UpdateTimeStamp int64 `json:"updateTimeStamp"`
}

func GetNamespaceQuota(namespace string) (*resourceQuota, error) {

	cli, err := client.NewEtcdClient()
	if err != nil {
		glog.Error(err)
	}
	defer cli.Close()

	key := constants.Root + "/" + constants.QuotaKey + "/" + namespace
	value, err := cli.Get(key)
	var data = v1.ResourceQuotaStatus{make(v1.ResourceList), make(v1.ResourceList)}
	var res = resourceQuota{Data: data}

	err = json.Unmarshal(value, &res)
	if time.Now().Unix()-res.UpdateTimeStamp > 5*constants.UpdateCircle {
		err = errors.New("internal server error")
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func GetClusterQuota() (*resourceQuota, error) {

	return GetNamespaceQuota("\"\"")
}
