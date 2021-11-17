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
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
)

// GetOAuthToken gets the pivotal and necessary secret to interact with the Casdoor server
func GetOAuthToken(code string, state string) (*oauth2.Token, error) {
	config := oauth2.Config{
		ClientID:     authConfig.ClientId,
		ClientSecret: authConfig.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("%s/api/login/oauth/authorize", authConfig.Endpoint),
			TokenURL:  fmt.Sprintf("%s/api/login/oauth/access_token", authConfig.Endpoint),
			AuthStyle: oauth2.AuthStyleInParams,
		},
		//RedirectURL: redirectUri,
		Scopes: nil,
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return token, err
	}

	if strings.HasPrefix(token.AccessToken, "error:") {
		return nil, errors.New(strings.TrimLeft(token.AccessToken, "error: "))
	}

	return token, err
}
