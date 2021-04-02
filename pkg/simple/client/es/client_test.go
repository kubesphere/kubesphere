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

package es

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"

	"kubesphere.io/kubesphere/pkg/simple/client/es/query"
)

func TestNewClient(t *testing.T) {
	var tests = []struct {
		fakeResp string
		expected string
	}{
		{
			fakeResp: "es6_detect_version_major_200.json",
			expected: ElasticV6,
		},
		{
			fakeResp: "es7_detect_version_major_200.json",
			expected: ElasticV7,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			es := mockElasticsearchService("/", test.fakeResp, http.StatusOK)
			defer es.Close()

			client := &Client{host: es.URL}
			err := client.loadClient()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(client.version, test.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}

func TestClient_Search(t *testing.T) {
	var tests = []struct {
		fakeVersion string
		fakeResp    string
		fakeCode    int
		expected    string
		expectedErr string
	}{
		{
			fakeVersion: ElasticV7,
			fakeResp:    "es7_search_200.json",
			fakeCode:    http.StatusOK,
			expected:    "es7_search_200_result.json",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var expected Response
			err := JsonFromFile(test.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}

			srv := mockElasticsearchService("/ks-logstash*/_search", test.fakeResp, test.fakeCode)
			defer srv.Close()

			c, err := NewClient(srv.URL, false, "", "", "ks-logstash", test.fakeVersion)
			if err != nil {
				t.Fatalf("create client error, %s", err)
			}
			result, err := c.Search(query.NewBuilder(), time.Time{}, time.Now(), false)
			if test.expectedErr != "" {
				if diff := cmp.Diff(fmt.Sprint(err), test.expectedErr); diff != "" {
					t.Fatalf("%T differ (-got, +want): %s", test.expectedErr, diff)
				}
			}
			if diff := cmp.Diff(result, &expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func mockElasticsearchService(pattern, fakeResp string, fakeCode int) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(res http.ResponseWriter, req *http.Request) {
		b, _ := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", fakeResp))
		res.WriteHeader(fakeCode)
		res.Write(b)
	})
	return httptest.NewServer(mux)
}

func JsonFromFile(expectedFile string, expectedJsonPtr interface{}) error {
	json, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", expectedFile))
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(json, expectedJsonPtr)
	if err != nil {
		return err
	}

	return nil
}
