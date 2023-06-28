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

package v1

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"

	"kubesphere.io/kubesphere/pkg/simple/client/es/versions"
)

type OpenSearch struct {
	client *opensearch.Client
	index  string
}

func New(address string, basicAuth bool, username, password, index string) (*OpenSearch, error) {
	var client *opensearch.Client
	var err error

	if !basicAuth {
		username = ""
		password = ""
	}

	client, err = opensearch.NewClient(opensearch.Config{
		Addresses: []string{address},
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})

	return &OpenSearch{client: client, index: index}, err
}

func (o *OpenSearch) Search(indices string, body []byte, scroll bool) ([]byte, error) {
	opts := []func(*opensearchapi.SearchRequest){
		o.client.Search.WithContext(context.Background()),
		o.client.Search.WithIndex(indices),
		o.client.Search.WithRestTotalHitsAsInt(true),
		o.client.Search.WithIgnoreUnavailable(true),
		o.client.Search.WithBody(bytes.NewBuffer(body)),
	}
	if scroll {
		opts = append(opts, o.client.Search.WithScroll(time.Minute))
	}

	response, err := o.client.Search(opts...)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.IsError() {
		return nil, parseError(response)
	}

	return io.ReadAll(response.Body)
}

func (o *OpenSearch) Scroll(id string) ([]byte, error) {
	body, err := jsoniter.Marshal(map[string]string{
		"scroll_id": id,
	})
	if err != nil {
		return nil, err
	}

	response, err := o.client.Scroll(
		o.client.Scroll.WithContext(context.Background()),
		o.client.Scroll.WithBody(bytes.NewBuffer(body)),
		o.client.Scroll.WithScroll(time.Minute))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.IsError() {
		return nil, parseError(response)
	}

	return io.ReadAll(response.Body)
}

func (o *OpenSearch) ClearScroll(scrollId string) {
	response, _ := o.client.ClearScroll(
		o.client.ClearScroll.WithContext(context.Background()),
		o.client.ClearScroll.WithScrollID(scrollId))
	defer response.Body.Close()
}

func (o *OpenSearch) GetTotalHitCount(v interface{}) int64 {
	f, _ := v.(float64)
	return int64(f)
}

func parseError(response *opensearchapi.Response) error {
	var e versions.Error
	if err := json.NewDecoder(response.Body).Decode(&e); err != nil {
		return err
	} else {
		// Print the response status and error information.
		if len(e.Details.RootCause) != 0 {
			return fmt.Errorf("type: %v, reason: %v", e.Details.Type, e.Details.RootCause[0].Reason)
		} else {
			return fmt.Errorf("type: %v, reason: %v", e.Details.Type, e.Details.Reason)
		}
	}
}
