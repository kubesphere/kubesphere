package iam

import (
	"github.com/spf13/pflag"
	"time"
)

type AuthenticationOptions struct {
	// authenticate rate limit will
	AuthenticateRateLimiterMaxTries int
	AuthenticateRateLimiterDuration time.Duration

	// maximum retries when authenticate failed
	MaxAuthenticateRetries int

	// token validation duration, will refresh token expiration for each user request
	TokenExpiration time.Duration

	// allow multiple users login at the same time
	MultipleLogin bool
}

func NewAuthenticateOptions() *AuthenticationOptions {
	return &AuthenticationOptions{
		AuthenticateRateLimiterMaxTries: 5,
		AuthenticateRateLimiterDuration: time.Minute * 30,
		MaxAuthenticateRetries:          0,
		TokenExpiration:                 0,
		MultipleLogin:                   false,
	}
}

func (options *AuthenticationOptions) Validate() []error {
	var errs []error
	return errs
}

func (options *AuthenticationOptions) AddFlags(fs *pflag.FlagSet, s *AuthenticationOptions) {
	fs.IntVar(&options.AuthenticateRateLimiterMaxTries, "authenticate-rate-limiter-max-retries", s.AuthenticateRateLimiterMaxTries, "")
	fs.DurationVar(&options.AuthenticateRateLimiterDuration, "authenticate-rate-limiter-duration", s.AuthenticateRateLimiterDuration, "")
	fs.IntVar(&options.MaxAuthenticateRetries, "authenticate-max-retries", s.MaxAuthenticateRetries, "")
	fs.DurationVar(&options.TokenExpiration, "token-expiration", s.TokenExpiration, "")
	fs.BoolVar(&options.MultipleLogin, "multiple-login", s.MultipleLogin, "")
}
