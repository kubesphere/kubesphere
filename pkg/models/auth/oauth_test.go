/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auth

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"

	"kubesphere.io/kubesphere/pkg/constants"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	runtimefakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/scheme"
	"kubesphere.io/kubesphere/pkg/server/options"
)

func Test_oauthAuthenticator_Authenticate(t *testing.T) {
	fakeIDP := &identityprovider.Configuration{
		Name:          "fake",
		MappingMethod: "auto",
		Type:          "FakeOAuthProvider",
		ProviderOptions: options.DynamicOptions{
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
	}

	marshal, err := yaml.Marshal(fakeIDP)
	if err != nil {
		return
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fake-idp",
			Namespace: "kubesphere-system",
			Labels: map[string]string{
				constants.GenericConfigTypeLabel: identityprovider.ConfigTypeIdentityProvider,
			},
		},
		Data: map[string][]byte{
			"configuration.yaml": marshal,
		},
		Type: identityprovider.SecretTypeIdentityProvider,
	}

	fakeCache := informertest.FakeInformers{Scheme: scheme.Scheme}
	err = fakeCache.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fakeSecretInformer, err := fakeCache.FakeInformerFor(context.Background(), &v1.Secret{})
	if err != nil {
		t.Fatal(err)
	}

	identityprovider.RegisterOAuthProviderFactory(&fakeProviderFactory{})
	identityprovider.SharedIdentityProviderController = identityprovider.NewController()
	err = identityprovider.SharedIdentityProviderController.WatchConfigurationChanges(context.Background(), &fakeCache)
	if err != nil {
		t.Fatal(err)
	}

	fakeSecretInformer.Add(secret)

	blockedUser := newUser("user2", "100002", "fake")
	blockedUser.Status = iamv1beta1.UserStatus{State: iamv1beta1.UserDisabled}

	client := runtimefakeclient.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithRuntimeObjects(newUser("user1", "100001", "fake"), secret, blockedUser).
		Build()

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
			name:               "Should successfully",
			oauthAuthenticator: NewOAuthAuthenticator(client),
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
			name:               "Blocked user test",
			oauthAuthenticator: NewOAuthAuthenticator(client),
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
			name:               "Should successfully",
			oauthAuthenticator: NewOAuthAuthenticator(client),
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

			userInfo, err := tt.oauthAuthenticator.Authenticate(tt.args.ctx, tt.args.provider, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(userInfo, tt.userInfo) {
				t.Errorf("Authenticate() got = %v, want %v", userInfo, tt.userInfo)
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

func newUser(username string, uid string, idp string) *iamv1beta1.User {
	return &iamv1beta1.User{
		TypeMeta: metav1.TypeMeta{
			Kind:       iamv1beta1.ResourceKindUser,
			APIVersion: iamv1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: username,
			Labels: map[string]string{
				iamv1beta1.IdentifyProviderLabel: idp,
				iamv1beta1.OriginUIDLabel:        uid,
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
	return "FakeOAuthProvider"
}

func (fakeProviderFactory) Create(dynamicOptions options.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var fakeProvider fakeProvider
	if err := mapstructure.Decode(dynamicOptions, &fakeProvider); err != nil {
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
