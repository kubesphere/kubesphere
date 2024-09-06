/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auth

import (
	"context"
	"reflect"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"

	"kubesphere.io/kubesphere/pkg/constants"

	"gopkg.in/yaml.v3"

	runtimefakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/scheme"

	"kubesphere.io/kubesphere/pkg/server/options"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
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
	identityprovider.RegisterGenericProviderFactory(&fakePasswordProviderFactory{})
	oauthOptions := &authentication.Options{
		Issuer: &oauth.IssuerOptions{},
	}

	fakepwd1 := &identityprovider.Configuration{
		Name:          "fakepwd1",
		MappingMethod: "manual",
		Type:          "fakePasswordProvider",
		ProviderOptions: options.DynamicOptions{
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
	}

	fakepwd2 := &identityprovider.Configuration{
		Name:          "fakepwd2",
		MappingMethod: "auto",
		Type:          "fakePasswordProvider",
		ProviderOptions: options.DynamicOptions{
			"identities": map[string]interface{}{
				"user5": map[string]string{
					"uid":      "100005",
					"email":    "user5@kubesphere.io",
					"username": "user5",
					"password": "password",
				},
			},
		},
	}

	fakepwd3 := &identityprovider.Configuration{
		Name:          "fakepwd3",
		MappingMethod: "lookup",
		Type:          "fakePasswordProvider",
		ProviderOptions: options.DynamicOptions{
			"identities": map[string]interface{}{
				"user6": map[string]string{
					"uid":      "100006",
					"email":    "user6@kubesphere.io",
					"username": "user6",
					"password": "password",
				},
			},
		},
	}

	marshal1, err := yaml.Marshal(fakepwd1)
	if err != nil {
		return
	}

	fakepwd1Secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fake-idp",
			Namespace: "kubesphere-system",
			Labels: map[string]string{
				constants.GenericConfigTypeLabel: identityprovider.ConfigTypeIdentityProvider,
			},
		},
		Data: map[string][]byte{
			"configuration.yaml": marshal1,
		},
		Type: identityprovider.SecretTypeIdentityProvider,
	}

	marshal2, err := yaml.Marshal(fakepwd2)
	if err != nil {
		return
	}
	fakepwd2Secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fake-idp2",
			Namespace: "kubesphere-system",
			Labels: map[string]string{
				constants.GenericConfigTypeLabel: identityprovider.ConfigTypeIdentityProvider,
			},
		},
		Data: map[string][]byte{
			"configuration.yaml": marshal2,
		},
		Type: identityprovider.SecretTypeIdentityProvider,
	}

	marshal3, err := yaml.Marshal(fakepwd3)
	if err != nil {
		return
	}
	fakepwd3Secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fake-idp3",
			Namespace: "kubesphere-system",
			Labels: map[string]string{
				constants.GenericConfigTypeLabel: identityprovider.ConfigTypeIdentityProvider,
			},
		},
		Data: map[string][]byte{
			"configuration.yaml": marshal3,
		},
		Type: identityprovider.SecretTypeIdentityProvider,
	}

	client := runtimefakeclient.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithRuntimeObjects(
			newUser("user1", "100001", "fakepwd1"),
			newUser("user3", "100003", ""),
			newActiveUser("user4", "password"),
			fakepwd1Secret,
			fakepwd2Secret,
			fakepwd3Secret,
		).
		Build()

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

	fakeSecretInformer.Add(fakepwd1Secret)
	fakeSecretInformer.Add(fakepwd2Secret)
	fakeSecretInformer.Add(fakepwd3Secret)

	authenticator := NewPasswordAuthenticator(client, oauthOptions)

	type args struct {
		ctx      context.Context
		username string
		password string
		provider string
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
				provider: "fakepwd1",
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
				provider: "fakepwd1",
			},
			want: &user.DefaultInfo{
				Name: "system:pre-registration",
				Extra: map[string][]string{
					"email":    {"user2@kubesphere.io"},
					"idp":      {"fakepwd1"},
					"uid":      {"100002"},
					"username": {"user2"},
				},
			},
			wantErr: false,
		},
		{
			name:                  "Should create user and return",
			passwordAuthenticator: authenticator,
			args: args{
				ctx:      context.Background(),
				username: "user5",
				password: "password",
				provider: "fakepwd2",
			},
			want:    &user.DefaultInfo{Name: "user5"},
			wantErr: false,
		},
		{
			name:                  "Should return user not found",
			passwordAuthenticator: authenticator,
			args: args{
				ctx:      context.Background(),
				username: "user6",
				password: "password",
				provider: "fakepwd3",
			},
			wantErr: true,
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
			got, err := p.Authenticate(tt.args.ctx, tt.args.provider, tt.args.username, tt.args.password)
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

func (fakePasswordProviderFactory) Create(dynamicOptions options.DynamicOptions) (identityprovider.GenericProvider, error) {
	var fakeProvider fakePasswordProvider
	if err := mapstructure.Decode(dynamicOptions, &fakeProvider); err != nil {
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

func newActiveUser(username string, password string) *iamv1beta1.User {
	u := newUser(username, "", "")
	password, _ = encrypt(password)
	u.Spec.EncryptedPassword = password
	u.Status.State = iamv1beta1.UserActive
	return u
}
