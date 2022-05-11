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
package auth

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"

	"github.com/mitchellh/mapstructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
)

func Test_oauthAuthenticator_Authenticate(t *testing.T) {

	oauthOptions := &authentication.Options{
		OAuthOptions: &oauth.Options{
			IdentityProviders: []oauth.IdentityProviderOptions{
				{
					Name:          "fake",
					MappingMethod: "auto",
					Type:          "FakeIdentityProvider",
					Provider: oauth.DynamicOptions{
						"identities": map[string]interface{}{
							"code1": map[string]string{
								"uid":      "100001",
								"email":    "user1@kubesphere.io",
								"username": "user1",
							},
							"code2": map[string]string{
								"uid":      "100002",
								"email":    "user2@kubesphere.io",
								"username": "user2",
							},
						},
					},
				},
			},
		},
	}

	identityprovider.RegisterOAuthProvider(&fakeProviderFactory{})
	if err := identityprovider.SetupWithOptions(oauthOptions.OAuthOptions.IdentityProviders); err != nil {
		t.Fatal(err)
	}

	ksClient := fakeks.NewSimpleClientset()
	ksInformerFactory := ksinformers.NewSharedInformerFactory(ksClient, 0)

	if err := ksInformerFactory.Iam().V1alpha2().Users().Informer().GetIndexer().Add(newUser("user1", "100001", "fake")); err != nil {
		t.Fatal(err)
	}

	blockedUser := newUser("user2", "100002", "fake")
	blockedUser.Status = iamv1alpha2.UserStatus{State: iamv1alpha2.UserDisabled}
	if err := ksInformerFactory.Iam().V1alpha2().Users().Informer().GetIndexer().Add(blockedUser); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx      context.Context
		provider string
		req      *http.Request
	}
	tests := []struct {
		name               string
		oauthAuthenticator OAuthAuthenticator
		args               args
		userInfo           user.Info
		provider           string
		wantErr            bool
	}{
		{
			name: "Should successfully",
			oauthAuthenticator: NewOAuthAuthenticator(
				nil,
				ksInformerFactory.Iam().V1alpha2().Users().Lister(),
				oauthOptions,
			),
			args: args{
				ctx:      context.Background(),
				provider: "fake",
				req:      must(http.NewRequest(http.MethodGet, "https://ks-console.kubesphere.io/oauth/callback/test?code=code1&state=100001", nil)),
			},
			userInfo: &user.DefaultInfo{
				Name: "user1",
			},
			provider: "fake",
			wantErr:  false,
		},
		{
			name: "Blocked user test",
			oauthAuthenticator: NewOAuthAuthenticator(
				nil,
				ksInformerFactory.Iam().V1alpha2().Users().Lister(),
				oauthOptions,
			),
			args: args{
				ctx:      context.Background(),
				provider: "fake",
				req:      must(http.NewRequest(http.MethodGet, "https://ks-console.kubesphere.io/oauth/callback/test?code=code2&state=100002", nil)),
			},
			userInfo: nil,
			provider: "",
			wantErr:  true,
		},
		{
			name: "Should successfully",
			oauthAuthenticator: NewOAuthAuthenticator(
				nil,
				ksInformerFactory.Iam().V1alpha2().Users().Lister(),
				oauthOptions,
			),
			args: args{
				ctx:      context.Background(),
				provider: "fake1",
				req:      must(http.NewRequest(http.MethodGet, "https://ks-console.kubesphere.io/oauth/callback/test?code=code1&state=100001", nil)),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			userInfo, provider, err := tt.oauthAuthenticator.Authenticate(tt.args.ctx, tt.args.provider, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(userInfo, tt.userInfo) {
				t.Errorf("Authenticate() got = %v, want %v", userInfo, tt.userInfo)
			}
			if provider != tt.provider {
				t.Errorf("Authenticate() got = %v, want %v", provider, tt.provider)
			}
		})
	}
}

func must(r *http.Request, err error) *http.Request {
	if err != nil {
		panic(err)
	}
	return r
}

func newUser(username string, uid string, idp string) *iamv1alpha2.User {
	return &iamv1alpha2.User{
		TypeMeta: metav1.TypeMeta{
			Kind:       iamv1alpha2.ResourceKindUser,
			APIVersion: iamv1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: username,
			Labels: map[string]string{
				iamv1alpha2.IdentifyProviderLabel: idp,
				iamv1alpha2.OriginUIDLabel:        uid,
			},
		},
	}
}

type fakeProviderFactory struct {
}

type fakeProvider struct {
	Identities map[string]fakeIdentity `json:"identities"`
}

type fakeIdentity struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (f fakeIdentity) GetUserID() string {
	return f.UID
}

func (f fakeIdentity) GetUsername() string {
	return f.Username
}

func (f fakeIdentity) GetEmail() string {
	return f.Email
}

func (fakeProviderFactory) Type() string {
	return "FakeIdentityProvider"
}

func (fakeProviderFactory) Create(options oauth.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var fakeProvider fakeProvider
	if err := mapstructure.Decode(options, &fakeProvider); err != nil {
		return nil, err
	}
	return &fakeProvider, nil
}

func (f fakeProvider) IdentityExchangeCallback(req *http.Request) (identityprovider.Identity, error) {
	code := req.URL.Query().Get("code")
	if identity, ok := f.Identities[code]; ok {
		return identity, nil
	}
	return nil, fmt.Errorf("authorization failed")
}
