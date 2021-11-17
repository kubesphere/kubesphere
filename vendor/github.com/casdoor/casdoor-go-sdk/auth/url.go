// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"fmt"
	"net/url"
	"strings"
)

func GetSignupUrl(enablePassword bool, redirectUri string) string {
	// redirectUri can be empty string if enablePassword == true (only password enabled signup page is required)
	if enablePassword {
		return fmt.Sprintf("%s/signup/%s", authConfig.Endpoint, authConfig.ApplicationName)
	} else {
		return strings.ReplaceAll(GetSigninUrl(redirectUri), "/login/oauth/authorize", "/signup/oauth/authorize")
	}
}

func GetSigninUrl(redirectUri string) string {
	// origin := "https://door.casbin.com"
	// redirectUri := fmt.Sprintf("%s/callback", origin)
	scope := "read"
	state := authConfig.ApplicationName
	return fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=%s",
		authConfig.Endpoint, authConfig.ClientId, url.QueryEscape(redirectUri), scope, state)
}

func GetUserProfileUrl(userName string, accessToken string) string {
	param := ""
	if accessToken != "" {
		param = fmt.Sprintf("?access_token=%s", accessToken)
	}
	return fmt.Sprintf("%s/users/%s/%s%s", authConfig.Endpoint, authConfig.OrganizationName, userName, param)
}

func GetMyProfileUrl(accessToken string) string {
	param := ""
	if accessToken != "" {
		param = fmt.Sprintf("?access_token=%s", accessToken)
	}
	return fmt.Sprintf("%s/account%s", authConfig.Endpoint, param)
}
