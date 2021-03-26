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
	"time"
)

const (
	AccessToken       TokenType = "access_token"
	RefreshToken      TokenType = "refresh_token"
	StaticToken       TokenType = "static_token"
	IDToken           TokenType = "id_token"
	AuthorizationCode TokenType = "authorization_code"
)

type TokenType string

// Issuer issues token to user, tokens are required to perform mutating requests to resources
type Issuer interface {
	// IssueTo issues a token, return error if issuing process failed
	IssueTo(claims *Claims, expiresIn time.Duration, customSecret string) (string, error)

	// Verify verifies a token, and return Claims if it's a valid token, otherwise return error
	Verify(token string, customSecret string) (*Claims, error)
}
