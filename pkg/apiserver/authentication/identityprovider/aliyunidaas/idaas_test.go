/*
Copyright 2020 The KubeSphere Authors.

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

package aliyunidaas

import (
	"reflect"
	"testing"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/server/options"
)

func Test_idaasProviderFactory_Create(t *testing.T) {
	type args struct {
		options options.DynamicOptions
	}

	mustUnmarshalYAML := func(data string) options.DynamicOptions {
		var dynamicOptions options.DynamicOptions
		_ = yaml.Unmarshal([]byte(data), &dynamicOptions)
		return dynamicOptions
	}

	tests := []struct {
		name    string
		args    args
		want    identityprovider.OAuthProvider
		wantErr bool
	}{
		{
			name: "should create successfully",
			args: args{options: mustUnmarshalYAML(`
clientID: xxxx
clientSecret: xxxx
endpoint:
  userInfoUrl: "https://xxxxx.login.aliyunidaas.com/api/bff/v1.2/oauth2/userinfo"
  authURL: "https://xxxx.login.aliyunidaas.com/oauth/authorize"
  tokenURL: "https://xxxx.login.aliyunidaas.com/oauth/token"
redirectURL: "https://ks-console.kubesphere-system.svc/oauth/redirect/idaas"
scopes:
- read
`)},
			want: &aliyunIDaaS{
				ClientID:     "xxxx",
				ClientSecret: "xxxx",
				Endpoint: endpoint{
					AuthURL:     "https://xxxx.login.aliyunidaas.com/oauth/authorize",
					TokenURL:    "https://xxxx.login.aliyunidaas.com/oauth/token",
					UserInfoURL: "https://xxxxx.login.aliyunidaas.com/api/bff/v1.2/oauth2/userinfo",
				},
				RedirectURL: "https://ks-console.kubesphere-system.svc/oauth/redirect/idaas",
				Scopes:      []string{"read"},
				Config: &oauth2.Config{
					ClientID:     "xxxx",
					ClientSecret: "xxxx",
					Endpoint: oauth2.Endpoint{
						AuthURL:   "https://xxxx.login.aliyunidaas.com/oauth/authorize",
						TokenURL:  "https://xxxx.login.aliyunidaas.com/oauth/token",
						AuthStyle: oauth2.AuthStyleAutoDetect,
					},
					RedirectURL: "https://ks-console.kubesphere-system.svc/oauth/redirect/idaas",
					Scopes:      []string{"read"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &idaasProviderFactory{}
			got, err := f.Create(tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() got = %v, want %v", got, tt.want)
			}
		})
	}
}
