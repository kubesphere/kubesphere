/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"k8s.io/apimachinery/pkg/api/errors"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
)

type passwordAuthenticator struct {
	userGetter                          *userMapper
	client                              runtimeclient.Client
	authOptions                         *authentication.Options
	identityProviderConfigurationGetter identityprovider.ConfigurationGetter
}

func NewPasswordAuthenticator(cacheClient runtimeclient.Client, options *authentication.Options) PasswordAuthenticator {
	passwordAuthenticator := &passwordAuthenticator{
		client:                              cacheClient,
		userGetter:                          &userMapper{cache: cacheClient},
		identityProviderConfigurationGetter: identityprovider.NewConfigurationGetter(cacheClient),
		authOptions:                         options,
	}
	return passwordAuthenticator
}

func (p *passwordAuthenticator) Authenticate(ctx context.Context, provider, username, password string) (authuser.Info, error) {
	// empty username or password are not allowed
	if username == "" || password == "" {
		return nil, IncorrectPasswordError
	}
	if provider != "" {
		return p.authByProvider(ctx, provider, username, password)
	}
	return p.authByKubeSphere(ctx, username, password)
}

// authByKubeSphere authenticate by the kubesphere user
func (p *passwordAuthenticator) authByKubeSphere(ctx context.Context, username, password string) (authuser.Info, error) {
	user, err := p.userGetter.Find(ctx, username)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, IncorrectPasswordError
		}
		return nil, fmt.Errorf("failed to find user: %s", err)
	}

	if user == nil {
		return nil, IncorrectPasswordError
	}

	// check user status
	if user.Status.State != iamv1beta1.UserActive {
		if user.Status.State == iamv1beta1.UserAuthLimitExceeded {
			return nil, RateLimitExceededError
		} else {
			return nil, AccountIsNotActiveError
		}
	}

	// if the password is not empty, means that the password has been reset, even if the user was mapping from IDP
	if user.Spec.EncryptedPassword == "" {
		return nil, IncorrectPasswordError
	}

	if err = PasswordVerify(user.Spec.EncryptedPassword, password); err != nil {
		return nil, err
	}

	info := &authuser.DefaultInfo{
		Name:   user.Name,
		Groups: user.Spec.Groups,
	}

	// check if the password is initialized
	if uninitialized := user.Annotations[iamv1beta1.UninitializedAnnotation]; uninitialized != "" {
		info.Extra = map[string][]string{
			iamv1beta1.ExtraUninitialized: {uninitialized},
		}
	}

	return info, nil
}

// authByProvider authenticate by the third-party identity provider user
func (p *passwordAuthenticator) authByProvider(ctx context.Context, provider, username, password string) (authuser.Info, error) {
	genericProvider, exist := identityprovider.SharedIdentityProviderController.GetGenericProvider(provider)
	if !exist {
		return nil, fmt.Errorf("generic identity provider %s not found", provider)
	}

	providerConfig, err := p.identityProviderConfigurationGetter.GetConfiguration(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity provider configuration: %s", err)
	}

	identity, err := genericProvider.Authenticate(username, password)
	if err != nil {
		if errors.IsUnauthorized(err) {
			return nil, IncorrectPasswordError
		}
		return nil, err
	}

	mappedUser, err := p.userGetter.FindMappedUser(ctx, provider, identity.GetUserID())
	if err != nil {
		return nil, fmt.Errorf("failed to find mapped user: %s", err)
	}

	if mappedUser == nil {
		if providerConfig.MappingMethod == identityprovider.MappingMethodLookup {
			return nil, fmt.Errorf("failed to find mapped user: %s", identity.GetUserID())
		}

		if providerConfig.MappingMethod == identityprovider.MappingMethodManual {
			return newRreRegistrationUser(providerConfig.Name, identity), nil
		}

		if providerConfig.MappingMethod == identityprovider.MappingMethodAuto {
			mappedUser = newMappedUser(providerConfig.Name, identity)

			if err = p.client.Create(ctx, mappedUser); err != nil {
				return nil, err
			}

			return &authuser.DefaultInfo{Name: mappedUser.GetName()}, nil
		}

		return nil, fmt.Errorf("invalid mapping method found %s", providerConfig.MappingMethod)
	}

	if mappedUser.Status.State == iamv1beta1.UserDisabled {
		return nil, AccountIsNotActiveError
	}

	return &authuser.DefaultInfo{Name: mappedUser.GetName()}, nil
}

func PasswordVerify(encryptedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(password)); err != nil {
		return IncorrectPasswordError
	}
	return nil
}
