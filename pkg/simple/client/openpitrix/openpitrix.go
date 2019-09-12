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
package openpitrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"net/http"
	"strings"
	"time"
)

func NewOpenPitrixClient(options *OpenPitrixOptions) (*OpenPitrixClient, error) {
	return &OpenPitrixClient{
		client: &http.Client{
			Timeout: time.Duration(3) * time.Second,
		},
		apiServer: options.APIServer,
		token:     options.Token,
	}, nil
}

func (c *OpenPitrixClient) CreateRuntime(runtime *RunTime) error {

	data, err := json.Marshal(runtime)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/runtimes", c.apiServer), bytes.NewReader(data))

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", c.token)

	resp, err := c.client.Do(req)

	if err != nil {
		klog.Error(err)
		return err
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error(err)
		return err
	}

	if resp.StatusCode > http.StatusOK {
		err = Error{resp.StatusCode, string(data)}
		klog.Error(err)
		return err
	}

	return nil
}

func (c *OpenPitrixClient) deleteClusters(clusters []cluster) error {
	clusterId := make([]string, 0)

	for _, cluster := range clusters {
		if cluster.Status != "deleted" && cluster.Status != "deleting" && !sliceutil.HasString(clusterId, cluster.ClusterId) {
			clusterId = append(clusterId, cluster.ClusterId)
		}
	}

	if len(clusterId) == 0 {
		return nil
	}

	deleteRequest := struct {
		ClusterId []string `json:"cluster_id"`
	}{
		ClusterId: clusterId,
	}
	data, _ := json.Marshal(deleteRequest)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/clusters/delete", c.apiServer), bytes.NewReader(data))

	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		klog.Error(err)
		return err
	}

	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error(err)
		return err
	}

	if resp.StatusCode > http.StatusOK {
		err = Error{resp.StatusCode, string(data)}
		klog.Error(err)
		return err
	}

	return nil
}

func (c *OpenPitrixClient) listClusters(runtimeId string) ([]cluster, error) {
	limit := 200
	offset := 0
	clusters := make([]cluster, 0)
	for {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v1/clusters?runtime_id=%s&limit=%d&offset=%d", c.apiServer, runtimeId, limit, offset), nil)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		req.Header.Add("Authorization", c.token)

		resp, err := c.client.Do(req)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		data, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		resp.Body.Close()

		if resp.StatusCode > http.StatusOK {
			err = Error{resp.StatusCode, string(data)}
			klog.Error(err)
			return nil, err
		}
		listClusterResponse := struct {
			TotalCount int       `json:"total_count"`
			ClusterSet []cluster `json:"cluster_set"`
		}{}
		err = json.Unmarshal(data, &listClusterResponse)

		if err != nil {
			klog.Error(err)
			return nil, err
		}

		clusters = append(clusters, listClusterResponse.ClusterSet...)

		if listClusterResponse.TotalCount <= limit+offset {
			break
		}

		offset += limit
	}

	return clusters, nil
}

func (c *OpenPitrixClient) DeleteRuntime(runtimeId string) error {
	clusters, err := c.listClusters(runtimeId)

	if err != nil {
		klog.Error(err)
		return err
	}

	err = c.deleteClusters(clusters)

	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func IsNotFound(err error) bool {
	if e, ok := err.(Error); ok {
		if e.status == http.StatusNotFound {
			return true
		}
		if strings.Contains(e.message, "not exist") {
			return true
		}
		if strings.Contains(e.message, "not found") {
			return true
		}
	}
	return false
}

func IsDeleted(err error) bool {
	if e, ok := err.(Error); ok {
		if strings.Contains(e.message, "is [deleted]") {
			return true
		}
	}
	return false
}
