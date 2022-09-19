/*
Copyright 2021 KubeSphere Authors

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

package kiali

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"kubesphere.io/kubesphere/pkg/simple/client/cache"
)

func TestClient_Get(t *testing.T) {
	type fields struct {
		Strategy     Strategy
		cache        cache.Interface
		client       HttpClient
		ServiceToken string
		Host         string
	}
	type args struct {
		url string
	}

	inMemoryCache, err := cache.NewInMemoryCache(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	token, _ := json.Marshal(
		&TokenResponse{
			Username: "test",
			Token:    "test",
		},
	)
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp *http.Response
		wantErr  bool
	}{
		{
			name: "Anonymous",
			fields: fields{
				Strategy: AuthStrategyAnonymous,
				cache:    nil,
				client: &MockClient{
					RequestResult: "fake",
				},
				ServiceToken: "token",
				Host:         "http://kiali.istio-system.svc",
			},
			args: args{url: "http://kiali.istio-system.svc"},
			wantResp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("fake"))),
			},
			wantErr: false,
		},
		{
			name: "Token",
			fields: fields{
				Strategy: AuthStrategyToken,
				cache:    nil,
				client: &MockClient{
					TokenResult:   token,
					RequestResult: "fake",
				},
				ServiceToken: "token",
				Host:         "http://kiali.istio-system.svc",
			},
			args: args{url: "http://kiali.istio-system.svc"},
			wantResp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("fake"))),
			},
			wantErr: false,
		},
		{
			name: "Token",
			fields: fields{
				Strategy: AuthStrategyToken,
				cache:    inMemoryCache,
				client: &MockClient{
					TokenResult:   token,
					RequestResult: "fake",
				},
				ServiceToken: "token",
				Host:         "http://kiali.istio-system.svc",
			},
			args: args{url: "http://kiali.istio-system.svc"},
			wantResp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("fake"))),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(
				tt.fields.Strategy,
				tt.fields.cache,
				tt.fields.client,
				tt.fields.ServiceToken,
				tt.fields.Host,
			)
			//nolint:bodyclose
			gotResp, err := c.Get(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("Client.Get() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}
