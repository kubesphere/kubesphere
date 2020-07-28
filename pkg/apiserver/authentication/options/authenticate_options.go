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
	// AuthenticateRateLimiter defines under which circumstances we will block user.
	// A user will be blocked if his/her failed login attempt reaches AuthenticateRateLimiterMaxTries in
	// AuthenticateRateLimiterDuration for about AuthenticateRateLimiterDuration. For example,
	//   AuthenticateRateLimiterMaxTries: 5
	//   AuthenticateRateLimiterDuration: 10m
	// A user will be blocked for 10m if he/she logins with incorrect credentials for at least 5 times in 10m.
	AuthenticateRateLimiterMaxTries int           `json:"authenticateRateLimiterMaxTries" yaml:"authenticateRateLimiterMaxTries"`
	AuthenticateRateLimiterDuration time.Duration `json:"authenticateRateLimiterDuration" yaml:"authenticateRateLimiterDuration"`
	// Token verification maximum time difference
	MaximumClockSkew time.Duration `json:"maximumClockSkew" yaml:"maximumClockSkew"`
	// retention login history, records beyond this amount will be deleted
	LoginHistoryRetentionPeriod time.Duration `json:"loginHistoryRetentionPeriod" yaml:"loginHistoryRetentionPeriod"`
	// allow multiple users login from different location at the same time
	MultipleLogin bool `json:"multipleLogin" yaml:"multipleLogin"`
	// secret to sign jwt token
	JwtSecret string `json:"-" yaml:"jwtSecret"`
	// OAuthOptions defines options needed for integrated oauth plugins
	OAuthOptions *oauth.Options `json:"oauthOptions" yaml:"oauthOptions"`
	// KubectlImage is the image address we use to create kubectl pod for users who have admin access to the cluster.
	KubectlImage string `json:"kubectlImage" yaml:"kubectlImage"`
}

func NewAuthenticateOptions() *AuthenticationOptions {
	return &AuthenticationOptions{
		AuthenticateRateLimiterMaxTries: 5,
		AuthenticateRateLimiterDuration: time.Minute * 30,
		MaximumClockSkew:                10 * time.Second,
		LoginHistoryRetentionPeriod:     time.Hour * 24 * 7,
		OAuthOptions:                    oauth.NewOptions(),
		MultipleLogin:                   false,
		JwtSecret:                       "",
		KubectlImage:                    "kubesphere/kubectl:v1.0.0",
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
	fs.DurationVar(&options.LoginHistoryRetentionPeriod, "login-history-retention-period", s.LoginHistoryRetentionPeriod, "login-history-retention-period defines how long login history should be kept.")
	fs.DurationVar(&options.OAuthOptions.AccessTokenMaxAge, "access-token-max-age", s.OAuthOptions.AccessTokenMaxAge, "access-token-max-age control the lifetime of access tokens, 0 means no expiration.")
	fs.StringVar(&s.KubectlImage, "kubectl-image", s.KubectlImage, "Setup the image used by kubectl terminal pod")
	fs.DurationVar(&options.MaximumClockSkew, "maximum-clock-skew", s.MaximumClockSkew, "The maximum time difference between the system clocks of the ks-apiserver that issued a JWT and the ks-apiserver that verified the JWT.")
}
