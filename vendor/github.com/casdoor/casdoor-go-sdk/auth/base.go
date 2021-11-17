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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

// doGetBytes is a general function to get response from param url through HTTP Get method.
func doGetBytes(url string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(authConfig.ClientId, authConfig.ClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func doPost(action string, queryMap map[string]string, postBytes []byte, isFile bool) (*Response, error) {
	client := &http.Client{}
	url := getUrl(action, queryMap)

	var resp *http.Response
	var err error
	var contentType string
	var body io.Reader
	if isFile {
		contentType, body, err = createForm(map[string][]byte{"file": postBytes})
		if err != nil {
			return nil, err
		}
	} else {
		contentType = "text/plain;charset=UTF-8"
		body = bytes.NewReader(postBytes)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(authConfig.ClientId, authConfig.ClientSecret)
	req.Header.Set("Content-Type", contentType)

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(respByte, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// modifyUser is an encapsulation of user CUD(Create, Update, Delete) operations.
// possible actions are `add-user`, `update-user`, `delete-user`,
func modifyUser(action string, user *User) (*Response, bool, error) {
	queryMap := map[string]string{
		"id": fmt.Sprintf("%s/%s", user.Owner, user.Name),
	}

	user.Owner = authConfig.OrganizationName
	postBytes, err := json.Marshal(user)
	if err != nil {
		return nil, false, err
	}

	resp, err := doPost(action, queryMap, postBytes, false)
	if err != nil {
		return nil, false, err
	}

	return resp, resp.Data == "Affected", nil
}
