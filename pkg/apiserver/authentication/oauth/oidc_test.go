/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
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
