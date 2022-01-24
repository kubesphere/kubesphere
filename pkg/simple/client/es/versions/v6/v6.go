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

package v6

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"

	"kubesphere.io/kubesphere/pkg/simple/client/es/versions"
)

type Elastic struct {
	Client *elasticsearch.Client
	index  string
}

func New(address string, basicAuth bool, username, password, index string) (*Elastic, error) {
	var client *elasticsearch.Client
	var err error

	if !basicAuth {
		username = ""
		password = ""
	}

	client, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})

	return &Elastic{Client: client, index: index}, err
}

func (e *Elastic) Search(indices string, body []byte, scroll bool) ([]byte, error) {
	opts := []func(*esapi.SearchRequest){
		e.Client.Search.WithContext(context.Background()),
		e.Client.Search.WithIndex(indices),
		e.Client.Search.WithIgnoreUnavailable(true),
		e.Client.Search.WithBody(bytes.NewBuffer(body)),
	}
	if scroll {
		opts = append(opts, e.Client.Search.WithScroll(time.Minute))
	}

	response, err := e.Client.Search(opts...)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.IsError() {
		return nil, parseError(response)
	}

	return ioutil.ReadAll(response.Body)
}

func (e *Elastic) Scroll(id string) ([]byte, error) {
	response, err := e.Client.Scroll(
		e.Client.Scroll.WithContext(context.Background()),
		e.Client.Scroll.WithScrollID(id),
		e.Client.Scroll.WithScroll(time.Minute))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.IsError() {
		return nil, parseError(response)
	}

	return ioutil.ReadAll(response.Body)
}

func (e *Elastic) ClearScroll(scrollId string) {
	response, _ := e.Client.ClearScroll(
		e.Client.ClearScroll.WithContext(context.Background()),
		e.Client.ClearScroll.WithScrollID(scrollId))
	defer response.Body.Close()
}

func (e *Elastic) GetTotalHitCount(v interface{}) int64 {
	f, _ := v.(float64)
	return int64(f)
}

func parseError(response *esapi.Response) error {
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
