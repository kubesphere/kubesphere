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

package token

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"time"
)

const DefaultIssuerName = "kubesphere"

var (
	errInvalidToken = errors.New("invalid token")
	errTokenExpired = errors.New("expired token")
)

type Claims struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
	// Currently, we are not using any field in jwt.StandardClaims
	jwt.StandardClaims
}

type jwtTokenIssuer struct {
	name    string
	options *authoptions.AuthenticationOptions
	cache   cache.Interface
	keyFunc jwt.Keyfunc
}

func (s *jwtTokenIssuer) Verify(tokenString string) (User, error) {
	if len(tokenString) == 0 {
		return nil, errInvalidToken
	}

	clm := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, clm, s.keyFunc)
	if err != nil {
		return nil, err
	}

	// accessTokenMaxAge = 0 or token without expiration time means that the token will not expire
	// do not validate token cache
	if s.options.OAuthOptions.AccessTokenMaxAge > 0 && clm.ExpiresAt > 0 {
		_, err = s.cache.Get(tokenCacheKey(tokenString))

		if err != nil {
			if err == cache.ErrNoSuchKey {
				return nil, errTokenExpired
			}
			return nil, err
		}
	}

	return &user.DefaultInfo{Name: clm.Username, UID: clm.UID}, nil
}

func (s *jwtTokenIssuer) IssueTo(user User, expiresIn time.Duration) (string, error) {
	clm := &Claims{
		Username: user.GetName(),
		UID:      user.GetUID(),
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			Issuer:    s.name,
			NotBefore: time.Now().Unix(),
		},
	}

	if expiresIn > 0 {
		clm.ExpiresAt = clm.IssuedAt + int64(expiresIn.Seconds())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, clm)

	tokenString, err := token.SignedString([]byte(s.options.JwtSecret))

	if err != nil {
		klog.Error(err)
		return "", err
	}

	// 0 means no expiration.
	// validate token cache
	if s.options.OAuthOptions.AccessTokenMaxAge > 0 {
		err = s.cache.Set(tokenCacheKey(tokenString), tokenString, s.options.OAuthOptions.AccessTokenMaxAge)
		if err != nil {
			klog.Error(err)
			return "", err
		}
	}

	return tokenString, nil
}

func (s *jwtTokenIssuer) Revoke(token string) error {
	if s.options.OAuthOptions.AccessTokenMaxAge > 0 {
		return s.cache.Del(tokenCacheKey(token))
	}
	return nil
}

func NewJwtTokenIssuer(issuerName string, options *authoptions.AuthenticationOptions, cache cache.Interface) Issuer {
	return &jwtTokenIssuer{
		name:    issuerName,
		options: options,
		cache:   cache,
		keyFunc: func(token *jwt.Token) (i interface{}, err error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				return []byte(options.JwtSecret), nil
			} else {
				return nil, fmt.Errorf("expect token signed with HMAC but got %v", token.Header["alg"])
			}
		},
	}
}

func tokenCacheKey(token string) string {
	return fmt.Sprintf("kubesphere:tokens:%s", token)
}
