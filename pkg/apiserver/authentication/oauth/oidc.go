/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
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
	ScopeProfile               = "profile"
	ResponseTypeCode           = "code"
	ResponseTypeIDToken        = "id_token"
	ResponseTypeToken          = "token"
	GrantTypePassword          = "password"
	GrantTypeRefreshToken      = "refresh_token"
	GrantTypeCode              = "code"
	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeOTP               = "otp"
)

var ValidScopes = []string{ScopeOpenID, ScopeEmail, ScopeProfile}
var ValidResponseTypes = []string{ResponseTypeCode, ResponseTypeIDToken, ResponseTypeToken}

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
