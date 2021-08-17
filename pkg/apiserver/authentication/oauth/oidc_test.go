/*

 Copyright 2021 The KubeSphere Authors.

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
package oauth

import "testing"

func TestIsValidResponseTypes(t *testing.T) {
	type args struct {
		responseTypes []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid response type",
			args: args{responseTypes: []string{"code", "id_token"}},
			want: true,
		},
		{
			name: "invalid response type",
			args: args{responseTypes: []string{"value"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidResponseTypes(tt.args.responseTypes); got != tt.want {
				t.Errorf("IsValidResponseTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidScopes(t *testing.T) {
	type args struct {
		scopes []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid scope",
			args: args{scopes: []string{"openid", "email"}},
			want: true,
		},
		{
			name: "invalid scope",
			args: args{scopes: []string{"user"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidScopes(tt.args.scopes); got != tt.want {
				t.Errorf("IsValidScopes() = %v, want %v", got, tt.want)
			}
		})
	}
}
