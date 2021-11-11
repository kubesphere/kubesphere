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

package auth

import (
	"context"
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
)

func TestEncryptPassword(t *testing.T) {
	password := "P@88w0rd"
	encryptedPassword, err := hashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if err = PasswordVerify(encryptedPassword, password); err != nil {
		t.Fatal(err)
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(bytes), err
}

func Test_passwordAuthenticator_Authenticate(t *testing.T) {

	oauthOptions := &authentication.Options{
		OAuthOptions: &oauth.Options{
			IdentityProviders: []oauth.IdentityProviderOptions{
				{
					Name:          "fakepwd",
					MappingMethod: "auto",
					Type:          "fakePasswordProvider",
					Provider: oauth.DynamicOptions{
						"identities": map[string]interface{}{
							"user1": map[string]string{
								"uid":      "100001",
								"email":    "user1@kubesphere.io",
								"username": "user1",
								"password": "password",
							},
							"user2": map[string]string{
								"uid":      "100002",
								"email":    "user2@kubesphere.io",
								"username": "user2",
								"password": "password",
							},
						},
					},
				},
			},
		},
	}

	identityprovider.RegisterGenericProvider(&fakePasswordProviderFactory{})
	if err := identityprovider.SetupWithOptions(oauthOptions.OAuthOptions.IdentityProviders); err != nil {
		t.Fatal(err)
	}

	ksClient := fakeks.NewSimpleClientset()
	ksInformerFactory := ksinformers.NewSharedInformerFactory(ksClient, 0)
	err := ksInformerFactory.Iam().V1alpha2().Users().Informer().GetIndexer().Add(newUser("user1", "100001", "fakepwd"))
	err = ksInformerFactory.Iam().V1alpha2().Users().Informer().GetIndexer().Add(newUser("user3", "100003", ""))
	err = ksInformerFactory.Iam().V1alpha2().Users().Informer().GetIndexer().Add(newActiveUser("user4", "password"))

	if err != nil {
		t.Fatal(err)
	}

	authenticator := NewPasswordAuthenticator(
		ksClient,
		ksInformerFactory.Iam().V1alpha2().Users().Lister(),
		oauthOptions,
	)

	type args struct {
		ctx      context.Context
		username string
		password string
	}
	tests := []struct {
		name                  string
		passwordAuthenticator PasswordAuthenticator
		args                  args
		want                  authuser.Info
		want1                 string
		wantErr               bool
	}{
		{
			name:                  "Should successfully with existing provider user",
			passwordAuthenticator: authenticator,
			args: args{
				ctx:      context.Background(),
				username: "user1",
				password: "password",
			},
			want: &user.DefaultInfo{
				Name: "user1",
			},
			wantErr: false,
		},
		{
			name:                  "Should return register user",
			passwordAuthenticator: authenticator,
			args: args{
				ctx:      context.Background(),
				username: "user2",
				password: "password",
			},
			want: &user.DefaultInfo{
				Name: "system:pre-registration",
				Extra: map[string][]string{
					"email":    {"user2@kubesphere.io"},
					"idp":      {"fakepwd"},
					"uid":      {"100002"},
					"username": {"user2"},
				},
			},
			wantErr: false,
		},
		{
			name:                  "Should failed login",
			passwordAuthenticator: authenticator,
			args: args{
				ctx:      context.Background(),
				username: "user3",
				password: "password",
			},
			wantErr: true,
		},
		{
			name:                  "Should successfully with internal user",
			passwordAuthenticator: authenticator,
			args: args{
				ctx:      context.Background(),
				username: "user4",
				password: "password",
			},
			want: &user.DefaultInfo{
				Name: "user4",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.passwordAuthenticator
			got, _, err := p.Authenticate(tt.args.ctx, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("passwordAuthenticator.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("passwordAuthenticator.Authenticate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakePasswordProviderFactory struct {
}

type fakePasswordProvider struct {
	Identities map[string]fakePasswordIdentity `json:"identities"`
}

type fakePasswordIdentity struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (f fakePasswordIdentity) GetUserID() string {
	return f.UID
}

func (f fakePasswordIdentity) GetUsername() string {
	return f.Username
}

func (f fakePasswordIdentity) GetEmail() string {
	return f.Email
}

func (fakePasswordProviderFactory) Type() string {
	return "fakePasswordProvider"
}

func (fakePasswordProviderFactory) Create(options oauth.DynamicOptions) (identityprovider.GenericProvider, error) {
	var fakeProvider fakePasswordProvider
	if err := mapstructure.Decode(options, &fakeProvider); err != nil {
		return nil, err
	}
	return &fakeProvider, nil
}

func (l fakePasswordProvider) Authenticate(username string, password string) (identityprovider.Identity, error) {
	if i, ok := l.Identities[username]; ok && i.Password == password {
		return i, nil
	}
	return nil, errors.NewUnauthorized("authorization failed")
}

func encrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func newActiveUser(username string, password string) *iamv1alpha2.User {
	u := newUser(username, "", "")
	password, _ = encrypt(password)
	u.Spec.EncryptedPassword = password
	s := iamv1alpha2.UserActive
	u.Status.State = &s
	return u
}
