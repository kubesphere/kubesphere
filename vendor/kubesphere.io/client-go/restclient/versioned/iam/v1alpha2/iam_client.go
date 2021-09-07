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

package v1alpha2

import (
	resty "github.com/go-resty/resty/v2"
	rest "k8s.io/client-go/rest"
)

type IamV1alpha2Interface interface {
	GroupsGetter
	RoleBindingsGetter
}
type IamV1alpha2Client struct {
	client *resty.Client
}

func (c *IamV1alpha2Client) Groups() GroupInterface {
	return newGroups(c)
}

func (c *IamV1alpha2Client) RoleBindings() RoleBindingInterface {
	return newRoleBindings(c)
}

// NewForConfig creates a new IamV1alpha2Client for the given config.
func NewForConfig(c *rest.Config) (*IamV1alpha2Client, error) {

	client := resty.New()

	client.SetHostURL(c.Host)
	if c.BearerToken != "" {
		client.SetAuthToken(c.BearerToken)
	}

	if c.Username != "" {
		client.SetBasicAuth(c.Username, c.Password)
	}

	return &IamV1alpha2Client{client}, nil
}
