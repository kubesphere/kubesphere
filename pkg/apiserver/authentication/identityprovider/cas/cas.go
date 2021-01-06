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

package cas

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"net/http"
	"strings"
)

func init() {
	identityprovider.RegisterOAuthProvider(&cas{})
}

type cas struct {
	RedirectURL        string `json:"redirectURL" yaml:"redirectURL"`
	CASServerURL       string `json:"casServerURL" yaml:"casServerURL"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
}
type casIdentity struct {
	User string `json:"user"`
}

func (c casIdentity) GetName() string {
	return strings.ToLower(c.User)
}

func (c casIdentity) GetEmail() string {
	return ""
}

func (c cas) Type() string {
	return "CASIdentityProvider"
}

func (c cas) Setup(options *oauth.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var cas cas
	if err := mapstructure.Decode(options, &cas); err != nil {
		return nil, err
	}
	return &cas, nil
}

func (c cas) IdentityExchange(ticket string) (identityprovider.Identity, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}
	serviceValidateURL := fmt.Sprintf("%s/serviceValidate?format=json&service=%s&ticket=%s", c.CASServerURL, c.RedirectURL, ticket)
	resp, err := client.Get(serviceValidateURL)
	if err != nil {
		return nil, fmt.Errorf("cas validate service failed: %v", err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cas read data failed: %v", err)
	}
	var casResponse casResponse
	err = json.Unmarshal(data, &casResponse)
	if err != nil {
		return nil, fmt.Errorf("cas cannot unmarshal response: %v", err)
	}
	if casResponse.ServiceResponse.AuthenticationFailure != nil {
		errorCode := casResponse.ServiceResponse.AuthenticationFailure["code"]
		return nil, fmt.Errorf("cas authentication failed: %v", errorCode)
	}

	return casResponse.ServiceResponse.AuthenticationSuccess, nil
}

type casResponse struct {
	ServiceResponse serviceResponse `json:"serviceResponse"`
}
type serviceResponse struct {
	AuthenticationFailure map[string]string `json:"authenticationFailure,omitempty"`
	AuthenticationSuccess casIdentity       `json:"authenticationSuccess,omitempty"`
}
