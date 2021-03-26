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
	"sync"
	"time"
)

var once = sync.Once{}

type Claims struct {
	Username  string              `json:"username,omitempty"`
	Groups    []string            `json:"groups,omitempty"`
	Extra     map[string][]string `json:"extra,omitempty"`
	TokenType TokenType           `json:"token_type"`
	// Currently, we are not using any field in jwt.StandardClaims
	jwt.StandardClaims `json:",inline"`

	// https://openid.net/specs/openid-connect-core-1_0.html#IDToken
	AuthorizingParty  string `json:"azp,omitempty"`
	AccessTokenHash   string `json:"at_hash,omitempty"`
	CodeHash          string `json:"c_hash,omitempty"`
	Nonce             string `json:"nonce,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     *bool  `json:"email_verified,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}

func (c *Claims) UserInfo() user.Info {
	userInfo := &user.DefaultInfo{Name: c.Username, Groups: c.Groups, Extra: c.Extra}
	return userInfo
}

func NewClaims(user user.Info) *Claims {
	claims := Claims{
		Username: user.GetName(),
		//Groups:   user.GetGroups(),
		Extra: user.GetExtra(),
	}
	return &claims
}

type jwtTokenIssuer struct {
	name   string
	secret []byte
}

func (s *jwtTokenIssuer) Verify(tokenString string, customSecret string) (*Claims, error) {
	clm := &Claims{}
	// verify token signature and expiration time
	_, err := jwt.ParseWithClaims(tokenString, clm, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			if customSecret != "" {
				return []byte(customSecret), nil
			} else {
				return s.secret, nil
			}
		} else {
			return nil, fmt.Errorf("expect token signed with HMAC but got %v", token.Header["alg"])
		}
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return clm, nil
}

func (s *jwtTokenIssuer) IssueTo(claims *Claims, expiresIn time.Duration, customSecret string) (string, error) {
	now := time.Now()
	claims.IssuedAt = now.Unix()
	claims.Issuer = s.name
	claims.NotBefore = now.Unix()

	if expiresIn > 0 {
		claims.ExpiresAt = now.Add(expiresIn).Unix()
	}

	signKey := s.secret
	if customSecret != "" {
		signKey = []byte(customSecret)
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(signKey)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	return tokenString, nil
}

func NewTokenIssuer(issuer, secret string, maximumClockSkew time.Duration) Issuer {
	once.Do(func() {
		jwt.TimeFunc = func() time.Time {
			return time.Now().Add(maximumClockSkew)
		}
	})
	return &jwtTokenIssuer{
		name:   issuer,
		secret: []byte(secret),
	}
}
