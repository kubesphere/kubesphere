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
	"encoding/json"
	"errors"
	"time"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
)

type workLoadStatus struct {
	NameSpace       string         `json:"namespace"`
	Data            map[string]int `json:"data"`
	UpdateTimeStamp int64          `json:"updateTimeStamp"`
}

var resourceList = []string{"deployments", "daemonsets", "statefulsets"}

func GetNamespacesResourceStatus(namespace string) (*workLoadStatus, error) {

	cli, err := client.NewEtcdClient()
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	defer cli.Close()

	res := workLoadStatus{Data: make(map[string]int)}

	key := constants.Root + "/" + constants.WorkloadStatusKey + "/" + namespace
	value, err := cli.Get(key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(value, &res)
	if err != nil {
		return nil, err
	}

	if time.Now().Unix()-res.UpdateTimeStamp > 5*constants.UpdateCircle {
		err = errors.New("data in etcd is too old")
		return nil, err
	}
	return &res, nil
}

func GetClusterResourceStatus() (*workLoadStatus, error) {

	return GetNamespacesResourceStatus("\"\"")
}
