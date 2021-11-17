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

// AuthConfig is the core configuration.
// The first step to use this SDK is to use the InitConfig function to initialize the global authConfig.
type AuthConfig struct {
	Endpoint         string
	ClientId         string
	ClientSecret     string
	JwtPublicKey     string
	OrganizationName string
	ApplicationName  string
}

var authConfig AuthConfig

func InitConfig(endpoint string, clientId string, clientSecret string, jwtPublicKey string, organizationName string, applicationName string) {
	authConfig = AuthConfig{
		Endpoint:         endpoint,
		ClientId:         clientId,
		ClientSecret:     clientSecret,
		JwtPublicKey:     jwtPublicKey,
		OrganizationName: organizationName,
		ApplicationName:  applicationName,
	}
}
