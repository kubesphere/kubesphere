/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package authentication

import (
	"errors"
	"time"

	"github.com/spf13/pflag"

	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/aliyunidaas"
	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/cas"
	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/github"
	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/ldap"
	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/oidc"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

type Options struct {
	// AuthenticateRateLimiter defines under which circumstances we will block user.
	// A user will be blocked if his/her failed login attempt reaches AuthenticateRateLimiterMaxTries in
	// AuthenticateRateLimiterDuration for about AuthenticateRateLimiterDuration. For example,
	//   AuthenticateRateLimiterMaxTries: 5
	//   AuthenticateRateLimiterDuration: 10m
	// A user will be blocked for 10m if he/she logins with incorrect credentials for at least 5 times in 10m.
	AuthenticateRateLimiterMaxTries int           `json:"authenticateRateLimiterMaxTries" yaml:"authenticateRateLimiterMaxTries"`
	AuthenticateRateLimiterDuration time.Duration `json:"authenticateRateLimiterDuration" yaml:"authenticateRateLimiterDuration"`

	// retention login history, records beyond this amount will be deleted
	LoginHistoryRetentionPeriod time.Duration `json:"loginHistoryRetentionPeriod" yaml:"loginHistoryRetentionPeriod"`
	// retention login history, records beyond this amount will be deleted
	// LoginHistoryMaximumEntries restricts for all kubesphere accounts and must be greater than AuthenticateRateLimiterMaxTries
	LoginHistoryMaximumEntries int `json:"loginHistoryMaximumEntries,omitempty" yaml:"loginHistoryMaximumEntries,omitempty"`
	// allow multiple users login from different location at the same time
	MultipleLogin bool `json:"multipleLogin" yaml:"multipleLogin"`

	// Issuer defines options needed for integrated oauth plugins
	Issuer *oauth.IssuerOptions `json:"issuer" yaml:"issuer"`
}

func NewOptions() *Options {
	return &Options{
		AuthenticateRateLimiterMaxTries: 5,
		AuthenticateRateLimiterDuration: time.Minute * 30,
		LoginHistoryRetentionPeriod:     time.Hour * 24 * 7,
		LoginHistoryMaximumEntries:      100,
		Issuer:                          oauth.NewIssuerOptions(),
		MultipleLogin:                   false,
	}
}

func (options *Options) Validate() []error {
	var errs []error
	if len(options.Issuer.JWTSecret) == 0 {
		errs = append(errs, errors.New("JWT secret MUST not be empty"))
	}
	if options.AuthenticateRateLimiterMaxTries > options.LoginHistoryMaximumEntries {
		errs = append(errs, errors.New("authenticateRateLimiterMaxTries MUST not be greater than loginHistoryMaximumEntries"))
	}
	return errs
}

func (options *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.IntVar(&options.AuthenticateRateLimiterMaxTries, "authenticate-rate-limiter-max-retries", s.AuthenticateRateLimiterMaxTries, "")
	fs.DurationVar(&options.AuthenticateRateLimiterDuration, "authenticate-rate-limiter-duration", s.AuthenticateRateLimiterDuration, "")
	fs.BoolVar(&options.MultipleLogin, "multiple-login", s.MultipleLogin, "Allow multiple login with the same account, disable means only one user can login at the same time.")
	fs.StringVar(&options.Issuer.JWTSecret, "jwt-secret", s.Issuer.JWTSecret, "Secret to sign jwt token, must not be empty.")
	fs.DurationVar(&options.LoginHistoryRetentionPeriod, "login-history-retention-period", s.LoginHistoryRetentionPeriod, "login-history-retention-period defines how long login history should be kept.")
	fs.IntVar(&options.LoginHistoryMaximumEntries, "login-history-maximum-entries", s.LoginHistoryMaximumEntries, "login-history-maximum-entries defines how many entries of login history should be kept.")
	fs.DurationVar(&options.Issuer.AccessTokenMaxAge, "access-token-max-age", s.Issuer.AccessTokenMaxAge, "access-token-max-age control the lifetime of access tokens, 0 means no expiration.")
	fs.DurationVar(&options.Issuer.MaximumClockSkew, "maximum-clock-skew", s.Issuer.MaximumClockSkew, "The maximum time difference between the system clocks of the ks-apiserver that issued a JWT and the ks-apiserver that verified the JWT.")
}
