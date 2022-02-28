/*
Copyright 2022 The KubeSphere Authors.

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

package devops

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
)

func TestAddToContainer(t *testing.T) {
	fakeResponse := "fake DevOps APIServer response"
	notFoundResponse := "404 page not found\n"
	type args struct {
		target         string
		mockAPIPattern string
		mockResponse   string
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantResponse string
	}{{
		name: "Should proxy devops.kubesphere.io/v1alpha1 API properly",
		args: args{
			target:         "/kapis/devops.kubesphere.io/v1alpha1/resources",
			mockAPIPattern: "/kapis/devops.kubesphere.io/v1alpha1/resources",
			mockResponse:   fakeResponse,
		},
		wantResponse: notFoundResponse,
	}, {
		name: "Should proxy devops.kubesphere.io/v1alpha2 API properly",
		args: args{
			target:         "/kapis/devops.kubesphere.io/v1alpha2/resources",
			mockAPIPattern: "/kapis/devops.kubesphere.io/v1alpha2/resources",
			mockResponse:   fakeResponse,
		},
		wantResponse: fakeResponse,
	}, {
		name: "Should proxy devops.kubesphere.io/v1alpha3 API properly",
		args: args{
			target:         "/kapis/devops.kubesphere.io/v1alpha3/resources",
			mockAPIPattern: "/kapis/devops.kubesphere.io/v1alpha3/resources",
			mockResponse:   fakeResponse,
		},
		wantResponse: fakeResponse,
	}, {
		name: "Should proxy gitops.kubesphere.io/v1alpha1 API properly",
		args: args{
			target:         "/kapis/gitops.kubesphere.io/v1alpha1/resources",
			mockAPIPattern: "/kapis/gitops.kubesphere.io/v1alpha1/resources",
			mockResponse:   fakeResponse,
		},
		wantResponse: fakeResponse,
	}, {
		name: "Should return 404 if versions miss match",
		args: args{
			target:         "/kapis/devops.kubesphere.io/v1alpha3/resources",
			mockAPIPattern: "/kapis/devops.kubesphere.io/v1alpha1/resources",
		},
		wantResponse: notFoundResponse,
	}, {
		name: "Should return 404 if groups miss match",
		args: args{
			target:         "/kapis/devops.kubesphere.io/v1alpha3/resources",
			mockAPIPattern: "/kapis/gitops.kubesphere.io/v1alpha3/resources",
		},
		wantResponse: notFoundResponse,
	},
		{
			name: "Should not proxy v1alpha123 API properly event if pattern matched",
			args: args{
				target:         "/kapis/devops.kubesphere.io/v1alpha123/resources",
				mockAPIPattern: "/kapis/devops.kubesphere.io/v1alpha123/resources",
			},
			wantResponse: notFoundResponse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a fresh mock DevOps APIServer.
			server := mockDevOpsAPIServer(tt.args.mockAPIPattern, 200, tt.args.mockResponse)
			defer server.Close()

			// mock to request DevOps API from KubeSphere APIServer
			container := restful.NewContainer()
			if err := AddToContainer(container, server.URL); (err != nil) != tt.wantErr {
				t.Errorf("AddToContainer() error = %v, wantErr %v", err, tt.wantErr)
			}
			request := httptest.NewRequest(http.MethodGet, tt.args.target, nil)
			recorder := &responseRecorder{*httptest.NewRecorder()}
			container.ServeHTTP(restful.NewResponse(recorder), request)

			assert.Equal(t, tt.wantResponse, recorder.Body.String())
		})
	}
}

func mockDevOpsAPIServer(pattern string, fakeCode int, fakeResp string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(fakeCode)
		_, _ = writer.Write([]byte(fakeResp))
	})
	return httptest.NewServer(mux)
}

// responseRecorder extends httptest.ResponseRecorder and implements CloseNotifier interface for generic proxy.
type responseRecorder struct {
	httptest.ResponseRecorder
}

func (*responseRecorder) CloseNotify() <-chan bool {
	return nil
}
