/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package anonymous

import (
	"net/http"
	"reflect"
	"testing"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
)

func TestNewAuthenticator(t *testing.T) {
	tests := []struct {
		name string
		want authenticator.Request
	}{
		{
			name: "general",
			want: &Authenticator{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAuthenticator(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuthenticator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthenticator_AuthenticateRequest(t *testing.T) {
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		a       *Authenticator
		args    args
		want    *authenticator.Response
		want1   bool
		wantErr error
	}{
		{
			name: "Not auth",
			args: args{req: &http.Request{}},
			want: &authenticator.Response{
				User: &user.DefaultInfo{
					Name:   user.Anonymous,
					UID:    "",
					Groups: []string{user.AllUnauthenticated},
				},
			},
			want1:   true,
			wantErr: nil,
		},
		{
			name:    "Authenticated",
			args:    args{req: &http.Request{Header: http.Header{"Authorization": {"Authenticated"}}}},
			want:    nil,
			want1:   false,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authenticator{}
			got, got1, err := a.AuthenticateRequest(tt.args.req)
			if err != tt.wantErr {
				t.Errorf("Authenticator.AuthenticateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Authenticator.AuthenticateRequest() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Authenticator.AuthenticateRequest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
