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
	"errors"
	"fmt"
	"time"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"

	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
)

// TokenManagementInterface Cache issued token, support revocation of tokens after issuance
type TokenManagementInterface interface {
	// Verify the given token and returns token.VerifiedResponse
	Verify(token string) (*token.VerifiedResponse, error)
	// IssueTo issue a token for the specified user
	IssueTo(request *token.IssueRequest) (string, error)
	// Revoke revoke the specified token
	Revoke(token string) error
	// RevokeAllUserTokens revoke all user tokens
	RevokeAllUserTokens(username string) error
	// Keys hold encryption and signing keys.
	Keys() *token.Keys
}

type tokenOperator struct {
	issuer  token.Issuer
	options *authentication.Options
	cache   cache.Interface
}

func (t tokenOperator) Revoke(token string) error {
	pattern := fmt.Sprintf("kubesphere:user:*:token:%s", token)
	if keys, err := t.cache.Keys(pattern); err != nil {
		klog.Error(err)
		return err
	} else if len(keys) > 0 {
		if err := t.cache.Del(keys...); err != nil {
			klog.Error(err)
			return err
		}
	}
	return nil
}

func NewTokenOperator(cache cache.Interface, issuer token.Issuer, options *authentication.Options) TokenManagementInterface {
	operator := &tokenOperator{
		issuer:  issuer,
		options: options,
		cache:   cache,
	}
	return operator
}

func (t *tokenOperator) Verify(tokenStr string) (*token.VerifiedResponse, error) {
	response, err := t.issuer.Verify(tokenStr)
	if err != nil {
		return nil, err
	}
	if t.options.OAuthOptions.AccessTokenMaxAge == 0 ||
		response.TokenType == token.StaticToken {
		return response, nil
	}
	if err := t.tokenCacheValidate(response.User.GetName(), tokenStr); err != nil {
		return nil, err
	}
	return response, nil
}

func (t *tokenOperator) IssueTo(request *token.IssueRequest) (string, error) {
	tokenStr, err := t.issuer.IssueTo(request)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	if request.ExpiresIn > 0 {
		if err = t.cacheToken(request.User.GetName(), tokenStr, request.ExpiresIn); err != nil {
			klog.Error(err)
			return "", err
		}
	}
	return tokenStr, nil
}

// RevokeAllUserTokens revoke all user tokens in the cache
func (t *tokenOperator) RevokeAllUserTokens(username string) error {
	pattern := fmt.Sprintf("kubesphere:user:%s:token:*", username)
	if keys, err := t.cache.Keys(pattern); err != nil {
		klog.Error(err)
		return err
	} else if len(keys) > 0 {
		if err := t.cache.Del(keys...); err != nil {
			klog.Error(err)
			return err
		}
	}
	return nil
}

func (t *tokenOperator) Keys() *token.Keys {
	return t.issuer.Keys()
}

// tokenCacheValidate verify that the token is in the cache
func (t *tokenOperator) tokenCacheValidate(username, token string) error {
	key := fmt.Sprintf("kubesphere:user:%s:token:%s", username, token)
	if exist, err := t.cache.Exists(key); err != nil {
		return err
	} else if !exist {
		return errors.New("token not found in cache")
	}
	return nil
}

// cacheToken cache the token for a period of time
func (t *tokenOperator) cacheToken(username, token string, duration time.Duration) error {
	key := fmt.Sprintf("kubesphere:user:%s:token:%s", username, token)
	if err := t.cache.Set(key, token, duration); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
