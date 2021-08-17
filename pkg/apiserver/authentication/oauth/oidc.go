/*

 Copyright 2021 The KubeSphere Authors.

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

package oauth

import (
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	// ScopeOpenID Verify that a scope parameter is present and contains the openid scope value.
	// If no openid scope value is present, the request may still be a valid OAuth 2.0 request,
	// but is not an OpenID Connect request.
	ScopeOpenID = "openid"
	// ScopeEmail This scope value requests access to the email and email_verified Claims.
	ScopeEmail = "email"
	// ScopeProfile This scope value requests access to the End-User's default profile Claims,
	// which are: name, family_name, given_name, middle_name, nickname, preferred_username,
	// profile, picture, website, gender, birthdate, zoneinfo, locale, and updated_at.
	ScopeProfile = "profile"
	// ScopePhone This scope value requests access to the phone_number and phone_number_verified Claims.
	ScopePhone = "phone"
	// ScopeAddress This scope value requests access to the address Claim.
	ScopeAddress    = "address"
	ResponseCode    = "code"
	ResponseIDToken = "id_token"
	ResponseToken   = "token"
)

var ValidScopes = []string{ScopeOpenID, ScopeEmail, ScopeProfile}
var ValidResponseTypes = []string{ResponseCode, ResponseIDToken, ResponseToken}

func IsValidScopes(scopes []string) bool {
	for _, scope := range scopes {
		if !sliceutil.HasString(ValidScopes, scope) {
			return false
		}
	}
	return true
}

func IsValidResponseTypes(responseTypes []string) bool {
	for _, responseType := range responseTypes {
		if !sliceutil.HasString(ValidResponseTypes, responseType) {
			return false
		}
	}
	return true
}
