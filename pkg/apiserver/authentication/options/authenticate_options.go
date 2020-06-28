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

package options

import (
	"fmt"
	"github.com/spf13/pflag"
	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/github"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"time"
)

type AuthenticationOptions struct {
	// authenticate rate limit will
	AuthenticateRateLimiterMaxTries int           `json:"authenticateRateLimiterMaxTries" yaml:"authenticateRateLimiterMaxTries"`
	AuthenticateRateLimiterDuration time.Duration `json:"authenticationRateLimiterDuration" yaml:"authenticationRateLimiterDuration"`
	// allow multiple users login at the same time
	MultipleLogin bool `json:"multipleLogin" yaml:"multipleLogin"`
	// secret to signed jwt token
	JwtSecret string `json:"-" yaml:"jwtSecret"`
	// oauth options
	OAuthOptions *oauth.Options `json:"oauthOptions" yaml:"oauthOptions"`
}

func NewAuthenticateOptions() *AuthenticationOptions {
	return &AuthenticationOptions{
		AuthenticateRateLimiterMaxTries: 5,
		AuthenticateRateLimiterDuration: time.Minute * 30,
		OAuthOptions:                    oauth.NewOptions(),
		MultipleLogin:                   false,
		JwtSecret:                       "",
	}
}

func (options *AuthenticationOptions) Validate() []error {
	var errs []error
	if len(options.JwtSecret) == 0 {
		errs = append(errs, fmt.Errorf("jwt secret is empty"))
	}

	return errs
}

func (options *AuthenticationOptions) AddFlags(fs *pflag.FlagSet, s *AuthenticationOptions) {
	fs.IntVar(&options.AuthenticateRateLimiterMaxTries, "authenticate-rate-limiter-max-retries", s.AuthenticateRateLimiterMaxTries, "")
	fs.DurationVar(&options.AuthenticateRateLimiterDuration, "authenticate-rate-limiter-duration", s.AuthenticateRateLimiterDuration, "")
	fs.BoolVar(&options.MultipleLogin, "multiple-login", s.MultipleLogin, "Allow multiple login with the same account, disable means only one user can login at the same time.")
	fs.StringVar(&options.JwtSecret, "jwt-secret", s.JwtSecret, "Secret to sign jwt token, must not be empty.")
	fs.DurationVar(&options.OAuthOptions.AccessTokenMaxAge, "access-token-max-age", s.OAuthOptions.AccessTokenMaxAge, "AccessTokenMaxAgeSeconds  control the lifetime of access tokens, 0 means no expiration.")
}
